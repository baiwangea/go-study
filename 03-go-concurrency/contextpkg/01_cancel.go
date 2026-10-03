package contextpkg

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"go-study/go-concurrency/level"
)

// L2-11：context —— 让 goroutine 能被叫停。
func L11() level.Level {
	// subscribe 模拟一个永远在收行情的订阅协程
	subscribe := func(ctx context.Context, done chan<- string) {
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

	return level.Level{
		ID:      "L2-11",
		Title:   "context 取消与 goroutine 回收",
		Tags:    "context · WithCancel · WithTimeout · 泄漏治理",
		Pre:     "L2-10",
		Goal:    "把「取消」变成参数一路传下去，让所有下游协程都能收工 —— 这是 L2-06 泄漏的唯一正解",
		Observe: "调用 cancel() 后订阅协程退出，runtime.NumGoroutine() 回落到调用前的水平",
		Questions: []string{
			"把 defer cancel() 改成不调用，最后打印的 goroutine 数会怎样？（订阅协程永远收不到取消）",
			"context.WithTimeout 与 WithCancel 的关系是什么？为什么 WithTimeout 也要求 defer cancel()？",
			"ctx.Err() 有两种取值：canceled 与 deadline exceeded，分别对应什么场景？日志里为什么值得区分？",
			"为什么规范禁止把 context 存进 struct 字段，而是作为第一个参数传？",
		},
		Check: "能给任意后台协程加上「可取消」能力，并用 goroutine 数量变化证明没泄漏",
		Run: func() {
			before := runtime.NumGoroutine()

			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan string, 1)
			go subscribe(ctx, done)

			time.Sleep(100 * time.Millisecond)
			cancel() // 下达取消令
			fmt.Println("  ", <-done)

			withTimeout, cancelTimeout := context.WithTimeout(context.Background(), 80*time.Millisecond)
			defer cancelTimeout()
			start := time.Now()
			<-withTimeout.Done()
			fmt.Printf("  WithTimeout 到期：%v（等待 %v）\n", withTimeout.Err(), time.Since(start).Round(10*time.Millisecond))

			time.Sleep(20 * time.Millisecond)
			fmt.Printf("  goroutine 数量：进入本关前 %d → 现在 %d（订阅协程已回收）\n", before, runtime.NumGoroutine())
		},
	}
}
