package main

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/gogf/gf/v2/os/gcron"

	"go-study/goframe/level"
)

// L3-10：gcron 定时任务 —— 轮询行情、日报、超时取消的入口。
func L10() level.Level {
	return level.Level{
		ID:      "L3-10",
		Title:   "gcron：秒级 cron 与任务去重",
		Tags:    "gcron · 六段表达式 · AddOnce/AddTimes/Remove",
		Pre:     "L2-10（ticker 与 Stop）",
		Goal:    "用 gcron 替代手写 ticker：支持秒级表达式、命名任务、单实例与自动退出控制",
		Observe: "命名任务跑了约 2~3 次后被 Remove（次数取决于与秒边界的对齐）；AddOnce 只执行一次；延迟任务 200ms 后触发",
		Questions: []string{
			"表达式为什么是 6 段而不是 5 段？把 \"* * * * * *\" 改成 \"*/2 * * * * *\" 频率如何变化？",
			"Add 与 AddSingleton 的差别是什么？行情拉取超过 1 秒时，用 Add 会发生什么？",
			"机器人里「下单 15 秒未成交就撤单」该用 DelayAddOnce 还是队列延迟任务（L3-11）？各自的可靠性差别？",
		},
		Check: "会注册/停止命名定时任务，并能解释为什么长任务必须用 Singleton",
		Run: func() {
			ctx := context.Background()
			var ticks, once, delayed atomic.Int64

			// 命名任务：便于后续 Remove
			if _, err := gcron.Add(ctx, "* * * * * *", func(ctx context.Context) {
				n := ticks.Add(1)
				fmt.Printf("  [%s] 轮询行情 第 %d 次\n", time.Now().Format("15:04:05"), n)
			}, "poll-quote"); err != nil {
				fmt.Println("  注册失败：", err)
				return
			}

			_, _ = gcron.AddOnce(ctx, "* * * * * *", func(ctx context.Context) {
				once.Add(1)
			}, "report-once")

			gcron.DelayAddOnce(ctx, 200*time.Millisecond, "* * * * * *", func(ctx context.Context) {
				delayed.Add(1)
			}, "delayed")

			time.Sleep(2500 * time.Millisecond)
			gcron.Remove("poll-quote") // 停止并移除命名任务
			before := ticks.Load()
			time.Sleep(1100 * time.Millisecond)

			fmt.Printf("  移除前跑了 %d 次，移除后仍是 %d 次 → Remove 生效（命名任务不清理会常驻）\n", before, ticks.Load())
			fmt.Printf("  AddOnce 执行 %d 次；延迟任务执行 %d 次（200ms 后触发）\n", once.Load(), delayed.Load())
		},
	}
}
