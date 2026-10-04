package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"

	"go-study/goframe/level"
)

// PlaceOrderReq 把「字段名映射」和「校验规则」都写在结构体 tag 上。
type PlaceOrderReq struct {
	Symbol string  `json:"symbol" v:"required|length:2,20#交易对必填|交易对长度 2-20"`
	Side   string  `json:"side" v:"required|in:BUY,SELL#方向必填#方向只能是 BUY 或 SELL"`
	Qty    float64 `json:"qty" v:"required|min:0.000001#数量必填#数量太小"`
	Note   string  `json:"note" v:"max-length:50#备注最长 50"`
}

// L3-02：参数接收、绑定与校验。
func L02() level.Level {
	return level.Level{
		ID:      "L3-02",
		Title:   "r.Parse 自动绑定与校验",
		Tags:    "结构体绑定 · v 校验 · 自定义文案",
		Pre:     "L3-01",
		Goal:    "告别手写 ParseForm + 逐个 if 校验：一个 r.Parse(&req) 完成接收、类型转换、校验",
		Observe: "四种输入分别得到：受理成功、缺字段、方向非法、数量过小，且都是中文提示",
		Questions: []string{
			"把 v tag 全删掉再请求「数量过小」，非法值就会进入业务逻辑 —— 这就是校验贴着字段写的原因。",
			"前端传 \"qty\":\"0.1\"（字符串）能绑定成功吗？为什么框架做得到而标准库 json 做不到？",
			"把 Note 的 max-length 去掉，用 200 字的备注请求一次；想想数据库字段长度与这里校验如何对齐。",
		},
		Check: "能用结构体 tag 同时定义 JSON 字段名与校验规则，并解释 # 分隔的自定义错误文案",
		Run: func() {
			withServer(func(s *ghttp.Server) {
				s.BindHandler("POST:/order", func(r *ghttp.Request) {
					var req *PlaceOrderReq
					if err := r.Parse(&req); err != nil {
						r.Response.WriteStatus(http.StatusBadRequest, err.Error())
						return
					}
					r.Response.WriteJson(g.Map{
						"order_id": fmt.Sprintf("ORD-%d", gtime.TimestampMilli()),
						"symbol":   req.Symbol,
						"side":     req.Side,
						"qty":      req.Qty,
					})
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
					out := strings.TrimSpace(string(body))
					if len(out) > 88 {
						out = out[:88] + "…"
					}
					fmt.Printf("  %-10s → %s %s\n", c.name, resp.Status, out)
				}
			})
		},
	}
}
