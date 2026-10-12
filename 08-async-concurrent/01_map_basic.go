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
A-01：async.Map 并发数据处理
═══════════════════════════════════════════════════════════════════════

【核心概念】
    async.Map() 对切片元素并发执行函数，一行代码搞定并发处理。
    
    自动管理：
      · goroutine 数量（通过 async.IO() 或 async.CPU()）
      · 错误收集
      · 结果聚合
    
    对比手写：50+ 行代码 → 1 行搞定！

【你会看到】
    ✓ 并发处理 10 个数字
    ✓ 自动收集结果
    ✓ 比串行快 N 倍

【通关标准】
    能用 async.Map 并发处理数据

【前置】L2-12              【标签】Map · 并发 · 数据处理
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n╔═══════════════════════════════════════╗")
	fmt.Println("║  A-01: async.Map 并发数据处理        ║")
	fmt.Println("╚═══════════════════════════════════════╝\n")

	ctx := context.Background()

	// ========== 示例 1：最简单的并发 Map ==========
	fmt.Println("【示例 1】并发处理数字（一行代码）")
	
	numbers := []int{1, 2, 3, 4, 5}
	
	start := time.Now()
	results := async.Map(ctx, numbers, async.IO(), func(ctx context.Context, n int) (int, error) {
		time.Sleep(100 * time.Millisecond)  // 模拟耗时操作
		return n * 2, nil
	})
	
	fmt.Printf("  输入: %v\n", numbers)
	fmt.Print("  输出: [")
	for i, r := range results {
		if i > 0 {
			fmt.Print(", ")
		}
		if r.Err != nil {
			fmt.Printf("error: %v", r.Err)
		} else {
			fmt.Printf("%d", r.Value)
		}
	}
	fmt.Printf("]\n")
	fmt.Printf("  并发耗时：%v\n", time.Since(start))
	fmt.Println("  ⚡ 比串行快约 5 倍！\n")

	// ========== 示例 2：对比串行执行 ==========
	fmt.Println("【示例 2】对比：串行执行")
	
	start = time.Now()
	var serialResults []int
	for _, n := range numbers {
		time.Sleep(100 * time.Millisecond)
		serialResults = append(serialResults, n*2)
	}
	fmt.Printf("  串行耗时：%v\n", time.Since(start))
	fmt.Println("  ❌ 慢 5 倍！\n")

	// ========== 示例 3：处理字符串 ==========
	fmt.Println("【示例 3】处理不同类型（泛型支持）")
	
	words := []string{"hello", "world", "async", "golang"}
	wordResults := async.Map(ctx, words, async.CPU(), func(ctx context.Context, word string) (string, error) {
		return fmt.Sprintf("%s!", word), nil
	})
	
	fmt.Print("  结果: ")
	for i, r := range wordResults {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Printf("%s", r.Value)
	}
	fmt.Println("\n")

	// ========== 示例 4：处理结构体 ==========
	fmt.Println("【示例 4】处理结构体数据")
	
	type Order struct {
		ID     int
		Amount float64
	}
	
	orders := []Order{
		{ID: 1, Amount: 100},
		{ID: 2, Amount: 200},
		{ID: 3, Amount: 300},
	}
	
	orderResults := async.Map(ctx, orders, async.IO(), func(ctx context.Context, order Order) (float64, error) {
		// 计算手续费 (1%)
		fee := order.Amount * 0.01
		return fee, nil
	})
	
	fmt.Println("  订单手续费:")
	for i, r := range orderResults {
		fmt.Printf("    订单 #%d: %.2f\n", orders[i].ID, r.Value)
	}
	fmt.Println()

	// ========== 示例 5：错误处理 ==========
	fmt.Println("【示例 5】错误处理")
	
	data := []int{1, 2, 0, 4, 5}  // 0 会导致除零错误
	divResults := async.Map(ctx, data, async.CPU(), func(ctx context.Context, n int) (int, error) {
		if n == 0 {
			return 0, fmt.Errorf("除数不能为 0")
		}
		return 100 / n, nil
	})
	
	fmt.Println("  结果:")
	for i, r := range divResults {
		if r.Err != nil {
			fmt.Printf("    [%d] ❌ 错误: %v\n", data[i], r.Err)
		} else {
			fmt.Printf("    [%d] ✓ 结果: %d\n", data[i], r.Value)
		}
	}
	fmt.Println()

	// ========== 关键总结 ==========
	fmt.Println("【关键总结】")
	fmt.Println("  • async.Map() 一行代码搞定并发处理")
	fmt.Println("  • async.IO() 适合 IO 密集型（网络、文件）")
	fmt.Println("  • async.CPU() 适合 CPU 密集型（计算、加密）")
	fmt.Println("  • 泛型支持：可处理任何类型")
	fmt.Println("  • 自动错误收集：每个元素独立")
	fmt.Println()
	fmt.Println("【下一关】02_foreach.go - ForEach 遍历处理")
}
