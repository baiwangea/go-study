// Package level 提供「一关一关学 Go」的最小运行时。
//
// 设计契约（全仓库通用，新增模块请照抄本文件）：
//   - 一个 Level = 一个知识点 = 一个小文件（30~50 行，能独立跑）
//   - 关卡必须自带：前置(pre) / 目标(goal) / 观察(observe) / 思考题(questions) / 通关标准(check)
//   - main.go 只负责导航：list 看目录、传编号跑单关、不传参数按序通关
package level

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Level 描述一个最小学习单元。
type Level struct {
	ID        string   // 关卡编号，如 "L1-03"
	Title     string   // 关卡标题
	Tags      string   // 关键词标签，用 · 分隔
	Pre       string   // 前置关卡编号；第一关填 "无"
	Goal      string   // 本关要搞定的一件事
	Observe   string   // 运行后应该在输出里看到什么
	Questions []string // 思考题（建议动手改代码验证，而不是空想）
	Check     string   // 通关标准：能做到什么算过关
	Run       func()   // 演示代码本体
}

// Print 打印关卡横幅并执行演示。
func (l Level) Print() {
	line := strings.Repeat("─", 72)
	fmt.Printf("\n%s\n", line)
	fmt.Printf("▶ %s  %s\n", l.ID, l.Title)
	if l.Tags != "" {
		fmt.Printf("  标签  %s\n", l.Tags)
	}
	fmt.Printf("  前置  %s\n", l.Pre)
	fmt.Printf("  目标  %s\n", l.Goal)
	fmt.Printf("  观察  %s\n", l.Observe)
	if len(l.Questions) > 0 {
		fmt.Printf("  思考  %s\n", l.Questions[0])
		for _, q := range l.Questions[1:] {
			fmt.Printf("         %s\n", q)
		}
	}
	fmt.Printf("  通关  %s\n", l.Check)
	fmt.Printf("%s\n", line)
	fmt.Println("  ↓ 运行结果")

	l.Run()
}

// Catalog 打印关卡目录，便于先规划再逐关运行。
// 注意：中文字符宽度不统一，所以这里不用 %-Ns 对齐列，改成两行一条。
func Catalog(levels []Level) {
	fmt.Printf("共 %d 关，按编号逐关通关：\n\n", len(levels))
	for _, l := range levels {
		fmt.Printf("  %s  %s\n", l.ID, l.Title)
		fmt.Printf("       前置 %s ｜ 目标 %s\n", l.Pre, l.Goal)
	}
	fmt.Printf("\n用法：\n")
	fmt.Printf("  go run .            # 按序通关全部关卡\n")
	fmt.Printf("  go run . list       # 只看这张目录\n")
	fmt.Printf("  go run . 3          # 只跑第 3 关\n")
	fmt.Printf("  go run . 3-5        # 跑第 3 到第 5 关\n")
}

// Play 是各模块 main.go 的统一导航：无参数全跑，list 看目录，支持 3 / 3-5 / "3 5"。
// args 传入 os.Args[1:]，stageName 用于开场标题。
func Play(stageName string, levels []Level, args []string) {
	switch {
	case len(args) == 0:
		fmt.Printf("====== %s：%d 关按序通关 ======\n", stageName, len(levels))
		for _, l := range levels {
			l.Print()
		}
		fmt.Printf("\n✅ 全部关卡运行完毕。思考题没动手改过代码，就不算通关。\n")
		return

	case args[0] == "list":
		Catalog(levels)
		return
	}

	picked, err := pick(levels, strings.Join(args, " "))
	if err != nil {
		fmt.Printf("❌ %v\n\n输入 go run . list 查看全部关卡\n", err)
		os.Exit(1)
	}

	for _, l := range picked {
		l.Print()
	}
	fmt.Printf("\n提示：本关思考题在源码注释里，改完再进下一关。\n")
}

// pick 解析关卡选择表达式，编号从 1 开始按注册顺序计数。
func pick(levels []Level, spec string) ([]Level, error) {
	var out []Level

	for _, part := range strings.FieldsFunc(spec, func(r rune) bool { return r == ' ' || r == ',' }) {
		lo, hi := part, ""
		if i := strings.Index(part, "-"); i > 0 {
			lo, hi = part[:i], part[i+1:]
		}

		start, err := strconv.Atoi(lo)
		if err != nil {
			return nil, fmt.Errorf("无效关卡编号 %q", part)
		}
		end := start
		if hi != "" {
			if end, err = strconv.Atoi(hi); err != nil {
				return nil, fmt.Errorf("无效关卡区间 %q", part)
			}
		}

		if start < 1 || end > len(levels) || start > end {
			return nil, fmt.Errorf("关卡范围 %s 越界（有效：1-%d）", part, len(levels))
		}
		out = append(out, levels[start-1:end]...)
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("没有匹配到关卡：%q", spec)
	}
	return out, nil
}
