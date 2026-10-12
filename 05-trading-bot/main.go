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
//	go run . -mode=backtest                            # 回测：零网络零密钥也能评估策略
//	go run . -mode=backtest -pattern=v -bars=800       # 指定形态与 K 线数
//	go run . -mode=backtest -csv=/tmp/btc.csv          # 用外部 CSV（列 timestamp,price）
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
	"sync/atomic"
	"time"

	"github.com/shopspring/decimal"

	"go-study/trading-bot/internal/backtest"
	"go-study/trading-bot/internal/executor"
	"go-study/trading-bot/internal/market"
	"go-study/trading-bot/internal/notify"
	"go-study/trading-bot/internal/order"
	"go-study/trading-bot/internal/risk"
	"go-study/trading-bot/internal/store"
	"go-study/trading-bot/internal/strategy"
)

// envDecimal 读**金额类**环境变量（如 BOT_MAX_NOTIONAL）。金额用 decimal 而不是 float64：
// 解析失败退回默认值而不是 panic —— 配置写错不该让机器人起不来。
func envDecimal(key, def string) decimal.Decimal {
	if v := os.Getenv(key); v != "" {
		if d, err := decimal.NewFromString(v); err == nil {
			return d
		}
	}
	return decimal.RequireFromString(def) // def 是代码里的常量，写错就是程序员的问题，直接 panic
}

func main() {
	var (
		feedName  = flag.String("feed", "mock", "行情源：mock｜bitget（REST 轮询）｜bitget-ws（WebSocket 推送）")
		symbol    = flag.String("symbol", "BTCUSDT", "交易对")
		ticks     = flag.Int("ticks", 16, "跑多少个 tick")
		interval  = flag.Duration("interval", 8*time.Millisecond, "轮询间隔；真实接口建议 ≥300ms 以免撞限频")
		threshold = flag.Float64("threshold", 0.0004, "MA 交叉触发阈值（真实行情波动小，建议 0.00005）")

		mode       = flag.String("mode", "live", "运行模式：live（接真实/模拟行情）｜backtest（用确定的价格序列回测）")
		csvPath    = flag.String("csv", "", "回测模式：外部行情 CSV 路径（列 timestamp,price）；填了就用它，忽略 -pattern")
		pattern    = flag.String("pattern", "range", "回测模式：内置形态 up｜down｜range｜v")
		bars       = flag.Int("bars", 500, "回测模式：行情笔数")
		feeBP      = flag.String("fee-bp", "5", "回测模式：单边手续费率（bp，1 bp = 万分之一）")
		slippageBP = flag.String("slippage-bp", "1", "回测模式：撮合滑点（bp）")
	)
	flag.Parse()

	// 下单数量是「量」，用 decimal；并按交易所最小变动单位（stepSize）向下对齐 ——
	// 不对齐会被拒单，而 float64 做对齐会在 0.3/0.1 这类边界上少卖一个 step。
	const qtyStep = "0.0001"
	qty := order.AlignQtyDown(decimal.RequireFromString("0.002"), decimal.RequireFromString(qtyStep))

	// 回测模式：同一根管线喂确定的价格序列，最后打印一份盈亏报告。
	// 提前 return，避免再走一遍行情/存储/Redis 的装配。
	if *mode == "backtest" {
		if err := runBacktest(backtestOpts{
			CSVPath: *csvPath, Pattern: *pattern, Bars: *bars,
			FeeBP: *feeBP, SlippageBP: *slippageBP,
			Symbol: *symbol, Qty: qty,
		}); err != nil {
			fmt.Fprintln(os.Stderr, "回测失败：", err)
			os.Exit(1) // 明确以非零码退出，而不是 panic
		}
		return
	}
	if *mode != "live" {
		fmt.Fprintf(os.Stderr, "未知运行模式 %q（可选：live｜backtest）\n", *mode)
		os.Exit(1)
	}

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
	case "bitget-ws":
		// 推送式行情：同一个 Feed 接口，下游代码一行都不用改
		feed = &market.BitgetWS{
			Symbol: *symbol,
			OnError: func(err error) {
				failedTicks.Add(1)
				fmt.Println("     ⚠️ WS 异常（将自动重连）：", err)
			},
		}
	default:
		feed = market.RandomWalk{
			Symbol: *symbol, Start: 63000, Step: *interval, Vol: 0.006,
			Rnd: rand.New(rand.NewSource(7)), // 固定种子，输出可复现
		}
	}

	strategies := []strategy.Strategy{strategy.NewMACross(5, *threshold)}
	guard := risk.New(envDecimal("BOT_MAX_NOTIONAL", "500"), 25*time.Millisecond, 0.02)

	// 幂等实现：配了 BOT_REDIS_ADDR 就用 Redis SetNX（跨进程有效），否则退回内存 map。
	// executor 只依赖 Deduper 接口，换实现不用改它一行代码（L1-09）。
	var dedup executor.Deduper = executor.NewMemoryDeduper()
	if addr := store.RedisAddrFromEnv(); addr != "" {
		rd, err := store.NewRedisDeduper(addr, 9, time.Hour) // 内部已做 2s PING，连不上就地报错
		if err != nil {
			fmt.Println("  ⚠ Redis 不可用，幂等降级为内存 map：", err)
		} else {
			dedup = rd
			defer rd.Close()
		}
	}

	// 落库实现：配了 BOT_DB_LINK 就走 GoFrame 的 g.DB()（同 L3-07），否则只存内存。
	var persist store.Store = store.NewMemoryStore()
	if link := store.LinkFromEnv(); link != "" {
		initCtx, initCancel := context.WithTimeout(context.Background(), 3*time.Second)
		ms, err := store.NewMySQL(initCtx, link)
		initCancel()
		if err != nil {
			fmt.Println("  ⚠ MySQL 不可用，订单只存内存：", err)
		} else {
			persist = ms
		}
	}

	// 随机源必须每个组件各自持有：以前共用一个 *rand.Rand 被 -race 抓出 8 处竞态。
	engine := executor.NewEngine(executor.NewMock(11, 0.35, 0.25, 300*time.Millisecond), dedup, 2)
	// 通知：token/chat id 从环境变量读（TG_BOT_TOKEN + TG_CHAT_ID，兼容旧的 TG_TOKEN）。
	// 没配就自动降级为本地打印 —— 通知挂了绝不能影响下单。
	var sink notify.Notifier = notify.NewTelegramFromEnv()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	before := runtime.NumGoroutine()
	quotes := feed.Subscribe(ctx, *ticks)

	var (
		gotQuotes int
		signals   int
		accepted  int
		persisted int
		denied    = map[string]int{}
	)
	fmt.Printf("开始跑管线：行情源=%s 交易对=%s，目标 %d 个 tick，间隔 %v，阈值 %g\n",
		*feedName, *symbol, *ticks, *interval, *threshold)
	fmt.Printf("  风控上限 %s USD（qty 已按 step=%s 对齐 → %s）｜幂等：%s｜存储：%s\n\n",
		guard.MaxNotional, qtyStep, qty, dedup.Name(), persist.Name())

	for q := range quotes {
		gotQuotes++
		for _, s := range strategies {
			sig, ok := s.OnQuote(q)
			if !ok {
				continue
			}
			signals++

			// 下单价格从**交易所原始字符串**转 decimal，不用 decimal.NewFromFloat(q.Price)：
			// 后者会把 float64 的误差重新带进金额（07-web3-trading W-02 的坑）。
			price, perr := decimal.NewFromString(q.PriceStr)
			if perr != nil {
				fmt.Println("     ⚠ 交易所原始价格无法解析，跳过下单：", perr)
				continue
			}

			fmt.Printf("  行情 #%02d %-9s 价 %.2f（原始 %q）→ %s 发出 %s 信号（强度 %g）\n",
				gotQuotes, q.Symbol, q.Price, q.PriceStr, s.Name(), sig.Side, sig.Strength)

			if err := guard.Check(sig, qty, price, time.Now()); err != nil {
				denied[err.Error()]++
				fmt.Println("     ⛔ 风控拦截：", err)
				continue
			}

			ord := &executor.Order{
				ID: fmt.Sprintf("ORD-%s-%d", sig.Side, gotQuotes), Symbol: sig.Symbol, Side: sig.Side, Qty: qty,
			}
			err := engine.Submit(ctx, ord)
			if err != nil {
				fmt.Printf("     ✗ 下单失败：%v（尝试 %d 次，状态 %s/%s）\n", err, ord.Attempts, ord.Status, ord.State)
			} else {
				accepted++
				fmt.Printf("     ✓ 下单成功：%s %s %s（尝试 %d 次，状态 %s）\n", ord.ID, ord.Symbol, ord.Qty, ord.Attempts, ord.State)
				sink.Notify(fmt.Sprintf("%s %s %s 已受理", ord.Symbol, ord.Side, ord.ID))
			}

			// 成功与失败都要落库：失败记录才是复盘时最需要的东西。
			// 这里**记录订单状态**（State），不再只写终态字符串 —— 状态会随成交推进，
			// 落库按订单号 upsert，同一条订单的中间态能被不断刷新。
			if serr := persist.Save(ctx, store.Record{
				OrderID: ord.ID, Symbol: ord.Symbol, Side: ord.Side, Qty: ord.Qty,
				State: ord.State, Status: ord.Status, Attempts: ord.Attempts, Ts: ord.Ts,
			}); serr != nil {
				fmt.Println("     ⚠ 落库失败：", serr)
			} else {
				persisted++
			}
		}
	}

	// 幂等演示：用零失败的独立交易所 + 自己的内存去重表，保证结论稳定可复现
	idem := executor.NewEngine(executor.NewMock(1, 0, 0, 10*time.Millisecond), executor.NewMemoryDeduper(), 2)
	first := &executor.Order{ID: "ORD-IDEM", Symbol: "BTCUSDT", Side: "BUY", Qty: qty}
	second := &executor.Order{ID: "ORD-IDEM", Symbol: "BTCUSDT", Side: "BUY", Qty: qty}
	err1 := idem.Submit(ctx, first)
	err2 := idem.Submit(ctx, second)
	fmt.Printf("\n幂等演示：首次 err=%v（%s/%s）；同一订单号再提 err=%v（%s/%s，状态不动）\n",
		err1, first.Status, first.State, err2, second.Status, second.State)

	// 订单状态机演示：submitted → partial → filled；最后再喂一次同样的成交量验证幂等。
	// 交易所的成交回调本来就会重复推送，重复应用不该报错、也不该把状态改回去。
	half := first.Qty.Div(decimal.NewFromInt(2))
	stPartial, ferr1 := idem.ApplyFill(ctx, first.ID, half)
	stFilled, ferr2 := idem.ApplyFill(ctx, first.ID, first.Qty)
	stAgain, ferr3 := idem.ApplyFill(ctx, first.ID, first.Qty) // 重复回调：幂等
	fmt.Printf("状态机演示：%s --部分成交(%s)--> %s(err=%v) --全部成交--> %s(err=%v) --重复回调--> %s(err=%v)\n",
		order.Submitted, half, stPartial, ferr1, stFilled, ferr2, stAgain, ferr3)

	// 同一订单号两次落库：内存与 MySQL 都按订单号 upsert，最终只留一行（状态被刷成 filled）。
	for _, st := range []order.State{order.Submitted, stFilled} {
		if serr := persist.Save(ctx, store.Record{
			OrderID: first.ID, Symbol: first.Symbol, Side: first.Side, Qty: first.Qty,
			State: st, Status: first.Status, Attempts: first.Attempts, Ts: first.Ts,
		}); serr != nil {
			fmt.Println("     ⚠ 落库失败：", serr)
		}
	}

	cancel() // 停掉行情协程（生产环境是收到退出信号后由 main 统一取消）
	// 行为接口断言：只有带这个方法的行情源需要释放 HTTP 空闲连接（L1-10 类型断言的实战用法）
	if idle, ok := feed.(interface{ CloseIdleConnections() }); ok {
		idle.CloseIdleConnections()
	}
	time.Sleep(50 * time.Millisecond)

	fmt.Println("\n汇总：")
	fmt.Printf("  收到行情 %d 笔｜行情拉取失败 %d 次｜产生信号 %d 个｜下单成功 %d 笔｜落库 %d 条（存储：%s）\n",
		gotQuotes, failedTicks.Load(), signals, accepted, persisted, persist.Name())

	// 行为接口断言：只有 MySQL 实现能回读库里的累计行数，证明状态真的留在进程之外（L1-10）
	if c, ok := persist.(store.Counter); ok {
		ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel2()
		if n, cerr := c.Count(ctx2); cerr == nil {
			fmt.Printf("  回读 bot_orders 累计行数 = %d（重启进程也还在；同一订单号按 upsert 覆盖，不会重复写入）\n", n)
		}
	}
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

// backtestOpts 是回测模式的全部入参，集中成一个结构体便于单独测试。
type backtestOpts struct {
	CSVPath    string
	Pattern    string
	Bars       int
	FeeBP      string
	SlippageBP string
	Symbol     string
	Qty        decimal.Decimal
}

// runBacktest 准备价格序列、跑管线、打印报告。
//
// 全程不碰网络也不碰密钥：行情要么来自内置形态（固定种子），要么来自本地 CSV。
// 任何一步出错都返回 error（由 main 打印并以非零码退出），不 panic。
func runBacktest(opts backtestOpts) error {
	var (
		quotes []market.Quote
		pat    backtest.Pattern
		err    error
	)

	if opts.CSVPath != "" {
		quotes, err = backtest.LoadCSV(opts.CSVPath)
		pat = backtest.PatternCSV
	} else {
		if pat, err = backtest.ParsePattern(opts.Pattern); err != nil {
			return err
		}
		quotes, err = backtest.Series(pat, opts.Bars, 7) // 固定种子 7，结果可复现
	}
	if err != nil {
		return err
	}

	feeBP, err := decimal.NewFromString(opts.FeeBP)
	if err != nil {
		return fmt.Errorf("手续费率 %q 不是数字: %w", opts.FeeBP, err)
	}
	slippageBP, err := decimal.NewFromString(opts.SlippageBP)
	if err != nil {
		return fmt.Errorf("滑点 %q 不是数字: %w", opts.SlippageBP, err)
	}

	cfg := backtest.DefaultConfig(quotes, pat)
	cfg.Symbol, cfg.Qty, cfg.FeeBP, cfg.SlippageBP = opts.Symbol, opts.Qty, feeBP, slippageBP

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Printf("开始回测：形态=%s（%s）｜交易对=%s｜行情 %d 笔｜单笔 %s\n\n",
		pat, backtest.PatternName(pat), opts.Symbol, len(quotes), opts.Qty)

	res, err := backtest.Run(ctx, cfg)
	if err != nil {
		return err
	}
	fmt.Print(res.Report())
	return nil
}
