// Package demoserver 提供关卡演示用的「起一个真 GoFrame 服务、跑完优雅关闭」辅助。
//
// 为什么单独抽成一个包：L3-01 / L3-02 / L3-03 / L3-12 / L3-15 五个关卡都要起一个
// 真实 HTTP 服务来演示，而本模块「一个文件 = 一关 = 完整 main()」的约定意味着每个
// 关卡文件都是各自独立的 package main，无法共享同目录下的私有函数。与其把这 35 行
// 复制五遍，不如抽成一个真正可导入的包；关卡文件里写 demoserver.WithServer(...) 即可。
package demoserver

import (
	"fmt"
	"net"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
)

// FreePort 向内核要一个空闲端口，避免关卡之间、以及与本机其它服务冲突。
func FreePort() int {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 18080
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port
}

// WithServer 起一个独立的 GoFrame 服务实例，跑完 fn 后优雅关闭。
// 关键点：必须 SetPort，光靠实例名不会指定端口；返回后用 GetListenedPort 取真实端口。
func WithServer(setup func(s *ghttp.Server), fn func(baseURL string)) {
	s := ghttp.GetServer(fmt.Sprintf("demo-%d", time.Now().UnixNano()))
	s.SetDumpRouterMap(false)
	s.SetAccessLogEnabled(false)
	s.SetPort(FreePort())
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
