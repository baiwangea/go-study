// ── L2 阶段 · Go 并发模型（关卡导航）───────────────────────────
//
// 前置说明：完成 L1（尤其 L1-04 闭包、L1-08~10 接口）。
// 本模块目标：13 关打通 goroutine / channel / select / context / 锁 五件事，
//
//	终点能力是「写一个能限流、能超时、能取消、不泄漏的并发骨架」。
//
// 使用契约：
//  1. 一关一个文件（channels/01_unbuffered.go 这种），一次只读一关；
//  2. 每关先看四行：目标 / 观察 / 思考 / 通关；
//  3. 关卡里的「思考题」必须动手改代码验证，改出编译错误最好 —— 那是最快的学习方式；
//  4. L2-13 结束后请务必跑一次 go run -race . 13，看懂 DATA RACE 报告。
//
// 关卡分布：
//
//	L2-01 ~ L2-02  goroutines/  go 语句与调度不确定性、WaitGroup
//	L2-03 ~ L2-06  channels/    无缓冲、缓冲与背压、close/range、死锁与泄漏
//	L2-07 ~ L2-10  selects/     多路复用、超时、非阻塞丢弃、ticker 轮询
//	L2-11          contextpkg/  context 取消与协程回收
//	L2-12          patterns/    worker pool 并发限流（机器人主力骨架）
//	L2-13          mutexes/     data race、Mutex 与 atomic 选型
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"go-study/go-concurrency/channels"
	"go-study/go-concurrency/contextpkg"
	"go-study/go-concurrency/goroutines"
	"go-study/go-concurrency/level"
	"go-study/go-concurrency/mutexes"
	"go-study/go-concurrency/patterns"
	"go-study/go-concurrency/selects"
)

// all 按学习顺序汇总关卡；新增主题在这里追加一行。
func all() []level.Level {
	var levels []level.Level
	levels = append(levels, goroutines.Levels()...)
	levels = append(levels, channels.Levels()...)
	levels = append(levels, selects.Levels()...)
	levels = append(levels, contextpkg.Levels()...)
	levels = append(levels, patterns.Levels()...)
	levels = append(levels, mutexes.Levels()...)
	return levels
}

func main() {
	levels := all()
	args := os.Args[1:]

	switch {
	case len(args) == 0:
		fmt.Printf("====== L2 Go 并发模型：%d 关按序通关 ======\n", len(levels))
		for _, l := range levels {
			l.Print()
		}
		fmt.Printf("\n✅ 全部关卡运行完毕。再跑一次 go run -race . 13，确认你看懂了 DATA RACE 报告。\n")
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
	fmt.Printf("\n提示：思考题在源码注释里，改完再进下一关。（go run . list 看目录）\n")
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
