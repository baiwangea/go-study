package main

import (
	"fmt"

	"go-study/go-pointers/level"
)

type tradeErr struct{ code int }

func (e *tradeErr) Error() string { return fmt.Sprintf("下单失败 code=%d", e.code) }

// badDo 故意演示「返回 nil 指针给 error 接口」这个最经典的坑。
func badDo(fail bool) error {
	var e *tradeErr
	if fail {
		e = &tradeErr{code: 500}
	}
	return e // e 为 nil 时，返回的接口值并不等于 nil！
}

func goodDo(fail bool) error {
	if fail {
		return &tradeErr{code: 500}
	}
	return nil // 明确返回 nil 接口
}

// L1-15：nil 的真面目。
func L15() level.Level {
	return level.Level{
		ID:      "L1-15",
		Title:   "nil 陷阱与接口 nil",
		Tags:    "nil map · nil slice · 接口 = (类型, 值)",
		Pre:     "L1-14",
		Goal:    "分清四种 nil 的行为差异，尤其「接口里装了一个 nil 指针，接口本身不是 nil」",
		Observe: "nil map 读安全写 panic；nil slice 可以 append；badDo(false) 返回的 err != nil",
		Questions: []string{
			"对 nil map 执行 m[\"k\"] = 1 报什么 panic？读 m[\"k\"] 呢？",
			"badDo(false) 已经证明 err != nil —— 把它改成 return nil 的写法（goodDo），说清接口值的 (类型, 值) 二元结构",
			"刚才 go vet ./... 是干净的 —— 标准 vet 抓不到这个 bug，能抓它的是 staticcheck 的 nilness 分析（SA4031 一类）。这说明 lint 工具选型为什么重要？",
		},
		Check: "能预判任意 nil 操作的结果，并保证函数返回 error 时永远返回 nil 接口而非 nil 指针",
		Run: func() {
			var s []int
			s = append(s, 1, 2) // nil slice 可以 append
			fmt.Printf("  nil slice append 后：%v（len=%d cap=%d）\n", s, len(s), cap(s))

			var m map[string]int
			fmt.Printf("  nil map 读：m[\"x\"]=%d（零值，不报错）\n", m["x"])
			fmt.Println("  nil map 写：会 panic: assignment to entry in nil map —— 这里故意不执行")

			fmt.Printf("  badDo(false)  → err != nil 为 %t ← 坑！\n", badDo(false) != nil)
			fmt.Printf("  goodDo(false) → err != nil 为 %t ← 正确写法\n", goodDo(false) != nil)
			fmt.Printf("  badDo(true)  内容：%s\n", badDo(true).Error())
		},
	}
}
