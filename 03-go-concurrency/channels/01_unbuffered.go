package channels

import (
	"fmt"
	"time"

	"go-study/go-concurrency/level"
)

// L2-03：无缓冲 channel —— 一次同步握手。
func L03() level.Level {
	return level.Level{
		ID:      "L2-03",
		Title:   "无缓冲 channel：发送即交接",
		Tags:    "chan · 同步 · 阻塞语义",
		Pre:     "L2-02",
		Goal:    "理解无缓冲 channel 的两端会「碰面」：发送阻塞到有人接收，接收阻塞到有人发送",
		Observe: "发送方被阻塞约 200ms（直到接收方起床取走），随后两端几乎同时打印",
		Questions: []string{
			"把接收方 goroutine 里的 time.Sleep 删掉，发送方还会阻塞 200ms 吗？为什么？（这就验证了「阻塞时长取决于对端」）",
			"ch := make(chan int) 然后 ch <- 1 就结束，程序报什么错？（all goroutines are asleep - deadlock!）",
			"机器人里什么时候该用无缓冲？（提示：下单指令必须被执行方取走才算交出去）",
		},
		Check: "能说出无缓冲 channel 的两个阻塞条件，并解释它为什么等价于一次「同步会合」",
		Run: func() {
			ch := make(chan string)

			// 接收方故意晚 200ms 才起床 —— 发送方必须原地等它
			go func() {
				time.Sleep(200 * time.Millisecond)
				got := <-ch
				fmt.Printf("  [接收方] 收到指令：%s\n", got)
			}()

			start := time.Now()
			ch <- "BUY BTC 0.1" // 无缓冲：没有接收者就堵住
			fmt.Printf("  [发送方] 阻塞了 %v 才完成交接\n", time.Since(start).Round(10*time.Millisecond))
			time.Sleep(100 * time.Millisecond) // 等接收方把日志打完
		},
	}
}
