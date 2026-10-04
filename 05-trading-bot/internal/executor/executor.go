// Package executor 负责真正提交订单：超时取消、重试退避、幂等去重。
// 这里是 L2（并发）与 L3（框架配置/日志）能力真正兑现的地方。
package executor

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

type Order struct {
	ID       string
	Symbol   string
	Side     string
	Qty      float64
	Status   string // accepted / canceled / failed
	Attempts int
}

// Exchange 抽象交易所：mock 实现会随机失败/慢响应，便于观察重试与超时分支。
type Exchange interface {
	Submit(ctx context.Context, o *Order) error
}

type Mock struct {
	Rnd        *rand.Rand
	FailRate   float64
	SlowRate   float64
	TimeoutDur time.Duration
}

func (m *Mock) Submit(ctx context.Context, o *Order) error {
	if m.Rnd.Float64() < m.SlowRate {
		select {
		case <-time.After(m.TimeoutDur):
			return nil
		case <-ctx.Done():
			return fmt.Errorf("交易所响应超时：%w", ctx.Err())
		}
	}
	if m.Rnd.Float64() < m.FailRate {
		return fmt.Errorf("交易所拒绝：限频")
	}
	return nil
}

type Engine struct {
	exch     Exchange
	maxRetry int
	seen     map[string]bool // 幂等表：生产环境换 Redis SetNX 或唯一索引
}

func NewEngine(exch Exchange, maxRetry int) *Engine {
	return &Engine{exch: exch, maxRetry: maxRetry, seen: map[string]bool{}}
}

// Submit 带幂等、指数退避重试与单次超时；返回最终状态。
func (e *Engine) Submit(ctx context.Context, o *Order) error {
	if e.seen[o.ID] {
		o.Status = "duplicate"
		return fmt.Errorf("重复提交 %s，已忽略", o.ID)
	}
	e.seen[o.ID] = true

	var lastErr error
	for attempt := 1; attempt <= e.maxRetry+1; attempt++ {
		o.Attempts = attempt
		callCtx, cancel := context.WithTimeout(ctx, 120*time.Millisecond) // 下单必须限时
		err := e.exch.Submit(callCtx, o)
		cancel()
		if err == nil {
			o.Status = "accepted"
			return nil
		}
		lastErr = err
		if attempt > e.maxRetry {
			break
		}
		backoff := time.Duration(20*(1<<(attempt-1))) * time.Millisecond
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			o.Status = "canceled"
			return ctx.Err()
		}
	}
	o.Status = "failed"
	return lastErr
}

func (e *Engine) Processed() int { return len(e.seen) }
