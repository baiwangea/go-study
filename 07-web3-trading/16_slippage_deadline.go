//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
//
// 跑法：cd 07-web3-trading && go run 16_slippage_deadline.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

/*
═══════════════════════════════════════════════════════════════════════
W-16：滑点与 deadline —— 链上成交价不是你看见的价
═══════════════════════════════════════════════════════════════════════

【核心概念】
    CEX 下单有你挂的价。DEX 的 swap 进块时，池子价格可能已经变了：
    你签名时看到的报价，和矿工执行时的报价，中间隔着 mempool。

    保护就两道，写进 calldata，EVM 自己强制：

      amountOutMin  最少拿到多少。实际少了就 revert（W-14 status=0）
      deadline      unix 秒。过了就 revert，避免交易在池子里躺几小时
                    被按过期行情成交

    计算必须用 decimal（W-01~W-03）：

        minOut = quoteOut * (1 - slippage)
        slippage 常用 5~50 bp（1 bp = 0.01%）

    回测里的 slippage-bp 是「模拟成交价往不利方向偏」；
    链上的 amountOutMin 是「偏过这条线就宁可不成交」。
    两者方向相反、目的相同：别在最差的价上成交。

【你会看到】
    ✓ 报价 1000，50 bp 滑点 → amountOutMin = 995
    ✓ 成交 996 通过、成交 994 revert
    ✓ deadline 用 unix 秒；过期的签名必须重签，不能只改节点时钟的预期

【思考题】
    1. 滑点设 0 会怎样？设 5% 又会怎样？夹子机器人吃哪一种？
    2. deadline = now+20 分钟和 now+20 秒，对 meltdown / 拥堵各有什么代价？
    3. 为什么 amountOutMin 要用代币最小单位的整数，而不是「995.0」这种十进制？

【通关标准】
    能用 decimal 算出 amountOutMin，能解释 deadline 防的是「过期成交」，
    并能把「回测滑点」和「链上 amountOutMin」区分开。

【前置】W-03 · W-08 · W-14   【标签】滑点 · amountOutMin · deadline · MEV
═══════════════════════════════════════════════════════════════════════
*/

func amountOutMin(quote, slippageBP decimal.Decimal) decimal.Decimal {
	// min = quote * (10000 - bp) / 10000，全程定点，禁止 float64
	keep := decimal.NewFromInt(10000).Sub(slippageBP)
	return quote.Mul(keep).Div(decimal.NewFromInt(10000))
}

func wouldRevert(actual, minOut decimal.Decimal, now, deadline int64) string {
	if now > deadline {
		return "❌ revert: Expired"
	}
	if actual.LessThan(minOut) {
		return fmt.Sprintf("❌ revert: InsufficientOutput actual=%s < min=%s", actual, minOut)
	}
	return fmt.Sprintf("✅ 成交 actual=%s ≥ min=%s", actual, minOut)
}

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 W-16 滑点与 deadline")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	quote := decimal.RequireFromString("1000")
	bp := decimal.NewFromInt(50) // 50 bp = 0.5%
	minOut := amountOutMin(quote, bp)

	fmt.Println("📝 示例 1：50 bp 滑点怎么算 amountOutMin")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("  报价 quoteOut     = %s\n", quote)
	fmt.Printf("  滑点              = %s bp = 0.5%%\n", bp)
	fmt.Printf("  amountOutMin      = %s\n", minOut)
	fmt.Println("  ➡️  公式：quote * (10000-bp) / 10000，全程 decimal。")
	fmt.Println("     回测里的 -slippage-bp 是把成交价往不利方向推；")
	fmt.Println("     这里是设下限，比下限差就 revert。")

	fmt.Println("\n📌 示例 2：池子价格动了，两条成交")
	fmt.Println(strings.Repeat("-", 60))
	deadline := time.Date(2026, 10, 14, 8, 0, 20, 0, time.UTC).Unix()
	now := time.Date(2026, 10, 14, 8, 0, 10, 0, time.UTC).Unix()
	fmt.Println("  " + wouldRevert(decimal.RequireFromString("996"), minOut, now, deadline))
	fmt.Println("  " + wouldRevert(decimal.RequireFromString("994"), minOut, now, deadline))
	fmt.Println("  ➡️  996 还能接受；994 说明夹子或波动已经超过你设的 50 bp。")

	fmt.Println("\n🔍 示例 3：deadline 过了，报价再好也 revert")
	fmt.Println(strings.Repeat("-", 60))
	late := deadline + 1
	fmt.Println("  " + wouldRevert(decimal.RequireFromString("1000"), minOut, late, deadline))
	fmt.Println("  ➡️  交易在 mempool 躺过 deadline，必须按 W-13 换 nonce 重签。")
	fmt.Println("     不能指望节点帮你改 calldata —— 签名已经锁死这两个数。")

	fmt.Println("\n🧪 示例 4：滑点 0 vs 滑点 500 bp")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("  0 bp   minOut=%s  （几乎必然 revert，除非你是出块者）\n", amountOutMin(quote, decimal.Zero))
	fmt.Printf("  500 bp minOut=%s  （5%%，夹子的舒适区）\n", amountOutMin(quote, decimal.NewFromInt(500)))
	fmt.Println("  ➡️  热路径常用 20~100 bp；再宽就是在请人夹你。")

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ amountOutMin = quoteOut * (1 - slippage)，用 decimal")
	fmt.Println("  ✅ 实际输出 < minOut → 合约 revert，receipt.status=0（W-14）")
	fmt.Println("  ✅ deadline 是 unix 秒，写进 calldata，过期必须重签")
	fmt.Println("  ✅ 回测滑点是模拟吃亏；链上 minOut 是拒绝吃亏")
	fmt.Println("  ✅ 滑点越宽越容易被夹；0 滑点在公共 mempool 几乎下不进去")
	fmt.Println(strings.Repeat("=", 60) + "\n")
}
