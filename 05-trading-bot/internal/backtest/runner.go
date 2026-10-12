package backtest

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"go-study/trading-bot/internal/executor"
	"go-study/trading-bot/internal/market"
	"go-study/trading-bot/internal/risk"
	"go-study/trading-bot/internal/store"
	"go-study/trading-bot/internal/strategy"
)

// Config 是一次回测的全部参数。
type Config struct {
	Quotes  []market.Quote // 价格序列（内置形态或外部 CSV）
	Pattern Pattern        // 形态，只用于报告
	Symbol  string

	Qty         decimal.Decimal // 每次下单数量
	Window      int             // 均线窗口
	Threshold   float64         // 信号阈值（统计量，float64）
	MaxNotional decimal.Decimal // 风控：单笔名义金额上限
	Cooldown    time.Duration   // 风控：同方向冷却
	MaxSlippage float64         // 风控：信号强度上限

	FeeBP      decimal.Decimal // 单边手续费率（bp）
	SlippageBP decimal.Decimal // 撮合滑点（bp）
	MaxRetry   int
}

// DefaultConfig 给出一份「开箱即用」的回测参数：零网络、零密钥即可跑。
func DefaultConfig(quotes []market.Quote, p Pattern) Config {
	return Config{
		Quotes:  quotes,
		Pattern: p,
		Symbol:  DefaultSymbol,

		Qty:         decimal.RequireFromString("0.002"),
		Window:      5,
		Threshold:   0.002, // 0.2%：过滤掉均线的日常抖动，只留像样的动量
		MaxNotional: decimal.RequireFromString("500"),
		Cooldown:    5 * time.Minute, // 5 根 K 线（BarInterval=1min）内不重复同方向下单
		MaxSlippage: 0.02,

		FeeBP:      decimal.NewFromInt(5), // 5 bp/边 ≈ 主流所 taker 费率
		SlippageBP: decimal.NewFromInt(1),
		MaxRetry:   2,
	}
}

// Result 是一次回测的汇总，报告所需的数字全在这里。
type Result struct {
	Pattern Pattern
	Symbol  string
	FeeBP   decimal.Decimal

	Quotes    int // 喂进去的行情笔数
	Signals   int // 策略给出的信号数
	Denied    int // 被风控拦下的次数
	Orders    int // 真正提交的订单数
	Filled    int // 成交笔数
	Failed    int // 下单失败笔数
	Persisted int // 落库条数

	Stats Stats
}

// Run 把序列喂进**同一条管线**：strategy → risk → executor（回测交易所）→ store。
//
// 与 live 模式的差别只有两处：行情来自序列而不是网络，交易所是立即成交的回测交易所。
// 策略、风控、幂等、落库用的都是生产代码。
func Run(ctx context.Context, cfg Config) (Result, error) {
	if len(cfg.Quotes) == 0 {
		return Result{}, errors.New("行情序列为空，无法回测")
	}
	if cfg.Window <= 0 {
		return Result{}, fmt.Errorf("均线窗口必须为正，实际 %d", cfg.Window)
	}

	strat := strategy.NewMACross(cfg.Window, cfg.Threshold)
	guard := risk.New(cfg.MaxNotional, cfg.Cooldown, cfg.MaxSlippage)
	exch := NewExchange(cfg.SlippageBP)
	engine := executor.NewEngine(exch, executor.NewMemoryDeduper(), cfg.MaxRetry)
	persist := store.NewMemoryStore()

	res := Result{Pattern: cfg.Pattern, Symbol: cfg.Symbol, FeeBP: cfg.FeeBP, Quotes: len(cfg.Quotes)}

	for i, q := range cfg.Quotes {
		sig, ok := strat.OnQuote(q)
		if !ok {
			continue
		}
		res.Signals++

		// 下单价格从交易所原始字符串转 decimal（不是 decimal.NewFromFloat）：
		// 序列与真实交易所一样，字符串才是金额的事实来源。
		price, err := decimal.NewFromString(q.PriceStr)
		if err != nil {
			return Result{}, fmt.Errorf("第 %d 笔行情价格 %q 无法解析: %w", i+1, q.PriceStr, err)
		}

		// 风控用的是**行情时间** q.Ts：回测跑得快不代表行情快。
		if err := guard.Check(sig, cfg.Qty, price, q.Ts); err != nil {
			res.Denied++
			continue
		}

		res.Orders++
		exch.SetPrice(price)
		ord := &executor.Order{
			ID:     fmt.Sprintf("BT-%s-%06d", sig.Side, i),
			Symbol: sig.Symbol,
			Side:   sig.Side,
			Qty:    cfg.Qty,
		}
		if err := engine.Submit(ctx, ord); err != nil {
			res.Failed++
			continue
		}

		if err := persist.Save(ctx, store.Record{
			OrderID: ord.ID, Symbol: ord.Symbol, Side: ord.Side, Qty: ord.Qty,
			State: ord.State, Status: ord.Status, Attempts: ord.Attempts, Ts: ord.Ts,
		}); err == nil {
			res.Persisted++
		}
	}

	res.Filled = len(exch.Fills)
	res.Stats = Compute(exch.Fills, cfg.FeeBP)
	return res, nil
}

// Report 生成一眼能看懂的回测报告。
//
// 盈亏单位是 USDT，展示保留 2 位小数；持仓是「量」，用 decimal 的原始表示。
func (r Result) Report() string {
	var b strings.Builder

	base := strings.TrimSuffix(r.Symbol, "USDT")
	if base == "" {
		base = r.Symbol
	}

	fmt.Fprintln(&b, "════════ 回测报告 ════════")
	fmt.Fprintf(&b, "形态：%s ｜ 行情 %d 笔 ｜ 手续费 %s bp/边\n",
		PatternName(r.Pattern), r.Quotes, r.FeeBP.StringFixed(0))
	fmt.Fprintf(&b, "信号 %d 个 ｜ 风控拦截 %d 次 ｜ 下单 %d 笔 ｜ 成交 %d 笔 ｜ 失败 %d 笔\n",
		r.Signals, r.Denied, r.Orders, r.Filled, r.Failed)
	fmt.Fprintf(&b, "毛利 %s USDT ｜ 手续费 %s USDT ｜ 净利 %s USDT ｜ 胜率 %s%%\n",
		r.Stats.Gross.StringFixed(2), r.Stats.Fee.StringFixed(2),
		r.Stats.Net.StringFixed(2), r.Stats.WinRate.StringFixed(1))
	fmt.Fprintf(&b, "最大回撤 %s USDT ｜ 期末持仓 %s %s\n",
		r.Stats.MaxDrawdown.StringFixed(2), r.Stats.OpenQty.String(), base)

	return b.String()
}
