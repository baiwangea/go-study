// Package strategy 把行情变成信号。注意它只有输入输出处方，不做任何风控与下单。
package strategy

import "go-study/trading-bot/internal/market"

type Signal struct {
	Symbol   string
	Side     string // BUY / SELL
	Strength float64
}

// Strategy 是接口而非基类 —— 新策略只要满足方法就能接进管线（L1-09 隐式实现）。
type Strategy interface {
	Name() string
	OnQuote(q market.Quote) (*Signal, bool)
}

// MACross 演示用法：价格穿过移动平均即出信号。
type MACross struct {
	window    int
	prev      float64 // 0 表示尚无基准
	Threshold float64 // 触发所需的最小相对偏离
}

func NewMACross(window int, threshold float64) *MACross {
	return &MACross{window: window, Threshold: threshold}
}

func (s *MACross) Name() string { return "ma-cross" }

func (s *MACross) OnQuote(q market.Quote) (*Signal, bool) {
	if s.prev == 0 {
		s.prev = q.Price
		return nil, false
	}
	dev := (q.Price - s.prev) / s.prev
	s.prev = q.Price

	_ = s.window
	threshold := s.Threshold
	if dev > threshold {
		return &Signal{Symbol: q.Symbol, Side: "BUY", Strength: dev}, true
	}
	if dev < -threshold {
		return &Signal{Symbol: q.Symbol, Side: "SELL", Strength: -dev}, true
	}
	return nil, false
}
