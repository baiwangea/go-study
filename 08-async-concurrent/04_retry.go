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
A-04：Retry 重试机制
═══════════════════════════════════════════════════════════════════════

【核心概念】
    async.Retry() 提供智能重试机制，适合不稳定的外部调用。

    重试策略：
      · 指数退避（Exponential Backoff）：1s → 2s → 4s → 8s
      · 线性退避（Linear Backoff）：1s → 2s → 3s → 4s
      · 固定间隔

    适用场景：
      · API 调用（网络抖动）
      · 数据库连接（临时不可用）
      · 消息队列（消费失败）
      · 任何需要容错的操作

【你会看到】
    ✓ 指数退避重试
    ✓ 最大重试次数
    ✓ 快速失败 vs 重试成功

【通关标准】
    能为不稳定的外部调用配置合适的重试策略

【前置】A-03              【标签】Retry · 指数退避 · 容错
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🚀 A-04: Retry 重试机制")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	ctx := context.Background()

	// ========== 示例 1：基本重试 ==========
	fmt.Println("📝 示例 1：基本重试（3 次）")
	fmt.Println(strings.Repeat("-", 60))

	attempt := 0
	result, err := async.Retry[string](ctx).
		MaxRetries(3).
		Execute(func(ctx context.Context) (string, error) {
			attempt++
			fmt.Printf("     🔁 尝试 #%d...\n", attempt)
			if attempt < 3 {
				return "", fmt.Errorf("临时失败")
			}
			return "成功!", nil
		})

	if err != nil {
		fmt.Printf("  ❌ 失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ 结果: %s\n", result)
	}
	fmt.Println()

	// ========== 示例 2：指数退避 ==========
	fmt.Println("🔁 示例 2：指数退避重试")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("  ⚙️  退避策略：100ms → 200ms → 400ms")
	fmt.Println()

	attempt = 0
	start := time.Now()

	async.Retry[any](ctx).
		MaxRetries(4).
		Exponential().
		Backoff(100*time.Millisecond, 5*time.Second).
		Execute(func(ctx context.Context) (any, error) {
			attempt++
			now := time.Since(start)
			fmt.Printf("     🔁 尝试 #%d (⏱️  耗时 %v)\n", attempt, now.Round(time.Millisecond))
			if attempt < 4 {
				return nil, fmt.Errorf("失败")
			}
			return nil, nil
		})

	fmt.Printf("  ⏱️  总耗时: %v\n", time.Since(start).Round(time.Millisecond))
	fmt.Println()

	// ========== 示例 3：实际场景 - API 调用 ==========
	fmt.Println("🎯 示例 3：实际场景 - 不稳定的 API 调用")
	fmt.Println(strings.Repeat("-", 60))

	callCount := 0
	result, err = async.Retry[string](ctx).
		MaxRetries(5).
		Exponential().
		Backoff(50*time.Millisecond, 2*time.Second).
		Execute(func(ctx context.Context) (string, error) {
			callCount++
			fmt.Printf("     ➡️  调用 API (第 %d 次)\n", callCount)

			// 模拟 API：前 2 次失败，第 3 次成功
			if callCount < 3 {
				return "", fmt.Errorf("503 Service Unavailable")
			}
			return `{"status":"ok","data":"user_info"}`, nil
		})

	if err != nil {
		fmt.Printf("  ❌ API 调用失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ API 响应: %s\n", result)
	}
	fmt.Println()

	// ========== 示例 4：快速失败 ==========
	fmt.Println("⚠️  示例 4：快速失败（超过最大重试次数）")
	fmt.Println(strings.Repeat("-", 60))

	attempt = 0
	_, err = async.Retry[any](ctx).
		MaxRetries(2).
		Execute(func(ctx context.Context) (any, error) {
			attempt++
			fmt.Printf("     🔁 尝试 #%d\n", attempt)
			return nil, fmt.Errorf("永久失败")
		})

	if err != nil {
		fmt.Printf("  ❌ 达到最大重试次数，放弃: %v\n", err)
	}
	fmt.Println()

	// ========== 示例 5：重试策略对比 ==========
	fmt.Println("🔍 示例 5：重试策略对比")
	fmt.Println(strings.Repeat("-", 60))

	fmt.Println("  🔁 指数退避（适合大多数场景）:")
	fmt.Println("     100ms → 200ms → 400ms → 800ms")
	fmt.Println("     ✅ 优点: 快速恢复 + 避免雪崩")
	fmt.Println()

	fmt.Println("  🔁 线性退避（适合已知恢复时间）:")
	fmt.Println("     1s → 2s → 3s → 4s")
	fmt.Println("     ✅ 优点: 可预测，适合定期任务")
	fmt.Println()

	fmt.Println("  🔁 固定间隔（适合快速重试）:")
	fmt.Println("     500ms → 500ms → 500ms")
	fmt.Println("     ✅ 优点: 简单，适合瞬时故障")
	fmt.Println()

	// ========== 关键总结 ==========
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ Retry 适合处理临时性故障")
	fmt.Println("  ✅ 指数退避是最常用的策略")
	fmt.Println("  ✅ 设置合理的最大重试次数（避免无限等待）")
	fmt.Println("  ✅ 永久性错误应该快速失败，不应重试")
	fmt.Println()
	fmt.Println("  📋 应用场景：")
	fmt.Println("     🌐 API 调用（网络抖动）")
	fmt.Println("     🗄️  数据库连接（临时不可用）")
	fmt.Println("     📨 消息队列（消费失败）")
	fmt.Println("     📁 文件操作（资源竞争）")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("➡️  下一关：05_rate_limiter.go - 限流保护")
	fmt.Println()
}
