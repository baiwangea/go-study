//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 03-go-concurrency && go run selects/01_multiplex.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"strings"
	"time"
)

/*
═══════════════════════════════════════════════════════════════════════
L2-07：select 多路复用
═══════════════════════════════════════════════════════════════════════

【核心概念】
    用 select 同时等多个 channel：谁先就绪就走谁，都阻塞才等待。
    多个分支同时就绪时会随机挑一个（故意防饿死）。

【你会看到】
    ✓ 两个源快慢不同 → 先命中快的那个
    ✓ 两个同时就绪时会随机挑（防饿死）

【思考题】（动手改代码验证，别空想）
    1. 把 fast 的延时从 50ms 改成 300ms，输出会换成哪个分支？
       说明 select 的判定依据是什么？
    2. 把两个 case 的延时改成完全相同并循环 10 次，命中分布如何？
       为什么 Go 要故意随机？
    3. select 里如果有 default 分支，语义会怎么变？（下一关实测）

【通关标准】
    能说出 select 的三条规则：有就绪就执行、多个就绪随机选、
    都未就绪则阻塞（除非有 default）

【前置】L2-06              【标签】select · case · 随机公平
═══════════════════════════════════════════════════════════════════════
*/

// pick4 连续从 a、b 两个已就绪的通道里各选 4 次，返回命中顺序，
// 用来观察「多个分支同时就绪时 select 会随机挑」。
func pick4(a, b chan int) []string {
	var hits []string
	for i := 0; i < 4; i++ {
		select {
		case <-a:
			hits = append(hits, "a")
		case <-b:
			hits = append(hits, "b")
		}
	}
	return hits
}

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L2-07 select 多路复用")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// ── 示例 1：谁先就绪就走谁 ──────────────────────────────────────────
	fmt.Println("📝 示例 1：快的那个先命中")
	fmt.Println(strings.Repeat("-", 60))
	fast := make(chan string)
	slow := make(chan string)

	go func() { time.Sleep(50 * time.Millisecond); fast <- "BTC 63000" }()
	go func() { time.Sleep(200 * time.Millisecond); slow <- "ETH 3000" }()

	start := time.Now()
	select {
	case p := <-fast:
		fmt.Printf("  命中 fast 分支：%s（等待 %v）\n", p, time.Since(start).Round(10*time.Millisecond))
	case p := <-slow:
		fmt.Printf("  命中 slow 分支：%s（等待 %v）\n", p, time.Since(start).Round(10*time.Millisecond))
	}

	// ── 示例 2：多个分支同时就绪 → 随机挑 ───────────────────────────────
	fmt.Println("📌 示例 2：同时就绪时随机选择")
	fmt.Println(strings.Repeat("-", 60))
	// 两个分枝同时就绪（预填缓冲，保证都有数据）：连选 4 次，看它随机挑选
	a, b := make(chan int, 2), make(chan int, 2)
	a <- 1
	a <- 1
	b <- 1
	b <- 1
	fmt.Println("  同时就绪时的选择（应为随机混排，如 [a b a b]）：", pick4(a, b))

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ select 同时等多个通道，有就绪就执行那个 case")
	fmt.Println("  ✅ 多个 case 同时就绪时，运行时随机挑一个（防某分支饿死）")
	fmt.Println("  ✅ 所有 case 都未就绪时，select 阻塞等待（除非写了 default）")
	fmt.Println("  ✅ 别假设 select 的命中顺序，它是故意不保证的")
}
