//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 07-web3-trading && go run 03_decimal_rounding.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"math"

	"github.com/shopspring/decimal"
)

/*
═══════════════════════════════════════════════════════════════════════
W-03：精度对齐与舍入
═══════════════════════════════════════════════════════════════════════

【核心概念】
    W-02 解决了「加减乘除算不准」，这一关解决「算准了之后怎么落到交易所能接受的数」。

    交易所给你的每一笔下单都受两条规则约束：
        StepSize：下单量必须是它的整数倍（如 0.001 → 只能下 0.123，不能下 0.1234）
        MinQty  ：下单量不能小于它（如 0.001）
    这两条都是**字符串下发**的，所以正确做法是：decimal.RequireFromString(filter.StepSize)，
    而不是 parseFloat 再转 decimal。

    舍入模式有九种，交易里常用的只有三种：
        Floor()       向下取整（对齐 stepSize 用它 —— 宁可少下也别超出余额）
        Round(n)      四舍五入（= 远离零方向，中国人最习惯的那种）
        RoundBank(n)  银行家舍入（= 四舍六入五成双，会计/对账场景用它）
    ⚠️ Round 和 RoundBank 在「恰好 .5」时才分道扬镳 —— 但一旦分道扬镳，
       一千笔累加下来能差出可观的金额。

【你会看到】
    ✓ 九种舍入模式在同一批数字上的完整对照表
    ✓ Round vs RoundBank：0.5/1.5/…/99.5 求和，前者得 5050，后者得 5000（精确值是 5000）
    ✓ decimal.Round(2) 把 1.005 正确算成 1.01，而 float64 的 math.Round(1.005*100)/100 给 1.00
    ✓ decimal 版 stepSize 对齐把 W-01 的 qty=0 变成正确的 0.1
    ✓ 10000 笔手续费：float64 累加 400.000000000057014，decimal 精确 400
    ✓ StringFixed(8) 生成 "0.10000000" 这种交易所一定能解析的字符串

【思考题】（动手改代码验证，别空想）
    1. 示例 1 的表里，RoundCeil / RoundFloor / RoundUp / RoundDown 四列在负数上
       表现得很有意思。请解释：
         · RoundUp 和 RoundCeil 在 2.4 上都给 3，在 -2.4 上一个给 -3 一个给 -2，为什么？
         · Truncate(0) 和 RoundDown(0) 结果完全一样，它们真的等价吗？
       （提示：一个朝零截断，一个朝负无穷 —— 在负数上差别才显形。）
    2. 示例 2 里 0.5/1.5/…/99.5 求和，Round 得 5050，RoundBank 得 5000，精确值也是 5000。
       为什么银行家舍入在「大量恰好 .5 的数据」上更准？它真的永远更准吗？
       （提示：如果数据都是 x.5 且 x 恒为偶数呢？）
    3. 示例 3 里 1.005 用 float64 算成 1.00，用 decimal 算成 1.01。
       请用 W-01 的知识解释：1.005 存成 float64 之后实际是多少？
       为什么 math.Round(x*100) 这个「标准写法」在这里会失效？
    4. **本关最重要的一题**：示例 4 的 NormalizeQty 用的是 Floor（向下对齐）。
       把 Floor 换成 Round 再跑一遍，观察 qty=0.0999、stepSize=0.01 这类输入的结果。
       然后回答：为什么下单量**必须**向下对齐而不是四舍五入？
       （提示：想一想「账户只有 0.0999，四舍五入成 0.1 会发生什么」。）
    5. 示例 6 的 10000 笔手续费，float64 差了 5.7e-11 USDT。这个数字对钱毫无影响，
       那它到底会带来什么麻烦？请举出至少两个真实会出问题的场景。
       （提示：== 比较、map key、写进数据库 DECIMAL(20,8) 后的回读校验、报表导出。）
    6. 跨语言对照：Java 的 BigDecimal 有 RoundingMode.HALF_UP / HALF_EVEN，
       Python 的 decimal 有 ROUND_HALF_UP / ROUND_HALF_EVEN。
       Go 的 decimal 用 Round / RoundBank 两个方法名来表达，你觉得哪种更好用？
       有没有可能把舍入模式写成配置项下发（很多交易框架都这么做）？

【通关标准】
    能说清 Round / RoundBank / Floor / Truncate 的差别并选对场景，
    能写出一个「读交易所 filter → 对齐 stepSize → 校验 minQty → 输出定长字符串」的完整函数，
    并且知道手续费累加必须用 decimal 而不是 float64。

【前置】W-01（float64 的坑）、W-02（decimal 基本用法）  【标签】舍入 · 银行家舍入 · stepSize · 对账
═══════════════════════════════════════════════════════════════════════
*/

// SymbolFilter 模拟交易所下发的「交易规则」（对应币安的 symbol filters、
// 火币的 precision、OKX 的 lotSz/tickSz）。
//
// 真实实现里这三个值都是从 exchangeInfo / instruments 接口拿到的**字符串**，
// 一定要原样保留成 string 再交给 decimal.RequireFromString，
// 中间绝不能经过 float64 —— 一旦过一遍，W-01 的误差就回来了。
type SymbolFilter struct {
	Symbol   string
	StepSize decimal.Decimal // 下单量必须是它的整数倍
	MinQty   decimal.Decimal // 最小下单量
	QtyScale int32           // 数量保留几位小数（生成 API 字符串用）
}

// NormalizeQty 把「策略算出来的原始下单量」规范化成交易所能接受的量。
//
// 流程：向下对齐 stepSize → 校验 minQty → 按 QtyScale 定长输出。
// 为什么向下对齐（Floor）而不是四舍五入（Round）？
//
//	账户里只有 0.0999 个币，四舍五入成 0.1 就超卖了，交易所会报
//	"insufficient balance"，而且这种错在回测里根本复现不出来。
//	宁可少下一点，也不要下不出去。
//
// 注意 decimal 的 Div 默认只保留 16 位小数（decimal.DivisionPrecision）。
// 本例的 qty/stepSize 比值都远小于 16 位，够用；如果你的 stepSize 小到
// 1e-18 而 qty 又很大，应该改成按 Exponent 做整数运算，别依赖 Div 的默认精度。
func NormalizeQty(raw decimal.Decimal, f SymbolFilter) (decimal.Decimal, error) {
	if f.StepSize.IsZero() || f.StepSize.IsNegative() {
		return decimal.Zero, fmt.Errorf("非法的 stepSize: %s", f.StepSize.String())
	}
	// 除以 stepSize → 向下取整 → 乘回 stepSize，得到「不超过 raw 的最大对齐值」
	qty := raw.Div(f.StepSize).Floor().Mul(f.StepSize)
	if qty.LessThan(f.MinQty) {
		return decimal.Zero, fmt.Errorf(
			"下单量 %s 小于最小交易量 %s（原始值 %s，stepSize %s）",
			qty.StringFixed(f.QtyScale), f.MinQty.String(), raw.String(), f.StepSize.String())
	}
	return qty, nil
}

// alignDownFloat 是 W-01 留下的**反面教材**：同样的对齐逻辑，用 float64 写。
// 这一关会把它和 NormalizeQty 放在一起跑，让你亲眼看到差别。
func alignDownFloat(qty, step float64) float64 {
	return math.Floor(qty/step) * step
}

func main() {
	fmt.Println("\n=== W-03 精度对齐与舍入 ===")

	// ── 示例 1：九种舍入模式总表 ─────────────────────────────────────────
	fmt.Println("\n【示例 1】九种舍入模式对照（places=0）")
	fmt.Println("  下面每一列都是对同一个 x 的舍入结果，请逐列横着看：")
	// 表头全是 ASCII，和数据列一样按字节对齐，不会出现中文宽度错位
	fmt.Printf("  %-7s %-8s %-10s %-10s %-11s %-9s %-10s %-9s %-6s %-6s\n",
		"x", "Round", "RoundBank", "RoundCeil", "RoundFloor", "RoundUp", "RoundDown", "Truncate", "Floor", "Ceil")
	for _, s := range []string{"2.5", "3.5", "0.5", "-0.5", "-2.5", "2.4", "2.6", "-2.4"} {
		d := decimal.RequireFromString(s)
		fmt.Printf("  %-7s %-8s %-10s %-10s %-11s %-9s %-10s %-9s %-6s %-6s\n",
			s,
			d.Round(0).String(),      // 四舍五入（half away from zero）
			d.RoundBank(0).String(),  // 银行家舍入（half to even）
			d.RoundCeil(0).String(),  // 朝 +∞
			d.RoundFloor(0).String(), // 朝 -∞
			d.RoundUp(0).String(),    // 朝远离零的方向
			d.RoundDown(0).String(),  // 朝靠近零的方向（= 朝零截断）
			d.Truncate(0).String(),   // 朝零截断
			d.Floor().String(),       // 朝 -∞
			d.Ceil().String())        // 朝 +∞
	}
	fmt.Println("  读表要点：")
	fmt.Println("    • Round 与 RoundBank 只在「恰好 .5」时不同（2.5→3 vs 2；-2.5→-3 vs -2）")
	fmt.Println("    • RoundUp 朝远离零走，RoundCeil 朝 +∞ 走 —— 负数上就分家了（-2.4 → -3 vs -2）")
	fmt.Println("    • Truncate(0) 和 RoundDown(0) 在整数位上都朝零截断，本例结果一致")
	fmt.Println("    • 交易里最常用的是 Floor()（对齐 stepSize）和 Round(n)（算手续费）")

	// ── 示例 2：银行家舍入 vs 四舍五入 ───────────────────────────────────
	fmt.Println("\n【示例 2】银行家舍入（RoundBank）为什么存在")
	fmt.Println("  四舍五入：遇到 .5 一律进位 → 长期看**系统性偏高**")
	fmt.Println("  银行家舍入：遇到 .5 时进到最近的偶数（2.5→2, 3.5→4, 4.5→4, 5.5→6）")
	fmt.Println("  设计目的是让大量数据的舍入误差**互相抵消**，而不是单向累积")
	fmt.Println("\n  实验：把 0.5, 1.5, 2.5, …, 99.5 这 100 个数分别用两种方式取整后求和")
	sumRound := decimal.Zero
	sumBank := decimal.Zero
	sumExact := decimal.Zero
	for i := 0; i < 100; i++ {
		v := decimal.NewFromInt(int64(i)).Add(decimal.RequireFromString("0.5"))
		sumRound = sumRound.Add(v.Round(0))
		sumBank = sumBank.Add(v.RoundBank(0))
		sumExact = sumExact.Add(v)
	}
	fmt.Printf("    精确值            = %s\n", sumExact.String())
	fmt.Printf("    Round(0) 后求和   = %s   ← 比精确值高了 %s\n",
		sumRound.String(), sumRound.Sub(sumExact).String())
	fmt.Printf("    RoundBank(0) 求和 = %s   ← 和精确值完全一致\n", sumBank.String())
	fmt.Println("    ↑ 100 个数就差出 50。手续费按笔累计时，这个偏差会一路传导到对账")
	fmt.Println("  注意：RoundBank 不是「永远更准」，只是「在恰好 .5 的数据上不偏」。")
	fmt.Println("        真实数据几乎不会恰好落在 .5，所以两者日常差别很小 —— ")
	fmt.Println("        但**一旦**落在 .5（比如手续费正好是 0.125），你就必须知道用的是哪种")

	// ── 示例 3：精度案例 —— decimal 准，float64 不准 ─────────────────────
	fmt.Println("\n【示例 3】1.005 该舍入成多少？decimal 和 float64 吵起来了")
	fmt.Println("  decimal.Round(2) 对 1.005 给 1.01（十进制下 1.005 就是 1.005）")
	fmt.Println("  float64 的 math.Round(1.005*100)/100 却给 1.00 —— 为什么？")
	fmt.Println()
	for _, s := range []string{"1.005", "1.015", "1.045", "2.675", "0.125", "0.135"} {
		d := decimal.RequireFromString(s)
		f, _ := d.Float64()
		fmt.Printf("    %-7s float64 实际是 %.17g  math.Round(x*100)/100 = %.2f  |  decimal Round(2)=%s  RoundBank(2)=%s\n",
			s, f, math.Round(f*100)/100, d.Round(2).String(), d.RoundBank(2).String())
	}
	fmt.Println("  ↑ 1.005 存成 float64 后是 1.00499999999999989…（比 1.005 小），")
	fmt.Println("    乘 100 得 100.499999…，math.Round 向下取成 100，于是得到 1.00 —— 错了一位")
	fmt.Println("    decimal 里 1.005 就是精确的 1.005，Round(2) 正确给 1.01")
	fmt.Println("  ⚠️  math.Round(x*100)/100 这个「标准写法」本身就是 float64 的坑，别再用它做金额")

	// ── 示例 4：stepSize 对齐 —— 用 decimal 重写 W-01 ────────────────────
	fmt.Println("\n【示例 4】stepSize 对齐：decimal 重写 W-01 的拒单场景")
	// 先把 W-01 的 float64 版本再跑一遍（必须用变量，否则会被编译期常量折叠救回来）
	var priceF, budgetF, stepF float64 = 7.0, 0.7, 0.1
	qtyF := budgetF / priceF
	fmt.Printf("  float64: 价格=%v 预算=%v stepSize=%v\n", priceF, budgetF, stepF)
	fmt.Printf("           qty     = %.20f\n", qtyF)
	fmt.Printf("           对齐后  = %.20f   ← 变成 0，交易所直接拒单\n", alignDownFloat(qtyF, stepF))

	// 再跑 decimal 版本
	priceD := decimal.RequireFromString("7")
	budgetD := decimal.RequireFromString("0.7")
	stepD := decimal.RequireFromString("0.1")
	qtyD := budgetD.Div(priceD)
	fmt.Printf("  decimal: 价格=%s 预算=%s stepSize=%s\n", priceD.String(), budgetD.String(), stepD.String())
	fmt.Printf("           qty     = %s   ← 精确的 0.1\n", qtyD.String())
	fmt.Printf("           对齐后  = %s   ← 正确\n", qtyD.Div(stepD).Floor().Mul(stepD).String())

	// 再来一个 W-01 里「悄悄少下 1/3」的例子
	var q3F, s3F float64 = 0.3, 0.1
	fmt.Printf("\n  float64: 0.3 对齐 stepSize 0.1 → %.20f  ← 应该是 0.3，下成了 0.2\n",
		alignDownFloat(q3F, s3F))
	q3D := decimal.RequireFromString("0.3")
	s3D := decimal.RequireFromString("0.1")
	fmt.Printf("  decimal: 0.3 对齐 stepSize 0.1 → %s   ← 正确\n",
		q3D.Div(s3D).Floor().Mul(s3D).String())

	// 最后看工程版：读 filter → 对齐 → 校验 → 输出
	fmt.Println("\n  工程版：NormalizeQty（向下对齐 + minQty 校验）")
	btc := SymbolFilter{
		Symbol:   "BTCUSDT",
		StepSize: decimal.RequireFromString("0.001"),
		MinQty:   decimal.RequireFromString("0.001"),
		QtyScale: 3,
	}
	doge := SymbolFilter{
		Symbol:   "DOGEUSDT",
		StepSize: decimal.RequireFromString("0.1"),
		MinQty:   decimal.RequireFromString("1"),
		QtyScale: 1,
	}
	// 第三个 filter 专门复刻 W-01 的拒单场景：stepSize 和 minQty 都是 0.1
	alt := SymbolFilter{
		Symbol:   "ALTUSDT",
		StepSize: decimal.RequireFromString("0.1"),
		MinQty:   decimal.RequireFromString("0.1"),
		QtyScale: 1,
	}
	for _, tc := range []struct {
		f   SymbolFilter
		raw string
	}{
		{btc, "0.1234567"},  // 正常，向下对齐到 0.123
		{btc, "0.0009999"},  // 对齐后是 0.000，低于 minQty
		{doge, "1.95"},      // 正常
		{doge, "0.0999999"}, // 对齐后为 0
		// ↓ W-01 的复刻：同一笔预算/价格，float64 算出来是 0.09999999999999999167，
		//   decimal 算出来是 0.1。把两个字符串分别喂进同一个 filter，结果天差地别
		{alt, "0.09999999999999999167"}, // float64 版本 → 对齐成 0 → 拒单
		{alt, "0.1"},                    // decimal 版本 → 0.1 → 通过
	} {
		raw := decimal.RequireFromString(tc.raw)
		got, err := NormalizeQty(raw, tc.f)
		if err != nil {
			fmt.Printf("    %-10s raw=%-12s → ❌ %v\n", tc.f.Symbol, tc.raw, err)
			continue
		}
		fmt.Printf("    %-10s raw=%-12s → ✅ qty=%s（发出去用 %q）\n",
			tc.f.Symbol, tc.raw, got.String(), got.StringFixed(tc.f.QtyScale))
	}
	fmt.Println("  ⚠️  注意报错信息里把「原始值 / stepSize / minQty」都打了出来 —— ")
	fmt.Println("      这类拒单如果不打原始值，你只会看到一个漂亮的 0.000，排查无从下手")

	// ── 示例 5：StringFixed 生成给交易所的字符串 ─────────────────────────
	fmt.Println("\n【示例 5】StringFixed(n)：生成给交易所的字符串")
	for _, s := range []string{"0.1", "1", "0.00000123", "1234.5"} {
		d := decimal.RequireFromString(s)
		fmt.Printf("  %-12s String()=%-14s StringFixed(8)=%-14s StringFixed(2)=%s\n",
			s, d.String(), d.StringFixed(8), d.StringFixed(2))
	}
	fmt.Println("  • String() 会去掉末尾的 0：1.00000000 打印成 \"1\"")
	fmt.Println("    有些交易所的字段校验是**按字符串长度**做的，\"1\" 可能被拒；")
	fmt.Println("    更麻烦的是它会让你的订单幂等 key 出现 \"1\" 和 \"1.00000000\" 两个版本")
	fmt.Println("  • StringFixed(n) 补零到 n 位，并**四舍五入**（走的是 Round，不是 RoundBank）")
	fmt.Printf("    验证：0.125 的 StringFixed(2)=%s，StringFixedBank(2)=%s（银行家舍入不进位）\n",
		decimal.RequireFromString("0.125").StringFixed(2),
		decimal.RequireFromString("0.125").StringFixedBank(2))
	fmt.Println("  • 给交易所的**数量用 QtyScale、价格用 PriceScale**，两个精度不一样，别混用")

	// ── 示例 6：对账 —— 10000 笔手续费 ──────────────────────────────────
	fmt.Println("\n【示例 6】对账：10000 笔手续费，float64 vs decimal")
	feeD := decimal.RequireFromString("0.04") // 每笔手续费 0.04 USDT
	feeF, _ := feeD.Float64()
	var sumF float64
	for i := 0; i < 10000; i++ {
		sumF += feeF
	}
	sumD := decimal.Zero
	for i := 0; i < 10000; i++ {
		sumD = sumD.Add(feeD)
	}
	expect := feeD.Mul(decimal.NewFromInt(10000))
	fmt.Printf("  每笔手续费      = %s USDT × 10000 笔\n", feeD.String())
	fmt.Printf("  float64 累加    = %.15f\n", sumF)
	fmt.Printf("  decimal 累加    = %s\n", sumD.String())
	fmt.Printf("  理论值          = %s\n", expect.String())
	fmt.Printf("  float64 的偏差  = %s USDT\n", decimal.NewFromFloat(sumF).Sub(expect).String())
	fmt.Printf("  float64 == 400      ? %t\n", sumF == 400)
	fmt.Printf("  decimal.Equal(400)  ? %t\n", sumD.Equal(decimal.NewFromInt(400)))
	fmt.Println("  偏差只有 5.7e-11 USDT，对钱没影响 —— 但它会让下面这些判断出错：")
	fmt.Println("    ① if total == expected  → 对账永远不平，告警天天响")
	fmt.Println("    ② 当 map key / 幂等 key → 同一个 400 存进去查不出来")
	fmt.Println("    ③ 写进 DECIMAL(20,8) 再读回 → 400.00000000 != 400.000000000057014")
	fmt.Println("  所以「误差小」不是理由，**类型选错**才是问题")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n【关键总结】")
	fmt.Println("  • Floor() 对齐 stepSize（宁可少下，不能超卖）；Round(n) 算手续费；")
	fmt.Println("    RoundBank(n) 会计/对账场景；Truncate(n) 朝零截断")
	fmt.Println("  • Round 与 RoundBank 只在「恰好 .5」时不同，但累计起来方向性偏差很可观")
	fmt.Println("    （0.5…99.5 求和：Round=5050，RoundBank=5000，精确值=5000）")
	fmt.Println("  • float64 的 math.Round(x*100)/100 在 1.005 上会算错（存的是 1.00499999999999989）")
	fmt.Println("  • stepSize 对齐必须用 decimal：float64 版会把 0.1 对齐成 0，直接拒单")
	fmt.Println("  • 校验链要完整：对齐 → minQty 检查 → StringFixed 定长输出，报错要带原始值")
	fmt.Println("  • 手续费累加必须用 decimal；float64 一万笔差 5.7e-11，破坏 == / map key / 回读校验")
	fmt.Println("  • 下一关 W-04 换赛道：从「钱的表示」转到「链上身份」—— 私钥、地址、keystore")
}
