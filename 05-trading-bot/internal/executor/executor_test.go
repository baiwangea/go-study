// Package executor 的测试：用「手写测试替身」给 Exchange 打桩，把 Engine 的
// 幂等 / 重试 / 超时 / 整体取消四条路径全部覆盖。
//
// 手法照搬 06-go-testing/07_stub_mock/trader_test.go：不引入任何 mock 框架，
// 三个桩加起来不到 40 行，全是普通代码，编译期就能检查。
//
// 关键点：**桩必须真正响应 ctx.Done()**。context 的取消是协作式的 ——
// 实现不检查 ctx，超时/取消就永远不会生效（L2-11）。
//
// 断言尽量落在「可观察行为」上：订单的 Status/Attempts、返回的错误、以及
// **交易所到底被调用了几次**（spy 才能告诉你的信息）。这样 float64 → decimal
// 重构时，Order 里的 Qty 换类型不影响这些断言。

package executor

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"go-study/trading-bot/internal/order"
)

// ─────────────────────────────────────────────────────────────────────────────
// 桩 1：计数桩 —— 记录被真正调用了几次，并能模拟「前 N 次失败」。
//
// 用 sync.Mutex 保护计数：即便将来测试里出现并发提交，也不会变成 DATA RACE。
// ─────────────────────────────────────────────────────────────────────────────

type countingExchange struct {
	mu       sync.Mutex
	calls    int
	failures int   // 前 failures 次调用返回 err，之后返回 nil
	err      error // 失败时返回的错误
}

func (c *countingExchange) Submit(_ context.Context, _ *Order) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls++
	if c.calls <= c.failures {
		return c.err
	}
	return nil
}

// Calls 返回交易所被调用的总次数 —— 这是「幂等有没有真的省掉一次下单」的唯一证据。
func (c *countingExchange) Calls() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.calls
}

// ─────────────────────────────────────────────────────────────────────────────
// 桩 2：阻塞桩 —— 一直等到 ctx.Done() 才返回，用来测「单次调用超时」。
//
// 它绝不会自己返回 nil，所以能确定地把 Engine 的 perCall 超时分支逼出来。
// ─────────────────────────────────────────────────────────────────────────────

type blockingExchange struct{ calls atomic.Int32 }

func (b *blockingExchange) Submit(ctx context.Context, _ *Order) error {
	b.calls.Add(1)
	<-ctx.Done() // 取消是协作式的：必须自己盯着 ctx
	return ctx.Err()
}

// ─────────────────────────────────────────────────────────────────────────────
// 桩 3：ctx 桩 —— 立刻反映传入 ctx 的状态，用来测「整体取消不再重试」。
// ─────────────────────────────────────────────────────────────────────────────

type ctxExchange struct{ calls atomic.Int32 }

func (c *ctxExchange) Submit(ctx context.Context, _ *Order) error {
	c.calls.Add(1)
	return ctx.Err() // ctx 没取消时是 nil（成功），取消后是 Canceled
}

// 测试用的哨兵错误：断言时用 errors.Is 比对，而不是比字符串。
var errExchangeRefused = errors.New("交易所拒绝：限频")

// newOrder 造一笔订单，ID 由调用方给，保证幂等键可控。
// Qty 是 decimal：数量是「量」不是统计量，构造方式变了但断言逻辑不变。
func newOrder(id string) *Order {
	return &Order{ID: id, Symbol: "BTCUSDT", Side: "BUY", Qty: decimal.RequireFromString("0.002")}
}

// ─────────────────────────────────────────────────────────────────────────────
// 幂等：同一个订单号第二次提交必须被拒，且**不能再打一次交易所**。
//
// 「返回了 duplicate」和「交易所没被再调一次」是两件事，后者只有计数桩能证明。
// ─────────────────────────────────────────────────────────────────────────────

func TestSubmitIdempotent(t *testing.T) {
	exch := &countingExchange{} // failures=0：从不失败
	eng := NewEngine(exch, NewMemoryDeduper(), 2)

	first := newOrder("ORD-IDEM-1")
	if err := eng.Submit(context.Background(), first); err != nil {
		t.Fatalf("首次提交应成功，实际: %v", err)
	}
	if first.Status != "accepted" {
		t.Errorf("首次提交后 Status = %q, want %q", first.Status, "accepted")
	}
	// 状态机：提交成功 → created → submitted
	if first.State != order.Submitted {
		t.Errorf("首次提交后 State = %q, want %q", first.State, order.Submitted)
	}

	second := newOrder("ORD-IDEM-1") // 同一个订单号
	err := eng.Submit(context.Background(), second)
	if err == nil {
		t.Fatal("同一订单号第二次提交应报错，实际成功")
	}
	if second.Status != "duplicate" {
		t.Errorf("第二次提交后 Status = %q, want %q", second.Status, "duplicate")
	}
	// 被幂等拦下的订单**根本没发往交易所**，状态不该被推进
	if second.State != order.Created {
		t.Errorf("被幂等拦下的订单 State = %q, want %q（保持原状态）", second.State, order.Created)
	}
	if !strings.Contains(err.Error(), "ORD-IDEM-1") {
		t.Errorf("错误信息 = %q, 期望包含订单号 %q", err, "ORD-IDEM-1")
	}
	// 核心断言：幂等生效时，交易所只被真正调用过一次
	if got := exch.Calls(); got != 1 {
		t.Errorf("交易所被调用 %d 次, want 1（幂等：第二次不该再打交易所）", got)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 重试：桩前 N 次失败、第 N+1 次成功 → Attempts == N+1、Status == "accepted"。
//
// ⚠ 退避是 20/40/80ms 指数级，所以这里刻意只让失败次数 ≤ 2，避免测试变慢。
// ─────────────────────────────────────────────────────────────────────────────

func TestSubmitRetryThenSucceed(t *testing.T) {
	tests := []struct {
		name     string
		failures int // 前几次失败
		maxRetry int
	}{
		{"首次即成功", 0, 2},
		{"第 2 次成功（前 1 次失败）", 1, 2},
		{"第 3 次成功（前 2 次失败）", 2, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exch := &countingExchange{failures: tt.failures, err: errExchangeRefused}
			eng := NewEngine(exch, NewMemoryDeduper(), tt.maxRetry)

			o := newOrder("ORD-RETRY-" + tt.name)
			if err := eng.Submit(context.Background(), o); err != nil {
				t.Fatalf("最终应成功，实际: %v", err)
			}
			if o.Status != "accepted" {
				t.Errorf("Status = %q, want %q", o.Status, "accepted")
			}
			if want := tt.failures + 1; o.Attempts != want {
				t.Errorf("Attempts = %d, want %d", o.Attempts, want)
			}
			if want := tt.failures + 1; exch.Calls() != want {
				t.Errorf("交易所被调用 %d 次, want %d", exch.Calls(), want)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 重试耗尽：一直失败 → Status == "failed"，且返回的是**最后一次**的错误。
// ─────────────────────────────────────────────────────────────────────────────

func TestSubmitRetryExhausted(t *testing.T) {
	const maxRetry = 2
	exch := &countingExchange{failures: 1 << 30, err: errExchangeRefused} // 永远失败
	eng := NewEngine(exch, NewMemoryDeduper(), maxRetry)

	o := newOrder("ORD-EXHAUST")
	err := eng.Submit(context.Background(), o)
	if err == nil {
		t.Fatal("一直失败时应返回错误，实际成功")
	}
	if o.Status != "failed" {
		t.Errorf("Status = %q, want %q", o.Status, "failed")
	}
	// 状态机：重试耗尽 / 单次超时耗尽重试 → created → rejected
	if o.State != order.Rejected {
		t.Errorf("State = %q, want %q", o.State, order.Rejected)
	}
	// 用 errors.Is 而不是比字符串：重构换错误包装方式时依然成立
	if !errors.Is(err, errExchangeRefused) {
		t.Errorf("返回的错误 = %v, 期望能 errors.Is 到最后一次的错误 %v", err, errExchangeRefused)
	}
	if want := maxRetry + 1; o.Attempts != want {
		t.Errorf("Attempts = %d, want %d（maxRetry+1）", o.Attempts, want)
	}
	if want := maxRetry + 1; exch.Calls() != want {
		t.Errorf("交易所被调用 %d 次, want %d", exch.Calls(), want)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 单次调用超时：桩阻塞到 ctx.Done()，Engine 的 perCall（120ms）到点后必须返回超时错误。
//
// 用真实网络测超时是噩梦，用桩就是几行 + 120ms。
// maxRetry 传 0，避免连超时好几次把测试拖慢。
// ─────────────────────────────────────────────────────────────────────────────

func TestSubmitPerCallTimeout(t *testing.T) {
	exch := &blockingExchange{}
	eng := NewEngine(exch, NewMemoryDeduper(), 0) // 不重试

	start := time.Now()
	o := newOrder("ORD-TIMEOUT")
	err := eng.Submit(context.Background(), o)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("桩一直阻塞，期望超时错误，实际成功")
	}
	// 错误链里能看出是超时，而不是别的失败
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("错误 = %v, 期望 errors.Is(err, context.DeadlineExceeded) 为 true", err)
	}
	if o.Status != "failed" {
		t.Errorf("Status = %q, want %q", o.Status, "failed")
	}
	// 状态机：重试耗尽 / 单次超时耗尽重试 → created → rejected
	if o.State != order.Rejected {
		t.Errorf("State = %q, want %q", o.State, order.Rejected)
	}
	// 必须在 perCall（120ms）附近返回，不能真的等桩自己结束（它永远不会）
	if elapsed > 2*time.Second {
		t.Errorf("耗时 %v，太久 —— blockingExchange 可能没有正确响应 ctx.Done()", elapsed)
	}
	if got := exch.calls.Load(); got != 1 {
		t.Errorf("交易所被调用 %d 次, want 1", got)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 整体取消：传入一个**已经 cancel** 的 ctx，不该继续重试。
//
// 这条守的是 Engine 里那个 `select { case <-time.After(backoff): case <-ctx.Done(): }`：
// 少了 ctx.Done() 分支，取消就只能在「所有重试跑完」之后才生效。
// ─────────────────────────────────────────────────────────────────────────────

func TestSubmitCanceledContextStopsRetry(t *testing.T) {
	exch := &ctxExchange{}
	eng := NewEngine(exch, NewMemoryDeduper(), 5) // maxRetry 故意给大：真重试的话会调很多次

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 提交前就取消

	o := newOrder("ORD-CANCEL")
	err := eng.Submit(ctx, o)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("错误 = %v, 期望 errors.Is(err, context.Canceled) 为 true", err)
	}
	if o.Status != "canceled" {
		t.Errorf("Status = %q, want %q", o.Status, "canceled")
	}
	// 状态机：订单始终没被交易所受理，所以落到 rejected（created → rejected），
	// 而不是 canceled —— canceled 在跃迁表里只从 submitted/partial 出发（那才是「撤单」）。
	if o.State != order.Rejected {
		t.Errorf("State = %q, want %q", o.State, order.Rejected)
	}
	if got := exch.calls.Load(); got != 1 {
		t.Errorf("已取消的 ctx 不该继续重试：交易所被调用 %d 次, want 1", got)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// MemoryDeduper 单独测：同一个 key 第一次 true、第二次 false；不同 key 互不影响。
// ─────────────────────────────────────────────────────────────────────────────

func TestMemoryDeduper(t *testing.T) {
	d := NewMemoryDeduper()
	ctx := context.Background()

	first, err := d.MarkOnce(ctx, "ORD-A")
	if err != nil {
		t.Fatalf("首次 MarkOnce 报错: %v", err)
	}
	if !first {
		t.Error("同一个 key 第一次 MarkOnce 应返回 true")
	}

	second, err := d.MarkOnce(ctx, "ORD-A")
	if err != nil {
		t.Fatalf("第二次 MarkOnce 报错: %v", err)
	}
	if second {
		t.Error("同一个 key 第二次 MarkOnce 应返回 false")
	}

	other, err := d.MarkOnce(ctx, "ORD-B")
	if err != nil {
		t.Fatalf("另一个 key MarkOnce 报错: %v", err)
	}
	if !other {
		t.Error("不同 key 应各自独立，ORD-B 第一次应返回 true")
	}

	if d.Name() == "" {
		t.Error("Name() 不该为空（日志里要靠它说清幂等状态存在哪）")
	}
}

// NewEngine 传 nil Deduper 时应自动兜底成内存实现，而不是 panic。
func TestNewEngineNilDeduperFallsBack(t *testing.T) {
	eng := NewEngine(&countingExchange{}, nil, 1)
	if eng.DeduperName() == "" {
		t.Error("nil Deduper 应兜底为内存实现，DeduperName() 不该为空")
	}
	o := newOrder("ORD-NIL-DEDUP")
	if err := eng.Submit(context.Background(), o); err != nil {
		t.Fatalf("兜底后应能正常提交，实际: %v", err)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Mock（生产代码里的假交易所）的分支：失败率/慢响应/超时。
//
// 它是 main 用来「逼出重试与超时分支」的实现，行为本身也值得钉住。
// 全部用固定种子，结果可复现。
// ─────────────────────────────────────────────────────────────────────────────

func TestMockSubmitBranches(t *testing.T) {
	t.Run("失败率与慢响应都为 0：总是成功", func(t *testing.T) {
		m := NewMock(11, 0, 0, 10*time.Millisecond)
		if err := m.Submit(context.Background(), newOrder("ORD-MOCK-OK")); err != nil {
			t.Errorf("期望成功，实际: %v", err)
		}
	})

	t.Run("失败率为 1：总是被拒", func(t *testing.T) {
		m := NewMock(11, 1, 0, 10*time.Millisecond)
		if err := m.Submit(context.Background(), newOrder("ORD-MOCK-FAIL")); err == nil {
			t.Error("失败率为 1 时应返回错误，实际成功")
		}
	})

	t.Run("慢响应但超时窗口够长：最终成功", func(t *testing.T) {
		m := NewMock(11, 0, 1, 5*time.Millisecond) // SlowRate=1，模拟慢响应
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		if err := m.Submit(ctx, newOrder("ORD-MOCK-SLOW")); err != nil {
			t.Errorf("超时窗口够长时应成功，实际: %v", err)
		}
	})

	t.Run("慢响应撞上更短的 ctx 超时：报超时", func(t *testing.T) {
		m := NewMock(11, 0, 1, time.Hour) // 慢到不可能在 ctx 到期前返回
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()
		err := m.Submit(ctx, newOrder("ORD-MOCK-TIMEOUT"))
		if err == nil {
			t.Fatal("ctx 超时应报错，实际成功")
		}
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("错误 = %v, 期望能 errors.Is 到 context.DeadlineExceeded", err)
		}
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// ApplyFill：按「累计成交量」推进订单状态机。
//
// 每条用例都从「提交成功（submitted）」起步，断言的是**状态与成交量的可观察结果**，
// 而不是内部实现。half 用 decimal 除 2 得到 0.001，精确无误差。
// ─────────────────────────────────────────────────────────────────────────────

// submitOnce 造一个引擎、提交一笔订单，返回引擎与订单（提交失败直接 Fatal）。
func submitOnce(t *testing.T, id string) (*Engine, *Order) {
	t.Helper()
	eng := NewEngine(&countingExchange{}, NewMemoryDeduper(), 0)
	o := newOrder(id)
	if err := eng.Submit(context.Background(), o); err != nil {
		t.Fatalf("提交应成功，实际: %v", err)
	}
	if o.State != order.Submitted {
		t.Fatalf("提交成功后 State = %q, want %q", o.State, order.Submitted)
	}
	return eng, o
}

// 部分成交 → 补满：submitted → partial → filled。
func TestApplyFillPartialThenFilled(t *testing.T) {
	eng, o := submitOnce(t, "ORD-FILL-STEP")
	half := o.Qty.Div(decimal.NewFromInt(2)) // 0.002 / 2 = 0.001（decimal 精确）

	st, err := eng.ApplyFill(context.Background(), o.ID, half)
	if err != nil {
		t.Fatalf("部分成交应成功，实际: %v", err)
	}
	if st != order.Partial {
		t.Errorf("部分成交后 State = %q, want %q", st, order.Partial)
	}
	if !o.FilledQty.Equal(half) {
		t.Errorf("FilledQty = %s, want %s", o.FilledQty, half)
	}

	st, err = eng.ApplyFill(context.Background(), o.ID, o.Qty)
	if err != nil {
		t.Fatalf("补满应成功，实际: %v", err)
	}
	if st != order.Filled {
		t.Errorf("补满后 State = %q, want %q", st, order.Filled)
	}
	if !o.FilledQty.Equal(o.Qty) {
		t.Errorf("FilledQty = %s, want %s", o.FilledQty, o.Qty)
	}
}

// 一次全成：submitted → filled。
func TestApplyFillFullInOneGo(t *testing.T) {
	eng, o := submitOnce(t, "ORD-FILL-ALL")

	st, err := eng.ApplyFill(context.Background(), o.ID, o.Qty)
	if err != nil {
		t.Fatalf("全部成交应成功，实际: %v", err)
	}
	if st != order.Filled {
		t.Errorf("State = %q, want %q", st, order.Filled)
	}
}

// 零成交：还没成交，状态保持 submitted（不应被推到 partial）。
func TestApplyFillZeroKeepsSubmitted(t *testing.T) {
	eng, o := submitOnce(t, "ORD-FILL-ZERO")

	st, err := eng.ApplyFill(context.Background(), o.ID, decimal.Zero)
	if err != nil {
		t.Fatalf("零成交不该报错，实际: %v", err)
	}
	if st != order.Submitted {
		t.Errorf("零成交后 State = %q, want %q", st, order.Submitted)
	}
}

// 幂等：交易所会把同一条成交回调重复推送，重复应用同一个成交量必须是无变化、不报错。
func TestApplyFillIdempotent(t *testing.T) {
	eng, o := submitOnce(t, "ORD-FILL-IDEM")
	ctx := context.Background()

	first, err := eng.ApplyFill(ctx, o.ID, o.Qty)
	if err != nil {
		t.Fatalf("首次全成应成功，实际: %v", err)
	}
	again, err := eng.ApplyFill(ctx, o.ID, o.Qty)
	if err != nil {
		t.Fatalf("重复回调不该报错，实际: %v", err)
	}
	if first != order.Filled || again != order.Filled {
		t.Errorf("两次回调状态 = %q / %q, want 都是 %q", first, again, order.Filled)
	}
}

// 不变量违反：成交量超过下单数量 → 报错，且状态与成交量都不许被改动。
func TestApplyFillOverFillRejected(t *testing.T) {
	eng, o := submitOnce(t, "ORD-FILL-OVER")
	over := o.Qty.Add(decimal.RequireFromString("0.001"))

	st, err := eng.ApplyFill(context.Background(), o.ID, over)
	if !errors.Is(err, order.ErrFillInvariant) {
		t.Fatalf("超额成交应报 ErrFillInvariant，实际 err=%v", err)
	}
	if st != order.Submitted {
		t.Errorf("报错后 State = %q, want 保持 %q", st, order.Submitted)
	}
	if !o.FilledQty.IsZero() {
		t.Errorf("报错后 FilledQty = %s, want 0（不能把非法值写进去）", o.FilledQty)
	}
}

// 不变量违反：负成交量同样要报错。
func TestApplyFillNegativeRejected(t *testing.T) {
	eng, o := submitOnce(t, "ORD-FILL-NEG")

	if _, err := eng.ApplyFill(context.Background(), o.ID, decimal.RequireFromString("-0.001")); !errors.Is(err, order.ErrFillInvariant) {
		t.Errorf("负成交量应报 ErrFillInvariant，实际 err=%v", err)
	}
}

// 终态出不去：已经 filled 的订单不能再退回 partial。
func TestApplyFillAfterTerminalRejected(t *testing.T) {
	eng, o := submitOnce(t, "ORD-FILL-TERM")
	ctx := context.Background()

	if _, err := eng.ApplyFill(ctx, o.ID, o.Qty); err != nil {
		t.Fatalf("全部成交应成功，实际: %v", err)
	}
	st, err := eng.ApplyFill(ctx, o.ID, o.Qty.Div(decimal.NewFromInt(2)))
	if !errors.Is(err, order.ErrIllegalTransition) {
		t.Errorf("filled 之后不能再回到 partial，错误 = %v", err)
	}
	if st != order.Filled {
		t.Errorf("报错后 State = %q, want 保持 %q", st, order.Filled)
	}
}

// 未登记的订单：提交失败的订单没被交易所受理，不该能更新成交。
func TestApplyFillUnknownOrderRejected(t *testing.T) {
	ctx := context.Background()

	eng := NewEngine(&countingExchange{failures: 1 << 30, err: errExchangeRefused}, NewMemoryDeduper(), 0)
	failed := newOrder("ORD-FILL-NEVER")
	if err := eng.Submit(ctx, failed); err == nil {
		t.Fatal("这个桩永远失败，提交不该成功")
	}
	if _, err := eng.ApplyFill(ctx, failed.ID, failed.Qty); err == nil {
		t.Error("提交失败的订单不在已受理列表里，ApplyFill 应报错")
	}

	// 完全陌生的订单号同样报错，而不是静默创建一个状态
	if _, err := eng.ApplyFill(ctx, "ORD-FILL-NOBODY", failed.Qty); err == nil {
		t.Error("陌生订单号应报错")
	}
}
