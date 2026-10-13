//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 01-go-fundamentals && go run strings/02_builder_strconv.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

/*
═══════════════════════════════════════════════════════════════════════
L1-23：strings.Builder / strconv / strings 包
═══════════════════════════════════════════════════════════════════════

【核心概念】
    1. 循环里用 += 拼字符串是 O(n²)：Go 的 string 不可变，每次 += 都要
       **重新分配内存 + 拷贝全部已有内容**。拼 n 次就拷贝了大约 n²/2 个字节。
       Java 的 String 同样不可变（所以有 StringBuilder），Python 的 str
       也一样（所以推荐 "".join）。三家坑法一致，解法也一致。

    2. strings.Builder 是官方解药：内部是一个可增长的 []byte，
       写入是追加（均摊 O(1)），最后 String() 一次性转成 string。
       两个注意点：
         · 零值即可用，不需要 new/make
         · 不要拷贝它（内含 noCopy），要用指针 *strings.Builder

    3. strconv 负责「字符串 ↔ 数字/布尔」的转换：
        Atoi / Itoa（int）、ParseInt / FormatInt（任意进制）、
        ParseFloat / FormatFloat（浮点，可控制精度）、ParseBool。
       注意：Parse 系列**总是返回两个值**（值, error），
       Java 的 Integer.parseInt 抛 NumberFormatException，Go 让你显式处理。

    4. strings 包里的高频函数：
        Join / Split / Fields（按空白切，自动忽略连续空白）
        TrimSpace / TrimPrefix / TrimSuffix / Trim（去空白、去前缀后缀）
        Cut（一次拿到「分隔符前 / 后 / 是否找到」，比 Split 更省事）
        Contains / HasPrefix / HasSuffix / Index
        ReplaceAll / ToUpper / ToLower / Repeat / EqualFold

【你会看到】
    ✓ += 拼 20000 次 vs Builder 拼 20000 次的耗时对比（差一个数量级）
    ✓ Cut 一次拿到 key/value/ok，不用 Split 再判长度
    ✓ Fields 把连续空白当作一个分隔符（Split(" ") 做不到）
    ✓ strconv.Atoi 失败返回 error 而不是 panic
    ✓ FormatFloat 把 3.14159 格式化成 "3.14"

【思考题】（动手改代码验证，别空想）
    1. 把耗时对比的 n 从 20000 改成 200000，观察两者差距怎么变 ——
       为什么 += 的耗时涨得远不止 10 倍？（提示：O(n²)）
    2. Java 的 StringBuilder 有 append(int)，Go 的 Builder 只有 WriteString/
       WriteByte/WriteRune。那拼数字时该怎么办？（提示：strconv.Itoa 或 fmt.Fprintf）
    3. strings.Split("a,,b", ",") 和 strings.Fields("a  b") 结果差在哪？
       空字符串会被保留吗？
    4. strings.Cut 相比 strings.SplitN(s, sep, 2) 好在哪？为什么它不需要处理
       「切出来不够 2 段」这个边界？

【通关标准】
    能在循环里条件反射地想到 Builder（或 Join），并熟练用 strconv 做转换。

【前置】L1-22              【标签】strings.Builder · strconv · O(n²) · 拼接
═══════════════════════════════════════════════════════════════════════
*/

// buildWithPlus 用 += 拼接，演示 O(n²) 的代价。
// 每次 += 都会：申请更大的新内存 → 拷贝旧内容 → 追加新内容 → 丢弃旧内存。
func buildWithPlus(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s += "x" // 每次都全量拷贝，越到后面越慢
	}
	return s
}

// buildWithBuilder 用 strings.Builder，内部是 []byte 追加，均摊 O(1)。
// 注意参数是指针：Builder 内部有 noCopy 保护，按值传递会被 go vet 警告。
func buildWithBuilder(n int) string {
	var b strings.Builder
	b.Grow(n) // 预分配容量，避免中途扩容（可选，但推荐）
	for i := 0; i < n; i++ {
		b.WriteByte('x')
	}
	return b.String()
}

// csvLine 演示 Builder 拼「混合类型」的一行：WriteString 拼字符串，
// 数字先用 strconv.Itoa 转成字符串再写（Go 的 Builder 没有 append(int)）。
func csvLine(fields []string) string {
	var b strings.Builder
	for i, f := range fields {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(f)
	}
	return b.String()
}

// parseConfig 演示 strings.Cut：一次拿到 key / value / 是否找到。
// 对比 SplitN(s, "=", 2)：Cut 不需要判 len(parts) == 2，也不会因为
// value 里含 "=" 而切错（Cut 只在第一个分隔符处切）。
func parseConfig(line string) (key, value string, ok bool) {
	key, value, ok = strings.Cut(line, "=")
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	return key, value, ok
}

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L1-23 strings.Builder / strconv / strings")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// ── 示例 1：+= 在循环里是 O(n²) ─────────────────────────────────────
	fmt.Println("📝 示例 1：循环拼接：+= vs strings.Builder")
	fmt.Println(strings.Repeat("-", 60))
	const n = 20000
	t0 := time.Now()
	sPlus := buildWithPlus(n)
	dPlus := time.Since(t0)

	t1 := time.Now()
	sBuilder := buildWithBuilder(n)
	dBuilder := time.Since(t1)

	fmt.Printf("  += 拼 %d 次：耗时 %v，长度 %d\n", n, dPlus, len(sPlus))
	fmt.Printf("  Builder 拼 %d 次：耗时 %v，长度 %d\n", n, dBuilder, len(sBuilder))
	fmt.Printf("  ➡️  结果一样，但 += 每次都要重新分配 + 全量拷贝，是 O(n²)\n")

	// ── 示例 2：Builder 的常规用法 ──────────────────────────────────────
	fmt.Println("📌 示例 2：Builder：零值可用、用指针、最后 String()")
	fmt.Println(strings.Repeat("-", 60))
	var b strings.Builder
	b.WriteString("订单")
	b.WriteByte('-')
	b.WriteString(strconv.Itoa(1001)) // 数字要先转字符串
	b.WriteRune('✓')                  // 写一个 rune（可写中文/符号）
	fmt.Printf("  拼接结果 = %q\n", b.String())
	fmt.Printf("  Builder 当前长度 = %d\n", b.Len())
	b.Reset() // 复用：清空内容但保留已分配的容量
	fmt.Printf("  Reset 后 = %q，长度 %d（容量还在，适合放进循环复用）\n", b.String(), b.Len())
	fmt.Printf("  csvLine = %q\n", csvLine([]string{"BTC", "65000", "2026-10-12"}))

	// ── 示例 3：strings 高频函数 ────────────────────────────────────────
	fmt.Println("🔍 示例 3：strings 包高频函数")
	fmt.Println(strings.Repeat("-", 60))
	path := "/api/v1/users"
	fmt.Printf("  TrimPrefix(%q, \"/api\") = %q\n", path, strings.TrimPrefix(path, "/api"))
	fmt.Printf("  Split(%q, \"/\") = %q\n", path, strings.Split(path, "/"))
	fmt.Printf("  Join([\"a\",\"b\",\"c\"], \"-\") = %q\n", strings.Join([]string{"a", "b", "c"}, "-"))
	fmt.Printf("  Fields(\"  a   b  c \") = %q  ← 连续空白算一个分隔符\n", strings.Fields("  a   b  c "))
	fmt.Printf("  Split(\"  a   b  c \", \" \") = %q  ← 对比：空串被保留了\n", strings.Split("  a   b  c ", " "))
	fmt.Printf("  Contains(\"hello.go\", \".go\") = %v\n", strings.Contains("hello.go", ".go"))
	fmt.Printf("  ReplaceAll(\"a-b-c\", \"-\", \"_\") = %q\n", strings.ReplaceAll("a-b-c", "-", "_"))
	fmt.Printf("  Repeat(\"ab\", 3) = %q\n", strings.Repeat("ab", 3))
	fmt.Printf("  EqualFold(\"GO\", \"go\") = %v  ← 忽略大小写比较\n", strings.EqualFold("GO", "go"))
	fmt.Printf("  ToUpper(\"go语言\") = %q  ← 只影响 ASCII 字母，中文原样\n", strings.ToUpper("go语言"))

	// ── 示例 4：Cut —— 一次拿到三段 ─────────────────────────────────────
	fmt.Println("🧪 示例 4：strings.Cut：切一次拿到「前 / 后 / 是否找到」")
	fmt.Println(strings.Repeat("-", 60))
	for _, line := range []string{"host = 127.0.0.1", "port=8080", "这一行没有等号", "url=a=b=c"} {
		if key, value, ok := parseConfig(line); ok {
			fmt.Printf("  %q → key=%q value=%q\n", line, key, value)
		} else {
			fmt.Printf("  %q → 不是 key=value 格式（ok=false）\n", line)
		}
	}
	fmt.Println("  ➡️  Cut 只在第一个分隔符处切，所以 value 里含 '=' 也不会切错；")
	fmt.Println("    而且它把「没找到」做成第三个返回值，不用像 Split 那样判 len == 2")

	// ── 示例 5：strconv —— 字符串与数字互转 ─────────────────────────────
	fmt.Println("⚙️ 示例 5：strconv：Parse 系列返回 (值, error)")
	fmt.Println(strings.Repeat("-", 60))
	if v, err := strconv.Atoi("8080"); err == nil {
		fmt.Printf("  Atoi(\"8080\") = %d\n", v)
	}
	if _, err := strconv.Atoi("80x0"); err != nil {
		fmt.Printf("  Atoi(\"80x0\") → err = %v  ← 返回 error，不 panic\n", err)
	}
	fmt.Printf("  Itoa(404) = %q\n", strconv.Itoa(404))

	// ParseInt 支持进制：把十六进制字符串按 16 进制解析
	if v, err := strconv.ParseInt("ff", 16, 64); err == nil {
		fmt.Printf("  ParseInt(\"ff\", 16, 64) = %d\n", v)
	}
	fmt.Printf("  FormatInt(255, 16) = %q  ← 转成十六进制字符串\n", strconv.FormatInt(255, 16))

	// 浮点：ParseFloat 与 FormatFloat（第二个参数是格式，第三位是精度）
	if f, err := strconv.ParseFloat("3.14159", 64); err == nil {
		fmt.Printf("  ParseFloat(\"3.14159\") = %v\n", f)
		fmt.Printf("  FormatFloat(%.5f, 'f', 2, 64) = %q  ← 保留 2 位小数\n", f, strconv.FormatFloat(f, 'f', 2, 64))
		fmt.Printf("  FormatFloat(%.5f, 'e', 3, 64) = %q  ← 科学计数法\n", f, strconv.FormatFloat(f, 'e', 3, 64))
	}
	if bv, err := strconv.ParseBool("true"); err == nil {
		fmt.Printf("  ParseBool(\"true\") = %v\n", bv)
	}
	fmt.Printf("  Quote(\"他说: \\\"hi\\\"\") = %s  ← 加引号并转义，日志/调试很有用\n",
		strconv.Quote("他说: \"hi\""))

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ 循环里别用 +=，用 strings.Builder（零值可用、传指针、最后 String()）")
	fmt.Println("  ✅ 已知总长度就 b.Grow(n) 预分配；要复用就 b.Reset()")
	fmt.Println("  ✅ strings.Cut 一次拿到「前/后/是否找到」，比 SplitN 更省心")
	fmt.Println("  ✅ strings.Fields 按任意空白切并丢弃空段；Split 会保留空段")
	fmt.Println("  ✅ strconv.Parse* 返回 (值, error)，绝不 panic；Format* 反向转字符串")
	fmt.Println("  ✅ Builder 没有 append(int)，数字要先 strconv.Itoa 再 WriteString")
}
