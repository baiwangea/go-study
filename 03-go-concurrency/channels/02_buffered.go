//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 03-go-concurrency && go run channels/02_buffered.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"strings"
	"time"
)

/*
═══════════════════════════════════════════════════════════════════════
L2-04：缓冲队列 —— 什么时候开始堵
═══════════════════════════════════════════════════════════════════════

【核心概念】
    make(chan T, n) 的语义：前 n 次发送不阻塞，第 n+1 次开始阻塞。
    这就是「背压」：缓冲区填满后，生产者会被迫慢下来等消费者。

【你会看到】
    ✓ 容量 2 的通道，前 2 条秒过，第 3 条卡住直到消费者取走一个

【常见陷阱案例】
    ❌ 错误写法 1：缓冲区设太大，积压过期数据
       priceCh := make(chan float64, 1000)  // 缓冲 1000 条行情
       go func() {
           for price := range priceCh {
               placeOrder(price)  // 价格可能是 5 秒前的，已经过期！
           }
       }()
       后果：用旧价格下单，亏损

    ❌ 错误写法 2：缓冲区设太小，生产者被拖慢
       logCh := make(chan string, 1)  // 只能缓冲 1 条
       // 日志写入慢（磁盘 I/O），第 2 条日志就会阻塞业务逻辑
       后果：业务线程被日志拖慢，延迟飙升

    ❌ 错误写法 3：忘记 close，消费者永远阻塞
       ch := make(chan int, 10)
       for i := 0; i < 10; i++ {
           ch <- i
       }
       // 忘了 close(ch)
       for v := range ch {  // range 永远不退出，等下一个值
           fmt.Println(v)
       }
       后果：goroutine 泄漏，内存泄漏

【实战场景】
    ✅ 行情通道（容忍短暂积压，但不能太大）
    ✅ 日志通道（缓冲区吸收突发流量，避免阻塞业务）
    ✅ 任务队列（缓冲区 = 排队等待的任务数）

【缓冲区大小选型】
    • 太小 → 生产者频繁阻塞，拖慢业务
    • 太大 → 数据积压过期，内存浪费
    • 经验值：
      - 行情通道：1-10（实时性优先）
      - 日志通道：100-1000（吸收突发）
      - 任务队列：worker 数量的 2-3 倍

【思考题】（动手改代码验证，别空想）
    1. 把容量改成 0 和改成 5，分别观察第几条开始阻塞 —— 容量与「谁先阻塞」的关系是什么？
    2. len(ch) 与 cap(ch) 分别告诉你什么？打印出来看看。
    3. 机器人里行情通道该设多大？设太大会怎样（价格过期还在下单），
       设太小会怎样（生产者被拖慢）？
    4. 如果消费者完全停止（panic 或 goroutine 退出），生产者会怎样？
       （提示：缓冲区满后，生产者永远阻塞 —— goroutine 泄漏）

【通关标准】
    能解释缓冲 = 生产者与消费者的速度差容忍度，并说清 len/cap 的区别

【前置】L2-03              【标签】buffered chan · 背压 · 队列深度
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L2-04 有缓冲 channel")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// ── 示例 1：容量 2，第 3 条开始被背压 ───────────────────────────────
	fmt.Println("📝 示例 1：容量 2 的队列，第 3 条卡住")
	fmt.Println(strings.Repeat("-", 60))
	ch := make(chan string, 2) // ⚠️ 缓冲容量 2：前 2 次发送不阻塞

	for i, tick := range []string{"tick-1", "tick-2", "tick-3"} {
		start := time.Now()
		// ⚠️ 关键行：第 1、2 次发送立即返回，第 3 次阻塞
		ch <- tick
		blocked := time.Since(start)
		// len(ch) = 当前排队的消息数；cap(ch) = 队列总容量
		fmt.Printf("  发送 %s 用时 %v，当前 len=%d cap=%d\n", tick, blocked.Round(10*time.Millisecond), len(ch), cap(ch))

		if i == 0 {
			// 启动消费者，200ms 后取走一个消息
			go func() {
				time.Sleep(200 * time.Millisecond)
				fmt.Printf("  [消费者] 取走 %s\n", <-ch)
			}()
		}
	}
	fmt.Println("  ➡️  第 3 条被「背压」住了：队列容量 2 已满，只能等消费者腾位置")
	time.Sleep(300 * time.Millisecond)

	// ── 示例 2：实战场景 —— 行情通道缓冲区选型 ───────────────────────────
	fmt.Println("\n🔥 示例 2：真实场景 —— 行情通道（模拟）")
	fmt.Println(strings.Repeat("-", 60))

	// 场景：行情推送很快（每 10ms 一条），策略计算慢（每 50ms 一次）
	// 缓冲区设置：5 条（容忍短暂积压，但不能太多）
	priceCh := make(chan float64, 5)

	// 模拟行情推送（生产者快）
	go func() {
		price := 10000.0
		for i := 0; i < 10; i++ {
			priceCh <- price + float64(i)*0.1
			fmt.Printf("  [行情] 推送价格 %.1f，队列 len=%d\n", price+float64(i)*0.1, len(priceCh))
			time.Sleep(10 * time.Millisecond)
		}
		close(priceCh) // ✅ 发送方关闭通道，告知消费者没有更多数据
	}()

	// 模拟策略计算（消费者慢）
	time.Sleep(30 * time.Millisecond) // 延迟启动，模拟初始化
	count := 0
	for price := range priceCh { // range 自动处理通道关闭
		count++
		fmt.Printf("  [策略] 处理价格 %.1f（第 %d 条）\n", price, count)
		time.Sleep(50 * time.Millisecond) // 策略计算耗时
		if count >= 3 {
			break // 演示提前退出
		}
	}
	fmt.Println("  ➡️  缓冲区吸收了速度差，但积压太多会导致用旧数据下单")

	// ── 反面教材：错误写法演示 ─────────────────────────────────────────
	fmt.Println("\n❌ 反面教材：不要这么写")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("  // 错误写法 1：缓冲区太大，数据过期")
	fmt.Println("  priceCh := make(chan float64, 1000)  // 缓冲 1000 条")
	fmt.Println("  // 策略慢，队列积压 → 用 5 秒前的价格下单")
	fmt.Println("  // 后果：价格过期，亏损")
	fmt.Println()
	fmt.Println("  // 错误写法 2：缓冲区太小，生产者被拖慢")
	fmt.Println("  logCh := make(chan string, 1)")
	fmt.Println("  logCh <- \"log1\"  // OK")
	fmt.Println("  logCh <- \"log2\"  // 阻塞，等消费者取走 log1")
	fmt.Println("  // 后果：业务线程被日志拖慢")
	fmt.Println()
	fmt.Println("  // 错误写法 3：忘记 close")
	fmt.Println("  ch := make(chan int, 10)")
	fmt.Println("  for i := 0; i < 10; i++ { ch <- i }")
	fmt.Println("  // 忘了 close(ch)")
	fmt.Println("  for v := range ch { /* 永远不退出 */ }")

	// ── 缓冲区大小选型指南 ───────────────────────────────────────────────
	fmt.Println("\n📐 缓冲区大小选型指南")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("  • 行情通道：1-10（实时性优先，过期数据无意义）")
	fmt.Println("  • 日志通道：100-1000（吸收突发流量，避免阻塞业务）")
	fmt.Println("  • 任务队列：worker 数量的 2-3 倍（平衡内存和响应）")
	fmt.Println()
	fmt.Println("  原则：")
	fmt.Println("    ✓ 太小 → 生产者频繁阻塞，拖慢业务")
	fmt.Println("    ✓ 太大 → 数据积压过期，内存浪费")
	fmt.Println("    ✓ 根据生产/消费速度差和数据时效性决定")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ make(chan T, n)：前 n 次发送立即返回，第 n+1 次开始阻塞")
	fmt.Println("  ✅ cap(ch) = 队列总容量；len(ch) = 当前排队条数")
	fmt.Println("  ✅ 缓冲 = 容忍生产/消费速度差的额度，满了就形成背压")
	fmt.Println("  ✅ 行情通道别设太大：积压的旧价格过期后还在下单是灾难")
	fmt.Println()
	fmt.Println("  📌 实战建议：")
	fmt.Println("    • 实时行情：缓冲 1-10，丢弃积压的旧数据（见 L2-09 default）")
	fmt.Println("    • 日志写入：缓冲 100-1000，避免阻塞业务线程")
	fmt.Println("    • 关闭通道：只有发送方 close，接收方 range 自动退出（见 L2-05）")
}
