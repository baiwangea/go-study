package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gogf/gf/v2/net/ghttp"

	"go-study/goframe/level"
)

// L3-01：第一个 GoFrame 服务。
func L01() level.Level {
	return level.Level{
		ID:      "L3-01",
		Title:   "g.Server 与路由注册",
		Tags:    "ghttp · 路由 · 优雅关闭",
		Pre:     "L2-11（context 取消）",
		Goal:    "跑起一个 GoFrame HTTP 服务，看清「Server 实例 + 路由表 + Start/Shutdown」三件事",
		Observe: "临时服务监听随机端口，两个路由都返回内容，关闭后端口释放",
		Questions: []string{
			"ghttp.GetServer(\"name\") 里换一个名字会多出一个实例；用同一个名字会怎样？动手打印 s.Name() 验证。",
			"SetDumpRouterMap(false) 与 SetAccessLogEnabled(false) 各关掉了什么？生产环境为什么建议开着访问日志？",
			"路由写法有 BindHandler / Group.Bind / REST 资源路由三种，同一路径重复注册会覆盖还是报错？",
		},
		Check: "能说清一个 Server 从创建到优雅关闭的全过程，并解释关卡里为什么要随机端口",
		Run: func() {
			withServer(func(s *ghttp.Server) {
				s.BindHandler("GET:/hello", func(r *ghttp.Request) {
					r.Response.Write("Hello GoFrame")
				})
				s.BindHandler("GET:/order/{id}", func(r *ghttp.Request) {
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
			fmt.Println("  ↑ withServer 退出时已调用 s.Shutdown()：停止接新连接并排空在途请求")
		},
	}
}
