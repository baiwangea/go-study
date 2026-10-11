//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 04-goframe && go run 10_cron_task.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/gogf/gf/v2/os/gcron"
)

/*
═══════════════════════════════════════════════════════════════════════
L3-10：gcron：秒级 cron 与任务去重
═══════════════════════════════════════════════════════════════════════

【核心概念】
    用 gcron 替代手写 ticker：支持秒级六段表达式、命名任务、单实例与自动退出控制。
    命名任务便于后续 Remove；AddSingleton 保证长任务不并发重入。

【你会看到】
    ✓ 命名任务跑了约 2~3 次后被 Remove（次数取决于与秒边界的对齐）；AddOnce 只执行一次；延迟任务 200ms 后触发

【思考题】（动手改代码验证，别空想）
    1. 表达式为什么是 6 段而不是 5 段？把每秒一次的写法改成「每 2 秒一次」（秒字段写 星号斜杠2），频率如何变化？
    2. Add 与 AddSingleton 的差别是什么？行情拉取超过 1 秒时，用 Add 会发生什么？
    3. 机器人里「下单 15 秒未成交就撤单」该用 DelayAddOnce 还是队列延迟任务（L3-11）？各自的可靠性差别？

【通关标准】
    会注册/停止命名定时任务，并能解释为什么长任务必须用 Singleton

【前置】L2-10（ticker 与 Stop）      【标签】gcron · 六段表达式 · AddOnce/AddTimes/Remove
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n=== L3-10 gcron：秒级 cron 与任务去重 ===")
	ctx := context.Background()
	var ticks, once, delayed atomic.Int64

	// ── 示例 1：命名周期任务（便于后续 Remove） ──────────────────────────
	fmt.Println("\n【示例 1】每秒一次的命名任务 poll-quote")
	if _, err := gcron.Add(ctx, "* * * * * *", func(ctx context.Context) {
		n := ticks.Add(1)
		fmt.Printf("  [%s] 轮询行情 第 %d 次\n", time.Now().Format("15:04:05"), n)
	}, "poll-quote"); err != nil {
		fmt.Println("  注册失败：", err)
		return
	}

	// ── 示例 2：AddOnce 只跑一次；DelayAddOnce 延迟触发 ──────────────────
	fmt.Println("\n【示例 2】AddOnce 只执行一次，延迟任务 200ms 后触发")
	_, _ = gcron.AddOnce(ctx, "* * * * * *", func(ctx context.Context) {
		once.Add(1)
	}, "report-once")

	gcron.DelayAddOnce(ctx, 200*time.Millisecond, "* * * * * *", func(ctx context.Context) {
		delayed.Add(1)
	}, "delayed")

	// ── 示例 3：跑一会儿后 Remove，验证命名任务被真正停掉 ────────────────
	time.Sleep(2500 * time.Millisecond)
	gcron.Remove("poll-quote") // 停止并移除命名任务
	before := ticks.Load()
	time.Sleep(1100 * time.Millisecond)

	fmt.Printf("  移除前跑了 %d 次，移除后仍是 %d 次 → Remove 生效（命名任务不清理会常驻）\n", before, ticks.Load())
	fmt.Printf("  AddOnce 执行 %d 次；延迟任务执行 %d 次（200ms 后触发）\n", once.Load(), delayed.Load())

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n【关键总结】")
	fmt.Println("  • gcron 表达式是 6 段（多出最左边的秒），\"* * * * * *\" 即每秒一次")
	fmt.Println("  • 命名任务必须 Remove，否则常驻；长任务要用 AddSingleton 防并发重入")
	fmt.Println("  • AddOnce / DelayAddOnce 用于一次性与延迟任务")
}
