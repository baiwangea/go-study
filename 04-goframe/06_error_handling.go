//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 04-goframe && go run 06_error_handling.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

/*
═══════════════════════════════════════════════════════════════════════
L3-06：gerror：错误码、包装与判断
═══════════════════════════════════════════════════════════════════════

【核心概念】
    把 PHP/Java 的异常思维换成「错误值 + 错误码」：错误可层层包装、可判类型、可带堆栈。
    gerror.NewCode 造带码错误，gerror.WrapCode 在保留内部原因的同时套一层对外提示。

【你会看到】
    ✓ 一次失败下单能同时看到：对外提示、内部原因、完整堆栈、错误码枚举

【思考题】（动手改代码验证，别空想）
    1. 对比 fmt.Errorf("%w") 与 gerror.Wrap：后者多给了什么信息（堆栈/码）？多花了什么代价？
    2. gerror.Code(err) 拿到错误码后，中间件如何据此决定 HTTP 状态码与对外文案？
    3. 什么时候该把 error 记入日志、什么时候该直接返回？两者都做会不会重复报警？

【通关标准】
    能用 gerror 写出带码、可包装、可判断的错误链，并通过 g.Log().Error 打印堆栈

【前置】L1-02, L3-05      【标签】error 链 · gcode · Wrap · 可观测
═══════════════════════════════════════════════════════════════════════
*/

// checkRisk 演示「带错误码 + 带堆栈 + 可包装」的错误链。
func checkRisk(usd float64, limit float64) error {
	if usd > limit {
		return gerror.NewCode(gcode.CodeValidationFailed, "超出单笔最大仓位限制")
	}
	return nil
}

func placeOrder(usd float64, limit float64) error {
	if err := checkRisk(usd, limit); err != nil {
		return gerror.WrapCode(gcode.CodeBusinessValidationFailed, err, "下单被风控拦截")
	}
	return nil
}

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L3-06 gerror：错误码、包装与判断")
	fmt.Println(strings.Repeat("=", 60) + "\n")
	ctx := gctx.New()

	// ── 示例 1：一次失败下单，拿到完整错误链 ─────────────────────────────
	fmt.Println("📝 示例 1：带码、可包装、可判断的错误链")
	fmt.Println(strings.Repeat("-", 60))
	err := placeOrder(900, 500)
	fmt.Println("  对外提示：", err.Error())
	fmt.Println("  错误码：  ", gerror.Code(err), "｜业务码：", gcode.CodeBusinessValidationFailed)
	fmt.Println("  是否同码：", gerror.HasCode(err, gcode.CodeBusinessValidationFailed))
	fmt.Println("  堆栈（截断展示）：")
	stack := gerror.Stack(err)
	if len(stack) > 220 {
		stack = stack[:220] + "…"
	}
	fmt.Print(stack)

	// ── 示例 2：错误记入日志，框架会把堆栈一起落盘 ───────────────────────
	fmt.Println("📌 示例 2：用 g.Log().Error 打印错误（含堆栈）")
	fmt.Println(strings.Repeat("-", 60))
	g.Log().Error(ctx, "下单失败（框架会把堆栈一起落盘）", err)

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ 用「错误值 + 错误码」代替 PHP/Java 的异常：可层层包装、可判类型、可带堆栈")
	fmt.Println("  ✅ gerror.WrapCode 比 fmt.Errorf(\"%w\") 多给了错误码与堆栈")
	fmt.Println("  ✅ gerror.Code(err) 拿到码后，中间件据此决定 HTTP 状态码与对外文案")
}
