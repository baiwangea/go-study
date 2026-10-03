package main

import (
	"fmt"

	"go-study/go-pointers/level"
)

type Counter struct{ n int }

func (c Counter) IncValue() int { c.n++; return c.n } // 值接收者：自增只作用于副本
func (c *Counter) IncPtr() int  { c.n++; return c.n } // 指针接收者：真正累加

// L1-13：方法接收者选值还是指针。
func L13() level.Level {
	return level.Level{
		ID:      "L1-13",
		Title:   "接收者：值 or 指针",
		Tags:    "method receiver · 状态修改",
		Pre:     "L1-12",
		Goal:    "记住「要改状态就用指针接收者」，并且同一类型的方法接收者风格要统一",
		Observe: "同一个 Counter 连调三次 IncValue 都是 1；IncPtr 依次得到 1 2 3",
		Questions: []string{
			"把 IncValue 改成接收 *Counter，但调用写成 c.IncValue()（c 是值变量，不是 &c），能编译吗？",
			"为什么 Go 官方建议「同一类型的方法不要一半值一半指针」？混用会带来什么困惑？",
			"sync.Mutex 为什么绝不能按值拷贝（vet 会报 copylocks）？把 Mutex 放进 L1-12 的 Position 试试",
		},
		Check: "能解释值接收者拿不到可写副本的原因，并说出含锁/含缓存的结构体必须用指针接收者",
		Run: func() {
			c := Counter{}
			fmt.Print("  IncValue（值接收者）：")
			for i := 0; i < 3; i++ {
				fmt.Printf("%d ", c.IncValue())
			}
			fmt.Println("← 永远是 1，因为每次都在改副本")

			fmt.Print("  IncPtr（指针接收者）：")
			for i := 0; i < 3; i++ {
				fmt.Printf("%d ", c.IncPtr()) // 等价于 (&c).IncPtr()，Go 自动取地址
			}
			fmt.Println("← 正常累加")
			fmt.Printf("  最终 c.n = %d\n", c.n)
		},
	}
}
