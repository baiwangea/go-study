//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 04-goframe && go run 15_project_layout.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"go-study/goframe/demoserver"
	"go-study/goframe/internal/controller"
)

/*
═══════════════════════════════════════════════════════════════════════
L3-15：分层结构与依赖方向
═══════════════════════════════════════════════════════════════════════

【核心概念】
    把一个接口拆成三层：controller 翻译协议、logic 写业务、model 定义结构，
    依赖方向单向向下（main → controller → logic → model），任何反向依赖都算分层破坏。

【你会看到】
    ✓ 同一个 HTTP 请求经过三层；正常单成功、超限额与下架交易对各返回自己的业务错误

【思考题】（动手改代码验证，别空想）
    1. 如果直接把风控判断写进 controller，定时任务和命令行工具想复用这段逻辑要付出什么代价？
    2. 为什么 logic 里不该出现 *ghttp.Request？出现之后单元测试要怎么写？
    3. 目录里的 internal/ 相比 pkg/ 有什么保护意义？（提示：跨模块导入会被编译器拒绝）
    4. 用 `gf gen dao` 生成数据层时，生成代码应该放在哪一层？为什么不要手改生成文件？

【通关标准】
    能画出 api→controller→logic→model 的依赖图，并说清每层的禁止事项

【前置】L3-02, L3-03      【标签】controller · logic · model · 可测试性
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n=== L3-15 分层结构与依赖方向 ===")

	// ── 示例 1：一个路由挂上分层的 controller，四种输入四种结果 ─────────
	fmt.Println("\n【示例 1】/api/order 经过 controller → logic → model 三层")
	demoserver.WithServer(func(s *ghttp.Server) {
		s.Use(ghttp.MiddlewareHandlerResponse)
		s.Group("/", func(group *ghttp.RouterGroup) {
			// group.Bind 会读取 controller.Order 上各方法的 g.Meta 路由声明
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

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n【关键总结】")
	fmt.Println("  依赖方向：main → controller → logic → model，任何反向依赖都算分层破坏")
	fmt.Println("  • controller 只做协议翻译，不写业务、不碰 SQL")
	fmt.Println("  • logic 只写业务，不出现 *ghttp.Request、不出现 SQL，便于单元测试与复用")
	fmt.Println("  • internal/ 限制跨模块导入，编译器会拒绝越界引用")
}
