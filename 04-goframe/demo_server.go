package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// freePort 向内核要一个空闲端口，避免关卡之间、以及与本机其它服务端口冲突。
func freePort() int {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 18080
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port
}

// withServer 启动一个独立的 GoFrame 服务实例，执行 fn 后再优雅关闭。
// 用「:端口」作为实例名，保证每一关拿到的是互不干扰的新服务对象。
func withServer(port int, setup func(s *ghttp.Server), fn func(baseURL string)) error {
	s := g.Server(fmt.Sprintf("127.0.0.1:%d", port))
	s.SetDumpRouterMap(false) // 关卡输出太多，关掉启动时的路由打印
	setup(s)

	go s.Start()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = s.Shutdown(ctx)
	}()

	// 等服务就绪（GoFrame 的 Start 是异步的）
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			conn.Close()
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	fn(fmt.Sprintf("http://127.0.0.1:%d", port))
	return nil
}
