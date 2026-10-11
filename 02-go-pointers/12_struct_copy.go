package main

import "fmt"

/*
═══════════════════════════════════════════════════════════════
L1-12：结构体的拷贝 vs 指针
═══════════════════════════════════════════════════════════════
【核心概念】
结构体传参会整个拷贝所有字段
需要就地修改或字段较多时，改用指针

【运行方式】
go run 12_struct_copy.go

【你会看到】
✓ resizeCopy 后 Size 仍是 0.5（改的是副本）
✓ resizePtr 后 Size 变 999（通过指针修改原件）
✓ 赋值操作也是拷贝，两个结构体互不影响

【思考题】
1. 给 Position 再加 10 个字段，用 %p 打印地址，观察拷贝
2. ptr.Label() 能编译吗？Go 会自动解引用吗？
3. 什么时候用值、什么时候用指针？
   • 要修改原件？→ 指针
   • 结构体很大（>几十字节）？→ 指针
   • 需要表示"不存在"（nil）？→ 指针
═══════════════════════════════════════════════════════════════
*/

// Position 交易持仓结构
type Position struct {
	Symbol string  // 交易对
	Size   float64 // 仓位大小
	Entry  float64 // 入场价格
}

// Label 返回标签（值接收者）
func (p Position) Label() string {
	return fmt.Sprintf("%s %.2f", p.Symbol, p.Size)
}

// resizeCopy 值传递：改副本，不影响原件
func resizeCopy(p Position) {
	fmt.Printf("    resizeCopy 收到的地址：%p\n", &p)
	p.Size = 999
	fmt.Println("    resizeCopy 内部：Size =", p.Size, "(改的是副本)")
}

// resizePtr 指针传递：改原件
func resizePtr(p *Position) {
	fmt.Printf("    resizePtr 收到的地址：%p\n", p)
	p.Size = 999
	fmt.Println("    resizePtr 内部：Size =", p.Size, "(修改了原件)")
}

func main() {
	fmt.Println("\n╔═══════════════════════════════════════╗")
	fmt.Println("║  L1-12: 结构体的拷贝 vs 指针         ║")
	fmt.Println("╚═══════════════════════════════════════╝\n")

	// ========== 示例 1：值传递拷贝结构体 ==========
	fmt.Println("【示例 1：值传递（整个结构体被拷贝）】")
	pos := Position{Symbol: "BTC", Size: 0.5, Entry: 63000}
	fmt.Printf("  原始 pos 地址：%p\n", &pos)
	fmt.Printf("  原始值：%+v\n", pos)

	resizeCopy(pos)
	fmt.Printf("  调用 resizeCopy 后：Size = %.2f (未变) %s\n\n", pos.Size, pos.Label())

	// ========== 示例 2：指针传递 ==========
	fmt.Println("【示例 2：指针传递（传地址，可修改原件）】")
	resizePtr(&pos)
	fmt.Printf("  调用 resizePtr 后：Size = %.2f (已改)\n\n", pos.Size)

	// ========== 示例 3：赋值也是拷贝 ==========
	fmt.Println("【示例 3：赋值操作也是拷贝】")
	copy2 := pos // 整体拷贝所有字段
	copy2.Symbol = "ETH"
	copy2.Size = 10

	fmt.Printf("  改 copy2 后：\n")
	fmt.Printf("    copy2: %+v\n", copy2)
	fmt.Printf("    pos:   %+v (未受影响)\n\n", pos)

	// ========== 示例 4：指针的自动解引用 ==========
	fmt.Println("【示例 4：Go 的语法糖 - 自动解引用】")
	ptr := &pos
	fmt.Println("  ptr.Label() =", ptr.Label(), "(Go 自动解引用，等价于 (*ptr).Label())")
	fmt.Println("  pos.Label() =", pos.Label())

	fmt.Println("\n【关键总结】")
	fmt.Println("  • 结构体赋值和传参都是整体拷贝")
	fmt.Println("  • 需要修改原件 → 用指针")
	fmt.Println("  • 结构体很大 → 用指针（避免拷贝开销）")
	fmt.Println("  • ptr.Method() 会自动解引用为 (*ptr).Method()")
}
