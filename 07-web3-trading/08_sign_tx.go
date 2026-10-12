//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 07-web3-trading && go run 08_sign_tx.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

/*
═══════════════════════════════════════════════════════════════════════
W-08：EIP-1559 交易构造与离线签名 —— 机器人自己签名
═══════════════════════════════════════════════════════════════════════

【核心概念】

	前面 W-05 拿到了私钥，W-06 学会了拼 calldata。
	现在把两样东西合成一笔**签好名的、可以直接广播的交易**。

	EIP-1559（2021-08 London 升级）把交易的 gas 模型改成了两段式：

	    GasFeeCap（fee cap，上限）   —— 你愿意为每个 gas 付的**最高价**
	    GasTipCap（priority fee，小费）—— 你愿意额外给矿工/验证者的**激励**

	实际单价 = min(GasFeeCap, baseFee + GasTipCap)

	其中 baseFee 是协议算出来的、会随区块拥堵自动升降的**基础费**（被销毁）。
	所以要保证交易能被打包，必须满足：

	    GasFeeCap >= GasTipCap + baseFee        ← 本关会演示违反的后果

	字段清单（DynamicFeeTx）：

	    ChainID    1          → 链 ID。防重放的核心，签名时参与哈希
	    Nonce      7          → 这个账户已发出的交易数，必须连续、不能跳号
	    GasTipCap  1.5 gwei   → 小费
	    GasFeeCap  30 gwei    → 单价上限
	    Gas        21000      → gas 额度上限（普通转账固定 21000）
	    To         0x7099...  → 收款地址（nil = 部署合约）
	    Value      1 ETH      → 转账金额（wei）
	    Data       nil        → calldata（W-06 拼的那段）

	签名的意义：**私钥签名 = 授权**。签名之后任何人拿着这段字节都能广播，
	因为验证者能从签名里恢复出「是谁授权的」，恢复不出来就拒绝。

【你会看到】

	✓ 未签名交易的哈希 ≠ 最终交易哈希（前者没人认）
	✓ 签名后 tx.Hash() 就是交易哈希（在区块浏览器上查到的那个）
	✓ types.Sender(signer, signedTx) 恢复出的地址 == 私钥对应地址
	✓ 用 chainID=56 的 signer 去恢复 chainID=1 的交易 → 报错，拿到零地址
	✓ DynamicFeeTx 的 RLP 以 0x02 开头，LegacyTx 以 0xf8 开头（RLP list）
	✓ 签名让 RLP 从 53 字节涨到 118 字节（多出的就是 V/R/S）

【思考题】（动手改代码验证，别空想）

 1. 把 Nonce 从 7 改成 8，交易哈希变了，恢复出的地址还是同一个。
    为什么 Nonce 变了地址不变？（提示：地址只由私钥决定。）
    那 Nonce 在防什么？如果你同时发两笔 Nonce 相同的交易，链上会怎样？

 2. 示例 5 里用 chainID=56 的 signer 恢复 chainID=1 的交易会失败。
    现在反过来：用 chainID=1 的 signer **签名**一笔 ChainID=56 的交易，
    会发生什么？先猜，再改代码验证。
    （提示：DynamicFeeTx 自己带着 ChainID 字段。）

 3. 把 GasFeeCap 改成 1 gwei、GasTipCap 改成 2 gwei（feeCap < tipCap）。
    SignTx 会报错吗？EffectiveGasTipValue 返回什么？
    **如果这笔交易真的广播出去，会发生什么？**（提示：RPC 节点会拒绝，
    或者永远不被打包。所以生产代码里必须自己先校验。）

 4. 示例 6 对比了 Legacy 和 DynamicFee。现在把 LegacyTx 的 GasPrice 设成
    30 gwei，算一下「这笔交易最坏情况花多少钱」。
    再算 DynamicFeeTx 的最坏情况。**为什么 1559 的「最坏情况」更难预测？**

 5. 签名后 MarshalBinary 是 118 字节。把这 118 字节的 hex 全部打印出来，
    按 RLP 结构手工拆一遍：0x02 | list 长度 | chainID | nonce | tipCap |
    feeCap | gas | to | value | data | accessList | v | r | s。
    数一数每个字段各占几个字节。

【通关标准】

	能说清 DynamicFeeTx 每个字段的含义和 GasFeeCap/GasTipCap/baseFee 的关系，
	能用 SignTx + LatestSignerForChainID 完成离线签名，
	能用 Sender 从签名恢复发送者并验证，
	能解释链 ID 保护怎么防重放，
	并且知道「交易哈希 = 签名后交易的 Keccak256」。

【前置】W-05（私钥）· W-06（calldata）   【标签】EIP-1559 · SignTx · 交易哈希 · 链 ID 重放
═══════════════════════════════════════════════════════════════════════
*/

// testPrivateKeyHex 是 W-05 里 m/44'/60'/0'/0/0 派生出的私钥。
//
// ⚠️ 它是全公开的测试私钥（Hardhat / Anvil 的账户 0），
// 里面**绝不能有真钱**。写在这里是为了让本关输出可复现。
const testPrivateKeyHex = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

// testToAddress 是 W-05 里 /1 的地址，当成本关的收款方。
const testToAddress = "0x70997970C51812dc3A010C7d01b50e0d17dc79C8"

// wei 把「多少个 ether」换成 wei 的 big.Int。
//
// 1 ETH = 10^18 wei。链上所有金额都用 wei 表示，且必须是整数 ——
// 所以交易机器人里**绝不能**用 float64 存金额（0.1 ETH 在 float64 里
// 是 0.1000000000000000055511151231257827，转成 wei 就差了几十 wei）。
// 正确做法是全程 big.Int / decimal.Decimal，只在展示时才格式化。
func wei(eth int64) *big.Int {
	return new(big.Int).Mul(big.NewInt(eth), big.NewInt(1e18))
}

// gwei 把「多少个 gwei」换成 wei。1 gwei = 10^9 wei。
//
// gas 价格用 gwei 报价是行业习惯（就像油价用「分/升」），
// 但链上字段全是 wei，所以每次都要乘 1e9。
func gwei(n int64) *big.Int {
	return new(big.Int).Mul(big.NewInt(n), big.NewInt(1e9))
}

// tipCap15gwei 返回 1.5 gwei = 1500000000 wei。
//
// 单独写成函数，是因为 1.5 不是整数 —— 用 gwei(1.5) 会让浮点数混进金额计算。
// 金额相关的常量一律写成 wei 的整数字面量，一个浮点都不许有。
func tipCap15gwei() *big.Int {
	return big.NewInt(1_500_000_000)
}

// mustKey 从 hex 还原私钥，出错直接 panic（本关是固定的测试私钥）。
func mustKey(hexKey string) *ecdsa.PrivateKey {
	key, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		panic(fmt.Sprintf("私钥解析失败: %v", err))
	}
	return key
}

func main() {
	fmt.Println("\n=== W-08 EIP-1559 交易构造与离线签名 ===")

	key := mustKey(testPrivateKeyHex)
	from := crypto.PubkeyToAddress(key.PublicKey)
	fmt.Printf("  私钥（测试用，公开的）: %s\n", testPrivateKeyHex)
	fmt.Printf("  对应地址              : %s\n", from.Hex())
	fmt.Println("  ↑ 和 W-05 里 m/44'/60'/0'/0/0 派生出的地址一致 —— 同一个私钥。")

	// ── 示例 1：DynamicFeeTx 各字段的含义 ────────────────────────────────
	fmt.Println("\n【示例 1】DynamicFeeTx 的字段")
	to := common.HexToAddress(testToAddress)
	chainID := big.NewInt(1) // 1 = 以太坊主网，56 = BSC

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     7,
		GasTipCap: tipCap15gwei(), // 1.5 gwei
		GasFeeCap: gwei(30),       // 30 gwei
		Gas:       21000,
		To:        &to,
		Value:     wei(1),
		Data:      nil,
	})

	fmt.Printf("  ChainID   : %s   （链 ID：1=主网，56=BSC，137=Polygon）\n", tx.ChainId())
	fmt.Printf("  Nonce     : %d   （这个账户已发出的交易数，必须连续）\n", tx.Nonce())
	fmt.Printf("  GasTipCap : %s wei = %s gwei  （小费，给验证者）\n", tx.GasTipCap(), toGwei(tx.GasTipCap()))
	fmt.Printf("  GasFeeCap : %s wei = %s gwei  （单价上限）\n", tx.GasFeeCap(), toGwei(tx.GasFeeCap()))
	fmt.Printf("  Gas       : %d   （gas 额度上限，普通转账固定 21000）\n", tx.Gas())
	fmt.Printf("  To        : %s\n", tx.To().Hex())
	fmt.Printf("  Value     : %s wei = 1 ETH\n", tx.Value())
	fmt.Printf("  Data      : %d 字节（空 = 纯转账）\n", len(tx.Data()))
	fmt.Printf("  交易类型  : %d（0=Legacy，2=EIP-1559 DynamicFee）\n", tx.Type())

	fmt.Println("\n  GasFeeCap 与 GasTipCap 的关系（本关最重要的一条）：")
	fmt.Println("    实际单价 = min(GasFeeCap, baseFee + GasTipCap)")
	fmt.Println("    GasTipCap 是给验证者的小费，GasFeeCap 是你出的最高单价。")
	fmt.Println("    要保证能被及时打包，必须 GasFeeCap >= GasTipCap + baseFee。")
	fmt.Println("    下面用 EffectiveGasTipValue 验证这个公式：")
	lowBase := gwei(20)   // 假设当前 baseFee = 20 gwei
	highBase := gwei(100) // 假设拥堵了，baseFee 涨到 100 gwei
	fmt.Printf("      baseFee = 20 gwei  → 实际小费 = %s gwei（= 你的 tipCap，正常）\n",
		toGwei(tx.EffectiveGasTipValue(lowBase)))
	fmt.Printf("      baseFee = 100 gwei → 实际小费 = %s gwei（**负数！**）\n",
		toGwei(tx.EffectiveGasTipValue(highBase)))
	fmt.Println("      baseFee 超过 feeCap 时，你的交易根本进不了区块（验证者不干亏本买卖），")
	fmt.Println("      于是就一直卡在 mempool 里 —— 这是新手最常见的「交易卡住」原因。")

	// ── 示例 2：未签名交易的哈希 ─────────────────────────────────────────
	fmt.Println("\n【示例 2】未签名交易的哈希不是交易哈希")
	fmt.Printf("  未签名 tx.Hash() = %s\n", tx.Hash().Hex())
	fmt.Printf("  未签名 RLP 长度   = %d 字节\n", mustBinary(tx))
	fmt.Println("  ⚠️ 这个哈希是「没有签名的交易」的 Keccak256，链上没有任何人认它。")
	fmt.Println("     你在区块浏览器里永远查不到这个值。它只是一个中间产物。")
	fmt.Println("     真正的交易哈希要等签名之后才有（见示例 3）。")

	// ── 示例 3：离线签名 ─────────────────────────────────────────────────
	fmt.Println("\n【示例 3】types.SignTx：离线签名")
	// LatestSignerForChainID 会根据链 ID 挑最合适的 signer：
	//   chainID = nil → HomesteadSigner（没有 EIP-155 保护，危险）
	//   chainID = 1   → modernSigner（内部带 London/EIP-1559 规则）
	// 一定要用 LatestSignerForChainID，不要用已废弃的 NewEIP155Signer。
	signer := types.LatestSignerForChainID(chainID)
	fmt.Printf("  signer 类型: %T（由 LatestSignerForChainID(1) 自动挑选）\n", signer)

	signed, err := types.SignTx(tx, signer, key)
	if err != nil {
		fmt.Println("  签名失败：", err)
		return
	}
	fmt.Printf("  签名后 tx.Type()  = %d（签名不改变交易类型）\n", signed.Type())
	fmt.Printf("  签名后 tx.Hash()  = %s\n", signed.Hash().Hex())
	fmt.Println("  ↑ **这就是交易哈希**，广播后能在区块浏览器上查到的那一个。")

	v, r, s := signed.RawSignatureValues()
	fmt.Printf("\n  V = %s\n", v)
	fmt.Printf("  R = %s\n", r)
	fmt.Printf("  S = %s\n", s)
	fmt.Println("  ⚠️ EIP-1559 交易里的 V 只有 0 或 1（叫 yParity），")
	fmt.Println("     不是 Legacy 的 27/28，也不是 EIP-155 的 37/38。")
	fmt.Println("     这里 V=0，表示公钥 Y 坐标是偶数。")
	fmt.Println("     你要是按「V 应该是 27 或 28」去校验，就会把正常交易判成非法。")

	fmt.Println("\n  签名做了什么？")
	fmt.Println("    ① 把交易各字段 RLP 编码，前面加上 0x02（类型字节）")
	fmt.Println("    ② Keccak256 得到「签名哈希」")
	fmt.Println("    ③ 用私钥做 secp256k1 ECDSA 签名 → 得到 (r, s, yParity)")
	fmt.Println("    ④ 把 (yParity, r, s) 追加进 RLP，得到最终字节")
	fmt.Println("    广播时发的就是第 ④ 步的字节 —— 私钥自始至终没离开本机。")

	// ── 示例 4：从签名恢复发送者 ─────────────────────────────────────────
	fmt.Println("\n【示例 4】从签名恢复发送者（链上就是这么验证你的）")
	// 以太坊没有「账户余额表」这种传统数据库的账户体系：
	// 链上只存 (nonce, balance, code, storage)，**没有公钥**。
	// 验证者要确认「这笔交易确实是 X 授权的」，只能从签名里**恢复**出公钥。
	// 恢复出来的地址就是这笔交易的 from —— 这也是「地址即身份」的由来。
	recovered, err := types.Sender(signer, signed)
	if err != nil {
		fmt.Println("  恢复失败：", err)
		return
	}
	fmt.Printf("  私钥对应地址      : %s\n", from.Hex())
	fmt.Printf("  Sender 恢复的地址 : %s\n", recovered.Hex())
	fmt.Printf("  两者一致          : %t\n", recovered == from)
	fmt.Println("  → 这就是「签名 = 授权」的全部含义：")
	fmt.Println("     任何人都能广播这笔交易，但只有私钥持有者能造出这个签名。")

	// ── 示例 5：链 ID 保护 —— 防重放 ─────────────────────────────────────
	fmt.Println("\n【示例 5】链 ID 保护：同一个签名不能在别的链上用")
	// 没有链 ID 时，一笔在以太坊主网签的交易可以被原样搬到 BSC 上重放，
	// 把你在 BSC 上的钱也转走 —— 这就是 2016 年 ETC 分叉后的重放攻击。
	// EIP-155 的解法：把 chainID 塞进签名哈希，链 ID 不同 → 签名校验不过。
	wrongSigner := types.LatestSignerForChainID(big.NewInt(56)) // BSC
	fmt.Printf("  正确 signer 的 chainID: %s\n", signer.ChainID())
	fmt.Printf("  错误 signer 的 chainID: %s\n", wrongSigner.ChainID())

	badSender, err := types.Sender(wrongSigner, signed)
	fmt.Printf("  用 chainID=56 恢复 chainID=1 的交易：\n")
	fmt.Printf("    地址 = %s\n", badSender.Hex())
	fmt.Printf("    错误 = %v\n", err)
	fmt.Println("  → 恢复失败时返回的是零地址 0x000...000，**必须检查 err**。")
	fmt.Println("    如果你的代码只看地址不看 err，就会把零地址当成发送者，")
	fmt.Println("    然后去查余额、记账、对账 —— 全都是错的。")

	// 反向验证：signer 的链 ID 和 tx 自带的 ChainID 不一致时，签名直接失败。
	_, err = types.SignTx(tx, wrongSigner, key)
	fmt.Printf("\n  反向：用 chainID=56 的 signer 去签 ChainID=1 的交易 → %v\n", err)
	fmt.Println("  注意 DynamicFeeTx 自己带 ChainID 字段，signer 会拿它俩比对，")
	fmt.Println("  不一致就拒绝 —— 这样你不可能「不小心签出一笔跨链可重放的交易」。")

	// 顺带看一下「不传 chainID」会发生什么。
	noChain := types.LatestSignerForChainID(nil)
	fmt.Printf("\n  LatestSignerForChainID(nil) → %T\n", noChain)
	fmt.Println("  ↑ 退化成 HomesteadSigner，没有 EIP-155 保护，生产代码绝不能用。")

	// ── 示例 6：LegacyTx vs DynamicFeeTx ─────────────────────────────────
	fmt.Println("\n【示例 6】LegacyTx 与 DynamicFeeTx 的差别")
	legacy := types.NewTx(&types.LegacyTx{
		Nonce:    7,
		GasPrice: gwei(30), // Legacy 只有一个 GasPrice，没有 tip/feeCap 之分
		Gas:      21000,
		To:       &to,
		Value:    wei(1),
	})
	legacySigned, err := types.SignTx(legacy, signer, key)
	if err != nil {
		fmt.Println("  Legacy 签名失败：", err)
		return
	}
	fmt.Printf("  %-14s type=%d  hash=%s\n", "LegacyTx", legacySigned.Type(), legacySigned.Hash().Hex())
	fmt.Printf("  %-14s type=%d  hash=%s\n", "DynamicFeeTx", signed.Type(), signed.Hash().Hex())

	dynBin, err := signed.MarshalBinary()
	if err != nil {
		fmt.Println("  MarshalBinary 失败：", err)
		return
	}
	legacyBin, err := legacySigned.MarshalBinary()
	if err != nil {
		fmt.Println("  MarshalBinary 失败：", err)
		return
	}
	fmt.Println("\n  广播时发的字节（MarshalBinary 的结果）：")
	fmt.Printf("    DynamicFeeTx: %d 字节，首字节 0x%02x（=2，EIP-2718 类型化交易的标志）\n",
		len(dynBin), dynBin[0])
	fmt.Printf("    LegacyTx    : %d 字节，首字节 0x%02x（>=0xc0，RLP list 的前缀）\n",
		len(legacyBin), legacyBin[0])
	fmt.Printf("    未签名的 DynamicFeeTx: %d 字节（签名的多了 %d 字节 = V/R/S）\n",
		mustBinary(tx), len(dynBin)-mustBinary(tx))

	fmt.Printf("\n  DynamicFeeTx 完整 RLP（%d 字节）:\n    %s\n", len(dynBin), hex.EncodeToString(dynBin))
	fmt.Println("  手工拆一遍（上面这串 hex 从左往右读）：")
	fmt.Println("    02              ← 交易类型（EIP-2718 的类型前缀，Legacy 没有这个字节）")
	fmt.Println("    f8 73           ← RLP list 头，说明后面还有 0x73 = 115 字节")
	fmt.Println("    01              ← ChainID = 1")
	fmt.Println("    07              ← Nonce = 7")
	fmt.Println("    84 59682f00     ← GasTipCap = 0x59682f00 = 1500000000 wei = 1.5 gwei")
	fmt.Println("    85 06fc23ac00   ← GasFeeCap = 0x06fc23ac00 = 30000000000 wei = 30 gwei")
	fmt.Println("    82 5208         ← Gas = 0x5208 = 21000")
	fmt.Println("    94 7099...79c8  ← To（0x94 = 0x80 + 20，表示后面是 20 字节）")
	fmt.Println("    88 0de0b6b3a7640000 ← Value = 10^18 wei = 1 ETH")
	fmt.Println("    80              ← Data 为空（空字节串在 RLP 里是 0x80）")
	fmt.Println("    c0              ← AccessList 为空（空列表在 RLP 里是 0xc0）")
	fmt.Println("    80              ← yParity = 0（0x80 表示整数 0）")
	fmt.Println("    a0 1ced...7a8b  ← R（32 字节，0xa0 = 0x80 + 32）")
	fmt.Println("    a0 5de0...bd9f  ← S（32 字节）")
	fmt.Println("  对照 Legacy：它没有 02 前缀、没有 ChainID 字段、只有 1 个 GasPrice，")
	fmt.Println("  但 V 不是一个字节的 yParity，而是 EIP-155 的 chainID*2+35/36（主网 37/38）。")

	fmt.Println("\n  两者怎么选？")
	fmt.Println("    · 2021-08 London 之后，主网上的新交易基本都是 1559。")
	fmt.Println("    · 但一些老链 / L2 / 私有链还没支持 1559，只能用 Legacy。")
	fmt.Println("    · 机器人的做法：先 eth_getBlockByNumber 看区块头有没有 baseFeePerGas，")
	fmt.Println("      有就发 1559，没有就退回 Legacy。")

	// ── 示例 7：GasFeeCap 必须 >= GasTipCap ──────────────────────────────
	fmt.Println("\n【示例 7】GasFeeCap 与 GasTipCap 的关系校验")
	type feeCase struct {
		tipCap, feeCap *big.Int
		baseFee        *big.Int
	}
	cases := []feeCase{
		{tipCap15gwei(), gwei(30), gwei(20)}, // 正常：feeCap > tip + base
		{tipCap15gwei(), gwei(30), gwei(40)}, // 拥堵：feeCap < tip + base，但仍 > base
		{gwei(2), gwei(1), gwei(20)},         // 写反了：feeCap < tipCap
	}
	for i, c := range cases {
		t := types.NewTx(&types.DynamicFeeTx{
			ChainID: chainID, Nonce: 7,
			GasTipCap: c.tipCap, GasFeeCap: c.feeCap,
			Gas: 21000, To: &to, Value: wei(1),
		})
		effTip := t.EffectiveGasTipValue(c.baseFee)
		// 实际单价 = min(feeCap, baseFee + tipCap)
		effPrice := new(big.Int).Add(c.baseFee, c.tipCap)
		if effPrice.Cmp(c.feeCap) > 0 {
			effPrice = c.feeCap
		}
		status := "可打包"
		if effTip.Sign() < 0 {
			status = "❌ 永远不会被打包（feeCap < baseFee）"
		}
		if c.feeCap.Cmp(c.tipCap) < 0 {
			status = "❌ 字段写反（feeCap < tipCap）"
		}
		fmt.Printf("  [%d] tip=%s gwei fee=%s gwei base=%s gwei\n", i,
			toGwei(c.tipCap), toGwei(c.feeCap), toGwei(c.baseFee))
		fmt.Printf("      实际小费=%s gwei，实际单价=%s gwei → %s\n",
			toGwei(effTip), toGwei(effPrice), status)
	}
	// 动手验证思考题 3：feeCap < tipCap 时 SignTx 会不会报错？
	badFeeTx := types.NewTx(&types.DynamicFeeTx{
		ChainID: chainID, Nonce: 7,
		GasTipCap: gwei(2), GasFeeCap: gwei(1), // 故意写反
		Gas: 21000, To: &to, Value: wei(1),
	})
	badSigned, badErr := types.SignTx(badFeeTx, signer, key)
	fmt.Printf("\n  故意构造 feeCap(1 gwei) < tipCap(2 gwei) 并签名：\n")
	fmt.Printf("    SignTx 返回的错误 = %v\n", badErr)
	fmt.Printf("    签名后 tx.Hash() = %s（照样签成功了）\n", badSigned.Hash().Hex())
	fmt.Println("  → SignTx **不报错**。它只负责密码学正确性，不做经济合理性校验。")
	fmt.Println("    所以这类校验必须写在你的业务代码里，不能指望库帮你兜底。")

	fmt.Println("\n  ⚠️ types.SignTx **不会**帮你校验 feeCap >= tipCap。")
	fmt.Println("     它只管签名，不管经济合理性 —— 这是 EVM 库的设计哲学：")
	fmt.Println("     库只保证「编码/签名正确」，业务校验必须你自己做。")
	fmt.Println("     所以生产代码里签名之前一定要自己加这一道校验。")

	// 最坏情况成本
	fmt.Println("\n  最坏情况成本（gas 额度 × 单价上限）：")
	dynWorst := new(big.Int).Mul(tx.GasFeeCap(), new(big.Int).SetUint64(tx.Gas()))
	legacyWorst := new(big.Int).Mul(legacy.GasPrice(), new(big.Int).SetUint64(legacy.Gas()))
	fmt.Printf("    DynamicFeeTx: 21000 × 30 gwei = %s wei ≈ %s ETH\n", dynWorst, weiToEther(dynWorst))
	fmt.Printf("    LegacyTx    : 21000 × 30 gwei = %s wei ≈ %s ETH\n", legacyWorst, weiToEther(legacyWorst))
	fmt.Printf("    实际成本（tx.Cost()）= %s wei（含转账金额 1 ETH）\n", signed.Cost())

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n【关键总结】")
	fmt.Println("  DynamicFeeTx = ChainID + Nonce + GasTipCap + GasFeeCap + Gas + To + Value + Data")
	fmt.Println("  实际单价 = min(GasFeeCap, baseFee + GasTipCap)")
	fmt.Println("  要能被打包：GasFeeCap >= GasTipCap + baseFee；TipCap 必须 <= FeeCap")
	fmt.Println("  签名三步：LatestSignerForChainID → SignTx → 得到 (yParity, R, S)")
	fmt.Println("  交易哈希 = 签名后交易的 Keccak256 = signedTx.Hash()（未签名的不算数）")
	fmt.Println("  types.Sender(signer, signedTx) 从签名恢复发送者；失败返回零地址 + err")
	fmt.Println("  链 ID 保护：chainID 参与签名，跨链重放会因 chainID 不匹配而失败")
	fmt.Println("  广播字节 = MarshalBinary()；1559 以 0x02 开头，Legacy 以 RLP list 前缀开头")
	fmt.Println("  ⚠️ 金额全程用 big.Int / decimal，绝不能用 float64")
	fmt.Println("  对照：PHP web3.php 用 `Transaction` + `eth_account` 签名；")
	fmt.Println("        Java web3j 用 `RawTransactionManager` + `TransactionEncoder.signMessage`；")
	fmt.Println("        Go 用 `types.SignTx` + `types.LatestSignerForChainID`")
}

// toGwei 把 wei 显示成 gwei，保留 2 位小数，纯为打印好看。
// 注意这是**显示层**的浮点运算 —— 金额计算本身全程是 big.Int。
func toGwei(v *big.Int) string {
	f := new(big.Float).Quo(new(big.Float).SetInt(v), new(big.Float).SetFloat64(1e9))
	return f.Text('f', 2)
}

// weiToEther 把 wei 显示成 ETH。
func weiToEther(v *big.Int) string {
	f := new(big.Float).Quo(new(big.Float).SetInt(v), new(big.Float).SetFloat64(1e18))
	return f.Text('f', 8)
}

// mustBinary 返回交易的 RLP 编码长度，出错 panic（用于打印对比）。
func mustBinary(t *types.Transaction) int {
	b, err := t.MarshalBinary()
	if err != nil {
		panic(fmt.Sprintf("MarshalBinary 失败: %v", err))
	}
	return len(b)
}
