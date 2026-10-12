// Package order 是订单生命周期的唯一事实来源：状态定义、合法跃迁、成交不变量。
//
// 它刻意只依赖 decimal，不依赖 executor / store —— 这样状态机可以被单独测试，
// 也能被 executor、store、回测器同时复用（L1-09：职责单一，实现可替换）。
//
// 跃迁表（第一列是起点，后面是允许到达的终点）：
//
//	created   → submitted, rejected
//	submitted → partial, filled, canceled
//	partial   → partial, filled, canceled
//	filled    → （终态，无出边）
//	canceled  → （终态，无出边）
//	rejected  → （终态，无出边）
//
// 两条容易踩的规则：
//   - **幂等**：Transition(s, s) 视为「无变化」而不是错误。交易所的成交回调会重复推送，
//     同一个 filled 再来一次不该把管线打成错误。
//   - **终态**：filled / canceled / rejected 出不去（同状态无变化除外）。
package order

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

// State 是订单状态。用字符串而不是 int：日志、落库、告警里都能直接读，
// 也不会因为加一个枚举值就把历史数据对不上。
type State string

const (
	Created   State = "created"   // 本地已创建，尚未提交交易所
	Submitted State = "submitted" // 交易所已受理
	Partial   State = "partial"   // 部分成交
	Filled    State = "filled"    // 全部成交
	Canceled  State = "canceled"  // 已撤单
	Rejected  State = "rejected"  // 被拒（提交失败 / 风控拒绝）
)

var (
	// ErrUnknownState 表示出现了跃迁表里没定义的状态（多为拼写错误或跨版本数据）。
	ErrUnknownState = errors.New("未知订单状态")
	// ErrIllegalTransition 表示跃迁表里明确禁止的跃迁（例如终态出边）。
	ErrIllegalTransition = errors.New("非法的订单状态跃迁")
	// ErrFillInvariant 表示成交量违反了不变量（负数，或超过下单数量）。
	ErrFillInvariant = errors.New("成交量不满足不变量")
)

// transitions 是跃迁表：key 为起点，value 为允许到达的终点（不含「同状态无变化」）。
var transitions = map[State][]State{
	Created:   {Submitted, Rejected},
	Submitted: {Partial, Filled, Canceled},
	Partial:   {Partial, Filled, Canceled},
	Filled:    {},
	Canceled:  {},
	Rejected:  {},
}

// stateOrder 固定状态的展示顺序，日志里打印跃迁表时顺序稳定，diff 才好看。
var stateOrder = []State{Created, Submitted, Partial, Filled, Canceled, Rejected}

// AllStates 返回全部状态（副本，调用方改不动内部表）。
func AllStates() []State {
	out := make([]State, len(stateOrder))
	copy(out, stateOrder)
	return out
}

// Valid 判断 s 是不是跃迁表里定义过的状态。
func Valid(s State) bool {
	_, ok := transitions[s]
	return ok
}

// IsTerminal 判断 s 是不是终态（终态出不去，只能保持原状态）。
func IsTerminal(s State) bool {
	targets, ok := transitions[s]
	return ok && len(targets) == 0
}

// Allowed 返回 from 允许到达的终点（副本）。方便日志里打印「当前状态还能去哪」。
// from 非法时返回 nil。
func Allowed(from State) []State {
	targets, ok := transitions[from]
	if !ok {
		return nil
	}
	out := make([]State, len(targets))
	copy(out, targets)
	return out
}

// CanTransition 判断 from → to 是否合法。同状态（含终态）视为「无变化」，返回 true。
func CanTransition(from, to State) bool {
	if !Valid(from) || !Valid(to) {
		return false
	}
	if from == to {
		return true // 幂等：重复回调不改变状态
	}
	for _, s := range transitions[from] {
		if s == to {
			return true
		}
	}
	return false
}

// Transition 执行一次跃迁：合法（含同状态幂等）返回新状态与 nil，非法返回原状态与明确错误。
//
// 返回原状态而不是零值，是为了让调用方能安全地 `s, err := Transition(...)` ——
// 出错时 s 仍是「当前状态」，日志里不会突然变成空字符串。
func Transition(from, to State) (State, error) {
	if !Valid(from) {
		return from, fmt.Errorf("%w: 起点 %q", ErrUnknownState, from)
	}
	if !Valid(to) {
		return from, fmt.Errorf("%w: 终点 %q", ErrUnknownState, to)
	}
	if from == to {
		return from, nil // 幂等：交易所重复推送
	}
	if !CanTransition(from, to) {
		if IsTerminal(from) {
			return from, fmt.Errorf("%w: %s 是终态，不能再到 %s", ErrIllegalTransition, from, to)
		}
		return from, fmt.Errorf("%w: %s → %s", ErrIllegalTransition, from, to)
	}
	return to, nil
}

// CheckFill 校验成交量不变量：0 ≤ FilledQty ≤ Qty。
// 违反时必须报错而不是静默截断 —— 静默截断会把「交易所多报了一笔」这种严重问题藏起来。
func CheckFill(qty, filledQty decimal.Decimal) error {
	if filledQty.IsNegative() {
		return fmt.Errorf("%w: FilledQty=%s 不能为负", ErrFillInvariant, filledQty)
	}
	if filledQty.GreaterThan(qty) {
		return fmt.Errorf("%w: FilledQty=%s 超过 Qty=%s", ErrFillInvariant, filledQty, qty)
	}
	return nil
}

// StateForFill 按「累计成交量」推导订单状态：
//
//	FilledQty == Qty → filled（补满或一次全成）
//	0 < FilledQty < Qty → partial
//	FilledQty == 0 → submitted（还没成交，状态不变）
//
// 成交量违反不变量时返回错误。
func StateForFill(qty, filledQty decimal.Decimal) (State, error) {
	if err := CheckFill(qty, filledQty); err != nil {
		return "", err
	}
	switch {
	case filledQty.Equal(qty):
		return Filled, nil
	case filledQty.IsZero():
		return Submitted, nil
	default:
		return Partial, nil
	}
}

// AlignQtyDown 把数量按交易所的最小变动单位 step 向下取整（不足一个 step 的部分直接丢掉）。
//
// 为什么不用 `qty - math.Mod(qty, step)` 那套 float64：0.3 / 0.1 在二进制浮点里是
// 2.9999999999999996，Floor 之后会少一个 step，下单数量直接少一档（07-web3-trading W-02 同款坑）。
// decimal 里 0.3 和 0.1 都是精确的十进制数，取模为 0，结果就是 0.3。
//
// step <= 0 时视为「无约束」，原样返回。
func AlignQtyDown(qty, step decimal.Decimal) decimal.Decimal {
	if !step.IsPositive() {
		return qty
	}
	return qty.Sub(qty.Mod(step))
}
