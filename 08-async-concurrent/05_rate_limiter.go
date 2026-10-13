//go:build level

package main

import (
	"context"
	"fmt"
	"strings"
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
    3. RateLimiter：通用限流器

    适用场景：
      · 保护外部 API（避免被限流/封禁）
      · 保护数据库（避免连接打满）
      · 保护下游服务（避免雪崩）

【你会看到】
    ✓ 令牌桶限流
    ✓ 请求被限流的效果
    ✓ 多种限流算法对比

【通关标准】
    能为不同场景选择合适的限流策略

【前置】A-04              【标签】RateLimiter · 限流 · 令牌桶
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🚀 A-05: RateLimiter 限流保护")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	ctx := context.Background()

	// ========== 示例 1：令牌桶基础 ==========
	fmt.Println("📝 示例 1：令牌桶限流（5 req/s）")
	fmt.Println(strings.Repeat("-", 60))

	limiter := async.Ratelimit(ctx).
		TokenBucket().
		Rate(10).
		Capacity(20).
		Build()

	fmt.Println("  ⚙️  配置：每秒 10 个令牌，桶容量 20")
	fmt.Println("  🚦 发送 15 个请求:")
	start := time.Now()
	successCount := 0
	for i := 1; i <= 15; i++ {
		if limiter.Allow() {
			elapsed := time.Since(start)
			fmt.Printf("     ✅ 请求 #%d 通过 (⏱️  耗时 %v)\n", i, elapsed.Round(time.Millisecond))
			successCount++
		} else {
			fmt.Printf("     ❌ 请求 #%d 被限流\n", i)
		}
		time.Sleep(50 * time.Millisecond) // 模拟请求间隔
	}
	fmt.Printf("  📊 成功: %d/15，⏱️  总耗时: %v\n", successCount, time.Since(start).Round(time.Millisecond))
	fmt.Println()

	// ========== 示例 2：RateLimiter 保护 API ==========
	fmt.Println("🎯 示例 2：RateLimiter 保护外部 API")
	fmt.Println(strings.Repeat("-", 60))

	apiLimiter := async.Ratelimit(ctx).
		RateLimiter().
		Rate(5).
		Per(time.Second).
		Build()

	users := []string{"Alice", "Bob", "Charlie", "David", "Eve", "Frank"}

	fmt.Printf("  📋 查询 %d 个用户信息（限流 5 req/s）:\n", len(users))
	start = time.Now()

	for _, user := range users {
		// 等待直到可以执行
		for !apiLimiter.Allow() {
			time.Sleep(50 * time.Millisecond)
		}
		fmt.Printf("     ➡️  查询用户: %s\n", user)
		// 模拟 API 调用
		time.Sleep(10 * time.Millisecond)
	}

	fmt.Printf("  ✅ 完成，⏱️  总耗时: %v\n", time.Since(start).Round(time.Millisecond))
	fmt.Println()

	// ========== 示例 3：滑动窗口限流 ==========
	fmt.Println("🔍 示例 3：滑动窗口限流（更精确）")
	fmt.Println(strings.Repeat("-", 60))

	slidingLimiter := async.Ratelimit(ctx).
		SlidingWindow().
		Limit(5).
		Window(time.Second).
		Build()

	fmt.Println("  🚦 连续发送 10 个请求:")
	successCount = 0
	for i := 1; i <= 10; i++ {
		if slidingLimiter.Allow() {
			fmt.Printf("     ✅ 请求 #%d 通过\n", i)
			successCount++
		} else {
			fmt.Printf("     ❌ 请求 #%d 被拒绝（超过窗口限制）\n", i)
		}
		time.Sleep(150 * time.Millisecond)
	}
	fmt.Printf("  📊 通过: %d/10\n", successCount)
	fmt.Println()

	// ========== 示例 4：配合 Map 使用 ==========
	fmt.Println("🔢 示例 4：Map + 限流（批量处理 + 保护下游）")
	fmt.Println(strings.Repeat("-", 60))

	type Task struct {
		ID int
	}

	tasks := []Task{{1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}}
	batchLimiter := async.Ratelimit(ctx).
		TokenBucket().
		Rate(3).
		Capacity(10).
		Build()

	fmt.Println("  🎭 并发处理 8 个任务（限流 3 req/s）:")
	start = time.Now()

	result := async.Slice(ctx, tasks).
		Worker(async.IO()).
		Map(func(ctx context.Context, task Task) (string, error) {
			// 等待限流器允许
			for !batchLimiter.Allow() {
				time.Sleep(100 * time.Millisecond)
			}

			elapsed := time.Since(start)
			resultStr := fmt.Sprintf("Task #%d done at %v", task.ID, elapsed.Round(time.Millisecond))
			fmt.Printf("     ✅ %s\n", resultStr)
			return resultStr, nil
		})

	fmt.Printf("  📊 完成 %d 个任务，⏱️  总耗时: %v\n", len(result.Values()), time.Since(start).Round(time.Millisecond))
	fmt.Println()

	// ========== 示例 5：限流策略对比 ==========
	fmt.Println("📊 示例 5：限流策略对比")
	fmt.Println(strings.Repeat("-", 60))

	fmt.Println("  🔹 TokenBucket（令牌桶）:")
	fmt.Println("     ✅ 允许突发流量（桶内有余额）")
	fmt.Println("     ✅ 平滑限流，适合大多数场景")
	fmt.Println("     📝 用法: .TokenBucket().Rate(100).Capacity(200)")
	fmt.Println()

	fmt.Println("  🔹 SlidingWindow（滑动窗口）:")
	fmt.Println("     ✅ 精确限流，严格控制窗口内请求数")
	fmt.Println("     ✅ 防止窗口边界突发")
	fmt.Println("     📝 用法: .SlidingWindow().Limit(100).Window(time.Second)")
	fmt.Println()

	fmt.Println("  🔹 RateLimiter（通用限流器）:")
	fmt.Println("     ✅ 简单易用")
	fmt.Println("     ✅ 适合基础限流需求")
	fmt.Println("     📝 用法: .RateLimiter().Rate(100).Per(time.Second)")
	fmt.Println()

	// ========== 关键总结 ==========
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ TokenBucket：最常用，平滑 + 突发")
	fmt.Println("  ✅ SlidingWindow：精确限流，严格控制")
	fmt.Println("  ✅ RateLimiter：简单通用限流")
	fmt.Println("  ✅ 所有限流器使用 .Allow() 方法检查是否允许")
	fmt.Println()
	fmt.Println("  📋 应用场景：")
	fmt.Println("     🌐 保护外部 API（避免被封禁）")
	fmt.Println("     🗄️  保护数据库（避免连接打满）")
	fmt.Println("     🔧 保护下游服务（避免雪崩）")
	fmt.Println("     📊 流量整形（平滑突发流量）")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("➡️  下一关：06_pipeline.go - 数据管道")
	fmt.Println()
}
