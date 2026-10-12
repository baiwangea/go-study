// 本文件是 T-07 的测试代码：手写测试替身（test double）。
//
// 全篇没有引入任何 mock 框架 —— 三个假实现加起来不到 60 行。
// 先读 trader.go 顶部的关卡卡，再回来看代码。

package trader

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// 桩 1：固定价格（stub）
//
// 只有 6 行。这就是 Go 不需要 mock 框架的原因 ——
// 接口只有一个方法，手写比学框架 API 还快。
//
// 对照 Java 的 Mockito：
//     when(feed.price("BTCUSDT")).thenReturn(60000.0);
// 一行搞定，但代价是：你要引入框架、要处理 Mockito 和 JDK 版本的兼容、
// 静态 import、以及「为什么这个 mock 没生效」的调试。
// 而 Go 这几行是**普通代码**，编译期就检查、断点能进、没有魔法。
// ─────────────────────────────────────────────────────────────────────────────

type stubFeed struct {
	price float64
	err   error
}

func (s stubFeed) Price(ctx context.Context, symbol string) (float64, error) {
	return s.price, s.err
}

// ─────────────────────────────────────────────────────────────────────────────
// 桩 2：阻塞到 context 取消（用来测超时）
//
// ⚠ 关键点：**它必须 select ctx.Done()**。
//
// context 的取消是「协作式」的 —— 它只是关掉一个 channel，
// 不会强杀任何 goroutine。如果你的实现不主动检查 ctx.Done()，
// 取消就永远不会生效，测试会一直等到真实超时（或者永远挂住）。
//
// 这是 Go 并发里最容易被忽略的一点，也是 L2-11 的核心。
// 现实中的坑：HTTP 客户端要设 Timeout、数据库查询要传 ctx、
// 循环里每轮都要看一眼 ctx.Err()。少写一处，整个链路就取消不掉。
// ─────────────────────────────────────────────────────────────────────────────

type slowFeed struct {
	delay time.Duration
}

func (s slowFeed) Price(ctx context.Context, symbol string) (float64, error) {
	select {
	case <-time.After(s.delay):
		return 60000, nil // 慢，但最终会返回
	case <-ctx.Done():
		return 0, ctx.Err() // ← 被取消时立刻返回，把 ctx 的错误原样带出去
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 桩 3：spy —— 既返回结果，又记录「被怎么调用了」
//
// 这是纯返回值断言做不到的部分。很多 bug 的表现不是「返回错了」，
// 而是「**不该被调用的东西被调用了**」——
// 比如风控该拦住却没拦、价格不合适却下了单。
//
// 用 sync.Mutex 保护记录，因为 T-02 学过的 t.Parallel 可能让多个测试共享它。
// ─────────────────────────────────────────────────────────────────────────────

type spySink struct {
	mu    sync.Mutex
	calls []placeCall
	err   error
}

type placeCall struct {
	symbol string
	qty    float64
	price  float64
}

func (s *spySink) Place(ctx context.Context, symbol string, qty, price float64) (string, error) {
	s.mu.Lock()
	s.calls = append(s.calls, placeCall{symbol, qty, price})
	s.mu.Unlock()

	if s.err != nil {
		return "", s.err
	}
	return "ord-0001", nil
}

// CallCount 返回被调用次数。
func (s *spySink) CallCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.calls)
}

// LastCall 返回最后一次调用的参数。
func (s *spySink) LastCall() (placeCall, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.calls) == 0 {
		return placeCall{}, false
	}
	return s.calls[len(s.calls)-1], true
}

// ─────────────────────────────────────────────────────────────────────────────
// 用桩跑测试。
//
// 注意每条用例都在**同一个测试函数里**用表驱动 + 子测试（复习 T-01、T-02），
// 但每条用例构造自己的桩 —— 桩是「每条用例的私有财产」，
// 不像 Java 那样要 @BeforeEach 重建、或者在 @AfterEach 里 reset()。
// ─────────────────────────────────────────────────────────────────────────────

func TestRunOnce(t *testing.T) {
	// ⚠ 为什么 refPrice 和 threshold 挑这两个值？
	//
	// 因为边界用例踩过浮点数的坑。第一版我写的是 refPrice=60000、threshold=0.01，
	// 想测「价格正好等于 60000×(1-0.01)」这个边界。但：
	//     0.01 在二进制里不能精确表示 → 1-0.01 = 0.9899999999999999911...
	//     60000 × 那个数 = 59399.9999999999946... 而不是 59400
	// 于是「59400 是否大于阈值」这个问题，答案取决于编译器怎么折叠常量 ——
	// 同一段代码换个 Go 版本可能就翻转了。
	//
	// 现在用的 0.25 和 100 都是二进制可精确表示的数：
	//     100 × (1-0.25) = 75 精确成立
	// **写边界测试时，先确认你的边界值是精确可表示的。** 否则你测的是一个假边界。
	const (
		symbol    = "BTCUSDT"
		refPrice  = 100.0
		threshold = 0.25 // 便宜 25% 才买
		qty       = 0.1
	)

	tests := []struct {
		name          string
		price         float64
		feedErr       error
		sinkErr       error
		wantBought    bool
		wantSinkCalls int // ← 和 wantBought 是两回事，见「下单接口报错」那条
		wantErrText   string
	}{
		{
			name:          "便宜 26%，买入",
			price:         74,
			wantBought:    true,
			wantSinkCalls: 1,
		},
		{
			name:          "正好便宜 25%（边界），买入",
			price:         75, // 100 × 0.75 = 75，精确
			wantBought:    true,
			wantSinkCalls: 1,
		},
		{
			name:          "只便宜 24%，不买",
			price:         76,
			wantBought:    false,
			wantSinkCalls: 0,
		},
		{
			name:          "价格等于参考价，不买",
			price:         100,
			wantBought:    false,
			wantSinkCalls: 0,
		},
		{
			name:          "行情接口报错，不下单",
			feedErr:       errors.New("ws 连接断开"),
			wantBought:    false,
			wantSinkCalls: 0,
			wantErrText:   "行情失败",
		},
		{
			// ⚠ 这一条是本关最容易搞错的地方：
			//   下单接口**被调用了**（1 次），但它返回了错误，
			//   所以最终 orderID 是空的、wantBought 是 false。
			//   如果只按 wantBought 去推「sink 该被调用几次」，就会写出错误的断言 ——
			//   我第一版就是这么写错的，被这个测试自己抓了出来。
			//   这正是 spy 的价值：**「调用了但失败了」和「压根没调用」是两件事。**
			name:          "下单接口报错：被调用了，但没成功",
			price:         74,
			sinkErr:       errors.New("余额不足"),
			wantBought:    false,
			wantSinkCalls: 1,
			wantErrText:   "下单",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 每条用例一套全新的桩 —— 零污染、零 reset
			feed := stubFeed{price: tt.price, err: tt.feedErr}
			sink := &spySink{err: tt.sinkErr}
			tr := NewTrader(feed, sink, threshold)

			orderID, err := tr.RunOnce(context.Background(), symbol, refPrice, qty)

			// ── 断言 1：错误 ──────────────────────────────────────────
			if tt.wantErrText != "" {
				if err == nil {
					t.Fatalf("期望报错，实际成功（orderID=%q）", orderID)
				}
				if !strings.Contains(err.Error(), tt.wantErrText) {
					t.Errorf("错误信息 = %q, 期望包含 %q", err, tt.wantErrText)
				}
			} else if err != nil {
				t.Fatalf("意外报错: %v", err)
			}

			// ── 断言 2：返回值 ────────────────────────────────────────
			if gotBought := orderID != ""; gotBought != tt.wantBought {
				t.Errorf("是否买入 = %t, want %t（orderID=%q）", gotBought, tt.wantBought, orderID)
			}

			// ── 断言 3：spy —— 「不该被调用」比「返回什么」更重要 ─────
			if got := sink.CallCount(); got != tt.wantSinkCalls {
				t.Errorf("下单接口被调用 %d 次, want %d", got, tt.wantSinkCalls)
			}

			// ── 断言 4：spy 记录的参数 ────────────────────────────────
			// 如果只看 RunOnce 的返回值，你永远不知道「下单用的是什么价格」。
			if tt.wantSinkCalls == 0 {
				return
			}
			call, ok := sink.LastCall()
			if !ok {
				t.Fatal("期望有一次下单调用")
			}
			if call.symbol != symbol {
				t.Errorf("下单 symbol = %q, want %q", call.symbol, symbol)
			}
			if call.qty != qty {
				t.Errorf("下单 qty = %v, want %v", call.qty, qty)
			}
			// 下单价格必须是行情价，不能是参考价 —— 这个 bug 很常见
			if call.price != tt.price {
				t.Errorf("下单 price = %v, want %v（应该用行情价而不是参考价）", call.price, tt.price)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 超时分支：只有桩能做到的测试。
//
// 用真实网络测超时是噩梦：你得搭一个「故意不回包」的服务，
// 还要等真实的秒级超时，CI 上必然 flaky。
// 用桩就是 10 行代码 + 50 毫秒。
//
// 这也解释了为什么「依赖注入」是写出好测试的前提 ——
// 没有接口，你连一个能阻塞的假实现都塞不进去。
// ─────────────────────────────────────────────────────────────────────────────

func TestContextTimeout(t *testing.T) {
	feed := slowFeed{delay: 1 * time.Second} // 行情要 1 秒才回来
	sink := &spySink{}
	tr := NewTrader(feed, sink, 0.01)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := tr.RunOnce(ctx, "BTCUSDT", 60000, 0.1)
	elapsed := time.Since(start)

	// ── 断言 1：必须报错 ─────────────────────────────────────────────
	if err == nil {
		t.Fatal("期望超时错误，实际成功了")
	}

	// ── 断言 2：错误必须是 context.DeadlineExceeded ──────────────────
	// 用 errors.Is 而不是比较错误字符串（复习 L1-19 错误链）。
	// RunOnce 用 %w 包装了错误，所以 Is 能穿透两层找到根因。
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("错误 = %v, 期望 errors.Is(err, context.DeadlineExceeded) 为 true", err)
	}

	// ── 断言 3：必须在 50ms 左右返回，不能等满 1 秒 ──────────────────
	// 如果 slowFeed 忘了 select ctx.Done()，这里会等到 1 秒，
	// 断言会失败 —— 这条断言就是在守护「取消是协作式的」这个认知。
	if elapsed > 500*time.Millisecond {
		t.Errorf("耗时 %v，太久了 —— slowFeed 可能没有正确响应 ctx.Done()", elapsed)
	}

	// ── 断言 4：超时时绝不能下单 ─────────────────────────────────────
	if sink.CallCount() != 0 {
		t.Errorf("超时情况下下单接口被调用了 %d 次，应该一次都不调", sink.CallCount())
	}
	t.Logf("超时用例耗时 %v，错误链正确，且未下单", elapsed)
}

// ─────────────────────────────────────────────────────────────────────────────
// 顺带演示一个常见误区：**不要为了「以后可能替换」而提前抽接口**。
//
// 下面这个接口只被一个实现满足、也没有第二个实现的计划 ——
// 它带来的只有间接层，没有可测试性收益（因为它本来就没有外部依赖）。
// 这种接口在生产仓库里很常见，通常叫 IOrderService / AbstractFactory 之类，
// 是从 Java 带过来的习惯。
//
// 判断标准很简单：**如果你说不出「第二个实现是什么」，就不该有接口。**
// ─────────────────────────────────────────────────────────────────────────────

// Calculator 是反面示例：纯计算，没有任何外部依赖，不需要接口。
//
// 它存在的唯一目的是让你看到「多余抽象」长什么样。
type Calculator struct{}

// Fee 计算手续费。直接调用即可 —— 测试它不需要任何桩。
func (Calculator) Fee(notional, rateBps float64) float64 {
	return notional * rateBps / 10000
}

// ICalculator 是**多余抽象**的示例：没有第二个实现，也没有注入需求。
//
// 如果哪天 Calculator 真的需要读配置文件、或者调用外部服务，
// 那时再抽接口也完全来得及 —— 而且那时候你才知道接口该有几个方法。
type ICalculator interface {
	Fee(notional, rateBps float64) float64
}

func TestNoUnnecessaryInterface(t *testing.T) {
	// 直接测具体类型就行，不需要任何替身
	c := Calculator{}
	if got, want := c.Fee(1000, 5), 0.5; got != want {
		t.Errorf("Fee(1000, 5) = %v, want %v", got, want)
	}

	// 注意：ICalculator 这个接口目前没有任何用处。
	// 保留它只是作为「反面教材」——真实项目里这种接口应该删掉。
	var _ ICalculator = c
}
