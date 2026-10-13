//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 01-go-fundamentals && go run generics/02_constraints.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"cmp"
	"fmt"
	"strings"
)

/*
═══════════════════════════════════════════════════════════════════════
L1-25：泛型约束与取舍
═══════════════════════════════════════════════════════════════════════

【核心概念】
    约束（constraint）就是「类型参数的合法取值范围」，写法是一个接口。
    接口里的元素有两种含义：

        interface{ int | string }    ← 类型**并集**：int 或 string
        interface{ ~int }            ← `~` 表示「底层类型是 int 的任意类型」

    `~` 是关键。Go 里 `type MyInt int` 声明的 MyInt 和 int 是**两个不同的
    类型**（不能互相赋值），但 MyInt 的**底层类型**是 int。
    · 约束写 `interface{ int }`     → 只有 int 本身满足，MyInt 不满足
    · 约束写 `interface{ ~int }`    → 底层是 int 的都满足，MyInt 也行

    三个常用约束：
        any              最宽松，等价 interface{}，什么都不能做（不能 == 也不能 +）
        comparable       内置，支持 == / != 的所有类型（指针、结构体、接口都行）
        cmp.Ordered      标准库 cmp 包，等价 ~int | ~float64 | ~string | ... 支持 < > 的

    什么时候**不该**用泛型（比什么时候该用更重要）：
        1. 只有一种类型实参 —— 直接写具体类型，别绕一圈泛型
        2. 函数体里还是要 type switch / 类型断言 —— 说明你处理的是「不同逻辑」，
           泛型解决的是「同一逻辑、不同类型」，用错了只会更难读
        3. 逻辑短到一屏能看完、且只写一次 —— 直接复制两份更直白
        4. 标准库已经有了 —— min / max 内建函数、cmp.Compare、slices 包

【你会看到】
    ✓ Sum[T Number] 同时接受 []int、[]float64、[]MyInt（因为 ~）
    ✓ MyInt 和 int 是两个不同类型：x.(int) 断言失败，x.(MyInt) 才成功
    ✓ Max[T cmp.Ordered] 对 int / float64 / string 都工作
    ✓ 一个「泛型反而更差」的例子：函数体里还是得 type switch

【思考题】（动手改代码验证，别空想）
    1. Java 的 `class Box<T extends Comparable<T>>` 也是约束，但因为有类型擦除，
       运行时 T 就是 Object；Go 的 cmp.Ordered 约束在编译期就展开成具体代码。
       这带来什么性能差异？什么限制？
    2. 把 Number 约束里的 ~ 全部去掉（写成 interface{ int | int64 | float64 }），
       Sum([]MyInt{...}) 还能编译吗？为什么？（动手试）
    3. 为什么 comparable 不允许 []int？但 comparable 允许 interface 类型？
       （提示：接口比较的是动态类型+值，底层可能是不可比较的，运行时会 panic）
    4. 自己写 min/max 之前先想想：标准库内建 min/max（Go 1.21+）和 cmp.Compare
       已经覆盖了什么？还有必要自己写泛型工具函数吗？

【通关标准】
    能说清 `~` 的作用，并能举出至少两个「不该用泛型」的场景。

【前置】L1-24              【标签】constraint · ~ · comparable · cmp.Ordered · 取舍
═══════════════════════════════════════════════════════════════════════
*/

// Number 是自定义约束：底层类型是 int / int64 / float64 之一的类型都满足。
// 注意必须用 ~，否则下面 Sum([]MyInt{...}) 编译不过。
type Number interface {
	~int | ~int64 | ~float64
}

// StrictInt 是对照组：没有 ~，只有 int 本身满足。
// 用它当约束时，SumStrict([]MyInt{...}) 会编译失败：
//
//	MyInt does not satisfy StrictInt (possibly missing ~ for int in StrictInt)
type StrictInt interface {
	int
}

// MyInt / MyFloat 是自定义的具名类型（底层分别是 int / float64）。
// 它们的用途是给数字加语义（比如「分值」「金额」），但底层还是数字。
type MyInt int
type MyFloat float64

// Sum 用自定义约束：因为 Number 里有 ~，MyInt 这类具名类型也能用。
// total += x 之所以合法，是因为约束里的类型都支持 +。
func Sum[T Number](xs []T) T {
	var total T
	for _, x := range xs {
		total += x
	}
	return total
}

// SumStrict 用无 ~ 的约束：只接受真正的 int。留在这里做对比，不调用。
func SumStrict[T StrictInt](xs []T) T {
	var total T
	for _, x := range xs {
		total += x
	}
	return total
}

// SumInts 是「只有一种类型实参」时的正确写法：直接写具体类型。
// 和 Sum[int](xs) 相比，调用方少写一层方括号，函数体也更直白。
func SumInts(xs []int) int {
	total := 0
	for _, x := range xs {
		total += x
	}
	return total
}

// Contains 用内置的 comparable 约束：能 == 比较的类型都可以。
// 注意 comparable 比 cmp.Ordered 宽：结构体、指针、接口都满足 comparable，
// 但它们不支持 < >，所以不能传给 Max。
func Contains[T comparable](xs []T, v T) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

// Max 用标准库的 cmp.Ordered：支持 < > 的类型（数字、字符串）。
// 自己写 Ordered 的话要列一长串 ~int | ~int8 | ... | ~string，
// 直接用 cmp.Ordered 更省事、也不会漏。
func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Describe 是「泛型反而更差」的反面教材：
// 它想按类型给出不同描述，于是函数体里只能把 T 转成 any 再 type switch ——
// 这正好抵消了泛型的意义（用泛型就是想避免类型分支）。
// 结论：这种需求应该用接口（给每个类型实现 Describe() 方法），而不是泛型。
func Describe[T any](v T) string {
	switch x := any(v).(type) { // 被迫转成 any 才能分支，泛型没帮上忙
	case int:
		return fmt.Sprintf("整数 %d", x)
	case string:
		return fmt.Sprintf("字符串 %q", x)
	default:
		return fmt.Sprintf("其他类型 %T", x)
	}
}

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L1-25 泛型约束与取舍")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// ── 示例 1：自定义约束 + ~ ──────────────────────────────────────────
	fmt.Println("📝 示例 1：自定义约束 ~int | ~int64 | ~float64")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("  Sum([]int{1,2,3})        = %v\n", Sum([]int{1, 2, 3}))
	fmt.Printf("  Sum([]int64{...})        = %v\n", Sum([]int64{10, 20}))
	fmt.Printf("  Sum([]float64{1.5,2.5})  = %v\n", Sum([]float64{1.5, 2.5}))
	fmt.Printf("  Sum([]MyInt{1,2,3})      = %v  ← MyInt 是自定义类型，靠 ~int 才通过\n", Sum([]MyInt{1, 2, 3}))
	fmt.Printf("  Sum([]MyFloat{1.5,2.5})  = %v\n", Sum([]MyFloat{1.5, 2.5}))
	fmt.Println("  ➡️  如果没有 ~，只有 int / float64 本身能用，MyInt / MyFloat 会被拒绝")

	// ── 示例 2：MyInt 和 int 确实是两个类型 ─────────────────────────────
	fmt.Println("📌 示例 2：为什么需要 ~：MyInt 和 int 不是同一个类型")
	fmt.Println(strings.Repeat("-", 60))
	var x any = MyInt(5)
	_, isInt := x.(int)     // MyInt 不是 int
	_, isMyInt := x.(MyInt) // 只有 MyInt 断言成功
	fmt.Printf("  var x any = MyInt(5)：x.(int) 成功？%v   x.(MyInt) 成功？%v\n", isInt, isMyInt)
	fmt.Println("  ➡️  所以 interface{ int } 匹配不到 MyInt，必须写 interface{ ~int }")
	fmt.Printf("  对照 SumInts([]int{1,2,3}) = %v  ← 只有一种类型时，直接写具体类型更清晰\n",
		SumInts([]int{1, 2, 3}))

	// ── 示例 3：cmp.Ordered ─────────────────────────────────────────────
	fmt.Println("🔍 示例 3：cmp.Ordered：需要 < > 时用标准库的约束")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("  Max(3, 5)           = %v\n", Max(3, 5))
	fmt.Printf("  Max(1.5, 0.5)       = %v\n", Max(1.5, 0.5))
	fmt.Printf("  Max(\"apple\",\"pear\") = %q\n", Max("apple", "pear"))
	fmt.Printf("  cmp.Compare(1, 2)   = %d  ← 返回 -1/0/1，适合当排序的比较器\n", cmp.Compare(1, 2))
	fmt.Printf("  cmp.Or(\"\", \"fallback\") = %q  ← 取第一个非零值\n", cmp.Or("", "fallback"))
	fmt.Printf("  内建 min(3,5)/max(3,5) = %d/%d  ← 标准库已经写好了，别自己造\n", min(3, 5), max(3, 5))

	// ── 示例 4：comparable 比 cmp.Ordered 宽 ────────────────────────────
	fmt.Println("🧪 示例 4：comparable：能 == 就行，不需要 < >")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("  Contains([]int, 3)                 = %v\n", Contains([]int{1, 2, 3}, 3))
	fmt.Printf("  Contains([]string, \"b\")            = %v\n", Contains([]string{"a", "b"}, "b"))
	type Point struct{ X, Y int }
	pts := []Point{{1, 2}, {3, 4}}
	fmt.Printf("  Contains([]Point, {3,4})           = %v  ← 结构体也能比较\n", Contains(pts, Point{3, 4}))
	fmt.Println("  ➡️  结构体、指针、接口都满足 comparable，但它们没有大小关系，不能给 Max 用")

	// ── 示例 5：反面教材 —— 泛型没帮上忙 ────────────────────────────────
	fmt.Println("⚙️ 示例 5：什么时候不该用泛型")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("  Describe(42)        = %s\n", Describe(42))
	fmt.Printf("  Describe(\"hi\")      = %s\n", Describe("hi"))
	fmt.Printf("  Describe(3.14)      = %s\n", Describe(3.14))
	fmt.Println("  ➡️  这个泛型函数体里还是 type switch —— 说明处理的是「不同逻辑」，")
	fmt.Println("    正确做法是定义 Describe() 方法的接口，让各类型自己实现（见 L1-27）")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ 约束就是一个接口：| 是并集，~ 表示「底层类型是它」")
	fmt.Println("  ✅ 具名类型（type MyInt int）不等于 int，想让它通过约束必须写 ~int")
	fmt.Println("  ✅ comparable = 支持 ==；cmp.Ordered = 支持 < >；any = 什么都不支持")
	fmt.Println("  ✅ 只有一种类型实参 → 直接写具体类型；函数体要 type switch → 改用接口")
	fmt.Println("  ✅ 标准库已有 min/max/cmp.Compare/slices，别重复造泛型工具")
	fmt.Println("  ✅ 泛型的价值是「同一逻辑 × 多种类型」，不是「一种逻辑里分辨多种类型」")
}
