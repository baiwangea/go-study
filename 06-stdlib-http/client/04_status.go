package client

import (
	"fmt"
	"io"
	"net/http"

	"go-study/stdlib-http/level"
)

// fetch 演示「err == nil 不等于成功」的正确判法。
func fetch(url string) (int, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, "", fmt.Errorf("网络层失败： %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
	if resp.StatusCode >= 400 {
		return resp.StatusCode, string(body), fmt.Errorf("HTTP 状态异常：%s", resp.Status)
	}
	return resp.StatusCode, string(body), nil
}

// L4-04：两个错误通道 —— 网络 err 与状态码。
func L04() level.Level {
	return level.Level{
		ID:      "L4-04",
		Title:   "err == nil 不代表成功",
		Tags:    "StatusCode · 错误分类 · 重试判断",
		Pre:     "L4-02",
		Goal:    "分清两类失败：传输层 err（连不上/超时）与业务层状态码（4xx/5xx），处理方式完全不同",
		Observe: "502 的请求 err 为 nil，状态码才是 502；不判断状态码就会把错误页当成成功",
		Questions: []string{
			"只写 if err != nil 就认为成功，把 /error 的响应体当数据解析会怎样？（拿到 HTML/错误文案继续跑）",
			"429（限频）、5xx（服务端）、4xx（参数错）分别该重试还是立刻失败？（提示：4xx 重试是雪崩制造机）",
			"把 fetch 的返回值改成 (body []byte, err error) 只透出错误状态，会不会丢掉 429 的 Retry-After 头？",
		},
		Check: "能写出一个区分「可重试 / 不可重试」的错误分类函数",
		Run: func() {
			srv := newDemoServer()
			defer srv.Close()

			code, body, err := fetch(srv.URL + "/ping")
			fmt.Printf("  /ping   → code=%d body=%q err=%v\n", code, body, err)

			code, body, err = fetch(srv.URL + "/error")
			fmt.Printf("  /error  → code=%d body=%q\n", code, body)
			fmt.Printf("          err=%v ← 注意 http.Get 的 err 其实是 nil\n", err)

			code, body, err = fetch("http://127.0.0.1:1/nope")
			fmt.Printf("  连不上  → code=%d err=%v ← 这才是传输层失败\n", code, err)
		},
	}
}
