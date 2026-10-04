// Package executor 负责真正提交订单：幂等去重、指数退避重试、单次超时。
// 这里是 L2（并发/超时）与 L3（Redis 幂等、g.DB 落库）能力真正兑现的地方。
package executor

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Order 是一笔提交的订单。
type Order struct {
	ID       string
	Symbol   string
	Side     string
	Qty      float64
	Status   string // accepted / failed / duplicate
	Attempts int
	Ts       time.Time
}

// Deduper 是幂等契约。**接口定义在使用方**，实现可以放别处：
// 内存 map、Redis SetNX、数据库唯一索引都能接上，调用方一行不用改（L1-09 隐式实现）。
type Deduper interface {
	MarkOnce(ctx context.Context, key string) (bool, error)
	Name() string
}

// MemoryDeduper 只在单进程内有效，重启即失效 —— 生产环境请换 RedisDeduper。
type MemoryDeduper struct {
	mu   sync.Mutex
	seen map[string]bool
}

func NewMemoryDeduper() *MemoryDeduper { return &MemoryDeduper{seen: map[string]bool{}} }

func (m *MemoryDeduper) MarkOnce(_ context.Context, key string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.seen[key] {
		return false, nil
	}
	m.seen[key] = true
	return true, nil
}

func (m *MemoryDeduper) Name() string { return "内存 map（重启失效）" }

// Exchange 抽象交易所。Submit 只负责「这一次调用成没成」，不管重试。
type Exchange interface {
	Submit(ctx context.Context, o *Order) error
}

// Mock 是假交易所：按概率失败与慢响应，用来把重试和超时分支逼出来。
type Mock struct {
	FailRate   float64
	SlowRate   float64
	TimeoutDur time.Duration

	// 随机源必须每个组件各自持有：之前把同一个 *rand.Rand 同时交给行情协程与下单逻辑，
	// 被 go run -race 抓出 8 处 DATA RACE —— math/rand.Rand 内部没有锁。
	rnd *rand.Rand
}

// NewMock 用固定种子构造，保证运行结果可复现。
func NewMock(seed int64, failRate, slowRate float64, timeout time.Duration) *Mock {
	return &Mock{
		FailRate: failRate, SlowRate: slowRate, TimeoutDur: timeout,
		rnd: rand.New(rand.NewSource(seed)),
	}
}

func (m *Mock) Submit(ctx context.Context, o *Order) error {
	if m.rnd.Float64() < m.SlowRate {
		select {
		case <-time.After(m.TimeoutDur):
			return nil
		case <-ctx.Done():
			return fmt.Errorf("交易所响应超时：%w", ctx.Err())
		}
	}
	if m.rnd.Float64() < m.FailRate {
		return fmt.Errorf("交易所拒绝：限频")
	}
	return nil
}

// Engine 把「幂等 → 限时调用 → 退避重试」编排成一条固定流程。
type Engine struct {
	exch     Exchange
	dedup    Deduper
	maxRetry int
	perCall  time.Duration
}

func NewEngine(exch Exchange, dedup Deduper, maxRetry int) *Engine {
	if dedup == nil {
		dedup = NewMemoryDeduper()
	}
	return &Engine{exch: exch, dedup: dedup, maxRetry: maxRetry, perCall: 120 * time.Millisecond}
}

// DeduperName 暴露当前幂等实现，方便运行日志里说清楚状态存在哪。
func (e *Engine) DeduperName() string { return e.dedup.Name() }

// Submit 先查幂等，再带超时提交，失败按指数退避重试。
func (e *Engine) Submit(ctx context.Context, o *Order) error {
	o.Ts = time.Now()

	// 1) 幂等：同一个订单号只允许真正提交一次
	first, err := e.dedup.MarkOnce(ctx, o.ID)
	if err != nil {
		o.Status = "dedupe-error"
		return fmt.Errorf("幂等检查失败（宁可不下单，也不能重复下单）: %w", err)
	}
	if !first {
		o.Status = "duplicate"
		return fmt.Errorf("重复提交 %s，已忽略", o.ID)
	}

	// 2) 提交 + 退避重试
	var lastErr error
	for attempt := 1; attempt <= e.maxRetry+1; attempt++ {
		o.Attempts = attempt

		callCtx, cancel := context.WithTimeout(ctx, e.perCall) // 每次调用都单独限时
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

		backoff := time.Duration(20*(1<<(attempt-1))) * time.Millisecond // 20/40/80ms 指数退避
		select {
		case <-time.After(backoff):
		case <-ctx.Done(): // 整体被取消：不再重试，但订单已被标记处理过
			o.Status = "canceled"
			return ctx.Err()
		}
	}
	o.Status = "failed"
	return lastErr
}
