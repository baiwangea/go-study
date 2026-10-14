//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
//
// 跑法：cd 07-web3-trading && go run 13_nonce_send.go
// 本关纯本地，不连链。假节点在进程里。
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"strings"
)

/*
═══════════════════════════════════════════════════════════════════════
W-13：Nonce 管理 —— 签完名还不能闭眼广播
═══════════════════════════════════════════════════════════════════════

【核心概念】
    W-08 把交易签好了。链上同一账户发出的交易必须 **nonce 连续**：
      · 账户历史上成功打包的交易数叫 confirmed nonce
      · 已经广播、还在 mempool 里的叫 pending
      · 下一笔该用的是 pending（不是 latest）

    机器人最容易踩的三个坑：

      1. 跳号：本地发了 nonce=5，却没发 nonce=4 → 5 永远卡在 mempool
      2. 重放：崩溃后从 latest 重新数，把已经 pending 的 nonce 再用一次
         → 节点返回 nonce too low，或者替换掉还没打包的那笔
      3. 替换：同一 nonce 再发一笔更高 gas 的，才会顶掉卡住的那笔
         （这是「加速」和「取消」的唯一正规手段）

    正确姿势：进程里维护一个 pending 计数器，启动时用
    eth_getTransactionCount(addr, "pending") 对齐链上，之后只在本地 +1。
    打包确认后再把 confirmed 往前推。

【你会看到】
    ✓ latest=3、pending=5：链上已确认 3 笔，mempool 还有 2 笔没打包
    ✓ 跳过 nonce=3 去发 4 → 4 卡住，confirmed 不动
    ✓ 同一 nonce 用更高 fee 替换 → 旧交易被顶掉
    ✓ 崩溃后若用 latest 对齐会把 pending 的 nonce 再用一遍

【思考题】
    1. 为什么对齐要用 pending 而不是 latest？
    2. 两个进程同时给同一个热钱包发交易会怎样？这就是为什么热钱包要单飞。
    3. 「取消一笔卡住的交易」为什么是「同样 nonce、To=自己、Value=0、更高 gas」？

【通关标准】
    能画出 confirmed / pending 两条水位，能解释跳号为什么卡死，
    知道加速和取消都是「同 nonce 替换」。

【前置】W-08 · W-12        【标签】nonce · mempool · 替换交易 · 热钱包单飞
═══════════════════════════════════════════════════════════════════════
*/

// chain 是进程内的假节点：只关心一个账户的 nonce 水位。
type chain struct {
	confirmed uint64 // 已经进块的交易数 = next latest
	mempool   map[uint64]tx
}

type tx struct {
	nonce uint64
	fee   int // 只用来比较谁能替换谁，单位随意
	note  string
}

func newChain(confirmed uint64) *chain {
	return &chain{confirmed: confirmed, mempool: map[uint64]tx{}}
}

func (c *chain) Latest() uint64  { return c.confirmed }
func (c *chain) Pending() uint64 { return c.confirmed + uint64(len(c.mempool)) }

func (c *chain) Broadcast(t tx) string {
	if t.nonce < c.confirmed {
		return fmt.Sprintf("❌ nonce too low（链上已确认到 %d，你却发 %d）", c.confirmed, t.nonce)
	}
	if t.nonce > c.Pending() {
		c.mempool[t.nonce] = t
		return fmt.Sprintf("⚠️  已进 mempool，但 nonce=%d 跳号：前面缺 %d，这笔会一直挂着", t.nonce, c.Pending()-1)
	}
	if old, ok := c.mempool[t.nonce]; ok {
		if t.fee <= old.fee {
			return fmt.Sprintf("❌ 同 nonce=%d 替换失败：新 fee=%d ≤ 旧 fee=%d", t.nonce, t.fee, old.fee)
		}
		c.mempool[t.nonce] = t
		return fmt.Sprintf("✅ 替换成功：nonce=%d 旧「%s」被「%s」顶掉（fee %d→%d）", t.nonce, old.note, t.note, old.fee, t.fee)
	}
	c.mempool[t.nonce] = t
	return fmt.Sprintf("✅ 进入 mempool：nonce=%d 「%s」", t.nonce, t.note)
}

func (c *chain) MineNext() string {
	want := c.confirmed
	t, ok := c.mempool[want]
	if !ok {
		return fmt.Sprintf("⏸  不出块：下一口该吃 nonce=%d，mempool 里没有", want)
	}
	delete(c.mempool, want)
	c.confirmed++
	return fmt.Sprintf("🧱 打包 nonce=%d 「%s」，confirmed 变成 %d", t.nonce, t.note, c.confirmed)
}

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 W-13 Nonce 管理")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	fmt.Println("📝 示例 1：latest 和 pending 不是同一个数")
	fmt.Println(strings.Repeat("-", 60))
	n := newChain(3)
	fmt.Println("  " + n.Broadcast(tx{nonce: 3, fee: 10, note: "转 1 ETH"}))
	fmt.Println("  " + n.Broadcast(tx{nonce: 4, fee: 10, note: "再转 0.1 ETH"}))
	fmt.Printf("  eth_getTransactionCount(latest)  = %d  ← 已进块的笔数\n", n.Latest())
	fmt.Printf("  eth_getTransactionCount(pending) = %d  ← 下一笔该用这个\n", n.Pending())
	fmt.Println("  ➡️  机器人启动时要对齐 pending。用 latest 会把 mempool 里的 nonce 再用一遍。")

	fmt.Println("\n📌 示例 2：跳号 → 后面的交易全部卡住")
	fmt.Println(strings.Repeat("-", 60))
	gap := newChain(3)
	fmt.Println("  " + gap.Broadcast(tx{nonce: 4, fee: 10, note: "跳过 3 直接发 4"}))
	fmt.Println("  " + gap.MineNext())
	fmt.Println("  ➡️  nonce=4 在池子里，但矿工下一口只吃 3。4 会一直挂到 3 出现，或者过期。")

	fmt.Println("\n🔍 示例 3：同 nonce + 更高 gas = 加速 / 取消")
	fmt.Println(strings.Repeat("-", 60))
	rep := newChain(3)
	fmt.Println("  " + rep.Broadcast(tx{nonce: 3, fee: 10, note: "gas 太低，卡住"}))
	fmt.Println("  " + rep.Broadcast(tx{nonce: 3, fee: 10, note: "同样 fee 再发一次"}))
	fmt.Println("  " + rep.Broadcast(tx{nonce: 3, fee: 25, note: "To=自己 Value=0 取消"}))
	fmt.Println("  " + rep.MineNext())
	fmt.Println("  ➡️  没有「cancel」RPC。取消 = 用同一 nonce 给自己转 0，并出更高的费。")

	fmt.Println("\n🧪 示例 4：崩溃重启，错误地用 latest 对齐")
	fmt.Println(strings.Repeat("-", 60))
	crash := newChain(3)
	fmt.Println("  " + crash.Broadcast(tx{nonce: 3, fee: 10, note: "崩溃前发出去了"}))
	fmt.Printf("  崩溃前：latest=%d pending=%d\n", crash.Latest(), crash.Pending())
	wrong := crash.Latest() // 错误：当下一笔 nonce
	fmt.Printf("  重启后若读 latest=%d 再发一笔 → %s\n", wrong,
		crash.Broadcast(tx{nonce: wrong, fee: 12, note: "重启后重复的 nonce=3"}))
	fmt.Println("  ➡️  热钱包同一时刻只能有一个签名进程。多进程 = 抢 nonce。")

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ 下一笔 nonce = eth_getTransactionCount(addr, \"pending\")")
	fmt.Println("  ✅ 本地维护 pending 计数，广播成功才 +1；进块后再推 confirmed")
	fmt.Println("  ✅ 跳号的交易会永久挂在 mempool，直到缺口被填上")
	fmt.Println("  ✅ 加速 / 取消 = 同一 nonce + 更高 fee（取消时 To=自己、Value=0）")
	fmt.Println("  ✅ 热钱包单飞：同一个私钥不要让两个进程同时签")
	fmt.Println(strings.Repeat("=", 60) + "\n")
}
