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
		Tags:    "ghttp · 路由 · 生命周期",
		Pre:     "L2-11（context 取消）",
		Goal:    "跑起一个 GoFrame HTTP 服务，理解「Server 实例 + 路由表 + Start/Shutdown」三件事",
		Observe: "同一进程里起了一个临时服务，GET /hello 返回文本，关闭后端口释放",
		Questions: []string{
			"把 g.Server() 换成 g.Server(\"name\")，两个实例会共用路由吗？动手打印两次 s.Name() 看看。",
			"SetDumpRouterMap(false) 关掉了什么输出？为什么生产环境建议开着？",
			"路由写法有三种：s.BindHandler / s.Group / s.SetHandler，各有什么差别？试着都换成同一路径看谁覆盖谁。",
		},
		Check: "能说清 Server 实例从创建到优雅关闭的完整过程，并解释为什么关卡里要随机端口",
		Run: func() {
			port := freePort()

			err := withServer(port, func(s *ghttp.Server) {
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
					fmt.Printf("  GET %-16s → %s %q\n", path, resp.Status, body)
				}
			})
			fmt.Println("  服务已优雅关闭，端口", port, "释放；withServer err =", err)
		},
	}
}
