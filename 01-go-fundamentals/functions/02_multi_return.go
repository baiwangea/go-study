package functions

import (
	"fmt"

	"go-study/go-fundamentals/level"
)

// L02 第二关：多返回值 —— Go 错误处理的地基。
func L02() level.Level {
	// 惯用写法：结果 + error 一起返回
	divide := func(a, b float64) (float64, error) {
		if b == 0 {
			return 0, fmt.Errorf("除数不能为 0")
		}
		return a / b, nil
	}

	return level.Level{
		ID:      "L1-02",
		Title:   "多返回值与 error",
		Tags:    "多返回值 · error 惯例",
		Pre:     "L1-01",
		Goal:    "习惯「(结果, error)」这个 Go 最常见的函数形状",
		Observe: "正常除法打印结果；除零打印 error，且第一个返回值为零值",
		Questions: []string{
			"为什么 Go 不用 try/catch，而把 error 当成普通返回值？两种风格差别在哪？",
			"如果只返回 error 不返回结果（func do() error），和 (结果, error) 比哪种更好传参？",
		},
		Check: "能不看资料写出一个返回 (int, error) 的函数，并正确处理 err != nil 分支",
		Run: func() {
			v, err := divide(10, 3)
			fmt.Printf("  10 / 3 = %.2f, err = %v\n", v, err)

			v, err = divide(10, 0)
			fmt.Printf("  10 / 0 = %.2f, err = %v\n", v, err)
			fmt.Println("  ↑ err 非 nil 时结果不可信，必须先看 err（这就是 Go 的 if err != nil）")
		},
	}
}
