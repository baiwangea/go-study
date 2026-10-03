package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"go-study/stdlib-http/level"
)

// L4-07：请求级 context —— 客户端走了，服务端别白干。
func L07() level.Level {
	canceled := make(chan string, 1)

	slowHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done(): // 客户端断开 / 超时，都会让这里就绪
			canceled <- "服务端收到取消信号：" + r.Context().Err().Error()
			return
		case <-time.After(200 * time.Millisecond):
			fmt.Fprint(w, "行情聚合完成")
		}
	})

	return level.Level{
		ID:      "L4-07",
		Title:   "r.Context() 的取消传播",
		Tags:    "context · 取消传播 · 成本控制",
		Pre:     "L2-11, L4-02",
		Goal:    "理解请求 context 的生命周期：连接断开即取消，长任务必须监听它",
		Observe: "客户端 50ms 就放弃了，服务端在 200ms 的活儿被取消，并打印出取消原因",
		Questions: []string{
			"把客户端超时改成 300ms（比服务端 200ms 长），服务端还会被取消吗？输出有什么不同？",
			"ctx.Err() 在这里会是 canceled 还是 deadline exceeded？谁触发的取消？",
			"如果你正在查数据库/调用交易所，收到 Done 之后应该做什么？（提示：提前返回，别让协程继续堆积）",
		},
		Check: "能在 handler 里正确传递 r.Context() 给下游，并解释「客户端断开≠服务端停止」这个常见事故",
		Run: func() {
			srv := httptest.NewServer(slowHandler)
			defer srv.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/aggregate", nil)
			start := time.Now()
			resp, err := (&http.Client{}).Do(req)
			if err != nil {
				fmt.Printf("  客户端 %v 后放弃：%v\n", time.Since(start).Round(10*time.Millisecond), err)
			} else {
				resp.Body.Close()
				fmt.Println("  客户端拿到了响应（说明服务端比客户端快）")
			}

			select {
			case msg := <-canceled:
				fmt.Println(" ", msg)
			case <-time.After(400 * time.Millisecond):
				fmt.Println("   服务端没走到取消分支")
			}
		},
	}
}
