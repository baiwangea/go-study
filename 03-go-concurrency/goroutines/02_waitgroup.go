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
	"strings"
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

【常见陷阱案例】
    ❌ 错误写法 1：Add 写在 goroutine 里面
       for i := 0; i < 10; i++ {
           go func() {
               wg.Add(1)        // ⚠️ 太晚了！
               defer wg.Done()
               doWork()
           }()
       }
       wg.Wait()  // 可能立即返回，因为 Add 还没执行
       后果：Wait 看到计数器为 0，直接通过，后台任务白启动

    ❌ 错误写法 2：忘记 Done
       wg.Add(1)
       go func() {
           doWork()  // 忘了 wg.Done()
       }()
       wg.Wait()  // 永远阻塞，报 deadlock
       后果：计数器永远不归零，程序卡死

    ❌ 错误写法 3：并发 append 到切片
       var results []float64
       for _, s := range symbols {
           wg.Add(1)
           go func(s string) {
               defer wg.Done()
               results = append(results, fetchPrice(s))  // ⚠️ data race
           }(s)
       }
       后果：多个 goroutine 同时操作切片，导致数据竞争或 panic

【实战场景】
    ✅ 并发拉取多个交易所的同一交易对价格（套利机器人）
    ✅ 并发发送多笔订单（批量下单）
    ✅ 并发执行多个无依赖的初始化任务（启动时）

【思考题】（动手改代码验证，别空想）
    1. 把 wg.Add(1) 挪进 goroutine 内部第一行，会偶发 panic: WaitGroup is reused...
       或提前 Wait —— 为什么必须在启动前 Add？
    2. 删掉 defer wg.Done()，程序会怎样？
       （永久阻塞在 Wait，报 all goroutines are asleep）
    3. 如果并发数由运行时决定（比如从数据库读交易对列表），Add 该写在哪？
       （提示：wg.Add(len(list)) 或每轮循环前）
    4. 为什么要用 defer wg.Done() 而不是直接写 wg.Done()？
       （提示：如果中间 panic，Done 不执行，计数器永远不归零）

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
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L2-02 WaitGroup")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// ── 示例 1：并发拉取多个交易对 ──────────────────────────────────────
	fmt.Println("📝 示例 1：6 个交易对并发拉取，等最慢的那个就行")
	fmt.Println(strings.Repeat("-", 60))
	symbols := []string{"BTC", "ETH", "SOL", "ARB", "OP", "DOGE"}
	start := time.Now()

	var wg sync.WaitGroup
	results := make([]float64, len(symbols)) // ⚠️ 预分配空间，按索引写入

	for i, s := range symbols {
		// ✅ 关键：Add 必须在 go 之前调用
		// 为什么？因为 Wait() 可能在任何 goroutine 开始前就执行
		// 如果 Add 在 goroutine 里，Wait 看到计数 0，立即返回，goroutine 白启动
		wg.Add(1)

		go func(i int, s string) {
			// ✅ 第一行写 defer Done，保证异常路径也能减计数
			// 如果不用 defer，中间 panic 了，Done 不执行，Wait 永远阻塞
			defer wg.Done()

			// 模拟行情请求耗时 200ms
			results[i] = price(s, 200*time.Millisecond)

			// ⚠️ 为什么要 results[i] 而不是 append(results, x)？
			// append 不是原子操作，多个 goroutine 同时 append 会 data race
			// 详见 L2-13 data race 实测
		}(i, s)
	}

	// ✅ Wait 阻塞到所有 goroutine 都执行完 Done
	wg.Wait()
	fmt.Printf("  并发拉取 %d 个价格，耗时 %v（串行需要约 1.2s）\n", len(symbols), time.Since(start).Round(time.Millisecond))
	fmt.Printf("  结果按索引回填，顺序稳定：%.2f %.2f %.2f\n", results[0], results[1], results[5])
	fmt.Println("  ➡️  关键：结果写进 results[i] 而不是 append(results, x) —— append 在并发下会 data race（L2-13 实测）")

	// ── 示例 2：实战场景 —— 套利机器人并发拉取多个交易所价格 ───────────────
	fmt.Println("\n🔥 示例 2：真实场景 —— 套利机器人（模拟）")
	fmt.Println(strings.Repeat("-", 60))

	type ExchangePrice struct {
		Exchange string
		Price    float64
		Err      error
	}

	exchanges := []string{"Binance", "OKX", "Bybit"}
	prices := make([]ExchangePrice, len(exchanges))
	var wg2 sync.WaitGroup

	for i, ex := range exchanges {
		wg2.Add(1)
		go func(i int, ex string) {
			defer wg2.Done()
			// 模拟 HTTP 请求延迟不同
			time.Sleep(time.Duration(50+i*30) * time.Millisecond)
			prices[i] = ExchangePrice{
				Exchange: ex,
				Price:    10000.0 + float64(i*10), // 模拟价差
			}
		}(i, ex)
	}

	wg2.Wait()
	fmt.Println("  价格对比（寻找套利机会）：")
	for _, p := range prices {
		fmt.Printf("    %s: $%.2f\n", p.Exchange, p.Price)
	}
	fmt.Println("  → Binance 和 Bybit 有 $20 价差，可以套利！")

	// ── 反面教材：错误写法演示 ─────────────────────────────────────────
	fmt.Println("\n❌ 反面教材：不要这么写")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("  // 错误写法 1：Add 在 goroutine 里")
	fmt.Println("  go func() {")
	fmt.Println("      wg.Add(1)  // ⚠️ 太晚了！Wait 可能已经返回")
	fmt.Println("      defer wg.Done()")
	fmt.Println("      doWork()")
	fmt.Println("  }()")
	fmt.Println()
	fmt.Println("  // 错误写法 2：忘记 Done")
	fmt.Println("  wg.Add(1)")
	fmt.Println("  go func() {")
	fmt.Println("      doWork()  // 忘了 wg.Done()")
	fmt.Println("  }()")
	fmt.Println("  wg.Wait()  // 永远阻塞，报 deadlock")
	fmt.Println()
	fmt.Println("  // 错误写法 3：并发 append")
	fmt.Println("  var results []float64")
	fmt.Println("  for _, s := range symbols {")
	fmt.Println("      go func(s string) {")
	fmt.Println("          results = append(results, fetch(s))  // ⚠️ data race")
	fmt.Println("      }(s)")
	fmt.Println("  }")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ Add 必须在 go 之前调用（否则 Wait 可能提前返回）")
	fmt.Println("  ✅ goroutine 第一行写 defer wg.Done()，保证异常路径也能减计数")
	fmt.Println("  ✅ 主流程 wg.Wait() 阻塞到计数归零，代替「靠 sleep 赌时间」")
	fmt.Println("  ✅ 并发写共享切片要按索引回填，append 不是并发安全的")
	fmt.Println()
	fmt.Println("  📌 实战建议：")
	fmt.Println("    • 套利机器人：并发拉取多个交易所价格，取最快返回的")
	fmt.Println("    • 批量下单：并发发送订单，等全部确认再继续")
	fmt.Println("    • 启动初始化：并发加载配置、连接数据库、订阅行情")
}
