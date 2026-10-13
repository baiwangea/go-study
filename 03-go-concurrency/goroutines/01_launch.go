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

【思考题】（动手改代码验证，别空想）
    1. 把循环里的 go func(){...}() 改成直接调用函数，顺序变成什么？
       为什么并发时不是这样？
    2. 删掉结尾的 time.Sleep，主协程会立刻退出 —— 后台 goroutine 的输出还在吗？
       （这就是「不等」的坑）
    3. 注意 i := i 这一行：Go 1.22 之前不写它会打印什么？为什么？

【通关标准】
    能说清 go 语句的语义，并解释为什么「靠 sleep 等并发结果」在生产里是错的

【前置】L1-04（闭包）      【标签】goroutine · 调度 · 不要假设顺序
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

	// ── 示例 2：启动顺序 ≠ 完成顺序 ─────────────────────────────────────
	fmt.Println("📌 示例 2：完成顺序由调度器决定（这里用 sleep 人为造出倒序）")
	fmt.Println(strings.Repeat("-", 60))
	for i := 0; i < 5; i++ {
		i := i // Go 1.22 起循环变量每轮新建；旧版本必须手动复制，否则闭包共享同一个 i
		go func() {
			time.Sleep(time.Duration(5-i) * 10 * time.Millisecond) // 故意让完成顺序倒过来
			fmt.Printf("  [goroutine] 编号 %d 完成\n", i)
		}()
	}
	fmt.Println("  [main] 我已经走到这里了 —— go 语句不会等任何人")

	wg.Wait()
	time.Sleep(120 * time.Millisecond) // 借用等待收干净输出，下一关换成正规写法

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ go 语句 = 提交任务后立即返回，不等待、不保证顺序")
	fmt.Println("  ✅ 循环里启动 goroutine 要小心闭包捕获：旧版本必须写 i := i")
	fmt.Println("  ✅ 靠 time.Sleep 等结果只是「赌时间」，生产里必须用 WaitGroup / channel")
	fmt.Println("  ✅ 主协程退出 = 整个程序结束，后台 goroutine 会被直接抛弃")
}
