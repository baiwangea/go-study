// Package risk 的表驱动测试：把「名义金额上限 / 信号偏离 / 同方向冷却」三条规则的
// 当前行为钉住，给接下来的 float64 → decimal 重构当安全网。
//
// 写法上刻意贴近**可观察行为**而不是内部字段：
//   - 一律通过 New(...) 构造 Guard，不直接拼结构体字面量；
//   - 只断言 Check 的返回值（error 是否为 nil、errors.Is 命中的哨兵、错误信息里的实际值）；
//   - 冷却相关的用例显式传 now，避免依赖 time.Sleep 或真实时钟（慢且 flaky）。
//
// 这样签名从 float64 换成 decimal.Decimal 时，这里只需要改参数类型，断言逻辑不用动。

package risk

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"go-study/trading-bot/internal/strategy"
)

// 测试里统一用的固定时刻：不依赖 time.Now()，冷却期判定完全可控。
var baseTime = time.Unix(1_700_000_000, 0)

// 冷却相关用例统一用这一对「数量、价格」：数量与价格现在是 decimal（都是「钱和量」）。
var (
	qtyDefault   = decimal.RequireFromString("0.002")
	priceDefault = decimal.RequireFromString("60000")
)

// newGuard 构造一个「除了被考察的那条规则，其它规则都放行」的 Guard。
// maxNotional 传字符串：decimal 没有字面量，字符串是唯一无歧义的构造方式。
func newGuard(maxNotional string, cooldown time.Duration, maxSlippage float64) *Guard {
	return New(decimal.RequireFromString(maxNotional), cooldown, maxSlippage)
}

// sig 构造一个最小可用的信号（符号/方向/强度是 Check 唯一关心的三个字段）。
func sig(symbol, side string, strength float64) *strategy.Signal {
	return &strategy.Signal{Symbol: symbol, Side: side, Strength: strength}
}

// ─────────────────────────────────────────────────────────────────────────────
// 规则一 + 规则二：名义金额上限、信号偏离上限。
//
// 这两条是无状态的纯判断，适合一次性放进同一张表里跑。
// 边界用例（恰好等于上限）是最容易在重构里被写反的地方 —— `>` 写成 `>=` 就会
// 把「刚好卡在上限」的合法订单误杀，所以这里专门留了两条边界。
// ─────────────────────────────────────────────────────────────────────────────

func TestCheckNotionalAndSlippage(t *testing.T) {
	tests := []struct {
		name        string
		maxNotional string // decimal 的构造源（字符串）
		maxSlippage float64
		qty         string
		price       string
		strength    float64
		wantErr     bool
		wantMsg     string // 错误信息里必须包含的子串（实际值，便于排障）
	}{
		{
			name: "名义金额超上限被拦", maxNotional: "100", maxSlippage: 0.02,
			qty: "1", price: "200", strength: 0.001,
			wantErr: true, wantMsg: "200.00", // 实际名义金额
		},
		{
			// 边界：qty*price == MaxNotional，规则用的是 `>`，所以必须放行
			name: "名义金额恰好等于上限（边界通过）", maxNotional: "100", maxSlippage: 0.02,
			qty: "1", price: "100", strength: 0.001,
			wantErr: false,
		},
		{
			name: "名义金额略超上限被拦", maxNotional: "100", maxSlippage: 0.02,
			qty: "1", price: "100.01", strength: 0.001,
			wantErr: true, wantMsg: "100.01",
		},
		{
			name: "信号偏离超 MaxSlippage 被拦", maxNotional: "1000000000", maxSlippage: 0.02,
			qty: "0.002", price: "60000", strength: 0.05,
			wantErr: true, wantMsg: "0.0500", // 实际偏离强度
		},
		{
			// 边界：Strength == MaxSlippage，规则用的是 `>`，必须放行
			name: "信号偏离恰好等于 MaxSlippage（边界通过）", maxNotional: "1000000000", maxSlippage: 0.02,
			qty: "0.002", price: "60000", strength: 0.02,
			wantErr: false,
		},
		{
			name: "两条规则都合规时通过", maxNotional: "1000000000", maxSlippage: 0.02,
			qty: "0.002", price: "60000", strength: 0.001,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // 每条用例自己 newGuard，互不共享状态，可以并行

			g := newGuard(tt.maxNotional, time.Minute, tt.maxSlippage)
			err := g.Check(sig("BTCUSDT", "BUY", tt.strength),
				decimal.RequireFromString(tt.qty), decimal.RequireFromString(tt.price), baseTime)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("Check(qty=%v, price=%v, strength=%v) 期望被拦，实际通过", tt.qty, tt.price, tt.strength)
				}
				if tt.wantMsg != "" && !strings.Contains(err.Error(), tt.wantMsg) {
					t.Errorf("错误信息 = %q, 期望包含实际值 %q（排障时要能看出拦的是哪笔）", err, tt.wantMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("Check(qty=%v, price=%v, strength=%v) 意外被拦: %v", tt.qty, tt.price, tt.strength, err)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 规则三：同 symbol + 同 side 的冷却期。
//
// 冷却是有状态的，必须按顺序调用，所以单独写。用 errors.Is(err, ErrCooldown)
// 而不是比较错误字符串 —— 重构时错误信息文案可能被改写，哨兵不会。
// ─────────────────────────────────────────────────────────────────────────────

func TestCheckCooldown(t *testing.T) {
	// 冷却窗口用 1 分钟，配合显式 now，判定完全确定。
	const cooldown = time.Minute
	// 上限给得足够大，确保被考察的永远是冷却规则而不是金额/偏离规则。
	newCooldownGuard := func() *Guard { return newGuard("1000000000", cooldown, 1) }

	t.Run("冷却期内同 symbol+side 第二次被拒", func(t *testing.T) {
		g := newCooldownGuard()
		s := sig("BTCUSDT", "BUY", 0.001)

		if err := g.Check(s, qtyDefault, priceDefault, baseTime); err != nil {
			t.Fatalf("第一次应通过，实际: %v", err)
		}
		err := g.Check(s, qtyDefault, priceDefault, baseTime.Add(30*time.Second))
		if !errors.Is(err, ErrCooldown) {
			t.Fatalf("冷却期内第二次应返回 ErrCooldown，实际 err=%v", err)
		}
	})

	t.Run("冷却期外通过", func(t *testing.T) {
		g := newCooldownGuard()
		s := sig("BTCUSDT", "BUY", 0.001)

		if err := g.Check(s, qtyDefault, priceDefault, baseTime); err != nil {
			t.Fatalf("第一次应通过: %v", err)
		}
		if err := g.Check(s, qtyDefault, priceDefault, baseTime.Add(cooldown+time.Second)); err != nil {
			t.Fatalf("冷却期外应通过，实际: %v", err)
		}
	})

	t.Run("恰好等于冷却期（边界通过）", func(t *testing.T) {
		// 规则是 `now.Sub(last) < Cooldown`，所以「恰好等于」不算冷却期内，应放行。
		g := newCooldownGuard()
		s := sig("BTCUSDT", "BUY", 0.001)

		if err := g.Check(s, qtyDefault, priceDefault, baseTime); err != nil {
			t.Fatalf("第一次应通过: %v", err)
		}
		if err := g.Check(s, qtyDefault, priceDefault, baseTime.Add(cooldown)); err != nil {
			t.Fatalf("恰好等于冷却期应通过，实际: %v", err)
		}
	})

	t.Run("不同 symbol 互不影响", func(t *testing.T) {
		g := newCooldownGuard()
		if err := g.Check(sig("BTCUSDT", "BUY", 0.001), qtyDefault, priceDefault, baseTime); err != nil {
			t.Fatalf("BTCUSDT 第一次应通过: %v", err)
		}
		// ETHUSDT 是另一把钥匙，即便在同一时刻也不该被 BTCUSDT 的冷却拖累
		if err := g.Check(sig("ETHUSDT", "BUY", 0.001), qtyDefault, priceDefault, baseTime); err != nil {
			t.Fatalf("不同 symbol 不该被冷却拦截，实际: %v", err)
		}
	})

	t.Run("不同 side 互不影响", func(t *testing.T) {
		g := newCooldownGuard()
		if err := g.Check(sig("BTCUSDT", "BUY", 0.001), qtyDefault, priceDefault, baseTime); err != nil {
			t.Fatalf("BUY 第一次应通过: %v", err)
		}
		// 同 symbol 但方向相反：冷却键是 symbol+":"+side，所以互不影响
		if err := g.Check(sig("BTCUSDT", "SELL", 0.001), qtyDefault, priceDefault, baseTime); err != nil {
			t.Fatalf("不同 side 不该被冷却拦截，实际: %v", err)
		}
	})

	t.Run("第三次仍被拦且不影响其它键", func(t *testing.T) {
		g := newCooldownGuard()
		buy := sig("BTCUSDT", "BUY", 0.001)

		if err := g.Check(buy, qtyDefault, priceDefault, baseTime); err != nil {
			t.Fatalf("第一次应通过: %v", err)
		}
		// 连续两次都在冷却期内 → 两次都应被拒
		for i, at := range []time.Time{baseTime.Add(10 * time.Second), baseTime.Add(20 * time.Second)} {
			if err := g.Check(buy, qtyDefault, priceDefault, at); !errors.Is(err, ErrCooldown) {
				t.Fatalf("第 %d 次重复应返回 ErrCooldown，实际 err=%v", i+2, err)
			}
		}
		// 期间 SELL 方向依然是干净的
		if err := g.Check(sig("BTCUSDT", "SELL", 0.001), qtyDefault, priceDefault, baseTime.Add(20*time.Second)); err != nil {
			t.Fatalf("SELL 不该被 BUY 的冷却影响: %v", err)
		}
	})

	t.Run("被金额规则拦下的信号不占用冷却窗口（当前行为）", func(t *testing.T) {
		// ⚠ 这一条记录的是**当前实现顺序**：金额/偏离检查排在冷却检查之前，
		//   而且拦截时直接 return，不会写入 LastBySymbol。
		//   也就是说「被金额拦下」不消耗冷却额度 —— 下一笔合规订单不会被它连累。
		//   重构时如果调整了检查顺序，这条测试会立刻变红，提醒我们这是个行为变更。
		g := newGuard("100", cooldown, 1)
		s := sig("BTCUSDT", "BUY", 0.001)

		if err := g.Check(s, decimal.RequireFromString("1"), decimal.RequireFromString("200"), baseTime); err == nil {
			t.Fatal("超上限的订单应被拦下")
		}
		if err := g.Check(s, decimal.RequireFromString("0.001"), decimal.RequireFromString("100"), baseTime); err != nil {
			t.Fatalf("上一笔被金额拦下后，合规订单不该进入冷却，实际: %v", err)
		}
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// 并发安全：Check 会被多个 worker 并发调用，而 LastBySymbol 是裸 map —— 必须加锁。
//
// 配 `go test -race` 跑：没加锁时下面的并发写 map 会被 -race 直接抓出来
// （Go 运行时对并发写 map 还有可能直接 fatal，不是「偶尔算错」那么温柔）。
//
// 除了「不崩」，这里还断言一条**语义**：冷却判定必须原子 —— 同一时刻 N 个同 key
// 请求只能有一个拿到放行。检查与写冷却时间之间有窗口的话，两个信号会双双放行。
// ─────────────────────────────────────────────────────────────────────────────

func TestCheckConcurrentSameKeyOnlyOnePasses(t *testing.T) {
	const workers = 64

	// 上限与偏离都给得足够宽，确保唯一能拦下请求的是冷却规则
	g := newGuard("1000000000", time.Minute, 1)
	s := sig("BTCUSDT", "BUY", 0.001)

	var (
		mu     sync.Mutex
		passed int
		denied int
	)
	start := make(chan struct{}) // 用闸门让所有 goroutine 尽量同时冲进 Check
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			<-start
			err := g.Check(s, qtyDefault, priceDefault, baseTime)

			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				passed++
			case errors.Is(err, ErrCooldown):
				denied++
			}
		}()
	}
	close(start)
	wg.Wait()

	if passed != 1 {
		t.Errorf("并发 %d 个同 key 请求放行 %d 个, want 恰好 1（冷却判定必须原子）", workers, passed)
	}
	if denied != workers-1 {
		t.Errorf("被冷却拦下 %d 个, want %d", denied, workers-1)
	}
}

// 不同 key 的并发读写：这是 -race 最直接的守护测试（并发写 map 必炸）。
func TestCheckConcurrentDistinctKeys(t *testing.T) {
	const (
		workers = 32
		perG    = 50
	)
	g := newGuard("1000000000", time.Minute, 1)

	var failures atomic.Int64
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()
			s := sig(fmt.Sprintf("SYM-%d", i), "BUY", 0.001)
			for j := 0; j < perG; j++ {
				// 每次往后推 1 分钟（= 冷却期），恰好越过冷却，不该被拦
				at := baseTime.Add(time.Duration(j) * time.Minute)
				if err := g.Check(s, qtyDefault, priceDefault, at); err != nil {
					failures.Add(1)
				}
			}
		}(i)
	}
	wg.Wait()

	if n := failures.Load(); n != 0 {
		t.Errorf("并发检查失败 %d 次, want 0（同一 key 每次间隔恰好一个冷却期，不该被拦）", n)
	}
}
