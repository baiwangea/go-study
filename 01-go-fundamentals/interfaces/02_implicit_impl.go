//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 01-go-fundamentals && go run interfaces/02_implicit_impl.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"math"
)

/*
═══════════════════════════════════════════════════════════════════════
L1-09：隐式实现与接收者
═══════════════════════════════════════════════════════════════════════

【核心概念】
    Go 没有 implements 关键字，实现是**隐式**的：方法集合对得上，就满足接口。
    真正的坑在**接收者类型**：

        · 值接收者  func (s Square) Area()      → Square 和 *Square 都满足接口
        · 指针接收者 func (r *Ring) Area()       → 只有 *Ring 满足接口，Ring 不行

    原因：方法集规则 —— *T 的方法集包含「值接收者 + 指针接收者」两套方法，
    而 T 的方法集只包含值接收者那一套。所以用指针接收者实现时，
    必须传 &T 才满足接口。

    编译期断言 var _ Shaper = Square{} 把「谁实现了接口」写成代码，
    接口一旦改动，错误立刻在这里暴露，比等到调用处才发现好得多。

【你会看到】
    ✓ Square 与 *Ring 都能装进 Shaper
    ✓ Blob 装不进去（注释掉的那行会报错：missing method Area）

【思考题】（动手改代码验证，别空想）
    1. 把 _ Shaper = (*Ring)(nil) 改成 _ Shaper = Ring{}，报错是什么？为什么？
    2. var s Shaper = Ring{...} 也编译不过，但 s := &Ring{...} 却可以 —— 说清规则
    3. 上面的 var _ Shaper = ... 这种「空白变量断言」在标准库里很常见，它解决什么问题？

【通关标准】
    能解释值接收者/指针接收者对接口实现的影响，并写出编译期断言。

【前置】L1-08              【标签】implements · 值/指针接收者
═══════════════════════════════════════════════════════════════════════
*/

// Shaper 要求实现 Area() float64。
type Shaper interface {
	Area() float64
}

// Square 用值接收者实现 Shaper —— Square 和 *Square 都满足接口。
type Square struct{ Side float64 }

func (s Square) Area() float64 { return s.Side * s.Side }

// Ring 用指针接收者实现 Shaper —— 只有 *Ring 才满足接口。
type Ring struct{ Radius float64 }

func (r *Ring) Area() float64 { return math.Pi * r.Radius * r.Radius }

// Blob 故意只实现了 Perimeter，没有 Area，因此不是 Shaper。
type Blob struct{ Base float64 }

func (b Blob) Perimeter() float64 { return 4 * b.Base }

// 编译期断言：把「谁实现了接口」写成代码，接口一改就立刻在这里报错。
var (
	_ Shaper = Square{}
	_ Shaper = (*Ring)(nil)
	// _ Shaper = Blob{} // 打开这行：cannot use Blob as Shaper (missing method Area)
)

func main() {
	fmt.Println("\n=== L1-09 隐式实现与接收者 ===")

	// ── 示例 1：值接收者与指针接收者都能装进接口 ─────────────────────────
	// Square 传值可以（值接收者），Ring 必须传指针 &Ring{...}（指针接收者）。
	fmt.Println("\n【示例 1】同一个接口，接住两种实现")
	shapes := []Shaper{Square{Side: 3}, &Ring{Radius: 2}}
	for _, sh := range shapes {
		fmt.Printf("  %-18s Area = %.2f\n", fmt.Sprintf("%T", sh), sh.Area())
	}

	// ── 示例 2：方法对不上，就不是这个接口 ───────────────────────────────
	// Blob 有 Perimeter() 但没有 Area()，所以它不满足 Shaper，装不进上面的切片。
	fmt.Println("\n【示例 2】Blob 缺 Area，不是 Shaper")
	b := Blob{Base: 5}
	fmt.Printf("  Blob 只有 Perimeter()=%.0f —— 它不是 Shaper，装不进上面的切片\n", b.Perimeter())

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n【关键总结】")
	fmt.Println("  • 实现是隐式的：方法集对得上就满足接口，无需 implements")
	fmt.Println("  • 指针接收者实现 → 只有 *T 满足接口；值接收者 → T 与 *T 都满足")
	fmt.Println("  • var _ I = T{} 是编译期断言，能在接口变动时第一时间报错")
}
