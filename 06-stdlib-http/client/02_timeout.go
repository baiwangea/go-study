package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"go-study/stdlib-http/level"
)

// L4-02：超时 —— 默认客户端没有超时，这比没请求更危险。
func L02() level.Level {
	return level.Level{
		ID:      "L4-01+1",
		Title:   "Client.Timeout 与 context 超时",
		Tags:    "超时 · http.Client · Transport",
		Pre:     "L4-01",
		Goal:    "任何外部调用都必须有超时：学会 client.Timeout（整体）与 context（单请求）两条线",
		Observe: "默认 http.Get 慢接口等满 300ms；带 120ms 超时的客户端立刻报错退出",
		Questions: []string{
			"删掉 Timeout 字段，用 http.DefaultClient 请求 /slow —— 它等多久？如果对方永远不返回呢？（这就是默认客户端的陷阱）",
			"client.Timeout 和 context.WithTimeout 的区别：前者含连接+读body全过程，后者只覆盖到哪里？报错文案分别是什么？",
			"行情轮询里超时设多少合适？（提示：小于你的轮询间隔，否则会叠加堆积）",
		},
		Check: "能给任意请求加两层超时，并解释报错里 Client.Timeout exceeded 与 context deadline exceeded 的差别",
		Run: func() {
			srv := newDemoServer()
			defer srv.Close()

			start := time.Now()
			resp, err := http.Get(srv.URL + "/slow") // 无超时
			if err == nil {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
			fmt.Printf("  无超时客户端：耗时 %v，err=%v（等满了对方 300ms）\n", time.Since(start).Round(10*time.Millisecond), err)

			fast := &http.Client{Timeout: 120 * time.Millisecond}
			start = time.Now()
			resp2, err := fast.Get(srv.URL + "/slow")
			if err != nil {
				fmt.Printf("  120ms 超时客户端：耗时 %v，err=%v ← 快速失败\n", time.Since(start).Round(10*time.Millisecond), err)
			} else {
				resp2.Body.Close()
			}

			ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
			defer cancel()
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/slow", nil)
			resp3, err := (&http.Client{}).Do(req)
			if err != nil {
				fmt.Println("  context 超时：", err)
			} else {
				resp3.Body.Close()
			}
		},
	}
}
