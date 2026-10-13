//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 01-go-fundamentals && go run errors/02_custom_error.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"errors"
	"fmt"
	"strings"
)

/*
═══════════════════════════════════════════════════════════════════════
L1-20：自定义错误与哨兵错误
═══════════════════════════════════════════════════════════════════════

【核心概念】
    只要一个类型有 `Error() string` 方法，它就是一个 error。
    Go 的 error 接口只有这一个方法 —— 所以「自定义错误」门槛极低：

        type QueryError struct { Op string; Code int; Err error }
        func (e *QueryError) Error() string { ... }

    两种最常用的形态：

    1. 哨兵错误（sentinel error）：`var ErrNotFound = errors.New("...")`
       适合「只有一种原因」的场景，调用方用 errors.Is 判断。
       缺点是没法带上下文（谁、什么参数），只能靠外层 %w 补。

    2. 自定义错误类型：带字段（错误码、操作名、原始错误），
       调用方用 errors.As 取出来读字段。适合需要结构化信息的场景。

    自定义类型还可以实现 Unwrap() error 把链接下去 ——
    这样「errors.Is 判断哨兵 + errors.As 取错误码」能同时成立。

    接收者选择：想用 errors.As 取出时，用**指针接收者**是更常见的做法
    （错误对象通常不可变、体积略大），但要注意：只有 *QueryError 才实现 error，
    QueryError 本身不算 error。

【你会看到】
    ✓ *QueryError 的 Error() 输出含操作名和错误码
    ✓ errors.As 把 *QueryError 从链里捞出来，读到 Code 字段
    ✓ errors.Is(err, ErrNotFound) 同时为 true —— 因为实现了 Unwrap
    ✓ 错误码模式：按 Code 分支给出不同的用户提示

【思考题】（动手改代码验证，别空想）
    1. Java 里自定义异常要继承 Exception，还得注意 checked/unchecked；
       Go 只要一个 Error() string 方法。这种「鸭子类型」的代价是什么？
       （提示：编译器不会强制你处理任何错误）
    2. 把 Error() 的接收者从 *QueryError 改成 QueryError，再跑一遍。
       errors.As 还能成功吗？为什么？（提示：看方法集，见 L1-09）
    3. 哨兵错误 vs 自定义类型：如果调用方只需要知道「是/不是」，用哪个？
       如果需要知道「第几次尝试失败、错误码是多少」，用哪个？
    4. 为什么哨兵错误要用包级变量而不是每次 errors.New 新建一个？
       （提示：errors.Is 底层是 == 比较）

【通关标准】
    能写一个带字段和 Unwrap 的自定义错误类型，
    并同时用 errors.As 取类型、errors.Is 判哨兵。

【前置】L1-19              【标签】custom error · sentinel · errors.As · 错误码
═══════════════════════════════════════════════════════════════════════
*/

// ── 哨兵错误：包级变量，用来给别人做 errors.Is 判断 ──────────────────────
// 注意是 var 而不是每次调用时新建 —— errors.Is 底层用的是 ==，
// 每次新建的 errors.New 是不同的指针，永远比不相等。
var (
	ErrNotFound = errors.New("记录不存在")
	ErrConflict = errors.New("记录已存在")
)

// ── 错误码：用常量把「机器可判断的语义」和「给人看的文字」分开 ────────────
// 这样调用方可以 switch Code，而不必去匹配字符串（字符串一改就崩）。
type Code int

const (
	CodeUnknown  Code = 0
	CodeNotFound Code = 404
	CodeConflict Code = 409
)

// QueryError 是自定义错误类型：把「哪一步操作」「什么错误码」「底层原因」都带上。
type QueryError struct {
	Op   string // 出错的操，比如 "GetUser"
	Code Code   // 结构化的错误码，供调用方 switch
	Err  error  // 底层原因（可能是哨兵错误，也可能还是包装过的）
}

// Error 实现 error 接口。注意接收者是**指针**：
// 所以只有 *QueryError 是 error，QueryError 值本身不是。
func (e *QueryError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("%s 失败（code=%d）", e.Op, e.Code)
	}
	return fmt.Sprintf("%s 失败（code=%d）: %v", e.Op, e.Code, e.Err)
}

// Unwrap 把链接下去。有了它，errors.Is(err, ErrNotFound) 才能穿透
// QueryError 找到底层的哨兵错误 —— 这就是「既能取字段、又能判哨兵」的关键。
func (e *QueryError) Unwrap() error { return e.Err }

// GetUser 演示「底层返回哨兵错误，上层包装成自定义类型」的典型分层。
// 返回类型是 error，但实际动态类型是 *QueryError。
func GetUser(id int) error {
	switch {
	case id == 0:
		// %w 挂链，让 ErrNotFound 留在链上
		return &QueryError{Op: "GetUser", Code: CodeNotFound, Err: fmt.Errorf("id=%d: %w", id, ErrNotFound)}
	case id == 1:
		return nil
	default:
		// 没有底层原因时 Err 留 nil，Unwrap 返回 nil，链到此为止
		return &QueryError{Op: "GetUser", Code: CodeUnknown}
	}
}

// CreateUser 演示另一种常见做法：先建再查，冲突时返回哨兵错误。
func CreateUser(name string, exists bool) error {
	if exists {
		return fmt.Errorf("CreateUser(%q): %w", name, ErrConflict)
	}
	return nil
}

// userMessage 演示「错误 → 用户提示」的映射，两条路都走一遍：
// 优先用 errors.As 拿结构化错误码（信息最全），
// 拿不到就退化成 errors.Is 判哨兵错误（信息少但够用）。
// 共同点是：判断只依赖 Code / 哨兵，文案随时能改，逻辑不受影响。
func userMessage(err error) string {
	var qerr *QueryError
	if errors.As(err, &qerr) { // 路 1：结构化错误，读 Code
		switch qerr.Code {
		case CodeNotFound:
			return "你查询的内容不存在，换个关键词试试"
		case CodeConflict:
			return "该名称已被占用，请换一个"
		default:
			return "服务暂时不可用，请稍后重试"
		}
	}
	// 路 2：只有哨兵错误，没有错误码，也能给出对的提示
	switch {
	case errors.Is(err, ErrNotFound):
		return "你查询的内容不存在，换个关键词试试"
	case errors.Is(err, ErrConflict):
		return "该名称已被占用，请换一个"
	}
	return "发生未知错误"
}

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L1-20 自定义错误与哨兵错误")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// ── 示例 1：自定义错误类型的 Error() 输出 ────────────────────────────
	fmt.Println("📝 示例 1：自定义错误类型：Error() 带上操作名和错误码")
	fmt.Println(strings.Repeat("-", 60))
	err := GetUser(0)
	fmt.Printf("  err = %v\n", err)
	fmt.Printf("  动态类型：%T  ← 注意是指针类型 *main.QueryError\n", err)

	// ── 示例 2：errors.As 取出具体类型，读字段 ───────────────────────────
	// 这是自定义错误类型最大的价值：拿到**结构化的字段**，而不是一段文字。
	fmt.Println("📌 示例 2：errors.As：取出 *QueryError 读字段")
	fmt.Println(strings.Repeat("-", 60))
	var qerr *QueryError
	if errors.As(err, &qerr) {
		fmt.Printf("  命中！Op=%q Code=%d\n", qerr.Op, qerr.Code)
		fmt.Printf("  还能顺着 Unwrap 拿到下一层：%v\n", errors.Unwrap(qerr))
	}

	// ── 示例 3：同时用 errors.Is 判哨兵 ─────────────────────────────────
	// 因为 QueryError 实现了 Unwrap，errors.Is 能一路剥到 ErrNotFound。
	// 于是「判断原因」和「读结构化字段」这两件事可以同时做到。
	fmt.Println("🔍 示例 3：errors.Is 穿透自定义错误，找到哨兵")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("  errors.Is(err, ErrNotFound) = %v\n", errors.Is(err, ErrNotFound))
	fmt.Printf("  errors.Is(err, ErrConflict) = %v  ← 不是这个原因\n", errors.Is(err, ErrConflict))

	// ── 示例 4：错误码模式 → 用户提示 ───────────────────────────────────
	// 三种错误形态：带错误码的 *QueryError、只有哨兵错误的包装错误、
	// 以及完全没有结构化信息的错误。userMessage 会依次降级处理。
	fmt.Println("🧪 示例 4：错误码模式：按 Code / 哨兵错误给出用户可读的提示")
	fmt.Println(strings.Repeat("-", 60))
	for _, e := range []error{GetUser(0), GetUser(99), CreateUser("alice", true)} {
		var qe *QueryError
		tag := "只有哨兵错误，没有错误码"
		if errors.As(e, &qe) {
			tag = fmt.Sprintf("有错误码 code=%d", qe.Code)
		}
		fmt.Printf("  %v\n      %s → %s\n", e, tag, userMessage(e))
	}
	fmt.Println("  ➡️  判断走 Code / errors.Is，文案只给人看 —— 两者分离，改文案不会改逻辑")

	// ── 示例 5：哨兵错误必须复用同一个变量 ──────────────────────────────
	fmt.Println("⚙️ 示例 5：哨兵错误为什么必须是包级变量")
	fmt.Println(strings.Repeat("-", 60))
	fresh := errors.New("记录不存在")
	fmt.Printf("  errors.Is(ErrNotFound, fresh) = %v  ← 文字一样但不是一个对象\n", errors.Is(ErrNotFound, fresh))
	fmt.Printf("  errors.Is(ErrNotFound, ErrNotFound) = %v  ← 同一个变量才相等\n", errors.Is(ErrNotFound, ErrNotFound))

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ error 接口只有一个方法，实现 Error() string 就能自定义错误")
	fmt.Println("  ✅ 哨兵错误（包级 var）适合「只有一种原因」，用 errors.Is 判断，必须复用同一变量")
	fmt.Println("  ✅ 自定义类型适合「需要结构化字段」，用 errors.As 取出（接收者用指针）")
	fmt.Println("  ✅ 实现 Unwrap() error 把链接下去，Is 和 As 就能同时生效")
	fmt.Println("  ✅ 错误码让逻辑判断与文案解耦：Code 给机器，Error() 给人")
}
