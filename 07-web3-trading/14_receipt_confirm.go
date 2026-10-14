//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
//
// 跑法：cd 07-web3-trading && go run 14_receipt_confirm.go
// 本关纯本地：假节点在进程里挖块、给回执。
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"context"
	"fmt"
	"strings"
	"time"
)

/*
═══════════════════════════════════════════════════════════════════════
W-14：等回执与确认数 —— 广播成功 ≠ 成交
═══════════════════════════════════════════════════════════════════════

【核心概念】
    eth_sendRawTransaction 只保证节点收下了字节，返回交易哈希。
    钱有没有转出去，要看回执：

      receipt == nil          还在 mempool / 节点还没看见
      receipt.Status == 0     进块了但 EVM 执行失败（gas 费照扣）
      receipt.Status == 1     执行成功
      确认数 = 当前高度 - 回执所在块 + 1

    机器人不能「发出去就当成交」。正确循环：

      1. 带超时的 context 里轮询 eth_getTransactionReceipt
      2. status=0 → 当失败，不要重发同一 nonce（已经进块）
      3. 等 N 个确认再改自己的订单状态（重组会把浅块打掉）
      4. 超时 → 走 W-13 的同 nonce 替换，而不是再签一笔新 nonce

【你会看到】
    ✓ 发出去后前几次 receipt 是 nil
    ✓ 成功回执 status=1，确认数随出块增加
    ✓ 失败回执 status=0：进块了，gas 没了，业务没做成
    ✓ context 超时：停止空转，转去替换 nonce

【思考题】
    1. 为什么 status=0 时不能再拿一个新 nonce 重发？
    2. 主网常见等 12 个确认，L2 为什么可以少很多？
    3. 轮询间隔怎么定？太密会打爆节点，太疏会让用户以为卡住。

【通关标准】
    能写出「poll receipt + 确认数 + 超时后替换 nonce」这条链路，
    并说清 status=0 和 receipt=nil 不是一回事。

【前置】W-13 · W-12        【标签】receipt · status · 确认数 · 重组
═══════════════════════════════════════════════════════════════════════
*/

type receipt struct {
	status      uint64
	blockNumber uint64
	hash        string
}

type node struct {
	head     uint64
	pool     map[string]uint64 // hash → 还要几块才打包，0 表示下一轮进块
	receipts map[string]receipt
	fail     map[string]bool // 进块后标失败（模拟 revert）
}

func newNode(head uint64) *node {
	return &node{
		head:     head,
		pool:     map[string]uint64{},
		receipts: map[string]receipt{},
		fail:     map[string]bool{},
	}
}

func (n *node) Send(hash string, blocksUntilMine uint64, revert bool) {
	n.pool[hash] = blocksUntilMine
	if revert {
		n.fail[hash] = true
	}
}

func (n *node) Mine() {
	n.head++
	for hash, wait := range n.pool {
		if wait > 0 {
			n.pool[hash] = wait - 1
			continue
		}
		st := uint64(1)
		if n.fail[hash] {
			st = 0
		}
		n.receipts[hash] = receipt{status: st, blockNumber: n.head, hash: hash}
		delete(n.pool, hash)
	}
}

func (n *node) Receipt(hash string) *receipt {
	r, ok := n.receipts[hash]
	if !ok {
		return nil
	}
	cp := r
	return &cp
}

func confirmations(n *node, r *receipt) uint64 {
	if r == nil || n.head < r.blockNumber {
		return 0
	}
	return n.head - r.blockNumber + 1
}

func waitReceipt(ctx context.Context, n *node, hash string, wantConf uint64, tick time.Duration) (receipt, error) {
	t := time.NewTicker(tick)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return receipt{}, fmt.Errorf("等回执超时：%w", ctx.Err())
		case <-t.C:
			n.Mine()
			r := n.Receipt(hash)
			if r == nil {
				fmt.Printf("     … %s 还在 mempool（高度 %d）\n", hash, n.head)
				continue
			}
			if r.status == 0 {
				return *r, fmt.Errorf("回执失败 status=0（EVM revert，gas 已扣）块=%d", r.blockNumber)
			}
			conf := confirmations(n, r)
			fmt.Printf("     … 已进块 %d，确认 %d/%d\n", r.blockNumber, conf, wantConf)
			if conf >= wantConf {
				return *r, nil
			}
		}
	}
}

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 W-14 等回执与确认数")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	fmt.Println("📝 示例 1：成功路径 —— nil → status=1 → 确认数够")
	fmt.Println(strings.Repeat("-", 60))
	n := newNode(1000)
	n.Send("0xaaa", 2, false) // 再挖 2 块才打包
	ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
	r, err := waitReceipt(ctx, n, "0xaaa", 3, 40*time.Millisecond)
	cancel()
	if err != nil {
		fmt.Println("  ❌", err)
	} else {
		fmt.Printf("  ✅ 成交 hash=%s 块=%d 确认=%d\n", r.hash, r.blockNumber, confirmations(n, &r))
	}

	fmt.Println("\n📌 示例 2：status=0 —— 进块了，但合约 revert")
	fmt.Println(strings.Repeat("-", 60))
	n2 := newNode(2000)
	n2.Send("0xbbb", 1, true)
	ctx2, cancel2 := context.WithTimeout(context.Background(), 400*time.Millisecond)
	_, err = waitReceipt(ctx2, n2, "0xbbb", 1, 40*time.Millisecond)
	cancel2()
	fmt.Printf("  %v\n", err)
	fmt.Println("  ➡️  这一笔的 nonce 已经消耗。重试必须用新 nonce，不能当「没发出去」。")

	fmt.Println("\n🔍 示例 3：超时 —— 别空转，去替换 nonce（W-13）")
	fmt.Println(strings.Repeat("-", 60))
	n3 := newNode(3000)
	n3.Send("0xccc", 99, false) // 几乎不会进块
	ctx3, cancel3 := context.WithTimeout(context.Background(), 120*time.Millisecond)
	_, err = waitReceipt(ctx3, n3, "0xccc", 1, 40*time.Millisecond)
	cancel3()
	fmt.Printf("  %v\n", err)
	fmt.Println("  ➡️  超时后用同一 nonce + 更高 gas 重签（加速或取消），不要再开一个新 nonce。")

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ sendRawTransaction 只返回 hash，成交看 receipt")
	fmt.Println("  ✅ receipt=nil：还在路上；status=0：进块但失败；status=1：执行成功")
	fmt.Println("  ✅ 确认数 = 当前高度 - 回执块高 + 1；重组会回滚浅确认")
	fmt.Println("  ✅ 轮询必须带 context 超时；超时走同 nonce 替换，不是新 nonce")
	fmt.Println("  ✅ status=0 已经消耗 nonce，重试要重新取 pending nonce")
	fmt.Println(strings.Repeat("=", 60) + "\n")
}
