package main

import (
	"fmt"

	"go-study/go-pointers/level"
)

func zeroVal(v int)  { v = 0 }  // 收到的是副本
func zeroPtr(p *int) { *p = 0 } // 收到的是地址

// L1-11：值传递 vs 指针传递。
func L11() level.Level {
	return level.Level{
		ID:      "L1-11",
		Title:   "值传递与指针传递",
		Tags:    "值拷贝 · & 取地址 · * 解引用",
		Pre:     "L1-01",
		Goal:    "建立 Go 的第一原则：所有赋值与传参都是拷贝；想改原件必须传指针",
		Observe: "zeroVal 后 i 仍是 1；zeroPtr(&i) 后 i 变成 0",
		Questions: []string{
			"打印 p := &i 与 *p，再打印 p 与 &i 是否相同？说明 & 的语义。",
			"把 zeroPtr 的参数写成 p int，函数内 *p = 0 会报什么编译错误？",
			"PHP/Python 里对象传进去随便改，Go 为什么坚持「全部是拷贝」？（提示：并发与所有权）",
		},
		Check: "能解释为什么 Go 没有「引用传递」只有「传指针」，并写出让调用方可见的修改",
		Run: func() {
			i := 1
			fmt.Println("  初始 i =", i)

			zeroVal(i)
			fmt.Println("  zeroVal(i)  之后 i =", i, "← 改的是副本，无效果")

			zeroPtr(&i)
			fmt.Println("  zeroPtr(&i) 之后 i =", i, "← 通过地址改到了原件")

			p := &i
			fmt.Printf("  p = %p，&i = %p，二者相等：%t；*p 解引用得到 %d\n", p, &i, p == &i, *p)
		},
	}
}
