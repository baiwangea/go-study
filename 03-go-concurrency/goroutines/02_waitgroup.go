package goroutines

import (
	"fmt"
	"sync"
	"time"

	"go-study/go-concurrency/level"
)

// L2-02：WaitGroup —— 等一组 goroutine 干完活。
func L02() level.Level {
	price := func(symbol string, delay time.Duration) float64 {
		time.Sleep(delay) // 模拟一次行情请求的耗时
		return float64(len(symbol)) * 3.14
	}

	return level.Level{
		ID:      "L2-02",
		Title:   "WaitGroup 三点规矩",
		Tags:    "WaitGroup · Add/Done/Wait",
		Pre:     "L2-01",
		Goal:    "掌握 Add / Done / Wait 的调用位置，告别用 sleep 赌时间",
		Observe: "6 个交易对并发拉取，总耗时接近最慢的一个（200ms），而不是 6 个相加",
		Questions: []string{
			"把 wg.Add(1) 挪进 goroutine 内部第一行，会偶发 panic: WaitGroup is reused... 或提前 Wait —— 为什么必须在启动前 Add？",
			"删掉 defer wg.Done()，程序会怎样？（永久阻塞在 Wait，报 all goroutines are asleep）",
			"如果并发数由运行时决定（比如从数据库读交易对列表），Add 该写在哪？（提示：wg.Add(len(list)) 或每轮循环前）",
		},
		Check: "能默写出正确姿势：循环外/循环内启动前 Add，goroutine 里第一行 defer Done，主流程 Wait",
		Run: func() {
			symbols := []string{"BTC", "ETH", "SOL", "ARB", "OP", "DOGE"}
			start := time.Now()

			var wg sync.WaitGroup
			results := make([]float64, len(symbols))

			for i, s := range symbols {
				wg.Add(1) // 必须在 go 之前
				go func(i int, s string) {
					defer wg.Done() // 放进 goroutine 内部
					results[i] = price(s, 200*time.Millisecond)
				}(i, s)
			}

			wg.Wait()
			fmt.Printf("  并发拉取 %d 个价格，耗时 %v（串行需要约 1.2s）\n", len(symbols), time.Since(start).Round(time.Millisecond))
			fmt.Printf("  结果按索引回填，顺序稳定：%.2f %.2f %.2f\n", results[0], results[1], results[5])
			fmt.Println("  ↑ 关键：结果写进 results[i] 而不是 append(results, x) —— append 在并发下会 data race（L2-13 实测）")
		},
	}
}
