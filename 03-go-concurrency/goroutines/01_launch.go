//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 03-go-concurrency && go run goroutines/01_launch.go
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
L2-01：go 关键字与调度不确定性
═══════════════════════════════════════════════════════════════════════

【核心概念】
    go 语句是「提交任务后立即返回」：它把函数交给调度器，自己马上往下走。
    执行顺序由调度器决定，不保证任何顺序 —— 谁先跑、谁先完成都无法假设。

【你会看到】
    ✓ 编号按 4→0 倒序完成（故意用 sleep 造出来的）
    ✓ 但 [main] 那行与后台输出的相对顺序完全不保证

【常见陷阱案例】
    ❌ 错误写法 1：主协程不等待
       func main() {
           go fetchData()  // 后台协程还没执行
           // main 退出 → 整个程序立即结束，fetchData 白启动
       }
       后果：后台任务根本没机会执行，程序就结束了

    ❌ 错误写法 2：循环闭包捕获变量（Go 1.22 之前）
       for i := 0; i < 5; i++ {
           go func() { fmt.Println(i) }() // 全部打印 5！
       }
       后果：所有 goroutine 共享同一个 i 变量，循环结束时 i=5

    ❌ 错误写法 3：用 time.Sleep 等待结果
       go doWork()
       time.Sleep(1 * time.Second) // 猜测 1 秒够了
       后果：任务可能需要 2 秒，结果丢失；或者 100ms 就完成，白等 900ms

【实战场景】
    ✅ 并发拉取多个交易对行情（每个交易对起一个 goroutine）
    ✅ 并发下单到多个交易所（降低延迟）
    ✅ 后台定时任务（监控、日志上报、健康检查）

【思考题】（动手改代码验证，别空想）
    1. 把循环里的 go func(){...}() 改成直接调用函数，顺序变成什么？
       为什么并发时不是这样？
    2. 删掉结尾的 time.Sleep，主协程会立刻退出 —— 后台 goroutine 的输出还在吗？
       （这就是「不等」的坑）
    3. 注意 i := i 这一行：Go 1.22 之前不写它会打印什么？为什么？
    4. 真实场景：如果要并发拉取 100 个交易对行情，该怎么写？
       （提示：不能直接起 100 个 goroutine，需要限流 —— 见 L2-12 worker pool）

【通关标准】
    能说清 go 语句的语义，并解释为什么「靠 sleep 等并发结果」在生产里是错的

【前置】L1-11              【标签】goroutine · 调度 · 不要假设顺序
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L2-01 启动 goroutine")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// ── 示例 1：go 语句提交任务后立刻返回 ───────────────────────────────
	fmt.Println("📝 示例 1：go 语句不会等任何人")
	fmt.Println(strings.Repeat("-", 60))
	var wg sync.WaitGroup // 这里先借用一下，L2-02 正式讲等待
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("  [goroutine] 我在后台跑")
	}()
	// 关键：go 语句执行到这里时，上面的 goroutine 可能还没开始执行
	// 调度器只是记录了「有一个任务待执行」，具体什么时候执行完全不确定

	// ── 示例 2：启动顺序 ≠ 完成顺序 ─────────────────────────────────────
	fmt.Println("\n📌 示例 2：完成顺序由调度器决定（这里用 sleep 人为造出倒序）")
	fmt.Println(strings.Repeat("-", 60))
	for i := 0; i < 5; i++ {
		// ⚠️ 关键行：Go 1.22+ 循环变量每轮自动新建；旧版本必须手动 i := i
		// 不写这行会怎样？旧版本所有 goroutine 共享同一个 i 变量，
		// 循环结束时 i=5，所有 goroutine 都打印 5！
		i := i

		go func() {
			// 故意让编号越小的任务 sleep 越久，制造「倒序完成」效果
			// 真实场景：任务耗时不可控，完成顺序更不可控
			time.Sleep(time.Duration(5-i) * 10 * time.Millisecond)
			fmt.Printf("  [goroutine] 编号 %d 完成\n", i)
		}()
	}
	fmt.Println("  [main] 我已经走到这里了 —— go 语句不会等任何人")

	wg.Wait()
	time.Sleep(120 * time.Millisecond) // 借用等待收干净输出，下一关换成正规写法

	// ── 示例 3：实战场景 —— 并发拉取多个交易对行情 ───────────────────────
	fmt.Println("\n🔥 示例 3：真实场景 —— 并发拉取行情（模拟）")
	fmt.Println(strings.Repeat("-", 60))
	symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT"}
	var wg2 sync.WaitGroup

	for _, symbol := range symbols {
		symbol := symbol // 循环变量捕获（Go 1.22+ 可省略，但写上兼容性更好）
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			// 模拟 HTTP 请求延迟（实际是网络 I/O）
			time.Sleep(time.Duration(30+len(symbol)*2) * time.Millisecond)
			fmt.Printf("  ✓ %s 行情已更新\n", symbol)
		}()
	}

	wg2.Wait() // 等待所有行情拉取完成
	fmt.Println("  → 所有行情拉取完成，可以开始计算策略")

	// ── 反面教材：错误写法演示 ─────────────────────────────────────────
	fmt.Println("\n❌ 反面教材：不要这么写")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("  // 错误写法 1：主协程不等待")
	fmt.Println("  func main() {")
	fmt.Println("      go doWork()  // 后台任务还没执行")
	fmt.Println("  }  // main 退出 → 程序结束，doWork 白启动")
	fmt.Println()
	fmt.Println("  // 错误写法 2：用 sleep 猜测时间")
	fmt.Println("  go doWork()")
	fmt.Println("  time.Sleep(1 * time.Second) // 猜测 1 秒够？")
	fmt.Println("  // 如果任务需要 2 秒 → 结果丢失")
	fmt.Println("  // 如果任务只需 100ms → 白等 900ms")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ go 语句 = 提交任务后立即返回，不等待、不保证顺序")
	fmt.Println("  ✅ 循环里启动 goroutine 要小心闭包捕获：旧版本必须写 i := i")
	fmt.Println("  ✅ 靠 time.Sleep 等结果只是「赌时间」，生产里必须用 WaitGroup / channel")
	fmt.Println("  ✅ 主协程退出 = 整个程序结束，后台 goroutine 会被直接抛弃")
	fmt.Println()
	fmt.Println("  📌 实战建议：")
	fmt.Println("    • 并发拉取行情：用 WaitGroup 等待所有完成（见 L2-02）")
	fmt.Println("    • 需要限流：用 worker pool 控制并发数（见 L2-12）")
	fmt.Println("    • 需要超时取消：用 context.WithTimeout（见 L2-08, L2-11）")
}
