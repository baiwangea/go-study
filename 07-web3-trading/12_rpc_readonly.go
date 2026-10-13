//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 07-web3-trading && go run 12_rpc_readonly.go
// 依赖：github.com/ethereum/go-ethereum（ethclient / rpc）、github.com/shopspring/decimal
// 网络：需要能访问一个以太坊 RPC 节点。连不上会自动降级（打印配置方法，退出码 0）。
//       想换节点：ETH_RPC_URL=https://your-node go run 12_rpc_readonly.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/shopspring/decimal"
)

/*
═══════════════════════════════════════════════════════════════════════
W-12：ethclient 只读（要能降级）
═══════════════════════════════════════════════════════════════════════

【核心概念】
    前 11 关都是本地的：金额精度、密钥、HD 钱包、ABI、签名。
    这一关第一次真的**连出去**，通过 JSON-RPC 读以太坊。

    ethclient 是 go-ethereum 对 JSON-RPC 的封装，读接口长这样：

        cli, _ := ethclient.Dial("https://...")     // 底层是 rpc.DialContext
        id, _   := cli.ChainID(ctx)                 // → 1（主网）
        n, _    := cli.BlockNumber(ctx)             // → 当前区块高度
        wei, _  := cli.BalanceAt(ctx, addr, nil)    // → *big.Int，单位 wei（nil = latest）
        nonce,_ := cli.PendingNonceAt(ctx, addr)    // → 下一笔交易的 nonce
        gas, _  := cli.SuggestGasPrice(ctx)         // → *big.Int，wei/gas
        tip, _  := cli.SuggestGasTipCap(ctx)        // → *big.Int，EIP-1559 的优先费
        est, _  := cli.EstimateGas(ctx, msg)        // → 这笔交易的 gas 估算

    这一关有两个「生产级」的要点，比 API 本身重要得多：

    ① **每一次远程调用都必须有超时。**
       没有超时的 RPC 调用遇到网络黑洞会永久挂住 —— 进程还在、日志不报错、
       指标不报警，但机器人已经死了。这比 panic 危险得多。
       所以：rpc.DialContext + context.WithTimeout（本关用 6 秒），
       连每次调用都要单独套一层超时。

    ② **连不上必须能降级。**
       学习材料要能在飞机上、在地铁里跑通；生产服务也不能因为「行情节点连不上」
       就整个崩掉。正确姿势：明确报错 + 打印怎么配置 + 跳过依赖它的功能。
       ⚠ 但生产代码里降级不等于假装没事 —— 该告警还是要告警，该重试还是要重试。

    另外注意：BalanceAt 返回的是 **wei 的 *big.Int**（1 ETH = 1e18 wei）。
    要展示成 ETH，必须用 decimal 做 18 位小数换算，**绝不能** float64(wei)/1e18 ——
    5.77e18 这个量级已经远超 float64 能精确表示的整数范围（2^53 ≈ 9e15），
    末几位会直接漂掉。这就是 W-01~W-03 学的东西第一次派上真实用场。

【你会看到】
    ✓ 依次尝试 5 个公共节点，连上第一个能回答 ChainID 的（实测第一个就通了）
    ✓ ChainID=1、区块高度、vitalik.eth 的余额（wei + 精确 ETH）
    ✓ 同一个余额用 decimal 算 vs 用 float64 算，差了多少 wei（固定示例里看得最清楚）
    ✓ PendingNonceAt（下一笔交易的 nonce）、SuggestGasPrice / SuggestGasTipCap（gwei）
    ✓ EstimateGas：纯转账 21000 gas，以及这笔交易大概要花多少 ETH
    ✓ 单项失败只跳过那一项（去查一个不存在的「未来区块」，被节点拒绝后程序继续）
    ✓ 节点不可达时：ETH_RPC_URL=http://127.0.0.1:1 实测 0.3 秒内降级退出、退出码 0
    ✓ 每一次调用都有自己的 6 秒超时（单点失败只跳过那一项）

    下面是本次实测（2026-10-12，走 https://eth.drpc.org；链上数据每次都会变，
    所以这些数字只是「当时真实的样子」，不是写死的期望值）：
        ChainID        1（Ethereum Mainnet 主网）
        BlockNumber    26,173,244
        余额           5768698980403065014 wei = 5.768698980403065014 ETH
        同一余额走 float64 → 5768698980403064832 wei，差 182 wei
        固定示例 1234567890123456789 wei → decimal 精确，float64 少 21 wei
        PendingNonce   5969
        SuggestGasPrice 66219438 wei = 0.066219438 gwei
        SuggestGasTipCap 7016 wei = 0.000007016 gwei
        EstimateGas    21000；成本 ≈ 1390608198000 wei ≈ 0.000001390608198 ETH

【思考题】（动手改代码验证，别空想）
    1. 把 perCallTimeout 改成 1 毫秒，再跑。你会看到哪些调用失败？
       如果这是实盘下单路径，1ms 超时意味着什么？（提示：宁可失败也不要挂住，但也不能误杀）
    2. 用 ETH_RPC_URL=http://127.0.0.1:1 跑一次，观察降级路径。
       生产代码里，降级之后你希望发生什么？（提示：告警 + 重试 + 切换备用节点）
    3. 把 BalanceAt 的第三个参数从 nil 改成 big.NewInt(某个历史高度)，会怎样？
       （提示：历史状态查询要归档节点，普通节点会报 "missing trie node"）
    4. 为什么 BalanceAt 返回 *big.Int 而不是 float64 或 int64？
       如果某个代币总量是 1e30（远超 int64 的 9.2e18），int64 会怎样？
    5. SuggestGasTipCap 在有些公共节点上返回 0。这正常吗？如果直接拿它去签名发交易，
       会发生什么？（提示：0 优先费 = 交易排在队尾，可能几小时不上链）
    6. 这一关全是「只读」。要发交易还缺哪几样东西？
       （提示：私钥/签名（W-08）、余额、nonce 管理、gas 上限、以及最重要的——别把 key 写进代码）

【通关标准】
    能默写「rpc.DialContext + context.WithTimeout + 每次调用单独超时」的只读客户端骨架，
    能把 wei 用 decimal 精确转成 ETH 展示，并且写出的代码在连不上节点时
    会明确降级而不是卡死或 panic。

【前置】W-01（float64 陷阱）· W-02/W-03（decimal）· W-08（签名交易）
【标签】ethclient · JSON-RPC · 超时 · 降级 · wei/big.Int · EIP-1559
═══════════════════════════════════════════════════════════════════════
*/

const (
	// perCallTimeout：单次 RPC 调用的超时。5~8 秒是公共节点的常见合理值 ——
	// 太短会误杀（公共节点偶尔要 2~3 秒），太长等于没超时。
	perCallTimeout = 6 * time.Second

	// overallBudget：整个程序的兜底预算。哪怕所有节点都在黑洞里，
	// 也保证这个程序会在 40 秒内退出，绝不挂死。
	overallBudget = 40 * time.Second
)

// defaultEndpoints 是内置的公共以太坊主网节点（免费、可能限频、随时可能失效）。
// 用列表而不是单个地址，是因为公共节点挂掉是常态 —— 挨个试，连上哪个用哪个。
// 生产环境请换成自己的付费节点（Alchemy / Infura / QuickNode / 自建 geth）。
var defaultEndpoints = []string{
	"https://eth.drpc.org",
	"https://ethereum-rpc.publicnode.com",
	"https://rpc.flashbots.net",
	"https://rpc.mevblocker.io",
	"https://eth-mainnet.public.blastapi.io",
}

// 两个演示用的地址：
//
//	probeAddr = vitalik.eth（余额充足，随便查，不会有人骂你）
//	toAddr    = 币安的一个热钱包（仅用于 EstimateGas 的收款方）
var (
	probeAddr = common.HexToAddress("0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045")
	toAddr    = common.HexToAddress("0x28C6c06298d514Db089934071355E5743bf21d60")
)

// chainNames 把 chainId 翻成人话。EIP-155 之后 chainId 是签名的一部分，
// 签错链的交易会被全网拒绝 —— 所以发交易前一定要确认这个数。
var chainNames = map[uint64]string{
	1:        "Ethereum Mainnet 主网",
	5:        "Goerli（已废弃的测试网）",
	11155111: "Sepolia 测试网",
	17000:    "Holesky 测试网",
	137:      "Polygon",
	56:       "BNB Smart Chain",
	42161:    "Arbitrum One",
	10:       "OP Mainnet",
	8453:     "Base",
	43114:    "Avalanche C-Chain",
}

// resolveEndpoints 决定用哪些节点：ETH_RPC_URL 优先，否则用内置列表。
// 返回 (节点列表, 来源说明)。
func resolveEndpoints() ([]string, string) {
	if u := strings.TrimSpace(os.Getenv("ETH_RPC_URL")); u != "" {
		return []string{u}, "环境变量 ETH_RPC_URL（覆盖了内置列表）"
	}
	return defaultEndpoints, "内置公共节点列表（可用 ETH_RPC_URL 覆盖）"
}

// shortErr 把可能很长的 RPC 错误压成一行，方便打印。
func shortErr(err error) string {
	s := strings.ReplaceAll(err.Error(), "\n", " ")
	if len(s) > 120 {
		s = s[:120] + "…"
	}
	return s
}

// callCtx 给单次调用套上 perCallTimeout。
// ⚠ 关键点：超时不是「给整个程序设一个」就够了 —— 每次远程调用都要有自己的，
// 否则第一次慢调用会把整个预算耗光，后面所有调用都直接失败。
func callCtx(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, perCallTimeout)
}

// withCall 执行一次带超时的只读调用，把「超时 + 错误降级」收在一处。
// T 是返回值类型 —— 泛型在这里刚好省掉一大坨 ctx/cancel/err 样板代码。
// 返回 ok=false 表示这一项失败了，调用方应该**跳过它继续往下走**（降级）。
func withCall[T any](ctx context.Context, name string, fn func(context.Context) (T, error)) (T, bool) {
	var zero T
	cctx, cancel := context.WithTimeout(ctx, perCallTimeout)
	defer cancel()
	v, err := fn(cctx)
	if err != nil {
		fmt.Printf("  ✗ %s 失败（降级跳过这一项）：%s\n", name, shortErr(err))
		return zero, false
	}
	return v, true
}

// dialFirst 依次尝试候选节点，返回第一个「能连上、并且能回答 ChainID」的客户端。
//
// 为什么要多探一次 ChainID？因为 HTTP 连上了不代表节点可用 ——
// 有的节点会返回 401/403（需要 API key），有的直接返回错误。
// 探一次 ChainID 是最便宜的健康检查。
func dialFirst(ctx context.Context, endpoints []string) (*ethclient.Client, string, error) {
	var lastErr error
	for _, ep := range endpoints {
		// rpc.DialContext 的 ctx 只作用于「建立连接」这一步，
		// 之后的调用不受它影响（go-ethereum 文档明确说明）—— 所以这里可以放心 cancel。
		dialCtx, cancelDial := callCtx(ctx)
		rpcCli, err := rpc.DialContext(dialCtx, ep)
		if err != nil {
			cancelDial()
			lastErr = err
			fmt.Printf("  ✗ %-42s 连接失败：%s\n", ep, shortErr(err))
			continue
		}
		cli := ethclient.NewClient(rpcCli)

		// ⚠ 注意：HTTP 传输的 rpc.DialContext 是「懒连接」—— 它根本不会真的去连，
		// 所以 dial 返回 nil 错误**不代表**节点可用。必须真发一个请求才能确认。
		// 这就是为什么健康检查不能只探端口，得探一个真实接口（这里用最便宜的 ChainID）。
		probeCtx, cancelProbe := callCtx(ctx)
		id, err := cli.ChainID(probeCtx)
		cancelProbe()
		cancelDial()
		if err != nil {
			cli.Close()
			lastErr = err
			fmt.Printf("  ✗ %-42s 不可用：%s\n", ep, shortErr(err))
			continue
		}
		fmt.Printf("  ✓ %-42s 连通（chainId=%d）\n", ep, id)
		return cli, ep, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("候选节点列表为空")
	}
	return nil, "", lastErr
}

// comma 给整数加千分位，纯为了好读：26173190 → 26,173,190。
func comma(n uint64) string {
	s := strconv.FormatUint(n, 10)
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	return b.String()
}

// weiToETH 用 decimal 把 wei（*big.Int）精确换算成 ETH 字符串（18 位小数）。
// ⚠ 绝不要写成 float64(wei) / 1e18 —— 见示例 3 的对比。
func weiToETH(wei *big.Int) string {
	return decimal.NewFromBigInt(wei, 0).Shift(-18).String()
}

// weiToGwei 把 wei 换算成 gwei（1 gwei = 1e9 wei），这是 gas 价格的常用单位。
func weiToGwei(wei *big.Int) string {
	return decimal.NewFromBigInt(wei, 0).Shift(-9).String()
}

// printDegradeHint 在「一个节点都连不上」时打印配置指南。
//
// 为什么降级而不是直接报错退出？因为这是学习材料，要能在没有网络的环境里跑通。
// ⚠ 生产代码里绝不能这样「假装没事」：连不上节点就该告警 + 重试 + 切备用节点，
//
//	因为读不到链上状态时继续下单，比停下来危险得多。
func printDegradeHint(endpoints []string, err error) {
	fmt.Println("📌 降级：连不上任何以太坊节点 → 跳过链上读取（不影响本模块其它关卡）")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("  最后一次错误：%s\n", shortErr(err))
	fmt.Println("  尝试过的地址：")
	for _, ep := range endpoints {
		fmt.Printf("    · %s\n", ep)
	}
	fmt.Println("  怎么配置（任选一种）：")
	fmt.Println("    1) 临时指定一个节点：")
	fmt.Println("         ETH_RPC_URL=https://你的节点 go run 12_rpc_readonly.go")
	fmt.Println("    2) 常用公共节点（免费、可能限频、随时失效）：")
	fmt.Println("         https://ethereum-rpc.publicnode.com")
	fmt.Println("         https://eth.drpc.org")
	fmt.Println("         https://rpc.flashbots.net")
	fmt.Println("    3) 生产环境用带 key 的付费节点（Alchemy / Infura / QuickNode / Ankr）：")
	fmt.Println("         ETH_RPC_URL=https://eth-mainnet.g.alchemy.com/v2/<你的key>")
	fmt.Println("    4) 自己跑一个本地节点（最可控，也最贵）：")
	fmt.Println("         geth --http --http.api eth,net,web3   →   ETH_RPC_URL=http://127.0.0.1:8545")
	fmt.Println("  提示：本关全是只读调用，任何节点都能做；要发交易还需要私钥签名和余额。")
	fmt.Println("  退出码 0：降级不算失败（生产代码里则应该告警）。")
}

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 W-12 ethclient 只读")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// 兜底预算：无论发生什么，40 秒后 ctx 一定会取消，程序一定会退出。
	ctx, cancel := context.WithTimeout(context.Background(), overallBudget)
	defer cancel()

	// ── 示例 1：连接（多节点 + 超时 + 降级）────────────────────────────
	fmt.Println("📝 示例 1：连接：rpc.DialContext + context.WithTimeout（单次 6 秒）")
	fmt.Println(strings.Repeat("-", 60))
	endpoints, source := resolveEndpoints()
	fmt.Printf("  节点来源：%s\n", source)
	if len(endpoints) > 1 {
		fmt.Println("  ➡️  用列表挨个试：公共节点挂掉是常态，连上哪个用哪个")
	}

	cli, used, err := dialFirst(ctx, endpoints)
	if err != nil {
		printDegradeHint(endpoints, err)
		return // 正常返回 → 退出码 0
	}
	defer cli.Close()
	fmt.Printf("  本次使用：%s\n", used)
	fmt.Println("  ⚠️ 每一次远程调用都单独套 6 秒超时：没有超时的 RPC 会在网络黑洞里永久挂住，")
	fmt.Println("     进程还活着、日志不报错、指标不报警 —— 那比 panic 危险得多。")

	// ── 示例 2：ChainID 与区块高度 ─────────────────────────────────────
	fmt.Println("📌 示例 2：ChainID / BlockNumber：最便宜的两个只读调用")
	fmt.Println(strings.Repeat("-", 60))
	if id, ok := withCall(ctx, "ChainID", cli.ChainID); ok {
		name := chainNames[id.Uint64()]
		if name == "" {
			name = "未知链（不是主网）"
		}
		fmt.Printf("  ChainID      : %s（%s）\n", id, name)
		fmt.Println("  ➡️  交易签名必须带 chainId（EIP-155），签错链的 tx 会被全网拒绝 —— 见 W-08")
	}
	if n, ok := withCall(ctx, "BlockNumber", cli.BlockNumber); ok {
		fmt.Printf("  BlockNumber  : %s（十六进制 0x%X）\n", comma(n), n)
		fmt.Println("  ➡️  区块高度是「链上时间」的刻度：查历史状态、算确认数、判断自己是否掉队都靠它")
	}

	// ── 示例 3：余额（wei → ETH，decimal vs float64）───────────────────
	fmt.Println("🔍 示例 3：BalanceAt：返回的是 wei 的 *big.Int，要自己换算成 ETH")
	fmt.Println(strings.Repeat("-", 60))
	if wei, ok := withCall(ctx, "BalanceAt", func(c context.Context) (*big.Int, error) {
		return cli.BalanceAt(c, probeAddr, nil) // nil = latest；也可以传 big.NewInt(高度) 查历史
	}); ok {
		fmt.Printf("  地址         : %s（vitalik.eth）\n", probeAddr.Hex())
		fmt.Printf("  余额（wei）  : %s\n", wei.String())
		fmt.Printf("  余额（ETH）  : %s   ← decimal 精确换算（wei ÷ 1e18）\n", weiToETH(wei))

		// 对比：如果偷懒用 float64 换算，会差多少？
		f64, _ := new(big.Float).SetInt(wei).Float64()
		back, _ := new(big.Float).SetFloat64(f64).Int(nil)
		diff := new(big.Int).Sub(wei, back)
		fmt.Printf("  用 float64 算: 会得到 %s wei，和真实值差 %s wei\n", back.String(), diff.String())
		fmt.Println("  ➡️  这个差值就是「精度丢失」的真身：float64 只有 53 位尾数，")
		fmt.Println("    能精确表示的整数上限是 2^53 ≈ 9.007e15，而余额是 5.77e18 量级。")

		// 固定示例：不随链上余额变化，看得更清楚
		sample, _ := new(big.Int).SetString("1234567890123456789", 10) // 1.234567890123456789 ETH
		fmt.Println("\n  固定示例（不随链上余额变化，看得更清楚）：")
		fmt.Printf("    wei 原值       : %s\n", sample.String())
		fmt.Printf("    decimal 精确算 : %s ETH\n", weiToETH(sample))
		sf, _ := new(big.Float).SetInt(sample).Float64()
		sBack, _ := new(big.Float).SetFloat64(sf).Int(nil)
		fmt.Printf("    float64 还原   : %s wei（少了 %s wei）\n",
			sBack.String(), new(big.Int).Sub(sample, sBack).String())
		fmt.Println("    ↑ 只要走一趟 float64，最后 3 位就没了。金额计算全程 decimal / 整数最小单位。")
	}

	// ── 示例 4：nonce 与 gas 价格 ──────────────────────────────────────
	fmt.Println("🧪 示例 4：PendingNonceAt / SuggestGasPrice / SuggestGasTipCap")
	fmt.Println(strings.Repeat("-", 60))
	if nonce, ok := withCall(ctx, "PendingNonceAt", func(c context.Context) (uint64, error) {
		return cli.PendingNonceAt(c, probeAddr)
	}); ok {
		fmt.Printf("  PendingNonce : %d   ← 这个地址「下一笔」交易该用的 nonce\n", nonce)
		fmt.Println("  ➡️  nonce 是账户的交易序号，从 0 开始、必须连续。重发同一 nonce 会覆盖旧交易；")
		fmt.Println("    用 pending 而不是 latest，是为了把「已发出但还没上链」的交易也算进去，")
		fmt.Println("    否则连续下单会全部撞在同一个 nonce 上（经典事故）")
	}
	if gp, ok := withCall(ctx, "SuggestGasPrice", cli.SuggestGasPrice); ok {
		fmt.Printf("  SuggestGasPrice: %s wei = %s gwei（传统 gasPrice 模式）\n", gp.String(), weiToGwei(gp))
	}
	if tip, ok := withCall(ctx, "SuggestGasTipCap", cli.SuggestGasTipCap); ok {
		fmt.Printf("  SuggestGasTipCap: %s wei = %s gwei（EIP-1559 的小费）\n", tip.String(), weiToGwei(tip))
		fmt.Println("  ➡️  EIP-1559 之后费用拆成两部分：baseFee（协议烧掉）+ tip（给矿工/验证者的小费）。")
		fmt.Println("    公共节点给出的 tip 常常是 0 或很小 —— 直接拿去签名会排到队尾，可能几小时不上链。")
		fmt.Println("    生产里要自己加保底：tip = max(建议值, 你的下限)")
	}

	// ── 示例 5：EstimateGas + 成本估算 ─────────────────────────────────
	fmt.Println("⚙️ 示例 5：EstimateGas：这笔交易要花多少 gas，以及大概多少 ETH")
	fmt.Println(strings.Repeat("-", 60))
	oneETH := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil) // 1e18 wei = 1 ETH
	gas, ok := withCall(ctx, "EstimateGas", func(c context.Context) (uint64, error) {
		// CallMsg 描述「如果发这笔交易会怎样」——注意它**不需要私钥**，
		// 因为节点只是在本地模拟执行，不会广播出去。
		return cli.EstimateGas(c, ethereum.CallMsg{
			From:  probeAddr,
			To:    &toAddr,
			Value: oneETH,
		})
	})
	if ok {
		fmt.Printf("  交易内容     : 从 vitalik.eth 转 1 ETH 到 %s\n", toAddr.Hex())
		fmt.Printf("  gas 估算     : %d\n", gas)
		fmt.Println("  ➡️  21000 是「纯转账」的固定成本（以太坊协议写死的）。")
		fmt.Println("    对比一下：ERC-20 transfer ≈ 51000，Uniswap 一笔 swap ≈ 150000~300000，")
		fmt.Println("    所以「gas 费」不是转账费，而是你占用了多少计算量。")
		fmt.Println("    ⚠ EstimateGas 只能估算，实际执行可能更多：上链前记得留 10~20% 余量")
	}
	if gp, ok := withCall(ctx, "SuggestGasPrice(成本估算用)", cli.SuggestGasPrice); ok {
		cost := new(big.Int).Mul(new(big.Int).SetUint64(gas), gp) // gas × 单价 = 总费用（wei）
		fmt.Printf("\n  成本估算     : %d gas × %s wei/gas = %s wei ≈ %s ETH\n",
			gas, gp.String(), cost.String(), weiToETH(cost))
		fmt.Println("  ➡️  这就是「一笔转账要花多少钱」的算法：gas 用量 × gas 单价。")
		fmt.Println("    注意 baseFee 每个区块都会变（拥堵时涨、空闲时跌），所以成本只在当下有效。")
	}

	// ── 示例 6：单项失败只跳过这一项 ───────────────────────────────────
	fmt.Println("🎯 示例 6：单项失败只跳过这一项：去查一个还不存在的「未来区块」")
	fmt.Println(strings.Repeat("-", 60))
	future := new(big.Int).SetUint64(99999999) // 远高于当前高度，节点一定会拒绝
	if _, ok := withCall(ctx, "BalanceAt(未来区块)", func(c context.Context) (*big.Int, error) {
		return cli.BalanceAt(c, probeAddr, future)
	}); !ok {
		fmt.Println("  ➡️  这一项失败了，但前面的输出全部照常 —— 这就是「降级不中断」：")
		fmt.Println("    单个接口挂掉只影响依赖它的功能，不能让整个机器人停摆。")
		fmt.Println("    真实场景里对应的错误还有：历史状态查归档节点失败（missing trie node）、")
		fmt.Println("    合约调用 revert、限频 429 —— 全都该走这条路：记日志、跳过、继续。")
	}

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ 只读三件套：rpc.DialContext（建连）+ 每次调用 context.WithTimeout（超时）")
	fmt.Println("    + 出错就跳过并打印原因（降级）—— 三者缺一，你的机器人就会在某个夜里静默死掉")
	fmt.Println("  ✅ ChainID 必须核对：EIP-155 之后它是签名的一部分，签错链的交易会被全网拒绝")
	fmt.Println("  ✅ BalanceAt 返回 wei 的 *big.Int，换算 ETH 必须用 decimal（float64 会丢末几位）")
	fmt.Println("  ✅ PendingNonceAt 而不是 NonceAt(latest)：连续下单必须把在途交易算进 nonce")
	fmt.Println("  ✅ EIP-1559：总费用 = gas 用量 × (baseFee + tip)；公共节点给的 tip 常为 0，要加保底")
	fmt.Println("  ✅ 21000 = 纯转账；ERC-20 transfer ≈ 51000；swap ≈ 15~30 万。EstimateGas 只是估算")
	fmt.Println("  ✅ 该用：查余额、查 nonce、估 gas、读合约状态（只读调用不需要私钥，最安全）")
	fmt.Println("  ✅ 不该用：把 RPC 调用放在没有超时、没有重试、没有备用节点的裸调用里")
}
