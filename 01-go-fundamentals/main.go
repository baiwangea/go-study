// ── L1 阶段 · Go 语言基础（关卡导航）───────────────────────────────
//
// 前置说明：只需要会用 go run 命令，不需要任何 Go 经验。
// 本模块目标：10 关打通「函数 → 包 → 接口」三个地基概念。
// 使用契约：
//  1. 一关一个文件（functions/01_xxx.go 这种命名），不要一次读完整模块；
//  2. 每关只看四行：目标 / 观察 / 思考 / 通关；
//  3. 先跑 go run . 4 看输出，再打开那个文件改代码验证思考题；
//  4. 思考题动手改过了，才算通关，可以进入下一关。
//
// 关卡分布：
//
//	L1-01 ~ L1-05  functions/   函数与返回值 · 多返回值 · 变长参数 · 闭包 · 递归
//	L1-06 ~ L1-07  packages/    导出规则 · 导入写法
//	L1-08 ~ L1-10  interfaces/  接口定义 · 隐式实现 · 多态与类型断言
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"go-study/go-fundamentals/functions"
	"go-study/go-fundamentals/interfaces"
	"go-study/go-fundamentals/level"
	"go-study/go-fundamentals/packages"
)

// all 按学习顺序汇总所有关卡。新增主题时在这里追加一行即可。
func all() []level.Level {
	var levels []level.Level
	levels = append(levels, functions.Levels()...)
	levels = append(levels, packages.Levels()...)
	levels = append(levels, interfaces.Levels()...)
	return levels
}

func main() {
	levels := all()
	args := os.Args[1:]

	switch {
	case len(args) == 0:
		fmt.Printf("====== L1 Go 语言基础：%d 关按序通关 ======\n", len(levels))
		for _, l := range levels {
			l.Print()
		}
		fmt.Printf("\n✅ 全部关卡运行完毕。别停在“看完了”——回到源码把每关的思考题改一遍才算通关。\n")
		return

	case args[0] == "list":
		level.Catalog(levels)
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
	fmt.Printf("\n提示：本关思考题在源码注释里，改完再进下一关。（go run . list 看目录）\n")
}

// pick 支持 "3"、"3-5"、"3 5" 三种写法，编号从 1 开始按注册顺序计数。
func pick(levels []level.Level, spec string) ([]level.Level, error) {
	var out []level.Level

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
