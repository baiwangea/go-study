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
A-01：async.Map 并发数据处理
═══════════════════════════════════════════════════════════════════════

【核心概念】
    async.Slice(ctx, items).Worker(async.IO()).Map(...) 对切片元素并发执行函数。

    自动管理：
      · goroutine 数量（Worker(async.IO()) 或 DefaultWorker()）
      · 错误收集
      · 结果聚合

    对比手写：50+ 行代码 → 链式几行搞定。

【你会看到】
    ✓ 并发处理 10 个数字
    ✓ 自动收集结果
    ✓ 比串行快 N 倍

【通关标准】
    能用 async.Slice(...).Map 并发处理数据

【前置】L2-12              【标签】Map · 并发 · 数据处理
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🚀 A-01: async.Map 并发数据处理")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	ctx := context.Background()

	// ========== 示例 1：最简单的并发 Map ==========
	fmt.Println("📝 示例 1：并发处理数字（一行代码）")
	fmt.Println(strings.Repeat("-", 60))

	numbers := []int{1, 2, 3, 4, 5}

	start := time.Now()
	result := async.Slice(ctx, numbers).
		DefaultWorker(). // 使用默认并发度（IO密集型）
		Map(func(ctx context.Context, n int) (int, error) {
			time.Sleep(100 * time.Millisecond) // 模拟耗时操作
			return n * 2, nil
		})

	fmt.Printf("  📥 输入: %v\n", numbers)
	fmt.Print("  📤 输出: [")
	values := result.Values() // 获取所有成功的值
	for i, v := range values {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Printf("%d", v)
	}
	fmt.Printf("]\n")
	fmt.Printf("  ⏱️  并发耗时：%v\n", time.Since(start))
	fmt.Println("  ⚡ 速度提升：约 5 倍！\n")

	// ========== 示例 2：对比串行执行 ==========
	fmt.Println("🐌 示例 2：对比 - 串行执行")
	fmt.Println(strings.Repeat("-", 60))

	start = time.Now()
	var serialResults []int
	for _, n := range numbers {
		time.Sleep(100 * time.Millisecond)
		serialResults = append(serialResults, n*2)
	}
	fmt.Printf("  ⏱️  串行耗时：%v\n", time.Since(start))
	fmt.Println("  ❌ 结论：慢 5 倍！\n")

	// ========== 示例 3：处理字符串 ==========
	fmt.Println("🔤 示例 3：处理不同类型（泛型支持）")
	fmt.Println(strings.Repeat("-", 60))

	words := []string{"hello", "world", "async", "golang"}
	wordResult := async.Slice(ctx, words). // string → string，不需要 SliceWith
						Worker(async.CPU()). // 使用CPU密集型并发度
						Map(func(ctx context.Context, word string) (string, error) {
			return fmt.Sprintf("%s!", word), nil
		})

	fmt.Print("  📤 结果: ")
	wordValues := wordResult.Values()
	for i, v := range wordValues {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Printf("%s", v)
	}
	fmt.Println("\n")

	// ========== 示例 4：处理结构体 ==========
	fmt.Println("🏗️  示例 4：处理结构体数据")
	fmt.Println(strings.Repeat("-", 60))

	type Order struct {
		ID     int
		Amount float64
	}

	orders := []Order{
		{ID: 1, Amount: 100},
		{ID: 2, Amount: 200},
		{ID: 3, Amount: 300},
	}

	orderResult := async.SliceWith[float64](ctx, orders).
		DefaultWorker().
		Map(func(ctx context.Context, order Order) (float64, error) {
			// 计算手续费 (1%)
			fee := order.Amount * 0.01
			return fee, nil
		})

	fmt.Println("  💰 订单手续费:")
	for i, fee := range orderResult.Values() {
		fmt.Printf("     订单 #%d → ¥%.2f\n", orders[i].ID, fee)
	}
	fmt.Println()

	// ========== 示例 5：错误处理 ==========
	fmt.Println("⚠️  示例 5：错误处理")
	fmt.Println(strings.Repeat("-", 60))

	data := []int{1, 2, 0, 4, 5} // 0 会导致除零错误
	divResult := async.Slice(ctx, data).
		DefaultWorker().
		Map(func(ctx context.Context, n int) (int, error) {
			if n == 0 {
				return 0, fmt.Errorf("除数不能为 0")
			}
			return 100 / n, nil
		})

	fmt.Println("  📊 结果:")
	results := divResult.Results() // 获取所有结果（含错误）
	for i, r := range results {
		if r.Err != nil {
			fmt.Printf("     [%d] ❌ 错误: %v\n", data[i], r.Err)
		} else {
			fmt.Printf("     [%d] ✅ 结果: %d\n", data[i], r.Value)
		}
	}
	fmt.Println()

	// ========== 关键总结 ==========
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ async.Slice(ctx, items).Map() 并发处理数据")
	fmt.Println("  ✅ DefaultWorker() 使用默认并发度（IO密集型）")
	fmt.Println("  ✅ Worker(n) 自定义并发度，async.IO()/CPU() 获取推荐值")
	fmt.Println("  ✅ .Values() 获取所有成功结果，.Results() 包含错误")
	fmt.Println("  ✅ 泛型支持：可处理任何类型")
	fmt.Println("  ✅ 自动错误收集：每个元素独立")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("➡️  下一关：02_foreach.go - ForEach 遍历处理")
	fmt.Println()
}
