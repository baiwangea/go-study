package main

import (
	"fmt"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"

	"go-study/goframe/level"
)

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

// L3-06：错误处理与错误码。
func L06() level.Level {
	ctx := gctx.New()

	return level.Level{
		ID:      "L3-06",
		Title:   "gerror：错误码、包装与判断",
		Tags:    "error 链 · gcode · Wrap · 可观测",
		Pre:     "L1-02, L3-05",
		Goal:    "把 PHP/Java 的异常思维换成「错误值 + 错误码」：错误可层层包装、可判类型、可带堆栈",
		Observe: "一次失败下单能同时看到：对外提示、内部原因、完整堆栈、错误码枚举",
		Questions: []string{
			"对比 fmt.Errorf(\"%w\") 与 gerror.Wrap：后者多给了什么信息（堆栈/码）？多花了什么代价？",
			"gerror.Code(err) 拿到错误码后，中间件如何据此决定 HTTP 状态码与对外文案？",
			"什么时候该把 error 记入日志、什么时候该直接返回？两者都做会不会重复报警？",
		},
		Check: "能用 gerror 写出带码、可包装、可判断的错误链，并通过 g.Log().Error 打印堆栈",
		Run: func() {
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

			g.Log().Error(ctx, "下单失败（框架会把堆栈一起落盘）", err)
		},
	}
}
