package backtest

import (
	"context"

	"github.com/shopspring/decimal"

	"go-study/trading-bot/internal/executor"
)

// bpDenom 是 bp（万分之一）的分母：费率 5 bp = 5/10000 = 0.0005。
// 手续费和滑点都以 bp 为单位配置，避免在命令行上敲一串 0.00005。
var bpDenom = decimal.NewFromInt(10000)

const (
	sideBuy  = "BUY"
	sideSell = "SELL"
)

// Exchange 是回测专用交易所：**立即成交、永不失败**，可选滑点。
//
// 它实现 executor.Exchange，所以能原封不动插进同一条 executor.Engine 管线 ——
// 幂等、重试、超时这些编排代码在回测里跑的是同一份实现，不是另写一套。
type Exchange struct {
	SlippageBP decimal.Decimal // 滑点，单位 bp；0 表示按信号价成交
	Fills      []Fill          // 按时间顺序记录每一笔成交

	price decimal.Decimal // 当前行情价，由 Runner 在每次下单前喂入
}

// 编译期契约：回测交易所就是 executor 认的那个 Exchange。
var _ executor.Exchange = (*Exchange)(nil)

func NewExchange(slippageBP decimal.Decimal) *Exchange {
	return &Exchange{SlippageBP: slippageBP}
}

// SetPrice 设置当前行情价。回测没有真实撮合，成交价完全由价格序列决定。
func (e *Exchange) SetPrice(p decimal.Decimal) { e.price = p }

// Submit 立即成交：把成交记进 Fills 并返回 nil。
// 永不失败 → 回测里不会触发重试分支，结论只反映策略本身。
func (e *Exchange) Submit(_ context.Context, o *executor.Order) error {
	e.Fills = append(e.Fills, Fill{
		OrderID: o.ID,
		Symbol:  o.Symbol,
		Side:    o.Side,
		Qty:     o.Qty,
		Price:   e.fillPrice(o.Side, e.price),
	})
	return nil
}

// fillPrice 把滑点算在「更差」的方向上：买入成交价上移、卖出成交价下移。
func (e *Exchange) fillPrice(side string, px decimal.Decimal) decimal.Decimal {
	if !e.SlippageBP.IsPositive() {
		return px
	}
	adj := px.Mul(e.SlippageBP).Div(bpDenom)
	if side == sideBuy {
		return px.Add(adj)
	}
	return px.Sub(adj)
}
