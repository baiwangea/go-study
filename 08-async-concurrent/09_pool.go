//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
//
// 跑法：cd 08-async-concurrent && go run 09_pool.go
// API 以 github.com/chichengyu/async v0.2.0 为准：
//   async.Pool[T]().Context(ctx).Worker(n).Run(func(ctx, p *async.Pool[T]) error)
// ─────────────────────────────────────────────────────────────────────────────

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
A-09：协程池 —— 任务比 Map 先到，worker 反复用
═══════════════════════════════════════════════════════════════════════

【核心概念】
    Map / ForEach 是「这一批切片，处理完就散」。
    协程池是「worker 先开好，任务陆续 Submit，最后 Wait」。

    这一关用的是库里真实的链式入口，不是手写 NewPool：

        async.Pool[int]().Context(ctx).Worker(4).Run(func(ctx context.Context, p *async.Pool[int]) error {
            p.Submit(ctx, func(ctx context.Context) (int, error) { ... })
            results := p.Wait()
            return nil
        })

    Run 负责把池建起来、把生命周期收掉。池子只活在 Run 的回调里。

    和 L2-12 手写 worker pool 的差别：不用自己 close(jobs)、不用自己 WaitGroup。
    和 A-07 BoundedRunner 的差别：池是复用固定 worker；BoundedRunner 是限制同时在飞的 goroutine。

【你会看到】
    ✓ 4 个 worker 吃 8 个任务，总耗时接近两批而不是八批
    ✓ Wait 拿到的每条结果带 Value / Err
    ✓ 任务返回 error 不会把整个池打崩，错误在结果里

【思考题】
    1. 行情是「一批已知的 tick」还是「连接不断往里推」？该用 Map 还是 Pool？
    2. Worker(4) 和 rate.Limiter 各限制什么？
    3. 为什么 Submit 的闭包不要直接捕获 for 循环变量（旧 Go）？

【通关标准】
    能写出 Pool().Context().Worker().Run，并说清它和 Map 的分工。

【前置】A-01 · L2-12      【标签】Pool · Submit · Wait · 复用 worker
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🚀 A-09: 协程池")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	ctx := context.Background()

	fmt.Println("📝 示例 1：4 个 worker 吃 8 个任务")
	fmt.Println(strings.Repeat("-", 60))
	start := time.Now()
	err := async.Pool[int]().Context(ctx).Worker(4).Run(func(ctx context.Context, p *async.Pool[int]) error {
		for i := 1; i <= 8; i++ {
			n := i
			if err := p.Submit(ctx, func(ctx context.Context) (int, error) {
				time.Sleep(40 * time.Millisecond)
				return n * 10, nil
			}); err != nil {
				return err
			}
		}
		results := p.Wait()
		fmt.Printf("  📤 收到 %d 条结果，耗时 %v\n", len(results), time.Since(start).Round(time.Millisecond))
		for _, r := range results {
			if !r.Ok() {
				fmt.Printf("     ❌ %v\n", r.Err)
				continue
			}
			fmt.Printf("     ✅ %d\n", r.Value)
		}
		return nil
	})
	if err != nil {
		fmt.Println("  ❌ Run:", err)
	}
	fmt.Println("  ➡️  8×40ms 串行要 320ms；4 个 worker 大约两批，100ms 上下。")

	fmt.Println("\n📌 示例 2：单个任务失败，记在结果里")
	fmt.Println(strings.Repeat("-", 60))
	err = async.Pool[string]().Context(ctx).Worker(2).Run(func(ctx context.Context, p *async.Pool[string]) error {
		_ = p.Submit(ctx, func(ctx context.Context) (string, error) {
			return "ok", nil
		})
		_ = p.Submit(ctx, func(ctx context.Context) (string, error) {
			return "", fmt.Errorf("下游超时")
		})
		for _, r := range p.Wait() {
			if r.Ok() {
				fmt.Printf("  ✅ %s\n", r.Value)
			} else {
				fmt.Printf("  ❌ %v\n", r.Err)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Println("  ❌ Run:", err)
	}
	fmt.Println("  ➡️  Submit 成功只表示进了队列。成败看 Wait 里的 Result.Err。")

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ 一批已知数据用 Slice.Map；任务陆续到达用 Pool.Submit")
	fmt.Println("  ✅ Worker(n) 是复用的工人数，不是每秒次数")
	fmt.Println("  ✅ Run 管生命周期；池子不要拿到 Run 外面用")
	fmt.Println("  ✅ 任务错误在 Result 里，不必让 Run 返回 error")
	fmt.Println(strings.Repeat("=", 60) + "\n")
}
