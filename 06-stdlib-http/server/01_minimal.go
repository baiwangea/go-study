package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"go-study/stdlib-http/level"
)

// L4-05：手写一个 HTTP 服务 —— 路由、优雅关闭。
func L05() level.Level {
	return level.Level{
		ID:      "L4-05",
		Title:   "http.Server 与优雅关闭",
		Tags:    "ServeMux · ListenAndServe · Shutdown",
		Pre:     "L4-01",
		Goal:    "理解框架帮你做的事：一个 Handler 接口 + 一个 Listener + 一个循环；并学会停机不丢请求",
		Observe: "真实端口（:0 随机分配）上完成一次请求；Shutdown 后服务不再接受新连接",
		Questions: []string{
			"把 srv.Shutdown(ctx) 换成什么都不做，程序会怎样？（goroutine 泄漏在 Serve 循环里）",
			"http.ListenAndServe 和 srv.Serve(ln) 的区别是什么？为什么生产代码要拿到 server 实例？",
			"Shutdown 与 Close 的差别：哪个会等待在途请求？（对照 14-asynq 的 srv.Shutdown 语义）",
		},
		Check: "能手写带超时配置的 http.Server，并解释优雅停机的三步（停止监听 → 等在途 → 退出）",
		Run: func() {
			mux := http.NewServeMux()
			mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, "ok")
			})
			mux.HandleFunc("POST /order", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
				fmt.Fprint(w, "order accepted")
			})

			ln, err := net.Listen("tcp", "127.0.0.1:0") // 端口 0 = 让内核分配，避免冲突
			if err != nil {
				fmt.Println("  监听失败：", err)
				return
			}

			srv := &http.Server{
				Handler:      mux,
				ReadTimeout:  2 * time.Second,
				WriteTimeout: 2 * time.Second,
			}
			go func() { _ = srv.Serve(ln) }()
			fmt.Printf("  服务已启动：%s（端口由系统分配）\n", ln.Addr())

			resp, err := http.Get("http://" + ln.Addr().String() + "/health")
			if err == nil {
				defer resp.Body.Close()
				fmt.Printf("  GET /health → %s\n", resp.Status)
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			if err := srv.Shutdown(ctx); err != nil {
				fmt.Println("  关闭异常：", err)
			}
			fmt.Println("  Shutdown 完成：不再接受新连接，在途请求已排空")
		},
	}
}
