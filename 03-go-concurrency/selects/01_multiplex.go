package selects

import (
	"fmt"
	"time"

	"go-study/go-concurrency/level"
)

// L2-07：select —— 同时等多个通道。
func L07() level.Level {
	return level.Level{
		ID:      "L2-07",
		Title:   "select 多路复用",
		Tags:    "select · case · 随机公平",
		Pre:     "L2-06",
		Goal:    "用 select 同时等多个 channel：谁先就绪就走谁，都阻塞才等待",
		Observe: "两个源快慢不同 → 先命中快的那个；两个同时就绪时会随机挑（防饿死）",
		Questions: []string{
			"把 fast 的延时从 50ms 改成 300ms，输出会换成哪个分支？说明 select 的判定依据是什么？",
			"把两个 case 的延时改成完全相同并循环 10 次，命中分布如何？为什么 Go 要故意随机？",
			"select 里如果有 default 分支，语义会怎么变？（下一关实测）",
		},
		Check: "能说出 select 的三条规则：有就绪就执行、多个就绪随机选、都未就绪则阻塞（除非有 default）",
		Run: func() {
			fast := make(chan string)
			slow := make(chan string)

			go func() { time.Sleep(50 * time.Millisecond); fast <- "BTC 63000" }()
			go func() { time.Sleep(200 * time.Millisecond); slow <- "ETH 3000" }()

			start := time.Now()
			select {
			case p := <-fast:
				fmt.Printf("  命中 fast 分支：%s（等待 %v）\n", p, time.Since(start).Round(10*time.Millisecond))
			case p := <-slow:
				fmt.Printf("  命中 slow 分支：%s（等待 %v）\n", p, time.Since(start).Round(10*time.Millisecond))
			}

			// 两个分枝同时就绪（预填缓冲，保证都有数据）：连选 4 次，看它随机挑选
			a, b := make(chan int, 2), make(chan int, 2)
			a <- 1
			a <- 1
			b <- 1
			b <- 1
			fmt.Println("  同时就绪时的选择（应为随机混排，如 [a b a b]）：", pick4(a, b))
		},
	}
}

func pick4(a, b chan int) []string {
	var hits []string
	for i := 0; i < 4; i++ {
		select {
		case <-a:
			hits = append(hits, "a")
		case <-b:
			hits = append(hits, "b")
		}
	}
	return hits
}
