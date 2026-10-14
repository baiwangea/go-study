//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
//
// 跑法：cd 07-web3-trading && go run 15_erc20_approve.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

/*
═══════════════════════════════════════════════════════════════════════
W-15：ERC20 approve / allowance —— 机器人动别人合约之前
═══════════════════════════════════════════════════════════════════════

【核心概念】
    现货机器人在 CEX 下单不需要授权。链上换 DEX、进金库、走路由，
    必须先让合约有权扣你的代币：

        allowance(owner, spender) → 还能扣多少
        approve(spender, amount)  → 把额度设成 amount（是「设成」，不是「加上」）

    选择器（无空格）：
        approve(address,uint256)   → 095ea7b3
        allowance(address,address) → dd62ed3e

    三个生产陷阱：

      1. 无限授权：amount = 2^256-1。省 gas，但 spender 被黑就等于钱包被黑。
         热钱包只给路由合约「这一笔要用的量」。
      2. USDT 一类老合约：非 0 额度再 approve 别的数会 revert。
         正确顺序：approve(spender, 0) 再 approve(spender, n)。
      3. approve 是「覆盖」不是「累加」。并发两笔 approve 会互相踩。
         和 W-13 一样：同一钱包的授权交易也要排队。

【你会看到】
    ✓ approve / allowance 的 calldata 拆槽
    ✓ 精确额度 vs type(uint256).max
    ✓ 「先清零再授权」的两笔 nonce

【思考题】
    1. 路由合约升级到新地址，旧 approve 还有效吗？
    2. increaseAllowance 解决了什么竞态？为什么很多 DEX 仍要你 approve max？
    3. 授权额度用 decimal 还是 *big.Int？链上只有 uint256。

【通关标准】
    能 Pack 出 approve/allowance，能说清无限授权的代价，
    以及 USDT「必须先 approve 0」的原因。

【前置】W-06 · W-08        【标签】ERC20 · approve · allowance · uint256.max
═══════════════════════════════════════════════════════════════════════
*/

var tokenABI = mustABI(`[
  {"name":"approve","type":"function",
   "inputs":[{"name":"spender","type":"address"},{"name":"amount","type":"uint256"}],
   "outputs":[{"name":"","type":"bool"}]},
  {"name":"allowance","type":"function",
   "inputs":[{"name":"owner","type":"address"},{"name":"spender","type":"address"}],
   "outputs":[{"name":"","type":"uint256"}]}
]`)

func mustABI(js string) abi.ABI {
	parsed, err := abi.JSON(strings.NewReader(js))
	if err != nil {
		panic(err)
	}
	return parsed
}

func dump(label string, data []byte) {
	fmt.Printf("  %s（%d 字节）\n", label, len(data))
	fmt.Printf("    [选择器] %s\n", hex.EncodeToString(data[:4]))
	for i := 0; i*32 < len(data)-4; i++ {
		start := 4 + i*32
		end := start + 32
		fmt.Printf("    [槽 %d ] %s\n", i, hex.EncodeToString(data[start:end]))
	}
}

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 W-15 ERC20 approve / allowance")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	router := common.HexToAddress("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D") // Uniswap V2
	owner := common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266")
	amount := new(big.Int).Mul(big.NewInt(100), new(big.Int).Exp(big.NewInt(10), big.NewInt(6), nil)) // 100 USDT（6 位）

	fmt.Println("📝 示例 1：精确授权 100 USDT")
	fmt.Println(strings.Repeat("-", 60))
	call, err := tokenABI.Pack("approve", router, amount)
	if err != nil {
		fmt.Println("  Pack 失败：", err)
		return
	}
	dump("approve(router, 100e6)", call)
	fmt.Println("  ➡️  槽 0 是 spender（地址左补齐），槽 1 是 uint256 额度。")
	fmt.Println("     这才是热钱包该发的量：这一笔 swap 要用多少，就批多少。")

	fmt.Println("\n📌 示例 2：type(uint256).max —— 省 gas，把钥匙交给路由")
	fmt.Println(strings.Repeat("-", 60))
	max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	inf, _ := tokenABI.Pack("approve", router, max)
	dump("approve(router, 2^256-1)", inf)
	fmt.Println("  ⚠️  spender 被攻破 = 你钱包里这种代币能被掏空。")
	fmt.Println("     交易机器人的热钱包本来就只放亏得起的量；即便如此也不该无限授权。")

	fmt.Println("\n🔍 示例 3：allowance 查询的 calldata")
	fmt.Println(strings.Repeat("-", 60))
	q, _ := tokenABI.Pack("allowance", owner, router)
	dump("allowance(owner, router)", q)
	fmt.Println("  ➡️  这是 eth_call 只读，不消耗 nonce。下单前先看还剩多少额度。")

	fmt.Println("\n🧪 示例 4：USDT 老合约必须先清零")
	fmt.Println(strings.Repeat("-", 60))
	zero, _ := tokenABI.Pack("approve", router, big.NewInt(0))
	fmt.Printf("  第 1 笔 nonce=N   approve(0)    selector=%s\n", hex.EncodeToString(zero[:4]))
	fmt.Printf("  第 2 笔 nonce=N+1 approve(100e6) selector=%s\n", hex.EncodeToString(call[:4]))
	fmt.Println("  ➡️  两笔都是 095ea7b3，差别只在槽 1。中间不能插别的 nonce。")
	fmt.Println("     并发两个 approve 覆盖同一额度，就是 W-13 抢 nonce 的代币版。")

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ approve(address,uint256) = 095ea7b3，额度是覆盖不是累加")
	fmt.Println("  ✅ 热钱包按笔授权；2^256-1 等于把钥匙交给 spender")
	fmt.Println("  ✅ 部分老 USDT：非 0 再改额度会 revert，先 approve(0)")
	fmt.Println("  ✅ 链上金额用 *big.Int（最小单位）；展示层再用 decimal")
	fmt.Println("  ✅ 授权交易同样消耗 nonce，必须排队（W-13）")
	fmt.Println(strings.Repeat("=", 60) + "\n")
}
