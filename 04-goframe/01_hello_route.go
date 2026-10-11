//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 04-goframe && go run 01_hello_route.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gogf/gf/v2/net/ghttp"

	"go-study/goframe/demoserver"
)

/*
═══════════════════════════════════════════════════════════════════════
L3-01：g.Server 与路由注册
═══════════════════════════════════════════════════════════════════════

【核心概念】
    跑起一个 GoFrame HTTP 服务，看清「Server 实例 + 路由表 + Start/Shutdown」三件事：
      · ghttp.GetServer("name") 按名字取/建单例 Server；
      · BindHandler("METHOD:/path", fn) 把处理函数挂到路由表；
      · Start() 后台监听、Shutdown() 优雅停止（不再接新连接、排空在途请求）。

【你会看到】
    ✓ 临时服务监听随机端口，两个路由都返回内容，关闭后端口释放

【思考题】（动手改代码验证，别空想）
    1. ghttp.GetServer("name") 里换一个名字会多出一个实例；用同一个名字会怎样？动手打印 s.Name() 验证。
    2. SetDumpRouterMap(false) 与 SetAccessLogEnabled(false) 各关掉了什么？生产环境为什么建议开着访问日志？
    3. 路由写法有 BindHandler / Group.Bind / REST 资源路由三种，同一路径重复注册会覆盖还是报错？

【通关标准】
    能说清一个 Server 从创建到优雅关闭的全过程，并解释关卡里为什么要随机端口

【前置】L2-11（context 取消）      【标签】ghttp · 路由 · 优雅关闭
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n=== L3-01 g.Server 与路由注册 ===")

	// ── 示例 1：注册两条路由，再起服务请求它们 ───────────────────────────
	// demoserver.WithServer 干三件事：起一个随机端口的临时服务 → 执行 fn 发请求
	// → 跑完自动优雅关闭。用随机端口是为了不和本机其它服务、以及其它关卡抢端口。
	fmt.Println("\n【示例 1】随机端口起服务 + 两条路由")
	demoserver.WithServer(func(s *ghttp.Server) {
		s.BindHandler("GET:/hello", func(r *ghttp.Request) {
			r.Response.Write("Hello GoFrame")
		})
		s.BindHandler("GET:/order/{id}", func(r *ghttp.Request) {
			// {id} 是路径参数，用 GetRouter("id") 取回
			r.Response.Writef("查询订单 %s", r.GetRouter("id").String())
		})
	}, func(base string) {
		for _, path := range []string{"/hello", "/order/ORD-1001"} {
			resp, err := http.Get(base + path)
			if err != nil {
				fmt.Println("  请求失败：", err)
				continue
			}
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			fmt.Printf("  GET %-18s → %s %q\n", path, resp.Status, body)
		}
	})

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n【关键总结】")
	fmt.Println("  ↑ demoserver.WithServer 退出时已调用 s.Shutdown()：停止接新连接并排空在途请求")
	fmt.Println("  • Server 实例由 ghttp.GetServer(name) 管理，同名即同实例")
	fmt.Println("  • BindHandler(\"METHOD:/path\", fn) 注册路由；{id} 用 GetRouter(\"id\") 取")
	fmt.Println("  • 关卡用随机端口，避免与其它服务/关卡冲突")
}
