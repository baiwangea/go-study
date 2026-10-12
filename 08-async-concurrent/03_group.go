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
A-03：async.Group 任务组
═══════════════════════════════════════════════════════════════════════

【核心概念】
    async.Group 并发执行一组任务，任一失败则全部失败。
    
    类似 golang.org/x/sync/errgroup，但零外部依赖。
    
    适用场景：
      · 一组相关任务必须全部成功
      · 任一失败需要立即终止其他任务
      · 并发执行多个独立操作

【你会看到】
    ✓ 多个任务并发执行
    ✓ 全部成功或任一失败
    ✓ 自动取消机制

【通关标准】
    能用 Group 组织相关任务的并发执行

【前置】A-01              【标签】Group · errgroup · 任务组
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n╔═══════════════════════════════════════╗")
	fmt.Println("║  A-03: async.Group 任务组            ║")
	fmt.Println("╚═══════════════════════════════════════╝\n")

	ctx := context.Background()

	// ========== 示例 1：全部成功 ==========
	fmt.Println("【示例 1】全部任务成功")
	
	start := time.Now()
	err := async.NewGroup[any](ctx).
		Add(func(ctx context.Context) (any, error) {
			time.Sleep(100 * time.Millisecond)
			fmt.Println("    ✓ 任务 1 完成")
			return nil, nil
		}).
		Add(func(ctx context.Context) (any, error) {
			time.Sleep(150 * time.Millisecond)
			fmt.Println("    ✓ 任务 2 完成")
			return nil, nil
		}).
		Add(func(ctx context.Context) (any, error) {
			time.Sleep(120 * time.Millisecond)
			fmt.Println("    ✓ 任务 3 完成")
			return nil, nil
		}).
		Wait()
	
	if err != nil {
		fmt.Printf("  ❌ 失败: %v\n", err)
	} else {
		fmt.Printf("  ✓ 全部完成，耗时：%v\n", time.Since(start))
	}
	fmt.Println()

	// ========== 示例 2：任一失败 ==========
	fmt.Println("【示例 2】任一任务失败")
	
	err = async.NewGroup[any](ctx).
		Add(func(ctx context.Context) (any, error) {
			time.Sleep(50 * time.Millisecond)
			fmt.Println("    ✓ 任务 A 完成")
			return nil, nil
		}).
		Add(func(ctx context.Context) (any, error) {
			time.Sleep(100 * time.Millisecond)
			fmt.Println("    ❌ 任务 B 失败")
			return nil, fmt.Errorf("任务 B 出错")
		}).
		Add(func(ctx context.Context) (any, error) {
			time.Sleep(200 * time.Millisecond)
			fmt.Println("    任务 C（可能被取消）")
			return nil, nil
		}).
		Wait()
	
	if err != nil {
		fmt.Printf("  ❌ Group 失败: %v\n", err)
	}
	fmt.Println()

	// ========== 示例 3：实际场景 - 多数据源查询 ==========
	fmt.Println("【示例 3】实际场景：同时查询多个数据源")
	
	type UserData struct {
		Name string
	}
	
	// 并发查询 3 个数据源
	results, err := async.NewGroup[*UserData](ctx).
		Add(func(ctx context.Context) (*UserData, error) {
			time.Sleep(50 * time.Millisecond)
			fmt.Println("    ✓ 数据库查询完成")
			return &UserData{Name: "Alice"}, nil
		}).
		Add(func(ctx context.Context) (*UserData, error) {
			time.Sleep(80 * time.Millisecond)
			fmt.Println("    ✓ 缓存查询完成")
			return &UserData{Name: "Bob"}, nil
		}).
		Add(func(ctx context.Context) (*UserData, error) {
			time.Sleep(60 * time.Millisecond)
			fmt.Println("    ✓ API 查询完成")
			return &UserData{Name: "Charlie"}, nil
		}).
		WaitAndGetResults()
	
	if err != nil {
		fmt.Printf("  ❌ 查询失败: %v\n", err)
	} else {
		fmt.Println("  ✓ 所有数据源查询成功:")
		for i, r := range results {
			if r.Value != nil {
				fmt.Printf("    数据源 %d: %s\n", i+1, r.Value.Name)
			}
		}
	}
	fmt.Println()

	// ========== 示例 4：对比 errgroup ==========
	fmt.Println("【示例 4】对比 golang.org/x/sync/errgroup")
	fmt.Println("  async.Group 的优势：")
	fmt.Println("    ✓ 零外部依赖")
	fmt.Println("    ✓ 链式 API 更优雅")
	fmt.Println("    ✓ 支持获取结果")
	fmt.Println("    ✓ 泛型支持")
	fmt.Println()

	// ========== 关键总结 ==========
	fmt.Println("【关键总结】")
	fmt.Println("  • Group 适合一组相关任务必须全部成功的场景")
	fmt.Println("  • 任一失败会导致其他任务被取消")
	fmt.Println("  • Wait() 等待全部完成并返回错误")
	fmt.Println("  • WaitAndGetResults() 可以获取每个任务的结果")
	fmt.Println()
	fmt.Println("【下一关】04_retry.go - 重试机制")
}
