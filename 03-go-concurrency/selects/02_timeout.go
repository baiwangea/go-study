//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 03-go-concurrency && go run selects/02_timeout.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

/*
═══════════════════════════════════════════════════════════════════════
L2-08：select + time.After 超时
═══════════════════════════════════════════════════════════════════════

【核心概念】
    给任何外部调用套上超时上限，宁可失败也不要无限等待。
    超时返回 ≠ 后端没执行 —— 这是最容易踩的坑。

【你会看到】
    ✓ 接口耗时 100ms < 超时 200ms → 成功
    ✓ 接口耗时 500ms > 超时 200ms → 立刻返回错误

【思考题】（动手改代码验证，别空想）
    1. 超时触发后，那个仍在 sleep 的 goroutine 怎么样了？
       （泄漏！这就是为什么真实项目要用 context.WithTimeout，见 L2-11）
    2. time.After 在 for 循环里反复调用会有什么副作用？
       （每次都会新建 timer，未触发前不会被 GC）
    3. 机器人下单超时后应该立即重试，还是先查订单状态？为什么？（防止重复下单）

【通关标准】
    能写出带超时的调用封装，并说清「超时返回 ≠ 后端没执行」这个坑

【前置】L2-07              【标签】timeout · 快速失败
═══════════════════════════════════════════════════════════════════════
*/

// slowAPI 模拟一个响应不确定的交易所接口：
// 在 timeout 内等到 latency 结束就算成功，否则返回超时错误。
func slowAPI(timeout time.Duration, latency time.Duration) error {
	ch := make(chan struct{})
	go func() {
		time.Sleep(latency)
		close(ch)
	}()

	select {
	case <-ch:
		return nil
	case <-time.After(timeout):
		return errors.New("context deadline exceeded（请求超时）")
	}
}

// mustLabel 把 error 转成一行中文说明，方便打印。
func mustLabel(err error) string {
	if err != nil {
		return "失败：" + err.Error()
	}
	return "成功：拿到响应"
}

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L2-08 超时控制")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// ── 示例 1：耗时小于超时 → 成功 ─────────────────────────────────────
	// ── 示例 2：耗时大于超时 → 快速失败 ─────────────────────────────────
	fmt.Println("📝 示例 1、2：快就成功，慢就立刻失败")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("  情况A：", mustLabel(slowAPI(200*time.Millisecond, 100*time.Millisecond)), "（接口 100ms，超时 200ms）")
	fmt.Println("  情况B：", mustLabel(slowAPI(200*time.Millisecond, 500*time.Millisecond)), "（接口 500ms，超时 200ms）")
	fmt.Println("  ➡️  情况B 只花了约 200ms 就失败返回，而不是等满 500ms —— 快速失败比死等安全")

	time.Sleep(600 * time.Millisecond) // 让两个后台 goroutine 收尾，避免污染下一关输出

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ select + time.After 是给任意调用加超时的通用写法")
	fmt.Println("  ✅ 超时只是「我这边不等了」，后端可能仍在执行 —— 别盲目重试")
	fmt.Println("  ✅ 超时后后台 goroutine 会泄漏，真实项目用 context.WithTimeout 收尾")
	fmt.Println("  ✅ time.After 放在 for 循环里会不断新建 timer，注意内存")
}
