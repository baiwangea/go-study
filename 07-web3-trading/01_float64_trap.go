//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 07-web3-trading && go run 01_float64_trap.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"math"
	"strings"
)

/*
═══════════════════════════════════════════════════════════════════════
W-01：float64 为什么不能表示钱
═══════════════════════════════════════════════════════════════════════

【核心概念】
    float64 是 IEEE-754 双精度二进制浮点：1 位符号 + 11 位指数 + 52 位尾数。
    它用「二进制小数」存数，而 0.1 / 0.2 / 0.3 这些十进制小数在二进制里
    是**无限循环小数**（就像 1/3 在十进制里是 0.333…），52 位尾数只能截断存放。
    于是「写进去的数」和「你以为的数」从一开始就不是同一个数。

    这不是 Go 的 bug，PHP / Java / Python / JS 的 double 全都一样。
    但交易机器人对这一点零容忍：下单量、价格、手续费、余额，
    任何一处算错都可能是真金白银。

【你会看到】
    ✓ 变量相加 0.1+0.2 打印出 0.30000000000000004，且 == 0.3 为 false
    ✓ 0.1+0.2 与 0.3 的二进制只差**最后 1 位**（1 个 ULP）
    ✓ 循环加 0.1 一千次得 99.99999999999859312538，而不是 100
    ✓ 循环加 0.04 一万次得 400.000000000057014，而不是 400
    ✓ 1e18 + 1 == 1e18 为 true（大数处 ULP 已经涨到 128）
    ✓ 价格 7 / 预算 0.7 / stepSize 0.1 → 对齐后下单量变成 0 → 直接被交易所拒单

【思考题】（动手改代码验证，别空想）
    1. 把示例 1 里的 0.1 / 0.2 换成 0.5 / 0.25（二进制能精确表示的数），
       0.5+0.25 == 0.75 是 true 还是 false？这说明了「哪些小数是安全的」？
    2. **本关最反直觉的一题**：直接写 fmt.Println(0.1+0.2 == 0.3) 居然打印 true！
       为什么把它拆成 a, b := 0.1, 0.2; a+b == 0.3 就变 false 了？
       提示：Go 的**无类型常量**是任意精度的，编译期就折叠算完了；
       只有落到 float64 变量上才降级成二进制浮点。
       （对照：Java 的 double 常量折叠用 IEEE 双精度，所以 Java 里是 false。）
    3. 累加误差为什么有时偏大有时偏小？把示例 3 的循环次数从 1000 改成
       100、10000、100000，观察误差是随机游走还是单调累积。
       这对「用 float64 累计手续费」意味着什么？
    4. 跨语言对照：PHP 的 bcmath、Java 的 BigDecimal、Python 的 decimal
       都是十进制定点/任意精度库。Go 标准库**没有**等价物 ——
       为什么 Go 选择不内置？（提示：标准库的极简主义 + 第三方库生态）
    5. 示例 5 里我用 math.Floor(qty/step)*step 做对齐。请解释：
       为什么加一个 1e-9 的「容差」能临时救回来？这种做法有什么隐患？
       （提示：容差该给多大取决于 step 的量级，而 step 是交易所动态下发的。）

【通关标准】
    能说清「float64 存的是二进制小数、十进制小数多半存不准」，
    能手算 0.1+0.2 != 0.3 的原因，
    并知道交易机器人里**任何金额/数量都不该用 float64 累加或比较**。

【前置】L1-01（基础类型）      【标签】IEEE-754 · 精度丢失 · stepSize · 拒单
═══════════════════════════════════════════════════════════════════════
*/

// alignDownFloat 是**反面教材**：用 float64 把下单量向下对齐到交易所的 stepSize。
//
// 交易所要求下单量必须是 stepSize 的整数倍（比如 stepSize=0.1，只能下
// 0.1 / 0.2 / 0.3 …）。这段代码的思路是「先除以 stepSize 取整，再乘回去」，
// 思路本身没错，但 float64 的除法会先引入误差，floor 就会往错的方向取整。
//
// 示例 5 会让你亲眼看到它把 0.1 的下单量算成 0 —— 也就是「下单量小于最小交易量」。
func alignDownFloat(qty, step float64) float64 {
	return math.Floor(qty/step) * step
}

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 W-01 float64 为什么不能表示钱")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// ── 示例 1：0.1 + 0.2 不等于 0.3 ─────────────────────────────────────
	fmt.Println("📝 示例 1：0.1 + 0.2 到底等于多少")
	fmt.Println(strings.Repeat("-", 60))
	// 注意：这里**必须**先赋值给变量，不能直接写字面量 0.1 + 0.2。
	// 原因见示例 1 末尾的「常量折叠」小实验。
	a, b := 0.1, 0.2
	sum := a + b
	fmt.Printf("  0.1 + 0.2        = %.20f\n", sum)
	fmt.Printf("  %%v 打印         = %v\n", sum)
	fmt.Printf("  字面量 0.3       = %.20f\n", 0.3)
	fmt.Printf("  sum == 0.3       = %t   ← 假的！\n", sum == 0.3)
	fmt.Printf("  sum - 0.3        = %.20f\n", sum-0.3)
	fmt.Println("  多出来的 4 在末位，肉眼几乎看不见，但 == 会毫不留情地判 false")

	// 一个更刺眼的例子：这个数字会直接出现在你的日志里
	x, y := 0.1, 0.7
	fmt.Printf("  换个组合 0.1+0.7 = %.20f（%v）← 日志里会印成 0.7999999999999999\n", x+y, x+y)

	// 【Go 特有的坑】无类型常量是任意精度的，编译期折叠算完才降级成 float64。
	// 所以直接写字面量时结果反而是「对」的 —— 这会让人误以为 Go 没问题。
	fmt.Printf("  常量 0.1+0.2     = %.20f   ← 编译期精确计算后取最近的 float64\n", 0.1+0.2)
	fmt.Printf("  常量 0.1+0.2 == 0.3 → %t   ← 居然是 true！\n", 0.1+0.2 == 0.3)
	fmt.Println("  ⚠️  真实代码里数字都来自变量/JSON/交易所，永远走不到这条「精确」的路")

	// ── 示例 2：为什么？看看内存里的 64 个 bit ───────────────────────────
	fmt.Println("📌 示例 2：为什么存不准：二进制小数是无限循环的")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("  十进制里 1/3 = 0.3333…（3 无限循环，写不完）")
	fmt.Println("  二进制里 1/10 也一样：0.1 = 0.0001100110011…（1100 无限循环，写不完）")
	fmt.Println("  float64 只有 52 位尾数，只能把后面截掉 —— 于是从第一步就错了")

	// 把三个数的 64 个 bit 摊开看：1 位符号 + 11 位指数 + 52 位尾数
	for _, v := range []struct {
		name string
		val  float64
	}{{"0.1", 0.1}, {"0.2", 0.2}, {"0.3", 0.3}} {
		bits := math.Float64bits(v.val)
		sign := bits >> 63
		exp := (bits >> 52) & 0x7ff // 11 位指数（偏移量 1023）
		frac := bits & ((1 << 52) - 1)
		fmt.Printf("  %-4s %016x  符号=%d 指数=%d(2^%d) 尾数=%013x\n",
			v.name, bits, sign, exp, int(exp)-1023, frac)
		fmt.Printf("       二进制 = %064b\n", bits)
	}
	fmt.Println("  ➡️  尾数那一串 100110011001100… 就是被截断的无限循环")

	// 决定性的一击：0.1+0.2 与 0.3 只差最后 1 位
	bitsSum := math.Float64bits(sum)
	bits03 := math.Float64bits(0.3)
	fmt.Printf("\n  0.1+0.2 的 bits = %064b\n", bitsSum)
	fmt.Printf("  0.3     的 bits = %064b\n", bits03)
	fmt.Printf("  两者的差值      = %d 个 ULP（Unit in the Last Place，最后一位的权重）\n", bitsSum-bits03)
	fmt.Println("  就这么 1 位的差距，让 == 从 true 变成 false —— 浮点比较的脆弱性全在这里")

	// ── 示例 3：累加误差 ────────────────────────────────────────────────
	fmt.Println("🔍 示例 3：累加误差：每笔都「差不多」，一万笔就差出来了")
	fmt.Println(strings.Repeat("-", 60))
	step := 0.1
	var loopSum float64
	for i := 0; i < 1000; i++ {
		loopSum += step // 每次都在「已经不准的数」上再加，误差会持续累积
	}
	fmt.Printf("  循环加 0.1 一千次  = %.20f\n", loopSum)
	fmt.Printf("  直接 0.1 * 1000    = %.20f\n", step*1000)
	fmt.Printf("  误差               = %.20f  (%.3e)\n", loopSum-100, loopSum-100)
	fmt.Printf("  1000 次累加 == 100 ? %t\n", loopSum == 100)
	fmt.Println("  有意思的是 0.1*1000 恰好 == 100（乘法只错一次，还错对了方向）")
	fmt.Println("  → 这就是为什么「累加」比「乘法」危险得多：误差会一直叠上去")

	// 换成手续费的真实数字：每笔 0.04 USDT，一万笔
	fee := 0.04
	var feeSum float64
	for i := 0; i < 10000; i++ {
		feeSum += fee
	}
	fmt.Printf("\n  循环加 0.04 一万次 = %.15f\n", feeSum)
	fmt.Printf("  期望 0.04 * 10000  = %.15f\n", fee*10000)
	fmt.Printf("  误差               = %.15f  (%.3e)\n", feeSum-fee*10000, feeSum-fee*10000)
	fmt.Printf("  feeSum == 400      ? %t   ← 对账时用 == 判断就会「不平账」\n", feeSum == 400)
	fmt.Println("  注意误差方向：0.1 是偏小，0.04 是偏大。它不守规矩，你没法预先补偿")

	// ── 示例 4：大数丢精度 ──────────────────────────────────────────────
	fmt.Println("🧪 示例 4：数一大，连整数都存不下")
	fmt.Println(strings.Repeat("-", 60))
	var big float64 = 1e18 // 1 ETH 的 wei 数，交易机器人里天天见
	fmt.Printf("  big       = %.0f\n", big)
	fmt.Printf("  big + 1   = %.0f\n", big+1)
	fmt.Printf("  big + 1 == big ? %t   ← 1 wei 凭空消失了\n", big+1 == big)
	fmt.Printf("  这一量级的 ULP = %.0f，也就是说 1e18 附近的最小刻度是 128\n",
		math.Nextafter(big, math.Inf(1))-big)
	fmt.Printf("  big + 64  = %.0f（还是原样：64 不足半个刻度，被舍掉）\n", big+64)
	fmt.Printf("  big + 127 = %.0f（超过半个刻度，一下跳到 +128）\n", big+127)
	fmt.Println("  ⚠️  再次提醒：直接写常量 1e18+1 == 1e18 是 false（编译期精确计算），")
	fmt.Println("      只有落到 float64 变量上才是 true —— 真实代码里必然走变量这条路")

	// 精度分界线：2^53
	var i53 float64 = 1 << 53
	fmt.Printf("\n  2^53 = %.0f 是 float64 能**精确表示所有整数**的上限\n", i53)
	fmt.Printf("  2^53 + 1 == 2^53 ? %t\n", i53+1 == i53)
	fmt.Printf("  float64(9007199254740993) = %.0f ← 奇数位丢了\n", 9007199254740993.0)
	fmt.Println("  而 uint256 余额、1e18 量级的 wei 全都远超 2^53，只能上 big.Int / decimal")

	// ── 示例 5：交易所场景 —— 对齐 stepSize 被拒单 ──────────────────────
	fmt.Println("⚙️ 示例 5：交易所场景：stepSize 对齐算错 → 下单量变成 0 → 拒单")
	fmt.Println(strings.Repeat("-", 60))
	// 交易所规则：下单量必须是 stepSize 的整数倍，且 >= minQty。
	// 机器人常见做法：用「预算 / 价格」算出下单量，再向下对齐到 stepSize。
	price, budget, stepSize := 7.0, 0.7, 0.1
	qty := budget / price
	fmt.Printf("  价格 = %v USDT，预算 = %v USDT，stepSize = %v\n", price, budget, stepSize)
	fmt.Printf("  理论下单量 = %.20f  （人眼看是 0.1，实际上小了一丁点）\n", qty)
	aligned := alignDownFloat(qty, stepSize)
	fmt.Printf("  对齐 stepSize 后 = %.20f\n", aligned)
	fmt.Printf("  aligned <= 0 ? %t\n", aligned <= 0)
	fmt.Println("  ⚠️  交易所收到 qty=0 → 直接报错「Order quantity is too small / 小于最小交易量」")
	fmt.Println("      而你在日志里看到的 qty 明明是 0.1，排查半天找不到原因")
	fmt.Println("      根因：floor(0.09999999999999999167 / 0.1) = floor(0.9999999999999999) = 0")

	// 同样的对齐逻辑，换一组数字会「少下单」而不是「被拒单」——更隐蔽
	// 注意必须用变量：0.3/0.1 写成常量会被编译期精确算成 3.0，看不到问题
	var qtyA, stepA float64 = 0.3, 0.1
	fmt.Printf("\n  另一组：qty=%v, stepSize=%v\n", qtyA, stepA)
	fmt.Printf("    qty/stepSize  = %.20f   ← 差一点点到 3\n", qtyA/stepA)
	fmt.Printf("    floor 后       = %.0f\n", math.Floor(qtyA/stepA))
	fmt.Printf("    对齐结果       = %.20f  ← 应该是 0.3，却下成了 0.2\n", alignDownFloat(qtyA, stepA))
	fmt.Println("    少下 1/3 的量不会报错，只会让你的策略慢慢跑偏 —— 这种 bug 最难发现")

	// 还有一种：用累加出来的成交量去对齐，误差让「对账」失败
	fmt.Println("\n  再一种：三笔 0.1 的成交累加后去对齐")
	var filled float64
	for i := 0; i < 3; i++ {
		filled += 0.1
	}
	fmt.Printf("    累加结果 = %.20f\n", filled)
	fmt.Printf("    == 0.3   ? %t   ← 幂等去重 / 对账比较会失败\n", filled == 0.3)
	fmt.Printf("    对齐后   = %.20f\n", alignDownFloat(filled, 0.1))

	// 「加个 1e-9 容差」这种土办法
	fmt.Printf("\n  网上常见的「土办法」：floor(qty/step + 1e-9) * step = %.20f\n",
		math.Floor(qty/stepSize+1e-9)*stepSize)
	fmt.Println("  这次救回来了，但容差该给多大完全靠猜：stepSize 可能是 1e-8，也可能是 100")
	fmt.Println("  → 正确的解法不是调容差，而是**换数据类型**。下一关（W-02）上 decimal")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ float64 = 1 位符号 + 11 位指数 + 52 位尾数，存的是**二进制**小数")
	fmt.Println("  ✅ 0.1 / 0.2 / 0.3 在二进制里是无限循环小数，52 位装不下 → 必然有误差")
	fmt.Println("  ✅ 0.1+0.2 与 0.3 只差 1 个 ULP，但 == 判 false；误差方向不可预测")
	fmt.Println("  ✅ 累加比乘法危险：误差会持续叠加，一万笔手续费能差到 1e-11 量级")
	fmt.Println("  ✅ 超过 2^53 ≈ 9.007e15 之后，float64 连整数都不能精确表示")
	fmt.Println("  ✅ 1e18 附近 1 个 ULP = 128，1 wei 直接被吃掉 —— 余额绝不能用 float64")
	fmt.Println("  ✅ 交易所 stepSize 对齐一旦用 float64，可能算出 qty=0 被拒单，或悄悄少下单")
	fmt.Println("  ✅ 铁律：**金额、数量、价格、手续费一律不用 float64**")
	fmt.Println("    下一关 W-02 用 shopspring/decimal 把上面这些例子重跑一遍，全部归零")
}
