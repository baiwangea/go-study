package main

import (
	"fmt"

	"go-study/go-pointers/level"
)

type Position struct {
	Symbol string
	Size   float64
	Entry  float64
}

func (p Position) Label() string { return fmt.Sprintf("%s %.2f", p.Symbol, p.Size) }

func resizeCopy(p Position) { p.Size = 999 } // 值接收者：改副本
func resizePtr(p *Position) { p.Size = 999 } // 指针接收者：改原件

// L1-12：结构体拷贝成本与修改可见性。
func L12() level.Level {
	return level.Level{
		ID:      "L1-12",
		Title:   "结构体：拷贝 vs 指针",
		Tags:    "struct · 拷贝成本 · 接收者",
		Pre:     "L1-11",
		Goal:    "明白结构体传参是整个拷贝，需要就地修改或字段较多时改用指针",
		Observe: "resizeCopy 后 Size 仍是 0.5；resizePtr 后 Size 变 999",
		Questions: []string{
			"给 Position 再加 10 个字段，用 %p 打印传参前后的地址，观察拷贝发生在哪里。",
			"p.Label() 用值接收者可以调用，那 ptr := &pos; ptr.Label() 能编译吗？为什么 Go 允许这样自动解引用？",
			"机器人里持仓 Position 应该用值还是指针传？（考虑：会被修改吗？字段大吗？并发访问吗？）",
		},
		Check: "能按「是否要改原件 / 拷贝成本 / 是否需要 nil」三条准则选值或指针",
		Run: func() {
			pos := Position{Symbol: "BTC", Size: 0.5, Entry: 63000}

			resizeCopy(pos)
			fmt.Printf("  resizeCopy 后：Size=%.2f（未变）%s\n", pos.Size, pos.Label())

			resizePtr(&pos)
			fmt.Printf("  resizePtr 后：Size=%.2f（已改）\n", pos.Size)

			copy2 := pos // 整体拷贝，改 copy2 不影响 pos
			copy2.Symbol = "ETH"
			fmt.Printf("  赋值 copy2 := pos 后改 copy2.Symbol：%q vs pos.Symbol=%q\n", copy2.Symbol, pos.Symbol)
		},
	}
}
