//go:build level

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/chichengyu/async"
)

/*
═══════════════════════════════════════════════════════════════════════
A-05：RateLimiter 限流保护
═══════════════════════════════════════════════════════════════════════

【核心概念】
    限流器保护下游服务不被打爆，三种算法：
    
    1. 令牌桶（Token Bucket）：平滑限流 + 允许突发
    2. 滑动窗口（Sliding Window）：精确限流
    3. 自适应限流（Adaptive）：根据延迟动态调整
    
    适用场景：
      · 保护外部 API（避免被限流/封禁）
      · 保护数据库（避免连接打满）
      · 保护下游服务（避免雪崩）

【你会看到】
    ✓ 令牌桶限流
    ✓ 请求被限流的效果
    ✓ 配合 Retry 使用

【通关标准】
    能为不同场景选择合适的限流策略

【前置】A-04              【标签】RateLimiter · 限流 · 令牌桶
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n╔═══════════════════════════════════════╗")
	fmt.Println("║  A-05: RateLimiter 限流保护          ║")
	fmt.Println("╚═══════════════════════════════════════╝\n")

	ctx := context.Background()

	// ========== 示例 1：令牌桶基础 ==========
	fmt.Println("【示例 1】令牌桶限流（5 req/s）")
	
	limiter := async.NewTokenBucketRateLimiter(
		5,              // 每秒 5 个令牌
		time.Second,    // 时间窗口
		10,             // 桶容量（允许突发 10 个）
	)
	
	fmt.Println("  发送 15 个请求:")
	start := time.Now()
	for i := 1; i <= 15; i++ {
		if err := limiter.Wait(ctx); err != nil {
			fmt.Printf("  请求 #%d 被限流\n", i)
			continue
		}
		elapsed := time.Since(start)
		fmt.Printf("    ✓ 请求 #%d 通过 (耗时 %v)\n", i, elapsed)
	}
	fmt.Printf("  总耗时: %v\n", time.Since(start))
	fmt.Println()

	// ========== 示例 2：保护 API 调用 ==========
	fmt.Println("【示例 2】保护外部 API（防止被封禁）")
	
	// 限制：每秒最多 3 个请求
	apiLimiter := async.NewTokenBucketRateLimiter(3, time.Second, 5)
	
	users := []string{"Alice", "Bob", "Charlie", "David", "Eve", "Frank"}
	
	fmt.Printf("  查询 %d 个用户信息（限流 3 req/s）:\n", len(users))
	start = time.Now()
	
	for _, user := range users {
		apiLimiter.Wait(ctx)
		fmt.Printf("    → 查询用户: %s\n", user)
		// 模拟 API 调用
		time.Sleep(10 * time.Millisecond)
	}
	
	fmt.Printf("  完成，总耗时: %v\n", time.Since(start))
	fmt.Println()

	// ========== 示例 3：配合 Map 使用 ==========
	fmt.Println("【示例 3】Map + 限流（批量处理 + 保护下游）")
	
	type Task struct {
		ID int
	}
	
	tasks := []Task{{1}, {2}, {3}, {4}, {5}}
	batchLimiter := async.NewTokenBucketRateLimiter(2, time.Second, 5)
	
	fmt.Println("  并发处理 5 个任务（限流 2 req/s）:")
	start = time.Now()
	
	results := async.Map(ctx, tasks, async.IO(), func(ctx context.Context, task Task) (string, error) {
		// 获取令牌
		if err := batchLimiter.Wait(ctx); err != nil {
			return "", err
		}
		
		elapsed := time.Since(start)
		result := fmt.Sprintf("Task #%d done at %v", task.ID, elapsed)
		fmt.Printf("    ✓ %s\n", result)
		return result, nil
	})
	
	fmt.Printf("  完成 %d 个任务，总耗时: %v\n", len(results), time.Since(start))
	fmt.Println()

	// ========== 示例 4：滑动窗口限流 ==========
	fmt.Println("【示例 4】滑动窗口限流（更精确）")
	
	slidingLimiter := async.NewSlidingWindowRateLimiter(
		3,           // 3 个请求
		time.Second, // 每秒
	)
	
	fmt.Println("  连续发送 6 个请求:")
	for i := 1; i <= 6; i++ {
		if err := slidingLimiter.Allow(); err != nil {
			fmt.Printf("    ✗ 请求 #%d 被拒绝\n", i)
			time.Sleep(400 * time.Millisecond)  // 等待一会儿
			continue
		}
		fmt.Printf("    ✓ 请求 #%d 通过\n", i)
		time.Sleep(200 * time.Millisecond)
	}
	fmt.Println()

	// ========== 示例 5：自适应限流 ==========
	fmt.Println("【示例 5】自适应限流（根据延迟动态调整）")
	
	fmt.Println("  场景：下游服务延迟增加时自动降低请求速率")
	fmt.Println("  原理：")
	fmt.Println("    • 延迟 < 100ms → 允许更多请求")
	fmt.Println("    • 延迟 > 500ms → 自动限流")
	fmt.Println("    • 延迟 > 1s    → 激进限流")
	fmt.Println()

	adaptiveLimiter := async.NewAdaptiveRateLimiter().
		WithInitialLimit(10).
		WithMinLimit(1).
		WithMaxLimit(100).
		WithTargetLatency(100 * time.Millisecond).
		Build()
	
	fmt.Println("  模拟请求（延迟逐渐增加）:")
	for i := 1; i <= 5; i++ {
		latency := time.Duration(i*100) * time.Millisecond
		
		if err := adaptiveLimiter.Wait(ctx); err != nil {
			fmt.Printf("    ✗ 请求 #%d 被限流\n", i)
			continue
		}
		
		// 模拟请求
		time.Sleep(latency)
		
		// 上报延迟，自动调整限流
		adaptiveLimiter.ObserveLatency(latency)
		
		currentLimit := adaptiveLimiter.CurrentLimit()
		fmt.Printf("    ✓ 请求 #%d 完成 (延迟 %v, 当前限制 %d req/s)\n", 
			i, latency, currentLimit)
	}
	fmt.Println()

	// ========== 关键总结 ==========
	fmt.Println("【关键总结】")
	fmt.Println("  • 令牌桶：最常用，平滑 + 突发")
	fmt.Println("  • 滑动窗口：精确限流")
	fmt.Println("  • 自适应：智能调整，适合不稳定的下游")
	fmt.Println()
	fmt.Println("【应用场景】")
	fmt.Println("  ✓ 保护外部 API（避免被封禁）")
	fmt.Println("  ✓ 保护数据库（避免连接打满）")
	fmt.Println("  ✓ 保护下游服务（避免雪崩）")
	fmt.Println("  ✓ 流量整形（平滑突发流量）")
	fmt.Println()
	fmt.Println("【下一关】06_pipeline.go - 数据管道")
}
