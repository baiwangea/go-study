package channels

import (
	"fmt"
	"time"

	"go-study/go-concurrency/level"
)

// L2-04：有缓冲 channel —— 队列深度与阻塞边界。
func L04() level.Level {
	return level.Level{
		ID:      "L2-04",
		Title:   "缓冲队列：什么时候开始堵",
		Tags:    "buffered chan · 背压 · 队列深度",
		Pre:     "L2-03",
		Goal:    "理解 make(chan T, n) 的语义：前 n 次发送不阻塞，第 n+1 次开始阻塞（这就是背压）",
		Observe: "容量 2 的通道，前 2 条秒过，第 3 条卡住直到消费者取走一个",
		Questions: []string{
			"把容量改成 0 和改成 5，分别观察第几条开始阻塞 —— 容量与「谁先阻塞」的关系是什么？",
			"len(ch) 与 cap(ch) 分别告诉你什么？打印出来看看。",
			"机器人里行情通道该设多大？设太大会怎样（价格过期还在下单），设太小会怎样（生产者被拖慢）？",
		},
		Check: "能解释缓冲 = 生产者与消费者的速度差容忍度，并说清 len/cap 的区别",
		Run: func() {
			ch := make(chan string, 2)

			for i, tick := range []string{"tick-1", "tick-2", "tick-3"} {
				start := time.Now()
				ch <- tick // 第 3 次会阻塞，直到下面的消费者取走一个
				blocked := time.Since(start)
				fmt.Printf("  发送 %s 用时 %v，当前 len=%d cap=%d\n", tick, blocked.Round(10*time.Millisecond), len(ch), cap(ch))

				if i == 0 {
					go func() {
						time.Sleep(200 * time.Millisecond)
						fmt.Printf("  [消费者] 取走 %s\n", <-ch)
					}()
				}
			}
			fmt.Println("  ↑ 第 3 条被「背压」住了：队列容量 2 已满，只能等消费者腾位置")
			time.Sleep(300 * time.Millisecond)
		},
	}
}
