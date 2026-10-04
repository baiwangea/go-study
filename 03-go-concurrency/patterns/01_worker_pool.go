package patterns

import (
	"fmt"
	"sync"
	"time"

	"go-study/go-concurrency/level"
)

// L2-12：worker pool —— 固定并发数跑一批任务（本关最接近真实机器人）。
func L12() level.Level {
	return level.Level{
		ID:      "L2-12",
		Title:   "Worker Pool 并发限流",
		Tags:    "worker pool · 限并发 · 结果收敛",
		Pre:     "L2-02, L2-05",
		Goal:    "把 WaitGroup + channel + close 三件套组合成可复用的并发骨架：N 个工人消费一批任务",
		Observe: "10 个任务由 3 个工人处理，日志里能看到 worker 编号交错；总耗时明显小于串行",
		Questions: []string{
			"把 workers 从 3 改成 1 和改成 10，耗时怎么变？为什么不是越大越好（交易所限频 / 连接池上限）？",
			"为什么必须先 close(jobs) 再 wg.Wait()？顺序反了会怎样？（提示：range 不会退出 → 死锁）",
			"results 通道容量改成 0 会怎样？（工人写完就阻塞 → 必须有人同时读，思考「收敛协程」的写法）",
			"如果某个 worker panic 了，整个池会怎样？怎么加 recover？（对照 04-goframe L3-03 的中间件洋葱模型）",
		},
		Check: "能默写 worker pool 骨架：生产 jobs → close → workers 消费写 results → close results → 收敛端 range",
		Run: func() {
			const workers = 3
			jobs := make(chan int, 10)
			results := make(chan string, 10)
			var wg sync.WaitGroup

			// 1) 起 N 个工人：jobs 关闭后 range 自动退出
			for w := 1; w <= workers; w++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					for j := range jobs {
						time.Sleep(30 * time.Millisecond) // 模拟一次 HTTP 请求
						results <- fmt.Sprintf("worker-%d 处理任务 %d", id, j)
					}
				}(w)
			}

			// 2) 投任务，投完立刻关闭（只有发送方有权 close）
			for j := 1; j <= 10; j++ {
				jobs <- j
			}
			close(jobs)

			// 3) 工人全部收工后关闭结果通道，收敛端 range 才能结束
			go func() {
				wg.Wait()
				close(results)
			}()

			start := time.Now()
			count := 0
			for r := range results {
				count++
				if count <= 3 || count == 10 {
					fmt.Printf("  %2d) %s\n", count, r)
				} else if count == 4 {
					fmt.Println("  ...（中间省略）")
				}
			}
			fmt.Printf("  共收到 %d 条结果，耗时 %v（串行需约 %dms）\n",
				count, time.Since(start).Round(time.Millisecond), 10*30)
		},
	}
}
