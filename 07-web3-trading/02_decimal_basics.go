//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 07-web3-trading && go run 02_decimal_basics.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/shopspring/decimal"
)

/*
═══════════════════════════════════════════════════════════════════════
W-02：decimal 基本用法
═══════════════════════════════════════════════════════════════════════

【核心概念】
    W-01 让你看到了 float64 的病，这一关给药。

    github.com/shopspring/decimal 用「一个大整数 + 一个十进制指数」表示数值：
        Decimal{ value: 12345, exp: -2 }  ==  123.45
    所以它存的是**十进制**数，0.1 就是精确的 0.1，不存在二进制转换误差。

    代价：运算慢（要走 big.Int）、内存占用大、除法必须指定精度。
    但在交易机器人里，这点开销换来的是「算出来的钱是对的」。

    对照其他语言：
        PHP    : bcmath（bcadd / bcmul，全程字符串）
        Java   : BigDecimal（必须用 new BigDecimal("0.1")，不能用 new BigDecimal(0.1)）
        Python : decimal.Decimal（必须用 Decimal("0.1")，不能用 Decimal(0.1)）
        Go     : shopspring/decimal（第三方，标准库没有）
    ⚠️ 注意上面三种语言的括号里都藏着一模一样的坑：**从字符串构造**。
       Go 的 NewFromFloat 就是 Java 的 new BigDecimal(0.1)、Python 的 Decimal(0.1)。

【你会看到】
    ✓ decimal 版 0.1+0.2 精确等于 0.3，Equal 判 true
    ✓ NewFromFloat 把 float64 的误差**原样搬进** decimal：0.1+0.7 → 0.7999999999999999
    ✓ 1/3 的默认精度是 16 位小数，想要别的精度得用 DivRound
    ✓ 1/0 会 panic（"decimal division by 0"）
    ✓ 1e18+1 在 decimal 里精确等于 1000000000000000001
    ✓ decimal.Decimal 是值类型（Add 不改原值），但**当 map key 会静默失效**

【思考题】（动手改代码验证，别空想）
    1. 示例 2 里 decimal.NewFromFloat(0.1).String() 打印出 "0.1"，看起来完全正常。
       请用 StringFixed(25) 再打一次，解释为什么它其实是准的，
       然后用 decimal.NewFromFloat(0.1+0.7)（**必须用变量**）再来一次。
       结论：NewFromFloat 什么时候安全、什么时候是陷阱？
    2. 把示例 4 里的 .Equal 换成 ==，观察哪些比较会静默出错。
       再打印两个 Decimal 的内部字段（Coefficient() / Exponent()），
       解释为什么 1.10 和 1.1 用 == 比是 false。
       （提示：Decimal 内部是 *big.Int，== 比的是指针。）
    3. 示例 8 展示了「decimal 当 map key」的静默失效。请设计一个正确的
       「按价格分档缓存」的 map，key 用什么？为什么用 String() 而不是 Float64()？
    4. 跨语言对照：Java 的 BigDecimal.divide() 不指定 scale 会抛
       ArithmeticException，Python 的 Decimal 除法用的是上下文精度。
       Go 的 decimal.Div 为什么选择「默认 16 位 + 不报错」？这样设计的风险是什么？
       （提示：静默截断 vs 显式报错，交易系统里你更想要哪个？）
    5. 用 decimal 重写 W-01 示例 5 的 stepSize 对齐（就是下面的 alignDownDecimal），
       确认 qty=0 的拒单问题消失了。然后想想：
       既然 decimal 这么准，为什么交易所的 API 文档还强调「数量要按 stepSize 对齐」？

【通关标准】
    能默写「从字符串构造 decimal」的正确姿势，
    知道比较必须用 Cmp/Equal、不能 ==，
    并且能说出 decimal 是值类型但内部持有指针这件事带来的两个后果。

【前置】W-01（float64 的坑）      【标签】shopspring/decimal · NewFromString · Cmp · 值类型
═══════════════════════════════════════════════════════════════════════
*/

// alignDownDecimal 用 decimal 把下单量向下对齐到交易所的 stepSize。
//
// 这是 W-01 里 alignDownFloat 的正确版本，思路一模一样（除以 stepSize → 向下取整 → 乘回去），
// 但因为每一步都是十进制定点运算，除法结果精确等于 1，floor 后就是 1，不会被截成 0。
//
// 注意 decimal 的 Div 默认只保留 16 位小数（decimal.DivisionPrecision），
// 这里刚好够用；真要严格控制精度应该用 DivRound(step, n) 显式指定。
func alignDownDecimal(qty, step decimal.Decimal) decimal.Decimal {
	if step.IsZero() {
		return qty
	}
	return qty.Div(step).Floor().Mul(step)
}

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 W-02 decimal 基本用法")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// ── 示例 1：先复现病，再上药 ─────────────────────────────────────────
	fmt.Println("📝 示例 1：同一个算式，float64 和 decimal 各跑一遍")
	fmt.Println(strings.Repeat("-", 60))
	f1, f2 := 0.1, 0.2
	fmt.Printf("  float64  0.1 + 0.2 = %.20f\n", f1+f2)
	fmt.Printf("  float64  == 0.3    ? %t\n", f1+f2 == 0.3)

	d1 := decimal.RequireFromString("0.1") // RequireFromString = 出错就 panic，适合常量
	d2 := decimal.RequireFromString("0.2")
	dSum := d1.Add(d2)
	fmt.Printf("  decimal  0.1 + 0.2 = %s\n", dSum.String())
	fmt.Printf("  decimal  Equal(0.3) ? %t\n", dSum.Equal(decimal.RequireFromString("0.3")))
	fmt.Println("  ➡️  药到病除。注意 RequireFromString 与 NewFromString 的区别：")
	fmt.Println("    NewFromString 返回 (Decimal, error)，RequireFromString 出错直接 panic")
	fmt.Println("    —— 常量用 Require 更省事，来自外部输入（JSON/API）的必须用 New 并检查 error")

	// ── 示例 2：NewFromString vs NewFromFloat（本关最大的坑） ────────────
	fmt.Println("📌 示例 2：NewFromString vs NewFromFloat —— 后者会把 float64 的误差搬进来")
	fmt.Println(strings.Repeat("-", 60))
	// 第一个观察：对「干净的字面量」，NewFromFloat 看起来完全正常
	fmt.Printf("  NewFromFloat(0.1).String()          = %s\n", decimal.NewFromFloat(0.1).String())
	fmt.Printf("  NewFromFloat(0.1).StringFixed(25)   = %s\n", decimal.NewFromFloat(0.1).StringFixed(25))
	fmt.Println("  → 看着没问题？因为 decimal 内部会对 float64 取「最短往返表示」，0.1 刚好能对上")

	// 第二个观察：一旦传进来的是**算出来的** float64，误差就被合法化了
	g1, g2 := 0.1, 0.7
	bad := decimal.NewFromFloat(g1 + g2)
	fmt.Printf("  float64 变量 0.1+0.7 = %.20f\n", g1+g2)
	fmt.Printf("  NewFromFloat(那个值) = %s   ← 误差原封不动进了 decimal\n", bad.String())
	fmt.Printf("  错误被放大：bad * 1e18 = %s wei\n", bad.Mul(decimal.NewFromInt(1000000000000000000)).String())
	fmt.Printf("  正确值            0.8 * 1e18 = %s wei（差 100 wei）\n",
		decimal.RequireFromString("0.8").Mul(decimal.NewFromInt(1000000000000000000)).String())
	fmt.Printf("  bad.Equal(0.8)       = %t\n", bad.Equal(decimal.RequireFromString("0.8")))
	fmt.Printf("  bad.Sub(0.8)         = %s   ← 差额是实实在在的\n", bad.Sub(decimal.RequireFromString("0.8")).String())
	fmt.Printf("  正确写法 RequireFromString(\"0.1\").Add(\"0.7\") = %s\n",
		decimal.RequireFromString("0.1").Add(decimal.RequireFromString("0.7")).String())
	fmt.Println("  ⚠️  这就是 Java 的 new BigDecimal(0.1)、Python 的 Decimal(0.1) 的同一个坑：")
	fmt.Println("      构造函数**不会**帮你修正浮点误差，它只是把误差原样接管过来")

	// 第三个观察：超出 float64 精度的大整数会被直接丢位
	fmt.Printf("\n  NewFromFloat(9007199254740993)  = %s   ← 丢了最后一位\n",
		decimal.NewFromFloat(9007199254740993).String())
	fmt.Printf("  NewFromInt(9007199254740993)    = %s   ← 整数走 NewFromInt，精确\n",
		decimal.NewFromInt(9007199254740993).String())
	fmt.Printf("  RequireFromString(\"9007199254740993\") = %s\n",
		decimal.RequireFromString("9007199254740993").String())
	fmt.Println("  口诀：**能拿字符串就拿字符串**。交易所的 REST/WS 返回的价格数量本来也是字符串")

	// 顺带提一下 NewFromFloatWithExponent：它按指定小数位转换，等于顺手做了舍入
	fmt.Printf("\n  NewFromFloatWithExponent(123.456, -2) = %s（保留 2 位小数）\n",
		decimal.NewFromFloatWithExponent(123.456, -2).String())
	fmt.Printf("  NewFromFloatWithExponent(123.456, 1)  = %s（保留到十位）\n",
		decimal.NewFromFloatWithExponent(123.456, 1).String())

	// ── 示例 3：四则运算 ────────────────────────────────────────────────
	fmt.Println("🔍 示例 3：加减乘除")
	fmt.Println(strings.Repeat("-", 60))
	p := decimal.RequireFromString("0.1")
	q := decimal.RequireFromString("0.2")
	fmt.Printf("  0.1 + 0.2 = %s\n", p.Add(q).String())
	fmt.Printf("  0.1 - 0.2 = %s\n", p.Sub(q).String())
	fmt.Printf("  0.1 * 0.2 = %s\n", p.Mul(q).String())
	fmt.Printf("  0.1 / 0.2 = %s\n", p.Div(q).String())
	fmt.Printf("  0.1^3     = %s   （Pow，指数是整数 Decimal）\n",
		p.Pow(decimal.NewFromInt(3)).String())
	fmt.Printf("  1.5 Shift(3)  = %s   （左移 3 位 = ×1000）\n",
		decimal.RequireFromString("1.5").Shift(3).String())
	fmt.Printf("  1.5 Shift(-3) = %s   （右移 3 位 = ÷1000）\n",
		decimal.RequireFromString("1.5").Shift(-3).String())

	// 除法要小心：十进制下 1/3 是无限循环小数，decimal 只能截断
	one := decimal.NewFromInt(1)
	three := decimal.NewFromInt(3)
	fmt.Printf("\n  1 / 3                = %s   ← 默认保留 %d 位\n", one.Div(three).String(), decimal.DivisionPrecision)
	fmt.Printf("  1 / 3 DivRound(3, 8) = %s   ← 显式指定 8 位\n", one.DivRound(three, 8).String())
	fmt.Println("  ⚠️  decimal 的除法默认是**截断**到 16 位，不报错。涉及除法的金额计算")
	fmt.Println("      一定要想清楚保留几位，宁可自己写 DivRound 也别吃默认值")

	// 除零是 panic，不是返回 error —— 生产代码里必须先判断
	fmt.Println("  1 / 0 → panic: decimal division by 0（decimal 用 panic 而不是 error 表达除零）")
	fmt.Printf("  安全的写法：先判 IsZero()，或者用 alignDownDecimal 里那种 if step.IsZero() 保护\n")

	// ── 示例 4：比较 —— 千万不要用 == ───────────────────────────────────
	fmt.Println("🧪 示例 4：比较：用 Cmp / Equal，不要用 ==")
	fmt.Println(strings.Repeat("-", 60))
	x := decimal.RequireFromString("1.10")
	y := decimal.RequireFromString("1.1")
	fmt.Printf("  x = 1.10, y = 1.1\n")
	fmt.Printf("  x == y          = %t   ← 假的！数值明明相等\n", x == y)
	fmt.Printf("  x.Equal(y)      = %t   ← 正确\n", x.Equal(y))
	fmt.Printf("  x.Cmp(y)        = %d   （-1 小于 / 0 相等 / 1 大于）\n", x.Cmp(y))
	fmt.Printf("  x.GreaterThan(y) = %t, x.LessThan(y) = %t\n", x.GreaterThan(y), x.LessThan(y))
	fmt.Printf("  内部真相：x.Coefficient()=%s exp=%d ; y.Coefficient()=%s exp=%d\n",
		x.Coefficient().String(), x.Exponent(), y.Coefficient().String(), y.Exponent())
	fmt.Println("  → 1.10 存成 {110, -2}，1.1 存成 {11, -1}，数值相等但内部表示不同")
	fmt.Println("  → 更要命的是 Decimal 内部的 value 是 *big.Int 指针，== 比的是**指针地址**，")
	fmt.Println("    所以连「同一个数、不同变量」都可能判 false。永远用 Equal / Cmp")

	// ── 示例 5：String() 与 StringFixed(n) ──────────────────────────────
	fmt.Println("⚙️ 示例 5：String() vs StringFixed(n)")
	fmt.Println(strings.Repeat("-", 60))
	for _, s := range []string{"1.5", "1.00000000", "0.00000001", "123456789.123456789"} {
		d := decimal.RequireFromString(s)
		fmt.Printf("  %-22s String()=%-22s StringFixed(8)=%-14s StringFixed(2)=%s\n",
			s, d.String(), d.StringFixed(8), d.StringFixed(2))
	}
	fmt.Println("  ✅ String() 是「最短表示」：末尾的 0 会被去掉，1.00000000 打印成 \"1\"")
	fmt.Println("  ✅ StringFixed(n) 强制 n 位小数，会**四舍五入**（第 4 行的 8 位就被进位了）")
	fmt.Println("  ✅ 发给交易所的数量/价格要用 StringFixed(精度)，否则 \"1\" 和 \"1.00000000\"")
	fmt.Println("    在有些交易所会触发格式校验失败；做幂等 key 时也会算出两个不同的字符串")

	// ── 示例 6：与 float64 / big.Int 互转 ───────────────────────────────
	fmt.Println("🎯 示例 6：与 float64 / big.Int 互转")
	fmt.Println(strings.Repeat("-", 60))
	// decimal → float64：会丢精度，第二个返回值告诉你是否精确
	for _, s := range []string{"0.1", "1e18", "0.1234567890123456789"} {
		d := decimal.RequireFromString(s)
		f, exact := d.Float64()
		fmt.Printf("  %-22s → float64=%-24.17g exact=%t\n", s, f, exact)
	}
	fmt.Println("  ➡️  exact=false 就说明这个转换**丢精度了**。把它当成编译器的 unused 警告：")
	fmt.Println("    能不用就不用；非要转（比如给只收 float64 的图表库）就把 exact 记进日志")

	// decimal ↔ big.Int：链上金额的必经之路
	fmt.Println()
	weiStr := "123456789012345678901234567890" // 一个 uint256 量级的余额
	wei, _ := new(big.Int).SetString(weiStr, 10)
	eth := decimal.NewFromBigInt(wei, -18) // 除以 10^18
	fmt.Printf("  wei  = %s\n", wei.String())
	fmt.Printf("  ETH  = %s   （NewFromBigInt(wei, -18) 就是除以 1e18）\n", eth.String())
	fmt.Printf("  反向 eth.BigInt() = %s（BigInt() 只取整数部分，会截断）\n", eth.BigInt().String())
	fmt.Printf("  整数构造 NewFromInt(100) = %s，New(15, -1) = %s\n",
		decimal.NewFromInt(100).String(), decimal.New(15, -1).String())

	// ── 示例 7：值类型语义 ──────────────────────────────────────────────
	fmt.Println("📦 示例 7：decimal.Decimal 是值类型：运算不改原值")
	fmt.Println(strings.Repeat("-", 60))
	orig := decimal.RequireFromString("100")
	plus := orig.Add(decimal.NewFromInt(1))
	fmt.Printf("  orig = %s, plus = %s   ← Add 返回新值，orig 纹丝不动\n", orig.String(), plus.String())
	fmt.Println("  这和 Go 的 int/string 一样，和 map/slice 的引用语义相反。")
	fmt.Println("  好处：不用防御性拷贝，传参天然安全（交易里这点很省心）")

	// ── 示例 8：值类型当 map key 的静默陷阱 ─────────────────────────────
	fmt.Println("🔗 示例 8：⚠️ decimal.Decimal 可以当 map key，但会静默失效")
	fmt.Println(strings.Repeat("-", 60))
	// 编译能过、不报错，但查不到 —— 这是本关最容易埋雷的地方
	cache := map[decimal.Decimal]string{}
	key := decimal.NewFromInt(1)
	cache[key] = "one"
	fmt.Printf("  用同一个变量查        : %q\n", cache[key])
	fmt.Printf("  用新构造的等价键查    : %q   ← 空！查不到\n", cache[decimal.NewFromInt(1)])
	fmt.Printf("  此时 len(cache) = %d\n", len(cache))
	cache[decimal.NewFromInt(1)] = "uno"
	fmt.Printf("  再用新键写一次后 len = %d   ← 同一个逻辑值占了两个槽位\n", len(cache))
	fmt.Println("  根因同示例 4：key 的比较用的是 ==，比的是内部 *big.Int 指针地址")
	fmt.Println("  正确做法：用 String() 当 key")

	byStr := map[string]string{}
	byStr[decimal.NewFromInt(1).String()] = "one"
	fmt.Printf("  用 String() 当 key 查 : %q   ← 命中\n", byStr["1"])
	fmt.Println("  （也别用 Float64() 当 key：0.1 和 0.10000000000000001 会变成两个 key）")

	// ── 示例 9：回到 W-01 的拒单场景 ────────────────────────────────────
	fmt.Println("🧭 示例 9：回到 W-01 的 stepSize 拒单场景，看 decimal 怎么解")
	fmt.Println(strings.Repeat("-", 60))
	price := decimal.RequireFromString("7")
	budget := decimal.RequireFromString("0.7")
	step := decimal.RequireFromString("0.1")
	qty := budget.Div(price)
	fmt.Printf("  价格=%s 预算=%s stepSize=%s\n", price.String(), budget.String(), step.String())
	fmt.Printf("  decimal qty      = %s   ← 精确的 0.1，没有 0.09999999999999999167\n", qty.String())
	fmt.Printf("  对齐 stepSize 后 = %s   ← 不再是 0，不会拒单\n", alignDownDecimal(qty, step).String())
	fmt.Println("  W-01 那个 0.3/0.1 的例子同理：decimal 里 Div 精确等于 3，floor 后就是 3")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ decimal = 大整数 + 十进制指数，存的**就是**十进制数，0.1 精确")
	fmt.Println("  ✅ **一律从字符串构造**：NewFromString / RequireFromString")
	fmt.Println("    NewFromFloat 会把 float64 的误差原样接管，且不报错（Java/Python 同款坑）")
	fmt.Println("  ✅ 比较永远用 Equal / Cmp / GreaterThan / LessThan，**不要用 ==**")
	fmt.Println("  ✅ String() 去末尾零，StringFixed(n) 补零并四舍五入；发给交易所用 StringFixed")
	fmt.Println("  ✅ 除法默认截断到 16 位且不报错，涉及金额要显式 DivRound；除零是 panic")
	fmt.Println("  ✅ 与 big.Int 互转：NewFromBigInt(wei, -18) 换算 ETH，BigInt() 只取整数部分")
	fmt.Println("  ✅ 是值类型，运算不改原值；但当 map key 会静默失效 —— 用 String() 当 key")
	fmt.Println("  ✅ 下一关 W-03 讲舍入模式（四舍五入 vs 银行家舍入）和按 stepSize 对齐的完整工程写法")
}
