//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 07-web3-trading && go run 09_order_state_machine.go
// 依赖：只用标准库
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"errors"
	"fmt"
	"strings"
)

/*
═══════════════════════════════════════════════════════════════════════
W-09：订单状态机
═══════════════════════════════════════════════════════════════════════

【核心概念】
    订单不是「一个数字」，而是一个有严格规矩的**状态机**。

    交易所会通过两条路反复告诉你订单的状态：REST 回执、WS 私有推送。
    这两条路的回调有三个特点，你必须先接受它们：
      · 会重复 —— 网络重试、服务端双写，同一笔成交被推两次很常见；
      · 会乱序 —— 断线重连补推时，旧状态可能后到；
      · 会迟到 —— 你断线 3 秒，重连后一口气收到 5 条历史推送。

    所以机器人**绝不能「收到什么就存什么」**。正确做法是自己维护状态机：
      · 只有合法跃迁才被接受（filled → submitted 永远不许发生）；
      · 重复事件按事件 ID 幂等丢弃，绝不改变状态；
      · 状态必须由**数量**推出来（remaining==0 才算 filled），
        不能无条件相信回执里那个 status 字符串；
      · 每次跃迁后校验不变量（filled + remaining == total）。

【你会看到】
    ✓ 一张打印出来的「合法跃迁表」，就是状态机的全部规则
    ✓ 完整序列：创建 → 提交 → 部分成交 30% → 重复事件（幂等丢弃）
      → 非法跃迁（拒绝）→ 成交 70% → filled
    ✓ filled → submitted、canceled → filled 被明确拒绝，状态纹丝不动
    ✓ 事件声明的状态和数量对不上时，以数量为准（真实交易所偶尔就这样）
    ✓ 人为把订单改坏（remaining 清零 / 状态硬改成 filled），
      CheckInvariants 立刻抓到 —— 这就是它在生产里的价值
    ✓ 全程整数运算（单位 1e-8），不碰 float64 —— 复习 W-01

【思考题】（动手改代码验证，别空想）
    1. 为什么 created → filled 必须被禁止，而不是「省事点直接允许」？
       如果允许了，你还能相信「filled 的订单一定是提交过的」吗？
    2. 示例 2 里我把「部分成交」和「完全成交」交给同一个分支处理，状态由
       remaining 算出来。如果交易所回执的 status 说 partial、成交量却是全额，
       你信谁？为什么？（提示：status 是「服务端当时的想法」，数量是「已发生的事实」）
    3. 幂等表现在用「事件 ID」去重。如果交易所回调根本不带唯一 ID，
       你还能拿什么当幂等键？（提示：(orderID, filledQty) 或 (orderID, tradeID)）
    4. rejected 现在是终态。但真实里「被拒 → 改价重试」很常见。
       这该建模成状态机的一部分，还是上层的重试策略？两者的边界在哪？
    5. 这台状态机只在内存里。进程重启后怎么恢复？
       （提示：把审计日志落库，启动时按顺序 replay；这就是「事件溯源」的雏形）
    6. 跨语言对照：Java 的 enum + switch、Python 的 transitions 库、
       PHP 的 Symfony Workflow —— Go 里为什么用 map[from]map[to]bool
       这种朴素写法就够？它少了什么、又多了什么？

【通关标准】
    能默写 map[from]map[to]bool 的状态机骨架，说清「为什么必须幂等」
    「为什么状态要从数量算」，并写出一个同时校验状态与数量的不变量函数。

【前置】L1-11（值/指针接收者）· W-01（金额用整数最小单位）
【标签】状态机 · 幂等 · 部分成交 · 不变量 · 审计日志
═══════════════════════════════════════════════════════════════════════
*/

// ─────────────────────────────────────────────────────────────────────────────
// 一、状态定义
// ─────────────────────────────────────────────────────────────────────────────

// OrderState 是订单状态。用 string 而不是 int，是为了：
//
//	· 日志、JSON、DB 里直接可读，不用查「3 是什么意思」；
//	· 跨语言/跨服务对接时不用维护枚举映射表。
//
// 代价是多几个字节、少一点点类型安全 —— 交易系统里可读性更重要。
type OrderState string

const (
	StateCreated   OrderState = "created"   // 本地已创建，还没提交给交易所
	StateSubmitted OrderState = "submitted" // 交易所已受理，等待成交
	StatePartial   OrderState = "partial"   // 部分成交，还有剩余在挂单
	StateFilled    OrderState = "filled"    // 全部成交（终态）
	StateCanceled  OrderState = "canceled"  // 已撤销（终态）
	StateRejected  OrderState = "rejected"  // 被交易所/风控拒绝（终态）
)

// unit 是金额/数量的最小单位：1e-8，也就是「聪」（satoshi）那一套思路。
// 1 BTC = 100_000_000 unit。全程用 int64 整数运算，永远不碰 float64。
// 为什么？见 W-01：float64 存不下 0.1 这种十进制小数，累加会漂。
const unit int64 = 100_000_000

// allStates 是**展示顺序**的固定切片。
// ⚠ 打印表格、遍历时必须用它，绝不能 `for k := range 某个map` ——
// Go 的 map 遍历顺序是随机的，同一份代码每次输出都不一样，
// 学习材料（以及任何要 diff 的日志/测试）里这是不能接受的。
var allStates = []OrderState{
	StateCreated, StateSubmitted, StatePartial,
	StateFilled, StateCanceled, StateRejected,
}

// allowedTransitions 就是这台状态机的**宪法**：from → 允许去到的 to 集合。
// 用 map[from]map[to]bool 而不是 map[from][]to，是为了 O(1) 判定，
// 顺便天然去重（切片里重复写一个目标状态，判定时就得多遍历一遍）。
//
// 设计要点：
//
//	· created 只能去 submitted / canceled / rejected
//	  → 「没提交就成交」在数据层就是不可能的状态；
//	· submitted → partial / filled / canceled / rejected
//	  → 直接全部成交（吃单）是合法的，所以 submitted → filled 必须允许；
//	· partial → partial（再来一笔部分成交）/ filled / canceled
//	  → partial → rejected 不合法：已经在挂的单不会被「拒绝」，只会被撤销；
//	· filled / canceled / rejected 是终态，出度为 0，进去就出不来。
var allowedTransitions = map[OrderState]map[OrderState]bool{
	StateCreated: {
		StateSubmitted: true,
		StateCanceled:  true,
		StateRejected:  true,
	},
	StateSubmitted: {
		StatePartial:  true,
		StateFilled:   true,
		StateCanceled: true,
		StateRejected: true,
	},
	StatePartial: {
		StatePartial:  true,
		StateFilled:   true,
		StateCanceled: true,
	},
	StateFilled:   {}, // 终态
	StateCanceled: {}, // 终态
	StateRejected: {}, // 终态
}

// canTransition 查表判定 from → to 是否合法。整个状态机只有这一个判定入口。
func canTransition(from, to OrderState) bool {
	return allowedTransitions[from][to]
}

// isTerminal 判断是不是终态（没有任何出边）。出度为 0 即为终态。
func isTerminal(s OrderState) bool {
	return len(allowedTransitions[s]) == 0
}

// 两个哨兵错误，方便调用方用 errors.Is 分类处理
// （比如：重复事件只记 debug 日志，非法跃迁要告警）。
var (
	ErrDuplicateEvent    = errors.New("重复事件")
	ErrIllegalTransition = errors.New("非法跃迁")
)

// ─────────────────────────────────────────────────────────────────────────────
// 二、订单与事件
// ─────────────────────────────────────────────────────────────────────────────

// Event 是「一个来自交易所（或本地）的状态变更请求」。
//
// ID 是幂等的关键：交易所回调通常会带一个唯一的事件/成交 ID，
// 重复推送时这个 ID 相同 —— 我们靠它把重复事件丢掉。
type Event struct {
	ID     string     // 事件唯一 ID，重复推送时相同
	To     OrderState // 事件声称要跃迁到的状态
	Fill   int64      // 本次成交量（单位 unit）；非成交类事件为 0
	Reason string     // 人话说明，写进审计日志
}

// Order 是一张订单。字段全部私有化思路：外部只能通过 Apply 改状态。
//
// ⚠ Filled / Remaining / State 三者必须始终自洽，改任何一个都要过 CheckInvariants。
type Order struct {
	ID     string
	Symbol string
	Side   string // BUY / SELL

	Total     int64 // 下单总量
	Filled    int64 // 累计成交量
	Remaining int64 // 剩余未成交量 = Total - Filled
	State     OrderState

	Audit []string // 审计日志：每一次「接受/拒绝」都留痕，生产里要落库

	seen map[string]bool // 已处理过的事件 ID —— 幂等表
}

// NewOrder 创建一张订单：初始状态 created、零成交、剩余 = 总量。
func NewOrder(id, symbol, side string, total int64) *Order {
	return &Order{
		ID:        id,
		Symbol:    symbol,
		Side:      side,
		Total:     total,
		Filled:    0,
		Remaining: total,
		State:     StateCreated,
		seen:      map[string]bool{},
	}
}

// Apply 是唯一的状态变更入口。它按固定顺序做四件事：
//
//	① 幂等：这个事件 ID 见过吗？见过就原样返回，绝不动状态；
//	② 合法性：canTransition(当前状态, 事件目标) 为 false 就拒绝；
//	③ 推进数量：成交类事件累加 Filled、递减 Remaining，
//	   **再由数量推出状态**（Remaining==0 才是 filled）；
//	④ 校验不变量：任何一步让订单变得自相矛盾，立即报错。
//
// 返回 nil 表示状态确实被改变了。
func (o *Order) Apply(ev Event) error {
	// ① 幂等去重：交易所会重复推送，同一个事件第二次到达必须无副作用。
	if o.seen[ev.ID] {
		o.log("⊘ 丢弃", ev, fmt.Sprintf("事件 ID 已处理过，状态保持 %s", o.State))
		return fmt.Errorf("%w：%s（订单 %s 保持 %s）", ErrDuplicateEvent, ev.ID, o.ID, o.State)
	}

	// ② 合法性：这是状态机的核心。非法跃迁连数量都不许碰。
	if !canTransition(o.State, ev.To) {
		o.log("✗ 拒绝", ev, fmt.Sprintf("%s 不允许跃迁到 %s", o.State, ev.To))
		return fmt.Errorf("%w：%s → %s（%s）",
			ErrIllegalTransition, o.State, ev.To, allowedText(o.State))
	}

	// ③ 推进状态与数量
	switch ev.To {
	case StatePartial, StateFilled:
		if ev.Fill <= 0 {
			o.log("✗ 拒绝", ev, "成交类事件必须带正数成交量")
			return fmt.Errorf("事件 %s：成交类事件 Fill 必须 > 0，实际 %d", ev.ID, ev.Fill)
		}
		if ev.Fill > o.Remaining {
			o.log("✗ 拒绝", ev, fmt.Sprintf("成交量 %s 超过剩余 %s", sat(ev.Fill), sat(o.Remaining)))
			return fmt.Errorf("事件 %s：成交量 %s 超过剩余 %s（超卖，说明上游数据有问题）",
				ev.ID, sat(ev.Fill), sat(o.Remaining))
		}
		o.Filled += ev.Fill
		o.Remaining -= ev.Fill

		// ★ 关键：状态由**数量**推出来，而不是照抄事件里的 To。
		// 真实交易所偶尔会给出自相矛盾的回执（status=partial 但数量已满），
		// 此时「已发生的数量」才是事实，「状态字符串」只是服务端当时的想法。
		next := StatePartial
		if o.Remaining == 0 {
			next = StateFilled
		}
		if ev.To != next {
			o.log("⚠ 修正", ev, fmt.Sprintf("事件声明 %s，按数量算应为 %s —— 以数量为准", ev.To, next))
		}
		o.State = next

	case StateCanceled:
		// 撤单：状态转 canceled，但 Remaining **原样保留**。
		// 它是「撤单时还剩多少」的答案，上层靠它决定要不要重新挂单。
		o.State = StateCanceled

	case StateRejected:
		// 被拒：一单未成。若已有部分成交却收到 rejected，下面的不变量会抓到。
		o.State = StateRejected

	case StateSubmitted:
		o.State = StateSubmitted

	default:
		o.log("✗ 拒绝", ev, "未知目标状态")
		return fmt.Errorf("事件 %s：未知目标状态 %q", ev.ID, ev.To)
	}

	o.seen[ev.ID] = true // 只有真正被接受的事件才记入幂等表
	o.log("✓ 接受", ev, fmt.Sprintf("→ %s", o.State))

	// ④ 不变量校验：宁可在这一步 panic/告警，也不能让坏状态流到下游。
	if err := o.CheckInvariants(); err != nil {
		return fmt.Errorf("事件 %s 应用后不变量被破坏：%w", ev.ID, err)
	}
	return nil
}

// CheckInvariants 校验「状态 + 数量」是否自洽。任何一条不满足都返回错误。
//
// 生产用法：每次 Apply 之后跑一遍（本文件就是这么做的），
// 另外再加一个定时任务对全量订单巡检 —— 因为 bug 也可能来自
// 「绕过 Apply 直接写库」的代码路径。
func (o *Order) CheckInvariants() error {
	if o.Filled < 0 {
		return fmt.Errorf("不变量#1 成交量不能为负：filled=%d", o.Filled)
	}
	if o.Remaining < 0 {
		return fmt.Errorf("不变量#2 剩余量不能为负：remaining=%d", o.Remaining)
	}
	if o.Filled+o.Remaining != o.Total {
		return fmt.Errorf("不变量#3 filled+remaining 必须等于 total：%s + %s != %s",
			sat(o.Filled), sat(o.Remaining), sat(o.Total))
	}

	switch o.State {
	case StateCreated, StateSubmitted, StateRejected:
		if o.Filled != 0 {
			return fmt.Errorf("不变量#4 %s 状态不该有成交量：filled=%s", o.State, sat(o.Filled))
		}
	case StatePartial:
		if o.Filled == 0 || o.Remaining == 0 {
			return fmt.Errorf("不变量#5 partial 必须两头都有量：filled=%s remaining=%s",
				sat(o.Filled), sat(o.Remaining))
		}
	case StateFilled:
		if o.Filled != o.Total || o.Remaining != 0 {
			return fmt.Errorf("不变量#6 filled 必须是全额成交：filled=%s total=%s remaining=%s",
				sat(o.Filled), sat(o.Total), sat(o.Remaining))
		}
	case StateCanceled:
		// 撤单允许「部分成交后撤销」：0 ≤ filled < total，remaining 就是没成交的那部分。
		// 这里没有额外约束，上面的 #1~#3 已经足够。
	default:
		return fmt.Errorf("不变量#7 未知状态：%q", o.State)
	}
	return nil
}

// log 往审计日志里追加一行。真实系统里这行要落库（订单事件表），
// 它是「重启后 replay 恢复状态」和「事后对账」的唯一依据。
func (o *Order) log(tag string, ev Event, detail string) {
	o.Audit = append(o.Audit, fmt.Sprintf("%s %-5s %-9s fill=%-8s %s",
		tag, ev.ID, ev.To, sat(ev.Fill), detail))
}

// ─────────────────────────────────────────────────────────────────────────────
// 三、展示用的小工具
// ─────────────────────────────────────────────────────────────────────────────

// sat 把「最小单位整数」格式化成人类可读的十进制字符串：
//
//	100_000_000 → "1"        30_000_000 → "0.3"
//	1           → "0.00000001"     0       → "0"
//
// ⚠ 全程整数除法/取余，不经过 float64 —— 这正是 W-01 教的写法。
func sat(v int64) string {
	neg := v < 0
	if neg {
		v = -v
	}
	s := fmt.Sprintf("%d.%08d", v/unit, v%unit)
	s = strings.TrimRight(s, "0")
	s = strings.TrimSuffix(s, ".")
	if neg {
		s = "-" + s
	}
	return s
}

// targetsOf 返回某个状态允许去的所有目标状态（按 allStates 固定顺序，输出稳定）。
func targetsOf(from OrderState) []string {
	out := []string{}
	for _, to := range allStates {
		if canTransition(from, to) {
			out = append(out, string(to))
		}
	}
	return out
}

// allowedText 把「某个状态允许去哪」说成一句人话，用在错误信息里。
// 终态会返回「终态，不允许再变」，避免打印出「只能去 」这种空列表。
func allowedText(from OrderState) string {
	targets := targetsOf(from)
	if len(targets) == 0 {
		return "终态，不允许再变"
	}
	return "只能去 " + strings.Join(targets, " / ")
}

// invariantText 把不变量校验的结果打成人话：nil 也要看得懂。
func invariantText(err error) string {
	if err == nil {
		return "nil（状态与数量完全自洽）"
	}
	return err.Error()
}

// snapshot 是订单的一行摘要，打印和日志都用它。
func (o *Order) snapshot() string {
	return fmt.Sprintf("state=%-9s filled=%-9s remaining=%-9s total=%-6s",
		o.State, sat(o.Filled), sat(o.Remaining), sat(o.Total))
}

// printTransitionTable 打印合法跃迁表 —— 这张表就是状态机的全部规则。
func printTransitionTable() {
	fmt.Println("  状态         → 允许跃迁到")
	fmt.Println("  ------------ → --------------------------------------------------")
	for _, from := range allStates {
		targets := targetsOf(from)
		if len(targets) == 0 {
			fmt.Printf("  %-12s → （终态，不允许再变）\n", from)
			continue
		}
		fmt.Printf("  %-12s → %s\n", from, strings.Join(targets, ", "))
	}
}

// applyAndPrint 执行一个事件并把结果打成人话。main 里所有示例都走它。
func applyAndPrint(o *Order, ev Event) {
	err := o.Apply(ev)
	switch {
	case err == nil:
		fmt.Printf("  ✓ 事件 %-5s 被接受：%s\n", ev.ID, o.snapshot())
		fmt.Printf("      %s\n", ev.Reason)
	case errors.Is(err, ErrDuplicateEvent):
		fmt.Printf("  ⊘ 事件 %-5s 幂等丢弃：%v\n", ev.ID, err)
		fmt.Printf("      状态保持 %s ← 重复投递没有改变任何东西\n", o.snapshot())
	case errors.Is(err, ErrIllegalTransition):
		fmt.Printf("  ✗ 事件 %-5s 非法跃迁被拒：%v\n", ev.ID, err)
		fmt.Printf("      状态保持 %s ← 非法跃迁连数量都不许碰\n", o.snapshot())
	default:
		fmt.Printf("  ✗ 事件 %-5s 被拒：%v\n", ev.ID, err)
		fmt.Printf("      状态保持 %s\n", o.snapshot())
	}
}

// printAudit 打印审计日志：每一次接受/拒绝都留痕。
func printAudit(o *Order) {
	fmt.Printf("  【审计日志】订单 %s（%s %s）\n", o.ID, o.Symbol, o.Side)
	for i, line := range o.Audit {
		fmt.Printf("    %2d) %s\n", i+1, line)
	}
	fmt.Printf("  审计日志 %d 条 —— 真实系统里它要落库，重启后靠它 replay 恢复状态\n", len(o.Audit))
}

func main() {
	fmt.Println("\n=== W-09 订单状态机 ===")

	// ── 示例 1：合法跃迁表 ───────────────────────────────────────────────
	fmt.Println("\n【示例 1】合法跃迁表：状态机的「宪法」")
	printTransitionTable()
	fmt.Println("  这张表就是全部规则，canTransition(from, to) 只查它，没有第二处判断。")
	fmt.Println("  ⚠ created 只能去 submitted / canceled / rejected ——")
	fmt.Println("    也就是说「没提交就成交」（created → filled）在数据层就不可能发生。")

	// ── 示例 2：一条真实订单的完整生命周期 ──────────────────────────────
	fmt.Println("\n【示例 2】完整序列：创建 → 提交 → 部分成交 30% → 重复事件 → 非法跃迁 → 成交 70% → filled")
	o := NewOrder("ORD-1001", "BTCUSDT", "BUY", 1*unit) // 1 BTC = 1e8 unit
	fmt.Printf("  创建订单 %s：%s\n", o.ID, o.snapshot())
	fmt.Printf("  CheckInvariants() → %s\n", invariantText(o.CheckInvariants()))

	// 提交：本地 created → 交易所 submitted
	applyAndPrint(o, Event{ID: "E-1", To: StateSubmitted, Reason: "POST /order 返回 orderId=998877"})

	// 第一次部分成交 30%：remaining 还有 0.7，所以状态是 partial 而不是 filled
	applyAndPrint(o, Event{
		ID: "E-2", To: StatePartial, Fill: 30 * unit / 100,
		Reason: "WS 私有推送：成交 0.3 BTC，剩余 0.7 仍在挂单",
	})

	// 交易所把同一个事件又推了一遍（网络重试 / 服务端双写），必须幂等
	applyAndPrint(o, Event{
		ID: "E-2", To: StatePartial, Fill: 30 * unit / 100,
		Reason: "同一条推送重复到达（真实里非常常见）",
	})

	// 非法跃迁：partial 不可能回到 submitted
	applyAndPrint(o, Event{ID: "E-X", To: StateSubmitted, Reason: "乱序到达的旧回执"})

	// 剩余 70% 成交：数量凑满总量，状态自动落到 filled
	applyAndPrint(o, Event{
		ID: "E-3", To: StateFilled, Fill: 70 * unit / 100,
		Reason: "WS 私有推送：剩余 0.7 BTC 全部成交",
	})

	// filled 是终态，任何后续事件都必须被拒
	applyAndPrint(o, Event{ID: "E-4", To: StateSubmitted, Reason: "迟到的旧状态推送"})
	applyAndPrint(o, Event{ID: "E-5", To: StatePartial, Fill: 30 * unit / 100, Reason: "重复的成交推送（换了新 ID，但状态已是终态）"})

	printAudit(o)
	fmt.Printf("  最终不变量校验：%s\n", invariantText(o.CheckInvariants()))

	// ── 示例 3：撤单的订单不能再成交 ────────────────────────────────────
	fmt.Println("\n【示例 3】撤单后剩余量作废，canceled → filled 必须被拒")
	o2 := NewOrder("ORD-1002", "ETHUSDT", "SELL", 50*unit/100) // 0.5 ETH
	fmt.Printf("  创建订单 %s：%s\n", o2.ID, o2.snapshot())
	applyAndPrint(o2, Event{ID: "F-1", To: StateSubmitted, Reason: "POST /order 已受理"})
	applyAndPrint(o2, Event{ID: "F-2", To: StatePartial, Fill: 20 * unit / 100, Reason: "成交 0.2 ETH（40%）"})
	applyAndPrint(o2, Event{ID: "F-3", To: StateCanceled, Reason: "用户撤单：剩余 0.3 立即作废"})
	applyAndPrint(o2, Event{ID: "F-4", To: StateFilled, Fill: 30 * unit / 100, Reason: "交易所延迟推送的成交回报"})
	fmt.Println("  ↑ 撤单后 Remaining 仍然是 0.3 —— 那是「撤单时还剩多少」，上层靠它决定要不要重新挂单；")
	fmt.Println("    但状态已经是 canceled，任何想把它推回 filled 的回调都会被拒绝。")
	fmt.Println("  ↑ 顺带一提：o2 是 *Order（指针），所以 Apply 改的是同一张订单；")
	fmt.Println("    如果 Apply 定义在值接收者上，改的就只是副本了 —— 复习 L1-11。")

	// 顺便验证 created 直连 filled 也不行
	o3 := NewOrder("ORD-1003", "SOLUSDT", "BUY", 10*unit)
	fmt.Println("\n  再验证 created 直接成交：")
	applyAndPrint(o3, Event{ID: "G-1", To: StateFilled, Fill: 10 * unit, Reason: "凭空出现的成交回报"})

	// ── 示例 4：不变量校验 ──────────────────────────────────────────────
	fmt.Println("\n【示例 4】不变量校验：状态和数量必须自洽")
	o4 := NewOrder("ORD-1004", "SOLUSDT", "BUY", 10*unit)
	applyAndPrint(o4, Event{ID: "H-1", To: StateSubmitted, Reason: "提交"})
	applyAndPrint(o4, Event{ID: "H-2", To: StatePartial, Fill: 4 * unit, Reason: "成交 4 个"})
	fmt.Printf("  正常订单：%s\n", o4.snapshot())
	fmt.Printf("  CheckInvariants() → %s\n", invariantText(o4.CheckInvariants()))

	// ① 破坏 #5：数量守恒仍然成立（4+0==4），但 partial 状态却已经没有剩余量了。
	//    这种「守恒但状态与数量矛盾」的坏数据，光靠对账是发现不了的。
	o4.Remaining, o4.Total = 0, 4*unit
	fmt.Println("\n  ① 把 Remaining 清零、Total 同步改成 4（守恒仍成立），状态却还是 partial：")
	fmt.Printf("     %s\n", o4.snapshot())
	fmt.Printf("     CheckInvariants() → %s\n", invariantText(o4.CheckInvariants()))

	// ② 破坏 #6：状态硬改成 filled，但数量根本没凑满。
	o4.Remaining, o4.Total, o4.State = 6*unit, 10*unit, StateFilled
	fmt.Println("\n  ② 恢复数量，再把状态硬改成 filled（数量其实没满）：")
	fmt.Printf("     %s\n", o4.snapshot())
	fmt.Printf("     CheckInvariants() → %s\n", invariantText(o4.CheckInvariants()))

	// ③ 破坏 #3：数量守恒被打破（filled + remaining != total）。
	o4.State, o4.Total = StatePartial, 12*unit
	fmt.Println("\n  ③ 恢复状态，只把 Total 改成 12（filled+remaining 就对不上 total 了）：")
	fmt.Printf("     %s\n", o4.snapshot())
	fmt.Printf("     CheckInvariants() → %s\n", invariantText(o4.CheckInvariants()))
	fmt.Println("  三种坏状态全被抓到 —— 这就是「不变量函数」在生产里的价值：")
	fmt.Println("  它不修数据，它只是让坏数据在**流到下游之前**大声报错。")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n【关键总结】")
	fmt.Println("  • 状态机 = 一张 map[from]map[to]bool 允许表 + 一个 Apply 方法，别的都是细节")
	fmt.Println("  • 幂等：交易所回调会重复推送，用事件 ID 去重，重复事件绝不改变状态")
	fmt.Println("  • 部分成交：Filled 累加、Remaining 递减，Remaining==0 才转 filled")
	fmt.Println("  • 状态要从数量算出来，别无条件相信回执里的 status 字符串")
	fmt.Println("  • 每次跃迁后跑 CheckInvariants：filled+remaining==total，且状态与数量自洽")
	fmt.Println("  • 金额/数量用整数最小单位（1e-8），不碰 float64 —— 见 W-01")
	fmt.Println("  • 审计日志要落库：它是重启后 replay 恢复状态、以及事后对账的唯一依据")
	fmt.Println("  • 该用：订单 / 仓位 / 任务等「有明确生命周期」的实体")
	fmt.Println("  • 不该用：没有生命周期约束的普通数据 —— 纯 CRUD 就够，别硬套状态机")
}
