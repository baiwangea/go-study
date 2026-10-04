// Package market 只负责一件事：把行情以 channel 的形式推给下游。
// 先接 mock 源（本地随机游走，无需网络），换成 CEX WebSocket 时接口不变。
package market

import (
	"context"
	"math/rand"
	"time"
)

type Quote struct {
	Symbol string
	Price  float64
	Ts     time.Time
}

// Feed 是行情源接口：实现它的任何类型（mock / Binance / OKX / 链上池子）都能插进管线。
type Feed interface {
	Subscribe(ctx context.Context, n int) <-chan Quote
}

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
		price := f.Start
		for i := 0; i < n; i++ {
			vol := f.Vol
			if vol == 0 {
				vol = 0.004
			}
			price += (f.Rnd.Float64() - 0.48) * price * vol // 轻微上偏的随机游走
			select {
			case <-ctx.Done(): // 取消传播（L2-11）
				return
			case <-time.After(f.Step):
				out <- Quote{Symbol: f.Symbol, Price: price, Ts: time.Now()}
			}
		}
	}()
	return out
}
