//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 03-go-concurrency && go run channels/03_close_range.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"strings"
	"time"
)

/*
═══════════════════════════════════════════════════════════════════════
L2-05：close / range / 逗号 ok
═══════════════════════════════════════════════════════════════════════

【核心概念】
    用 close 广播「没有更多数据了」，接收端用 for range 自动收尾。
    通道关闭后仍可读完剩余数据，但再写入会 panic。

【你会看到】
    ✓ range 打印完 3 条后自动结束；已关闭通道读出的 x, ok 里 ok=false

【思考题】（动手改代码验证，别空想）
    1. 把 close(ch) 删掉：range 取完 3 条后就会永久阻塞，最终报
       fatal error: all goroutines are asleep - deadlock! ——
       为什么这种「忘关闭」比一般 bug 难发现？
    2. 读取的两种写法：v := <-ch 与 v, ok := <-ch，通道关闭后分别得到什么？
       为什么只看零值不可信？
    3. 发送方和接收方谁该负责 close？
       （提示：只有发送方关；两个发送者都关会 panic: send on closed channel
       —— 怎么做单发送者？）

【通关标准】
    能说清「通道关闭后仍可读剩余数据，但再写入会 panic」，
    并解释为什么机器人里禁止接收方 close

【前置】L2-04              【标签】close · range · 结束信号
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L2-05 close 与 range")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// ── 示例 1：range 自动收尾 ──────────────────────────────────────────
	fmt.Println("📝 示例 1：close 之后 range 自动结束")
	fmt.Println(strings.Repeat("-", 60))
	ch := make(chan int, 3)
	for i := 1; i <= 3; i++ {
		ch <- i * 100
	}
	close(ch) // 生产完毕，通知接收方

	for v := range ch { // 通道关闭且取空后自动退出循环
		fmt.Printf("  收到价格 tick：%d\n", v)
	}

	// ── 示例 2：逗号 ok 区分「零值」和「已关闭」 ─────────────────────────
	fmt.Println("📌 示例 2：逗号 ok：判断通道是否已关闭")
	fmt.Println(strings.Repeat("-", 60))
	v, ok := <-ch // 已关闭且已排空
	fmt.Printf("  再读一次：v=%d ok=%v（ok=false 表示通道已关闭，零值不可信）\n", v, ok)

	time.Sleep(10 * time.Millisecond)
	fmt.Println("  ➡️  机器人里常用套路：行情源关闭 → 所有订阅者的 range 自动结束 → 优雅停机")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ close(ch) 是广播「不会再有数据」，不是清空通道")
	fmt.Println("  ✅ for range ch 在通道关闭且取空后自动退出，是标准的收尾写法")
	fmt.Println("  ✅ 关闭后再读得到零值，靠 v, ok := <-ch 的 ok 判断真假")
	fmt.Println("  ✅ 只有发送方有权 close；对已关闭通道再发送会 panic")
}
