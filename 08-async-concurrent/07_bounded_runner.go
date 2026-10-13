//go:build level

package main

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/chichengyu/async"
	"github.com/chichengyu/async/task"
)

/*
═══════════════════════════════════════════════════════════════════════
A-07：BoundedRunner 千万级并发管控
═══════════════════════════════════════════════════════════════════════

【核心概念】
    BoundedRunner 用于管控千万级 goroutine 并发。

    问题场景：
      · 百万级任务提交
      · goroutine 数量失控
      · 内存爆炸
      · 系统崩溃

    解决方案：
      · 全局并发限制
      · 背压控制
      · 优雅降级

    适用场景：
      · 爬虫系统（百万 URL）
      · 批量处理（千万数据）
      · 消息消费（高并发）

【你会看到】
    ✓ 控制最大并发 goroutine 数
    ✓ 背压保护
    ✓ 内存可控

【通关标准】
    能用 BoundedRunner 管控大规模并发

【前置】A-01              【标签】BoundedRunner · 并发管控 · 背压
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🚀 A-07: BoundedRunner 并发管控")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	ctx := context.Background()

	// ========== 示例 1：无限制 vs 有限制 ==========
	fmt.Println("📝 示例 1：对比 - 无限制 vs 有限制")
	fmt.Println(strings.Repeat("-", 60))

	fmt.Println("  📋 场景：处理 100 个任务")
	fmt.Println()

	// 无限制（危险！）
	fmt.Println("  ❌ 无限制方式:")
	var count1 atomic.Int32
	for i := 1; i <= 100; i++ {
		go func() {
			count1.Add(1)
			time.Sleep(10 * time.Millisecond)
		}()
	}
	fmt.Println("     ⚠️  创建了 100 个 goroutine")
	fmt.Println("     ⚠️  问题：如果是 100 万个任务，会创建 100 万个 goroutine!")
	fmt.Println()

	time.Sleep(100 * time.Millisecond)

	// 有限制（安全）
	fmt.Println("  ✅ BoundedRunner 方式:")
	runner := async.NewBoundedRunnerBuilder().
		Max(10). // 最多 10 个并发
		Build()

	var count2 atomic.Int32
	for i := 1; i <= 100; i++ {
		i := i
		task.BoundedGo(runner, ctx, func(ctx context.Context) (any, error) {
			count2.Add(1)
			if i <= 5 {
				fmt.Printf("     🎭 处理任务 #%d\n", i)
			}
			time.Sleep(10 * time.Millisecond)
			return nil, nil
		})
	}

	runner.Wait()
	fmt.Printf("     ✅ 完成 %d 个任务，最多同时 10 个 goroutine\n", count2.Load())
	fmt.Println()

	// ========== 示例 2：大规模任务 ==========
	fmt.Println("🔢 示例 2：模拟大规模任务")
	fmt.Println(strings.Repeat("-", 60))

	taskCount := 10000 // 1 万任务

	fmt.Printf("  📊 处理 %d 个任务，最大并发 100\n", taskCount)

	runner2 := async.NewBoundedRunnerBuilder().
		Max(100).
		Build()

	var processed atomic.Int64
	start := time.Now()

	// 提交任务
	for i := 0; i < taskCount; i++ {
		task.BoundedGo(runner2, ctx, func(ctx context.Context) (any, error) {
			// 模拟任务处理
			processed.Add(1)
			time.Sleep(time.Microsecond) // 极短延迟模拟工作
			return nil, nil
		})
	}

	runner2.Wait()

	fmt.Printf("  ✅ 完成 %d 个任务\n", processed.Load())
	fmt.Printf("  ⏱️  总耗时: %v\n", time.Since(start).Round(time.Millisecond))
	if time.Since(start).Seconds() > 0 {
		fmt.Printf("  📊 吞吐量: %.0f 任务/秒\n", float64(taskCount)/time.Since(start).Seconds())
	}
	fmt.Println()

	// ========== 示例 3：错误处理 ==========
	fmt.Println("⚠️  示例 3：错误处理")
	fmt.Println(strings.Repeat("-", 60))

	runner3 := async.NewBoundedRunnerBuilder().
		Max(3).
		Build()

	tasks := []int{1, 2, 0, 4, 5} // 0 会导致错误

	fmt.Println("  🎭 提交 5 个任务（其中 1 个会失败）:")
	for _, n := range tasks {
		n := n
		task.BoundedGo(runner3, ctx, func(ctx context.Context) (any, error) {
			if n == 0 {
				fmt.Printf("     ❌ 任务 %d 失败\n", n)
				return nil, fmt.Errorf("无效值: %d", n)
			}
			fmt.Printf("     ✅ 任务 %d 成功\n", n)
			return nil, nil
		})
	}

	runner3.Wait()
	fmt.Println("  📊 部分任务失败，但其他任务继续执行")
	fmt.Println()

	// ========== 示例 4：实际应用 - 爬虫 ==========
	fmt.Println("🎯 示例 4：实际应用 - 大规模 URL 爬虫")
	fmt.Println(strings.Repeat("-", 60))

	fmt.Println("  📋 场景：爬取大量 URL")
	fmt.Println("  ⚙️  配置：")
	fmt.Println("     🎭 最大并发：1000")
	fmt.Println("     🛡️  内存可控：是")
	fmt.Println()

	crawler := async.NewBoundedRunnerBuilder().
		Max(1000).
		Build()

	// 模拟 URL 列表
	urls := []string{
		"https://example.com/1",
		"https://example.com/2",
		"https://example.com/3",
		"https://example.com/4",
		"https://example.com/5",
	}

	var crawled atomic.Int32

	fmt.Println("  🕷️  开始爬取（模拟 5 个 URL）:")
	for _, url := range urls {
		url := url
		task.BoundedGo(crawler, ctx, func(ctx context.Context) (any, error) {
			// 模拟爬取
			fmt.Printf("     ➡️  爬取: %s\n", url)
			time.Sleep(50 * time.Millisecond)
			crawled.Add(1)
			return nil, nil
		})
	}

	crawler.Wait()
	fmt.Printf("  ✅ 完成 %d 个 URL\n", crawled.Load())
	fmt.Println()

	// ========== 示例 5：与 Pool 的区别 ==========
	fmt.Println("🔍 示例 5：BoundedRunner vs Pool")
	fmt.Println(strings.Repeat("-", 60))

	fmt.Println("  📊 BoundedRunner:")
	fmt.Println("     ✅ 控制最大并发 goroutine 数量")
	fmt.Println("     ✅ 适合百万/千万级任务")
	fmt.Println("     ✅ 简单的背压控制")
	fmt.Println("     📝 用法: task.BoundedGo(runner, ctx, func)")
	fmt.Println()

	fmt.Println("  📊 Pool:")
	fmt.Println("     ✅ 复用 worker goroutine")
	fmt.Println("     ✅ 支持 AutoScale、优先级等高级特性")
	fmt.Println("     ✅ 更灵活的配置")
	fmt.Println("     📝 用法: pool.Submit(task)")
	fmt.Println()

	// ========== 关键总结 ==========
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ BoundedRunner 控制最大并发 goroutine 数")
	fmt.Println("  ✅ 适合百万/千万级任务处理")
	fmt.Println("  ✅ 使用 task.BoundedGo() 提交任务")
	fmt.Println("  ✅ 使用 runner.Wait() 等待所有任务完成")
	fmt.Println()
	fmt.Println("  📋 典型应用：")
	fmt.Println("     🕷️  爬虫系统（百万 URL）")
	fmt.Println("     📊 批量数据处理（千万数据）")
	fmt.Println("     📨 消息消费（高并发）")
	fmt.Println("     🎯 任何需要大规模并发的场景")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("➡️  下一关：08_real_crawler.go - 实战：完整爬虫系统")
	fmt.Println()
}
