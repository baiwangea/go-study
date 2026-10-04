package main

import (
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"

	"go-study/goframe/level"
)

// L3-05：日志 —— 分级、字段、落盘。
func L05() level.Level {
	ctx := gctx.New()

	return level.Level{
		ID:      "L3-05",
		Title:   "glog：分级、字段与落盘",
		Tags:    "日志级别 · 结构化字段 · 按天分文件",
		Pre:     "L3-04",
		Goal:    "用框架日志代替 fmt.Println：分级、带字段、按日期落盘、可按环境调整级别",
		Observe: "四行不同级别日志；把级别调到 ERRO 之后，Info 不再输出、Error 仍在",
		Questions: []string{
			"把 SetLevel 换成 glog.LEVEL_PROD（只留 WARN/ERRO/CRIT），线上为什么会这么配？",
			"g.Log() 与 ctx 有什么关系？同一请求链路里传同一个 ctx 能带来什么便利？",
			"机器人里「下单被交易所拒绝」该用 Error 还是 Warning？判断标准是什么（是否需要人立即处理）？",
		},
		Check: "会设置日志路径与级别，并能在日志里看到自己传入的结构化字段",
		Run: func() {
			log := g.Log()
			_ = log.SetPath("runtime/log") // 落盘目录，按天滚动
			log.Info(ctx, "策略启动", g.Map{"symbol": "BTCUSDT", "strategy": "ma-cross"})
			log.Notice(ctx, "信号触发", g.Map{"signal": "BUY", "price": 63000.5})
			log.Warning(ctx, "滑点偏大", g.Map{"slippagePct": 0.42})
			log.Error(ctx, "下单被交易所拒绝", g.Map{"code": -1111, "msg": "Invalid order rate limit."})
			fmt.Println("  ↑ 以上四行同时写入 runtime/log/ 下按日期命名的文件")

			log.SetLevel(glog.LEVEL_ERRO) // 只保留 ERRO 及以上
			fmt.Println("  ↓ 级别调到 ERRO 后，这行 Info 不应出现：")
			log.Info(ctx, "这条应该被过滤掉")
			log.Error(ctx, "这条 ERRO 仍会输出")

			log.SetLevel(glog.LEVEL_ALL) // 复原，避免影响其它关卡
		},
	}
}
