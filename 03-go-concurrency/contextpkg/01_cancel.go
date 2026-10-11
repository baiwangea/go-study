//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 03-go-concurrency && go run contextpkg/01_cancel.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

/*
═══════════════════════════════════════════════════════════════════════
L2-11：context 取消与 goroutine 回收
═══════════════════════════════════════════════════════════════════════

【核心概念】
    把「取消」变成参数一路传下去，让所有下游协程都能收工 ——
    这是 L2-06 泄漏的唯一正解。

【你会看到】
    ✓ 调用 cancel() 后订阅协程退出
    ✓ runtime.NumGoroutine() 回落到调用前的水平

【思考题】（动手改代码验证，别空想）
    1. 把 defer cancel() 改成不调用，最后打印的 goroutine 数会怎样？
       （订阅协程永远收不到取消）
    2. context.WithTimeout 与 WithCancel 的关系是什么？
       为什么 WithTimeout 也要求 defer cancel()？
    3. ctx.Err() 有两种取值：canceled 与 deadline exceeded，
       分别对应什么场景？日志里为什么值得区分？
    4. 为什么规范禁止把 context 存进 struct 字段，而是作为第一个参数传？

【通关标准】
    能给任意后台协程加上「可取消」能力，并用 goroutine 数量变化证明没泄漏

【前置】L2-10              【标签】context · WithCancel · WithTimeout · 泄漏治理
═══════════════════════════════════════════════════════════════════════
*/

// subscribe 模拟一个永远在收行情的订阅协程：
// 只有 ctx.Done() 是它唯一的正常出口。
func subscribe(ctx context.Context, done chan<- string) {
	for {
		select {
		case <-ctx.Done(): // 取消信号：唯一能让我们离开的出口
			done <- "订阅已退出，原因：" + ctx.Err().Error()
			return
		default:
			time.Sleep(20 * time.Millisecond) // 假装在处理一条 tick
		}
	}
}

func main() {
	fmt.Println("\n=== L2-11 context 取消 ===")

	// ── 示例 1：WithCancel 主动叫停订阅协程 ─────────────────────────────
	fmt.Println("\n【示例 1】cancel() 让订阅协程收工")
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan string, 1)
	go subscribe(ctx, done)

	time.Sleep(100 * time.Millisecond)
	cancel() // 下达取消令
	fmt.Println("  ", <-done)

	// ── 示例 2：WithTimeout 到期自动取消 ────────────────────────────────
	fmt.Println("\n【示例 2】WithTimeout 到点自动取消")
	withTimeout, cancelTimeout := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancelTimeout()
	start := time.Now()
	<-withTimeout.Done()
	fmt.Printf("  WithTimeout 到期：%v（等待 %v）\n", withTimeout.Err(), time.Since(start).Round(10*time.Millisecond))

	time.Sleep(20 * time.Millisecond)
	fmt.Printf("  goroutine 数量：进入本关前 %d → 现在 %d（订阅协程已回收）\n", before, runtime.NumGoroutine())

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n【关键总结】")
	fmt.Println("  • context 是「取消信号」的载体，沿调用链一路传下去")
	fmt.Println("  • WithCancel / WithTimeout 都要 defer cancel()，否则自身会泄漏")
	fmt.Println("  • ctx.Err() 区分 canceled（主动）与 deadline exceeded（超时）")
	fmt.Println("  • context 作为函数第一个参数传，不要塞进 struct 字段")
}
