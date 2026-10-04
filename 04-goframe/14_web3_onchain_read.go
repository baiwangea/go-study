package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"go-study/goframe/level"
)

// ethCall 手搓一次 JSON-RPC 并取回 result 字段。
// 用 geth 的 ethclient 更省事，但先手写一遍才知道它替你做了什么（超时、错误、hex 解析）。
func ethCall(ctx context.Context, endpoint, method string, params ...any) (string, error) {
	payload, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0", "id": 1, "method": method, "params": params,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("构造请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("RPC 网络层失败: %w", err) // 连不上/超时
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP 状态异常 %d", resp.StatusCode)
	}

	var out struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", fmt.Errorf("响应不是合法 JSON: %w", err)
	}
	if out.Error != nil { // JSON-RPC 层错误：HTTP 200 也可能带着 error
		return "", fmt.Errorf("RPC 错误 code=%d msg=%s", out.Error.Code, out.Error.Message)
	}
	var s string
	if err := json.Unmarshal(out.Result, &s); err != nil {
		return "", fmt.Errorf("result 不是字符串: %w", err)
	}
	return s, nil
}

// hexToBig 把 "0x..." 解析成大整数；uint256 远超 int64，只能用 big.Int。
func hexToBig(h string) (*big.Int, bool) {
	return new(big.Int).SetString(strings.TrimPrefix(h, "0x"), 16)
}

// L3-14：链上只读（区块高度、余额、calldata 拼装）。
func L14() level.Level {
	return level.Level{
		ID:      "L3-14",
		Title:   "链上只读：JSON-RPC 与 calldata",
		Tags:    "eth_blockNumber · eth_getBalance · selector · big.Int",
		Pre:     "L3-02, L3-03",
		Goal:    "直接向节点发 JSON-RPC：取区块高度、查余额、拼出 ERC20 调用数据，看清 ethclient 帮你藏了什么",
		Observe: "打印当前区块号（十进制）、以太坊基金会地址余额（ETH）、transfer 的 4 字节 selector 与完整 calldata",
		Questions: []string{
			"为什么必须用 big.Int 解析 hex，而不能用 strconv.ParseInt？（uint256 上限 2^256-1）",
			"HTTP 200 但 body 里有 error 字段 —— 这一层判断漏掉会发生什么？（对照 L3-03 两层成功）",
			"eth_call 与 eth_getBalance 的差别是什么？为什么说读操作不消耗 gas？",
			"selector 只有 4 字节，撞了怎么办？（用 geth 的 crypto.Keccak256 亲手算一遍 transfer 的选择器验证）",
			"公共节点会限频、也可能给你旧区块：机器人该用多节点 + 区块高度回退怎么设计？",
		},
		Check: "能手写 JSON-RPC 调用、正确解析 hex 金额，并解释一次 ERC20 调用的数据布局",
		Run: func() {
			endpoint := os.Getenv("ETH_RPC_URL")
			if endpoint == "" {
				endpoint = "https://ethereum-rpc.publicnode.com"
				fmt.Println("  未设置 ETH_RPC_URL，用公共节点：", endpoint)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			defer cancel()

			blk, err := ethCall(ctx, endpoint, "eth_blockNumber")
			if err != nil {
				fmt.Println("  节点不可达，本关跳过：", err)
				return
			}
			n, _ := hexToBig(blk)
			fmt.Printf("  eth_blockNumber → %s = 区块 %s\n", blk, n)

			bal, err := ethCall(ctx, endpoint, "eth_getBalance",
				"0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe", "latest")
			if err != nil {
				fmt.Println("  eth_getBalance 失败：", err)
			} else {
				wei, _ := hexToBig(bal)
				eth := new(big.Float).Quo(new(big.Float).SetInt(wei), new(big.Float).SetFloat64(1e18))
				fmt.Printf("  eth_getBalance  → %s wei ≈ %s ETH（以太坊基金会地址）\n", wei, eth.Text('f', 4))
			}

			// ERC20 transfer(address,uint256)：selector(4B) + 地址左补齐 32B + 金额 32B
			const selector = "a9059cbb" // keccak256("transfer(address,uint256)") 前 4 字节
			addr := strings.TrimPrefix("0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045", "0x")
			amt := new(big.Int).Mul(big.NewInt(2), big.NewInt(1e18)) // 2 枚代币（18 位小数）
			data := "0x" + selector +
				pad64(addr) +
				pad64(hex.EncodeToString(amt.Bytes()))
			fmt.Printf("  calldata        → %s\n", data)
			fmt.Printf("                    selector=%s 参数1(地址)=%s 参数2(wei=%s)\n",
				data[2:10], data[10:74], amt)
			fmt.Println("  ↑ 一次链上调用本质就是这段 calldata；ethclient/bind 只是替你算了 selector 和补齐")
		},
	}
}

// pad64 把 hex 片段左补 0 到 32 字节（64 个 hex 字符）。
func pad64(s string) string {
	const width = 64
	if len(s) >= width {
		return s[len(s)-width:]
	}
	return strings.Repeat("0", width-len(s)) + s
}
