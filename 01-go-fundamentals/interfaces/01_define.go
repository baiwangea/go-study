//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 01-go-fundamentals && go run interfaces/01_define.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"strings"
)

/*
═══════════════════════════════════════════════════════════════════════
L1-08：接口 = 方法集合
═══════════════════════════════════════════════════════════════════════

【核心概念】
    接口是一份**行为契约**：它只规定「要提供哪些方法」，不关心「谁来实现、
    怎么实现」。接口里不能有字段，只能有方法签名。

        type Speaker interface { Speak() string }

    谁拥有 Speak() string 这个方法，谁就满足 Speaker —— 不需要写任何
    implements 声明（这点在下一关 L1-09 展开）。
    空接口 interface{}（也就是 any）不含任何方法，所以**所有类型**都满足它。

【你会看到】
    ✓ Dog 没有声明自己实现 Speaker，却能赋值给 Speaker 类型的变量
    ✓ 接口值本身只暴露 Speak() 这一个方法

【思考题】（动手改代码验证，别空想）
    1. 接口里写 0 个方法（interface{} / any）意味着什么？为什么所有类型都能满足它？
    2. 给 Speaker 再加一个方法 Sleep()，Dog 会发生什么？（试试，然后看编译错误）
    3. 接口字段和方法混着写行吗？（提示：接口只能有方法）

【通关标准】
    能用自己的话说明「接口是行为契约，不是数据结构」。

【前置】L1-06              【标签】interface · 契约
═══════════════════════════════════════════════════════════════════════
*/

// Speaker 是只含一个方法的接口：谁有 Speak() 方法，谁就是 Speaker。
type Speaker interface {
	Speak() string
}

// Dog 并没有写 "implements Speaker"，只是因为有了 Speak 方法就满足了接口。
type Dog struct{ Name string }

func (d Dog) Speak() string { return d.Name + "：汪汪！" }

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L1-08 接口 = 方法集合")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// ── 示例 1：把 Dog 装进 Speaker 接口变量 ─────────────────────────────
	// 这一步能编译通过，就说明 Dog 满足 Speaker —— 没有任何 implements 声明。
	fmt.Println("📝 示例 1：Dog 隐式满足 Speaker")
	fmt.Println(strings.Repeat("-", 60))
	var s Speaker = Dog{Name: "旺财"}
	fmt.Println("  s.Speak() =", s.Speak())

	// ── 示例 2：接口值只暴露契约里声明的方法 ─────────────────────────────
	// %T 打出的是动态类型 Dog；但通过 s 你只能调 Speak()，调不到 Dog 的字段。
	fmt.Println("📌 示例 2：接口只暴露契约中的方法")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("  s 的动态类型 = %T，接口值本身只暴露 Speak() 这一个方法\n", s)

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ 接口 = 方法集合，只规定行为，不存数据")
	fmt.Println("  ✅ 有对应方法就满足接口，不需要 implements 声明")
	fmt.Println("  ✅ 空接口 interface{} / any 无方法，所有类型都满足")
}
