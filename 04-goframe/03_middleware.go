package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"go-study/goframe/level"
)

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

// L3-03：标准出入参 + 统一响应中间件。
func L03() level.Level {
	return level.Level{
		ID:      "L3-03",
		Title:   "标准出入参与统一响应",
		Tags:    "g.Meta · MiddlewareHandlerResponse · 洋葱模型",
		Pre:     "L3-02",
		Goal:    "让 handler 只剩业务代码：入参绑定校验、返回包装、错误转 JSON 全部交给框架",
		Observe: "成功与失败都是同一种 JSON（code/message/data）；缺参数时框架直接给出校验错误",
		Questions: []string{
			"把 s.Use 的顺序换成 (timingMiddleware, ghttp.MiddlewareHandlerResponse)，耗时日志和响应包装的先后关系会变吗？",
			"handler 里改成 panic(\"boom\")，看框架默认怎么处理（对照 03-go-concurrency 的 withRecover）。",
			"对比 L3-02 的手写 r.Parse：标准出入参少了哪些样板代码？代价是什么（必须先掌握 g.Meta 规则）？",
		},
		Check: "能用 g.Meta + 标准出入参写一个接口，并解释统一响应中间件对签名的要求",
		Run: func() {
			_ = withServer(func(s *ghttp.Server) {
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
		},
	}
}
