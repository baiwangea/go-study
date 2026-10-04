// Package strategy 把行情变成信号：只有输入输出处方，不做风控、不下单、不通知。
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

// MACross 是真正的移动平均交叉：比较「本根均线」与「上一根均线」的变化幅度。
type MACross struct {
	window    int
	prices    []float64
	peak      float64 // 实测到的最大均线偏离，用于校准阈值
	Threshold float64 // 均线变化的最小相对偏离，才认为是有效信号
}

func NewMACross(window int, threshold float64) *MACross {
	return &MACross{window: window, prices: make([]float64, 0, window+1), Threshold: threshold}
}

func (s *MACross) Name() string { return "ma-cross" }

// Peak 返回本次运行实测的最大均线偏离 —— 调阈值前先看看真实波动到底有多大，
// 否则阈值设高了就是「看起来在跑、其实永远不出信号」。
func (s *MACross) Peak() float64 { return s.peak }

func (s *MACross) OnQuote(q market.Quote) (*Signal, bool) {
	s.prices = append(s.prices, q.Price)
	if len(s.prices) > s.window+1 {
		s.prices = s.prices[len(s.prices)-(s.window+1):] // 只保留必要长度，避免无限增长
	}
	if len(s.prices) < s.window+1 {
		return nil, false // 数据还不够，算不出两根均线
	}

	cur, prev := avg(s.prices[len(s.prices)-s.window:]), avg(s.prices[:s.window])
	dev := (cur - prev) / prev

	abs := dev
	if abs < 0 {
		abs = -abs
	}
	if abs > s.peak { // 先记录实测幅度，方便下次把阈值调到合理区间
		s.peak = abs
	}
	if abs < s.Threshold {
		return nil, false
	}
	if dev > 0 {
		return &Signal{Symbol: q.Symbol, Side: "BUY", Strength: abs}, true
	}
	return &Signal{Symbol: q.Symbol, Side: "SELL", Strength: abs}, true
}

func avg(xs []float64) float64 {
	var sum float64
	for _, v := range xs {
		sum += v
	}
	return sum / float64(len(xs))
}
