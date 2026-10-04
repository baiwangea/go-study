package main

import (
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"

	"go-study/goframe/level"
)

// L3-04：配置 —— g.Cfg 的自动查找与类型安全读取。
func L04() level.Level {
	ctx := gctx.New()

	return level.Level{
		ID:      "L3-04",
		Title:   "g.Cfg 读取配置",
		Tags:    "manifest/config · 点号路径 · 类型转换",
		Pre:     "L3-01",
		Goal:    "把策略参数、密钥、端口从代码里搬进配置文件，用点号路径一行读到强类型值",
		Observe: "config.yaml 里的 app / trade.maxPositionUsd / trade.risk.stopLossPct / symbols 都能直接取出",
		Questions: []string{
			"Get 与 MustGet 的差别是什么？启动期与请求期分别该用哪个？",
			"把 key 写成 trade.maxPositionUsd2，err 是什么？*gvar.Var 的零值会带来什么隐患？",
			"生产环境的 API Key 不该进 Git —— 动手试：把值改成从环境变量读取（os.Getenv）或挂载独立配置文件。",
		},
		Check: "能把一组策略参数写进 manifest/config/config.yaml 并在代码中强类型读出",
		Run: func() {
			name, _ := g.Cfg().Get(ctx, "app.name")
			debug, _ := g.Cfg().Get(ctx, "app.debug")
			maxUsd, _ := g.Cfg().Get(ctx, "trade.maxPositionUsd")
			stopLoss, _ := g.Cfg().Get(ctx, "trade.risk.stopLossPct")
			symbols, _ := g.Cfg().Get(ctx, "trade.symbols")

			fmt.Printf("  app.name             = %q\n", name.String())
			fmt.Printf("  app.debug            = %v（字符串自动转 bool）\n", debug.Bool())
			fmt.Printf("  trade.maxPositionUsd = %.2f（自动转 float64）\n", maxUsd.Float64())
			fmt.Printf("  trade.risk.stopLossPct = %.1f%%\n", stopLoss.Float64())
			list := symbols.Strings()
			fmt.Printf("  trade.symbols        = %v（共 %d 个）\n", list, len(list))

			missing, err := g.Cfg().Get(ctx, "trade.notExist")
			fmt.Printf("  不存在的 key → err=%v，值=%q ← 不判空就会静默兜底成零值\n", err, missing.String())
		},
	}
}
