package interfaces

import (
	"fmt"
	"math"

	"go-study/go-fundamentals/level"
)

// Shaper 要求实现 Area() float64。
type Shaper interface {
	Area() float64
}

// Square 用值接收者实现 Shaper。
type Square struct{ Side float64 }

func (s Square) Area() float64 { return s.Side * s.Side }

// Ring 用指针接收者实现 Shaper —— 只有 *Ring 才满足接口。
type Ring struct{ Radius float64 }

func (r *Ring) Area() float64 { return math.Pi * r.Radius * r.Radius }

// Blob 故意只实现了 Perimeter，没有 Area，因此不是 Shaper。
type Blob struct{ Base float64 }

func (b Blob) Perimeter() float64 { return 4 * b.Base }

// 编译期断言：把「谁实现了接口」写成代码，接口一改就立刻在这里报错。
var (
	_ Shaper = Square{}
	_ Shaper = (*Ring)(nil)
	// _ Shaper = Blob{} // 打开这行：cannot use Blob as Shaper (missing method Area)
)

// L09 第九关：隐式实现与「接收者类型」的坑。
func L09() level.Level {
	return level.Level{
		ID:      "L1-09",
		Title:   "隐式实现与接收者",
		Tags:    "implements · 值/指针接收者",
		Pre:     "L1-08",
		Goal:    "理解 Go 没有 implements 关键字，以及指针接收者导致只有 *T 满足接口",
		Observe: "Square 与 *Ring 都能装进 Shaper；Blob 装不进去（注释掉的那行会报错）",
		Questions: []string{
			"把 _ Shaper = (*Ring)(nil) 改成 _ Shaper = Ring{}，报错是什么？为什么？",
			"var s Shaper = Ring{...} 也编译不过，但 s := &Ring{...} 却可以 —— 说清规则",
			"上面的 var _ Shaper = ... 这种「空白变量断言」在标准库里很常见，它解决什么问题？",
		},
		Check: "能解释值接收者/指针接收者对接口实现的影响，并写出编译期断言",
		Run: func() {
			shapes := []Shaper{Square{Side: 3}, &Ring{Radius: 2}}
			for _, sh := range shapes {
				fmt.Printf("  %-18s Area = %.2f\n", fmt.Sprintf("%T", sh), sh.Area())
			}
			b := Blob{Base: 5}
			fmt.Printf("  Blob 只有 Perimeter()=%.0f —— 它不是 Shaper，装不进上面的切片\n", b.Perimeter())
		},
	}
}
