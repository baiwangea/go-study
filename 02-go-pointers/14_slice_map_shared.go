//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 02-go-pointers && go run 14_slice_map_shared.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"strings"
)

/*
═══════════════════════════════════════════════════════════════════════
L1-14：切片的共享与失联
═══════════════════════════════════════════════════════════════════════

【核心概念】
    slice = (指向底层数组的指针, len, cap) 三元组。
    赋值只拷贝这个三元组，底层数组仍然是共享的 —— 所以改一个会动到另一个。
    append 一旦触发扩容，会申请新数组并拷贝旧数据，此后两个切片彻底失联。
    map 本身就是引用语义（内部是指针），赋值不拷贝内容。

【你会看到】
    ✓ s2[0] = 100 之后 s1[0] 也变成 100        —— 共享底层数组
    ✓ append 未扩容时写进了 s1 的空位，s1 被改脏 —— 失联前的共享
    ✓ 扩容后 cap 3 → 8，改 big[0] 不再影响 s1   —— 已失联
    ✓ 改 m2 后 m1 同步变化                      —— map 是引用语义

【思考题】（动手改代码验证，别空想）
    1. s3 := append(s2, 999) 之后，s1 的第三个元素变成了什么？为什么？
    2. 打印每次 append 前后的 cap，找出「从哪一次开始换新数组」。
    3. 把函数参数写成 func fix(s []int) { s[0] = 0 }，调用后原切片变吗？
       如果函数体里写的是 s = append(s, 1) 呢？—— 这就是「切片当参数」的经典歧义。
    4. PHP 数组按值传递、JS 数组按引用，Go 的 slice 为什么两者都像、又都不完全是？

【通关标准】
    能用三元组模型解释任何切片共享/失联现象，
    并知道 nil map 必须 make 之后才能写。

【前置】L1-12              【标签】slice header · append 扩容 · map 引用语义
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L1-14 切片的共享与失联")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// ── 示例 1：切片赋值只拷贝「三元组」，底层数组共享 ───────────────────
	fmt.Println("📝 示例 1：切片头是值拷贝，底层数组共享")
	fmt.Println(strings.Repeat("-", 60))
	s1 := []int{1, 2, 3}
	s2 := s1[:2] // s2 是 s1 前两个元素的视图，共用同一块底层数组
	s2[0] = 100  // 改 s2 就是在改底层数组
	fmt.Printf("  s2[0]=100 后，s1 = %v（len=%d cap=%d）← s1 被改了！\n", s1, len(s1), cap(s1))

	// ── 示例 2：append 没扩容时会占用原数组的空位 ────────────────────────
	fmt.Println("📌 示例 2：append 未扩容时，直接写进同一块底层数组")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("  s2 的 len=%d cap=%d —— 还有 1 个空位\n", len(s2), cap(s2))
	s3 := append(s2, 999)
	fmt.Printf("  append(s2, 999) 后，s3 = %v\n", s3)
	fmt.Printf("  而 s1 = %v ← 第三个元素被 s3 覆盖了\n", s1)
	fmt.Println("  ⚠️  这就是「共享」：只要没扩容，两个切片就是同一块内存的两个视图")

	// ── 示例 3：扩容后换新数组，两个切片彻底失联 ─────────────────────────
	fmt.Println("🔍 示例 3：容量不足触发扩容 → 换新数组 → 失联")
	fmt.Println(strings.Repeat("-", 60))
	big := append(s3, 1, 2, 3, 4, 5) // 需要 len=8 > cap=3，触发扩容
	fmt.Printf("  cap 从 %d 涨到 %d（扩容 = 申请新数组 + 把旧数据拷过去）\n", cap(s3), cap(big))
	big[0] = -1
	fmt.Printf("  改 big[0]=-1 后，big = %v\n", big)
	fmt.Printf("  而 s1 = %v ← 已经失联，再也改不动了\n", s1)

	// ── 示例 4：map 本身就是引用语义 ─────────────────────────────────────
	fmt.Println("🧪 示例 4：map 是引用语义：赋值不拷贝内容")
	fmt.Println(strings.Repeat("-", 60))
	m1 := map[string]int{"BTC": 1}
	m2 := m1 // m2 和 m1 指向同一个 map，没有拷贝任何键值
	m2["BTC"] = 2
	fmt.Printf("  改 m2 后，m1 = %v ← 同步变化\n", m1)
	fmt.Println("  ⚠️  但 nil map 只能读不能写：var m map[string]int; m[\"x\"]=1 → panic（见 L1-15）")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ slice = (指针, len, cap)，赋值只拷贝三元组，底层数组共享")
	fmt.Println("  ✅ append 未扩容 → 写进原数组（可能污染别的切片）")
	fmt.Println("  ✅ append 已扩容 → 换新数组，从此失联")
	fmt.Println("  ✅ 想让调用方拿到 append 结果，必须 return 新切片，或传 *[]int")
	fmt.Println("  ✅ map 是引用语义，但必须 make 之后才能写")
}
