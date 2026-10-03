package interfaces

import (
	"fmt"

	"go-study/go-fundamentals/level"
)

// Speaker 是一个只含一个方法的接口：谁有 Speak() 方法，谁就是 Speaker。
type Speaker interface {
	Speak() string
}

// Dog 并没有写 "implements Speaker"，只是因为有了 Speak 方法就满足了接口。
type Dog struct{ Name string }

func (d Dog) Speak() string { return d.Name + "：汪汪！" }

// L08 第八关：接口是什么 —— 一组方法签名。
func L08() level.Level {
	return level.Level{
		ID:      "L1-08",
		Title:   "接口 = 方法集合",
		Tags:    "interface · 契约",
		Pre:     "L1-06",
		Goal:    "把接口理解成「契约」：只规定要提供哪些方法，不关心谁来实现",
		Observe: "Dog 没有声明自己实现 Speaker，却能赋值给 Speaker 类型的变量",
		Questions: []string{
			"接口里写 0 个方法（interface{} / any）意味着什么？为什么所有类型都能满足它？",
			"给 Speaker 再加一个方法 Sleep()，Dog 会发生什么？（试试，然后看编译错误）",
			"接口字段和方法混着写行吗？（提示：接口只能有方法）",
		},
		Check: "能用自己的话说明「接口是行为契约，不是数据结构」",
		Run: func() {
			var s Speaker = Dog{Name: "旺财"}
			fmt.Println("  s.Speak() =", s.Speak())
			fmt.Printf("  s 的动态类型 = %T，接口值本身只暴露 Speak() 这一个方法\n", s)
		},
	}
}
