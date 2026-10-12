//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 08-async-concurrent && go run 01_map_reduce.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/chichengyu/async"
)

/*
═══════════════════════════════════════════════════════════════════════
A-01：async.Map 并发数据处理
═══════════════════════════════════════════════════════════════════════

【核心概念】
    async 库的核心功能：用简洁的 API 实现并发数据处理。

    async.Map() 对切片元素并发执行函数，自动管理：
      · goroutine 数量（通过 async.IO() 或 async.CPU()）
      · 错误收集
      · 结果聚合
      · 超时控制

    对比手写并发处理（50+ 行代码）：
      · 需要手动创建 goroutine
      · 需要管理 WaitGroup
      · 需要 channel 收集结果
      · 需要处理错误

    async.Map() 只需一行！

【你会看到】
    ✓ 并发处理 10 个数字，每个数字 * 2
    ✓ 使用 async.IO() 控制并发度
    ✓ 自动收集结果和错误
    ✓ 并发执行比串行快 N 倍

【思考题】
    1. async.IO() 和 async.CPU() 有什么区别？什么时候用哪个？
    2. 如果处理函数返回 error，结果会怎样？
    3. 对比手写 goroutine + channel，这个库省略了哪些代码？
    4. 如果数据量是 100 万，该如何优化？

【通关标准】
    能用 async.Map 并发处理数据，
    并说清 IO 密集型和 CPU 密集型的并发度选择

【前置】L2-12              【标签】Map · 并发 · 数据处理 · 泛型
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n╔═══════════════════════════════════════╗")
	fmt.Println("║  A-01: async.Map 并发数据处理        ║")
	fmt.Println("╚═══════════════════════════════════════╝\n")

	ctx := context.Background()

	// ========== 示例 1：最简单的 Map（一行代码！） ==========
	fmt.Println("【示例 1】最简单的并发 Map")

	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	start := time.Now()
	results := async.Map(ctx, numbers, async.IO(), func(ctx context.Context, n int) (int, error) {
		time.Sleep(100 * time.Millisecond) // 模拟 IO 操作
		return n * 2, nil
	})

	fmt.Printf("  输入: %v\n", numbers)
	fmt.Print("  输出: [")
	for i, r := range results {
		if r.Err != nil {
			fmt.Printf("error")
		} else {
			fmt.Printf("%d", r.Value)
		}
		if i < len(results)-1 {
			fmt.Print(", ")
		}
	}
	fmt.Printf("]\n")
	fmt.Printf("  并发耗时：%v\n", time.Since(start))
	fmt.Println("  ⚡ 比串行快约 10 倍！\n")

	// ========== 示例 2：对比串行执行 ==========
	fmt.Println("【示例 2】对比：串行执行同样的任务")

	start = time.Now()
	for i := 1; i <= 20; i++ {
		time.Sleep(10 * time.Millisecond)
	}
	fmt.Printf("  串行执行耗时：%v\n", time.Since(start))
	fmt.Println("  ❌ 慢 10 倍！\n")

	// ========== 示例 3：泛型支持 - 不同类型的任务 ==========
	fmt.Println("【示例 3】泛型支持 - 处理不同类型")

	// 字符串任务池
	stringPool := async.NewPool[string]().
		WithWorkers(3).
		WithHandler(func(task string) error {
			fmt.Printf("    处理字符串任务: %s\n", task)
			return nil
		})

	stringPool.Submit("hello")
	stringPool.Submit("world")
	stringPool.Submit("async")
	stringPool.Wait()

	// 结构体任务池
	type Order struct {
		ID     int
		Amount float64
	}

	orderPool := async.NewPool[Order]().
		WithWorkers(2).
		WithHandler(func(order Order) error {
			fmt.Printf("    处理订单: ID=%d, Amount=%.2f\n", order.ID, order.Amount)
			return nil
		})

	orderPool.Submit(Order{ID: 1, Amount: 100.5})
	orderPool.Submit(Order{ID: 2, Amount: 200.8})
	orderPool.Wait()

	fmt.Println()

	// ========== 示例 4：动态提交任务 ==========
	fmt.Println("【示例 4】动态提交任务（模拟实时任务到达）")

	taskPool := async.NewPool[string]().
		WithWorkers(5).
		WithHandler(func(task string) error {
			time.Sleep(100 * time.Millisecond)
			fmt.Printf("    ✓ 完成任务: %s\n", task)
			return nil
		})

	// 模拟任务陆续到达
	go func() {
		for i := 1; i <= 10; i++ {
			taskPool.Submit(fmt.Sprintf("task-%d", i))
			time.Sleep(20 * time.Millisecond) // 每 20ms 来一个新任务
		}
	}()

	time.Sleep(300 * time.Millisecond) // 等待一些任务完成
	taskPool.Wait()                    // 等待所有剩余任务
	fmt.Println()

	// ========== 关键总结 ==========
	fmt.Println("【关键总结】")
	fmt.Println("  • 5 行代码：NewPool + WithWorkers + WithHandler + Submit + Wait")
	fmt.Println("  • 泛型支持：可以处理任何类型的任务")
	fmt.Println("  • 链式 API：代码优雅易读")
	fmt.Println("  • 自动管理：无需手写 channel、WaitGroup、goroutine")
	fmt.Println()
	fmt.Println("  对比手写协程池，省略了：")
	fmt.Println("    ✓ 创建 jobs channel")
	fmt.Println("    ✓ 创建 results channel")
	fmt.Println("    ✓ 创建 WaitGroup")
	fmt.Println("    ✓ 启动 N 个 goroutine")
	fmt.Println("    ✓ defer wg.Done()")
	fmt.Println("    ✓ close(jobs)")
	fmt.Println("    ✓ wg.Wait()")
	fmt.Println()
	fmt.Println("【下一关】02_pool_advanced.go - 协程池高级配置")
}
