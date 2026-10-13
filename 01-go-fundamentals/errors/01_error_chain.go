//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 01-go-fundamentals && go run errors/01_error_chain.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

/*
═══════════════════════════════════════════════════════════════════════
L1-19：错误链 —— errors.Is / As / Unwrap
═══════════════════════════════════════════════════════════════════════

【核心概念】
    每一层函数都可能想给错误补充上下文（"读取配置失败" → "解析 YAML 失败" →
    "第 12 行缩进错误"）。但一旦你把错误**转成字符串**再往上抛，底层那个
    "原始错误"就消失了，调用方再也没法判断「这到底是不是超时」。

    Go 1.13 起，fmt.Errorf 支持 %w 动词：它不是拼接字符串，而是**把原错误
    挂在新错误上**，形成一条链：

        err3 (loadConfig) → err2 (parseYAML) → err1 (ErrSyntax)

    配套的三个工具（都在 errors 包）：
        errors.Unwrap(err)   手动剥一层，拿到链上的下一个
        errors.Is(err, target)   顺着链一路剥，问「链上有没有 == target 的」
        errors.As(err, &target)  顺着链一路剥，问「链上有没有能塞进 target 的类型」

    关键区别：
        %w → 保留链，errors.Is/As 能穿透
        %v → 只留文字，链断了，errors.Is/As 只能看到最外层

【你会看到】
    ✓ err.Error() 打印出 "loadConfig: parseYAML: 第 12 行缩进错误" 整条链
    ✓ errors.Is(err, ErrSyntax) == true —— 穿了 3 层找到最底层的哨兵错误
    ✓ 用 %v 包装的版本 errors.Is 变成 false —— 链断了
    ✓ errors.As 把链上的 *os.PathError 捞出来，拿到它自己的字段

【思考题】（动手改代码验证，别空想）
    1. Java 里 `new RuntimeException("包装", cause)` + `getCause()` 逐层找，
       和 Go 的 %w + errors.Is 相比，哪一步是 Go 省掉的？为什么 Go 不需要
       自定义异常类型也能保留链路？
    2. 一个错误最多能被 %w 包装几次？试试同一个 fmt.Errorf 里写两个 %w，
       然后 errors.Is 分别判断两个底层错误，看结果。
    3. errors.Is 判断的是「相等」还是「是不是同一条链」？读一下它的文档，
       想想为什么自定义错误类型可以实现 Is(error) bool 方法来定制判断。
    4. os.ErrNotExist 配合 errors.Is 判断「文件不存在」，比 os.IsNotExist 好在哪？
       （提示：用 %w 包装过一层之后，老 API 还认得出来吗？）

【通关标准】
    能说清 %w 和 %v 的区别，并能用 errors.Is 穿透多层包装判断底层错误。

【前置】L1-10              【标签】error wrapping · errors.Is · errors.As · Unwrap
═══════════════════════════════════════════════════════════════════════
*/

// ErrSyntax 是一个「哨兵错误」（sentinel error）：一个包级变量，
// 专门用来给别人做 errors.Is 判断的锚点。命名惯例是 Err 前缀。
var ErrSyntax = errors.New("语法错误")

// parseYAML 是最内层：它产生并返回哨兵错误本身（第 1 层）。
func parseYAML(line int) error {
	return fmt.Errorf("第 %d 行缩进错误: %w", line, ErrSyntax)
}

// loadConfig 是中间层：用 %w 把下层错误挂到自己身上（第 2 层）。
func loadConfig(line int) error {
	if err := parseYAML(line); err != nil {
		return fmt.Errorf("parseYAML: %w", err)
	}
	return nil
}

// loadConfigLoose 是对照组：同样的包装，但用 %v 而不是 %w。
// 字符串看起来一模一样，但链断在这里，errors.Is 再也穿不下去。
func loadConfigLoose(line int) error {
	if err := parseYAML(line); err != nil {
		return fmt.Errorf("parseYAML: %v", err) // %v 只保留文字，丢弃原错误
	}
	return nil
}

// readFile 演示 errors.As：它需要一个**指向目标类型的指针**，
// 这样 errors.As 才能把找到的值写回去。
func readFile(path string) error {
	_, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("readFile(%s): %w", path, err) // 包装 os 包的错误
	}
	return nil
}

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L1-19 错误链")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// ── 示例 1：%w 保留链路，Error() 打印完整上下文 ──────────────────────
	fmt.Println("📝 示例 1：%w 包装：Error() 显示整条链")
	fmt.Println(strings.Repeat("-", 60))
	err := loadConfig(12)
	fmt.Printf("  最外层错误：%v\n", err)
	fmt.Println("  ➡️  三层信息拼在一起，读日志的人一眼知道发生了什么")

	// ── 示例 2：errors.Is 穿透多层找到哨兵错误 ──────────────────────────
	// 注意 err 的静态类型只是 error，我们完全不知道链里有什么，
	// 但 errors.Is 会一层层 Unwrap 下去比对。
	fmt.Println("📌 示例 2：errors.Is：穿透多层包装")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("  errors.Is(err, ErrSyntax) = %v  ← 穿透了 2 层包装\n", errors.Is(err, ErrSyntax))
	fmt.Printf("  err == ErrSyntax          = %v  ← 直接比较只能看最外层，必然失败\n", err == ErrSyntax)

	// ── 示例 3：%v 包装会断链 ────────────────────────────────────────────
	fmt.Println("🔍 示例 3：%v 包装：字符串一样，链却断了")
	fmt.Println(strings.Repeat("-", 60))
	loose := loadConfigLoose(12)
	fmt.Printf("  %v\n", loose)
	fmt.Printf("  errors.Is(loose, ErrSyntax) = %v  ← 同样的文字，判断失败！\n", errors.Is(loose, ErrSyntax))
	fmt.Println("  ➡️  结论：想保留可判断性必须用 %w，只有纯粹想拼消息时才用 %v")

	// ── 示例 4：errors.Unwrap 手动剥壳 ──────────────────────────────────
	fmt.Println("🧪 示例 4：errors.Unwrap：手动一层层剥")
	fmt.Println(strings.Repeat("-", 60))
	for cur, depth := err, 0; cur != nil; cur, depth = errors.Unwrap(cur), depth+1 {
		fmt.Printf("  第 %d 层：%v\n", depth, cur)
	}
	fmt.Println("  ➡️  剥到 errors.Unwrap 返回 nil，说明到底了（哨兵错误没有下一层）")

	// ── 示例 5：errors.As 取回具体类型 ──────────────────────────────────
	fmt.Println("⚙️ 示例 5：errors.As：从链里捞出具体类型")
	fmt.Println(strings.Repeat("-", 60))
	rerr := readFile("/no/such/file/here.txt")
	fmt.Printf("  原始错误：%v\n", rerr)

	var pathErr *os.PathError // 声明目标类型的零值变量
	if errors.As(rerr, &pathErr) {
		// 命中后 pathErr 已经被填好，可以直接用它的字段
		fmt.Printf("  命中 *os.PathError：Op=%q Path=%q\n", pathErr.Op, pathErr.Path)
	}
	fmt.Printf("  errors.Is(rerr, os.ErrNotExist) = %v  ← 判断「文件不存在」这类语义\n", errors.Is(rerr, os.ErrNotExist))
	fmt.Println("  ➡️  这就是为什么现代代码用 errors.Is(err, os.ErrNotExist) 而不是 os.IsNotExist(err)")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ %w 包装 = 挂链，%v 包装 = 拼字符串丢链，选错就失去可判断性")
	fmt.Println("  ✅ errors.Is 问「链上有没有这个值」（哨兵错误、os.ErrNotExist）")
	fmt.Println("  ✅ errors.As 问「链上有没有这个类型」，第二个参数必须是指针")
	fmt.Println("  ✅ errors.Unwrap 剥一层；err == target 只能比最外层，别用它做判断")
	fmt.Println("  ✅ 包装时补上下文（哪一层、什么参数），这是日志可读性的关键")
}
