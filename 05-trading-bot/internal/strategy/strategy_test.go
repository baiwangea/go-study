// Package strategy 的测试：给 MACross 喂**自己构造的确定价格序列**，
// 断言它吐出的信号方向、是否出信号、以及 Peak/Strength 的量级。
//
// 刻意不依赖随机数：随机游走跑出来的信号方向不可预测，写不出稳定断言
// （这正是 README 里「阈值与波动率不匹配 → 信号永远为 0」那个坑的测试版教训）。
//
// 只通过 NewMACross(...) 构造、只断言 OnQuote 的返回值，所以内部 prices 从
// []float64 换成别的类型也不影响这些用例。

package strategy

import (
	"math"
	"testing"
	"time"

	"go-study/trading-bot/internal/market"
)

// quote 造一根行情。Symbol 会原样出现在信号里，价格才是策略真正吃的输入。
func quote(symbol string, price float64) market.Quote {
	return market.Quote{Symbol: symbol, Price: price, Ts: time.Now()}
}

// assertNear 比较浮点是否足够接近（照抄 06/01 的 helper 写法）。
func assertNear(t *testing.T, name string, got, want float64) {
	t.Helper()
	const eps = 1e-9
	if math.Abs(got-want) > eps {
		t.Errorf("%s: got %.12f, want %.12f (diff=%.3e)", name, got, want, got-want)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 核心行为：给定价格序列，产出哪些信号。
//
// window=2 时，第 3 根行情开始才够算「两根均线」：
//
//	cur  = avg(最近 2 根)      prev = avg(前 2 根)
//	dev  = (cur - prev) / prev → dev>0 发 BUY，dev<0 发 SELL，|dev|<阈值 不发
//
// 期望值全部手算写在注释里，避免「照着实现写断言」这种自证循环。
// ─────────────────────────────────────────────────────────────────────────────

func TestMACrossSignals(t *testing.T) {
	const (
		window    = 2
		threshold = 0.01
	)

	tests := []struct {
		name    string
		prices  []float64
		wantSid []string // 按顺序期望产出的方向；nil 表示不该有任何信号
	}{
		{
			// [100,100,110]：cur=avg(100,110)=105，prev=avg(100,100)=100，dev=+0.05 ≥ 0.01 → BUY
			name:    "上涨触发 BUY",
			prices:  []float64{100, 100, 110},
			wantSid: []string{"BUY"},
		},
		{
			// [100,100,90]：cur=avg(100,90)=95，prev=100，dev=-0.05 → SELL
			name:    "下跌触发 SELL",
			prices:  []float64{100, 100, 90},
			wantSid: []string{"SELL"},
		},
		{
			// 完全横盘：dev=0，不够阈值 → 不出信号
			name:    "横盘不触发",
			prices:  []float64{100, 100, 100},
			wantSid: nil,
		},
		{
			// dev=(100.25-100)/100=0.0025 < 0.01 → 波动不足，不出信号
			name:    "波动不足阈值不触发",
			prices:  []float64{100, 100, 100.5},
			wantSid: nil,
		},
		{
			// 3rd：[100,100,110] → BUY(0.05)
			// 4th：窗口滑动成 [100,110,120]，cur=115，prev=105，dev=10/105≈0.0952 → BUY
			name:    "连续上涨连发两次 BUY",
			prices:  []float64{100, 100, 110, 120},
			wantSid: []string{"BUY", "BUY"},
		},
		{
			// 3rd → BUY(0.05)；4th 窗口 [100,110,100]，cur=105，prev=105，dev=0 → 不再发
			name:    "冲高回落后停止发信号",
			prices:  []float64{100, 100, 110, 100},
			wantSid: []string{"BUY"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := NewMACross(window, threshold)
			var got []string
			for _, p := range tt.prices {
				sig, ok := s.OnQuote(quote("BTCUSDT", p))
				if !ok {
					continue
				}
				if sig == nil {
					t.Fatal("OnQuote 返回 ok=true 却给了 nil 信号")
				}
				// 信号强度必须是非负的偏离幅度，否则下游风控的 MaxSlippage 比较会失真
				if sig.Strength < 0 {
					t.Errorf("Strength = %v, 不该为负", sig.Strength)
				}
				if sig.Symbol != "BTCUSDT" {
					t.Errorf("Signal.Symbol = %q, want %q", sig.Symbol, "BTCUSDT")
				}
				got = append(got, sig.Side)
			}

			if len(got) != len(tt.wantSid) {
				t.Fatalf("信号序列 = %v, want %v", got, tt.wantSid)
			}
			for i := range got {
				if got[i] != tt.wantSid[i] {
					t.Errorf("第 %d 个信号 = %q, want %q（完整序列 %v）", i+1, got[i], tt.wantSid[i], got)
				}
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 数据不足：喂不够 window+1 根时，一根信号都不该出（算不出两根均线）。
// ─────────────────────────────────────────────────────────────────────────────

func TestMACrossNotEnoughData(t *testing.T) {
	const window = 5
	s := NewMACross(window, 0.0001) // 阈值给得很小，排除「因为波动不够而不发」的干扰

	// 喂满 window 根（=5），依然差 1 根才算得出两根均线
	for i := 0; i < window; i++ {
		if sig, ok := s.OnQuote(quote("BTCUSDT", 100+float64(i)*10)); ok {
			t.Fatalf("喂了 %d 根（window=%d）不该出信号，却得到 %+v", i+1, window, sig)
		}
	}

	// 第 window+1 根：价格明显跳升，应立刻出 BUY
	sig, ok := s.OnQuote(quote("BTCUSDT", 1000))
	if !ok {
		t.Fatalf("喂满 %d 根后应能出信号", window+1)
	}
	if sig.Side != "BUY" {
		t.Errorf("价格跳升后 Side = %q, want BUY", sig.Side)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Peak：实测到的最大均线偏离，用来校准阈值（README 里反复强调的读数）。
//
// 它必须反映「所有 tick 里见过的最大 |dev|」，且**只增不减**。
// ─────────────────────────────────────────────────────────────────────────────

func TestMACrossPeak(t *testing.T) {
	const (
		window    = 2
		threshold = 0.01
	)
	s := NewMACross(window, threshold)

	if got := s.Peak(); got != 0 {
		t.Errorf("还没喂数据时 Peak = %v, want 0", got)
	}

	// 3 根后 dev = +0.05
	s.OnQuote(quote("BTCUSDT", 100))
	s.OnQuote(quote("BTCUSDT", 100))
	s.OnQuote(quote("BTCUSDT", 110))
	assertNear(t, "首次偏离", s.Peak(), 0.05)

	// 第 4 根：窗口 [100,110,120]，cur=115，prev=105，dev = 10/105 ≈ 0.095238...
	s.OnQuote(quote("BTCUSDT", 120))
	assertNear(t, "更大的偏离被记录", s.Peak(), 10.0/105.0)

	// 第 5 根：窗口 [110,120,121]，cur=120.5，prev=115，dev = 5.5/115 ≈ 0.047826
	// 比峰值小 → Peak 必须保持不变（只记录最大值）
	s.OnQuote(quote("BTCUSDT", 121))
	assertNear(t, "更小的偏离不该拉低峰值", s.Peak(), 10.0/105.0)
}

// 阈值可以高到永远不出信号 —— 这正是 README 里那个「管线在空转」的坑，
// 但 Peak 依然会把真实波动记录下来，供人回头调阈值。
func TestMACrossHighThresholdStillReportsPeak(t *testing.T) {
	s := NewMACross(2, 1.0) // 阈值 100%，绝无可能触发

	s.OnQuote(quote("BTCUSDT", 100))
	s.OnQuote(quote("BTCUSDT", 100))
	if _, ok := s.OnQuote(quote("BTCUSDT", 110)); ok {
		t.Fatal("阈值远高于实际波动时不该出信号")
	}
	if got := s.Peak(); got <= 0 {
		t.Errorf("即便不出信号，Peak 也应记录到真实波动，实际 = %v", got)
	}
	assertNear(t, "高阈值下的实测峰值", s.Peak(), 0.05)
}

// Name() 是行为接口探测用的（main 里靠它打印策略名），固定值钉一下。
func TestMACrossName(t *testing.T) {
	if got := NewMACross(5, 0.0004).Name(); got != "ma-cross" {
		t.Errorf("Name() = %q, want %q", got, "ma-cross")
	}
}
