package main

import (
	"fmt"

	"go-study/go-pointers/level"
)

func newCounter() *int {
	n := 0  // 局部变量
	p := &n // 取了地址 —— Go 会自动把它搬到堆上，返回完全安全
	return p
}

func valueOf(v int) *int { return &v }

// L1-16：取地址与逃逸分析。
func L16() level.Level {
	return level.Level{
		ID:      "L1-16",
		Title:   "逃逸：返回局部变量地址为什么安全",
		Tags:    "escape analysis · 栈/堆 · GC",
		Pre:     "L1-11",
		Goal:    "消除 C/C++ 式的悬垂指针焦虑：Go 编译器自动决定变量放栈还是堆",
		Observe: "newCounter() 返回的指针可用；同一次调用返回的地址稳定；逃逸分析输出 moved to heap",
		Questions: []string{
			"跑 go build -gcflags='-m' .，找出哪些变量被 marked as escaping、moved to heap: n。",
			"把 newCounter 改成返回值 int（不取地址），还能看到 moved to heap 吗？说明堆分配是谁决定的。",
			"既然逃逸自动做，为什么还要关心？（提示：堆分配有 GC 成本，高频行情循环里 new 太多会拖慢）",
		},
		Check: "能说出「Go 里没有悬垂指针」的原因，并用 -gcflags=-m 验证一次逃逸",
		Run: func() {
			p1 := newCounter()
			*p1 = 42
			fmt.Printf("  newCounter() 返回后仍可用：%d（地址 %p）\n", *p1, p1)

			p2 := newCounter()
			*p2 = 7
			fmt.Printf("  第二次调用是新变量：%d vs %d，地址 %p vs %p\n", *p1, *p2, p1, p2)

			q := valueOf(100)
			fmt.Printf("  参数取地址也没问题：%d（编译期已决定是否逃逸）\n", *q)
			fmt.Println("  自己动手：go build -gcflags='-m' . 2>&1 | grep escape")
		},
	}
}
