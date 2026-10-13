//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 01-go-fundamentals && go run generics/01_type_params.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"strings"
)

/*
═══════════════════════════════════════════════════════════════════════
L1-24：泛型类型参数入门
═══════════════════════════════════════════════════════════════════════

【核心概念】
    泛型 = 「把类型也当参数」。Go 1.18 引入，写法是在函数/类型名后面加
    一对方括号声明**类型参数**：

        func Map[T, U any](xs []T, f func(T) U) []U      // 泛型函数
        type Stack[T any] struct { items []T }           // 泛型类型

    [T any] 里的 `any` 是**约束（constraint）**：它规定 T 可以是哪些类型。
    `any` 是最宽松的约束，等价于 `interface{}` —— 任何类型都行。

    泛型类型的方法也要带上类型参数：func (s *Stack[T]) Push(v T)。
    实例化时写 Stack[int]{}、Stack[string]{}，编译器会为每种实参生成代码。

    ⚠️ T 不能直接比较！`if a == b` 在 T 是 any 时编译不过：
        invalid operation: a == b (type parameter T is not comparable with ==)
    原因：T 可能是 slice/map/func 这些**不可比较**的类型，Go 必须静态保证
    运算符合法。想要 == 就把约束换成 `comparable`（见 L1-25）。

    和 Java 泛型的最大差别：Java 是**类型擦除**（运行时 T 就是 Object，
    到处插入强转，取错了运行时才 ClassCastException）；Go 保留静态类型，
    实例化时按具体类型生成代码，编译期就能发现类型错误，也没有装箱开销。

【你会看到】
    ✓ Map / Filter / Reduce 一套代码同时处理 []int 和 []string
    ✓ Stack[int] 与 Stack[string] 各自独立、类型安全（Push 错类型编译不过）
    ✓ Index[T comparable] 能用 == 比较；改成 [T any] 会编译失败（见注释）
    ✓ 用 any + 类型断言的老写法要等到**运行时**才发现类型不对

【思考题】（动手改代码验证，别空想）
    1. Java 的 List<T> 因为擦除，运行时拿不到 T 是什么（所以 new T[] 不合法）；
       Go 的 Stack[T] 能吗？试试在方法里写 `var zero T`，能不能用？
    2. 把 Index 的约束从 comparable 改回 any，把编译错误信息抄下来，
       然后解释「为什么 Go 不像 Java 那样在运行时替你判断」。
    3. any + 类型断言 的 SumAny 和泛型版 Sum 相比，各有什么代价？
       （提示：前者编译期不管，运行时可能 panic/报错；后者编译期管，但会生成多份代码）
    4. 如果只想支持 int 和 float64 两种类型，用泛型更好，还是写两个函数更好？
       （先自己想，L1-25 会给判断标准）

【通关标准】
    能写出一个泛型函数和一个泛型类型，并说清 `[T any]` 里 any 的含义。

【前置】L1-08              【标签】generics · type parameter · any · 泛型类型
═══════════════════════════════════════════════════════════════════════
*/

// Map 把 []T 逐元素变成 []U：T 和 U 是两个独立的类型参数。
// 没有泛型时，这个函数要么为每种类型写一遍，要么用 []any + 类型断言。
func Map[T, U any](xs []T, f func(T) U) []U {
	out := make([]U, 0, len(xs))
	for _, x := range xs {
		out = append(out, f(x))
	}
	return out
}

// Filter 按谓词筛选：pred 的类型是 func(T) bool，T 由调用方决定。
func Filter[T any](xs []T, pred func(T) bool) []T {
	out := make([]T, 0, len(xs))
	for _, x := range xs {
		if pred(x) {
			out = append(out, x)
		}
	}
	return out
}

// Reduce 把切片折叠成单个值：T 是元素类型，A 是累积值类型（两者可以不同）。
func Reduce[T, A any](xs []T, init A, f func(acc A, x T) A) A {
	acc := init
	for _, x := range xs {
		acc = f(acc, x)
	}
	return acc
}

// Index 需要**比较**元素，所以约束必须是 comparable 而不是 any。
// comparable 是语言内置的约束：所有可以用 == / != 的类型（含指针、结构体、接口）。
func Index[T comparable](xs []T, target T) int {
	for i, x := range xs {
		if x == target { // 只有在 comparable 约束下，== 才合法
			return i
		}
	}
	return -1
}

// 反例（把上面改成 [T any] 会编译失败，错误信息大致如下，可以自己动手验证）：
//
//	invalid operation: x == target (type parameter T is not comparable with ==)
//
// 因为 T 可能是 []int / map / func 这类不可比较的类型，Go 必须在编译期排除。

// ── 泛型类型：一个可以装任意元素的栈 ────────────────────────────────────
// 方法名后面也要写 [T]，才能引用类型参数 T。
type Stack[T any] struct {
	items []T
}

// Push 入栈。接收者用指针，因为要修改 items。
func (s *Stack[T]) Push(v T) {
	s.items = append(s.items, v)
}

// Pop 出栈，返回 (值, 是否成功)。
// 注意零值问题：栈空时只能返回 T 的零值 + false —— 泛型里没法返回 nil
// （T 可能是 int 这种非指针类型），所以「comma-ok」是必须的。
func (s *Stack[T]) Pop() (T, bool) {
	var zero T // 声明一个 T 类型的零值变量
	if len(s.items) == 0 {
		return zero, false
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, true
}

// Len 返回元素个数。
func (s *Stack[T]) Len() int { return len(s.items) }

// ── 对照组：没有泛型时怎么写 ────────────────────────────────────────────
// 用 []any + 类型断言：编译期完全不管，类型错了只能等运行时。
func SumAny(xs []any) (int, error) {
	total := 0
	for i, x := range xs {
		v, ok := x.(int) // 类型断言：运行时才知道能不能转
		if !ok {
			return 0, fmt.Errorf("SumAny: 第 %d 个元素是 %T，不是 int", i, x)
		}
		total += v
	}
	return total, nil
}

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L1-24 泛型类型参数入门")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// ── 示例 1：一个 Map 处理多种类型 ───────────────────────────────────
	fmt.Println("📝 示例 1：泛型函数：一套 Map 处理 []int 和 []string")
	fmt.Println(strings.Repeat("-", 60))
	nums := []int{1, 2, 3, 4}
	doubled := Map(nums, func(x int) int { return x * 2 })
	fmt.Printf("  Map([]int, x*2)       = %v\n", doubled)

	names := []string{"alice", "bob"}
	upper := Map(names, func(s string) string { return s + "!" })
	fmt.Printf("  Map([]string, s+\"!\")  = %q\n", upper)

	// T 和 U 可以不同：int → string
	strs := Map(nums, func(x int) string { return fmt.Sprintf("<%d>", x) })
	fmt.Printf("  Map([]int, <n>)       = %q  ← 元素类型从 int 变成了 string\n", strs)

	// ── 示例 2：Filter 与 Reduce ────────────────────────────────────────
	fmt.Println("📌 示例 2：Filter / Reduce：约束仍是 any，因为不需要比较")
	fmt.Println(strings.Repeat("-", 60))
	evens := Filter(nums, func(x int) bool { return x%2 == 0 })
	fmt.Printf("  Filter(偶数) = %v\n", evens)
	fmt.Printf("  Reduce(求和) = %v\n", Reduce(nums, 0, func(acc, x int) int { return acc + x }))
	fmt.Printf("  Reduce(拼接) = %q  ← 元素 int，累积值 string，T 和 A 不同\n",
		Reduce(nums, "", func(acc string, x int) string { return fmt.Sprintf("%s%d-", acc, x) }))

	// ── 示例 3：Index 需要 comparable 约束 ──────────────────────────────
	fmt.Println("🔍 示例 3：comparable：需要 == 比较时必须换约束")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("  Index([]int, 3)      = %d\n", Index(nums, 3))
	fmt.Printf("  Index([]string, bob) = %d\n", Index(names, "bob"))
	fmt.Printf("  Index([]int, 99)     = %d  ← 没找到\n", Index(nums, 99))
	fmt.Println("  ➡️  把约束从 comparable 改成 any，这行 if x == target 就编译不过")

	// ── 示例 4：泛型类型 Stack[T] ───────────────────────────────────────
	fmt.Println("🧪 示例 4：泛型类型：Stack[int] 和 Stack[string] 各自独立")
	fmt.Println(strings.Repeat("-", 60))
	var si Stack[int]
	si.Push(1)
	si.Push(2)
	si.Push(3)
	fmt.Printf("  Stack[int] 长度 = %d\n", si.Len())
	for {
		v, ok := si.Pop()
		if !ok {
			break // 栈空，Pop 返回 (零值, false)
		}
		fmt.Printf("    出栈 %d\n", v)
	}

	var ss Stack[string]
	ss.Push("first")
	ss.Push("second")
	top, _ := ss.Pop()
	fmt.Printf("  Stack[string] 出栈 = %q  ← 完全独立的另一份代码\n", top)
	fmt.Println("  ➡️  类型安全：si.Push(\"字符串\") 直接编译报错，不会等到运行时")

	// ── 示例 5：对照 any + 类型断言的老写法 ─────────────────────────────
	fmt.Println("⚙️ 示例 5：对照：any + 类型断言要等运行时才发现问题")
	fmt.Println(strings.Repeat("-", 60))
	if total, err := SumAny([]any{1, 2, 3}); err == nil {
		fmt.Printf("  SumAny([1 2 3]) = %d\n", total)
	}
	_, err := SumAny([]any{1, "二", 3}) // 编译期完全合法
	fmt.Printf("  SumAny([1 \"二\" 3]) → err = %v\n", err)
	fmt.Println("  ➡️  同样是「算和」，泛型版在编译期就拦住，any 版要等运行时")
	fmt.Println("    但 any 版也有优势：能接受异质切片（元素类型各不相同）")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ 泛型 = 把类型当参数：[T any] 是函数/类型后面的类型参数列表")
	fmt.Println("  ✅ any 是最宽松的约束，等价于 interface{}，任何类型都满足")
	fmt.Println("  ✅ 泛型类型的方法要写 (s *Stack[T])，实例化写 Stack[int]{}")
	fmt.Println("  ✅ T 在 any 约束下不能 ==：可能是 slice/map，必须换成 comparable")
	fmt.Println("  ✅ 泛型类型里取零值用 `var zero T`，栈空只能返回 (零值, false)")
	fmt.Println("  ✅ 对比 Java：Go 不做类型擦除，类型错误在编译期暴露，也没有装箱开销")
}
