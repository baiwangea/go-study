package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"go-study/stdlib-http/level"
)

type middleware func(http.Handler) http.Handler

// withLogging 计时 + 状态码记录；next 是内层 http.Handler。
func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		fmt.Printf("  [中间件] %s %s 耗时 %v\n", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}

// withRecover 把 handler 的 panic 挡住，避免整个进程挂掉。
func withRecover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				fmt.Printf("  [中间件] 捕获 panic：%v → 返回 500\n", rec)
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// chain 按顺序把多个中间件套在 h 外面。
func chain(h http.Handler, mws ...middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// L4-06：中间件就是 http.Handler 接口的嵌套。
func L06() level.Level {
	return level.Level{
		ID:      "L4-06",
		Title:   "用接口实现中间件",
		Tags:    "http.Handler · 洋葱模型 · recover",
		Pre:     "L1-09, L4-05",
		Goal:    "看懂 Gin/Echo 中间件的真身：func(http.Handler) http.Handler，并自己实现一个",
		Observe: "先执行 recover 再进 logging（注册顺序决定洋葱层次）；panic 的接口仍返回 500 而不崩进程",
		Questions: []string{
			"把 chain 的注册顺序换成 (withRecover, withLogging)，日志还会打印 500 那一行吗？为什么？",
			"删掉 withRecover 让 handler panic，整个服务会怎样？（对照 14-asynq 的 RecoveryMiddleware 同样的道理）",
			"http.HandlerFunc 是什么？为什么函数能当 Handler？（提示：单方法接口的适配器）",
		},
		Check: "能独立写出一个统计耗时/记录状态码的中间件，并解释顺序带来的行为差异",
		Run: func() {
			biz := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/boom" {
					panic("行情字段为空")
				}
				time.Sleep(10 * time.Millisecond)
				fmt.Fprint(w, "strategy ok")
			})

			srv := httptest.NewServer(chain(biz, withLogging, withRecover))
			defer srv.Close()

			for _, path := range []string{"/tick", "/boom"} {
				resp, err := http.Get(srv.URL + path)
				if err != nil {
					fmt.Println("  请求失败：", err)
					continue
				}
				fmt.Printf("  GET %s → %s\n", path, resp.Status)
				resp.Body.Close()
			}
		},
	}
}
