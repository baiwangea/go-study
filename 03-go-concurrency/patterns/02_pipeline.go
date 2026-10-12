//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 03-go-concurrency && go run patterns/02_pipeline.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"sync"
	"time"
)

/*
═══════════════════════════════════════════════════════════════════════
L2-18：pipeline 与 fan-out / fan-in
═══════════════════════════════════════════════════════════════════════

【核心概念】
    把「一个 channel 进、一个 channel 出」的处理步骤串成流水线，
    每个阶段跑在自己的 goroutine 里。想加速就把某个阶段复制多份（fan-out），
    再用 merge 把多路结果合并回一路（fan-in）。整条线必须能被 ctx 取消，
    否则消费者提前退出时，上游会永远卡在发送上 —— 那就是 goroutine 泄漏。

【你会看到】
    ✓ 三阶段流水线 gen → square → square
    ✓ 3 个 worker 抢同一个输入（fan-out），再合并成一路（fan-in）
    ✓ 中途 cancel() 后 goroutine 数回落到基线（不泄漏）
    ✓ 忘记 select ctx.Done() 的阶段，取消后仍然卡死（泄漏）

【思考题】（动手改代码验证，别空想）
    1. 示例 2 里把 worker 从 3 个改成 1 个、6 个，耗时怎么变？
       为什么不是越多越好？
    2. 示例 3 里把 `break` 换成「读完全部再 cancel」，goroutine 数还会回落吗？
    3. 为什么每个阶段只能由「发送方」close？两个阶段都 close 会怎样？
    4. 如果某个阶段 panic 了，整条流水线会怎样？该怎么加 recover？

【通关标准】
    能默写带 ctx 的 pipeline 骨架，并说清 merge 为什么要「等所有上游
    关闭后再 close(out)」才不会漏协程

【前置】L2-12              【标签】pipeline · fan-out · fan-in · 取消 · 泄漏
═══════════════════════════════════════════════════════════════════════
*/

// ── 流水线的三个阶段与两个工具函数 ──────────────────────────────────────

// gen 是流水线的源头：把 nums 依次发到输出 channel，发完就 close。
// 每次发送都 select ctx.Done()，保证取消时能立刻退出（不泄漏）。
func gen(ctx context.Context, nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			select {
			case out <- n:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// genSlow 和 gen 一样，只是每发一个值睡 50ms，方便演示「中途取消」。
func genSlow(ctx context.Context, nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			select {
			case out <- n:
			case <-ctx.Done():
				return
			}
			select {
			case <-time.After(50 * time.Millisecond):
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// square 是一个普通处理阶段：读入 v，送出 v*v。
// 它自己起一个 goroutine，所以「每个阶段都是独立并发的」。
func square(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for v := range in {
			select {
			case out <- v * v:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// merge 是 fan-in：把多个上游 channel 合并成一路。
// 关键点：用 WaitGroup 等所有上游都关闭之后，才 close(out) ——
// 这样下游的 range 才能正常结束，既不会死锁也不会漏协程。
func merge(ctx context.Context, ins ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	for _, in := range ins {
		wg.Add(1)
		go func(in <-chan int) {
			defer wg.Done()
			for v := range in {
				select {
				case out <- v:
				case <-ctx.Done():
					return
				}
			}
		}(in)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// leakySquare 是「忘记 ctx」的反面教材：只写不检查取消。
// 一旦下游不再接收，它就永久阻塞在发送上 —— goroutine 泄漏。
func leakySquare(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for v := range in {
			out <- v * v // ← 少了 select ctx.Done()
		}
	}()
	return out
}

// waitGoroutines 轮询等待 goroutine 数量回落到 target 以下，最多等 timeout，
// 返回最终观察到的数量 —— 用它来证明「到底有没有泄漏」。
func waitGoroutines(target int, timeout time.Duration) int {
	deadline := time.Now().Add(timeout)
	for {
		n := runtime.NumGoroutine()
		if n <= target || time.Now().After(deadline) {
			return n
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func main() {
	fmt.Println("\n=== L2-18 pipeline / fan-out / fan-in ===")

	// ── 示例 1：三阶段流水线 ────────────────────────────────────────────
	fmt.Println("\n【示例 1】三阶段流水线：gen → square → square")
	ctx := context.Background()
	c1 := gen(ctx, 1, 2, 3, 4, 5)
	c2 := square(ctx, c1)
	c3 := square(ctx, c2)
	got := []int{}
	for v := range c3 {
		got = append(got, v)
	}
	fmt.Printf("  输入 1..5 → 输出 %v\n", got)
	fmt.Println("  每个阶段都是「一个 channel 进、一个 channel 出」+ 自己的 goroutine，")
	fmt.Println("  数据边走边算 —— 这就是 pipeline 的全部秘密")

	// ── 示例 2：fan-out + fan-in ────────────────────────────────────────
	fmt.Println("\n【示例 2】fan-out（3 个 worker 抢同一个输入）+ fan-in（合并回一路）")
	ctx2 := context.Background()
	in := gen(ctx2, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12)
	start := time.Now()
	// fan-out：同一个 in 交给 3 个 square 阶段，谁空闲谁拿
	w1 := square(ctx2, in)
	w2 := square(ctx2, in)
	w3 := square(ctx2, in)
	// fan-in：把 3 路合并回一路
	results := []int{}
	for v := range merge(ctx2, w1, w2, w3) {
		results = append(results, v)
	}
	sort.Ints(results) // fan-in 后顺序不确定，排序只是为了让输出稳定
	fmt.Printf("  12 个输入 → 3 个 worker → 合并出 %d 个结果，耗时 %v\n",
		len(results), time.Since(start).Round(time.Microsecond))
	fmt.Printf("  结果（已排序）：%v\n", results)
	fmt.Println("  ↑ 顺序不确定是 fan-in 的固有特征：谁先算完谁先到。")
	fmt.Println("    需要顺序就在数据里带序号，在最后一步按序号归位")

	// ── 示例 3：ctx 取消 + 无泄漏 ───────────────────────────────────────
	fmt.Println("\n【示例 3】用 ctx 取消整条流水线，并验证 goroutine 没泄漏")
	time.Sleep(50 * time.Millisecond) // 等前面几个示例的协程彻底退干净
	base := runtime.NumGoroutine()
	fmt.Printf("  goroutine 基线 = %d\n", base)

	ctx3, cancel3 := context.WithCancel(context.Background())
	slow := genSlow(ctx3, 1, 2, 3, 4, 5, 6, 7, 8)
	out3 := square(ctx3, square(ctx3, slow))

	read := 0
	for range out3 {
		read++
		if read == 2 {
			cancel3() // 只读 2 个就取消
			break
		}
	}
	after := waitGoroutines(base, 2*time.Second)
	fmt.Printf("  读了 %d 个就 cancel() → goroutine 数回落到 %d（基线 %d）\n", read, after, base)
	fmt.Println("  每个阶段都 select ctx.Done()，取消后立刻退出 —— 这就是「不泄漏」的写法")

	// ── 示例 4：忘记 ctx 的后果 ─────────────────────────────────────────
	fmt.Println("\n【示例 4】反面教材：阶段里忘记 select ctx.Done() → 取消也收不回来")
	time.Sleep(50 * time.Millisecond)
	base = runtime.NumGoroutine()
	ctx4, cancel4 := context.WithCancel(context.Background())
	out4 := leakySquare(gen(ctx4, 1, 2, 3, 4, 5, 6))
	<-out4    // 只读 1 个
	cancel4() // 取消 —— 但 leakySquare 根本不看 ctx
	after = waitGoroutines(base, 500*time.Millisecond)
	fmt.Printf("  只读了 1 个就 cancel() → 等 0.5 秒后 goroutine 数 = %d（基线 %d，多出来的是泄漏）\n", after, base)
	fmt.Println("  ↑ 多出来的那个 goroutine 永久卡在 `out <- v*v` 上，到程序结束都收不回来")
	fmt.Println("    修正办法只有一个：所有可能阻塞的收发都塞进 select，并带上 <-ctx.Done()")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n【关键总结】")
	fmt.Println("  • pipeline = 一串「chan 进、chan 出」的阶段，每阶段一个 goroutine")
	fmt.Println("  • fan-out：多个 worker 抢同一个输入；fan-in：用 merge 合并回一路")
	fmt.Println("  • merge 必须等所有上游关闭后再 close(out)，否则下游 range 会死锁")
	fmt.Println("  • 所有阻塞的收发都写成 select { case out<-v: case <-ctx.Done(): return }")
	fmt.Println("  • 该用：多步骤数据处理（拉行情 → 计算指标 → 落库/下单）要分阶段并发时")
	fmt.Println("  • 不该用：阶段只有一步、或必须严格保序（流水线的乱序是固有的）")
}
