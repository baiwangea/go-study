package functions

import (
	"fmt"

	"go-study/go-fundamentals/level"
)

// L01 第一关：函数与返回值。
func L01() level.Level {
	// 两个参数类型相同时可以只写一次：func plus(a, b int) int
	plus := func(a, b int) int { return a + b }

	return level.Level{
		ID:      "L1-01",
		Title:   "函数与返回值",
		Tags:    "参数 · 返回类型位置",
		Pre:     "无（第一关）",
		Goal:    "写出带参数与返回值的函数，理解返回类型为什么写在参数表后面",
		Observe: "plus(1, 2) 输出 3；丢弃返回值也能编译通过",
		Questions: []string{
			"func plus(a, b int) int 与 func plus(a int, b int) int 等价吗？漏写一个 int 会怎样？",
			"调用 plus(1, 2) 却不接收返回值，Go 报错吗？和 Java/JS 有什么差别？",
		},
		Check: "能口述返回类型的位置规则，并亲手验证两道思考题",
		Run: func() {
			fmt.Println("  1 + 2 =", plus(1, 2))
			plus(1, 2) // 返回值可以直接丢弃，Go 不强制接收
			fmt.Println("  plus(1, 2) 被调用但结果丢弃 —— 编译通过（Go 不强制接收返回值）")
		},
	}
}
