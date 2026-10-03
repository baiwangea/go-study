package client

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"
)

// newDemoServer 起一个本地测试服务器，让关卡不依赖外网、也不会阻塞。
// 三个分支：/ping 正常、/slow 故意慢、/error 返回 500。
func newDemoServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Demo", "stdlib-http")
		fmt.Fprint(w, "pong")
	})
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond) // 模拟交易所慢响应
		fmt.Fprint(w, "slow-but-ok")
	})
	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
	})
	return httptest.NewServer(mux)
}
