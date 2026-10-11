//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 04-goframe && go run 04_config.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

/*
═══════════════════════════════════════════════════════════════════════
L3-04：g.Cfg 读取配置
═══════════════════════════════════════════════════════════════════════

【核心概念】
    把策略参数、密钥、端口从代码里搬进配置文件（manifest/config/config.yaml），
    用点号路径一行读到强类型值：g.Cfg().Get(ctx, "a.b.c") 返回 *gvar.Var，再 .String()/.Bool()/...

【你会看到】
    ✓ config.yaml 里的 app / trade.maxPositionUsd / trade.risk.stopLossPct / symbols 都能直接取出

【思考题】（动手改代码验证，别空想）
    1. Get 与 MustGet 的差别是什么？启动期与请求期分别该用哪个？
    2. 把 key 写成 trade.maxPositionUsd2，err 是什么？*gvar.Var 的零值会带来什么隐患？
    3. 生产环境的 API Key 不该进 Git —— 动手试：把值改成从环境变量读取（os.Getenv）或挂载独立配置文件。

【通关标准】
    能把一组策略参数写进 manifest/config/config.yaml 并在代码中强类型读出

【前置】L3-01      【标签】manifest/config · 点号路径 · 类型转换
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n=== L3-04 g.Cfg 读取配置 ===")
	ctx := gctx.New()

	// ── 示例 1：用点号路径读取强类型值 ───────────────────────────────────
	fmt.Println("\n【示例 1】config.yaml 里的值按点号路径取出，自动做类型转换")
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

	// ── 示例 2：读不存在的 key 不报错，静默兜底成零值 ────────────────────
	fmt.Println("\n【示例 2】读不存在的 key：err 有值，但值本身是零值")
	missing, err := g.Cfg().Get(ctx, "trade.notExist")
	fmt.Printf("  不存在的 key → err=%v，值=%q ← 不判空就会静默兜底成零值\n", err, missing.String())

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n【关键总结】")
	fmt.Println("  • 策略参数/密钥/端口都搬进 manifest/config/config.yaml，用点号路径一行读到强类型值")
	fmt.Println("  • Get 返回 (*gvar.Var, error)；启动期用 MustGet（缺失即早暴露），请求期用 Get")
	fmt.Println("  • 配置缺失会静默兜底成零值，必须自己判 err")
}
