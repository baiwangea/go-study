//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 03-go-concurrency && go run channels/01_unbuffered.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"strings"
	"time"
)

/*
═══════════════════════════════════════════════════════════════════════
L2-03：无缓冲 channel —— 发送即交接
═══════════════════════════════════════════════════════════════════════

【核心概念】
    无缓冲 channel 的两端会「碰面」：发送阻塞到有人接收，
    接收阻塞到有人发送。一次收发 = 一次同步会合（handshake）。

【你会看到】
    ✓ 发送方被阻塞约 200ms（直到接收方起床取走），随后两端几乎同时打印

【常见陷阱案例】
    ❌ 错误写法 1：单协程自己发自己收
       ch := make(chan int)
       ch <- 1          // 阻塞在这里，永远等不到接收方
       x := <-ch        // 永远执行不到这一行
       后果：fatal error: all goroutines are asleep - deadlock!

    ❌ 错误写法 2：发送后没人接收
       ch := make(chan int)
       go func() {
           ch <- 42     // 发送后阻塞，等接收方
       }()
       // 主协程退出 → 整个程序结束，接收方永远不出现
       后果：发送协程泄漏，订单指令丢失

    ❌ 错误写法 3：接收方比发送方晚启动太久
       ch := make(chan int)
       time.Sleep(5 * time.Second)  // 延迟 5 秒
       go func() { <-ch }()
       ch <- 1  // 发送方阻塞 5 秒，等接收方起床
       后果：下单延迟 5 秒，价格已经变了

【实战场景】
    ✅ 下单指令通道（必须交到执行方手里才算成功）
    ✅ 任务分发（worker 取走任务才能继续发下一个）
    ✅ 信号量同步（确保某段代码严格串行执行）

【无缓冲 vs 有缓冲 对比】
    ┌──────────────┬────────────────────┬────────────────────┐
    │              │  无缓冲 make(chan T)│ 有缓冲 make(chan T, n)│
    ├──────────────┼────────────────────┼────────────────────┤
    │ 发送阻塞条件  │ 没有接收方          │ 队列满了            │
    │ 接收阻塞条件  │ 没有发送方          │ 队列空了            │
    │ 适用场景     │ 严格同步、指令流     │ 流量削峰、异步通知   │
    │ 容忍速度差   │ 0（必须配对）        │ n（队列容量）        │
    └──────────────┴────────────────────┴────────────────────┘

【思考题】（动手改代码验证，别空想）
    1. 把接收方 goroutine 里的 time.Sleep 删掉，发送方还会阻塞 200ms 吗？
       为什么？（这就验证了「阻塞时长取决于对端」）
    2. ch := make(chan int) 然后 ch <- 1 就结束，程序报什么错？
       （all goroutines are asleep - deadlock!）
    3. 机器人里什么时候该用无缓冲？
       （提示：下单指令必须被执行方取走才算交出去）
    4. 如果接收方 panic 了，发送方会怎样？
       （提示：永远阻塞，goroutine 泄漏 —— 见 L2-11 context 取消）

【通关标准】
    能说出无缓冲 channel 的两个阻塞条件，并解释它为什么等价于一次「同步会合」

【前置】L2-02              【标签】chan · 同步 · 阻塞语义
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L2-03 无缓冲 channel")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// ── 示例 1：发送方原地等接收方起床 ──────────────────────────────────
	fmt.Println("📝 示例 1：没有接收者，发送方就堵在原地")
	fmt.Println(strings.Repeat("-", 60))
	ch := make(chan string) // ⚠️ 无缓冲：cap(ch) == 0

	// 接收方故意晚 200ms 才起床 —— 发送方必须原地等它
	go func() {
		time.Sleep(200 * time.Millisecond)
		got := <-ch // 接收操作：从 channel 取值
		fmt.Printf("  [接收方] 收到指令：%s\n", got)
	}()

	start := time.Now()
	// ⚠️ 关键行：无缓冲 channel 的发送操作会阻塞，直到有接收方取走
	// 这里会等待约 200ms，直到上面的 goroutine 执行 <-ch
	ch <- "BUY BTC 0.1"
	fmt.Printf("  [发送方] 阻塞了 %v 才完成交接\n", time.Since(start).Round(10*time.Millisecond))
	time.Sleep(100 * time.Millisecond) // 等接收方把日志打完

	// ── 示例 2：实战场景 —— 下单指令必须交到执行方手里 ─────────────────
	fmt.Println("\n🔥 示例 2：真实场景 —— 下单指令通道（模拟）")
	fmt.Println(strings.Repeat("-", 60))

	type Order struct {
		Symbol string
		Side   string
		Qty    float64
	}

	orderCh := make(chan Order) // 无缓冲：确保指令被接收方拿到

	// 模拟订单执行器
	go func() {
		order := <-orderCh // 阻塞等待指令
		fmt.Printf("  [执行器] 收到订单：%s %s %.2f\n", order.Side, order.Symbol, order.Qty)
		time.Sleep(50 * time.Millisecond) // 模拟下单延迟
		fmt.Println("  [执行器] ✓ 订单已发送到交易所")
	}()

	time.Sleep(30 * time.Millisecond) // 模拟策略计算延迟
	fmt.Println("  [策略] 生成订单...")
	orderCh <- Order{Symbol: "BTCUSDT", Side: "BUY", Qty: 0.1}
	// ✅ 发送操作返回时，我们确定执行器已经收到订单
	fmt.Println("  [策略] ✓ 订单已交给执行器")
	time.Sleep(60 * time.Millisecond) // 等执行器完成

	// ── 反面教材：错误写法演示 ─────────────────────────────────────────
	fmt.Println("\n❌ 反面教材：不要这么写")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("  // 错误写法 1：单协程自己发自己收")
	fmt.Println("  ch := make(chan int)")
	fmt.Println("  ch <- 1     // 阻塞在这里，永远等不到接收方")
	fmt.Println("  x := <-ch   // 永远执行不到")
	fmt.Println("  // 后果：fatal error: all goroutines are asleep - deadlock!")
	fmt.Println()
	fmt.Println("  // 错误写法 2：发送后没人接收")
	fmt.Println("  ch := make(chan int)")
	fmt.Println("  go func() {")
	fmt.Println("      ch <- 42  // 永远阻塞，等不到接收方")
	fmt.Println("  }()")
	fmt.Println("  // main 退出 → goroutine 泄漏，订单丢失")
	fmt.Println()
	fmt.Println("  // 错误写法 3：接收方太晚启动")
	fmt.Println("  ch := make(chan int)")
	fmt.Println("  time.Sleep(5 * time.Second)  // 延迟 5 秒")
	fmt.Println("  go func() { <-ch }()")
	fmt.Println("  ch <- 1  // 发送方阻塞 5 秒，价格已经变了")

	// ── 对比表格 ─────────────────────────────────────────────────────────
	fmt.Println("\n📊 无缓冲 vs 有缓冲 对比")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("  ┌──────────────┬────────────────────┬────────────────────┐")
	fmt.Println("  │              │  无缓冲 make(chan T)│ 有缓冲 make(chan T, n)│")
	fmt.Println("  ├──────────────┼────────────────────┼────────────────────┤")
	fmt.Println("  │ 发送阻塞条件  │ 没有接收方          │ 队列满了            │")
	fmt.Println("  │ 接收阻塞条件  │ 没有发送方          │ 队列空了            │")
	fmt.Println("  │ 适用场景     │ 严格同步、指令流     │ 流量削峰、异步通知   │")
	fmt.Println("  │ 容忍速度差   │ 0（必须配对）        │ n（队列容量）        │")
	fmt.Println("  └──────────────┴────────────────────┴────────────────────┘")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ 无缓冲 channel 的发送：没有接收者就阻塞")
	fmt.Println("  ✅ 无缓冲 channel 的接收：没有发送者就阻塞")
	fmt.Println("  ✅ 阻塞时长由「对端什么时候到」决定，等价于一次同步会合")
	fmt.Println("  ✅ 适合「必须交到对方手里」的指令流，如下单")
	fmt.Println()
	fmt.Println("  📌 实战建议：")
	fmt.Println("    • 下单指令：用无缓冲 channel，确保执行方收到")
	fmt.Println("    • 行情通知：用有缓冲 channel，允许短暂积压（见 L2-04）")
	fmt.Println("    • 超时控制：配合 select + time.After（见 L2-08）")
}
