//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 01-go-fundamentals && go run composition/01_embedding.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import "fmt"

/*
═══════════════════════════════════════════════════════════════════════
L1-26：结构体嵌入与方法提升
═══════════════════════════════════════════════════════════════════════

【核心概念】
    Go 没有继承，但有一个更朴素的复用手段：**嵌入（embedding）**。
    把一个类型**不带字段名**地放进结构体里，就叫嵌入：

        type Service struct {
            BaseLogger   // 匿名字段，字段名自动就是 BaseLogger
            Name string
        }

    两个直接后果：
    · **方法提升**：BaseLogger 的方法会「提升」到 Service 上，
      s.Log("x") 等价于 s.BaseLogger.Log("x")，不用手写转发方法。
    · **字段名就是类型名**：s.BaseLogger 就是那个匿名字段，
      可以显式写 s.BaseLogger.Log("x")。

    ⚠️ 嵌入**不是继承**，三点必须记牢：
    1. 没有多态：func use(b BaseLogger) 不能接受 Service，必须写 use(s.BaseLogger)。
    2. 没有方法重写：Service 定义同名方法只是**遮蔽（shadow）**，不是覆盖。
       BaseLogger 内部调用自己的方法时，永远走 BaseLogger 的版本 ——
       因为 Go 的方法调用是静态确定的，没有虚函数表。
    3. 没有向上转型：Service 不是 BaseLogger 的子类，两者只是「有」的关系。

    字段名冲突规则（选择器解析，由浅入深）：
    · 同一层有同名成员 → 编译错误（ambiguous selector），必须写全路径
    · 外层与外层嵌入的类型同名 → 外层的遮蔽内层的（浅的优先）

    嵌入值 vs 嵌入指针：
    · 嵌入值 BaseLogger  → 每个 Service 各持有一份独立的副本
    · 嵌入指针 *BaseLogger → 多个 Service 共享同一个对象；零值是 nil，
      调用需要解引用接收者的方法会 panic

【你会看到】
    ✓ s.Log("hi") 不用写转发方法就能用（方法提升）
    ✓ d.CallGreet() 打印的是 Base.Greet，而不是 Derived.Greet（没有多态）
    ✓ useBase(s.BaseLogger) 必须显式取出内层，不能直接传 s
    ✓ u.ID 取外层，u.Meta.ID 取内层（遮蔽规则）
    ✓ 嵌入指针时改的是同一个对象，嵌入值改的是副本

【思考题】（动手改代码验证，别空想）
    1. Java 里 `class Service extends BaseLogger` 会得到 is-a 关系，
       可以赋值给父类引用；Go 的嵌入只有 has-a。这对「面向接口编程」有什么影响？
       （提示：Go 用接口而不是继承来做多态，见 L1-10）
    2. 如果 Service 定义了 Log，而某个函数接收 Loggable 接口，s.Log 会走哪个版本？
       那 s.BaseLogger.Log 呢？动手打印验证。
    3. 嵌入两个都有 ID 字段的结构体，u.ID 会怎样？把外层 ID 删掉再试一次。
    4. 嵌入 *BaseLogger 时，如果它一直是 nil，调用 Log 会怎样？动手试试
       （提示：方法里只要不解引用接收者就不会 panic —— 想想为什么）。

【通关标准】
    能说清「方法提升」和「遮蔽」的区别，并能解释为什么嵌入不能当继承用。

【前置】L1-09              【标签】embedding · 方法提升 · 组合 · has-a
═══════════════════════════════════════════════════════════════════════
*/

// BaseLogger 是被嵌入的「基类」：注意它只是普通结构体，没有任何继承语义。
type BaseLogger struct {
	Prefix string
}

// Log 会因嵌入而被提升到外层类型上。
func (b BaseLogger) Log(msg string) string {
	return b.Prefix + " " + msg
}

// Kind 也是被提升的方法。
func (b BaseLogger) Kind() string { return "BaseLogger" }

// CallOwnLog 用来证明「没有多态」：它在 BaseLogger 内部调用自己的 Log，
// 即使外层类型定义了同名 Log，这里也永远走 BaseLogger.Log。
// Java/C++ 的虚函数会分发到子类，Go 不会 —— 因为 Go 没有虚函数表。
func (b BaseLogger) CallOwnLog(msg string) string {
	return "BaseLogger 内部调用自己的 Log → " + b.Log(msg)
}

// Loggable 是接口，用来演示「方法提升让外层类型也满足接口」。
type Loggable interface{ Log(string) string }

// Service 嵌入 BaseLogger（匿名字段），于是自动获得 Log / Kind / CallOwnLog。
type Service struct {
	BaseLogger // ← 匿名字段：没有字段名，字段名就是类型名
	Name       string
}

// Service 自己定义 Log：这不是「重写」，只是**遮蔽**了提升上来的那个。
func (s Service) Log(msg string) string {
	return "[Service:" + s.Name + "] " + msg
}

// ── 编译期断言：靠方法提升，Service 也满足 Loggable ─────────────────────
var (
	_ Loggable = BaseLogger{}
	_ Loggable = Service{} // Service 自己有 Log，当然满足
)

// useBase 接收 BaseLogger。注意它**不能**接收 Service —— 嵌入不是继承。
// 想传 Service 必须写 useBase(s.BaseLogger)。
func useBase(b BaseLogger) string { return b.Kind() + " says " + b.Log("hello") }

// ── 字段名冲突演示 ──────────────────────────────────────────────────────
type Meta struct {
	ID      int // 与外层 User.ID 同名
	Created string
}

// User 同时有外层 ID 和（通过嵌入带来的）Meta.ID。
// 规则：外层优先，u.ID 取外层那个；想拿内层必须写 u.Meta.ID。
type User struct {
	Meta
	ID   int // 遮蔽掉 Meta.ID（选择器只看深度，浅的赢）
	Name string
}

// ── 嵌入值 vs 嵌入指针 ──────────────────────────────────────────────────
// Counter 用指针接收者，方便观察「改的是谁」。
type Counter struct{ N int }

func (c *Counter) Inc() { c.N++ }

// HolderValue 嵌入**值**：每个 HolderValue 各有一份独立的 Counter 副本。
type HolderValue struct{ Counter }

// HolderPtr 嵌入**指针**：多个 HolderPtr 可以共享同一个 Counter。
type HolderPtr struct{ *Counter }

func main() {
	fmt.Println("\n=== L1-26 结构体嵌入与方法提升 ===")

	// ── 示例 1：方法提升 ────────────────────────────────────────────────
	fmt.Println("\n【示例 1】方法提升：不用写转发方法")
	s := Service{BaseLogger: BaseLogger{Prefix: "[info]"}, Name: "api"}
	fmt.Printf("  s.Log(\"started\")            = %q  ← 走 Service 自己的 Log\n", s.Log("started"))
	fmt.Printf("  s.BaseLogger.Log(\"started\") = %q  ← 显式走内层\n", s.BaseLogger.Log("started"))
	fmt.Printf("  s.Kind()                     = %q  ← Service 没定义 Kind，用的是提升上来的\n", s.Kind())
	fmt.Printf("  s.BaseLogger 就是匿名字段本身：%q\n", s.BaseLogger.Prefix)
	fmt.Println("  ↑ 提升是编译期语法糖：s.Kind() 完全等价于 s.BaseLogger.Kind()")

	// ── 示例 2：嵌入不是继承（没有方法重写/多态）─────────────────────────
	fmt.Println("\n【示例 2】嵌入不是继承：遮蔽 ≠ 重写")
	fmt.Printf("  s.Log(\"x\")               = %q  ← 外层遮蔽了内层\n", s.Log("x"))
	fmt.Printf("  s.BaseLogger.Log(\"x\")    = %q  ← 内层还在，随时能调\n", s.BaseLogger.Log("x"))
	fmt.Printf("  s.CallOwnLog(\"x\")        = %q\n", s.CallOwnLog("x"))
	fmt.Println("  ↑ 关键：CallOwnLog 是 BaseLogger 的方法，它内部调的是 **BaseLogger.Log**，")
	fmt.Println("    不会分发到 Service.Log —— Go 没有虚函数表，这就是「没有多态」的证据")
	fmt.Println("  ↑ Java 里同样的写法会输出 [Service:api] x，因为 Java 的方法是虚调用")

	// ── 示例 3：没有向上转型，接口才是多态的正道 ────────────────────────
	fmt.Println("\n【示例 3】useBase 只能收 BaseLogger，收不了 Service")
	fmt.Printf("  useBase(s.BaseLogger) = %q\n", useBase(s.BaseLogger))
	// useBase(s) ← 这行编译不过：cannot use s (Service) as BaseLogger value
	fmt.Println("  ↑ 必须显式取出内层。Go 用接口做多态：")
	var l Loggable = s // 靠 Service 自己的 Log 满足接口
	fmt.Printf("     var l Loggable = s; l.Log(\"via interface\") = %q\n", l.Log("via interface"))
	l = s.BaseLogger
	fmt.Printf("     l = s.BaseLogger;  l.Log(\"via interface\") = %q  ← 接口分发到内层版本\n", l.Log("via interface"))

	// ── 示例 4：字段名冲突（遮蔽规则）───────────────────────────────────
	fmt.Println("\n【示例 4】字段名冲突：外层优先，内层要写全路径")
	u := User{
		Meta: Meta{ID: 100, Created: "2026-10-12"}, // 内层 ID
		ID:   7,                                    // 外层 ID
		Name: "alice",
	}
	fmt.Printf("  u.ID        = %d  ← 取外层（浅的赢）\n", u.ID)
	fmt.Printf("  u.Meta.ID   = %d  ← 显式取内层\n", u.Meta.ID)
	fmt.Printf("  u.Created   = %q  ← 提升的字段也能直接用\n", u.Created)
	fmt.Println("  ↑ 如果两个嵌入类型都有同名字段，u.ID 会报 ambiguous selector，必须写全路径")

	// ── 示例 5：嵌入值 vs 嵌入指针 ──────────────────────────────────────
	fmt.Println("\n【示例 5】嵌入值 = 副本，嵌入指针 = 共享")
	hv := HolderValue{}
	hv.Inc()
	hv.Inc()
	fmt.Printf("  嵌入值 HolderValue：hv.Counter.N = %d（副本被改了）\n", hv.Counter.N)

	shared := &Counter{}
	hp1 := HolderPtr{Counter: shared}
	hp2 := HolderPtr{Counter: shared}
	hp1.Inc()
	hp2.Inc()
	fmt.Printf("  嵌入指针 HolderPtr：hp1 和 hp2 共享同一个 Counter，N = %d\n", shared.N)
	fmt.Println("  ↑ 改任一个 HolderPtr 都会影响另一个（因为里面存的是同一个指针）")

	// 陷阱：嵌入指针的零值是 nil，解引用就 panic
	func() {
		defer func() {
			fmt.Printf("  嵌入指针零值调用 Inc() → panic: %v\n", recover())
		}()
		var bad HolderPtr // Counter 是 nil
		bad.Inc()         // Inc 里要改 c.N，必须解引用 nil → panic
	}()

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n【关键总结】")
	fmt.Println("  • 匿名字段 = 嵌入，字段名就是类型名，方法会被提升到外层")
	fmt.Println("  • 提升只是语法糖，s.M() 等价于 s.Inner.M()，编译期就确定了")
	fmt.Println("  • 嵌入不是继承：没有向上转型、没有虚函数、同名方法只是遮蔽")
	fmt.Println("  • 内层方法内部调用自己的方法，永远不会分发到外层的同名方法")
	fmt.Println("  • 选择器由浅到深：外层遮蔽内层；同层冲突则编译报错，必须写全路径")
	fmt.Println("  • 嵌入值各持一份副本；嵌入指针共享对象，但零值是 nil 会 panic")
}
