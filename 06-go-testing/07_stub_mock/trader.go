/*
═══════════════════════════════════════════════════════════════════════
T-07：接口打桩与依赖注入（Go 不需要 mock 框架）
═══════════════════════════════════════════════════════════════════════

【核心概念】

	被测代码如果直接 new 出一个真实的 HTTP 客户端、连真实的交易所，
	测试就没法写：慢、要网络、结果不可控、还可能真下单。

	解法是**依赖注入**：把「外部世界」抽象成接口，让调用方传进来。
	测试时传一个假实现（stub），生产时传真实现。

	    // 生产代码依赖接口，不依赖具体类型
	    type PriceFeed interface {
	        Price(ctx context.Context, symbol string) (float64, error)
	    }

	    type Trader struct {
	        feed PriceFeed   // ← 接口，不是 *BitgetClient
	    }

	为什么 Go 社区很少用 Mockito / gomock 这类框架？

	    Java  : 需要一个框架在运行时用字节码技术生成子类（Mockito），
	            或者编译期代码生成（gomock）。因为 Java 的接口方法多、
	            构造依赖重，手写假实现太累。
	    PHP   : PHPUnit 内建 createMock()，同样靠反射生成。
	    Go    : **手写假实现通常只要 5~15 行**，因为 Go 的接口天生小。
	            一个接口 1~3 个方法，手写一个 struct 就完事了 ——
	            比学一套框架的 API、调试代码生成的问题更省事。

	这不是「Go 没有工具」，而是**接口设计得好就不需要工具**。
	Go 社区的口诀是：「接口应该由调用方定义，而且越小越好。」
	io.Reader 只有一个方法，所以你手写一个假 Reader 只要 3 行。

	⚠ 三条纪律（本关的重点）：
	  ① **接口定义在调用方**，不要为了「以后可能替换」而提前抽象。
	     定义接口的理由永远是「我需要在这个位置注入假实现来做测试」。
	  ② **假实现要能被断言**。只返回固定值的叫 stub；
	     还记录「被调用了几次、参数是什么」的叫 spy。两者常合并在一个 struct 里。
	  ③ **测错误分支和超时分支**，不要只测 happy path。
	     生产事故几乎全在错误路径上 —— 而错误路径恰恰是最难手工复现的。

【运行方式】

	cd 06-go-testing && go test -v ./07_stub_mock/

	    go test -v -run 'TestRunOnce/价格太高' ./07_stub_mock/   # 只跑一条
	    go test -race ./07_stub_mock/                            # 桩也会并发调用

【你会看到】

	✓ 三种桩：固定值（stub）、返回错误、阻塞到 context 超时
	✓ spy 断言「价格不合适时，下单接口一次都没被调用」——
	  这是纯返回值断言做不到的
	✓ 超时用例完全可控：不依赖真实网络，毫秒级结束，CI 上不会 flaky

【思考题】（动手改代码验证，别空想）

 1. 把 Trader.feed 的类型从 PriceFeed 接口改成 *BitgetFeed 具体类型，
    再试着写现在的测试。你会遇到什么问题？
    这就是「依赖具体类型」的代价 —— 也是依赖注入存在的唯一理由。

 2. TestRunOnce 里的 spySink 记录了调用参数。想清楚：
    如果要断言「下单价格等于行情价」，只看返回值够不够？
    （提示：返回值是订单 ID，价格根本没出现在返回值里。）

 3. 看 TestContextTimeout 里的 slowFeed：它为什么要 select ctx.Done()？
    如果写成 time.Sleep(1*time.Second) 然后返回价格，测试会怎样失败？
    （提示：context 取消是**协作式**的 —— 没人检查 ctx.Done()，
    取消就永远不会生效。这是 Go 并发里最容易被忽略的一点。）

 4. 假设要给 Trader 加一个新依赖「风控检查器」。
    你会把它做成接口参数、还是塞进 Trader 的字段？
    如果想不清，先看 trader.go 里 NewTrader 的签名设计。

【通关标准】

	能说清「为什么 Go 手写桩比用 mock 框架更常见」，
	能写出「记录调用参数的 spy」并断言「某方法不该被调用」，
	并且知道「测错误分支和超时分支，不能只测 happy path」。

【前置】L1-08（接口）、L2-11（context 取消）   【标签】测试替身 · 依赖注入 · spy · 超时
═══════════════════════════════════════════════════════════════════════
*/
package trader

import (
	"context"
	"fmt"
)

// ─────────────────────────────────────────────────────────────────────────────
// 生产代码：一个「看到便宜就买」的极简交易器。
//
// 注意这里的接口是**定义在调用方（这个包）**的，不是定义在实现方。
// 这是 Go 的一条重要惯例，和 Java 的「先定义 IUserService 再写实现」相反。
// 理由：接口应该由「需要它的代码」来定义，这样接口天然贴合调用方的需要，
// 不会长成一个大而全的怪物。
// ─────────────────────────────────────────────────────────────────────────────

// PriceFeed 提供行情价格。只有 1 个方法 —— 刻意保持最小。
//
// 小接口的价值：任何东西只要「能给出价格」就能满足它。
// 真实实现走 WebSocket，测试实现返回固定值，回测实现读历史数据 —— 全都能接。
type PriceFeed interface {
	Price(ctx context.Context, symbol string) (float64, error)
}

// OrderSink 接收下单请求。
//
// 名字叫 Sink（水槽）而不是 Client：它只负责「把订单送出去」，
// 不负责重试、不负责状态跟踪 —— 那些是别的组件的事。
// 一个接口只做一件事，是它能被轻松替换的前提。
type OrderSink interface {
	Place(ctx context.Context, symbol string, qty, price float64) (orderID string, err error)
}

// Trader 是策略本体。
//
// 它只依赖两个接口，完全不知道外面是 Bitget 还是 mock。
// 这就是「可测试性」的来源 —— 不需要任何框架，只需要接口。
type Trader struct {
	feed      PriceFeed
	sink      OrderSink
	threshold float64 // 低于参考价多少比例才买，如 0.01 表示便宜 1% 才买
}

// NewTrader 构造 Trader。
//
// 依赖从参数传入（依赖注入），而不是在内部 new 出来。
// 这一个设计决定，就决定了这个包能不能被测试。
func NewTrader(feed PriceFeed, sink OrderSink, threshold float64) *Trader {
	return &Trader{feed: feed, sink: sink, threshold: threshold}
}

// RunOnce 跑一轮：取价 → 判断 → 可能下单。
//
// 返回的 orderID 为空表示「没买」。
//
// ⚠ 注意错误处理：**行情取不到时直接返回错误，绝不下单。**
//
//	这是交易系统的铁律 —— 拿不到价格就下单，等于盲赌。
func (t *Trader) RunOnce(ctx context.Context, symbol string, refPrice, qty float64) (string, error) {
	price, err := t.feed.Price(ctx, symbol)
	if err != nil {
		return "", fmt.Errorf("取 %s 行情失败，本轮跳过: %w", symbol, err)
	}

	// 只有「市价低于参考价 × (1 - threshold)」时才买
	if price > refPrice*(1-t.threshold) {
		return "", nil // 不够便宜，不买 —— 注意这不是错误
	}

	orderID, err := t.sink.Place(ctx, symbol, qty, price)
	if err != nil {
		return "", fmt.Errorf("下单 %s 失败: %w", symbol, err)
	}
	return orderID, nil
}
