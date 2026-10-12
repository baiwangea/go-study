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
A-02：async.ForEach 并发遍历
═══════════════════════════════════════════════════════════════════════

【核心概念】
    async.ForEach() 并发遍历切片，适合「只执行不返回」的场景。

    与 Map 的区别：
      · Map: 有返回值，收集结果
      · ForEach: 无返回值，只执行副作用（写文件、发请求、打日志）

    常用场景：
      · 批量发送通知
      · 批量写文件
      · 批量调用 API（不关心返回值）

【你会看到】
    ✓ 并发遍历执行
    ✓ 无返回值收集
    ✓ 错误收集

【通关标准】
    能区分 Map 和 ForEach 的使用场景

【前置】A-01              【标签】ForEach · 副作用 · 遍历
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n╔═══════════════════════════════════════╗")
	fmt.Println("║  A-02: async.ForEach 并发遍历        ║")
	fmt.Println("╚═══════════════════════════════════════╝\n")

	ctx := context.Background()

	// ========== 示例 1：批量发送通知 ==========
	fmt.Println("【示例 1】批量发送通知")

	users := []string{"Alice", "Bob", "Charlie", "David", "Eve"}
	var sentCount atomic.Int32

	start := time.Now()
	result := async.Slice(ctx, users).
		Worker(async.IO()).
		ForEach(func(ctx context.Context, user string) error {
			// 模拟发送通知
			time.Sleep(100 * time.Millisecond)
			sentCount.Add(1)
			fmt.Printf("    ✓ 通知已发送给: %s\n", user)
			return nil
		})

	fmt.Printf("  总计发送：%d 条\n", sentCount.Load())
	fmt.Printf("  耗时：%v\n", time.Since(start))
	if result.Ok() {
		fmt.Printf("  全部成功\n")
	} else {
		fmt.Printf("  有错误: %v\n", result.Error())
	}
	fmt.Println()

	// ========== 示例 2：批量写文件（模拟） ==========
	fmt.Println("【示例 2】批量写文件")

	files := []string{"data1.txt", "data2.txt", "data3.txt"}

	async.Slice(ctx, files).
		Worker(async.CPU()).
		ForEach(func(ctx context.Context, filename string) error {
			// 模拟写文件
			time.Sleep(50 * time.Millisecond)
			fmt.Printf("    ✓ 文件已写入: %s\n", filename)
			return nil
		})
	fmt.Println()

	// ========== 示例 3：错误处理 ==========
	fmt.Println("【示例 3】错误处理（部分失败）")

	tasks := []int{1, 2, 3, 4, 5}

	result = async.Slice(ctx, tasks).
		Worker(async.IO()).
		ForEach(func(ctx context.Context, n int) error {
			if n == 3 {
				return fmt.Errorf("任务 %d 失败", n)
			}
			fmt.Printf("    ✓ 任务 %d 完成\n", n)
			return nil
		})

	if !result.Ok() {
		fmt.Printf("  有错误发生: %v\n", result.Error())
	} else {
		fmt.Printf("  全部完成\n")
	}
	fmt.Println()

	// ========== 示例 4：Map vs ForEach 对比 ==========
	fmt.Println("【示例 4】Map vs ForEach 对比")

	numbers := []int{1, 2, 3}

	// 用 Map（需要返回值）
	fmt.Println("  Map（有返回值）:")
	mapResult := async.Slice(ctx, numbers).
		Worker(async.CPU()).
		Map(func(ctx context.Context, n int) (int, error) {
			return n * 2, nil
		})
	for _, v := range mapResult.Values() {
		fmt.Printf("    结果: %d\n", v)
	}

	// 用 ForEach（只执行副作用）
	fmt.Println("  ForEach（无返回值）:")
	async.Slice(ctx, numbers).
		Worker(async.CPU()).
		ForEach(func(ctx context.Context, n int) error {
			fmt.Printf("    执行副作用: %d\n", n)
			return nil
		})
	fmt.Println()

	// ========== 关键总结 ==========
	fmt.Println("【关键总结】")
	fmt.Println("  • Map: 有返回值，收集结果")
	fmt.Println("  • ForEach: 无返回值，只执行副作用")
	fmt.Println("  • 选择依据：需要结果？→ Map；只需执行？→ ForEach")
	fmt.Println("  • 错误都会被收集，但 ForEach 不中断其他任务")
	fmt.Println()
	fmt.Println("【下一关】03_reduce.go - Reduce 聚合")
}
