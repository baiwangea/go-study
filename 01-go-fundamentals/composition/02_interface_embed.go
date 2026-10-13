//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 01-go-fundamentals && go run composition/02_interface_embed.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

/*
═══════════════════════════════════════════════════════════════════════
L1-27：接口嵌入与小接口组合
═══════════════════════════════════════════════════════════════════════

【核心概念】
    接口可以嵌入接口 —— 这就是「用小的接口拼出大的接口」：

        type Reader interface { Read() string }
        type Writer interface { Write(s string) }
        type ReadWriter interface {
            Reader   // 嵌入：ReadWriter 的方法集 = Reader ∪ Writer
            Writer
        }

    标准库最著名的例子就是 io 包：

        type Reader interface { Read(p []byte) (int, error) }
        type Writer interface { Write(p []byte) (int, error) }
        type ReadWriter interface { Reader; Writer }

    **为什么 io.Reader 只有一个方法？**
    方法越少，能实现它的类型就越多 —— 文件、网络连接、字符串、bytes.Buffer、
    压缩流、加密流……全都满足 io.Reader。于是 io.Copy(dst, src) 一个函数
    就能处理所有组合。这就是 Go 社区那句「接口越小越强」的含义。
    反面教材：一个 10 方法的接口，几乎没人能实现，只能自己写 mock。

    any 与 interface{}：
        `any` 是 Go 1.18 加的**类型别名**，完全等价于 `interface{}`，
        两者可以互相赋值、互相比较，`%T` 输出的动态类型也完全一样。
        新代码统一写 any 即可（更短、更表意）。

    字面量接口：`interface{ M() }` 可以直接当类型用，不用先起名字。
    适合「只在这一个函数签名里用一次」的场景。

【你会看到】
    ✓ ReadWriter 由 Reader + Writer 拼成，*buffer 一次满足三个接口
    ✓ 自己写一个只有 Write 方法的类型，就能接进 fmt.Fprintf / io.Copy 生态
    ✓ io.Copy 把 strings.Reader 的内容搬到 strings.Builder，全程只用 io 的两个小接口
    ✓ any 和 interface{} 可以互相赋值，reflect 看到的类型一致
    ✓ interface{ Read() string } 这种字面量接口可以直接用

【思考题】（动手改代码验证，别空想）
    1. Java 的接口倾向于「大而全」（List 有几十个方法），Java 8 之后靠 default
       方法缓解。Go 的小接口为什么能成立？这和「隐式实现」有什么关系？（见 L1-09）
    2. 接口嵌入时，如果 Reader 和 Writer 里都有一个叫 Close() 但签名不同的方法，
       编译会怎样？签名相同呢？动手试。
    3. 写一个只接收 io.Writer 的函数，然后依次传 os.Stdout、&strings.Builder{}、
       自己写的类型 —— 体会「小接口 + 隐式实现」带来的通用性。
    4. any 是别名而不是新类型，怎么用代码证明？（提示：赋值给 interface{} 变量，
       再看 %T；如果是新类型，赋值就需要显式转换）

【通关标准】
    能自己用 2~3 个小接口拼出一个大接口，并说清「小接口 + 隐式实现」的威力。

【前置】L1-08 · L1-10       【标签】interface embedding · io.Reader · any · 小接口
═══════════════════════════════════════════════════════════════════════
*/

// ── 两个最小接口：各只有一个方法 ────────────────────────────────────────
// 故意不用 io 包的同名接口，先用自定义的看清楚「拼装」这件事。
type Reader interface{ Read() string }
type Writer interface{ Write(s string) }

// ReadWriter 不定义任何新方法，只是把两个小接口拼起来。
// 方法集 = Reader ∪ Writer，所以任何同时有 Read 和 Write 的类型都满足它。
type ReadWriter interface {
	Reader
	Writer
}

// buffer 是内存里的一行行日志缓冲，同时实现 Read 和 Write。
type buffer struct {
	lines []string
}

func (b *buffer) Write(s string) { b.lines = append(b.lines, s) }

func (b *buffer) Read() string {
	if len(b.lines) == 0 {
		return ""
	}
	s := b.lines[0]
	b.lines = b.lines[1:]
	return s
}

// 编译期断言：一个类型同时满足三个接口，不用写任何额外代码。
var (
	_ Reader     = (*buffer)(nil)
	_ Writer     = (*buffer)(nil)
	_ ReadWriter = (*buffer)(nil)
)

// drain 只要求 Reader —— 参数越小，能接的类型越多。
func drain(r Reader) []string {
	var out []string
	for {
		line := r.Read()
		if line == "" {
			break
		}
		out = append(out, line)
	}
	return out
}

// ── 自定义 io.Writer：只写一个方法就能接入整个 io / fmt 生态 ─────────────
// prefixWriter 把每次写入都加上前缀，可以想象成「日志加时间戳」的简化版。
type prefixWriter struct {
	prefix string
	sb     strings.Builder
}

// Write 是 io.Writer 唯一要求的方法。返回值是 (写了多少字节, 错误)。
func (w *prefixWriter) Write(p []byte) (int, error) {
	w.sb.WriteString(w.prefix)
	return w.sb.Write(p) // strings.Builder.Write 恰好就是 io.Writer 的形状
}

// String 取出累积的内容（顺便让它也满足 fmt.Stringer）。
func (w *prefixWriter) String() string { return w.sb.String() }

// 编译期断言：只实现一个方法，就自动是 io.Writer 了。
var _ io.Writer = (*prefixWriter)(nil)

// report 接收 io.Writer —— 这是「小接口」的通用性演示：
// os.Stdout、*bytes.Buffer、*strings.Builder、*prefixWriter 全都能传进来。
func report(w io.Writer, name string, score int) {
	fmt.Fprintf(w, "%-6s score=%d\n", name, score)
}

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L1-27 接口嵌入与小接口组合")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// ── 示例 1：接口嵌入 = 小接口拼成大接口 ─────────────────────────────
	fmt.Println("📝 示例 1：ReadWriter = Reader + Writer")
	fmt.Println(strings.Repeat("-", 60))
	var rw ReadWriter = &buffer{}
	rw.Write("line-1")
	rw.Write("line-2")
	rw.Write("line-3")
	fmt.Printf("  rw.Read() = %q\n", rw.Read())
	fmt.Printf("  剩下的行（用只收 Reader 的 drain）= %q\n", drain(rw))
	fmt.Println("  ➡️  *buffer 只写了 Read 和 Write 两个方法，就同时满足了三个接口")

	// ── 示例 2：一个方法就能接进 io 生态 ────────────────────────────────
	fmt.Println("📌 示例 2：只实现 Write 一个方法，就能当 io.Writer 用")
	fmt.Println(strings.Repeat("-", 60))
	pw := &prefixWriter{prefix: "[log] "}
	fmt.Fprintf(pw, "连接建立，耗时 %dms\n", 42) // fmt 只要求 io.Writer
	fmt.Fprintf(pw, "查询完成，命中 %d 条\n", 3)
	fmt.Print("  输出：")
	fmt.Print(pw.String())
	fmt.Println("  ➡️  fmt.Fprintf 的第一个参数就是 io.Writer，所以任何实现了 Write 的类型都能接")

	// ── 示例 3：同一个 report 输出到不同目的地 ──────────────────────────
	fmt.Println("🔍 示例 3：io.Writer 的通用性：一处代码，多种目的地")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("  → 输出到 os.Stdout：")
	report(os.Stdout, "alice", 90)
	fmt.Println("  → 输出到 bytes.Buffer：")
	var bb bytes.Buffer
	report(&bb, "bob", 85)
	fmt.Printf("    buffer 里是 %q\n", bb.String())
	fmt.Println("  ➡️  report 只依赖 io.Writer 这一个方法，就能写到终端、内存、文件、网络")

	// ── 示例 4：io.Copy 只用两个最小接口搬运数据 ────────────────────────
	fmt.Println("🧪 示例 4：io.Copy(dst, src)：全靠两个单方法接口")
	fmt.Println(strings.Repeat("-", 60))
	var out strings.Builder
	n, err := io.Copy(&out, strings.NewReader("从 Reader 流到 Writer"))
	fmt.Printf("  搬运了 %d 字节，err=%v，内容=%q\n", n, err, out.String())
	// 编译期断言：标准库这两个类型本来就是 io 生态里的实现
	var (
		_ io.Reader = strings.NewReader("")
		_ io.Writer = &out
		_ io.Reader = &bytes.Buffer{}
	)
	fmt.Println("  ➡️  io.Copy 不关心具体是什么类型，只要满足 io.Reader / io.Writer 就能用")
	fmt.Println("    这就是「小接口 + 隐式实现」带来的组合爆炸式复用")

	// ── 示例 5：any 与 interface{} 完全等价 ─────────────────────────────
	fmt.Println("⚙️ 示例 5：any 就是 interface{} 的别名")
	fmt.Println(strings.Repeat("-", 60))
	var a any = 42
	var b interface{} = a // 直接赋值，不需要转换 —— 说明是别名不是新类型
	fmt.Printf("  any 变量 %v，%%T = %T\n", a, a)
	fmt.Printf("  interface{} 变量 %v，%%T = %T\n", b, b)
	fmt.Printf("  两者类型相同？%v\n", fmt.Sprintf("%T", a) == fmt.Sprintf("%T", b))
	fmt.Println("  ➡️  any 是 Go 1.18 引入的 type alias，编译后完全一样；新代码统一写 any")

	// ── 示例 6：字面量接口 ──────────────────────────────────────────────
	fmt.Println("🎯 示例 6：字面量接口：不用起名字也能当类型")
	fmt.Println(strings.Repeat("-", 60))
	// 只在这个位置用一次，就没必要单独定义具名接口。
	var inline interface {
		Read() string
		Write(s string)
	} = &buffer{}
	inline.Write("临时接口也能用")
	fmt.Printf("  字面量接口变量 %T 的 Read() = %q\n", inline, inline.Read())
	fmt.Println("  ➡️  interface{ M() } 是合法类型；方法多的场景还是具名接口更好读")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ 接口嵌入接口 = 方法集求并集，用小接口拼大接口（io.ReadWriter）")
	fmt.Println("  ✅ 方法越少，能实现的类型越多 —— io.Reader / io.Writer 都只有一个方法")
	fmt.Println("  ✅ 小接口 + 隐式实现 → 一个 io.Copy 通吃文件、网络、内存、压缩流")
	fmt.Println("  ✅ any 是 interface{} 的别名，可互相赋值，新代码统一写 any")
	fmt.Println("  ✅ interface{ M() } 字面量接口可直接当类型，适合只用一次的场景")
	fmt.Println("  ✅ 别设计「大接口」：10 个方法的接口几乎没人能实现，也没法 mock")
}
