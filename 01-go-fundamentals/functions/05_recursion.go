package functions

import (
	"fmt"

	"go-study/go-fundamentals/level"
)

// L05 第五关：递归 —— 函数调用自己。
func L05() level.Level {
	var fact func(n int) int
	fact = func(n int) int { // 具名函数也能写成变量，这里故意用匿名函数体现递归需先声明
		if n <= 1 {
			return 1
		}
		return n * fact(n-1)
	}

	return level.Level{
		ID:      "L1-05",
		Title:   "递归与终止条件",
		Tags:    "recursion · 栈",
		Pre:     "L1-02",
		Goal:    "写递归时必须先定「终止条件」，否则就是无限递归",
		Observe: "fact(7)=5040；把终止条件删掉会立刻栈溢出",
		Questions: []string{
			"为什么 var fact func(int) int 要先声明再赋值？直接 fact := func(n int) int { ... fact(n-1) } 为什么不行？",
			"把 if n <= 1 分支删掉再运行，报什么错？（panic: stack overflow）这就是终止条件的作用",
			"用 for 循环改写阶乘，性能差别在哪？（提示：每层递归都要一个栈帧）",
		},
		Check: "能说出递归三要素（终止条件、递推、返回值），并解释栈溢出成因",
		Run: func() {
			for _, n := range []int{0, 1, 5, 7} {
				fmt.Printf("  fact(%d) = %d\n", n, fact(n))
			}
			fmt.Println("  ↑ 0! 与 1! 都等于 1，因为终止条件写的是 n <= 1")
		},
	}
}
