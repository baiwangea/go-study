//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 01-go-fundamentals && go run packages/02_import.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"

	h "go-study/go-fundamentals/packages/helper" // 别名导入：调用时写成 h.XXX
)

/*
═══════════════════════════════════════════════════════════════════════
L1-07：四种导入写法
═══════════════════════════════════════════════════════════════════════

【核心概念】
    Go 的 import 有四种形态，用途各不相同：

        1. 普通导入   import "path/to/pkg"      → 用 pkg.XXX 调用
        2. 别名导入   import h "path/to/pkg"    → 用 h.XXX 调用（解决重名 / 简化前缀）
        3. 点导入     import . "path/to/pkg"    → 直接用 XXX（可读性差，项目里几乎不用）
        4. 空白导入   import _ "path/to/pkg"    → 不调用，只为触发该包的 init()

    本关演示的是**别名导入**：同一个 helper 包，上一关用 helper.XXX 调，
    这里改名为 h，调用前缀变成 h.XXX —— 包的身份没变，只是名字变了。

    另外：导入了却不使用，Go 会直接报错（不是警告），这点和「未使用的局部变量」
    是同一条设计哲学 —— 编译器替你抓无用代码。

【你会看到】
    ✓ 同一个 helper 包，本文件用 h.XXX 调用，上一关用 helper.XXX 调用

【思考题】（动手改代码验证，别空想）
    1. import _ "database/sql" 的作用是什么？为什么驱动包常被这样导入？（提示：只为触发 init）
    2. import . "fmt" 叫点导入，Println 可以不带包名调用 —— 为什么实际项目里几乎没人这么写？
    3. 导入了却没使用，Go 会报错还是警告？这和「未使用的局部变量」是同一条规则吗？

【通关标准】
    能说出别名导入的适用场景（包名冲突、调用太长），并解释 _ 导入只为执行 init。

【前置】L1-06              【标签】import · 别名 · 点导入 · 空白导入
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n=== L1-07 四种导入写法 ===")

	// ── 示例 1：别名导入 —— 用 h.XXX 代替 helper.XXX ─────────────────────
	fmt.Println("\n【示例 1】别名导入：h 就是 helper")
	fmt.Println("  h.Version         =", h.Version)
	fmt.Println("  h.PublicFunction  =", h.PublicFunction("aliased import works"))

	// ── 示例 2：别名不改变包的身份，只改调用前缀 ─────────────────────────
	fmt.Println("\n【示例 2】别名只影响调用前缀")
	p := h.NewPair("beta", "s3cr3t")
	fmt.Printf("  h.NewPair(...)    → Name=%q Secret()=%q\n", p.Name, p.Secret())
	fmt.Println("  ↑ 本文件把 helper 改名为 h 使用；包的身份没变，只是调用前缀变了")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n【关键总结】")
	fmt.Println("  • 普通导入 pkg.XXX / 别名导入 h.XXX / 点导入直接用 / 空白导入只为 init")
	fmt.Println("  • 别名常用于包名冲突，或把长前缀写短")
	fmt.Println("  • 导入了不用会编译报错，点导入和空白导入是例外场景")
}
