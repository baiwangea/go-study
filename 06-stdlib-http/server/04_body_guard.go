package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"

	"go-study/stdlib-http/level"
)

type placeOrder struct {
	Symbol string  `json:"symbol"`
	Qty    float64 `json:"qty"`
}

// guardHandler 演示服务端三道基础防护：Body 限长、Content-Type、参数校验。
func guardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只接受 POST", http.StatusMethodNotAllowed)
		return
	}
	if ct := r.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		http.Error(w, "Content-Type 必须是 application/json", http.StatusUnsupportedMediaType)
		return
	}

	var in placeOrder
	limited := http.MaxBytesReader(w, r.Body, 128) // 关键：限制请求体大小
	body, err := io.ReadAll(limited)
	if err != nil {
		http.Error(w, "请求体过大或读取失败："+err.Error(), http.StatusRequestEntityTooLarge)
		return
	}
	if err := json.Unmarshal(body, &in); err != nil {
		http.Error(w, "JSON 解析失败："+err.Error(), http.StatusBadRequest)
		return
	}
	if in.Symbol == "" || in.Qty <= 0 {
		http.Error(w, "symbol 必填且 qty 必须大于 0", http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "已受理 %s x %s", in.Symbol, strconv.FormatFloat(in.Qty, 'f', -1, 64))
}

// L4-08：服务端输入防护 —— 任何外部输入都不可信。
func L08() level.Level {
	return level.Level{
		ID:      "L4-08",
		Title:   "Body 限长与参数校验",
		Tags:    "MaxBytesReader · 状态码语义 · 输入校验",
		Pre:     "L4-05",
		Goal:    "把后台接口写安全：限大小、校验类型与范围、返回正确的状态码而不是笼统 500",
		Observe: "同一接口对四种输入分别返回 200 / 413 / 400 / 415，四种情况互不混淆",
		Questions: []string{
			"去掉 MaxBytesReader，把一个 10MB 的 body 直接 io.ReadAll 会怎样？（内存被打爆，DoS 入门）",
			"为什么先判 Content-Type 再解析？直接解析会得到什么错误？状态码应该是 400 还是 415？",
			"把校验写成返回 {\"code\":-1,...} 的 200 响应，和返回真实 4xx 相比，运维与客户端各自付出什么代价？",
		},
		Check: "能独立写出一个带限流/限长/校验的控制后台接口，并解释每个状态码的选择",
		Run: func() {
			srv := httptest.NewServer(http.HandlerFunc(guardHandler))
			defer srv.Close()

			cases := []struct {
				name, method, ctype, body string
			}{
				{"正常下单", http.MethodPost, "application/json", `{"symbol":"BTC","qty":0.1}`},
				{"超大 body", http.MethodPost, "application/json", `{"symbol":"` + strings.Repeat("X", 300) + `"}`},
				{"数量非法", http.MethodPost, "application/json", `{"symbol":"BTC","qty":0}`},
				{"类型错误", http.MethodPost, "text/plain", `symbol=BTC`},
			}
			for _, c := range cases {
				resp, err := http.Post(srv.URL, c.ctype, strings.NewReader(c.body))
				if err != nil {
					fmt.Println("  ", c.name, "请求失败：", err)
					continue
				}
				b, _ := io.ReadAll(io.LimitReader(resp.Body, 80))
				resp.Body.Close()
				fmt.Printf("  %-10s → %s｜%s\n", c.name, resp.Status, strings.TrimSpace(string(b)))
			}
		},
	}
}
