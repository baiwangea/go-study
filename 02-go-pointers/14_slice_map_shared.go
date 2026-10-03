package main

import (
	"fmt"

	"go-study/go-pointers/level"
)

// L1-14：切片头是值拷贝，底层数组共享；map 是引用类型。
func L14() level.Level {
	return level.Level{
		ID:      "L1-14",
		Title:   "切片的共享与失联",
		Tags:    "slice header · append 扩容 · map 引用",
		Pre:     "L1-12",
		Goal:    "理解切片 = (指针, len, cap) 三元组，赋值只拷贝三元组，底层数组仍是共享的",
		Observe: "改 s2[0] 会影响 s1；append 触发扩容后两者失联；map 赋值后两边同步变化",
		Questions: []string{
			"s3 := append(s2, 999) 之后 s1 的第三个元素变成了什么？为什么？（还没扩容，写进了同一块底层数组）",
			"打印每次 append 前后的 cap，找出「什么时候开始换新数组」。",
			"把函数参数写成 func fix(s []int) { s[0] = 0 }，调用后原切片变吗？写成 append 呢？（这就是「切片当参数」的经典歧义）",
			"PHP 数组按值传递、JS 数组按引用，Go 的 slice 为什么两者都像又不完全是？",
		},
		Check: "能用三元组模型解释任何切片共享/失联现象，并知道 map 需要 make 才能写",
		Run: func() {
			s1 := []int{1, 2, 3}
			s2 := s1[:2]
			s2[0] = 100
			fmt.Printf("  s2[0]=100 后 s1=%v（len %d cap %d）← 共享底层数组\n", s1, len(s1), cap(s1))

			s3 := append(s2, 999) // 未扩容：直接占用 s1 的第三个位置
			fmt.Printf("  append(s2,999) 后 s3=%v，s1=%v ← 竟互相影响\n", s3, s1)

			big := append(s3, 1, 2, 3, 4, 5) // 容量不足，扩容到新数组
			big[0] = -1
			fmt.Printf("  扩容后改 big[0]=-1，s1=%v（已失联），cap 从 %d 涨到 %d\n", s1, cap(s3), cap(big))

			m1 := map[string]int{"BTC": 1}
			m2 := m1
			m2["BTC"] = 2
			fmt.Printf("  map：改 m2 后 m1=%v ← map 本身就是指针语义（但必须 make 后才能写）\n", m1)
		},
	}
}
