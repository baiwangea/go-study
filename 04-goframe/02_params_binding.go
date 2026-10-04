package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"

	"go-study/goframe/level"
)

// PlaceOrderReq 演示参数绑定 + 校验规则（v tag 是 GoFrame 的招牌能力）。
type PlaceOrderReq struct {
	Symbol string  `json:"symbol" v:"required|length:2,20#交易对必填|交易对长度 2-20"`
	Side   string  `json:"side" v:"required|in:BUY,SELL#方向必填#方向只能是 BUY 或 SELL"`
	Qty    float64 `json:"qty" v:"required|min:0.000001#数量必填#数量太小"`
	Note   string  `json:"note" v:"max-length:50#备注最长 50"`
}

type PlaceOrderRes struct {
	OrderID string      `json:"order_id"`
	At      *gtime.Time `json:"at"`
}

// L3-02：参数接收、绑定与校验。
func L02() level.Level {
	return level.Level{
		ID:      "L3-02",
		Title:   "r.Parse 自动绑定与校验",
		Tags:    "结构体绑定 · v 校验 · 错误文案",
		Pre:     "L3-01",
		Goal:    "告别手写 ParseForm + if 校验：一个 r.Parse(&req) 完成接收、转换、校验",
		Observe: "合法请求返回受理结果；缺字段、方向非法、数量过小分别给出对应的中文校验文案",
		Questions: []string{
			"把 v tag 去掉，非法参数会怎样进入业务逻辑？（这就是为什么校验要贴着字段声明写）",
			"r.Parse 与 r.GetStruct 的区别是什么？后者会校验吗？",
			"如果交易所字段叫 price（float），前端传来的却是字符串 \"63000.5\"，绑定还能成功吗？为什么？",
		},
		Check: "能用结构体 tag 同时定义 JSON 字段名与校验规则，并解释 # 分隔的自定义错误文案",
		Run: func() {
			port := freePort()

			err := withServer(port, func(s *ghttp.Server) {
				s.BindHandler("POST:/order", func(r *ghttp.Request) {
					var req *PlaceOrderReq
					if err := r.Parse(&req); err != nil {
						r.Response.WriteStatus(http.StatusBadRequest, err.Error())
						return
					}
					res := &PlaceOrderRes{
						OrderID: fmt.Sprintf("ORD-%d", gtime.TimestampMilli()),
						At:      gtime.Now(),
					}
					r.Response.WriteJson(res)
				})
			}, func(base string) {
				cases := []struct{ name, body string }{
					{"合法下单", `{"symbol":"BTCUSDT","side":"BUY","qty":0.1}`},
					{"缺 symbol", `{"side":"BUY","qty":0.1}`},
					{"方向非法", `{"symbol":"BTCUSDT","side":"HOLD","qty":0.1}`},
					{"数量过小", `{"symbol":"BTCUSDT","side":"BUY","qty":0.000000001}`},
				}
				for _, c := range cases {
					resp, err := http.Post(base+"/order", "application/json", strings.NewReader(c.body))
					if err != nil {
						fmt.Println("  ", c.name, "请求失败：", err)
						continue
					}
					body, _ := io.ReadAll(resp.Body)
					resp.Body.Close()
					out := string(body)
					if len(out) > 96 {
						out = out[:96] + "…"
					}
					fmt.Printf("  %-10s → %s %s\n", c.name, resp.Status, out)
				}
			})
			fmt.Println("  withServer err =", err)
		},
	}
}
