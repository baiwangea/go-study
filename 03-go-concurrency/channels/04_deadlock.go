//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 03-go-concurrency && go run channels/04_deadlock.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"time"
)

/*
═══════════════════════════════════════════════════════════════════════
L2-06：死锁 / 泄漏 / nil channel
═══════════════════════════════════════════════════════════════════════

【核心概念】
    三种最常见的卡死形态：无人接收的发送、无人发送的接收、nil channel。
    本关用「超时探测」代替「让它挂着」，安全地观察它们。

【你会看到】
    ✓ 三种情况都在 150ms 内被判定为「阻塞」，程序不会挂住
    ✓ 泄漏的 goroutine 依然存在（内存里收不回来）

【思考题】（动手改代码验证，别空想）
    1. 把上面的超时探测全部删掉，只留阻塞语句 ——
       报错是 fatal error: all goroutines are asleep - deadlock!，
       为什么它不可 recover？
    2. 本关留下的 3 个阻塞 goroutine 什么时候释放内存？
       （这就是泄漏，L2-11 用 context 才能真正收回来）
    3. nil channel（var ch chan int）收发都永久阻塞 ——
       所以 select 里可以用它「临时禁用某个分支」，怎么构造这个用法？

【通关标准】
    能说出无人接收的发送、无人发送的接收、nil channel 三种阻塞的成因，
    以及 fatal 与 panic 的区别

【前置】L2-05              【标签】deadlock · goroutine leak · nil chan
═══════════════════════════════════════════════════════════════════════
*/

// wait 制造一次「只有一端」的收发：send=true 时只发不收，否则只收不发，
// 两种情况都会让对端永久阻塞。用超时探测避免程序真的挂死。
func wait(who string, ch chan int, send bool) {
	if send {
		go func() { ch <- 42 }() // 发送方将永久阻塞
		time.Sleep(150 * time.Millisecond)
		fmt.Printf("  %s：无人接收 → 发送方 goroutine 永久阻塞（泄漏，仍在内存里）\n", who)
		return
	}
	got := make(chan bool, 1)
	go func() { <-ch; got <- true }() // 接收方将永久阻塞
	select {
	case <-got:
		fmt.Printf("  %s：竟然收到了\n", who)
	case <-time.After(150 * time.Millisecond):
		fmt.Printf("  %s：无人发送 → 接收方永久阻塞\n", who)
	}
}

func main() {
	fmt.Println("\n=== L2-06 死锁 ===")

	// ── 示例 1、2：只有发送 / 只有接收 ──────────────────────────────────
	fmt.Println("\n【示例 1、2】只有一端在场 → 另一端永久阻塞")
	wait("情况一", make(chan int), true)  // 只有发送
	wait("情况二", make(chan int), false) // 只有接收

	// ── 示例 3：nil channel 收发都阻塞 ──────────────────────────────────
	fmt.Println("\n【示例 3】nil channel 收发都永久阻塞")
	var nilCh chan int // nil channel
	done := make(chan string)
	go func() { <-nilCh; done <- "收到" }()
	select {
	case msg := <-done:
		fmt.Println("  情况三：", msg)
	case <-time.After(150 * time.Millisecond):
		fmt.Println("  情况三：nil channel 收发都永久阻塞 —— 可用来禁用 select 分支")
	}

	fmt.Println("  ↓ 若把上面三处超时删掉，程序会立刻 fatal error（deadlock），且无法 recover")
	time.Sleep(10 * time.Millisecond)

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n【关键总结】")
	fmt.Println("  • 无人接收的发送、无人发送的接收、nil channel 收发 —— 三种永久阻塞")
	fmt.Println("  • fatal error: all goroutines are asleep 不可 recover，是运行时死锁检测")
	fmt.Println("  • panic 可被 recover，fatal error 不能 —— 别指望兜住死锁")
	fmt.Println("  • 永久阻塞的 goroutine = 泄漏，要靠 context 取消（L2-11）才能收回")
}
