//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 04-goframe && go run 02_params_binding.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"

	"go-study/goframe/demoserver"
)

/*
═══════════════════════════════════════════════════════════════════════
L3-02：r.Parse 自动绑定与校验
═══════════════════════════════════════════════════════════════════════

【核心概念】
    告别手写 ParseForm + 逐个 if 校验：一个 r.Parse(&req) 完成
    「接收 → 类型转换 → 按 v 标签校验」，非法值根本进不了业务逻辑。

【你会看到】
    ✓ 四种输入分别得到：受理成功、缺字段、方向非法、数量过小，且都是中文提示

【思考题】（动手改代码验证，别空想）
    1. 把 v tag 全删掉再请求「数量过小」，非法值就会进入业务逻辑 —— 这就是校验贴着字段写的原因。
    2. 前端传 "qty":"0.1"（字符串）能绑定成功吗？为什么框架做得到而标准库 json 做不到？
    3. 把 Note 的 max-length 去掉，用 200 字的备注请求一次；想想数据库字段长度与这里校验如何对齐。

【通关标准】
    能用结构体 tag 同时定义 JSON 字段名与校验规则，并解释 # 分隔的自定义错误文案

【前置】L3-01      【标签】结构体绑定 · v 校验 · 自定义文案
═══════════════════════════════════════════════════════════════════════
*/

// PlaceOrderReq 把「字段名映射」和「校验规则」都写在结构体 tag 上。
// json tag 决定入参字段名；v tag 里 | 分隔多条规则、# 分隔自定义中文文案。
type PlaceOrderReq struct {
	Symbol string  `json:"symbol" v:"required|length:2,20#交易对必填|交易对长度 2-20"`
	Side   string  `json:"side" v:"required|in:BUY,SELL#方向必填#方向只能是 BUY 或 SELL"`
	Qty    float64 `json:"qty" v:"required|min:0.000001#数量必填#数量太小"`
	Note   string  `json:"note" v:"max-length:50#备注最长 50"`
}

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L3-02 r.Parse 自动绑定与校验")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// ── 示例 1：一个 handler 完成「接收 + 类型转换 + 校验」 ──────────────
	fmt.Println("📝 示例 1：r.Parse 一步到位，四种输入四种结果")
	fmt.Println(strings.Repeat("-", 60))
	demoserver.WithServer(func(s *ghttp.Server) {
		s.BindHandler("POST:/order", func(r *ghttp.Request) {
			var req *PlaceOrderReq
			if err := r.Parse(&req); err != nil {
				// 校验失败：框架已经把 v 标签里的中文文案拼好了，直接回给客户端
				r.Response.WriteStatus(http.StatusBadRequest, err.Error())
				return
			}
			// 通过校验才进入业务：这里只回显受理结果
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

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ r.Parse(&req) = 接收 + 类型转换 + 按 v 标签校验，三件事一步完成")
	fmt.Println("  ✅ v tag 里用 | 分隔规则、# 分隔自定义中文文案；未写文案的规则退回英文默认值")
	fmt.Println("  ✅ 校验贴着字段写，非法值根本进不了业务逻辑")
}
