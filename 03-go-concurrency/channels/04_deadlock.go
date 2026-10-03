package channels

import (
	"fmt"
	"time"

	"go-study/go-concurrency/level"
)

// L2-06：三种典型「卡住」，以及如何安全地观察它们。
func L06() level.Level {
	wait := func(who string, ch chan int, send bool) {
		if send {
			go func() { ch <- 42 }() // 发送方将永久阻塞
			time.Sleep(150 * time.Millisecond)
			fmt.Printf("  %s：无人接收 → 发送方 goroutine 永久阻塞（泄漏，仍在内存里）\n", who)
			return
		}
		got := make(chan bool, 1)
		go func() { <-ch; got <- true }() // 接收方将永久阻塞
		select {
		case <-got:
			fmt.Printf("  %s：竟然收到了\n", who)
		case <-time.After(150 * time.Millisecond):
			fmt.Printf("  %s：无人发送 → 接收方永久阻塞\n", who)
		}
	}

	return level.Level{
		ID:      "L2-06",
		Title:   "死锁 / 泄漏 / nil channel",
		Tags:    "deadlock · goroutine leak · nil chan",
		Pre:     "L2-05",
		Goal:    "认出三种最常见的卡死形态，并学会用超时探测代替「让它挂着」",
		Observe: "三种情况都在 150ms 内被判定为「阻塞」，程序不会挂住；泄漏的 goroutine 依然存在",
		Questions: []string{
			"把上面的超时探测全部删掉，只留阻塞语句 —— 报错是 fatal error: all goroutines are asleep - deadlock!，为什么它不可 recover？",
			"本关留下的 3 个阻塞 goroutine 什么时候释放内存？（这就是泄漏，L2-11 用 context 才能真正收回来）",
			"nil channel（var ch chan int）收发都永久阻塞 —— 所以 select 里可以用它「临时禁用某个分支」，怎么构造这个用法？",
		},
		Check: "能说出无人接收的发送、无人发送的接收、nil channel 三种阻塞的成因，以及 fatal 与 panic 的区别",
		Run: func() {
			wait("情况一", make(chan int), true)  // 只有发送
			wait("情况二", make(chan int), false) // 只有接收

			var nilCh chan int // nil channel
			done := make(chan string)
			go func() { <-nilCh; done <- "收到" }()
			select {
			case msg := <-done:
				fmt.Println("  情况三：", msg)
			case <-time.After(150 * time.Millisecond):
				fmt.Println("  情况三：nil channel 收发都永久阻塞 —— 可用来禁用 select 分支")
			}

			fmt.Println("  ↓ 若把上面三处超时删掉，程序会立刻 fatal error（deadlock），且无法 recover")
			time.Sleep(10 * time.Millisecond)
		},
	}
}
