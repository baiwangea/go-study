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
		Observe: "先打印 sent，再打印 got；中间有 300ms 延迟 —— 说明发送方一直等到接收方取走",
		Questions: []string{
			"把接收挪到发送之前（调换两段代码顺序），为什么还是同样输出？反过来：只有发送没有接收会怎样？",
			"ch := make(chan int) 然后 ch <- 1 就结束，程序报什么错？（all goroutines are asleep - deadlock!）",
			"机器人里什么时候该用无缓冲？（提示：下单指令必须被执行方取走才算交出去）",
		},
		Check: "能说出无缓冲 channel 的两个阻塞条件，并解释它为什么等价于一次「同步会合」",
		Run: func() {
			ch := make(chan string)

			go func() {
				got := <-ch
				fmt.Printf("  [接收方] 收到指令：%s\n", got)
			}()

			time.Sleep(300 * time.Millisecond) // 故意让接收方慢一点，好观察阻塞
			start := time.Now()
			ch <- "BUY BTC 0.1"
			fmt.Printf("  [发送方] 交接完成，等待了 %v\n", time.Since(start).Round(10*time.Millisecond))
			time.Sleep(100 * time.Millisecond) // 等接收方把日志打完
		},
	}
}
