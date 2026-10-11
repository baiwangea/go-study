package main

import "fmt"

/*
═══════════════════════════════════════════════════════════════
L1-11：值传递 vs 指针传递
═══════════════════════════════════════════════════════════════
【核心概念】
Go 的第一原则：所有赋值与传参都是拷贝；想改原件必须传指针

【运行方式】
cd 02-go-pointers
go run 11_value_vs_pointer.go

【你会看到】
✓ zeroVal 后 i 仍是 1 (改的是副本)
✓ zeroPtr(&i) 后 i 变成 0 (通过地址改到了原件)

【思考题】
1. 打印 p := &i 与 *p，再打印 p 与 &i 是否相同？
2. 把 zeroPtr 的参数写成 p int，函数内 *p = 0 会报什么编译错误？
3. PHP/Python 里对象传进去随便改，Go 为什么坚持「全部是拷贝」？
   提示：并发与所有权

【通关标准】
能解释为什么 Go 没有「引用传递」只有「传指针」
═══════════════════════════════════════════════════════════════
*/

// zeroVal 收到的是副本，修改不影响原变量
func zeroVal(v int) {
	v = 0
	fmt.Println("    zeroVal 内部：v =", v, "(这是副本)")
}

// zeroPtr 收到的是地址，可以修改原变量
func zeroPtr(p *int) {
	*p = 0 // * 是解引用操作符，访问指针指向的值
	fmt.Println("    zeroPtr 内部：*p =", *p, "(通过指针修改原变量)")
}

func main() {
	fmt.Println("\n╔═══════════════════════════════════════╗")
	fmt.Println("║  L1-11: 值传递 vs 指针传递           ║")
	fmt.Println("╚═══════════════════════════════════════╝\n")

	// ========== 示例 1：值传递 ==========
	fmt.Println("【示例 1：值传递】")
	i := 1
	fmt.Println("  初始 i =", i)

	zeroVal(i)
	fmt.Println("  调用 zeroVal(i) 后 i =", i, "← 改的是副本，无效果\n")

	// ========== 示例 2：指针传递 ==========
	fmt.Println("【示例 2：指针传递】")
	j := 1
	fmt.Println("  初始 j =", j)

	zeroPtr(&j) // & 是取地址操作符
	fmt.Println("  调用 zeroPtr(&j) 后 j =", j, "← 通过地址改到了原件\n")

	// ========== 示例 3：指针的本质 ==========
	fmt.Println("【示例 3：指针的本质】")
	k := 42
	p := &k // p 是指向 k 的指针

	fmt.Printf("  k 的值：%d\n", k)
	fmt.Printf("  k 的地址：%p (使用 &k)\n", &k)
	fmt.Printf("  p 的值（就是 k 的地址）：%p\n", p)
	fmt.Printf("  p 和 &k 相等吗？%t\n", p == &k)
	fmt.Printf("  *p 解引用得到：%d (访问 p 指向的值)\n", *p)

	*p = 100 // 通过指针修改 k
	fmt.Printf("  执行 *p = 100 后，k = %d\n\n", k)

	// ========== 关键总结 ==========
	fmt.Println("【关键总结】")
	fmt.Println("  & 操作符：取地址（&变量 → 指针）")
	fmt.Println("  * 操作符：解引用（*指针 → 值）")
	fmt.Println("  值传递：传副本，函数内修改不影响外部")
	fmt.Println("  指针传递：传地址，函数内可以修改原变量")
	fmt.Println("  ⚠️  Go 没有引用传递，只有值传递和指针传递")
}
