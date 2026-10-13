//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 04-goframe && go run 03_middleware.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"go-study/goframe/demoserver"
)

/*
═══════════════════════════════════════════════════════════════════════
L3-03：标准出入参与统一响应
═══════════════════════════════════════════════════════════════════════

【核心概念】
    让 handler 只剩业务代码：入参绑定校验、返回包装、错误转 JSON 全部交给框架。
    g.Meta 声明路由与方法，字段上的 v 标签负责校验，返回值由统一响应中间件包装。

【你会看到】
    ✓ 成功与失败都是同一种 JSON（code/message/data）；缺参数时框架直接给出校验错误

【思考题】（动手改代码验证，别空想）
    1. 把 s.Use 的顺序换成 (timingMiddleware, ghttp.MiddlewareHandlerResponse)，耗时日志和响应包装的先后关系会变吗？
    2. handler 里改成 panic("boom")，看框架默认怎么处理（对照 03-go-concurrency 的 withRecover）。
    3. 对比 L3-02 的手写 r.Parse：标准出入参少了哪些样板代码？代价是什么（必须先掌握 g.Meta 规则）？

【通关标准】
    能用 g.Meta + 标准出入参写一个接口，并解释统一响应中间件对签名的要求

【前置】L3-02      【标签】g.Meta · MiddlewareHandlerResponse · 洋葱模型
═══════════════════════════════════════════════════════════════════════
*/

// TickReq / TickRes 是 GoFrame 的「标准出入参」写法：
// g.Meta 声明路由与方法，字段上的 v 标签负责校验。
type TickReq struct {
	g.Meta `path:"/tick" method:"get" tags:"行情"`
	Symbol string `json:"symbol" v:"required|length:2,20#交易对必填|交易对长度需在 2-20"`
}

type TickRes struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
	At     int64   `json:"at"`
}

// tickHandler 只写业务：入参已被框架绑定并校验，返回值由中间件统一包装。
func tickHandler(ctx context.Context, req *TickReq) (res *TickRes, err error) {
	if req.Symbol == "FAKE" {
		return nil, gerror.NewCode(gcode.CodeValidationFailed, "该交易对不存在: "+req.Symbol)
	}
	return &TickRes{Symbol: req.Symbol, Price: 63000.5, At: time.Now().Unix()}, nil
}

// timingMiddleware 手写中间件：Next() 之前是「进」，之后是「出」——洋葱模型。
func timingMiddleware(r *ghttp.Request) {
	start := time.Now()
	r.Middleware.Next()
	fmt.Printf("  [中间件] %s → %d，耗时 %v\n", r.URL.Path, r.Response.Status, time.Since(start).Round(time.Microsecond*100))
}

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L3-03 标准出入参与统一响应")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// ── 示例 1：注册中间件与 handler，再请求三种输入 ────────────────────
	fmt.Println("📝 示例 1：成功 / 业务错误 / 缺参数，都走同一种 JSON")
	fmt.Println(strings.Repeat("-", 60))
	demoserver.WithServer(func(s *ghttp.Server) {
		// 中间件顺序：先统一响应包装，再计时（洋葱模型，Next() 之前进、之后出）
		s.Use(ghttp.MiddlewareHandlerResponse, timingMiddleware)
		s.BindHandler("/tick", tickHandler)
	}, func(base string) {
		for _, path := range []string{"/tick?symbol=BTCUSDT", "/tick?symbol=FAKE", "/tick"} {
			resp, err := http.Get(base + path)
			if err != nil {
				fmt.Println("  请求失败：", err)
				continue
			}
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			fmt.Printf("  GET %-24s → %s %s\n", path, resp.Status, body)
		}
	})

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ handler 只写业务：入参绑定校验、返回包装、错误转 JSON 全交给框架")
	fmt.Println("  ✅ 业务错误仍返回 HTTP 200，信息在 body 的 code/message（51 = Validation Failed）")
	fmt.Println("  ✅ 中间件是洋葱模型：Next() 之前是「进」，之后是「出」")
}
