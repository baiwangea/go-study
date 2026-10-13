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
A-06：Pipeline 数据管道
═══════════════════════════════════════════════════════════════════════

【核心概念】
    Pipeline 多阶段串行数据处理，适合 ETL 场景。

    典型流程：
      Extract（提取）→ Transform（转换）→ Load（加载）

    特点：
      · 每个阶段独立处理
      · 支持并发（每阶段内部并发）
      · 自动背压控制
      · AutoScale 自动扩缩容

    适用场景：
      · ETL 数据处理
      · 日志处理管道
      · 消息处理链
      · 实时数据流

【你会看到】
    ✓ 多阶段数据处理
    ✓ 每个阶段的并发控制
    ✓ 自动背压

【通关标准】
    能用 Pipeline 构建多阶段数据处理流程

【前置】A-01              【标签】Pipeline · ETL · 流式处理
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🚀 A-06: Pipeline 数据管道")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	ctx := context.Background()

	// ========== 示例 1：简单的 3 阶段管道 ==========
	fmt.Println("📝 示例 1：简单的 ETL 管道")
	fmt.Println(strings.Repeat("-", 60))

	type RawData struct {
		ID    int
		Value string
	}

	// 原始数据
	rawData := []RawData{
		{ID: 1, Value: "  raw1  "},
		{ID: 2, Value: "  raw2  "},
		{ID: 3, Value: "  raw3  "},
	}

	fmt.Println("  ⛓️  阶段 1: Extract（提取）")
	fmt.Println("  ⛓️  阶段 2: Transform（清洗）")
	fmt.Println("  ⛓️  阶段 3: Load（加载）")
	fmt.Println()

	start := time.Now()

	// 构建管道
	results, err := async.Pipeline(rawData).
		Context(ctx).
		Stage("extract", 2).
		Stage("transform", 2).
		Stage("load", 1).
		Execute(func(ctx context.Context, stage string, data RawData) (RawData, error) {
			switch stage {
			case "extract":
				fmt.Printf("     [Extract] ID=%d\n", data.ID)
				return data, nil
			case "transform":
				fmt.Printf("     [Transform] 清洗 ID=%d\n", data.ID)
				time.Sleep(50 * time.Millisecond) // 模拟处理时间
				data.Value = fmt.Sprintf("cleaned_%s", strings.TrimSpace(data.Value))
				return data, nil
			case "load":
				fmt.Printf("     [Load] 加载 ID=%d\n", data.ID)
				time.Sleep(30 * time.Millisecond)
				return data, nil
			}
			return data, nil
		})

	if err != nil {
		fmt.Printf("  ❌ 管道失败: %v\n", err)
	} else {
		fmt.Println("\n  📊 结果:")
		for _, r := range results {
			if r.Ok() {
				fmt.Printf("     ✅ ID=%d Value=%s\n", r.Value.ID, r.Value.Value)
			} else {
				fmt.Printf("     ❌ ID=%d 失败: %v\n", r.Value.ID, r.Err)
			}
		}
	}

	fmt.Printf("\n  ⏱️  总耗时: %v\n", time.Since(start).Round(time.Millisecond))
	fmt.Println()

	// ========== 示例 2：带并发的管道 ==========
	fmt.Println("🔢 示例 2：并发处理（每阶段内部并发）")
	fmt.Println(strings.Repeat("-", 60))

	data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	fmt.Printf("  📋 处理 %d 条数据，每阶段并发度 3\n", len(data))
	start = time.Now()

	results2, err := async.Pipeline(data).
		Context(ctx).
		Stage("stage1", 3). // 并发度 3
		Stage("stage2", 3).
		Execute(func(ctx context.Context, stage string, n int) (int, error) {
			time.Sleep(50 * time.Millisecond)
			switch stage {
			case "stage1":
				return n * 2, nil
			case "stage2":
				return n + 10, nil
			}
			return n, nil
		})

	if err != nil {
		fmt.Printf("  ❌ 失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ 完成 %d 条数据，⏱️  总耗时: %v\n", len(results2), time.Since(start).Round(time.Millisecond))
		fmt.Println("     ⚡ （如果串行处理需要 1 秒，并发只需约 0.3 秒）")
	}
	fmt.Println()

	// ========== 示例 3：流式管道 ==========
	fmt.Println("⛓️  示例 3：流式管道（边处理边消费）")
	fmt.Println(strings.Repeat("-", 60))

	bigData := make([]int, 20)
	for i := range bigData {
		bigData[i] = i + 1
	}

	fmt.Println("  🔄 流式处理 20 条数据:")
	start = time.Now()

	ch := async.Pipeline(bigData).
		Context(ctx).
		Stage("process", 5).
		ExecutePipe(func(ctx context.Context, stage string, n int) (int, error) {
			time.Sleep(50 * time.Millisecond)
			return n * 2, nil
		}, 1024) // buffer size

	count := 0
	for r := range ch {
		if r.Ok() {
			if count < 5 {
				fmt.Printf("     📤 处理完成: %d → %d\n", count+1, r.Value)
			}
			count++
		} else {
			fmt.Printf("     ❌ 错误: %v\n", r.Err)
		}
	}

	fmt.Printf("  ✅ 流式处理完成 %d 条，⏱️  总耗时: %v\n", count, time.Since(start).Round(time.Millisecond))
	fmt.Println()

	// ========== 示例 4：错误处理 ==========
	fmt.Println("⚠️  示例 4：管道中的错误处理")
	fmt.Println(strings.Repeat("-", 60))

	errorData := []int{1, 2, 0, 4, 5} // 0 会导致错误

	errorResults, _ := async.Pipeline(errorData).
		Context(ctx).
		Stage("validate", 2).
		Stage("process", 2).
		Execute(func(ctx context.Context, stage string, n int) (int, error) {
			switch stage {
			case "validate":
				if n == 0 {
					return 0, fmt.Errorf("无效数据: %d", n)
				}
				return n, nil
			case "process":
				return n * 10, nil
			}
			return n, nil
		})

	fmt.Println("  📊 结果:")
	for i, r := range errorResults {
		if r.Ok() {
			fmt.Printf("     [%d] ✅ 结果: %d\n", errorData[i], r.Value)
		} else {
			fmt.Printf("     [%d] ❌ 错误: %v\n", errorData[i], r.Err)
		}
	}
	fmt.Println()

	// ========== 关键总结 ==========
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ Pipeline 适合多阶段串行数据处理")
	fmt.Println("  ✅ 每阶段独立，可以独立配置并发度")
	fmt.Println("  ✅ Execute() 同步等待全部完成")
	fmt.Println("  ✅ ExecutePipe() 流式处理，边处理边消费")
	fmt.Println("  ✅ 错误会被传递到最终结果，不影响其他数据")
	fmt.Println()
	fmt.Println("  📋 典型应用：")
	fmt.Println("     📊 ETL 数据处理")
	fmt.Println("     📝 日志处理管道")
	fmt.Println("     📨 消息处理链")
	fmt.Println("     🔄 实时数据流处理")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("➡️  下一关：07_bounded_runner.go - 并发控制")
	fmt.Println()
}
