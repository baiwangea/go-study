//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 01-go-fundamentals && go run interfaces/03_polymorphism.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"math"
	"strings"
)

/*
═══════════════════════════════════════════════════════════════════════
L1-10：多态与类型断言
═══════════════════════════════════════════════════════════════════════

【核心概念】
    多态 = 同一段代码处理不同具体类型，靠接口参数实现：

        func measure(s Shape) { ... }   // 它不知道也不关心是 Rect 还是 Circle

    当接口不够用、需要拿回**具体类型**时，用类型断言 / type switch：
        v, ok := s.(Rect)          // comma-ok：失败不 panic，只得到 ok=false
        switch v := s.(type) {...} // type switch：按动态类型分支

    额外彩蛋：fmt 打印一个值时会检查它是否满足 fmt.Stringer（有 String() string
    方法）。满足就自动调用 String() 来决定打印内容 —— 所以 Rect 打印出来是
    Rect{10.0 x 5.0} 而不是 {10 5}。

【你会看到】
    ✓ 同一个 measure 函数处理两种类型
    ✓ %v 打印 Rect 时自动调用了它的 String()
    ✓ comma-ok 断言得到 ok=true

【思考题】（动手改代码验证，别空想）
    1. 给 Circle 也加一个 String() 方法，输出会怎样变化？这说明 fmt 包在做什么？
    2. s.(Rect) 断言失败会 panic，comma-ok 写法怎么避免？两种写法分别什么时候用？
    3. 如果一个接口有 10 个方法，调用方真的都需要吗？（提示：接口应尽量小，
       「接受接口，返回结构体」）

【通关标准】
    能写出接收接口的函数，并用 type switch 区分具体类型；能说清 Stringer 的魔法在哪。

【前置】L1-09              【标签】polymorphism · type switch · Stringer
═══════════════════════════════════════════════════════════════════════
*/

// Shape 是本关复用的接口：只要有 Area 方法就算形状。
type Shape interface{ Area() float64 }

// Rect 除了 Area，还额外实现了 fmt.Stringer（有个 String() 方法）。
type Rect struct{ Width, Height float64 }

func (r Rect) Area() float64  { return r.Width * r.Height }
func (r Rect) String() string { return fmt.Sprintf("Rect{%.1f x %.1f}", r.Width, r.Height) }

// Circle 只实现了 Area。
type Circle struct{ Radius float64 }

func (c Circle) Area() float64 { return math.Pi * c.Radius * c.Radius }

// measure 接收接口类型：它不知道也不关心具体是哪种形状。
// 这正是多态：一份代码，两种类型通用。
func measure(s Shape) {
	fmt.Printf("  measure(%-22v) Area=%.2f\n", s, s.Area())
}

// describe 用类型断言把接口值还原成具体类型。
func describe(s Shape) string {
	switch v := s.(type) { // type switch：按动态类型分支，v 是对应具体类型
	case Rect:
		return fmt.Sprintf("矩形，面积 %.2f", v.Area())
	case Circle:
		return fmt.Sprintf("圆形，面积 %.2f", v.Area())
	default:
		return fmt.Sprintf("未知形状 %T", v)
	}
}

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L1-10 多态与类型断言")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	r := Rect{Width: 10, Height: 5}
	c := Circle{Radius: 5}

	// ── 示例 1：多态 —— 同一个函数处理两种类型 ───────────────────────────
	// measure 的参数是 Shape，传 Rect 或 Circle 都能编译、都能正确算面积。
	fmt.Println("📝 示例 1：多态：一个 measure 处理两种类型")
	fmt.Println(strings.Repeat("-", 60))
	measure(r)
	measure(c)

	// ── 示例 2：type switch 还原具体类型 ─────────────────────────────────
	fmt.Println("📌 示例 2：type switch：按动态类型分支")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("  describe(r) =", describe(r))
	fmt.Println("  describe(c) =", describe(c))

	// ── 示例 3：fmt.Stringer 的魔法 ──────────────────────────────────────
	// Rect 有 String()，于是 fmt 打印它时会自动调用 String()。
	fmt.Println("🔍 示例 3：Rect 满足 fmt.Stringer")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("  直接打印 Rect 得到 %v —— 因为 Rect 满足 fmt.Stringer\n", r)

	// ── 示例 4：comma-ok 断言，失败不 panic ──────────────────────────────
	fmt.Println("🧪 示例 4：comma-ok 类型断言")
	fmt.Println(strings.Repeat("-", 60))
	var s Shape = r
	v, ok := s.(Rect) // comma-ok 写法：断言失败不 panic，只得到 ok == false
	fmt.Printf("  类型断言：s.(Rect) → ok=%v, v=%v\n", ok, v)

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ 接口作参数 = 多态：调用方不必知道具体类型")
	fmt.Println("  ✅ 需要具体类型时用类型断言 s.(T)，安全写法是 comma-ok 两返回值")
	fmt.Println("  ✅ 满足 fmt.Stringer 的类型，%v 打印时自动调用 String()")
}
