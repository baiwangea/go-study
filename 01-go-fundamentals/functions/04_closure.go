package functions

import (
	"fmt"

	"go-study/go-fundamentals/level"
)

// L04 第四关：闭包 —— 捕获的是变量本身。
func L04() level.Level {
	// intSeq 返回的匿名函数“记住”了它所在作用域里的 i
	intSeq := func() func() int {
		i := 0
		return func() int {
			i++
			return i
		}
	}

	return level.Level{
		ID:      "L1-04",
		Title:   "闭包捕获变量",
		Tags:    "closure · 作用域",
		Pre:     "L1-01",
		Goal:    "理解闭包持有的是变量的引用，不是创建时的值快照",
		Observe: "同一个闭包连续调用得到 1 2 3；新建闭包从 1 重新开始",
		Questions: []string{
			"nextInt 和 nextInt2 两个闭包会共享计数器吗？为什么？",
			"把 i++ / return i 改成 return i; i++（写完 i++ 放最后）会编译报错吗？输出如何变化？",
			"for 循环里创建闭包时捕获循环变量，Go 1.22 前后行为有什么不同？（进阶，可先记住结论）",
		},
		Check: "能解释「共享 vs 独立」，并用两个闭包亲手验证",
		Run: func() {
			nextInt := intSeq()
			fmt.Println("  同一个闭包：", nextInt(), nextInt(), nextInt())

			nextInt2 := intSeq()
			fmt.Println("  新闭包重新计数：", nextInt2(), "（与 nextInt 互不影响，各自持有独立的 i）")
			fmt.Println("  再调老闭包：", nextInt(), "（它自己的 i 已经累加到 4）")
		},
	}
}
