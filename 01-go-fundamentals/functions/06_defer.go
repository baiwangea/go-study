//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 01-go-fundamentals && go run functions/06_defer.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"errors"
	"fmt"
)

/*
═══════════════════════════════════════════════════════════════════════
L1-18：defer 的语义与陷阱
═══════════════════════════════════════════════════════════════════════

【核心概念】
    defer 把一次调用「预约」到当前函数 return 之前执行。三条铁律：

    1. 后进先出（LIFO）：多个 defer 按注册的逆序执行，像叠盘子。
    2. 参数在 defer 语句执行那一刻就求值完毕，之后变量再变也不影响。
       这一点和 Java 的 try-with-resources、Python 的 with 都不一样：
       它们是「延迟执行整段逻辑」，Go 是「延迟调用，但参数立即快照」。
    3. 如果函数有**命名返回值**，defer 里可以读写它 —— 返回值是最后一次被改写的结果。
       这就是「用 defer 兜底修改 err」的经典写法（也是 recover 能改返回值的原理）。

    常见坑：在 for 循环里 defer，defer 不会在本次迭代结束时执行，
    而是**全部堆到函数返回时**才按逆序执行 —— 资源会一直不释放。

【你会看到】
    ✓ 三个 defer 逆序打印 3 2 1
    ✓ defer 里打印的 i 是 0，尽管函数结束前 i 已经变成 100（参数快照）
    ✓ 命名返回值被 defer 改写，返回的是被包装后的错误
    ✓ 循环里 defer：3 个资源全部「打开」完才逐个「关闭」（逆序）

【思考题】（动手改代码验证，别空想）
    1. Java 的 try-with-resources / Python 的 with 是「作用域结束就关」，
       Go 的 defer 是「函数返回才关」。写一个大循环处理 1 万个文件时，
       这个差异会带来什么后果？怎么修？
    2. 把 `defer fmt.Println(i)` 换成 `defer func() { fmt.Println(i) }()`
       输出会变成什么？为什么？（提示：闭包捕获变量 vs 参数求值）
    3. `func f() (err error)` 里 defer 把 err 包了一层，
       如果函数体里写的是 `return errors.New("x")` 而不是 `return`，
       defer 还能改到吗？动手试试。
    4. os.Exit(0) 和 log.Fatal 会执行 defer 吗？（提示：它们不 return，直接结束进程）

【通关标准】
    能说出 defer 的三条铁律，并能手写「命名返回值 + defer 包装错误」这个惯用法。

【前置】L1-02              【标签】defer · LIFO · 命名返回值 · 资源释放
═══════════════════════════════════════════════════════════════════════
*/

// withNamedReturn 演示「defer 修改命名返回值」这个惯用法。
// 注意返回值列表里写的是 (err error) —— 有名字，才叫命名返回值，
// 函数体内的 err 就是这个变量本身，defer 闭包能读写它。
func withNamedReturn(fail bool) (err error) {
	// defer 注册时先不求值（这里没有参数），等到函数 return 前才执行闭包。
	defer func() {
		if err != nil {
			// 改的是命名返回值 err 本身，所以调用方拿到的是包装后的错误
			err = fmt.Errorf("withNamedReturn 包装层: %w", err)
		}
	}()

	if fail {
		return errors.New("底层错误")
	}
	return nil
}

// noNameReturn 是对照组：返回值没有名字，defer 就无从下手改它。
// 这里的 err 是**局部变量**，defer 改它不影响返回值。
func noNameReturn(fail bool) error {
	var err error
	defer func() {
		if err != nil {
			err = fmt.Errorf("这层包装会被丢弃: %w", err) // 白改了
		}
	}()

	if fail {
		return errors.New("底层错误")
	}
	return nil
}

// openResources 演示「循环里 defer」的错误姿势。
// defer 只是登记，函数不返回就不执行 —— 所以三个资源会全部打开完，
// 最后才在函数退出时按 LIFO 逆序关闭。真实项目里这就是「文件句柄耗尽」。
func openResources(names []string) {
	for _, name := range names {
		fmt.Printf("  打开 %s（defer 登记，但不会立刻执行）\n", name)
		defer fmt.Printf("  关闭 %s\n", name)
	}
	fmt.Println("  ↑ 循环结束了，但一个都还没关；现在函数要返回，才开始逆序关闭")
}

// openResourcesFixed 是正确姿势：把「一次迭代」包成一个函数，
// 每次调用结束就触发那一层的 defer，资源立刻释放。
// 多出的这层匿名函数，就是给 defer 制造一个「函数边界」。
func openResourcesFixed(names []string) {
	for _, name := range names {
		func() {
			fmt.Printf("  打开 %s\n", name)
			defer fmt.Printf("  关闭 %s\n", name)
		}() // 调用结束 = 这一层 return = 立即执行 defer
	}
}

func main() {
	fmt.Println("\n=== L1-18 defer 的语义与陷阱 ===")

	// ── 示例 1：LIFO —— 后进先出 ─────────────────────────────────────────
	// 三个 defer 依次登记，函数返回时按相反顺序执行，像叠盘子取盘子。
	fmt.Println("\n【示例 1】LIFO：后注册的先执行")
	func() {
		defer fmt.Println("  第 1 个 defer 登记（最后执行）")
		defer fmt.Println("  第 2 个 defer 登记（中间执行）")
		defer fmt.Println("  第 3 个 defer 登记（最先执行）")
		fmt.Println("  函数体结束，准备 return")
	}()
	fmt.Println("  ↑ 打印顺序是 3 → 2 → 1，这就是 LIFO")

	// ── 示例 2：参数在 defer 语句执行时求值 ──────────────────────────────
	// 这是最反直觉的一条：defer 后面的函数调用，参数**当场**就算好存起来，
	// 真正执行时用的还是那个旧值。
	fmt.Println("\n【示例 2】参数在 defer 那一刻求值（快照）")
	func() {
		i := 0
		defer fmt.Printf("  defer 里看到 i = %d（快照值）\n", i) // 此刻 i 是 0，就存 0
		i = 100                                          // 改了也没用，快照已经存好了
		fmt.Println("  函数体把 i 改成了 100")
	}()
	fmt.Println("  ↑ 如果换成 defer func(){ fmt.Println(i) }() 就会打印 100 —— 闭包捕获的是变量")

	// ── 示例 3：defer 修改命名返回值 ─────────────────────────────────────
	// 有名字的返回值 = 函数作用域里的一个变量，defer 闭包能改它。
	// 这是 Go 里「统一在一处包装/记录错误」的惯用法。
	fmt.Println("\n【示例 3】命名返回值可以被 defer 改写")
	fmt.Printf("  withNamedReturn(true)  → %v\n", withNamedReturn(true))
	fmt.Printf("  noNameReturn(true)     → %v  ← 注意：没有包装层，被丢弃了\n", noNameReturn(true))
	fmt.Println("  ↑ 同样想包装，命名返回值成功，匿名返回值失败 —— 差别只在 err 是不是那个变量")

	// ── 示例 4：循环里 defer 的坑 ────────────────────────────────────────
	fmt.Println("\n【示例 4】循环里 defer：全部堆到函数返回时才执行")
	openResources([]string{"a.log", "b.log", "c.log"})

	fmt.Println("\n  —— 修法：把每次迭代包进一个立即执行的匿名函数 ——")
	openResourcesFixed([]string{"a.log", "b.log", "c.log"})

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n【关键总结】")
	fmt.Println("  • defer 按 LIFO 逆序执行，适合「打开后立刻写 defer 关闭」")
	fmt.Println("  • 参数在 defer 语句处求值（快照）；要延迟取值就用闭包捕获变量")
	fmt.Println("  • 命名返回值能被 defer 改写 —— 包装错误、改返回值都靠它")
	fmt.Println("  • 循环里别裸写 defer，用匿名函数包一层，让每轮迭代各自归还资源")
	fmt.Println("  • os.Exit / log.Fatal 直接终止进程，defer 不会执行")
}
