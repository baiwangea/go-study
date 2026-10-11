//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 02-go-pointers && go run 15_nil_traps.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import "fmt"

/*
═══════════════════════════════════════════════════════════════════════
L1-15：nil 的真面目（四种 nil 的行为差异）
═══════════════════════════════════════════════════════════════════════

【核心概念】
    四种 nil 的行为完全不同，必须分开记：

        nil slice   → len/cap 都是 0，读安全，append 安全（会自动分配）
        nil map     → 读安全（返回零值），写 panic
        nil 指针    → 不能解引用（*p 会 panic），但可以比较、可以当参数传
        nil 接口    → 只有「类型和值都为 nil」时接口才等于 nil

    最经典的坑：接口值 = (动态类型, 动态值) 的二元组。
    把一个 nil 指针装进 error 接口后，接口的「类型」字段不是 nil，
    所以整个接口 != nil，调用方 if err != nil 会误判成「出错了」。

【你会看到】
    ✓ nil slice 直接 append 成功，自动分配底层数组
    ✓ nil map 读 m["x"] 返回 0，但写会 panic
    ✓ badDo(false) != nil 为 true  ← 坑！
    ✓ goodDo(false) == nil 为 true ← 正确写法

【思考题】（动手改代码验证，别空想）
    1. 对 nil map 执行 m["k"] = 1 报什么 panic？读 m["k"] 呢？
    2. badDo(false) 已经证明 err != nil —— 打开 badDo，用 (类型, 值) 二元结构
       解释清楚它为什么不是 nil。
    3. 本关的 bug 标准 go vet 抓不到（跑 `go vet 15_nil_traps.go` 是干净的）。
       能抓它的是 staticcheck 的 nilness 分析（SA4031 一类）。
       这说明 lint 工具选型为什么重要？

【通关标准】
    能预判任意 nil 操作的结果，
    并保证「函数返回 error 时，永远返回 nil 接口而不是 nil 指针」。

【前置】L1-14              【标签】nil map · nil slice · 接口 = (类型, 值)
═══════════════════════════════════════════════════════════════════════
*/

// tradeErr 是一个自定义错误类型，用指针接收者实现 error 接口。
type tradeErr struct{ code int }

func (e *tradeErr) Error() string { return fmt.Sprintf("下单失败 code=%d", e.code) }

// badDo 故意演示「返回 nil 指针给 error 接口」这个最经典的坑。
// 注意返回类型是 error（接口），不是 *tradeErr。
func badDo(fail bool) error {
	var e *tradeErr // e 是 *tradeErr 类型的 nil 指针
	if fail {
		e = &tradeErr{code: 500}
	}
	return e // ← 把 (*tradeErr)(nil) 装进 error 接口：接口的类型字段不为 nil！
}

// goodDo 是正确写法：要么返回具体错误，要么明确返回 nil 接口。
func goodDo(fail bool) error {
	if fail {
		return &tradeErr{code: 500}
	}
	return nil // 明确返回 nil 接口，而不是某个 nil 指针
}

func main() {
	fmt.Println("\n=== L1-15 nil 陷阱与接口 nil ===")

	// ── 示例 1：nil slice —— 读安全、append 安全 ─────────────────────────
	fmt.Println("\n【示例 1】nil slice：可以读、可以 append")
	var s []int
	fmt.Printf("  s == nil 是 %t，len=%d cap=%d\n", s == nil, len(s), cap(s))
	s = append(s, 1, 2) // append 会自动分配底层数组，不需要先 make
	fmt.Printf("  append 后 s = %v（len=%d cap=%d）\n", s, len(s), cap(s))

	// ── 示例 2：nil map —— 读安全、写 panic ──────────────────────────────
	fmt.Println("\n【示例 2】nil map：能读不能写")
	var m map[string]int
	fmt.Printf("  m == nil 是 %t\n", m == nil)
	fmt.Printf("  读 m[\"x\"] = %d（返回零值，不报错）\n", m["x"])
	fmt.Println("  写 m[\"x\"] = 1 → panic: assignment to entry in nil map（这里故意不执行）")
	fmt.Println("  正确做法：m = make(map[string]int) 或 m := map[string]int{}")

	// ── 示例 3：最经典的坑 —— 接口里装了一个 nil 指针 ────────────────────
	fmt.Println("\n【示例 3】接口值 = (类型, 值)：装了 nil 指针的接口 ≠ nil")
	fmt.Printf("  badDo(false)  == nil 吗？%t  ← 坑！本该是 nil 却判成了「有错」\n", badDo(false) == nil)
	fmt.Printf("  goodDo(false) == nil 吗？%t  ← 正确写法\n", goodDo(false) == nil)
	fmt.Printf("  badDo(true).Error() = %s\n", badDo(true).Error())

	fmt.Println("\n  原因：badDo 里返回的 e 是 (*tradeErr)(nil)，装进 error 接口后：")
	fmt.Println("        badDo(false) 的接口值 = (类型=*tradeErr, 值=nil)")
	fmt.Println("        类型字段不是 nil，所以整个接口值 != nil")
	fmt.Println("  ⚠️  标准 go vet 抓不到这个 bug，要 staticcheck 的 nilness 分析（SA4031）")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n【关键总结】")
	fmt.Println("  • nil slice：len=cap=0，读安全、append 安全")
	fmt.Println("  • nil map  ：读安全（零值），写 panic，必须先 make")
	fmt.Println("  • nil 指针 ：不能解引用，但能比较、能当参数传")
	fmt.Println("  • nil 接口 ：只有「类型和值都为空」时才 == nil")
	fmt.Println("  • 铁律：返回 error 的函数，失败就 return err，成功就 return nil")
}
