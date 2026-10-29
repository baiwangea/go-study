//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
//
// 跑法：cd 03-go-concurrency && go run xlib/04_semaphore.go
// 依赖：golang.org/x/sync/semaphore
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
)

/*
═══════════════════════════════════════════════════════════════════════
L2-22：加权信号量 —— 同时只能有 N 个在飞
═══════════════════════════════════════════════════════════════════════

【核心概念】
    三样东西经常被混用，目的不一样：

      rate.Limiter     限制「每秒多少次」（L2-20）
      errgroup.SetLimit 限制「这组任务最多几个 goroutine」（L2-19）
      semaphore        限制「任意代码段同时占几份资源」

    Weighted 可以一次占多份：拉 1 个行情占 1，拉 K 线占 3。
    Acquire 会等到有空位或 ctx 取消；TryAcquire 立刻返回 false。
    占了必须 Release，通常 defer。忘了 Release，后面的人永远进不来。

【你会看到】
    ✓ 权重 3 的信号量，同时最多 3 个单位在跑
    ✓ 一个任务占 2，剩下的 1 不够再进一个占 2 的
    ✓ ctx 取消后 Acquire 立刻返回，不会死等

【思考题】
    1. 交易所「最多 10 个并发请求」用信号量还是 rate.Limiter？
    2. Acquire 之后 panic 了、没走到 Release，会怎样？
    3. 为什么信号量本身不负责开 goroutine？

【通关标准】
    能说出信号量、限频、errgroup.SetLimit 各管哪一层，
    并能写出 Acquire + defer Release。

【前置】L2-19 · L2-20     【标签】semaphore · 并发上限 · TryAcquire
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L2-22 加权信号量")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	fmt.Println("📝 示例 1：权重 3，三个轻任务同时跑")
	fmt.Println(strings.Repeat("-", 60))
	sem := semaphore.NewWeighted(3)
	var wg sync.WaitGroup
	start := time.Now()
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if err := sem.Acquire(context.Background(), 1); err != nil {
				fmt.Println("  获取失败：", err)
				return
			}
			defer sem.Release(1)
			fmt.Printf("  ✅ 任务 %d 占 1 份，t=%v\n", id, time.Since(start).Round(time.Millisecond))
			time.Sleep(80 * time.Millisecond)
		}(i)
	}
	wg.Wait()
	fmt.Println("  ➡️  三个任务几乎同时开始：各自只占 1，加起来正好 3。")

	fmt.Println("\n📌 示例 2：一次占 2，第二个占 2 的要等")
	fmt.Println(strings.Repeat("-", 60))
	sem = semaphore.NewWeighted(3)
	order := make(chan string, 4)
	run := func(name string, weight int64, hold time.Duration) {
		defer wg.Done()
		if err := sem.Acquire(context.Background(), weight); err != nil {
			order <- name + " 失败"
			return
		}
		defer sem.Release(weight)
		order <- fmt.Sprintf("%s 占 %d，t=%v", name, weight, time.Since(start).Round(time.Millisecond))
		time.Sleep(hold)
	}
	start = time.Now()
	wg.Add(3)
	go run("A", 2, 80*time.Millisecond)
	go run("B", 2, 20*time.Millisecond) // 剩下 1，进不去
	go run("C", 1, 20*time.Millisecond) // 剩下 1，能进
	wg.Wait()
	close(order)
	for line := range order {
		fmt.Println("  ", line)
	}
	fmt.Println("  ➡️  A 占 2、C 占 1 先跑；B 要等 A 释放出 2 份。")

	fmt.Println("\n🔍 示例 3：TryAcquire 满了就走，ctx 取消则 Acquire 返回")
	fmt.Println(strings.Repeat("-", 60))
	sem = semaphore.NewWeighted(1)
	_ = sem.Acquire(context.Background(), 1)
	fmt.Printf("  TryAcquire(1) = %v  ← 已经满了\n", sem.TryAcquire(1))
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	err := sem.Acquire(ctx, 1)
	cancel()
	fmt.Printf("  带超时的 Acquire → %v\n", err)
	sem.Release(1)
	fmt.Println("  ➡️  行情可以丢就 TryAcquire；下单必须发出去就用带 ctx 的 Acquire。")

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ 信号量限制「同时占几份资源」，不负责开 goroutine")
	fmt.Println("  ✅ Acquire + defer Release；忘了 Release 后面全部饿死")
	fmt.Println("  ✅ Weighted 可以一次占多份（拉 K 线比拉一笔 tick 更重）")
	fmt.Println("  ✅ 限频管速率，SetLimit 管一组任务，信号量管任意临界段")
	fmt.Println(strings.Repeat("=", 60) + "\n")
}
