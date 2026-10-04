// ── L4 阶段 · 交易机器人骨架（端到端管线）─────────────────────
//
// 前置说明：L2 并发（channel / context 取消 / worker pool）+ L3 GoFrame 主线。
// 本模块目标：把「行情 → 信号 → 风控 → 下单 → 通知」串成一条能跑、能测、能停的链路。
//
// 分层与接口（每层只依赖下一层，替换实现不影响上游）：
//
//	market.Feed   ──▶ strategy.Strategy ──▶ risk.Guard ──▶ executor.Engine ──▶ notify.Notifier
//	  行情源            信号                 守门            提交/重试/幂等         通知
//
// 当前接的是 mock：本地随机游走行情 + 会随机失败/慢响应的假交易所，
// 因此不需要任何网络与密钥即可跑通；换成 CEX/链上时只替换接口实现。
//
// 运行：
//
//	cd 05-trading-bot && go run .                      # mock 行情，零依赖
//	go run . -feed=bitget -ticks=30 -interval=400ms -threshold=0.00000001   # 真实行情（Bitget REST）
//	go run -race .                                     # 验证管线没有竞态
//
// 行情源只是 Feed 接口的一个实现，换源不需要改下游一行代码。
package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"

	"go-study/trading-bot/internal/executor"
	"go-study/trading-bot/internal/market"
	"go-study/trading-bot/internal/notify"
	"go-study/trading-bot/internal/risk"
	"go-study/trading-bot/internal/strategy"
)

func envFloat(key string, def float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}

func main() {
	var (
		feedName  = flag.String("feed", "mock", "行情源：mock（本地随机游走）或 bitget（真实 REST tickers）")
		symbol    = flag.String("symbol", "BTCUSDT", "交易对")
		ticks     = flag.Int("ticks", 16, "跑多少个 tick")
		interval  = flag.Duration("interval", 8*time.Millisecond, "轮询间隔；真实接口建议 ≥300ms 以免撞限频")
		threshold = flag.Float64("threshold", 0.0004, "MA 交叉触发阈值（真实行情波动小，建议 0.00005）")
	)
	flag.Parse()

	qty := 0.002
	// 失败计数会在行情协程里被写、在 main 里被读，用 atomic 避免 data race（L2-13）
	var failedTicks atomic.Int64

	var feed market.Feed
	switch *feedName {
	case "bitget":
		feed = &market.Bitget{
			Symbol: *symbol, Every: *interval, Timeout: 800 * time.Millisecond,
			OnError: func(err error) {
				failedTicks.Add(1)
				fmt.Println("     ⚠️ 行情拉取失败（丢掉这个 tick，管线继续跑）：", err)
			},
		}
	default:
		feed = market.RandomWalk{
			Symbol: *symbol, Start: 63000, Step: *interval, Vol: 0.006,
			Rnd: rand.New(rand.NewSource(7)), // 固定种子，输出可复现
		}
	}

	strategies := []strategy.Strategy{strategy.NewMACross(5, *threshold)}
	guard := risk.New(envFloat("BOT_MAX_NOTIONAL", 500), 25*time.Millisecond, 0.02)
	// 随机源必须每个组件各自持有：之前把同一个 *rand.Rand 同时交给行情协程与下单逻辑，
	// 被 go run -race 抓出 8 处 DATA RACE（math/rand.Rand 内部无锁）。现在各自 NewMock(seed)。
	engine := executor.NewEngine(executor.NewMock(11, 0.35, 0.25, 300*time.Millisecond), 2)
	var sink notify.Notifier = notify.Telegram{Token: os.Getenv("TG_BOT_TOKEN")}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	before := runtime.NumGoroutine()
	quotes := feed.Subscribe(ctx, *ticks)

	var (
		gotQuotes int
		signals   int
		accepted  int
		denied    = map[string]int{}
	)
	fmt.Printf("开始跑管线：行情源=%s 交易对=%s，目标 %d 个 tick，间隔 %v，阈值 %g，风控上限 %.0f USD\n\n",
		*feedName, *symbol, *ticks, *interval, *threshold, guard.MaxNotional)

	for q := range quotes {
		gotQuotes++
		for _, s := range strategies {
			sig, ok := s.OnQuote(q)
			if !ok {
				continue
			}
			signals++
			fmt.Printf("  行情 #%02d %-9s 价 %.2f → %s 发出 %s 信号（强度 %g）\n",
				gotQuotes, q.Symbol, q.Price, s.Name(), sig.Side, sig.Strength)

			if err := guard.Check(sig, qty, q.Price, time.Now()); err != nil {
				denied[err.Error()]++
				fmt.Println("     ⛔ 风控拦截：", err)
				continue
			}

			order := &executor.Order{
				ID: fmt.Sprintf("ORD-%s-%d", sig.Side, gotQuotes), Symbol: sig.Symbol, Side: sig.Side, Qty: qty,
			}
			err := engine.Submit(ctx, order)
			if err != nil {
				fmt.Printf("     ✗ 下单失败：%v（尝试 %d 次，状态 %s）\n", err, order.Attempts, order.Status)
				continue
			}
			accepted++
			fmt.Printf("     ✓ 下单成功：%s %s %g（尝试 %d 次）\n", order.ID, order.Symbol, order.Qty, order.Attempts)
			sink.Notify(fmt.Sprintf("%s %s %s 已受理", order.Symbol, order.Side, order.ID))
		}
	}

	// 幂等演示：用零失败的独立交易所，确保结论稳定可复现（生产环境换成 Redis SetNX / 唯一索引）
	idem := executor.NewEngine(executor.NewMock(1, 0, 0, 10*time.Millisecond), 2)
	first := &executor.Order{ID: "ORD-IDEM", Symbol: "BTCUSDT", Side: "BUY", Qty: qty}
	second := &executor.Order{ID: "ORD-IDEM", Symbol: "BTCUSDT", Side: "BUY", Qty: qty}
	err1 := idem.Submit(ctx, first)
	err2 := idem.Submit(ctx, second)
	fmt.Printf("\n幂等演示：首次 err=%v（状态 %s）；同一订单号再提 err=%v（状态 %s）\n", err1, first.Status, err2, second.Status)

	cancel() // 停掉行情协程（生产环境是收到退出信号后由 main 统一取消）
	// 行为接口断言：只有带这个方法的行情源需要释放 HTTP 空闲连接（L1-10 类型断言的实战用法）
	if idle, ok := feed.(interface{ CloseIdleConnections() }); ok {
		idle.CloseIdleConnections()
	}
	time.Sleep(50 * time.Millisecond)

	fmt.Println("\n汇总：")
	fmt.Printf("  收到行情 %d 笔｜行情拉取失败 %d 次｜产生信号 %d 个｜下单成功 %d 笔｜提交过的唯一订单号 %d 个（含最终失败的）\n",
		gotQuotes, failedTicks.Load(), signals, accepted, engine.Processed())
	for reason, n := range denied {
		fmt.Printf("  风控拦截 %d 次：%s\n", n, reason)
	}
	after := runtime.NumGoroutine()
	note := "cancel 生效，行情协程已退出"
	if after > before {
		note = "仍有残留协程，多为 HTTP 空闲连接或未返的 in-flight 请求，需继续查"
	}
	fmt.Printf("  协程数：启动前 %d → 结束后 %d（%s）\n", before, after, note)
	if signals == 0 {
		fmt.Println("  ⚠ 本轮没出任何信号：阈值相对真实波动太高，等于什么都没验证。参考下面的实测峰值调小 -threshold")
	}
	for _, s := range strategies {
		if p, ok := s.(interface{ Peak() float64 }); ok { // 行为接口断言：能报峰值的策略才打印
			fmt.Printf("  校准参考：%s 实测最大均线偏离 = %g（当前阈值 %g）\n", s.Name(), p.Peak(), *threshold)
		}
	}
}
