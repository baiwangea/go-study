//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 03-go-concurrency && go run goroutines/02_waitgroup.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"sync"
	"time"
)

/*
═══════════════════════════════════════════════════════════════════════
L2-02：WaitGroup 三点规矩
═══════════════════════════════════════════════════════════════════════

【核心概念】
    WaitGroup 用来等一组 goroutine 全部干完。
    三条规矩：启动前 Add、goroutine 里第一行 defer Done、主流程 Wait。

【你会看到】
    ✓ 6 个交易对并发拉取，总耗时接近最慢的一个（200ms），而不是 6 个相加

【思考题】（动手改代码验证，别空想）
    1. 把 wg.Add(1) 挪进 goroutine 内部第一行，会偶发 panic: WaitGroup is reused...
       或提前 Wait —— 为什么必须在启动前 Add？
    2. 删掉 defer wg.Done()，程序会怎样？
       （永久阻塞在 Wait，报 all goroutines are asleep）
    3. 如果并发数由运行时决定（比如从数据库读交易对列表），Add 该写在哪？
       （提示：wg.Add(len(list)) 或每轮循环前）

【通关标准】
    能默写出正确姿势：循环外/循环内启动前 Add，goroutine 里第一行 defer Done，
    主流程 Wait

【前置】L2-01              【标签】WaitGroup · Add/Done/Wait
═══════════════════════════════════════════════════════════════════════
*/

// price 模拟一次行情请求：睡 delay 后返回一个和交易对名字长度相关的价格。
func price(symbol string, delay time.Duration) float64 {
	time.Sleep(delay) // 模拟一次行情请求的耗时
	return float64(len(symbol)) * 3.14
}

func main() {
	fmt.Println("\n=== L2-02 WaitGroup ===")

	// ── 示例 1：并发拉取多个交易对 ──────────────────────────────────────
	fmt.Println("\n【示例 1】6 个交易对并发拉取，等最慢的那个就行")
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

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n【关键总结】")
	fmt.Println("  • Add 必须在 go 之前调用（否则 Wait 可能提前返回）")
	fmt.Println("  • goroutine 第一行写 defer wg.Done()，保证异常路径也能减计数")
	fmt.Println("  • 主流程 wg.Wait() 阻塞到计数归零，代替「靠 sleep 赌时间」")
	fmt.Println("  • 并发写共享切片要按索引回填，append 不是并发安全的")
}
