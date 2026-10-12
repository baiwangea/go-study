// Package risk 是独立的守门层：所有下单请求都必须经过它，且它可以被单独测试。
package risk

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/shopspring/decimal"

	"go-study/trading-bot/internal/strategy"
)

var ErrCooldown = errors.New("冷却期内重复信号")

type Guard struct {
	// MaxNotional 是「钱」：单笔名义金额上限（USD），用 decimal 避免 float64 累加误差。
	MaxNotional decimal.Decimal
	Cooldown    time.Duration // 同方向最小间隔
	// MaxSlippage 是「概率/统计量」不是钱，保持 float64。
	MaxSlippage float64
	// LastBySymbol 是裸 map：Check 会被多个 worker 并发调用，必须加锁保护。
	LastBySymbol map[string]time.Time

	mu sync.Mutex
}

func New(maxNotional decimal.Decimal, cooldown time.Duration, maxSlippage float64) *Guard {
	return &Guard{
		MaxNotional: maxNotional, Cooldown: cooldown, MaxSlippage: maxSlippage,
		LastBySymbol: map[string]time.Time{},
	}
}

// Check 返回 nil 才允许下单；返回 error 说明被哪条规则拦下。
//
// 整段持锁：检查与「写冷却时间」必须原子，否则两个并发的同方向信号会同时看到「冷却期外」
// 而双双放行 —— 冷却就形同虚设。这里没有 I/O，持锁时间可以忽略。
func (g *Guard) Check(sig *strategy.Signal, qty, price decimal.Decimal, now time.Time) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// 名义金额 = 数量 × 价格，全程 decimal：qty 与 price 都来自交易所的原始字符串
	if notional := qty.Mul(price); notional.GreaterThan(g.MaxNotional) {
		return fmt.Errorf("名义金额 %s 超出上限 %s",
			notional.StringFixed(2), g.MaxNotional.StringFixed(2))
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
