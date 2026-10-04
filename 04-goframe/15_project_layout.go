package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"go-study/goframe/internal/controller"
	"go-study/goframe/level"
)

// L3-15：工程分层 —— api / controller / logic / model 各层只允许向下依赖。
func L15() level.Level {
	return level.Level{
		ID:      "L3-15",
		Title:   "分层结构与依赖方向",
		Tags:    "controller · logic · model · 可测试性",
		Pre:     "L3-02, L3-03",
		Goal:    "把一个接口拆成三层：controller 翻译协议、logic 写业务、model 定义结构，依赖方向单向向下",
		Observe: "同一个 HTTP 请求经过三层；正常单成功、超限额与下架交易对各返回自己的业务错误",
		Questions: []string{
			"如果直接把风控判断写进 controller，定时任务和命令行工具想复用这段逻辑要付出什么代价？",
			"为什么 logic 里不该出现 *ghttp.Request？出现之后单元测试要怎么写？",
			"目录里的 internal/ 相比 pkg/ 有什么保护意义？（提示：跨模块导入会被编译器拒绝）",
			"用 `gf gen dao` 生成数据层时，生成代码应该放在哪一层？为什么不要手改生成文件？",
		},
		Check: "能画出 api→controller→logic→model 的依赖图，并说清每层的禁止事项",
		Run: func() {
			withServer(func(s *ghttp.Server) {
				s.Use(ghttp.MiddlewareHandlerResponse)
				s.Group("/", func(group *ghttp.RouterGroup) {
					group.Bind(controller.Order)
				})
			}, func(base string) {
				cases := []struct{ name, body string }{
					{"正常下单", `{"symbol":"BTCUSDT","side":"BUY","qty":0.001}`},
					{"超出限额", `{"symbol":"BTCUSDT","side":"BUY","qty":1}`},
					{"交易对下架", `{"symbol":"DELISTED","side":"BUY","qty":0.001}`},
					{"参数非法", `{"symbol":"B","side":"HOLD","qty":0}`},
				}
				for _, c := range cases {
					resp, err := http.Post(base+"/api/order", "application/json", strings.NewReader(c.body))
					if err != nil {
						fmt.Println("  ", c.name, "请求失败：", err)
						continue
					}
					b, _ := io.ReadAll(resp.Body)
					resp.Body.Close()
					fmt.Printf("  %-10s → %s %s\n", c.name, resp.Status, strings.TrimSpace(string(b)))
				}
			})
			fmt.Println("  依赖方向：main → controller → logic → model，任何反向依赖都算分层破坏")
		},
	}
}
