//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 03-go-concurrency && go run xlib/01_errgroup.go
// 依赖：golang.org/x/sync（本关需要，见 03-go-concurrency/go.mod）
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

/*
═══════════════════════════════════════════════════════════════════════
L2-19：errgroup —— 并发任务 + 第一个错误
═══════════════════════════════════════════════════════════════════════

【核心概念】
    errgroup.Group ≈ WaitGroup + 错误传递 + 失败即取消。
    它并发跑一组任务，Wait() 返回「第一个」出错的任务的错误；
    配合 WithContext，一个任务失败会取消其余任务的 ctx，让大家尽快收工，
    而不是白等到各自超时。

【你会看到】
    ✓ 3 个任务并发跑，Wait() 返回最先失败的那个错误
    ✓ 一个任务 30ms 就失败 → 另外两个不会白等 2 秒
    ✓ 手写 WaitGroup + Mutex 聚合错误要多写多少代码、缺哪些能力
    ✓ SetLimit 给并发封顶，TryGo 在满了的时候直接返回 false

【思考题】（动手改代码验证，别空想）
    1. 示例 1 里三个任务都返回错误，Wait() 返回哪一个？为什么不是全部？
    2. errgroup 只返回第一个错误 —— 想收集「全部错误」该怎么写？
       （提示：errors.Join，或者在闭包里自己往 slice 里塞）
    3. 示例 4 里 SetLimit 为什么必须在第一次 g.Go() 之前调用？
    4. Go 1.22 之后循环变量是每轮独立的，示例里的 `i := i` 还需要吗？
       为什么老代码里到处都是它？

【通关标准】
    能说清「WaitGroup 为什么不够用」，并默写出
    「errgroup + WithContext + SetLimit」的并发骨架

【前置】L2-12              【标签】errgroup · 错误传递 · 失败取消 · 并发上限
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n=== L2-19 errgroup ===")

	// ── 示例 1：并发跑 3 个任务，只拿第一个错误 ─────────────────────────
	fmt.Println("\n【示例 1】并发跑 3 个任务，Wait() 只返回第一个错误")
	var g errgroup.Group
	g.Go(func() error {
		time.Sleep(30 * time.Millisecond)
		return nil // 成功
	})
	g.Go(func() error {
		time.Sleep(10 * time.Millisecond)
		return errors.New("任务 B 失败（10ms 就返回了）")
	})
	g.Go(func() error {
		time.Sleep(20 * time.Millisecond)
		return errors.New("任务 C 失败（20ms 才返回）")
	})
	err := g.Wait()
	fmt.Printf("  Wait() 返回：%v\n", err)
	fmt.Println("  ↑ B 先返回，所以拿到的是 B 的错误 —— errgroup 只保留「第一个」，")
	fmt.Println("    C 的错误被丢弃了。要全部错误就得自己在闭包里聚合")

	// ── 示例 2：WithContext —— 一个失败，全部取消 ───────────────────────
	fmt.Println("\n【示例 2】WithContext：一个任务失败，其余任务立刻收工")
	start := time.Now()
	g2, ctx := errgroup.WithContext(context.Background())
	for i := 1; i <= 3; i++ {
		i := i // Go 1.22 起循环变量每轮独立，这里保留是为了让意图更明确
		g2.Go(func() error {
			if i == 2 {
				time.Sleep(30 * time.Millisecond)
				return fmt.Errorf("任务 %d 失败", i)
			}
			select {
			case <-time.After(2 * time.Second):
				return fmt.Errorf("任务 %d 竟然跑完了", i) // 不该走到这里
			case <-ctx.Done():
				return fmt.Errorf("任务 %d 收到取消信号：%w", i, ctx.Err())
			}
		})
	}
	err = g2.Wait()
	fmt.Printf("  总耗时 %v（没人取消的话，1 号和 3 号要各跑满 2 秒）\n",
		time.Since(start).Round(time.Millisecond))
	fmt.Printf("  Wait() 返回：%v\n", err)
	fmt.Println("  ↑ ctx 被 errgroup 自动取消：先记录错误、再 cancel，所以拿到的是「2 号」的错误")

	// ── 示例 3：手写 WaitGroup 版，看看差在哪 ───────────────────────────
	fmt.Println("\n【示例 3】手写版：WaitGroup + Mutex 聚合错误（对比代码量和能力）")
	start = time.Now()
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error
	record := func(e error) { // 自己实现「只记第一个错误」
		mu.Lock()
		if firstErr == nil {
			firstErr = e
		}
		mu.Unlock()
	}
	ctx3, cancel3 := context.WithCancel(context.Background()) // 自己建可取消的 ctx
	defer cancel3()
	tasks := []func() error{
		func() error { time.Sleep(30 * time.Millisecond); return nil },
		func() error { time.Sleep(10 * time.Millisecond); return errors.New("任务 B 失败") },
		func() error {
			select {
			case <-time.After(2 * time.Second):
				return nil
			case <-ctx3.Done():
				return ctx3.Err()
			}
		},
	}
	for _, task := range tasks {
		wg.Add(1)
		go func(task func() error) {
			defer wg.Done()
			if e := task(); e != nil {
				record(e)
				cancel3() // 想「失败即取消」，就得自己补这一句
			}
		}(task)
	}
	wg.Wait()
	fmt.Printf("  手写版 firstErr = %v，耗时 %v\n", firstErr, time.Since(start).Round(time.Millisecond))
	fmt.Println("  这段能跑，但比 errgroup 版多写了十来行，而且每加一个需求就多一层胶水：")
	fmt.Println("    1) 错误传递：WaitGroup.Wait() 没有返回值，必须自己加锁聚合")
	fmt.Println("    2) 失败即取消：得自己 WithCancel + 在每个出错处手动 cancel")
	fmt.Println("    3) 并发上限：得自己拿带缓冲 channel 当信号量")
	fmt.Println("    4) WaitGroup 还有 Add/Wait 复用时的竞态约束，很容易写错")

	// ── 示例 4：SetLimit / TryGo ────────────────────────────────────────
	fmt.Println("\n【示例 4】SetLimit：给并发数封顶")
	var g4 errgroup.Group
	g4.SetLimit(2) // 必须在第一次 Go() 之前调用
	start = time.Now()
	for i := 1; i <= 4; i++ {
		g4.Go(func() error {
			time.Sleep(100 * time.Millisecond)
			return nil
		})
	}
	if err := g4.Wait(); err != nil {
		fmt.Println("  出错：", err)
	}
	fmt.Printf("  4 个任务 × 100ms、并发上限 2 → 耗时 %v（≈200ms，而不是 100ms 或 400ms）\n",
		time.Since(start).Round(time.Millisecond))

	var g5 errgroup.Group
	g5.SetLimit(1)
	_ = g5.TryGo(func() error { time.Sleep(50 * time.Millisecond); return nil })
	accepted := g5.TryGo(func() error { return nil })
	fmt.Printf("  上限为 1 时，第二个 TryGo 立刻返回 %v（不阻塞，由调用方决定排队还是丢弃）\n", accepted)
	_ = g5.Wait()

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n【关键总结】")
	fmt.Println("  • g.Go(f) 起一个任务；g.Wait() 等全部结束并返回第一个错误")
	fmt.Println("  • WithContext 让「任一失败」自动取消其余任务，避免白等")
	fmt.Println("  • SetLimit(n) 限并发（要在 Go 之前调）；TryGo 满了直接返回 false")
	fmt.Println("  • WaitGroup 不够用的三点：没有错误返回、没有失败即取消、没有并发上限")
	fmt.Println("  • 该用：并发拉多个交易对行情、并发初始化多个组件 —— 任一失败就整体失败")
	fmt.Println("  • 不该用：需要「部分成功 / 收集全部错误」（用 WaitGroup + errors.Join 自己聚合）")
}
