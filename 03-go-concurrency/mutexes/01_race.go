package mutexes

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go-study/go-concurrency/level"
)

// L2-13：data race —— 不加锁的计数、Mutex、atomic，以及 -race 怎么用。
func L13() level.Level {
	return level.Level{
		ID:      "L2-13",
		Title:   "Mutex / atomic / -race",
		Tags:    "data race · Mutex · atomic · 选型",
		Pre:     "L2-12",
		Goal:    "看见 data race 的真实后果（计数丢失），并掌握三种正确写法与检测工具",
		Observe: "无锁版本的计数小于预期值；Mutex 与 atomic 版本等于预期值",
		Questions: []string{
			"运行 go run -race . 13 —— 无锁那段会打出 WARNING: DATA RACE，读一读它指出的两个栈分别是谁",
			"把 goroutine 数从 100 提到 10000，无锁版本的丢失会更严重吗？为什么有时它也「看起来正确」？",
			"什么时候用 Mutex（保护一段逻辑/多个字段），什么时候用 atomic（单个计数器/标志位），什么时候用 channel（传递所有权）？",
			"锁的粒度：把整个循环放进一个临界区 vs 每次自增加锁，性能差多少？（提示：可先用无竞争版本对比）",
		},
		Check: "能用 -race 定位自己写出的竞态，并说出三种同步手段的选型依据",
		Run: func() {
			const workers, each = 50, 100
			want := workers * each

			var unsafeCount int // 故意不加锁
			var mu sync.Mutex
			var muCount int
			var atomicCount int64

			start := time.Now()
			var wg sync.WaitGroup
			for w := 0; w < workers; w++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for i := 0; i < each; i++ {
						unsafeCount++ // ← 竞态：读-改-写非原子
						mu.Lock()
						muCount++
						mu.Unlock()
						atomic.AddInt64(&atomicCount, 1)
					}
				}()
			}
			wg.Wait()

			fmt.Printf("  期望值        ：%d\n", want)
			fmt.Printf("  无锁（错误）  ：%d  ← 丢了 %d 次自增（每次运行都可能不同）\n", unsafeCount, want-unsafeCount)
			fmt.Printf("  Mutex（正确） ：%d\n", muCount)
			fmt.Printf("  atomic（正确）：%d\n", atomicCount)
			fmt.Printf("  耗时 %v；想看到官方警告就跑：go run -race . 13\n", time.Since(start).Round(time.Millisecond))
		},
	}
}
