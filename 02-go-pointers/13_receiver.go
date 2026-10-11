package main

import "fmt"

/*
═══════════════════════════════════════════════════════════════
L1-13：方法接收者 - 值 or 指针？
═══════════════════════════════════════════════════════════════
【核心概念】
要改状态就用指针接收者 (c *Counter)
不改状态可用值接收者 (c Counter)

【运行方式】
go run 13_receiver.go

【你会看到】
✓ IncValue 值接收者：连调三次都是 1（每次改副本）
✓ IncPtr 指针接收者：依次得到 1 2 3（真正累加）

【思考题】
1. 把 IncValue 改成 *Counter，调用 c.IncValue() 能编译吗？
2. 为什么不建议一个类型一半值接收者一半指针接收者？
3. sync.Mutex 为什么不能按值拷贝？

【接收者选择规则】
✓ 需要修改接收者 → 指针
✓ 接收者很大 → 指针
✓ 需要一致性（有一个是指针，其他也用指针）→ 指针
✓ 小的只读结构体 → 值
═══════════════════════════════════════════════════════════════
*/

type Counter struct {
	n int
}

// IncValue 值接收者：修改只作用于副本
func (c Counter) IncValue() int {
	c.n++ // 修改的是 c 的副本
	return c.n
}

// IncPtr 指针接收者：真正修改原对象
func (c *Counter) IncPtr() int {
	c.n++ // 通过指针修改原对象
	return c.n
}

// GetValue 读取计数（值接收者就够用）
func (c Counter) GetValue() int {
	return c.n
}

func main() {
	fmt.Println("\n╔═══════════════════════════════════════╗")
	fmt.Println("║  L1-13: 方法接收者 - 值 or 指针？   ║")
	fmt.Println("╚═══════════════════════════════════════╝\n")

	// ========== 示例 1：值接收者 ==========
	fmt.Println("【示例 1：值接收者（每次调用都拷贝接收者）】")
	c := Counter{n: 0}
	fmt.Printf("  初始值：c.n = %d\n", c.GetValue())

	fmt.Print("  调用 IncValue 三次：")
	for i := 0; i < 3; i++ {
		result := c.IncValue()
		fmt.Printf("%d ", result)
	}
	fmt.Println()
	fmt.Printf("  c.n 仍然是：%d (没变！因为每次改的是副本)\n\n", c.n)

	// ========== 示例 2：指针接收者 ==========
	fmt.Println("【示例 2：指针接收者（修改真正的对象）】")
	fmt.Print("  调用 IncPtr 三次：")
	for i := 0; i < 3; i++ {
		result := c.IncPtr() // Go 自动转换为 (&c).IncPtr()
		fmt.Printf("%d ", result)
	}
	fmt.Println()
	fmt.Printf("  c.n 现在是：%d (正常累加)\n\n", c.n)

	// ========== 示例 3：Go 的自动转换 ==========
	fmt.Println("【示例 3：Go 的自动转换（语法糖）】")

	// 值变量调用指针方法：Go 自动取地址
	c2 := Counter{n: 100}
	c2.IncPtr() // 等价于 (&c2).IncPtr()
	fmt.Printf("  值变量 c2 调用指针方法：c2.n = %d ✓\n", c2.n)

	// 指针变量调用值方法：Go 自动解引用
	c3 := &Counter{n: 200}
	val := c3.GetValue() // 等价于 (*c3).GetValue()
	fmt.Printf("  指针变量 c3 调用值方法：返回 %d ✓\n\n", val)

	// ========== 关键对比 ==========
	fmt.Println("【关键对比】")
	fmt.Println("  值接收者 func (c Counter)")
	fmt.Println("    • 接收整个结构体的副本")
	fmt.Println("    • 修改不影响原对象")
	fmt.Println("    • 适合小的只读操作")
	fmt.Println()
	fmt.Println("  指针接收者 func (c *Counter)")
	fmt.Println("    • 接收指向结构体的指针")
	fmt.Println("    • 可以修改原对象")
	fmt.Println("    • 避免大结构体拷贝")
	fmt.Println("    • 必须用于：需要修改状态、大结构体、包含 sync.Mutex")
}
