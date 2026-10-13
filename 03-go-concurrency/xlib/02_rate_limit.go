//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 03-go-concurrency && go run xlib/02_rate_limit.go
// 依赖：golang.org/x/time（本关需要，见 03-go-concurrency/go.mod）
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

/*
═══════════════════════════════════════════════════════════════════════
L2-20：限频 —— golang.org/x/time/rate
═══════════════════════════════════════════════════════════════════════

【核心概念】
    rate.Limiter 是「令牌桶」：桶里最多装 burst 个令牌，按 limit/s 的速度补充。
    取到令牌就放行；取不到时，Wait 会阻塞等，Allow 会直接返回 false。
    交易所 API 的「10 req/s」限频就是标准令牌桶场景。

【你会看到】
    ✓ 前 burst 个请求立刻通过，之后被均匀拉开间隔
    ✓ Wait 阻塞等待 vs Allow 立即判断
    ✓ 20 req/s 发 25 个请求 ≈ 1 秒（(25-5)/20）
    ✓ WaitN 批量申请令牌；n 超过 burst 会直接报错
    ✓ 手搓 time.Tick 为什么不行（没有突发容量、不能共享、不能调速率）

【思考题】（动手改代码验证，别空想）
    1. 把 burst 从 3 改成 1，示例 1 的前三个请求还会「立刻」通过吗？
    2. 多个 goroutine 共用一个 Limiter，总速率是多少？各用一个 Limiter 呢？
    3. 交易所把限频从 20/s 提到 50/s，代码怎么改？（提示：SetLimit）
    4. Wait 传入的 ctx 被取消会怎样？为什么工程上不该随手传
       context.Background()？

【通关标准】
    能说清「令牌桶 = 平均速率 + 突发容量」，并会用 Wait / Allow / WaitN
    三种姿势应对不同请求

【前置】L2-10              【标签】rate.Limiter · 令牌桶 · burst · 交易所限频
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L2-20 rate 限频")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// ── 示例 1：令牌桶与 burst ──────────────────────────────────────────
	fmt.Println("📝 示例 1：令牌桶：桶里先有 burst 个令牌，谁先来谁拿走")
	fmt.Println(strings.Repeat("-", 60))
	lim := rate.NewLimiter(rate.Limit(10), 3) // 10 req/s，桶容量 3
	fmt.Printf("  limit = %v/s，burst = %d，初始令牌 = %.0f 个\n", lim.Limit(), lim.Burst(), lim.Tokens())
	for i := 1; i <= 3; i++ {
		fmt.Printf("  第 %d 个请求：Allow() = %v（消耗一个令牌，立刻放行）\n", i, lim.Allow())
	}
	fmt.Printf("  第 4 个请求：Allow() = %v（桶空了，要等 100ms 才补回 1 个）\n", lim.Allow())
	fmt.Println("  ➡️  burst 就是「能吸收多大的瞬时突发」：10/s 是平均速率，3 是突发容量")

	// ── 示例 2：Wait 阻塞等待 ───────────────────────────────────────────
	fmt.Println("📌 示例 2：Wait：取不到令牌就阻塞等待（适合「必须发出去」的请求）")
	fmt.Println(strings.Repeat("-", 60))
	lim2 := rate.NewLimiter(rate.Limit(10), 1)
	start := time.Now()
	for i := 1; i <= 5; i++ {
		if err := lim2.Wait(context.Background()); err != nil {
			fmt.Println("  出错：", err)
			break
		}
		fmt.Printf("  第 %d 个请求放行，t=%v\n", i, time.Since(start).Round(time.Millisecond))
	}
	fmt.Println("  ➡️  间隔稳定在 100ms：Wait 把「速率限制」变成了「排队等待」")

	// ── 示例 3：交易所限频的真实场景 ────────────────────────────────────
	fmt.Println("🔍 示例 3：交易所场景：20 req/s 拉 25 个交易对的行情")
	fmt.Println(strings.Repeat("-", 60))
	const qps, burst, total = 20, 5, 25
	lim3 := rate.NewLimiter(rate.Limit(qps), burst)
	start = time.Now()
	for i := 1; i <= total; i++ {
		if err := lim3.Wait(context.Background()); err != nil {
			break
		}
		// 真实项目里这里就是 client.Do(req) —— 请求本身也要耗时间
	}
	elapsed := time.Since(start)
	want := time.Duration(total-burst) * time.Second / qps
	fmt.Printf("  25 个请求共耗时 %v（理论值 (25-5)/20 = %v）\n",
		elapsed.Round(time.Millisecond), want)
	fmt.Println("  ➡️  前 5 个吃掉 burst 立刻走，剩下 20 个按 20/s 排队 —— 全程不会触发交易所的 429")

	// ── 示例 4：WaitN 批量申请 ──────────────────────────────────────────
	fmt.Println("🧪 示例 4：WaitN：一次申请多个令牌（批量下单、批量查询）")
	fmt.Println(strings.Repeat("-", 60))
	lim4 := rate.NewLimiter(rate.Limit(10), 10)
	start = time.Now()
	if err := lim4.WaitN(context.Background(), 10); err != nil {
		fmt.Println("  出错：", err)
	}
	fmt.Printf("  第一次 WaitN(10)：桶是满的，立刻返回，t=%v\n", time.Since(start).Round(time.Millisecond))
	start = time.Now()
	if err := lim4.WaitN(context.Background(), 10); err != nil {
		fmt.Println("  出错：", err)
	}
	fmt.Printf("  第二次 WaitN(10)：桶空了，等了 %v（10 个令牌 ÷ 10/s）\n",
		time.Since(start).Round(time.Millisecond))
	errN := lim4.WaitN(context.Background(), 11)
	fmt.Printf("  WaitN(11) 而 burst=10 → 立刻返回错误：%v\n", errN)
	fmt.Println("  ➡️  超过 burst 的批量请求永远等不到，rate 选择立刻报错而不是傻等")

	// ── 示例 5：为什么不能手搓 time.Tick ────────────────────────────────
	fmt.Println("⚙️ 示例 5：为什么不能手搓 time.Tick")
	fmt.Println(strings.Repeat("-", 60))
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	start = time.Now()
	for i := 1; i <= 3; i++ {
		<-tick.C
		fmt.Printf("  手搓版第 %d 个请求放行，t=%v\n", i, time.Since(start).Round(time.Millisecond))
	}
	fmt.Println("  对比 Limiter(10/s, burst=3)：那三个请求在 t=0 就全走完了")
	fmt.Println("  手搓 time.Tick 的另外几个硬伤：")
	fmt.Println("    • 固定间隔，吸收不了抖动，突发流量只能一个个排队")
	fmt.Println("    • 每个调用方各自 Tick，总速率直接翻倍（没有全局共享的桶）")
	fmt.Println("    • 不能动态调整（交易所改限频只能重启）；rate 有 SetLimit / SetBurst")
	fmt.Println("    • time.Tick 无法 Stop → timer 泄漏；只能改用 NewTicker + Stop")
	fmt.Println("    • 不能批量申请 n 个令牌，也分不出「等待」和「拒绝」两种策略")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ 令牌桶 = 平均速率(limit) + 突发容量(burst)；令牌按 limit/s 匀速补充")
	fmt.Println("  ✅ Wait 阻塞等令牌（必须发出去）；Allow 立即判断（可以丢）；WaitN 批量等")
	fmt.Println("  ✅ 一个 Limiter 应该被所有调用方共享，否则总速率 = 各桶之和")
	fmt.Println("  ✅ 该用：调用外部 API / 数据库前的限速，保护下游也保护自己不被封")
	fmt.Println("  ✅ 不该用：需要「固定周期触发」用 time.Ticker（心跳、定时任务）；")
	fmt.Println("    需要「严格并发数」用信号量或 errgroup.SetLimit —— 三者目的不同")
}
