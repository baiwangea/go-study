package backtest

import "github.com/shopspring/decimal"

// Fill 是一笔回测成交。金额与数量一律 decimal。
type Fill struct {
	OrderID string
	Symbol  string
	Side    string // BUY / SELL
	Qty     decimal.Decimal
	Price   decimal.Decimal
}

// Stats 是回测的盈亏统计。
//
// **金额全部是 decimal**（单位 USDT），float64 只出现在统计量上（策略强度、均线）。
// 用 float64 累加几十笔成交的毛利，末位误差会一路放大，最后「净利 9.129999999」这种数字
// 让人没法判断到底是策略亏了还是浮点亏了。
type Stats struct {
	Fills        int             // 成交笔数
	ClosedTrades int             // 平仓笔数（胜率的分母）
	Wins         int             // 盈利的平仓笔数（胜率的分子）
	WinRate      decimal.Decimal // 胜率，百分比 0~100
	Gross        decimal.Decimal // 毛利（未扣手续费）
	Fee          decimal.Decimal // 手续费合计（双边）
	Net          decimal.Decimal // 净利 = 毛利 − 手续费
	MaxDrawdown  decimal.Decimal // 最大回撤
	OpenQty      decimal.Decimal // 期末持仓，有符号：正=多，负=空
}

// Compute 把成交流水折成盈亏。
//
// 记账方式是「持仓 + 开仓均价」，一个仓位从头跟到尾：
//   - 空仓时，直接以本次成交价开仓；
//   - 与持仓同向时，加仓，开仓均价按成交量加权；
//   - 与持仓反向时，先平掉 min(|持仓|, |本次量|)，毛利 = 平掉量 ×（平仓价 − 开仓均价）× 持仓方向；
//     平不完的部分反手开新仓，开仓价就是本次成交价。
//
// 手续费按**每笔成交**的名义金额单边收取：fee = qty × price × feeBP / 10000。
// 一开一平就是双边收费，与真实交易所一致。
//
// 最大回撤基于**逐笔成交后的已实现净权益曲线**（累计毛利 − 累计手续费）：
// peak 记历史最高权益，回撤 = peak − 当前权益，取其中最大值。
// 起点权益视为 0，所以「先亏一笔」本身就构成回撤。
func Compute(fills []Fill, feeBP decimal.Decimal) Stats {
	var (
		posQty = decimal.Zero // 有符号持仓
		posAvg = decimal.Zero // 开仓均价

		gross  = decimal.Zero
		fees   = decimal.Zero
		trades int
		wins   int

		peak  = decimal.Zero
		maxDD = decimal.Zero
	)

	for _, f := range fills {
		fees = fees.Add(f.Qty.Mul(f.Price).Mul(feeBP).Div(bpDenom))

		signed := f.Qty
		if f.Side == sideSell {
			signed = f.Qty.Neg()
		}

		switch {
		case posQty.IsZero():
			posQty, posAvg = signed, f.Price

		case posQty.IsPositive() == signed.IsPositive():
			// 同向加仓：按成交量加权更新开仓均价
			total := posQty.Add(signed)
			posAvg = posAvg.Mul(posQty).Add(f.Price.Mul(signed)).Div(total)
			posQty = total

		default:
			// 反向：先平仓。方向系数保证做空也能算对（持仓为负时，价格跌才是赚）。
			dir := decimal.NewFromInt(1)
			if posQty.IsNegative() {
				dir = decimal.NewFromInt(-1)
			}
			closeQty := decimal.Min(posQty.Abs(), signed.Abs())
			pnl := f.Price.Sub(posAvg).Mul(closeQty).Mul(dir)
			gross = gross.Add(pnl)
			trades++
			if pnl.IsPositive() {
				wins++
			}

			posQty = posQty.Add(signed)
			if posQty.IsZero() {
				posAvg = decimal.Zero
			} else if posQty.IsPositive() == signed.IsPositive() {
				posAvg = f.Price // 反手：剩下的是本次成交开的新仓
			}
			// 平不完仍有剩余时，posQty 与本次成交反向 → 开仓均价不变
		}

		equity := gross.Sub(fees)
		if equity.GreaterThan(peak) {
			peak = equity
		}
		if dd := peak.Sub(equity); dd.GreaterThan(maxDD) {
			maxDD = dd
		}
	}

	winRate := decimal.Zero
	if trades > 0 {
		winRate = decimal.NewFromInt(int64(wins)).
			Div(decimal.NewFromInt(int64(trades))).
			Mul(decimal.NewFromInt(100))
	}

	return Stats{
		Fills:        len(fills),
		ClosedTrades: trades,
		Wins:         wins,
		WinRate:      winRate,
		Gross:        gross,
		Fee:          fees,
		Net:          gross.Sub(fees),
		MaxDrawdown:  maxDD,
		OpenQty:      posQty,
	}
}
