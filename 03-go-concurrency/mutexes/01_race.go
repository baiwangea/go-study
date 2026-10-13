//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 03-go-concurrency && go run mutexes/01_race.go
//       想看到官方 DATA RACE 报告，加 -race：
//       cd 03-go-concurrency && go run -race mutexes/01_race.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

/*
═══════════════════════════════════════════════════════════════════════
L2-13：Mutex / atomic / -race
═══════════════════════════════════════════════════════════════════════

【核心概念】
    看见 data race 的真实后果（计数丢失），并掌握三种正确写法与检测工具。
    Mutex 保护一段逻辑/多个字段；atomic 保护单个计数器/标志位；
    channel 用于传递所有权。

【你会看到】
    ✓ 无锁版本的计数小于预期值
    ✓ Mutex 与 atomic 版本等于预期值

【思考题】（动手改代码验证，别空想）
    1. 运行 go run -race mutexes/01_race.go ——
       无锁那段会打出 WARNING: DATA RACE，读一读它指出的两个栈分别是谁
    2. 把 goroutine 数从 100 提到 10000，无锁版本的丢失会更严重吗？
       为什么有时它也「看起来正确」？
    3. 什么时候用 Mutex（保护一段逻辑/多个字段），
       什么时候用 atomic（单个计数器/标志位），什么时候用 channel（传递所有权）？
    4. 锁的粒度：把整个循环放进一个临界区 vs 每次自增加锁，性能差多少？
       （提示：可先用无竞争版本对比）

【通关标准】
    能用 -race 定位自己写出的竞态，并说出三种同步手段的选型依据

【前置】L2-12              【标签】data race · Mutex · atomic · 选型
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L2-13 竞态与 Mutex/atomic")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// ── 示例 1：同一份计数，三种写法对比 ────────────────────────────────
	fmt.Println("📝 示例 1：50 个 goroutine 各自增 100 次，共 5000 次")
	fmt.Println(strings.Repeat("-", 60))
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
	fmt.Printf("  耗时 %v；想看到官方警告就跑：go run -race mutexes/01_race.go\n", time.Since(start).Round(time.Millisecond))

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ unsafeCount++ 是「读-改-写」三步，并发下会互相覆盖 → 丢计数")
	fmt.Println("  ✅ Mutex 保护临界区，适合一段逻辑或多个字段的一致性")
	fmt.Println("  ✅ atomic 适合单个计数器/标志位，比锁更轻")
	fmt.Println("  ✅ channel 用于在 goroutine 间传递所有权，而不是共享内存")
	fmt.Println("  ✅ go run -race 能抓到竞态，CI 里应该常开")
}
