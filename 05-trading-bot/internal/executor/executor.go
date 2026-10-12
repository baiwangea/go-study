// Package executor 负责真正提交订单：幂等去重、指数退避重试、单次超时。
// 这里是 L2（并发/超时）与 L3（Redis 幂等、g.DB 落库）能力真正兑现的地方。
package executor

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/shopspring/decimal"

	"go-study/trading-bot/internal/order"
)

// Order 是一笔提交的订单。
//
// Qty / FilledQty 是 decimal：**钱和数量不用 float64**。0.1 + 0.2 != 0.3 这种误差
// 在累加名义金额、比对成交量时会被放大成「对不上账」。
// State 是生命周期状态（order.State，由状态机推进）；Status 是面向日志/落库的提交结果字符串
// （accepted / failed / duplicate ...），两者关注点不同，都保留。
type Order struct {
	ID        string
	Symbol    string
	Side      string
	Qty       decimal.Decimal
	FilledQty decimal.Decimal
	State     order.State
	Status    string // accepted / failed / duplicate / canceled / dedupe-error
	Attempts  int
	Ts        time.Time
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

	// orders 记录本引擎提交成功的订单，ApplyFill 靠它拿到 Qty 来推导 partial/filled。
	// 加锁：Submit 写、ApplyFill 读，成交回报通常来自另一个协程。
	mu     sync.Mutex
	orders map[string]*Order
}

func NewEngine(exch Exchange, dedup Deduper, maxRetry int) *Engine {
	if dedup == nil {
		dedup = NewMemoryDeduper()
	}
	return &Engine{
		exch: exch, dedup: dedup, maxRetry: maxRetry, perCall: 120 * time.Millisecond,
		orders: map[string]*Order{},
	}
}

// DeduperName 暴露当前幂等实现，方便运行日志里说清楚状态存在哪。
func (e *Engine) DeduperName() string { return e.dedup.Name() }

// Submit 先查幂等，再带超时提交，失败按指数退避重试。
//
// 状态机接入点（与 order 包的跃迁表对应）：
//   - 提交成功 → created → submitted
//   - 重试耗尽 → created → rejected
//   - 被幂等拦下 / 幂等检查出错 → **保持原状态**（这笔单根本没往交易所发）
func (e *Engine) Submit(ctx context.Context, o *Order) error {
	o.Ts = time.Now()
	if o.State == "" { // 零值归一化，保证后面每一步跃迁都有合法起点
		o.State = order.Created
	}

	// 1) 幂等：同一个订单号只允许真正提交一次
	first, err := e.dedup.MarkOnce(ctx, o.ID)
	if err != nil {
		o.Status = "dedupe-error"
		return fmt.Errorf("幂等检查失败（宁可不下单，也不能重复下单）: %w", err)
	}
	if !first {
		o.Status = "duplicate" // 状态不动：重复提交不改订单生命周期
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
			o.State, _ = order.Transition(o.State, order.Submitted) // 起点已归一化，必定合法
			e.register(o)                                           // 登记后才能收成交回报
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
			o.State, _ = order.Transition(o.State, order.Rejected) // 没提交成功 ≈ 被拒
			return ctx.Err()
		}
	}
	o.Status = "failed"
	o.State, _ = order.Transition(o.State, order.Rejected)
	return lastErr
}

// register 记录一笔已受理的订单，供后续成交回报更新状态。
func (e *Engine) register(o *Order) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.orders[o.ID] = o
}

// ApplyFill 根据「累计成交量」推进订单状态，返回推进后的状态。
//
//	filledQty == Qty       → filled（一次全成，或 partial 之后补满）
//	0 < filledQty < Qty    → partial（首次部分成交；继续部分成交时 partial → partial 幂等）
//	filledQty == 0         → 保持 submitted（还没成交）
//
// 幂等：交易所的成交回调会重复推送，同一个 filledQty 再来一次是「无变化」而不是错误。
// 不变量：0 ≤ filledQty ≤ Qty，违反时返回 ErrFillInvariant 且不改动订单。
// 订单必须是在本引擎提交成功的（否则拿不到 Qty），未登记时返回错误。
//
// ctx 预留：真实交易所实现会用它去查订单的成交明细；本地状态机不需要。
func (e *Engine) ApplyFill(_ context.Context, orderID string, filledQty decimal.Decimal) (order.State, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	o, ok := e.orders[orderID]
	if !ok {
		return "", fmt.Errorf("订单 %s 不在本引擎的已受理列表中，无法更新成交", orderID)
	}

	next, err := order.StateForFill(o.Qty, filledQty) // 先校验不变量，再推导目标状态
	if err != nil {
		return o.State, err
	}
	st, err := order.Transition(o.State, next)
	if err != nil {
		return o.State, err
	}
	o.State = st
	o.FilledQty = filledQty
	return st, nil
}
