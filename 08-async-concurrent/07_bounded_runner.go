//go:build level

package main

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/chichengyu/async"
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
	fmt.Println("\n╔═══════════════════════════════════════╗")
	fmt.Println("║  A-07: BoundedRunner 并发管控        ║")
	fmt.Println("╚═══════════════════════════════════════╝\n")

	ctx := context.Background()

	// ========== 示例 1：无限制 vs 有限制 ==========
	fmt.Println("【示例 1】对比：无限制 vs 有限制")
	
	fmt.Println("  场景：处理 100 个任务")
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
	fmt.Println("    创建了 100 个 goroutine")
	fmt.Println("    问题：如果是 100 万个任务，会创建 100 万个 goroutine!")
	fmt.Println()
	
	time.Sleep(100 * time.Millisecond)
	
	// 有限制（安全）
	fmt.Println("  ✓ BoundedRunner 方式:")
	runner := async.NewBoundedRunner(10)  // 最多 10 个并发
	
	var count2 atomic.Int32
	for i := 1; i <= 100; i++ {
		i := i
		runner.Submit(func() error {
			count2.Add(1)
			if i <= 5 {
				fmt.Printf("    处理任务 #%d\n", i)
			}
			time.Sleep(10 * time.Millisecond)
			return nil
		})
	}
	
	runner.Wait()
	fmt.Printf("    完成 %d 个任务，最多同时 10 个 goroutine\n", count2.Load())
	fmt.Println()

	// ========== 示例 2：百万级任务 ==========
	fmt.Println("【示例 2】模拟百万级任务")
	
	taskCount := 100000  // 10 万任务（模拟百万）
	
	fmt.Printf("  处理 %d 个任务，最大并发 100\n", taskCount)
	
	runner2 := async.NewBoundedRunner(100).
		WithQueueSize(1000).  // 队列容量
		WithBackpressure(true)  // 开启背压
	
	var processed atomic.Int64
	start := time.Now()
	
	// 提交任务
	for i := 0; i < taskCount; i++ {
		runner2.Submit(func() error {
			// 模拟任务处理
			processed.Add(1)
			return nil
		})
	}
	
	runner2.Wait()
	
	fmt.Printf("  完成 %d 个任务\n", processed.Load())
	fmt.Printf("  总耗时: %v\n", time.Since(start))
	fmt.Printf("  吞吐量: %.0f 任务/秒\n", float64(taskCount)/time.Since(start).Seconds())
	fmt.Println()

	// ========== 示例 3：背压保护 ==========
	fmt.Println("【示例 3】背压保护（避免内存爆炸）")
	
	fmt.Println("  场景：生产速度 > 消费速度")
	fmt.Println("  背压策略：")
	fmt.Println("    • 队列满时，Submit 会阻塞")
	fmt.Println("    • 避免无限堆积")
	fmt.Println("    • 保护系统内存")
	fmt.Println()
	
	runner3 := async.NewBoundedRunner(2).
		WithQueueSize(5).
		WithBackpressure(true)
	
	fmt.Println("  生产者快速提交 10 个任务:")
	for i := 1; i <= 10; i++ {
		i := i
		fmt.Printf("    → 提交任务 #%d\n", i)
		runner3.Submit(func() error {
			fmt.Printf("      ✓ 执行任务 #%d\n", i)
			time.Sleep(500 * time.Millisecond)  // 慢速消费
			return nil
		})
	}
	
	fmt.Println("  等待完成...")
	runner3.Wait()
	fmt.Println("  全部完成（背压保护生效，内存可控）")
	fmt.Println()

	// ========== 示例 4：错误收集 ==========
	fmt.Println("【示例 4】错误收集")
	
	runner4 := async.NewBoundedRunner(3)
	
	tasks := []int{1, 2, 0, 4, 5}  // 0 会导致错误
	
	fmt.Println("  提交 5 个任务（其中 1 个会失败）:")
	for _, n := range tasks {
		n := n
		runner4.Submit(func() error {
			if n == 0 {
				fmt.Printf("    ✗ 任务 %d 失败\n", n)
				return fmt.Errorf("无效值: %d", n)
			}
			fmt.Printf("    ✓ 任务 %d 成功\n", n)
			return nil
		})
	}
	
	errors := runner4.WaitAndGetErrors()
	
	fmt.Printf("  总计 %d 个错误:\n", len(errors))
	for _, err := range errors {
		fmt.Printf("    - %v\n", err)
	}
	fmt.Println()

	// ========== 示例 5：实际应用 - 爬虫 ==========
	fmt.Println("【示例 5】实际应用：百万 URL 爬虫")
	
	fmt.Println("  场景：爬取 100 万个 URL")
	fmt.Println("  配置：")
	fmt.Println("    • 最大并发：1000")
	fmt.Println("    • 队列容量：10000")
	fmt.Println("    • 背压保护：开启")
	fmt.Println()
	
	crawler := async.NewBoundedRunner(1000).
		WithQueueSize(10000).
		WithBackpressure(true)
	
	// 模拟 URL 列表
	urls := []string{
		"https://example.com/1",
		"https://example.com/2",
		"https://example.com/3",
		// ... 100 万个 URL
	}
	
	var crawled atomic.Int32
	
	fmt.Println("  开始爬取（模拟 3 个 URL）:")
	for _, url := range urls {
		url := url
		crawler.Submit(func() error {
			// 模拟爬取
			fmt.Printf("    → 爬取: %s\n", url)
			time.Sleep(100 * time.Millisecond)
			crawled.Add(1)
			return nil
		})
	}
	
	crawler.Wait()
	fmt.Printf("  完成 %d 个 URL\n", crawled.Load())
	fmt.Println()

	// ========== 关键总结 ==========
	fmt.Println("【关键总结】")
	fmt.Println("  • BoundedRunner 控制最大并发 goroutine 数")
	fmt.Println("  • 背压保护避免内存爆炸")
	fmt.Println("  • 适合百万/千万级任务处理")
	fmt.Println("  • 可以收集错误，不中断其他任务")
	fmt.Println()
	fmt.Println("【典型应用】")
	fmt.Println("  ✓ 爬虫系统（百万 URL）")
	fmt.Println("  ✓ 批量数据处理（千万数据）")
	fmt.Println("  ✓ 消息消费（高并发）")
	fmt.Println("  ✓ 任何需要大规模并发的场景")
	fmt.Println()
	fmt.Println("【下一关】08_real_crawler.go - 实战：完整爬虫系统")
}
