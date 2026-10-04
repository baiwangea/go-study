// Package risk 是独立的守门层：所有下单请求都必须经过它，且它可以被单独测试。
package risk

import (
	"errors"
	"fmt"
	"time"

	"go-study/trading-bot/internal/strategy"
)

var ErrCooldown = errors.New("冷却期内重复信号")

type Guard struct {
	MaxNotional  float64       // 单笔名义金额上限（USD）
	Cooldown     time.Duration // 同方向最小间隔
	MaxSlippage  float64       // 允许的最大偏离强度
	LastBySymbol map[string]time.Time
}

func New(maxNotional float64, cooldown time.Duration, maxSlippage float64) *Guard {
	return &Guard{
		MaxNotional: maxNotional, Cooldown: cooldown, MaxSlippage: maxSlippage,
		LastBySymbol: map[string]time.Time{},
	}
}

// Check 返回 nil 才允许下单；返回 error 说明被哪条规则拦下。
func (g *Guard) Check(sig *strategy.Signal, qty, price float64, now time.Time) error {
	if notional := qty * price; notional > g.MaxNotional {
		return fmt.Errorf("名义金额 %.2f 超出上限 %.2f", notional, g.MaxNotional)
	}
	if sig.Strength > g.MaxSlippage {
		return fmt.Errorf("信号偏离 %.4f 异常，疑似脏数据", sig.Strength)
	}
	key := sig.Symbol + ":" + sig.Side
	if last, ok := g.LastBySymbol[key]; ok && now.Sub(last) < g.Cooldown {
		return ErrCooldown
	}
	g.LastBySymbol[key] = now
	return nil
}
