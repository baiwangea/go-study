//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 02-go-pointers && go run 12_struct_copy.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"strings"
)

/*
═══════════════════════════════════════════════════════════════════════
L1-12：结构体的拷贝 vs 指针
═══════════════════════════════════════════════════════════════════════

【核心概念】
    结构体是值类型：赋值和传参都会「整个拷贝所有字段」。
    需要就地修改、或结构体很大（拷贝开销高）时，改用指针。

【你会看到】
    ✓ resizeCopy(pos) 后 Size 仍是 0.5 —— 改的是副本
    ✓ resizePtr(&pos) 后 Size 变成 999 —— 通过指针改了原件
    ✓ copy2 := pos 之后改 copy2，pos 不受影响 —— 赋值也是拷贝
    ✓ ptr.Label() 能编译 —— Go 自动解引用成 (*ptr).Label()

【思考题】（动手改代码验证，别空想）
    1. 用 %p 打印 pos 和函数内参数的地址，观察「拷贝」发生在哪一步。
    2. ptr.Label() 为什么能编译？Go 在背后帮你做了什么？
    3. 什么时候用值、什么时候用指针？把下面三条对号入座：
         · 要修改原件        → ?
         · 结构体很大        → ?
         · 需要表示「不存在」 → ?

【通关标准】
    能一眼看出一个函数「会不会改到我的结构体」，
    并说清 ptr.Method() 自动解引用的规则。

【前置】L1-11              【标签】结构体拷贝 · 指针接收 · 自动解引用
═══════════════════════════════════════════════════════════════════════
*/

// Position 是交易持仓结构，本关的演示对象。
type Position struct {
	Symbol string  // 交易对
	Size   float64 // 仓位大小
	Entry  float64 // 入场价格
}

// Label 是值接收者方法：只读，不会改到原对象。
func (p Position) Label() string {
	return fmt.Sprintf("%s %.2f", p.Symbol, p.Size)
}

// resizeCopy 值传递：参数是整个结构体的副本，改它不影响原件。
func resizeCopy(p Position) {
	fmt.Printf("    resizeCopy 收到参数的地址：%p（和外面那个不是同一个）\n", &p)
	p.Size = 999
	fmt.Println("    resizeCopy 内部：Size =", p.Size, "（改的是副本）")
}

// resizePtr 指针传递：参数是地址，改的是原件。
func resizePtr(p *Position) {
	fmt.Printf("    resizePtr 收到参数的地址：%p（就是外面那个）\n", p)
	p.Size = 999
	fmt.Println("    resizePtr 内部：Size =", p.Size, "（改到了原件）")
}

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L1-12: 结构体的拷贝 vs 指针")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// ── 示例 1：值传递会拷贝整个结构体 ───────────────────────────────────
	fmt.Println("📝 示例 1：值传递 - 整个结构体被拷贝一份")
	fmt.Println(strings.Repeat("-", 60))
	pos := Position{Symbol: "BTC", Size: 0.5, Entry: 63000}
	fmt.Printf("  原始 pos 的地址：%p\n", &pos)
	fmt.Printf("  原始值：%+v\n", pos)

	resizeCopy(pos)
	fmt.Printf("  调用 resizeCopy(pos) 后：Size = %.2f（没变），Label() = %s\n", pos.Size, pos.Label())

	// ── 示例 2：指针传递才能改到原件 ─────────────────────────────────────
	fmt.Println("\n📌 示例 2：指针传递 - 传地址，可以改原件")
	fmt.Println(strings.Repeat("-", 60))
	resizePtr(&pos)
	fmt.Printf("  调用 resizePtr(&pos) 后：Size = %.2f（变了）\n", pos.Size)

	// ── 示例 3：赋值运算符同样是拷贝 ─────────────────────────────────────
	fmt.Println("\n🔄 示例 3：赋值也是拷贝 - 两个结构体从此互不相干")
	fmt.Println(strings.Repeat("-", 60))
	copy2 := pos // 整体拷贝所有字段
	copy2.Symbol = "ETH"
	copy2.Size = 10
	fmt.Printf("    改 copy2 后 → copy2 = %+v\n", copy2)
	fmt.Printf("    而 pos        = %+v（完全没受影响）\n", pos)

	// ── 示例 4：Go 的语法糖 —— 自动解引用 ────────────────────────────────
	fmt.Println("\n🍬 示例 4：指针调方法会自动解引用")
	fmt.Println(strings.Repeat("-", 60))
	ptr := &pos
	fmt.Println("    ptr.Label() =", ptr.Label(), "（等价于 (*ptr).Label()，Go 自动帮你解引用）")
	fmt.Println("    pos.Label() =", pos.Label())

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ 结构体赋值、传参、返回，都是整体拷贝")
	fmt.Println("  ✅ 要修改原件 → 用指针；结构体很大 → 用指针（省拷贝开销）")
	fmt.Println("  ✅ 需要表达「不存在 / 未设置」→ 用指针（可以是 nil）")
	fmt.Println("  ✅ ptr.Method() 自动解引用为 (*ptr).Method()，是编译器给的糖")
	fmt.Println(strings.Repeat("=", 60) + "\n")
}
