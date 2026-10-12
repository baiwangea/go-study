//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 01-go-fundamentals && go run errors/03_panic_recover.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

/*
═══════════════════════════════════════════════════════════════════════
L1-21：panic 与 recover 的边界
═══════════════════════════════════════════════════════════════════════

【核心概念】
    panic  = 抛出；recover = 捕获。但它们**不是** try/catch：

    1. recover 只在 **defer 里的直接调用**才有效。
       写在普通代码里、或者包了一层别的函数调用，都只会返回 nil。
    2. panic(nil) 是个坑：Go 1.21 之前 recover() 拿到 nil，
       调用方会误以为「没发生 panic」；Go 1.21 起会包装成
       *runtime.PanicNilError，于是 r != nil 判断终于可靠了。
    3. **跨 goroutine 的 panic 无法被 recover。**
       recover 只对自己所在的 goroutine 有效；别的 goroutine 里 panic 且没被
       自己拦住，整个进程直接退出（退出码 2），堆栈打到 stderr。
       本关用一个子进程真跑一遍给你看，主进程本身照常退出 0。

    该用哪个？
    · 返回 error：一切**可预期的失败** —— 文件不存在、网络超时、参数非法。
    · panic：**不可恢复的编程错误** —— 初始化配置缺失、不可能出现的分支、
      nil 参数、除零。这类问题继续跑下去只会产生更诡异的错误，早点炸掉更好。

    库代码尤其不要跨 API 边界 panic：调用方看不到你的文档时，
    一个 panic 会让整个服务挂掉。要么返回 error，要么在边界用 recover 转成 error。

【你会看到】
    ✓ 普通代码里调 recover() 得到 <nil>（无效）
    ✓ defer 里的 recover() 成功拦下「除零 panic」并转成 error
    ✓ panic(nil) 被 recover 后得到 *runtime.PanicNilError（r != nil）
    ✓ 子进程里跨 goroutine panic → 退出码 2、堆栈在 stderr；父进程退出码仍是 0

【思考题】（动手改代码验证，别空想）
    1. Java 的 catch(Throwable t) 能跨线程吗？Go 的 recover 不能跨 goroutine，
       这算是缺点还是「让你不敢乱起 goroutine 然后忘记兜底」的设计？
    2. 把 catchPanicNil 里的 panic(nil) 改成 panic("真正的错误")，输出有什么变化？
       再去读一下 runtime.PanicNilError 的文档，理解 Go 1.21 为什么要改这个行为。
    3. 在子 goroutine 的第一行写 `defer func(){ recover() }()`，子进程还会崩吗？
       动手改 childGoroutinePanic 验证。
    4. 为什么 json.Unmarshal、database/sql 这类库内部用 panic 实现递归，
       却对外只返回 error？（提示：API 边界用 recover 兜住）

【通关标准】
    能说出 recover 的三个限制，并会写「goroutine 入口先 defer recover」这个保命模板。

【前置】L1-18 · L1-19       【标签】panic · recover · goroutine · 退出码
═══════════════════════════════════════════════════════════════════════
*/

// childFlag 是内部约定：带上这个参数运行时，本程序扮演「会被 panic 打死的子进程」。
// 这样演示跨 goroutine panic 时，牺牲的是子进程，主进程还能正常打印结论并退出 0。
const childFlag = "--child-goroutine-panic"

// safeDivide 演示「用 recover 把 panic 转成 error」的标准模板。
// 注意 recover 必须在 defer 的匿名函数里直接调用，中间不能再包一层。
func safeDivide(a, b int) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			// 把 panic 值转成普通 error 交回调用方 —— 这就是「库的边界」
			err = fmt.Errorf("safeDivide 捕获到 panic: %v", r)
		}
	}()
	return a / b, nil // b == 0 时这里会 panic: integer divide by zero
}

// catchPanicNil 专门演示 panic(nil) 的行为。
// 命名返回值 r 让 defer 能把 recover 的结果带出去。
func catchPanicNil() (r any) {
	defer func() { r = recover() }()
	panic(nil)
}

// badRecover 演示错误写法：recover 被包在另一层函数里，
// 中间多了一层调用栈，recover 就「看不到」panic 了 —— 它只会返回 nil，
// 而 panic 会**继续向外传播**（这里由 outerCatchBadRecover 兜住，否则进程就挂了）。
func badRecover() {
	defer func() { _ = indirectRecover() }() // 不是 defer 里直接调用 recover()
	panic("这个 panic 不会被间接调用拦住")
}

// indirectRecover 是 badRecover 的帮凶：recover 在这里调用，不在 defer 直接调用位置。
func indirectRecover() any { return recover() }

// outerCatchBadRecover 在外层用正确的写法接住 badRecover 逃逸出来的 panic，
// 这样我们才能安全地观察到「间接调用 recover 无效」这个事实。
func outerCatchBadRecover() (escaped any) {
	defer func() { escaped = recover() }() // 正确的写法：defer 里直接调用
	badRecover()
	return nil
}

// guardedGoroutine 是保命模板：每个 goroutine 的入口先挂一个 defer recover，
// 这样即使内部炸了，也只影响这一个 goroutine，不会拖垮整个进程。
func guardedGoroutine(name string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("  [%s] 自己拦住了 panic：%v（进程还活着）\n", name, r)
		}
	}()
	panic(name + " 内部出错")
}

// childGoroutinePanic 只会在子进程里运行。
// 它故意起一个没有 recover 的 goroutine，让 panic 把整个进程带走。
func childGoroutinePanic() {
	fmt.Println("  [子进程] 主 goroutine 健在，另起一个 goroutine 并在里面 panic（无 recover）")
	go func() {
		panic("子 goroutine 里的 panic：没人能救它")
	}()
	time.Sleep(300 * time.Millisecond) // 留点时间让 panic 发生
	fmt.Println("  [子进程] 这行永远不会打印：整个进程已经被那个 panic 终止了")
}

// mustParsePort 是「初始化期 fail fast」的典型：配置错了就该当场炸掉。
// 这种场景 panic 是合适的 —— 继续跑下去毫无意义。
func mustParsePort(s string) int {
	p, err := strconv.Atoi(s)
	if err != nil {
		panic(fmt.Sprintf("配置项 PORT=%q 不是合法端口: %v", s, err))
	}
	return p
}

// parsePort 是「处理外部输入」的典型：用户可能输错，必须返回 error 而不是 panic。
func parsePort(s string) (int, error) {
	p, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("端口 %q 非法: %w", s, err)
	}
	if p < 1 || p > 65535 {
		return 0, fmt.Errorf("端口 %d 超出范围 1-65535", p)
	}
	return p, nil
}

func main() {
	// 子进程模式：演示「跨 goroutine panic 会把进程打死」，然后直接退出。
	if len(os.Args) > 1 && os.Args[1] == childFlag {
		childGoroutinePanic()
		return
	}

	fmt.Println("\n=== L1-21 panic 与 recover 的边界 ===")

	// ── 示例 1：recover 只在 defer 里有效 ────────────────────────────────
	// 直接调用 recover() 只是普通函数调用，没有任何 panic 上下文，返回 nil。
	fmt.Println("\n【示例 1】recover 只在 defer 里生效")
	fmt.Printf("  普通代码里 recover() = %v  ← 无效，只是个返回 nil 的函数\n", recover())
	fmt.Printf("  包一层函数调用 recover() 的结果 = %v  ← 拦不住，panic 逃逸到外层被接住：%v\n",
		indirectRecover(), outerCatchBadRecover())
	fmt.Println("  ↑ 记住：只有「defer func(){ if r := recover(); r != nil {...} }()」这个形状能用")

	// ── 示例 2：用 recover 把 panic 转成 error ───────────────────────────
	// 整数除零在 Go 里是 panic，不是 error —— 这是「编程错误」，但库的边界要兜住。
	fmt.Println("\n【示例 2】defer + recover：panic → error")
	if v, err := safeDivide(10, 2); err == nil {
		fmt.Printf("  10 / 2 = %d（正常路径，recover 没被触发）\n", v)
	}
	v, err := safeDivide(10, 0)
	fmt.Printf("  10 / 0 = %d, err = %v\n", v, err)
	fmt.Println("  ↑ 调用方拿到的是普通 error，不需要知道底层发生过 panic")

	// ── 示例 3：panic(nil) 的坑 ──────────────────────────────────────────
	fmt.Println("\n【示例 3】panic(nil)：Go 1.21 之后才不再骗人")
	r := catchPanicNil()
	fmt.Printf("  recover() 的返回值 = %v\n", r)
	fmt.Printf("  类型 = %T\n", r)
	fmt.Printf("  r != nil ？ %v  ← 非 nil，判断可靠\n", r != nil)
	fmt.Println("  ⚠️  Go 1.21 之前这里是 nil，用 r != nil 判断会误以为「没发生 panic」")

	// ── 示例 4：跨 goroutine panic 无法 recover ─────────────────────────
	// 本进程自己不能 panic（会挂），所以起一个子进程去死，我们看它的退出码。
	fmt.Println("\n【示例 4】跨 goroutine 的 panic：recover 救不了，进程直接死")
	fmt.Println("  （下面启动一个子进程做实验，父进程不受影响）")
	cmd := exec.Command(os.Args[0], childFlag)
	out, cmdErr := cmd.CombinedOutput()
	fmt.Printf("  子进程退出码 = %d  ← 非 0，说明整个进程被 panic 终止了\n", cmd.ProcessState.ExitCode())
	fmt.Printf("  返回的错误类型 = %T\n", cmdErr)
	fmt.Println("  子进程 stderr 摘录：")
	printed := 0
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) == "" {
			continue // 跳过空行，只摘关键几行
		}
		if printed == 3 {
			fmt.Println("    ...（后面是完整调用栈）")
			break
		}
		fmt.Printf("    %s\n", line)
		printed++
	}

	// ── 示例 5：正确姿势 —— 每个 goroutine 自己兜底 ─────────────────────
	fmt.Println("\n【示例 5】保命模板：goroutine 入口先 defer recover")
	go guardedGoroutine("worker-1")
	time.Sleep(50 * time.Millisecond) // 等 goroutine 跑完
	fmt.Println("  ↑ 这个 goroutine 炸了，但进程毫发无伤")

	// ── 示例 6：什么时候 panic，什么时候返回 error ──────────────────────
	fmt.Println("\n【示例 6】panic 还是 error？看「是不是可预期的失败」")
	fmt.Println("  初始化期配置错误 → panic（fail fast）：")
	func() {
		defer func() { fmt.Printf("    mustParsePort(\"abc\") 触发了 panic：%v\n", recover()) }()
		mustParsePort("abc")
	}()
	fmt.Println("  处理外部输入 → 返回 error：")
	for _, s := range []string{"8080", "abc", "99999"} {
		if p, err := parsePort(s); err != nil {
			fmt.Printf("    parsePort(%q) → err = %v\n", s, err)
		} else {
			fmt.Printf("    parsePort(%q) → %d ✓\n", s, p)
		}
	}

	// ── 示例 7：recover 之后可以重新 panic（不想吞掉错误时）─────────────
	fmt.Println("\n【示例 7】re-panic：记录日志后再抛出去")
	func() {
		defer func() {
			if rr := recover(); rr != nil {
				fmt.Printf("    先记录日志：发生 %v\n", rr)
				// 处理完再抛，让更上层决定 —— 别静默吞掉 panic
				fmt.Println("    再包一层错误抛出去（这里为了演示只打印）")
			}
		}()
		panic(errors.New("底层 goroutine 池满了"))
	}()

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n【关键总结】")
	fmt.Println("  • recover 只在 defer 里被直接调用才有效，普通代码里返回 nil")
	fmt.Println("  • panic(nil) 在 Go 1.21+ 变成 *runtime.PanicNilError，r != nil 才可靠")
	fmt.Println("  • recover 不跨 goroutine：别人的 goroutine panic，你的进程直接退出（码 2）")
	fmt.Println("  • 每个 goroutine 入口先写 defer recover，这是并发代码的保命模板")
	fmt.Println("  • 可预期的失败返回 error；不可恢复的编程错误才 panic")
	fmt.Println("  • 库不要在 API 边界外泄 panic，边界处用 recover 转成 error")
}
