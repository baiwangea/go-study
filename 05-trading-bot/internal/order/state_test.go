// Package order 的测试：把状态机的**整张跃迁表**钉死。
//
// 断言策略：合法跃迁必须放行（且返回目标状态）、非法跃迁必须被拒（且返回原状态），
// 6×6 共 36 种组合全部覆盖 —— 少写一条边、多写一条边都会立刻变红。
//
// 另外三组：
//   - 终态不可出（filled / canceled / rejected 只能保持自身）；
//   - 幂等（Transition(s, s) 是「无变化」，不是错误）；
//   - 不变量（FilledQty < 0 或 > Qty 必须报错）。
//
// 只用标准库 + decimal，不引入 testify。

package order

import (
	"errors"
	"slices"
	"testing"

	"github.com/shopspring/decimal"
)

// legal 是期望的完整跃迁表（含「同状态 = 无变化」这一条幂等边）。
var legal = map[State]map[State]bool{
	Created:   {Created: true, Submitted: true, Rejected: true},
	Submitted: {Submitted: true, Partial: true, Filled: true, Canceled: true},
	Partial:   {Partial: true, Filled: true, Canceled: true},
	Filled:    {Filled: true},
	Canceled:  {Canceled: true},
	Rejected:  {Rejected: true},
}

// ─────────────────────────────────────────────────────────────────────────────
// 全跃迁表：36 种组合逐一验证 CanTransition 与 Transition 一致。
// ─────────────────────────────────────────────────────────────────────────────

func TestTransitionTable(t *testing.T) {
	states := AllStates()
	if len(states) != 6 {
		t.Fatalf("AllStates() 返回 %d 个状态, want 6", len(states))
	}

	for _, from := range states {
		for _, to := range states {
			want := legal[from][to]
			name := string(from) + "->" + string(to)

			t.Run(name, func(t *testing.T) {
				t.Parallel()

				if got := CanTransition(from, to); got != want {
					t.Errorf("CanTransition(%s, %s) = %t, want %t", from, to, got, want)
				}

				got, err := Transition(from, to)
				if want {
					if err != nil {
						t.Fatalf("Transition(%s, %s) 应放行，实际报错: %v", from, to, err)
					}
					if got != to {
						t.Errorf("Transition(%s, %s) 返回 %s, want %s", from, to, got, to)
					}
					return
				}
				if err == nil {
					t.Fatalf("Transition(%s, %s) 应被拒，实际放行并返回 %s", from, to, got)
				}
				if !errors.Is(err, ErrIllegalTransition) {
					t.Errorf("Transition(%s, %s) 错误 = %v, 期望 errors.Is(_, ErrIllegalTransition)", from, to, err)
				}
				// 出错时必须返回**原状态**，调用方才能安全地沿用
				if got != from {
					t.Errorf("非法跃迁返回 %s, want 原状态 %s", got, from)
				}
			})
		}
	}
}

// 跃迁表里声明的每条边都必须真的可达，不能有「表里有、函数不认」的死边。
func TestAllowedMatchesTable(t *testing.T) {
	for _, from := range AllStates() {
		set := map[State]bool{}
		for _, to := range Allowed(from) {
			set[to] = true
		}
		// 非「同状态」的边必须与跃迁表完全一致（同状态是幂等，未必要出现在出边列表里）
		for _, to := range AllStates() {
			if to == from {
				continue
			}
			if legal[from][to] != set[to] {
				t.Errorf("Allowed(%s) 含 %s = %t, 跃迁表 = %t", from, to, set[to], legal[from][to])
			}
		}
	}

	// partial → partial 是显式声明的边（继续部分成交），必须出现在出边列表里
	if !slices.Contains(Allowed(Partial), Partial) {
		t.Errorf("Allowed(partial) = %v, 应包含 partial（继续部分成交）", Allowed(Partial))
	}
	// 终态没有出边
	for _, term := range []State{Filled, Canceled, Rejected} {
		if got := Allowed(term); len(got) != 0 {
			t.Errorf("Allowed(%s) = %v, 终态不该有出边", term, got)
		}
	}
	// 非法起点返回 nil
	if got := Allowed(State("nope")); got != nil {
		t.Errorf("Allowed(非法状态) = %v, want nil", got)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 终态不可出：filled / canceled / rejected 不能到任何**别的**状态。
// ─────────────────────────────────────────────────────────────────────────────

func TestTerminalStatesCannotLeave(t *testing.T) {
	terminals := []State{Filled, Canceled, Rejected}

	for _, term := range terminals {
		if !IsTerminal(term) {
			t.Errorf("IsTerminal(%s) = false, want true", term)
		}
		for _, to := range AllStates() {
			if to == term {
				continue // 同状态是幂等无变化，允许
			}
			if CanTransition(term, to) {
				t.Errorf("CanTransition(%s, %s) = true, 终态不该有出边", term, to)
			}
			if _, err := Transition(term, to); err == nil {
				t.Errorf("Transition(%s, %s) 应被拒，实际放行", term, to)
			}
		}
	}

	// 非终态也要判对
	for _, s := range []State{Created, Submitted, Partial} {
		if IsTerminal(s) {
			t.Errorf("IsTerminal(%s) = true, want false", s)
		}
	}
	if IsTerminal(State("nope")) {
		t.Error("IsTerminal(非法状态) 应为 false")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 幂等：Transition(s, s) 对每个状态都是「无变化」而不是错误 ——
// 交易所的成交回调会重复推送，重复的 filled 不该把管线打成错误。
// ─────────────────────────────────────────────────────────────────────────────

func TestTransitionIdempotent(t *testing.T) {
	for _, s := range AllStates() {
		if !CanTransition(s, s) {
			t.Errorf("CanTransition(%s, %s) = false, 同状态应视为无变化", s, s)
		}
		got, err := Transition(s, s)
		if err != nil {
			t.Errorf("Transition(%s, %s) 报错: %v, 幂等场景不该报错", s, s, err)
		}
		if got != s {
			t.Errorf("Transition(%s, %s) = %s, want 原状态 %s", s, s, got, s)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 未知状态：跃迁表里没定义的状态必须报 ErrUnknownState，而不是被当成合法。
// ─────────────────────────────────────────────────────────────────────────────

func TestTransitionUnknownState(t *testing.T) {
	tests := []struct {
		name     string
		from, to State
		invalid  State // 这一对里哪个是非法状态
	}{
		{"起点未知", State("nope"), Submitted, State("nope")},
		{"终点未知", Created, State("nope"), State("nope")},
		{"空字符串起点", State(""), Submitted, State("")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if CanTransition(tt.from, tt.to) {
				t.Errorf("CanTransition(%q, %q) = true, want false", tt.from, tt.to)
			}
			got, err := Transition(tt.from, tt.to)
			if !errors.Is(err, ErrUnknownState) {
				t.Errorf("Transition(%q, %q) 错误 = %v, 期望 errors.Is(_, ErrUnknownState)", tt.from, tt.to, err)
			}
			if got != tt.from {
				t.Errorf("返回 %q, want 原状态 %q", got, tt.from)
			}
			if Valid(tt.invalid) {
				t.Errorf("Valid(%q) = true, 该状态不该出现在跃迁表里", tt.invalid)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 成交不变量：0 ≤ FilledQty ≤ Qty。
// ─────────────────────────────────────────────────────────────────────────────

func TestCheckFillInvariant(t *testing.T) {
	tests := []struct {
		name    string
		qty     string
		filled  string
		wantErr bool
	}{
		{"零成交合法", "1", "0", false},
		{"部分成交合法", "1", "0.4", false},
		{"全部成交合法", "1", "1", false},
		{"负成交量非法", "1", "-0.1", true},
		{"超过下单数量非法", "1", "1.00000001", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := CheckFill(decimal.RequireFromString(tt.qty), decimal.RequireFromString(tt.filled))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("CheckFill(%s, %s) 应报错，实际通过", tt.qty, tt.filled)
				}
				if !errors.Is(err, ErrFillInvariant) {
					t.Errorf("错误 = %v, 期望 errors.Is(_, ErrFillInvariant)", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("CheckFill(%s, %s) 意外报错: %v", tt.qty, tt.filled, err)
			}
		})
	}
}

// StateForFill：按累计成交量推导目标状态。
func TestStateForFill(t *testing.T) {
	tests := []struct {
		name    string
		qty     string
		filled  string
		want    State
		wantErr bool
	}{
		{"还没成交 → submitted", "1", "0", Submitted, false},
		{"部分成交 → partial", "1", "0.4", Partial, false},
		{"补满 → filled", "1", "1", Filled, false},
		{"超出数量 → 报错", "1", "2", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := StateForFill(decimal.RequireFromString(tt.qty), decimal.RequireFromString(tt.filled))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("StateForFill(%s, %s) 应报错，实际返回 %s", tt.qty, tt.filled, got)
				}
				if !errors.Is(err, ErrFillInvariant) {
					t.Errorf("错误 = %v, 期望 errors.Is(_, ErrFillInvariant)", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("StateForFill(%s, %s) 意外报错: %v", tt.qty, tt.filled, err)
			}
			if got != tt.want {
				t.Errorf("StateForFill(%s, %s) = %s, want %s", tt.qty, tt.filled, got, tt.want)
			}
		})
	}
}

// AllStates 返回的是副本：调用方改它不该污染内部表。
func TestAllStatesReturnsCopy(t *testing.T) {
	first := AllStates()
	first[0] = State("mutated")
	if second := AllStates(); second[0] == State("mutated") {
		t.Error("AllStates() 返回了内部切片，调用方能改坏跃迁表")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 数量按 stepSize 向下对齐：decimal 版。
//
// 关键边界 0.3 / 0.1：float64 下 math.Mod(0.3, 0.1) ≈ 0.09999999999999998，
// 对齐结果会掉到 0.2（少卖一个 step）。decimal 里 0.3 与 0.1 都是精确十进制数，取模为 0。
// ─────────────────────────────────────────────────────────────────────────────

func TestAlignQtyDown(t *testing.T) {
	tests := []struct {
		name string
		qty  string
		step string
		want string
	}{
		{"刚好整数个 step：0.3 / 0.1 必须原样保留", "0.3", "0.1", "0.3"},
		{"多出半个 step 要向下取整", "0.35", "0.1", "0.3"},
		{"常规币安 step", "0.002", "0.0001", "0.002"},
		{"不足一个 step 直接丢掉", "0.00009", "0.0001", "0"},
		{"step 不是 10 的幂", "1", "0.3", "0.9"},
		{"step 为 0 视为无约束", "5", "0", "5"},
		{"step 为负视为无约束", "5", "-1", "5"},
		{"零数量", "0", "0.1", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := AlignQtyDown(decimal.RequireFromString(tt.qty), decimal.RequireFromString(tt.step))
			want := decimal.RequireFromString(tt.want)
			if !got.Equal(want) {
				t.Errorf("AlignQtyDown(%s, %s) = %s, want %s", tt.qty, tt.step, got, want)
			}
		})
	}

	// 与 float64 的对照：同样想「对齐到 0.1」，float64 会把 0.3 削成 0.2。
	t.Run("float64 对照：0.3 会被削掉一个 step", func(t *testing.T) {
		const qty, step = 0.3, 0.1
		gotFloat := qty - floatMod(qty, step)
		if gotFloat == qty {
			t.Errorf("float64 对齐 0.3/0.1 得到 %v（恰好没误差），本用例的前提不成立", gotFloat)
		}
		if got := AlignQtyDown(decimal.RequireFromString("0.3"), decimal.RequireFromString("0.1")); !got.Equal(decimal.RequireFromString("0.3")) {
			t.Errorf("decimal 对齐 0.3/0.1 = %s, want 0.3", got)
		}
	})
}

// floatMod 是 math.Mod 的等价实现，写成局部函数免得为了一行引入 math。
func floatMod(a, b float64) float64 {
	q := a / b
	trunc := float64(int64(q))
	return a - trunc*b
}

// ─────────────────────────────────────────────────────────────────────────────
// 金额累加不产生误差：decimal 精确，float64 会漂。
// ─────────────────────────────────────────────────────────────────────────────

func TestDecimalSumHasNoError(t *testing.T) {
	const n = 10

	var decSum decimal.Decimal
	var fSum float64
	for i := 0; i < n; i++ {
		decSum = decSum.Add(decimal.RequireFromString("0.1"))
		fSum += 0.1
	}

	// decimal：10 个 0.1 精确等于 1
	if !decSum.Equal(decimal.NewFromInt(1)) {
		t.Errorf("decimal 累加 10 次 0.1 = %s, want 1", decSum)
	}

	// float64：同样累加会漂，差多少就打多少，这是「金额不能用 float64」的直接证据
	diff := fSum - 1
	t.Logf("对照：float64 累加 10 次 0.1 = %.20f，与 1 相差 %.3e；decimal = %s（精确）", fSum, diff, decSum)
	if diff == 0 {
		t.Error("float64 累加恰好等于 1，本用例的前提（存在误差）不成立，请换一组数字")
	}

	// 另一个更直观的例子：0.1 + 0.2 != 0.3。
	// 注意必须用 float64 变量：写成常量 `0.1+0.2 == 0.3` 会被 Go 的常量精确运算折叠成 true，
	// 那样根本测不出浮点误差（这个坑本身也值得记一笔）。
	var a, b float64 = 0.1, 0.2
	decTwo := decimal.RequireFromString("0.1").Add(decimal.RequireFromString("0.2"))
	if !decTwo.Equal(decimal.RequireFromString("0.3")) {
		t.Errorf("decimal 0.1+0.2 = %s, want 0.3", decTwo)
	}
	if a+b == 0.3 {
		t.Error("float64 的 0.1+0.2 恰好等于 0.3，本用例的前提不成立")
	}
	t.Logf("对照：float64 的 0.1+0.2 = %.20f（!= 0.3）；decimal = %s（精确）", a+b, decTwo)
}
