/*
═══════════════════════════════════════════════════════════════════════
T-02：子测试与并行（t.Run / t.Parallel / t.Cleanup）
═══════════════════════════════════════════════════════════════════════

【核心概念】

	T-01 用一张表 + 一个 for 循环跑完所有用例。问题是：**整张表只是一个测试**。
	想只跑其中一条、想给每条独立超时、想让它们并行 —— 都做不到。

	子测试（subtest）解决这个：把循环体包进 t.Run(name, func(t *testing.T){...})。

	    for _, tt := range tests {
	        t.Run(tt.name, func(t *testing.T) {   // ← 每个用例变成一个独立的测试
	            ...
	        })
	    }

	得到三个能力：
	  ① 选择性执行：go test -run 'TestCheck/名义金额' 只跑匹配的子测试（支持正则）；
	  ② 独立生命周期：某个子测试里 t.Fatalf 只终止它自己，不会连坐其他用例；
	  ③ 并行：子测试里调 t.Parallel()，go test 会并发跑它们。

	三个必须记住的细节：
	  · 循环变量捕获：Go 1.22 起 for-range 变量是**每次迭代新建**的，
	    所以 t.Run 的闭包可以安全地直接用 tt，不需要再写 tt := tt。
	    （老代码里那句 tt := tt 是给 1.21 及以前打的补丁，现在已经多余。）
	  · t.Parallel() 会**暂停**当前子测试，等所有兄弟子测试都启动后再一起跑，
	    所以父测试里 t.Run 返回之后不能立刻断言「子测试已经跑完」。
	  · t.Cleanup(fn) 注册的清理函数在**当前测试结束**时逆序执行，
	    比 defer 更好用：写在 helper 里的 Cleanup 会在调用方的测试结束时才跑，
	    而 helper 里的 defer 会在 helper 返回时就跑掉。

【运行方式】

	cd 06-go-testing && go test -v ./02_subtests/

	一定要亲手试这几个：
	    go test -v -run 'TestCheck/名义金额' ./02_subtests/   # 只跑一条子测试
	    go test -v -run 'TestCheck/(名义|数量)' ./02_subtests/ # 正则匹配多条
	    go test -v ./02_subtests/                             # 看子测试的并行交错
	    go test -v -parallel 1 ./02_subtests/                 # 串行跑，对比输出顺序

【你会看到】

	✓ 每个子测试单独一行 PASS/FAIL，缩进显示在父测试下面
	✓ t.Parallel() 让「互不影响的用例」交错执行，父测试结束时才汇总
	✓ t.Cleanup 逆序执行，且「注册在 helper 里、执行在调用方结束时」
	✓ t.Setenv 自动隔离环境变量，测试之间不会互相污染

【思考题】（动手改代码验证，别空想）

 1. 给 TestCheck 里的每个子测试加上 t.Parallel()，用 -v 跑两次。
    输出里子测试的「=== RUN / --- PASS」顺序还是固定的吗？
    然后加 -parallel 1 再跑一次，顺序变回固定了吗？
    对照 PHPUnit：为什么 PHPUnit 的测试默认是串行的、且很难并行？
    （提示：共享的 static 状态和全局数据库连接。）

 2. 把 TestCheck 里 t.Run 的闭包改成捕获外层变量（比如在循环外定义
    `var got error` 然后闭包里赋值），跑 go test -race ./02_subtests/。
    会报 data race 吗？为什么？—— 这解释了「为什么每个子测试要有自己的局部变量」。

 3. 删掉 newGuard 里的 t.Cleanup，改成 t.Cleanup(func(){ log... }) 以外什么都不做，
    观察输出顺序。然后把同一段逻辑改成 defer，再观察一次：
    helper 里的 defer 是在 helper 返回时执行，还是在整个测试结束时执行？
    这两种语义分别在什么场景下是你想要的？

 4. 表驱动 + 子测试组合起来，代码比 T-01 长了大约 3 行。
    想清楚：什么情况下值得用子测试，什么情况下 T-01 的裸循环就够了？
    提示：需要「选择性执行」或「独立并行」时才值得。

【通关标准】

	能默写「表驱动 + t.Run 子测试」的组合骨架，
	能说清 t.Cleanup 和 defer 在 helper 里的语义差异，
	并且知道 Go 1.22 之后不再需要写 tt := tt。

【前置】T-01（表驱动测试）   【标签】testing · 子测试 · t.Parallel · t.Cleanup
═══════════════════════════════════════════════════════════════════════
*/
package risk

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ─────────────────────────────────────────────────────────────────────────────
// 本关的被测代码：交易机器人的风控前置校验。
// 每一关都围绕真实零件出题 —— 这是机器人里「下单前必须过的那道闸」。
// ─────────────────────────────────────────────────────────────────────────────

// Order 是一笔待下单的委托。
type Order struct {
	Symbol   string
	Qty      float64
	Notional float64 // 名义金额 = 价格 × 数量
}

// Guard 是风控规则集合。
//
// 用值接收者而不是指针接收者：Guard 是只读配置，值语义更安全
// （对照 T-01 里 Side 的 String 方法，同样是只读）。
type Guard struct {
	MaxNotional float64         // 单笔名义金额上限
	MinQty      float64         // 最小下单量
	Allowed     map[string]bool // 白名单交易对
}

// Check 按顺序校验一笔委托，返回第一个不通过的规则。
//
// ⚠ 本函数故意留了一个「可优化点」（思考题 4 会让你处理）：
//
//	错误信息里没有带上实际值。排障时只看到「超过单笔上限」，
//	却不知道到底下了多少 —— 这在真实事故复盘里是灾难。
func (g Guard) Check(o Order) error {
	if !g.Allowed[o.Symbol] {
		return fmt.Errorf("交易对不在白名单")
	}
	if o.Qty < g.MinQty {
		return fmt.Errorf("下单量低于最小限制")
	}
	if o.Notional > g.MaxNotional {
		return fmt.Errorf("超过单笔名义金额上限")
	}
	return nil
}

// GuardFromEnv 从环境变量构造 Guard。
//
// 环境变量前缀 RISK_，逗号分隔的白名单。测试里用 t.Setenv 设置，
// 每个测试自动隔离 —— 不需要 PHP 的 putenv/putenv 还原、也不需要 Java 的
// System.setProperty + @AfterEach 手动清理。
func GuardFromEnv() (Guard, error) {
	g := Guard{
		MaxNotional: 10000,
		MinQty:      0.001,
		Allowed:     map[string]bool{"BTCUSDT": true},
	}

	if v := os.Getenv("RISK_MAX_NOTIONAL"); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return g, fmt.Errorf("RISK_MAX_NOTIONAL=%q 不是合法数字", v)
		}
		g.MaxNotional = f
	}

	if v := os.Getenv("RISK_ALLOWED"); v != "" {
		allowed := map[string]bool{}
		for _, s := range strings.Split(v, ",") {
			if s = strings.TrimSpace(s); s != "" {
				allowed[s] = true
			}
		}
		g.Allowed = allowed
	}

	return g, nil
}
