package channels

import (
	"fmt"
	"time"

	"go-study/go-concurrency/level"
)

// L2-05：close、range 与「取到零值」的歧义。
func L05() level.Level {
	return level.Level{
		ID:      "L2-05",
		Title:   "close / range / 逗号 ok",
		Tags:    "close · range · 结束信号",
		Pre:     "L2-04",
		Goal:    "学会用 close 广播「没有更多数据了」，接收端用 for range 自动收尾",
		Observe: "range 打印完 3 条后自动结束；已关闭通道读出的 x, ok 里 ok=false",
		Questions: []string{
			"把 close(ch) 删掉：range 取完 3 条后就会永久阻塞，最终报 fatal error: all goroutines are asleep - deadlock! —— 为什么这种“忘关闭”比一般 bug 难发现？",
			"读取的两种写法：v := <-ch 与 v, ok := <-ch，通道关闭后分别得到什么？为什么只看零值不可信？",
			"发送方和接收方谁该负责 close？（提示：只有发送方关；两个发送者都关会 panic: send on closed channel —— 怎么做单发送者？）",
		},
		Check: "能说清「通道关闭后仍可读剩余数据，但再写入会 panic」，并解释为什么机器人里禁止接收方 close",
		Run: func() {
			ch := make(chan int, 3)
			for i := 1; i <= 3; i++ {
				ch <- i * 100
			}
			close(ch) // 生产完毕，通知接收方

			for v := range ch { // 通道关闭且取空后自动退出循环
				fmt.Printf("  收到价格 tick：%d\n", v)
			}

			v, ok := <-ch // 已关闭且已排空
			fmt.Printf("  再读一次：v=%d ok=%v（ok=false 表示通道已关闭，零值不可信）\n", v, ok)

			time.Sleep(10 * time.Millisecond)
			fmt.Println("  ↑ 机器人里常用套路：行情源关闭 → 所有订阅者的 range 自动结束 → 优雅停机")
		},
	}
}
