//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 本仓库「直观学习模式」约定（所有关卡文件通用）：
//
//  1. 一个文件 = 一关 = 一个完整的 package main + func main()，
//     不需要任何导航框架，直接 `go run 文件名.go` 就能跑。
//  2. 第一行的 `//go:build level` 是关键，别删：
//        · 它让 `go build ./...` / `go vet ./...` 跳过本文件
//          ——否则同一目录下多个 func main() 会报 "main redeclared" 编译失败；
//        · 而 `go run 11_value_vs_pointer.go` 显式指定文件名时会绕过构建约束，
//          所以照样能单独运行。
//  3. 跑法：cd 02-go-pointers && go run 11_value_vs_pointer.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"strings"
)

/*
═══════════════════════════════════════════════════════════════════════
L1-11：值传递 vs 指针传递
═══════════════════════════════════════════════════════════════════════

【核心概念】
    Go 的第一原则：所有赋值与传参都是拷贝。
    想让函数内的修改对外可见，必须传指针。
    Go 里没有「引用传递」，只有值传递和指针传递。

【你会看到】
    ✓ zeroVal(i)  之后 i 仍然是 1  —— 改的是副本
    ✓ zeroPtr(&i) 之后 i 变成 0    —— 通过地址改到了原件
    ✓ p == &k 为 true，*p 可读可写 —— 指针的本质就是一个地址

【思考题】（动手改代码验证，别空想）
    1. 把 zeroPtr 的参数从 *int 改成 int，函数内再写 *p = 0，
       会报什么编译错误？为什么？
    2. 打印 p 和 &i，它们是同一个值吗？这说明了 & 的什么语义？
    3. PHP/Python 里对象传进函数可以随便改，Go 为什么坚持「一切都是拷贝」？
       提示：并发安全与所有权清晰。

【通关标准】
    能解释为什么 Go 没有「引用传递」只有「传指针」，
    并亲手写出一个「让调用方可见的修改」。

【前置】L1-01              【标签】值拷贝 · & 取地址 · * 解引用
═══════════════════════════════════════════════════════════════════════
*/

// zeroVal 收到的是 int 的一份副本：函数内怎么改，都动不到外面。
func zeroVal(v int) {
	v = 0
	fmt.Println("    zeroVal 内部：v =", v, "（这是副本）")
}

// zeroPtr 收到的是地址：*p = 0 会写回调用方的那个变量。
func zeroPtr(p *int) {
	*p = 0 // * 是解引用：访问「p 指向的那个变量」
	fmt.Println("    zeroPtr 内部：*p =", *p, "（通过指针改到了原变量）")
}

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L1-11: 值传递 vs 指针传递")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// ── 示例 1：值传递 —— 改的是副本 ─────────────────────────────────────
	fmt.Println("📝 示例 1：值传递")
	fmt.Println(strings.Repeat("-", 60))
	i := 1
	fmt.Println("  📥 初始 i =", i)

	zeroVal(i) // 传的是 i 的值，不是 i 本身
	fmt.Printf("  📤 调用 zeroVal(i) 后 i = %d ← ❌ 没变，改的是副本\n", i)

	// ── 示例 2：指针传递 —— 改的是原件 ───────────────────────────────────
	fmt.Println("\n📌 示例 2：指针传递")
	fmt.Println(strings.Repeat("-", 60))
	j := 1
	fmt.Println("  📥 初始 j =", j)

	zeroPtr(&j) // & 是取地址：把 j 的地址交给函数
	fmt.Printf("  📤 调用 zeroPtr(&j) 后 j = %d ← ✅ 变了，改的是原件\n", j)

	// ── 示例 3：指针的本质就是一个地址 ───────────────────────────────────
	fmt.Println("\n🔍 示例 3：指针的本质")
	fmt.Println(strings.Repeat("-", 60))
	k := 42
	p := &k // p 的类型是 *int，它的值就是 k 的地址

	fmt.Printf("  k 的值    : %d\n", k)
	fmt.Printf("  k 的地址  : %p   （&k）\n", &k)
	fmt.Printf("  p 的值    : %p   （和 &k 一样，因为 p 存的就是 k 的地址）\n", p)
	fmt.Printf("  p == &k   : %t\n", p == &k)
	fmt.Printf("  *p 解引用 : %d   （读 p 指向的那个值）\n", *p)

	*p = 100 // 通过指针写回 k
	fmt.Printf("  ✅ 执行 *p = 100 后，k = %d\n", k)

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ &变量  → 取地址，得到指针")
	fmt.Println("  ✅ *指针  → 解引用，读写指针指向的那个变量")
	fmt.Println("  ✅ 值传递 → 传副本，函数内修改对外不可见")
	fmt.Println("  ✅ 指针传递 → 传地址，函数内修改对外可见")
	fmt.Println("  ⚠️  Go 没有引用传递，只有值传递和指针传递")
	fmt.Println(strings.Repeat("=", 60) + "\n")
}
