package functions

import (
	"fmt"

	"go-study/go-fundamentals/level"
)

// L03 第三关：变长参数（variadic）。
func L03() level.Level {
	// nums ...int 在函数内部就是一个 []int，不用自己建切片
	sum := func(nums ...int) (int, int) {
		total := 0
		for _, n := range nums {
			total += n
		}
		return len(nums), total
	}

	return level.Level{
		ID:      "L1-03",
		Title:   "变长参数 ...",
		Tags:    "variadic · 切片展开",
		Pre:     "L1-01",
		Goal:    "理解 ...int 在函数内就是切片，以及用 s... 把切片展开传参",
		Observe: "传 2 个与 4 个参数都能调用；切片用 nums... 展开后结果一致",
		Questions: []string{
			"已有 []int{1,2,3}，直接 sum(nums) 会编译失败吗？为什么必须写 sum(nums...)？",
			"变长参数必须放在参数表最后，试试 sum(prefix string, nums ...int, suffix string) 看报错",
		},
		Check: "能用 s... 展开传参，并说清函数内 nums 的类型",
		Run: func() {
			count, total := sum(1, 2)
			fmt.Printf("  sum(1, 2)      → 个数 %d，合计 %d\n", count, total)

			count, total = sum(1, 2, 3, 4)
			fmt.Printf("  sum(1, 2, 3, 4) → 个数 %d，合计 %d\n", count, total)

			nums := []int{5, 6, 7}
			count, total = sum(nums...)
			fmt.Printf("  sum(nums...)    → 个数 %d，合计 %d（切片展开传参）\n", count, total)
		},
	}
}
