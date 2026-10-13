//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 07-web3-trading && go run 07_abi_events.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

/*
═══════════════════════════════════════════════════════════════════════
W-07：事件与日志解析 —— 交易机器人真正的数据源
═══════════════════════════════════════════════════════════════════════

【核心概念】

	合约的「函数返回值」只有发起交易的人能看到（eth_call 不算链上状态）。
	如果你想知道别人在链上干了什么，唯一的路子是**读事件日志（Log）**。

	一条 Log 长这样：

	    type Log struct {
	        Address common.Address   // 哪个合约发出的
	        Topics  []common.Hash    // 索引字段，最多 4 个（含 topic0）
	        Data    []byte           // 非索引字段，ABI 编码
	    }

	Event 声明的写法：

	    event Transfer(address indexed from, address indexed to, uint256 value)

	                      ↑ indexed            ↑ indexed          ↑ 没有 indexed

	- **topic0 永远是事件签名的哈希**：Keccak256("Transfer(address,address,uint256)")
	- 带 `indexed` 的参数，按声明顺序依次进 topics[1]、topics[2]、topics[3]
	- 不带 `indexed` 的参数，按 ABI 编码进 data（W-06 学的就是这套编码）

	为什么要分 topics 和 data？两个理由：
	  ① topics 可以**过滤**。RPC 的 eth_getLogs 支持 topics 数组做精确匹配，
	     节点能直接按 topic 建索引，扫「某个地址的所有转账」非常快。
	     data 是字节串，没法做等值过滤。
	  ② topics 数量有上限（4 个），data 理论上无限。所以只把「需要过滤」的字段
	     放 topics，其余都塞 data 省 gas。

【你会看到】

	✓ 手算的 topic0 和 abi.Events["Transfer"].ID 完全一致
	✓ 手工解析（左填充截断）与 abi 包解析结果逐字段一致
	✓ 同一条日志两种解法：手工 15 行 vs abi 包 5 行
	✓ indexed 的 address / uint256 在 topic 里是「左填充 32 字节」，能还原
	✓ indexed 的 string / bytes / 数组在 topic 里是「哈希」，**永远还原不回来**

【思考题】（动手改代码验证，别空想）

 1. 把 Transfer 的 `to` 改成非 indexed（去掉 indexed），重新构造日志。
    现在 topics 有几个？data 有多少字节？手工解析该怎么写？
    **如果 ERC20 真的这么设计，机器人扫转账会不会更难？**为什么？

 2. 示例 4 演示了 indexed string 只能拿到哈希。那么问题来了：
    一个 DeFi 合约把 `string symbol` 设成 indexed，
    你想读「这个事件里的 symbol 是什么」，有没有办法？
    （提示：只有一种办法 —— 穷举猜。把猜测值哈希后和 topic 比对。
    这也是为什么 indexed 只该用在 address/uint256 这类低熵字段上。）

 3. 把 abi.ParseTopicsIntoMap 的第二个参数从「只含 indexed 的 Arguments」
    改成「全部 Inputs」，看报什么错。为什么库要这么设计？
    （提示：读一下 go-ethereum 的 accounts/abi/topics.go。）

 4. topic0 是 4 字节选择器吗？还是 32 字节？
    为什么函数调用只取 4 字节，事件签名却要整个 32 字节？
    （提示：函数调用有 calldata 可以浪费，topic 是要花 gas 存链上的。）

 5. 一个合约 emit 了 5 个 indexed 参数，能编过吗？为什么？
    如果非要传第 5 个「可过滤」的值，你会怎么设计事件？

【通关标准】

	能说清 topic0 / indexed / data 三者的关系，
	能手工解析一条 ERC20 Transfer 日志，
	能用 abi.Events + UnpackIntoMap + ParseTopicsIntoMap 解析同一条日志，
	并且牢牢记住「indexed 的动态类型在 topic 里只剩哈希」。

【前置】W-06（ABI 编码）   【标签】event · log · topic · indexed · eth_getLogs
═══════════════════════════════════════════════════════════════════════
*/

// erc20EventsABI 只声明 Transfer 事件，用来演示标准 ERC20 的日志结构。
//
// 注意 "anonymous": false —— 非匿名事件才有 topic0。
// 如果声明成 anonymous: true，就没有 topic0，所有 indexed 参数往前挪一位。
// 匿名事件省 gas 但没法按签名过滤，实际项目里极少用。
var erc20EventsABI = mustABI(`[
  {"anonymous":false,"name":"Transfer","type":"event",
   "inputs":[
     {"indexed":true,  "name":"from",  "type":"address"},
     {"indexed":true,  "name":"to",    "type":"address"},
     {"indexed":false, "name":"value", "type":"uint256"}
   ]}
]`)

// namedEventABI 故意把 string 设成 indexed，用来演示「动态类型进 topic 只剩哈希」。
var namedEventABI = mustABI(`[
  {"anonymous":false,"name":"Named","type":"event",
   "inputs":[
     {"indexed":true,  "name":"tag", "type":"string"},
     {"indexed":false, "name":"v",   "type":"uint256"}
   ]}
]`)

// mustABI 解析 ABI JSON，出错直接 panic（本关用的是固定字符串）。
func mustABI(js string) abi.ABI {
	parsed, err := abi.JSON(strings.NewReader(js))
	if err != nil {
		panic(fmt.Sprintf("解析 ABI 失败: %v", err))
	}
	return parsed
}

// topic0Of 手算事件签名哈希：Keccak256("事件名(类型,类型)")，取完整 32 字节。
//
// 和 W-06 的函数选择器唯一的区别：**不截断**。
// 函数只取前 4 字节（省 calldata 空间），事件要完整的 32 字节。
func topic0Of(signature string) common.Hash {
	return crypto.Keccak256Hash([]byte(signature))
}

// indexedArgs 从事件定义里筛出所有 indexed 参数。
//
// 这个筛子必须自己写，因为 abi 包的 ParseTopicsIntoMap 要求
// 传入的 Arguments **只含 indexed 字段**，且数量必须和 topics[1:] 完全相等。
// 直接传全部 Inputs 会报 "topic/field count mismatch"（示例 5 会演示）。
func indexedArgs(event abi.Event) abi.Arguments {
	var indexed abi.Arguments
	for _, arg := range event.Inputs {
		if arg.Indexed {
			indexed = append(indexed, arg)
		}
	}
	return indexed
}

// dumpLog 把一条日志的原始字节打出来，方便逐段对照解析结果。
func dumpLog(lg types.Log) {
	fmt.Printf("    合约地址 : %s\n", lg.Address.Hex())
	fmt.Printf("    topics   : %d 个\n", len(lg.Topics))
	for i, t := range lg.Topics {
		fmt.Printf("      [%d] %s\n", i, t.Hex())
	}
	fmt.Printf("    data     : %s（%d 字节）\n", hex.EncodeToString(lg.Data), len(lg.Data))
}

// buildTransferLog 构造一条「看起来像真的」ERC20 Transfer 日志。
//
// 为什么要自己构造而不是连节点？因为本关要离线可跑、结果可复现。
// 字段值全部照着实链的真实布局摆，一个字节都不差。
func buildTransferLog() types.Log {
	from := common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266")
	to := common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8")
	value := big.NewInt(1000000) // 1 USDC（USDC 是 6 位小数）

	return types.Log{
		// USDC 合约地址，纯属装饰 —— 日志本身不关心合约是谁。
		Address: common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"),
		Topics: []common.Hash{
			topic0Of("Transfer(address,address,uint256)"),
			// indexed 的 address：20 字节**左填充**到 32 字节
			common.BytesToHash(common.LeftPadBytes(from.Bytes(), 32)),
			common.BytesToHash(common.LeftPadBytes(to.Bytes(), 32)),
		},
		// 非 indexed 的 uint256：ABI 编码就是左填充到 32 字节
		Data: common.LeftPadBytes(value.Bytes(), 32),
	}
}

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 W-07 事件与日志解析")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// ── 示例 1：topic0 —— 事件签名哈希 ───────────────────────────────────
	fmt.Println("📝 示例 1：topic0 = Keccak256(事件签名)，完整 32 字节")
	fmt.Println(strings.Repeat("-", 60))

	sigs := []string{
		"Transfer(address,address,uint256)",
		"Approval(address,address,uint256)",
		"Swap(address,uint256,uint256,uint256,uint256,address)",
	}
	for _, s := range sigs {
		fmt.Printf("  %-53s → %s\n", s, topic0Of(s).Hex())
	}

	transfer := erc20EventsABI.Events["Transfer"]
	fmt.Printf("\n  abi 包给的 Transfer.ID            → %s\n", transfer.ID.Hex())
	fmt.Printf("  手工算的 topic0                   → %s\n", topic0Of("Transfer(address,address,uint256)").Hex())
	fmt.Printf("  两者一致                          → %t\n", transfer.ID == topic0Of("Transfer(address,address,uint256)"))
	fmt.Println("\n  ⚠️ 和函数选择器的区别：")
	fmt.Println("     函数调用取 Keccak 的**前 4 字节**（省 calldata，调用方付钱）")
	fmt.Println("     事件签名取**全部 32 字节**（topic 存链上，4 字节碰撞风险太高）")

	// 看看有多少个 indexed 参数 —— 它决定了 topics 的长度。
	fmt.Printf("\n  Transfer 的 indexed 参数个数 = %d → topics 应该有 %d 个（含 topic0）\n",
		len(indexedArgs(transfer)), len(indexedArgs(transfer))+1)

	// ── 示例 2：构造一条日志 ─────────────────────────────────────────────
	fmt.Println("📌 示例 2：构造一条 ERC20 Transfer 日志")
	fmt.Println(strings.Repeat("-", 60))
	lg := buildTransferLog()
	dumpLog(lg)
	fmt.Println("\n  布局对照（记住这张图，本关就通了）：")
	fmt.Println("    topics[0] = 事件签名哈希            ← 32B，永远是哈希")
	fmt.Println("    topics[1] = from（indexed address）  ← 32B，左填充 12 个 0 + 20B 地址")
	fmt.Println("    topics[2] = to  （indexed address）  ← 32B，左填充 12 个 0 + 20B 地址")
	fmt.Println("    data      = value（非 indexed）      ← 32B，就是 W-06 学的 ABI 编码")

	// ── 示例 3：手工解析 —— 一个字节一个字节地抠 ─────────────────────────
	fmt.Println("🔍 示例 3：手工解析：只用 LeftPad 的逆运算")
	fmt.Println(strings.Repeat("-", 60))
	// topics[1] 是 32 字节，地址只有 20 字节。还原方式：取**后 20 字节**。
	// common.BytesToAddress 做的就是这个（它是 [20]byte 的定长类型，会截取尾部）。
	manualFrom := common.BytesToAddress(lg.Topics[1].Bytes())
	manualTo := common.BytesToAddress(lg.Topics[2].Bytes())
	// data 是 32 字节大端整数，直接 SetBytes。
	manualValue := new(big.Int).SetBytes(lg.Data)

	fmt.Printf("  from  : topics[1] 后 20 字节 → %s\n", manualFrom.Hex())
	fmt.Printf("  to    : topics[2] 后 20 字节 → %s\n", manualTo.Hex())
	fmt.Printf("  value : data 直接当大端整数 → %s\n", manualValue)

	// 用字符串表达「怎么抠出来的」，方便对照其他语言。
	fmt.Printf("\n  topics[1] 原始 = %s\n", hex.EncodeToString(lg.Topics[1].Bytes()))
	fmt.Printf("                    ↑ 前 12 字节全 0（填充）      ↑ 后 20 字节才是地址\n")
	fmt.Printf("  data      原始 = %s\n", hex.EncodeToString(lg.Data))
	fmt.Printf("                    ↑ 这就是 uint256 的 1000000，右端对齐、左侧补 0\n")

	// ── 示例 4：abi 包解析 —— 同一个答案，更少的代码 ─────────────────────
	fmt.Println("🧪 示例 4：abi 包解析：结果必须和手工解析一致")
	fmt.Println(strings.Repeat("-", 60))
	// 第一步：解 data（非 indexed 部分）。
	// UnpackIntoMap 的语义是「按事件定义解 data，把结果塞进 map」。
	fields := map[string]interface{}{}
	if err := erc20EventsABI.UnpackIntoMap(fields, "Transfer", lg.Data); err != nil {
		fmt.Println("  UnpackIntoMap 失败：", err)
		return
	}
	fmt.Printf("  UnpackIntoMap(只解 data) → %v\n", fields)

	// 第二步：解 topics[1:]（indexed 部分）。
	// ⚠️ 第二个参数必须**只含 indexed 的 Arguments**，数量要等于 len(topics[1:])。
	if err := abi.ParseTopicsIntoMap(fields, indexedArgs(transfer), lg.Topics[1:]); err != nil {
		fmt.Println("  ParseTopicsIntoMap 失败：", err)
		return
	}
	fmt.Printf("  ParseTopicsIntoMap(topics[1:]) → from=%v to=%v value=%v\n",
		fields["from"], fields["to"], fields["value"])

	// 逐字段和手工解析比对 —— 这是本关最重要的一个断言。
	abiFrom, okFrom := fields["from"].(common.Address)
	abiTo, okTo := fields["to"].(common.Address)
	abiValue, okValue := fields["value"].(*big.Int)
	fmt.Println("\n  手工解析 vs abi 包解析：")
	fmt.Printf("    from  %s == %s → %t\n", manualFrom.Hex(), abiFrom.Hex(), okFrom && manualFrom == abiFrom)
	fmt.Printf("    to    %s == %s → %t\n", manualTo.Hex(), abiTo.Hex(), okTo && manualTo == abiTo)
	fmt.Printf("    value %s == %s → %t\n", manualValue, abiValue, okValue && manualValue.Cmp(abiValue) == 0)
	fmt.Println("  → 完全一致。abi 包没做什么魔法，它只是把上面那 3 行手工代码")
	fmt.Println("    加上「类型检查 + 错误处理 + 动态类型分支」后包了起来。")

	// 真实机器人里通常还会先做一次 topic0 过滤，避免把别的事件当 Transfer 解。
	fmt.Printf("\n  过滤：lg.Topics[0] == transfer.ID → %t\n", lg.Topics[0] == transfer.ID)

	// ── 示例 5：indexed 的动态类型 —— 最容易搞错的地方 ───────────────────
	fmt.Println("⚙️ 示例 5：indexed 的 string / bytes / 数组在 topic 里只剩哈希")
	fmt.Println(strings.Repeat("-", 60))
	named := namedEventABI.Events["Named"]
	tagHash := topic0Of("hello") // 合约会把 "hello" 哈希后放进 topic
	lg2 := types.Log{
		Topics: []common.Hash{named.ID, tagHash},
		Data:   common.LeftPadBytes(big.NewInt(9).Bytes(), 32),
	}
	fmt.Printf("  合约 emit Named(\"hello\", 9)，topic 里存的是：%s\n", tagHash.Hex())
	fmt.Printf("  它正好等于 Keccak256(\"hello\")                ：%t\n", tagHash == topic0Of("hello"))

	fields2 := map[string]interface{}{}
	if err := namedEventABI.UnpackIntoMap(fields2, "Named", lg2.Data); err != nil {
		fmt.Println("  UnpackIntoMap 失败：", err)
		return
	}
	if err := abi.ParseTopicsIntoMap(fields2, indexedArgs(named), lg2.Topics[1:]); err != nil {
		fmt.Println("  ParseTopicsIntoMap 失败：", err)
		return
	}
	fmt.Printf("  ParseTopicsIntoMap 解出来的 tag：%v（类型 %T）\n", fields2["tag"], fields2["tag"])
	fmt.Printf("  解出来的 v（非 indexed，正常）    ：%v\n", fields2["v"])
	fmt.Println("\n  ⚠️ 注意 tag 的类型是 common.Hash，不是 string。")
	fmt.Println("     因为哈希是单向的，库里也变不回来 —— 它只能把 topic 原样给你。")
	fmt.Println("     所以：indexed 的 address / uint256 / bool / bytesN 可以还原")
	fmt.Println("     （它们是定长且可逆的），但 string / bytes / 数组 / tuple")
	fmt.Println("     进了 topic 就只剩哈希，**永远拿不回原值**。")
	fmt.Println("     这就是为什么 ERC20 的 Transfer 把 from/to 设成 indexed（地址可还原），")
	fmt.Println("     却没人把 symbol 设成 indexed（字符串还原不回来）。")

	// 反面演示：直接把全部 Inputs 传给 ParseTopicsIntoMap 会怎样。
	fmt.Println("\n  反面演示（故意传错）：")
	errAll := abi.ParseTopicsIntoMap(map[string]interface{}{}, named.Inputs, lg2.Topics[1:])
	fmt.Printf("    传全部 Inputs（2 个字段）给 topics[1:]（1 个 topic）→ %v\n", errAll)
	fmt.Println("    → 必须自己先筛出 indexed 字段。go-ethereum 内部（bind/v2）")
	fmt.Println("      也是这么做的：它遍历 Inputs 过滤 Indexed 后才调用这个函数。")

	// ── 示例 6：为什么 from/to 在 topics，value 在 data ──────────────────
	fmt.Println("🎯 示例 6：为什么 ERC20 的 Transfer 要这么设计")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("  event Transfer(address indexed from, address indexed to, uint256 value)")
	fmt.Println("\n  from / to 放 topics（indexed）：")
	fmt.Println("    · 可以过滤。机器人最常问的两个问题是「某个地址收过什么钱」和")
	fmt.Println("      「某个地址付过什么钱」，对应 topics[1]=X 或 topics[2]=X。")
	fmt.Println("      节点按 topic 建了索引，eth_getLogs 能直接命中，不用扫全链。")
	fmt.Println("    · 地址是定长 20 字节，能还原，不损失信息。")
	fmt.Println("  value 放 data（非 indexed）：")
	fmt.Println("    · 没人会去过滤「金额等于 1000000 的所有转账」—— 这种查询没意义。")
	fmt.Println("    · 每个 indexed 参数要多花约 375 gas（LOG3 指令 + 存储）。")
	fmt.Println("      省下的 gas 由全网的转账方分摊，所以能省就省。")
	fmt.Println("\n  gas 账（EVM 的 LOGn 指令）：")
	fmt.Println("    · 每个 topic 约 375 gas，每个字节的 data 约 8 gas")
	fmt.Println("    · topics 最多 4 个（LOG0~LOG4），data 无上限")
	fmt.Println("    · 所以规则是：**要过滤的进 topics，其余进 data**")

	// ── 示例 7：日志里还有什么 ───────────────────────────────────────────
	fmt.Println("📦 示例 7：真实 RPC 返回的日志还带这些字段")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("  type Log struct {")
	fmt.Println("      Address common.Address  // 发事件的合约（本关示例是 USDC 地址）")
	fmt.Println("      Topics  []common.Hash   // 本关重点")
	fmt.Println("      Data    []byte          // 本关重点")
	fmt.Println("      BlockNumber uint64      // 在哪个区块 —— 重组时用来回滚")
	fmt.Println("      TxHash      common.Hash // 哪笔交易触发的")
	fmt.Println("      TxIndex     uint        // 交易在区块里的位置")
	fmt.Println("      BlockHash   common.Hash // 区块哈希")
	fmt.Println("      Index       uint        // 这条日志在区块里的序号")
	fmt.Println("      Removed     bool        // ⚠️ true = 这条日志因链重组被撤销了")
	fmt.Println("  }")
	fmt.Println("  ⚠️ Removed 字段是机器人必须处理的：链重组会让已确认的日志消失。")
	fmt.Println("     典型做法是「等 N 个区块确认」再入账（N 取 12 左右，视链而定）。")
	fmt.Println("  本关的日志是手工构造的，所以这些字段全是零值 —— 不影响解析逻辑。")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  Log = { Address, Topics[], Data, BlockNumber, TxHash, ... }")
	fmt.Println("  topic0        = Keccak256(\"事件名(类型,类型)\")  ← 完整 32 字节，不是 4 字节")
	fmt.Println("  topics[1..3]  = 按声明顺序排列的 indexed 参数")
	fmt.Println("  data          = 非 indexed 参数，走 W-06 那套 ABI 编码")
	fmt.Println("  indexed 定长类型（address/uint256/bool/bytesN）→ 可还原，取后 N 字节")
	fmt.Println("  indexed 动态类型（string/bytes/数组/tuple）  → 只剩 Keccak256 哈希，不可还原")
	fmt.Println("  解析套路：UnpackIntoMap(解 data) + ParseTopicsIntoMap(解 topics[1:])")
	fmt.Println("            第二个参数必须是自己筛出来的「只含 indexed 的 Arguments」")
	fmt.Println("  对照：PHP web3.php 用 `Contract::decodeEventData`；")
	fmt.Println("        Java web3j 用 `EthFilter` + `EventDecoder.decode`；")
	fmt.Println("        Go 用 `abi.Events[name]` + UnpackIntoMap + ParseTopicsIntoMap")
}
