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
//	cd 05-trading-bot && go run .            # 跑一轮完整链路
//	go run -race .                            # 顺带验证管线没有竞态
package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strconv"
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
	const ticks = 12
	qty := 0.002

	rnd := rand.New(rand.NewSource(7)) // 固定种子，输出可复现
	feed := market.RandomWalk{Symbol: "BTCUSDT", Start: 63000, Step: 8 * time.Millisecond, Rnd: rnd}
	strategies := []strategy.Strategy{strategy.NewMACross(5)}
	guard := risk.New(envFloat("BOT_MAX_NOTIONAL", 500), 25*time.Millisecond, 0.02)
	// 注意前面的 &：Mock.Submit 用的是指针接收者，只有 *Mock 才满足 Exchange 接口（复盘 L1-13）
	engine := executor.NewEngine(&executor.Mock{
		Rnd: rnd, FailRate: 0.35, SlowRate: 0.25, TimeoutDur: 300 * time.Millisecond,
	}, 2)
	var sink notify.Notifier = notify.Telegram{Token: os.Getenv("TG_BOT_TOKEN")}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	before := runtime.NumGoroutine()
	quotes := feed.Subscribe(ctx, ticks)

	var (
		gotQuotes int
		signals   int
		accepted  int
		denied    = map[string]int{}
	)
	fmt.Printf("开始跑管线：%d 个行情 tick，%d 个策略，风控上限 %.0f USD\n\n", ticks, len(strategies), guard.MaxNotional)

	for q := range quotes {
		gotQuotes++
		for _, s := range strategies {
			sig, ok := s.OnQuote(q)
			if !ok {
				continue
			}
			signals++
			fmt.Printf("  行情 #%02d %-9s 价 %.2f → %s 发出 %s 信号（强度 %.4f）\n",
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

	// 故意重复提交一个已处理过的订单，验证幂等（生产环境换成 Redis SetNX / 唯一索引）
	dup := &executor.Order{ID: "ORD-BUY-01", Symbol: "BTCUSDT", Side: "BUY", Qty: qty}
	fmt.Printf("\n重复提交同一个订单号：")
	if err := engine.Submit(ctx, dup); err != nil {
		fmt.Println(" →", err)
	}

	cancel() // 停掉行情协程（生产环境是收到退出信号后由 main 统一取消）
	time.Sleep(20 * time.Millisecond)

	fmt.Println("\n汇总：")
	fmt.Printf("  收到行情 %d 笔｜产生信号 %d 个｜下单成功 %d 笔｜引擎去重后唯一订单 %d 个\n",
		gotQuotes, signals, accepted, engine.Processed())
	for reason, n := range denied {
		fmt.Printf("  风控拦截 %d 次：%s\n", n, reason)
	}
	fmt.Printf("  协程数：启动前 %d → 结束后 %d（cancel 生效，行情协程已退出）\n",
		before, runtime.NumGoroutine())
}
