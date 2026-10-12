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
	fmt.Println("\n╔═══════════════════════════════════════╗")
	fmt.Println("║  A-06: Pipeline 数据管道             ║")
	fmt.Println("╚═══════════════════════════════════════╝\n")

	ctx := context.Background()

	// ========== 示例 1：简单的 3 阶段管道 ==========
	fmt.Println("【示例 1】简单的 ETL 管道")
	
	type RawData struct {
		ID    int
		Value string
	}
	
	type CleanedData struct {
		ID    int
		Value string
	}
	
	type Result struct {
		ID     int
		Status string
	}
	
	// 原始数据
	rawData := []RawData{
		{ID: 1, Value: "  raw1  "},
		{ID: 2, Value: "  raw2  "},
		{ID: 3, Value: "  raw3  "},
	}
	
	fmt.Println("  阶段 1: Extract（提取）")
	fmt.Println("  阶段 2: Transform（清洗）")
	fmt.Println("  阶段 3: Load（加载）")
	fmt.Println()
	
	start := time.Now()
	
	// 构建管道
	pipeline := async.NewPipeline[RawData]().
		// 阶段 1: Extract
		AddStage("extract", func(ctx context.Context, data RawData) (CleanedData, error) {
			fmt.Printf("    [Extract] ID=%d\n", data.ID)
			return CleanedData{ID: data.ID, Value: data.Value}, nil
		}).
		// 阶段 2: Transform
		AddStage("transform", func(ctx context.Context, data CleanedData) (CleanedData, error) {
			fmt.Printf("    [Transform] 清洗 ID=%d\n", data.ID)
			// 清洗数据（去空格）
			cleaned := CleanedData{
				ID:    data.ID,
				Value: fmt.Sprintf("cleaned_%s", data.Value),
			}
			time.Sleep(50 * time.Millisecond)  // 模拟处理时间
			return cleaned, nil
		}).
		// 阶段 3: Load
		AddStage("load", func(ctx context.Context, data CleanedData) (Result, error) {
			fmt.Printf("    [Load] 加载 ID=%d\n", data.ID)
			time.Sleep(30 * time.Millisecond)
			return Result{ID: data.ID, Status: "loaded"}, nil
		})
	
	// 执行管道
	results := pipeline.Run(ctx, rawData)
	
	fmt.Println("\n  结果:")
	for _, r := range results {
		if r.Err != nil {
			fmt.Printf("    ✗ ID=%d 失败: %v\n", r.Value.ID, r.Err)
		} else {
			fmt.Printf("    ✓ ID=%d %s\n", r.Value.ID, r.Value.Status)
		}
	}
	
	fmt.Printf("\n  总耗时: %v\n", time.Since(start))
	fmt.Println()

	// ========== 示例 2：带并发的管道 ==========
	fmt.Println("【示例 2】并发处理（每阶段内部并发）")
	
	data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	
	fmt.Printf("  处理 %d 条数据，每阶段并发度 3\n", len(data))
	start = time.Now()
	
	pipelineWithConcurrency := async.NewPipeline[int]().
		WithConcurrency(3).  // 每阶段并发度 3
		AddStage("stage1", func(ctx context.Context, n int) (int, error) {
			time.Sleep(50 * time.Millisecond)
			return n * 2, nil
		}).
		AddStage("stage2", func(ctx context.Context, n int) (int, error) {
			time.Sleep(50 * time.Millisecond)
			return n + 10, nil
		})
	
	results2 := pipelineWithConcurrency.Run(ctx, data)
	
	fmt.Printf("  完成 %d 条数据，总耗时: %v\n", len(results2), time.Since(start))
	fmt.Println("  （如果串行处理需要 1 秒，并发只需 0.3 秒）")
	fmt.Println()

	// ========== 示例 3：AutoScale 自动扩缩容 ==========
	fmt.Println("【示例 3】AutoScale 自动扩缩容")
	
	fmt.Println("  场景：处理速度跟不上时自动增加并发度")
	fmt.Println("  配置：")
	fmt.Println("    • 初始并发度: 2")
	fmt.Println("    • 最大并发度: 10")
	fmt.Println("    • 队列积压 > 5 时自动扩容")
	fmt.Println()
	
	autoscalePipeline := async.NewPipeline[int]().
		WithAutoScale(
			2,   // 初始并发度
			10,  // 最大并发度
			5,   // 队列积压阈值
		).
		AddStage("process", func(ctx context.Context, n int) (int, error) {
			time.Sleep(100 * time.Millisecond)
			return n * 2, nil
		})
	
	bigData := make([]int, 20)
	for i := range bigData {
		bigData[i] = i + 1
	}
	
	start = time.Now()
	results3 := autoscalePipeline.Run(ctx, bigData)
	
	fmt.Printf("  完成 %d 条数据，总耗时: %v\n", len(results3), time.Since(start))
	fmt.Println("  （AutoScale 自动从 2 扩容到 10）")
	fmt.Println()

	// ========== 示例 4：错误处理 ==========
	fmt.Println("【示例 4】管道中的错误处理")
	
	errorData := []int{1, 2, 0, 4, 5}  // 0 会导致错误
	
	errorPipeline := async.NewPipeline[int]().
		AddStage("validate", func(ctx context.Context, n int) (int, error) {
			if n == 0 {
				return 0, fmt.Errorf("无效数据: %d", n)
			}
			return n, nil
		}).
		AddStage("process", func(ctx context.Context, n int) (int, error) {
			return n * 10, nil
		})
	
	errorResults := errorPipeline.Run(ctx, errorData)
	
	fmt.Println("  结果:")
	for i, r := range errorResults {
		if r.Err != nil {
			fmt.Printf("    [%d] ✗ 错误: %v\n", errorData[i], r.Err)
		} else {
			fmt.Printf("    [%d] ✓ 结果: %d\n", errorData[i], r.Value)
		}
	}
	fmt.Println()

	// ========== 关键总结 ==========
	fmt.Println("【关键总结】")
	fmt.Println("  • Pipeline 适合多阶段串行数据处理")
	fmt.Println("  • 每阶段独立，可以独立配置并发度")
	fmt.Println("  • AutoScale 自动根据负载调整并发度")
	fmt.Println("  • 错误会被传递到最终结果，不影响其他数据")
	fmt.Println()
	fmt.Println("【典型应用】")
	fmt.Println("  ✓ ETL 数据处理")
	fmt.Println("  ✓ 日志处理管道")
	fmt.Println("  ✓ 消息处理链")
	fmt.Println("  ✓ 实时数据流处理")
	fmt.Println()
	fmt.Println("【下一关】07_bounded_runner.go - 并发控制")
}
