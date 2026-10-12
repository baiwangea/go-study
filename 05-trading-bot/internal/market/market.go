// Package market 只负责一件事：把行情以 channel 的形式推给下游。
// 先接 mock 源（本地随机游走，无需网络），换成 CEX WebSocket 时接口不变。
package market

import (
	"context"
	"math/rand"
	"strconv"
	"time"
)

// Quote 是一根行情。
//
// Price 与 PriceStr 是**同一价格的两种表示**，用途不同：
//   - Price（float64）：给策略算均线用。均线、偏离度是统计量，浮点误差无所谓；
//   - PriceStr（string）：下单时转 decimal 用。交易所返回的本来就是 "lastPr":"84800.63"
//     这样的字符串，它自己在避开浮点误差；用 decimal.NewFromFloat(Price) 会把
//     float64 的误差重新带进金额（07-web3-trading W-02 讲过的坑）。
type Quote struct {
	Symbol   string
	Price    float64
	PriceStr string
	Ts       time.Time
}

// Feed 是行情源接口：实现它的任何类型（mock / Binance / OKX / 链上池子）都能插进管线。
type Feed interface {
	Subscribe(ctx context.Context, n int) <-chan Quote
}

// defaultSeed 是 RandomWalk 没给随机源时的兜底种子：固定值，保证「忘了设 Rnd」时
// 依然是可复现的，而不是每次跑都不一样。
const defaultSeed = 1

// RandomWalk 是本地 mock 源：价格随机游走，用于跑通链路与写测试。
type RandomWalk struct {
	Symbol string
	Start  float64
	Step   time.Duration
	Vol    float64 // 每 tick 的相对波动幅度
	Rnd    *rand.Rand
}

func (f RandomWalk) Subscribe(ctx context.Context, n int) <-chan Quote {
	out := make(chan Quote)
	go func() {
		defer close(out) // 只有生产者有权关闭通道（L2-05）

		// Rnd 为 nil 时兜底：裸着解引用会 panic，把整条管线带走。
		rnd := f.Rnd
		if rnd == nil {
			rnd = rand.New(rand.NewSource(defaultSeed))
		}

		price := f.Start
		for i := 0; i < n; i++ {
			vol := f.Vol
			if vol == 0 {
				vol = 0.004
			}
			price += (rnd.Float64() - 0.48) * price * vol // 轻微上偏的随机游走
			select {
			case <-ctx.Done(): // 取消传播（L2-11）
				return
			case <-time.After(f.Step):
				// PriceStr 用 -1 精度（最短且能无损还原的表示），模拟交易所返回的原始字符串
				out <- Quote{
					Symbol:   f.Symbol,
					Price:    price,
					PriceStr: strconv.FormatFloat(price, 'f', -1, 64),
					Ts:       time.Now(),
				}
			}
		}
	}()
	return out
}
