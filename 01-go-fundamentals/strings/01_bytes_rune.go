//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 01-go-fundamentals && go run strings/01_bytes_rune.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

/*
═══════════════════════════════════════════════════════════════════════
L1-22：string / []byte / rune 与 UTF-8
═══════════════════════════════════════════════════════════════════════

【核心概念】
    Go 的 string 是**只读的字节序列**，不是「字符数组」。它内部就是
    (指针, 长度)，存的是 UTF-8 编码后的原始字节。由此推出三个后果：

    1. len(s) 返回**字节数**，不是字符数。len("中文") == 6，不是 2。
       想知道字符数要用 utf8.RuneCountInString(s) 或 len([]rune(s))。
    2. s[i] 取出来的是**第 i 个字节**（byte = uint8），不是第 i 个字符。
       对多字节字符，单看某个字节毫无意义。
    3. for range 一个 string 时，Go 会**按 UTF-8 解码**，每次给出
       (字节下标, rune)。这是遍历「字符」的正确方式。
       rune = int32，表示一个 Unicode 码点（一个「字符」）。

    两种转换都有**拷贝代价**：
        []byte(s)  ← 拷一份字节（可写）
        []rune(s)  ← 解码成 int32 数组（一般比 []byte 大 4 倍）
        string(b)  ← 反向拷贝，且拷贝之后不可变

    经典 bug：按字节切片截断中文字符串。s[:3] 只切了 1.5 个汉字，
    剩下的半个字符就是乱码（U+FFFD 或直接是无效字节）。
    正确做法是先转 []rune 再切，或用 strings 包 / utf8 包的函数。

【你会看到】
    ✓ len("Go语言") == 8（字节）而 utf8.RuneCountInString 是 4（字符）
    ✓ s[2] 是一个字节值，而 for range 给出的第 3 个值是汉字 '语'
    ✓ s[:3] 的 %q 打印出 "Go\xe8" —— 半个汉字变成无效字节
    ✓ 转 []rune 再切片，"Go语言"[:2] 变成 "Go"（字符维度），不再乱码

【思考题】（动手改代码验证，别空想）
    1. Java 的 String 是 UTF-16 码元数组，s.length() 返回的是码元数不是字符数；
       Go 的 len(s) 返回字节数。两种「长度」各在什么场景下会骗人？
       （试试 emoji "👨‍👩‍👧" 和带音标的 "é"）
    2. Python 3 的 str 是码点序列，len("中文") == 2。Go 为什么要暴露字节这一层？
       （提示：网络、文件、加密算法的接口都是字节）
    3. 在 100 万次循环里频繁 []rune(s) 和 []byte(s) 会怎样？写个循环感受一下开销，
       想想什么情况下该一直持有 []byte 而不是来回转。
    4. 如何安全地把字符串截断到最多 N 个字符？写出一个函数，
       并用它处理 "Go语言很棒" 截断到 4、5、6 个字符的情况。

【通关标准】
    能解释「len 是字节、range 是字符、s[i] 是字节」这三句话，
    并知道中文截断 bug 的成因和修法。

【前置】—                  【标签】UTF-8 · rune · byte · 字符串不可变
═══════════════════════════════════════════════════════════════════════
*/

// truncateRunes 按**字符**截断，最多保留 n 个 rune。
// 对比 s[:n] 的按字节截断：这里不会产生半个汉字。
func truncateRunes(s string, n int) string {
	r := []rune(s) // 先解码成码点数组
	if n >= len(r) {
		return s
	}
	return string(r[:n])
}

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L1-22 string / []byte / rune 与 UTF-8")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	s := "Go语言" // G, o 各 1 字节；语, 言 各 3 字节（UTF-8 中文通常是 3 字节）

	// ── 示例 1：len 返回字节数 ───────────────────────────────────────────
	fmt.Println("📝 示例 1：len(s) 是字节数，不是字符数")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("  s = %q\n", s)
	fmt.Printf("  len(s) = %d  ← 字节数：2 + 3 + 3\n", len(s))
	fmt.Printf("  utf8.RuneCountInString(s) = %d  ← 字符数（码点数）\n", utf8.RuneCountInString(s))
	fmt.Printf("  len([]rune(s)) = %d  ← 等价写法，但会拷贝一份\n", len([]rune(s)))
	fmt.Println("  ⚠️  和 Python 3 的 len(\"中文\")==2 不同，Go 的 len 是字节")

	// ── 示例 2：s[i] 是字节，for range 是字符 ────────────────────────────
	fmt.Println("📌 示例 2：s[i] 取字节，for range 取字符")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("  s[0] = %d (%c)  ← 下标访问得到 byte\n", s[0], s[0])
	fmt.Printf("  s[2] = %d  ← 这是「语」的第一个字节，单独看没有意义\n", s[2])
	fmt.Println("  用 for range 遍历（正确姿势）：")
	for i, r := range s {
		// i 是字节下标，r 是 rune（码点）；一个汉字会让 i 跳 3
		fmt.Printf("    字节下标 i=%-2d rune=%q(%c) 码点=%U 占 %d 字节\n",
			i, r, r, r, utf8.RuneLen(r))
	}
	fmt.Println("  ➡️  i 不是 0,1,2,3 而是 0,1,2,5 —— 因为 i 是字节位置，不是字符序号")

	// ── 示例 3：按字节遍历会看到「半个汉字」 ─────────────────────────────
	fmt.Println("🔍 示例 3：按字节遍历：把 UTF-8 拆碎了")
	fmt.Println(strings.Repeat("-", 60))
	for i := 0; i < len(s); i++ {
		fmt.Printf("    s[%d] = 0x%02x  %q\n", i, s[i], s[i])
	}
	fmt.Println("  ➡️  0xe8 0xaf 0xad 是「语」的 3 个字节，拆开各自都不是合法字符")

	// ── 示例 4：经典 bug —— 按字节截断中文 ──────────────────────────────
	fmt.Println("🧪 示例 4：中文截断 bug：s[:3] 得到半个汉字")
	fmt.Println(strings.Repeat("-", 60))
	broken := s[:3]
	fmt.Printf("  s[:3] = %q  ← \\xe8 就是那个残缺字节\n", broken)
	fmt.Printf("  utf8.ValidString(s[:3]) = %v  ← 不是合法 UTF-8\n", utf8.ValidString(s[:3]))
	fmt.Printf("  打印出来看到的是：%s（后面会跟一个替换字符）\n", broken)

	fixed := truncateRunes(s, 2) // 要 2 个字符，而不是 2 个字节
	fmt.Printf("  truncateRunes(s, 2) = %q  ← 按字符切，干净\n", fixed)
	fmt.Printf("  utf8.ValidString(truncateRunes(s, 2)) = %v\n", utf8.ValidString(fixed))
	fmt.Println("  ➡️  修法：先 []rune(s) 再切片；或者用 strings 包里按 rune 操作的函数")

	// ── 示例 5：string / []byte / []rune 三种转换 ───────────────────────
	fmt.Println("⚙️ 示例 5：三种转换：每一次都发生拷贝")
	fmt.Println(strings.Repeat("-", 60))
	b := []byte(s)  // 拷一份字节，可写
	r := []rune(s)  // 解码成 int32 数组
	s2 := string(r) // 再编码回字符串
	fmt.Printf("  []byte(s) = %v  （len=%d）\n", b, len(b))
	fmt.Printf("  []rune(s) = %v  （len=%d，每元素是 int32）\n", r, len(r))
	b[0] = 'g' // 改字节切片，原字符串不受影响 —— 说明是拷贝
	fmt.Printf("  改 b[0]='g' 后：b = %q，而 s 仍然是 %q\n", b, s)
	fmt.Printf("  string([]rune(s)) = %q  ← 绕一圈又回来了，但白拷了两次\n", s2)
	fmt.Println("  ⚠️  字符串不可变（不能 s[0]='g'），所以每次「改字符串」都是新建一个")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ string = 只读的 UTF-8 字节序列，len(s) 是字节数不是字符数")
	fmt.Println("  ✅ s[i] 取字节；for range 按 UTF-8 解码，给出 (字节下标, rune)")
	fmt.Println("  ✅ rune = int32 = 一个 Unicode 码点；byte = uint8 = 一个字节")
	fmt.Println("  ✅ 按字节切多字节字符串会产生乱码，要切就 []rune(s) 再切")
	fmt.Println("  ✅ []byte(s) / []rune(s) / string(b) 都拷贝，别在热循环里来回转")
	fmt.Println("  ✅ utf8.RuneCountInString 数字符，utf8.ValidString 验合法性")
}
