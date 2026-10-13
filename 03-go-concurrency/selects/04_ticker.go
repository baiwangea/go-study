//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 03-go-concurrency && go run selects/04_ticker.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"strings"
	"time"
)

/*
═══════════════════════════════════════════════════════════════════════
L2-10：Ticker 轮询与 Stop
═══════════════════════════════════════════════════════════════════════

【核心概念】
    用 time.NewTicker 做周期性任务（每 60ms 拉一次价），并且必须 Stop 释放资源。
    ticker 通道容量为 1：一轮处理超时的话，未消费的 tick 会被丢弃。

【你会看到】
    ✓ ticker 每 60ms 触发一次，循环 3 次后 break
    ✓ 不 Stop 时 timer 仍挂在运行时里

【思考题】（动手改代码验证，别空想）
    1. time.Tick(ch) 与 time.NewTicker 有什么区别？
       为什么 Tick 的通道没法关、容易泄漏？
    2. 把 defer ticker.Stop() 删掉并循环 100 万次，
       用 runtime.NumGoroutine() 或 pprof 看看差异
    3. 如果一轮拉取耗时超过 60ms，未消费的 tick 会堆积吗？
       （提示：ticker 通道容量为 1，会丢 tick ——
       所以要做「处理完再等下一次」）

【通关标准】
    能写出正确的轮询循环（带 Stop），并解释为什么「tick 丢失」
    在行情场景需要特殊处理

【前置】L2-09              【标签】ticker · 定时 · 泄漏
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L2-10 Ticker 定时")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// ── 示例 1：每 60ms 轮询一次，3 次后退出 ────────────────────────────
	fmt.Println("📝 示例 1：周期性轮询并记得 Stop")
	fmt.Println(strings.Repeat("-", 60))
	ticker := time.NewTicker(60 * time.Millisecond)
	defer ticker.Stop() // 关键：离开作用域就释放

	rounds := 0
	start := time.Now()
	for range ticker.C {
		rounds++
		fmt.Printf("  第 %d 次轮询：假装请求交易所（耗时 10ms）\n", rounds)
		time.Sleep(10 * time.Millisecond)
		if rounds == 3 {
			break // 真实机器人里换成 <-ctx.Done()，见 L2-11
		}
	}
	fmt.Printf("  共轮询 %d 次，耗时 %v；已 defer ticker.Stop()，无残留 timer\n", rounds, time.Since(start).Round(time.Millisecond))

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ time.NewTicker 每隔固定间隔往 C 里发一次时间戳")
	fmt.Println("  ✅ 一定要 defer ticker.Stop()，否则底层 timer 一直挂着（泄漏）")
	fmt.Println("  ✅ ticker 通道容量为 1，处理慢了会丢 tick，不会堆积")
	fmt.Println("  ✅ 真实循环的退出条件应该是 <-ctx.Done()，而不是数到 3")
}
