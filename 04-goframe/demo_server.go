package main

import (
	"fmt"
	"net"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
)

// freePort 向内核要一个空闲端口，避免关卡之间、以及与本机其它服务冲突。
func freePort() int {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 18080
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port
}

// withServer 起一个独立的 GoFrame 服务实例，跑完 fn 后优雅关闭。
// 关键点：必须 SetPort，光靠实例名不会指定端口；返回后用 GetListenedPort 取真实端口。
func withServer(setup func(s *ghttp.Server), fn func(baseURL string)) {
	s := ghttp.GetServer(fmt.Sprintf("demo-%d", time.Now().UnixNano()))
	s.SetDumpRouterMap(false)
	s.SetAccessLogEnabled(false)
	s.SetPort(freePort())
	setup(s)

	go s.Start()
	defer s.Shutdown()

	port := 0
	for i := 0; i < 100 && port == 0; i++ {
		if p := s.GetListenedPort(); p > 0 {
			port = p
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if port == 0 {
		fmt.Println("  服务未能按时监听端口，跳过本关演示")
		return
	}

	fn(fmt.Sprintf("http://127.0.0.1:%d", port))
	time.Sleep(50 * time.Millisecond) // 等响应写回，避免关闭时截断
}
