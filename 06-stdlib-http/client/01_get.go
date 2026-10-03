package client

import (
	"fmt"
	"io"
	"net/http"

	"go-study/stdlib-http/level"
)

// L4-01：HTTP 客户端最小闭环。
func L01() level.Level {
	return level.Level{
		ID:      "L4-01",
		Title:   "Get / 读 Body / 必须 Close",
		Tags:    "http.Get · Response.Body · io.ReadAll",
		Pre:     "L1-02（多返回值与 error）",
		Goal:    "跑通一次请求-响应，并记住 Body 必须读完且 Close，否则连接不能复用",
		Observe: "状态码 200、自定义响应头、body 内容 pong；resp.Body 是 io.ReadCloser 不是字符串",
		Questions: []string{
			"删掉 defer resp.Body.Close()，连跑 1000 次会怎样？（提示：连接泄漏，配合 runtime 或 ss 观察）",
			"resp.Body 不读直接 Close，和读完再 Close，对连接复用（keep-alive）有什么区别？",
			"http.Get 与 http.NewRequest + client.Do 相比少了什么控制能力？（下一关补上）",
		},
		Check: "能说清 Body 的读取与关闭责任，以及为什么官方要求「读完并关闭」",
		Run: func() {
			srv := newDemoServer()
			defer srv.Close()

			resp, err := http.Get(srv.URL + "/ping")
			if err != nil {
				fmt.Println("  请求失败：", err)
				return
			}
			defer resp.Body.Close() // 必须：连接才能归还复用

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("  读取失败：", err)
				return
			}

			fmt.Printf("  %s %s\n", resp.Proto, resp.Status)
			fmt.Printf("  自定义响应头 X-Demo = %q\n", resp.Header.Get("X-Demo"))
			fmt.Printf("  body = %q（%d 字节，类型是 []byte，需自行解析）\n", body, len(body))
		},
	}
}
