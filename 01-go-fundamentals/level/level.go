// Package level 提供「一关一关学 Go」的最小运行时。
//
// 设计契约（全仓库通用，新增模块请照抄本文件）：
//   - 一个 Level = 一个知识点 = 一个小文件（30~50 行，能独立跑）
//   - 关卡必须自带：前置(pre) / 目标(goal) / 观察(observe) / 思考题(questions) / 通关标准(check)
//   - main.go 只负责导航：list 看目录、传编号跑单关、不传参数按序通关
package level

import (
	"fmt"
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
