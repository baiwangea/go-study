package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"

	"go-study/goframe/level"
)

// tickRes 业务只关心数据本身，包装结构交给统一响应中间件。
type tickRes struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
}

// timingMiddleware 手写中间件：GoFrame 的中间件就是 func(*ghttp.Request)，
// 调用 r.Middleware.Next() 之前的代码是「进」，之后的是「出」—— 洋葱模型。
func timingMiddleware(r *ghttp.Request) {
	start := time.Now()
	r.Middleware.Next()
	fmt.Printf("  [中间件] %s 耗时 %v\n", r.URL.Path, time.Since(start).Round(time.Millisecond))
}

// L3-03：中间件与统一响应。
func L03() level.Level {
	return level.Level{
		ID:      "L3-03",
		Title:   "中间件与 MiddlewareHandlerResponse",
		Tags:    "洋葱模型 · 统一响应 · 错误转 JSON",
		Pre:     "L3-02",
		Goal:    "让 handler 只返回 (数据, error)，包装、耗时日志、错误码全部交给中间件",
		Observe: "成功与失败输出同一种 JSON 形状（code/message/data），业务代码里一次手写包装都没有",
		Questions: []string{
			"把 Use 顺序换成 (timingMiddleware, ghttp.MiddlewareHandlerResponse)，耗时日志打在响应之前还是之后？为什么？",
			"handler 里 panic 会怎样？GoFrame 默认帮你做了什么？（对照 03-go-concurrency 里的 withRecover）",
			"错误的 message 直接透给前端安全吗？机器人后台该怎么区分「对外提示」与「内部错误详情」？",
		},
		Check: "能写自定义中间件，并解释 ghttp.MiddlewareHandlerResponse 对 handler 返回值的要求",
		Run: func() {
			port := freePort()

			_ = withServer(port, func(s *ghttp.Server) {
				s.Use(ghttp.MiddlewareHandlerResponse, timingMiddleware)

				s.BindHandler("GET:/tick/{symbol}", func(r *ghttp.Request) (any, error) {
					sym := r.GetRouter("symbol").String()
					if sym == "FAKE" {
						return nil, errors.New("该交易对不存在: " + sym)
					}
					time.Sleep(5 * time.Millisecond)
					return &tickRes{Symbol: sym, Price: 63000.5}, nil
				})
			}, func(base string) {
				for _, path := range []string{"/tick/BTCUSDT", "/tick/FAKE"} {
					resp, err := http.Get(base + path)
					if err != nil {
						fmt.Println("  请求失败：", err)
						continue
					}
					body, _ := io.ReadAll(resp.Body)
					resp.Body.Close()
					fmt.Printf("  GET %-18s → %s %s\n", path, resp.Status, body)
				}
			})
		},
	}
}
