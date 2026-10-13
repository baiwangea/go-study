//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 04-goframe && go run 05_logging.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
)

/*
═══════════════════════════════════════════════════════════════════════
L3-05：glog：分级、字段与落盘
═══════════════════════════════════════════════════════════════════════

【核心概念】
    用框架日志代替 fmt.Println：分级（Info/Notice/Warning/Error）、
    带结构化字段、按日期落盘、可按环境调整级别。

【你会看到】
    ✓ 四行日志带同一 trace id；ERRO 级别框架自动附调用堆栈；级别调到 ERRO 后 Info 被过滤

【思考题】（动手改代码验证，别空想）
    1. 把 SetLevel 换成 glog.LEVEL_PROD（只留 WARN/ERRO/CRIT），线上为什么会这么配？
    2. g.Log() 与 ctx 有什么关系？同一请求链路里传同一个 ctx 能带来什么便利？
    3. 机器人里「下单被交易所拒绝」该用 Error 还是 Warning？判断标准是什么（是否需要人立即处理）？

【通关标准】
    会设置日志路径与级别，并能在日志里看到自己传入的结构化字段

【前置】L3-04      【标签】日志级别 · 结构化字段 · 按天分文件
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L3-05 glog：分级、字段与落盘")
	fmt.Println(strings.Repeat("=", 60) + "\n")
	ctx := gctx.New()

	// ── 示例 1：四个级别 + 结构化字段 ────────────────────────────────────
	fmt.Println("📝 示例 1：分级写日志，同一 ctx 共享 trace id")
	fmt.Println(strings.Repeat("-", 60))
	log := g.Log()
	_ = log.SetPath("runtime/log") // 落盘目录，按天滚动
	log.Info(ctx, "策略启动", g.Map{"symbol": "BTCUSDT", "strategy": "ma-cross"})
	log.Notice(ctx, "信号触发", g.Map{"signal": "BUY", "price": 63000.5})
	log.Warning(ctx, "滑点偏大", g.Map{"slippagePct": 0.42})
	log.Error(ctx, "下单被交易所拒绝", g.Map{"code": -1111, "msg": "Invalid order rate limit."})
	fmt.Println("  ➡️  以上四行同时写入 runtime/log/ 下按日期命名的文件")

	// ── 示例 2：调高级别，低级别日志被过滤 ───────────────────────────────
	fmt.Println("📌 示例 2：级别调到 ERRO 后，Info 不再输出")
	fmt.Println(strings.Repeat("-", 60))
	log.SetLevel(glog.LEVEL_ERRO) // 只保留 ERRO 及以上
	fmt.Println("  ↓ 级别调到 ERRO 后，这行 Info 不应出现：")
	log.Info(ctx, "这条应该被过滤掉")
	log.Error(ctx, "这条 ERRO 仍会输出")

	log.SetLevel(glog.LEVEL_ALL) // 复原，避免影响其它关卡

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ 用框架日志代替 fmt.Println：分级、带字段、按日期落盘")
	fmt.Println("  ✅ ERRO 级别框架自动附调用堆栈；同一链路传同一个 ctx 共享 trace id")
	fmt.Println("  ✅ SetLevel 可按环境调整：线上常配 LEVEL_PROD（只留 WARN/ERRO/CRIT）")
}
