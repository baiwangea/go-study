//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 03-go-concurrency && go run selects/03_nonblock.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"strings"
	"time"
)

/*
═══════════════════════════════════════════════════════════════════════
L2-09：非阻塞 select 与丢弃策略
═══════════════════════════════════════════════════════════════════════

【核心概念】
    加 default 让 select 立刻返回：宁可丢数据，也不把主流程拖死。
    default 让 select 从「阻塞等待」变成「试一下就走」。

【你会看到】
    ✓ 通道空时走 default 丢弃；有数据时正常读取
    ✓ 发送端队列满也走 default 丢单

【思考题】（动手改代码验证，别空想）
    1. 把 default 删掉，第一次读取会发生什么？（阻塞到 100ms 后有人发送）
       这说明 default 改变了什么？
    2. 行情场景里丢 tick 可以接受吗？什么数据绝对不能丢（订单回报）？
       两者分别该用哪种通道？
    3. 用 default 做「非阻塞发送」时，被丢弃的任务去哪了？
       怎么让它至少留下日志/指标？

【通关标准】
    能写出非阻塞收/发两种写法，并说清「可丢」与「不可丢」数据的通道选型

【前置】L2-08              【标签】default · 非阻塞 · 降级
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L2-09 非阻塞收发")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// ── 示例 1：非阻塞接收（通道为空 → 走 default） ─────────────────────
	fmt.Println("📝 示例 1：通道为空，立刻走 default")
	fmt.Println(strings.Repeat("-", 60))
	ticks := make(chan string, 1)

	// 接收侧：此刻没有行情，立刻走 default，不浪费时间
	select {
	case t := <-ticks:
		fmt.Println("  读到 tick：", t)
	default:
		fmt.Println("  接收 default：暂无行情 → 直接返回去做别的事（不阻塞）")
	}

	// ── 示例 2：非阻塞接收（有数据 → 正常读到） ─────────────────────────
	fmt.Println("📌 示例 2：有数据时正常读取")
	fmt.Println(strings.Repeat("-", 60))
	go func() { time.Sleep(50 * time.Millisecond); ticks <- "BTC 63001" }()
	time.Sleep(100 * time.Millisecond)
	select {
	case t := <-ticks:
		fmt.Println("  读到 tick：", t)
	default:
		fmt.Println("  接收 default：暂无行情")
	}

	// ── 示例 3：非阻塞发送（队列满 → 丢弃） ─────────────────────────────
	fmt.Println("🔍 示例 3：队列满时非阻塞发送 → 丢弃")
	fmt.Println(strings.Repeat("-", 60))
	// 发送侧：队列满（容量 1 已占），下 2 条会被丢弃
	orders := make(chan string, 1)
	orders <- "ORD-1 已入队"
	dropped := 0
	for _, o := range []string{"ORD-2", "ORD-3"} {
		select {
		case orders <- o:
			fmt.Printf("  %s 入队成功\n", o)
		default:
			dropped++
			fmt.Printf("  %s 队列已满 → 丢弃（不可丢的数据不能这么写！）\n", o)
		}
	}
	fmt.Printf("  当前队列长度 %d，丢弃 %d 条\n", len(orders), dropped)

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ default 让 select 不阻塞：有就绪就走，没有就立刻返回")
	fmt.Println("  ✅ 非阻塞接收：select { case v := <-ch: ...; default: ... }")
	fmt.Println("  ✅ 非阻塞发送：select { case ch <- v: ...; default: ... }")
	fmt.Println("  ✅ 可丢数据（行情 tick）才用 default 丢弃；订单回报必须可靠送达")
}
