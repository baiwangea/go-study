//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 07-web3-trading && go run 06_abi_encode.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

/*
═══════════════════════════════════════════════════════════════════════
W-06：ABI 编解码 —— 和合约对话的语言
═══════════════════════════════════════════════════════════════════════

【核心概念】

	以太坊合约是一台跑在 EVM 上的程序。你调用它，靠的是往一笔交易的 data
	字段里塞一段字节。这段字节的格式就是 ABI 编码。

	一次调用的 data 永远是三段拼接：

	    ┌──────────┬────────────────────────────────────────┐
	    │ 选择器 4B │            参数区（32 字节对齐）        │
	    └──────────┴────────────────────────────────────────┘

	选择器（selector）= Keccak256("函数名(参数类型1,参数类型2,...)")[:4]

	注意签名串的细节：**类型之间不能有空格，参数名不参与**。
	    ✓ balanceOf(address)
	    ✓ transfer(address,uint256)
	    ✗ balanceOf(address a)
	    ✗ transfer(address, uint256)     ← 有空格，算出来的选择器是错的

	参数区有两种编码：
	    静态类型（uint/int/address/bool/bytesN）→ 直接占 1 个 32 字节槽
	    动态类型（string/bytes/数组）          → 槽里放「偏移量」，真正的数据在后面

	这就是 ABI 里最反直觉的地方：**动态参数的值不在它自己的槽里，
	槽里只是一个指针**。头（head）全是 32 字节槽，尾（tail）才是实际内容。

【你会看到】

	✓ 手算的 selector 和 abi.JSON 的 Pack 结果逐字节一致（70a08231）
	✓ balanceOf 的 calldata = 4 字节选择器 + 32 字节地址 = 36 字节
	✓ transfer 的 calldata = 4 + 32 + 32 = 68 字节，全部静态
	✓ setName(string) 的 calldata = 4 + 32(偏移) + 32(长度) + 32(内容) = 100 字节
	✓ 三个参数里夹一个 string 时，偏移量是 0x60 而不是 0x20 —— 因为头有三槽
	✓ abi.Arguments 手工编解码，结果和 abi.JSON 走同一套规则

【思考题】（动手改代码验证，别空想）

 1. 把 balanceOf 的签名改成 "balanceOf(address a)"（带参数名）再算 selector。
    和 70a08231 一样吗？为什么 ABI 规范要排除参数名？
    （提示：参数名只是给人看的，改个名不该让链上字节变化。）

 2. 把 setName 的入参从 "abc" 换成 "abcd"，calldata 长度变成多少？
    再换成 32 个字符、33 个字符，长度各是多少？
    总结公式：动态类型的长度 = 32(偏移) + 32(长度) + ceil(n/32)*32。

 3. 示例 6 里 f(uint256,string,uint256) 的偏移量是 0x60 = 96 = 3×32。
    现在把签名改成 f(string,uint256)，偏移量该是多少？
    先自己算，再改代码验证。**为什么偏移量取决于「头有几个槽」，而不是
    「前面有几个动态参数」？**

 4. 用 abi.JSON 解析一份真实 ERC20 的完整 ABI（transfer/approve/
    balanceOf/allowance/totalSupply/decimals 六个函数），
    把每个函数的选择器都算出来。它们分别是多少？

 5. 为什么 selector 只有 4 字节？碰撞概率有多大？
    （提示：生日问题，2^32 空间里约 7.7 万个函数就有 50% 概率撞。
    现实中真撞过：`transfer(address,uint256)` 和某些冷门函数签名撞过。
    这也是为什么「只靠 4 字节 selector 调合约」是有风险的。）

【通关标准】

	能手算任意函数的选择器，能解释 calldata 的头尾结构与偏移量规则，
	能说清静态类型和动态类型的编码差别，
	并且知道 abi.JSON 的 Pack/Unpack 背后做了什么。

【前置】W-05（keccak 与 hex）   【标签】selector · ABI · calldata · Pack · Unpack
═══════════════════════════════════════════════════════════════════════
*/

// erc20ABI 是一份最小 ERC20 ABI，只留 balanceOf 和 transfer 两个函数。
//
// 为什么要手写 JSON 而不是引入一份完整 ABI？因为 ABI 就是 JSON 文本，
// 看懂它长什么样，你才知道 abigen 生成的那些代码是从哪来的。
var erc20ABI = mustABI(`[
  {"name":"balanceOf","type":"function",
   "inputs":[{"name":"a","type":"address"}],
   "outputs":[{"name":"","type":"uint256"}]},
  {"name":"transfer","type":"function",
   "inputs":[{"name":"to","type":"address"},{"name":"amount","type":"uint256"}],
   "outputs":[{"name":"","type":"bool"}]}
]`)

// nameABI 用来演示动态类型 string 的编码。
var nameABI = mustABI(`[
  {"name":"setName","type":"function",
   "inputs":[{"name":"n","type":"string"}],
   "outputs":[]}
]`)

// mixABI 用来演示「静态 + 动态 + 静态」混合时的偏移量。
var mixABI = mustABI(`[
  {"name":"f","type":"function",
   "inputs":[{"name":"a","type":"uint256"},{"name":"s","type":"string"},{"name":"b","type":"uint256"}],
   "outputs":[]}
]`)

// mustABI 解析 ABI JSON，出错直接 panic。
// 生产代码里应该返回 error（ABI 可能来自外部文件），但本关是固定字符串。
func mustABI(js string) abi.ABI {
	parsed, err := abi.JSON(strings.NewReader(js))
	if err != nil {
		panic(fmt.Sprintf("解析 ABI 失败: %v", err))
	}
	return parsed
}

// selectorOf 手算函数选择器：Keccak256(签名)[:4]。
//
// 这是本关的核心 —— 不调用任何 ABI 库，只用哈希函数。
// 理解了这 5 行，你就理解了「为什么合约地址 + calldata 就能调合约」。
func selectorOf(signature string) []byte {
	full := crypto.Keccak256([]byte(signature))
	return full[:4]
}

// chunk 把「一次函数调用」的 calldata 按 32 字节（64 个 hex 字符）切段打印，
// 用来肉眼观察头尾结构。第一段单独处理，因为它是 4 字节选择器。
func chunk(label string, data []byte) {
	fmt.Printf("  %s（共 %d 字节）\n", label, len(data))
	if len(data) < 4 {
		fmt.Printf("    原始 hex: %s\n", hex.EncodeToString(data))
		return
	}
	fmt.Printf("    [选择器 4B ] %s\n", hex.EncodeToString(data[:4]))
	chunkArgs(data[4:])
}

// chunkArgs 把「不含选择器的裸参数区」按 32 字节切段打印。
//
// 单独写一个函数而不是复用 chunk，是因为 abi.Arguments.Pack 的结果里
// **没有选择器** —— 那不是一次函数调用，只是参数的 ABI 编码。
// 如果拿 chunk 去打印它，前 4 个字节会被当成选择器吃掉，输出就错位了。
func chunkArgs(body []byte) {
	for i := 0; i*32 < len(body); i++ {
		end := (i + 1) * 32
		if end > len(body) {
			end = len(body)
		}
		fmt.Printf("    [参数槽 %2d ] %s\n", i, hex.EncodeToString(body[i*32:end]))
	}
}

func main() {
	fmt.Println("\n=== W-06 ABI 编解码 ===")

	// ── 示例 1：函数选择器 —— 自己算一遍 ─────────────────────────────────
	fmt.Println("\n【示例 1】函数选择器 = Keccak256(签名)[:4]")

	sigs := []string{
		"balanceOf(address)",
		"transfer(address,uint256)",
		"setName(string)",
	}
	for _, s := range sigs {
		full := crypto.Keccak256([]byte(s))
		fmt.Printf("  Keccak256(%-28s) = %s\n", `"`+s+`"`, hex.EncodeToString(full))
		fmt.Printf("  %-39s → 取前 4 字节 = %s\n", "", hex.EncodeToString(full[:4]))
	}

	// 完整展示 balanceOf 的推导过程，一个字节都不省。
	sig := "balanceOf(address)"
	fmt.Printf("\n  手工推导 balanceOf(address)：\n")
	fmt.Printf("    ① 把签名写成 ASCII 字节: %s\n", hex.EncodeToString([]byte(sig)))
	fmt.Printf("    ② Keccak256 得到 32 字节 : %s\n", hex.EncodeToString(crypto.Keccak256([]byte(sig))))
	fmt.Printf("    ③ 取前 4 字节作为 selector: %s\n", hex.EncodeToString(selectorOf(sig)))

	// 对照：让 abi 包自己算，看是否一致。
	// ABI 的 Method.ID 就是 Keccak256(签名)[:4]，和我们的手工算法是同一个东西。
	manual := hex.EncodeToString(selectorOf(sig))
	fromABI := hex.EncodeToString(erc20ABI.Methods["balanceOf"].ID)
	fmt.Printf("    ④ abi 包算出来的 ID      : %s\n", fromABI)
	fmt.Printf("    ⑤ 两者一致               : %t  ← 手算没算错\n", manual == fromABI)

	fmt.Println("\n  ⚠️ 签名字符串的格式是硬规矩：")
	fmt.Println("     ✓ \"balanceOf(address)\"        → 70a08231")
	fmt.Printf("     ✗ \"balanceOf(address a)\"      → %s  ← 带参数名就全错了\n",
		hex.EncodeToString(selectorOf("balanceOf(address a)")))
	fmt.Printf("     ✗ \"transfer(address, uint256)\" → %s（带空格）\n",
		hex.EncodeToString(selectorOf("transfer(address, uint256)")))
	fmt.Printf("     ✓ \"transfer(address,uint256)\"  → %s\n",
		hex.EncodeToString(selectorOf("transfer(address,uint256)")))

	// ── 示例 2：Pack —— 拼出 balanceOf 的 calldata ───────────────────────
	fmt.Println("\n【示例 2】abi.Pack：拼出 calldata")
	holder := common.HexToAddress("0x0000000000000000000000000000000000000001")
	calldata, err := erc20ABI.Pack("balanceOf", holder)
	if err != nil {
		fmt.Println("  Pack 失败：", err)
		return
	}
	chunk("balanceOf(0x...0001)", calldata)
	fmt.Println("  拆解：")
	fmt.Println("    · 前 4 字节 70a08231          → 调用哪个函数")
	fmt.Println("    · 后 32 字节 000...0001        → 参数 address，左补齐到 32 字节")
	fmt.Println("  address 只有 20 字节，却占 1 个 32 字节槽：**左边补 12 个 0**。")
	fmt.Println("  这是「32 字节对齐」规则，所有静态类型都占整槽。")

	// ── 示例 3：Unpack —— 解合约返回值 ───────────────────────────────────
	fmt.Println("\n【示例 3】abi.Unpack：解返回值")
	// 合约返回的也是一段 32 字节对齐的 ABI 编码（不带选择器）。
	// 这里模拟合约返回 uint256 = 100（0x64）。
	rawReturn, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000064")
	out, err := erc20ABI.Unpack("balanceOf", rawReturn)
	if err != nil {
		fmt.Println("  Unpack 失败：", err)
		return
	}
	fmt.Printf("  原始返回值: %s\n", hex.EncodeToString(rawReturn))
	fmt.Printf("  Unpack 后 : %v（类型 %T）\n", out[0], out[0])
	fmt.Println("  ⚠️ uint256 解出来是 *big.Int，不是 uint64。")
	fmt.Println("     uint256 上限 2^256-1，远超 uint64 的 2^64-1，只能用 big.Int。")

	// ── 示例 4：静态类型 transfer(address,uint256) ───────────────────────
	fmt.Println("\n【示例 4】transfer(address,uint256)：全静态，最好懂")
	to := common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8")
	amount := big.NewInt(1000000) // 1 USDC（6 位小数）
	xfer, err := erc20ABI.Pack("transfer", to, amount)
	if err != nil {
		fmt.Println("  Pack 失败：", err)
		return
	}
	chunk("transfer(to, 1000000)", xfer)
	fmt.Println("  拆解：")
	fmt.Println("    · selector   a9059cbb")
	fmt.Printf("    · 参数 1     %s  （to，左补 12 字节 0）\n", hex.EncodeToString(xfer[4:36]))
	fmt.Printf("    · 参数 2     %s  （1000000 = 0xf4240，左补齐）\n", hex.EncodeToString(xfer[36:68]))
	fmt.Println("  两个参数都是静态类型 → 各占 1 槽，没有偏移量，没有尾巴。")
	fmt.Println("  这是最常见的转账调用，也是机器人发得最多的一笔交易。")

	// 手工拼一遍，验证「ABI 编码就是 32 字节左补齐」。
	manualXfer := hex.EncodeToString(selectorOf("transfer(address,uint256)")) +
		hex.EncodeToString(common.LeftPadBytes(to.Bytes(), 32)) +
		hex.EncodeToString(common.LeftPadBytes(amount.Bytes(), 32))
	fmt.Printf("\n  手工拼接结果与 Pack 一致: %t\n", manualXfer == hex.EncodeToString(xfer))

	// ── 示例 5：动态类型 setName(string) —— 头尾 + 偏移量 ────────────────
	fmt.Println("\n【示例 5】动态类型 string：头尾结构 + 偏移量")
	dynData, err := nameABI.Pack("setName", "abc")
	if err != nil {
		fmt.Println("  Pack 失败：", err)
		return
	}
	chunk(`setName("abc")`, dynData)
	fmt.Println("  拆解（动态类型把「值」搬到了后面的尾巴里）：")
	fmt.Printf("    · selector   %s\n", hex.EncodeToString(dynData[:4]))
	fmt.Printf("    · 槽 0       %s  ← 偏移量 0x20 = 32，意思是「值在头之后 32 字节处」\n", hex.EncodeToString(dynData[4:36]))
	fmt.Printf("    · 槽 1       %s  ← 长度 3（字符串有 3 个字节）\n", hex.EncodeToString(dynData[36:68]))
	fmt.Printf("    · 槽 2       %s  ← 内容 \"abc\" = 616263，右侧补 0 到 32 字节\n", hex.EncodeToString(dynData[68:100]))
	fmt.Println("\n  关键理解：**头里的槽放的是偏移量（指针），不是值本身**。")
	fmt.Println("  偏移量的基准点是「参数区的起点」，也就是紧跟在 4 字节选择器后面的位置。")
	fmt.Println("  这里头只有 1 个槽（32 字节），所以尾巴从偏移 32 开始 → 0x20。")

	// 手工拼一遍，逐字节对照。
	manualDyn := hex.EncodeToString(selectorOf("setName(string)")) +
		hex.EncodeToString(common.LeftPadBytes(big.NewInt(32).Bytes(), 32)) + // 偏移 0x20
		hex.EncodeToString(common.LeftPadBytes(big.NewInt(3).Bytes(), 32)) + // 长度 3
		hex.EncodeToString(common.RightPadBytes([]byte("abc"), 32)) // 内容右补 0
	fmt.Printf("\n  手工拼接结果与 Pack 一致: %t\n", manualDyn == hex.EncodeToString(dynData))
	fmt.Println("  注意内容区是**右侧**补 0（RightPad），而数值是**左侧**补 0（LeftPad）。")
	fmt.Println("  数值左补齐是为了保持大小关系；字节串右补齐是因为它本来就是「从左往右读」。")

	// ── 示例 6：混合参数 —— 偏移量取决于「头有几个槽」 ───────────────────
	fmt.Println("\n【示例 6】f(uint256,string,uint256)：偏移量 = 头槽数 × 32")
	mix, err := mixABI.Pack("f", big.NewInt(1), "hi", big.NewInt(2))
	if err != nil {
		fmt.Println("  Pack 失败：", err)
		return
	}
	chunk(`f(1, "hi", 2)`, mix)
	fmt.Printf("  槽 0  %s  ← a = 1\n", hex.EncodeToString(mix[4:36]))
	fmt.Printf("  槽 1  %s  ← s 的偏移量 0x60 = 96 = 3 × 32\n", hex.EncodeToString(mix[36:68]))
	fmt.Printf("  槽 2  %s  ← b = 2\n", hex.EncodeToString(mix[68:100]))
	fmt.Printf("  槽 3  %s  ← s 的长度 2\n", hex.EncodeToString(mix[100:132]))
	fmt.Printf("  槽 4  %s  ← s 的内容 \"hi\" = 6869\n", hex.EncodeToString(mix[132:164]))
	fmt.Println("\n  ⚠️ 最容易算错的地方：偏移量 0x60 是 **96 = 3 个槽**，")
	fmt.Println("     因为头有 3 个参数槽（a、s、b 各一个），而不是「s 前面有几个动态参数」。")
	fmt.Println("     头的大小 = 参数个数 × 32（不管静态还是动态），")
	fmt.Println("     所以偏移量永远指向「整个头之后」。")
	fmt.Println("     记住这条，ABI 编码就再也不会算错了。")

	// ── 示例 7：abi.Arguments 手工编解码 ─────────────────────────────────
	fmt.Println("\n【示例 7】abi.Arguments：绕开函数，直接操作类型")
	// abi.Arguments 是比 abi.ABI 更底层的接口：
	// 它只认「一串类型」，不管函数名和选择器。
	// 合约返回值的解析、事件 data 的解析，底层用的都是它。
	uint256Type, err := abi.NewType("uint256", "", nil)
	if err != nil {
		fmt.Println("  NewType 失败：", err)
		return
	}
	addressType, err := abi.NewType("address", "", nil)
	if err != nil {
		fmt.Println("  NewType 失败：", err)
		return
	}

	args := abi.Arguments{
		{Name: "amount", Type: uint256Type},
		{Name: "holder", Type: addressType},
	}
	packed, err := args.Pack(big.NewInt(7), holder)
	if err != nil {
		fmt.Println("  Pack 失败：", err)
		return
	}
	fmt.Printf("  Arguments.Pack(uint256(7), address(0x...01))（共 %d 字节，无选择器）\n", len(packed))
	chunkArgs(packed)

	unpacked, err := args.Unpack(packed)
	if err != nil {
		fmt.Println("  Unpack 失败：", err)
		return
	}
	fmt.Printf("  Unpack 回来: amount=%v holder=%v\n", unpacked[0], unpacked[1])
	fmt.Println("  → 一个 32 字节槽一个参数，顺序就是 Arguments 里的顺序。")
	fmt.Println("    没有选择器（那是函数才有的东西），也没有函数名。")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n【关键总结】")
	fmt.Println("  calldata = selector(4B) + 参数区")
	fmt.Println("  selector = Keccak256(\"函数名(类型,类型)\")[:4]  ← 类型间不能有空格，不含参数名")
	fmt.Println("  静态类型 → 占 1 槽；数值/地址/布尔左补齐，定长 bytesN 右补齐")
	fmt.Println("  动态类型 → 槽里放偏移量，真值在尾巴里：长度(32B) + 内容(右补齐到 32 的倍数)")
	fmt.Println("  偏移量基准 = 参数区起点；值 = 头槽数 × 32（头槽数 = 参数个数）")
	fmt.Println("  abi.JSON 的 Pack/Unpack 帮你做的，就是上面这些补齐和偏移计算")
	fmt.Println("  对照：PHP web3.php 用 `Ethabi::encodeFunctionSignature`；")
	fmt.Println("        Java web3j 用 `FunctionEncoder.encode(Function)`；")
	fmt.Println("        Go 用 `abi.JSON(...).Pack(...)` —— 三者输出完全一样（都是规范）")
}
