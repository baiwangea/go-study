package interfaces

import (
	"fmt"
	"math"

	"go-study/go-fundamentals/level"
)

// Shape 是本关复用的接口：只要有 Area 方法就算形状。
type Shape interface{ Area() float64 }

// Rect 除了 Area，还额外实现了 fmt.Stringer（有个 String() 方法）。
type Rect struct{ Width, Height float64 }

func (r Rect) Area() float64  { return r.Width * r.Height }
func (r Rect) String() string { return fmt.Sprintf("Rect{%.1f x %.1f}", r.Width, r.Height) }

// Circle 只实现了 Area。
type Circle struct{ Radius float64 }

func (c Circle) Area() float64 { return math.Pi * c.Radius * c.Radius }

// measure 接收接口类型：它不知道也不关心具体是哪种形状。
func measure(s Shape) {
	fmt.Printf("  measure(%-22v) Area=%.2f\n", s, s.Area())
}

// describe 用类型断言把接口值还原成具体类型。
func describe(s Shape) string {
	switch v := s.(type) {
	case Rect:
		return fmt.Sprintf("矩形，面积 %.2f", v.Area())
	case Circle:
		return fmt.Sprintf("圆形，面积 %.2f", v.Area())
	default:
		return fmt.Sprintf("未知形状 %T", v)
	}
}

// L10 第十关：多态、类型断言与 fmt.Stringer。
func L10() level.Level {
	return level.Level{
		ID:      "L1-10",
		Title:   "多态与类型断言",
		Tags:    "polymorphism · type switch · Stringer",
		Pre:     "L1-09",
		Goal:    "学会用接口做参数解耦，并在需要时用类型断言取回具体类型",
		Observe: "同一个 measure 函数处理两种类型；%v 打印 Rect 时自动调用了它的 String()",
		Questions: []string{
			"给 Circle 也加一个 String() 方法，输出会怎样变化？这说明 fmt 包在做什么？",
			"s.(Rect) 断言失败会 panic，s.(Rect) 的 comma-ok 写法怎么避免？两种写法分别什么时候用？",
			"如果一个接口有 10 个方法，调用方真的都需要吗？（提示：接口应尽量小，「接受接口，返回结构体」）",
		},
		Check: "能写出接收接口的函数，并用 type switch 区分具体类型；能说清 Stringer 的魔法在哪",
		Run: func() {
			r := Rect{Width: 10, Height: 5}
			c := Circle{Radius: 5}

			measure(r)
			measure(c)

			fmt.Println("  describe(r) =", describe(r))
			fmt.Println("  describe(c) =", describe(c))
			fmt.Printf("  直接打印 Rect 得到 %v —— 因为 Rect 满足 fmt.Stringer\n", r)

			var s Shape = r
			v, ok := s.(Rect) // comma-ok 写法：断言失败不 panic，只得到 ok == false
			fmt.Printf("  类型断言：s.(Rect) → ok=%v, v=%v\n", ok, v)
		},
	}
}
