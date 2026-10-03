package selects

import (
	"fmt"
	"time"

	"go-study/go-concurrency/level"
)

// L2-10：ticker 周期轮询 —— 记得 Stop，否则 goroutine 永不退出。
func L10() level.Level {
	return level.Level{
		ID:      "L2-10",
		Title:   "Ticker 轮询与 Stop",
		Tags:    "ticker · 定时 · 泄漏",
		Pre:     "L2-09",
		Goal:    "用 time.NewTicker 做周期性任务（每 60ms 拉一次价），并且必须 Stop 释放资源",
		Observe: "ticker 每 60ms 触发一次，循环 3 次后 break；不 Stop 时 timer 仍挂在运行时里",
		Questions: []string{
			"time.Tick(ch) 与 time.NewTicker 有什么区别？为什么 Tick 的通道没法关、容易泄漏？",
			"把 defer ticker.Stop() 删掉并循环 100 万次，用 runtime.NumGoroutine() 或 pprof 看看差异",
			"如果一轮拉取耗时超过 60ms，未消费的 tick 会堆积吗？（提示：ticker 通道容量为 1，会丢 tick —— 所以要做「处理完再等下一次」）",
		},
		Check: "能写出正确的轮询循环（带 Stop），并解释为什么「tick 丢失」在行情场景需要特殊处理",
		Run: func() {
			ticker := time.NewTicker(60 * time.Millisecond)
			defer ticker.Stop() // 关键：离开作用域就释放

			rounds := 0
			start := time.Now()
			for range ticker.C {
				rounds++
				fmt.Printf("  第 %d 次轮询：假装请求交易所（耗时 10ms）\n", rounds)
				time.Sleep(10 * time.Millisecond)
				if rounds == 3 {
					break // 真实机器人里换成 <-ctx.Done()，见 L2-11
				}
			}
			fmt.Printf("  共轮询 %d 次，耗时 %v；已 defer ticker.Stop()，无残留 timer\n", rounds, time.Since(start).Round(time.Millisecond))
		},
	}
}
