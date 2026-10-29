//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
//
// 跑法：cd 03-go-concurrency && go run synclib/05_atomic_value.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

/*
═══════════════════════════════════════════════════════════════════════
L2-23：atomic.Value —— 热路径读配置，不用锁
═══════════════════════════════════════════════════════════════════════

【核心概念】
    行情循环每秒读几千次「阈值、交易对列表」。
    用 Mutex 包住整份配置，读也要抢锁。
    atomic.Value 的用法是：写的人准备好一份完整新值再 Store，
    读的人 Load 拿到的永远是某一份完整快照，不会读到写了一半的结构体。

    两条铁律：
      1. 第一次 Store 之后，具体类型不能变（先存 Config 再存 *Config 会 panic）
      2. 不要改已经 Store 出去的那份。要改就复制一份再 Store。

    计数器用 atomic.Int64（L2-13）。
    发布一整份配置用 atomic.Value。别用 Value 存 int。

【你会看到】
    ✓ 读者一直 Load，中途 Store 新阈值，之后读到的是新快照
    ✓ 改旧副本不会影响已经发出去的快照
    ✓ 类型不一致时 Store panic（本关用 recover 接住演示）

【思考题】
    1. 为什么不能 Load 出来后改字段再当原件用？
    2. 配置很大时，每次全量复制值不值？什么时候退回 RWMutex？
    3. atomic.Pointer[Config]（Go 1.19）和 Value 差在哪？

【通关标准】
    能写出「复制 → 修改 → Store」的发布流程，
    并说清为什么热路径只 Load、不持锁。

【前置】L2-13 · L2-15     【标签】atomic.Value · 快照 · 配置热更新
═══════════════════════════════════════════════════════════════════════
*/

type botCfg struct {
	threshold string
	symbols   []string
}

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L2-23 atomic.Value 发布配置快照")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	var live atomic.Value
	live.Store(botCfg{threshold: "0.001", symbols: []string{"BTCUSDT"}})

	fmt.Println("📝 示例 1：读者只 Load，写者整份替换")
	fmt.Println(strings.Repeat("-", 60))
	got := live.Load().(botCfg)
	fmt.Printf("  热路径读到 threshold=%s symbols=%v\n", got.threshold, got.symbols)

	next := botCfg{threshold: "0.002", symbols: []string{"BTCUSDT", "ETHUSDT"}}
	live.Store(next)
	got = live.Load().(botCfg)
	fmt.Printf("  Store 之后读到 threshold=%s symbols=%v\n", got.threshold, got.symbols)
	fmt.Println("  ➡️  行情循环里不要加锁，Load 一次拿整份。")

	fmt.Println("\n📌 示例 2：改已经发出去的切片，读者会跟着脏")
	fmt.Println(strings.Repeat("-", 60))
	next.symbols[0] = "DOGEUSDT" // 和 Store 出去的是同一块底层数组
	fmt.Printf("  改写者手里的切片后，Load 看到 %v\n", live.Load().(botCfg).symbols)
	fmt.Println("  ➡️  Store 之前要复制切片。发出去的快照当成只读。")

	owned := []string{"BTCUSDT"}
	live.Store(botCfg{threshold: "0.003", symbols: append([]string(nil), owned...)})
	owned[0] = "改不到读者"
	fmt.Printf("  复制进 Store 之后再改自己的切片，Load 仍是 %v\n", live.Load().(botCfg).symbols)

	fmt.Println("\n🔍 示例 3：具体类型必须始终一致")
	fmt.Println(strings.Repeat("-", 60))
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("  ⚠️  Store(*botCfg) panic：%v\n", r)
			}
		}()
		live.Store(&botCfg{threshold: "boom"})
	}()
	fmt.Println("  ➡️  一开始存的是 botCfg 值，后面就不能改存指针。")

	fmt.Println("\n🧪 示例 4：热路径连续读，中间换配置")
	fmt.Println(strings.Repeat("-", 60))
	done := make(chan struct{})
	go func() {
		defer close(done)
		var last string
		deadline := time.Now().Add(30 * time.Millisecond)
		for time.Now().Before(deadline) {
			c := live.Load().(botCfg)
			if c.threshold != last {
				fmt.Printf("  读者看到阈值从 %q 变成 %q\n", last, c.threshold)
				last = c.threshold
			}
		}
	}()
	time.Sleep(10 * time.Millisecond)
	live.Store(botCfg{threshold: "0.010", symbols: []string{"SOLUSDT"}})
	<-done

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ 发布整份配置：复制 → Store；热路径只 Load")
	fmt.Println("  ✅ 读到的是某一份完整快照，不会读到写一半的字段")
	fmt.Println("  ✅ 具体类型一旦确定不能换；切片要复制，发出去就只读")
	fmt.Println("  ✅ 单个计数用 atomic.Int64；整份结构用 Value / atomic.Pointer")
	fmt.Println("  ✅ 配置巨大、又要局部改字段时，用 RWMutex 更合适")
	fmt.Println(strings.Repeat("=", 60) + "\n")
}
