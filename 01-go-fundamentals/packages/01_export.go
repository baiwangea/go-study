//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 01-go-fundamentals && go run packages/01_export.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"

	"go-study/go-fundamentals/packages/helper" // 被导入的示例包
)

/*
═══════════════════════════════════════════════════════════════════════
L1-06：导出与封装
═══════════════════════════════════════════════════════════════════════

【核心概念】
    Go 只有一条可见性规则：**标识符首字母大写 = 对外可见（导出），
    小写 = 包内私有（非导出）**。没有 public / private 关键字。

    这条规则同时管三样东西：
        · 函数：PublicFunction 可调，privateFunction 不行
        · 常量/变量：Version 可读
        · 结构体字段：Name 可读，secret 不行

    想访问非导出字段，只能由包自己提供方法（如 p.Secret()）或构造函数
    （如 helper.NewPair）—— 这叫「受控暴露」，就是 Go 的封装手段。

【你会看到】
    ✓ helper.Version / PublicFunction / NewPair 都能调用
    ✓ 被注释掉的两行如果取消注释，会编译失败

【思考题】（动手改代码验证，别空想）
    1. 把源码里注释掉的两行（helper.privateFunction() 与 _ = p.secret）取消注释，
       两条报错信息分别是什么？
    2. 为什么外部能读 p.Name 却不能读 p.secret？想用 secret 只能走 p.Secret() ——
       这种「受控暴露」和你熟悉的 getter/setter 有何异同？
    3. 同一个目录里放两个 package 名不同的文件会怎样？（提示：一个目录 = 一个包）

【通关标准】
    能说清「大写导出」这条唯一规则，并解释为什么需要 NewPair 这样的构造函数。

【前置】L1-01              【标签】package · 可见性
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n=== L1-06 导出与封装 ===")

	// ── 示例 1：调用导出的常量与函数 ─────────────────────────────────────
	fmt.Println("\n【示例 1】导出标识符：首字母大写，外部可见")
	fmt.Println("  helper.Version          =", helper.Version)
	fmt.Println("  helper.PublicFunction(...) =", helper.PublicFunction("hello from helper package"))

	// ── 示例 2：字段级的可见性 + 受控暴露 ────────────────────────────────
	// NewPair 之所以存在，正是因为 secret 是非导出字段，外部没法直接构造它。
	fmt.Println("\n【示例 2】结构体字段：Name 可读，secret 只能走方法")
	p := helper.NewPair("alpha", "hidden-token")
	fmt.Printf("  p.Name = %q（导出字段，可直接访问）\n", p.Name)
	fmt.Printf("  p.Secret() = %q（非导出字段，只能走方法）\n", p.Secret())

	// ── 示例 3：反例留在这里，动手取消注释就能看到编译错误 ───────────────
	// 下面两行故意保持注释：取消注释会得到 undefined / cannot refer to unexported 错误
	// helper.privateFunction()
	// _ = p.secret
	fmt.Println("\n【示例 3】非导出标识符：取消注释即编译失败")
	fmt.Println("  源码里两行注释掉的调用是故意的：取消注释即可看到编译错误")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n【关键总结】")
	fmt.Println("  • 首字母大写 = 导出，小写 = 包内私有，这是 Go 唯一的可见性规则")
	fmt.Println("  • 非导出字段靠方法 / 构造函数受控暴露（p.Secret()、NewPair）")
	fmt.Println("  • 一个目录 = 一个包，同目录文件必须同一个 package")
}
