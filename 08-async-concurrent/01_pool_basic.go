//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 08-async-concurrent && go run 01_pool_basic.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/chichengyu/async"
)

/*
═══════════════════════════════════════════════════════════════════════
A-01：5 行代码搭建协程池
═══════════════════════════════════════════════════════════════════════

【核心概念】
    async 库最核心的功能：用链式 API，5 行代码就能搭建生产可用的协程池。
    
    对比手写协程池（50+ 行代码）：
      · 需要手动管理 channel
      · 需要手动管理 WaitGroup
      · 需要处理优雅关闭
      · 容易出现 goroutine 泄漏
    
    async.Pool 的三个核心方法：
      · WithWorkers(n)  - 设置 worker 数量
      · WithHandler(fn) - 设置任务处理函数
      · Submit(task)    - 提交任务
      · Wait()          - 等待所有任务完成

【你会看到】
    ✓ 10 个 worker 并发处理 100 个任务
    ✓ 每个任务耗时 100ms，但总耗时约 1 秒（并发效果）
    ✓ 任务完成顺序不确定（并发特性）
    ✓ 所有任务都能正确完成

【思考题】
    1. 把 worker 数量从 10 改成 2，总耗时会变成多少？
    2. 如果不调用 Wait()，程序会怎样？
    3. Submit 100 个任务，但只有 10 个 worker，任务是如何分配的？
    4. 对比手写协程池，这个库帮你省略了哪些繁琐代码？

【通关标准】
    能用 5 行代码写出并发处理任务的协程池，
    并说清 Worker、Task、Handler 的关系

【前置】L2-12              【标签】Pool · Worker · 链式 API · 泛型
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n╔═══════════════════════════════════════╗")
	fmt.Println("║  A-01: 5 行代码搭建协程池            ║")
	fmt.Println("╚═══════════════════════════════════════╝\n")

	// ========== 示例 1：最简单的协程池（5 行代码！） ==========
	fmt.Println("【示例 1】最简单的协程池（5 行代码）")
	
	var count atomic.Int32
	
	// 创建协程池：10 个 worker
	pool := async.NewPool[int]().
		WithWorkers(10).
		WithHandler(func(task int) error {
			count.Add(1)
			time.Sleep(10 * time.Millisecond)
			fmt.Printf("    Worker 处理任务 #%d\n", task)
			return nil
		})
	
	// 提交 20 个任务
	start := time.Now()
	for i := 1; i <= 20; i++ {
		pool.Submit(i)
	}
	
	// 等待所有任务完成
	pool.Wait()
	
	fmt.Printf("  完成 %d 个任务，总耗时：%v\n", count.Load(), time.Since(start))
	fmt.Println("  ⚡ 10 个 worker 并发处理，比串行快了约 10 倍！\n")

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
			time.Sleep(20 * time.Millisecond)  // 每 20ms 来一个新任务
		}
	}()
	
	time.Sleep(300 * time.Millisecond)  // 等待一些任务完成
	taskPool.Wait()  // 等待所有剩余任务
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
