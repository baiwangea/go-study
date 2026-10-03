package selects

import (
	"fmt"
	"time"

	"go-study/go-concurrency/level"
)

// L2-09：default 分支 —— 非阻塞收发，来不及就丢弃。
func L09() level.Level {
	return level.Level{
		ID:      "L2-09",
		Title:   "非阻塞 select 与丢弃策略",
		Tags:    "default · 非阻塞 · 降级",
		Pre:     "L2-08",
		Goal:    "加 default 让 select 立刻返回：宁可丢数据，也不把主流程拖死",
		Observe: "通道空时走 default 丢弃；有数据时正常读取；发送端队列满也走 default 丢单",
		Questions: []string{
			"把 default 删掉，第一次读取会发生什么？（阻塞到 100ms 后有人发送）这说明 default 改变了什么？",
			"行情场景里丢 tick 可以接受吗？什么数据绝对不能丢（订单回报）？两者分别该用哪种通道？",
			"用 default 做「非阻塞发送」时，被丢弃的任务去哪了？怎么让它至少留下日志/指标？",
		},
		Check: "能写出非阻塞收/发两种写法，并说清「可丢」与「不可丢」数据的通道选型",
		Run: func() {
			ticks := make(chan string, 1)

			// 接收侧：此刻没有行情，立刻走 default，不浪费时间
			select {
			case t := <-ticks:
				fmt.Println("  读到 tick：", t)
			default:
				fmt.Println("  接收 default：暂无行情 → 直接返回去做别的事（不阻塞）")
			}

			go func() { time.Sleep(50 * time.Millisecond); ticks <- "BTC 63001" }()
			time.Sleep(100 * time.Millisecond)
			select {
			case t := <-ticks:
				fmt.Println("  读到 tick：", t)
			default:
				fmt.Println("  接收 default：暂无行情")
			}

			// 发送侧：队列满（容量 1 已占），下 2 条会被丢弃
			orders := make(chan string, 1)
			orders <- "ORD-1 已入队"
			dropped := 0
			for _, o := range []string{"ORD-2", "ORD-3"} {
				select {
				case orders <- o:
					fmt.Printf("  %s 入队成功\n", o)
				default:
					dropped++
					fmt.Printf("  %s 队列已满 → 丢弃（不可丢的数据不能这么写！）\n", o)
				}
			}
			fmt.Printf("  当前队列长度 %d，丢弃 %d 条\n", len(orders), dropped)
		},
	}
}
