// 本文件是 T-03 的测试代码，演示两种测 HTTP 的姿势。
//
// 记住这条分界线：
//
//	直测 handler（NewRecorder + NewRequest） → 快、不占端口、覆盖 90% 用例
//	起真服务（NewServer + http.Client）      → 慢、占端口、但能测网络层行为
//
// 默认用第一种；只有「测客户端代码」「测超时」「测中间件链路」时才用第二种。

package orderapi

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// 姿势一：直测 handler。
//
// 三行骨架，背下来：
//
//	req := httptest.NewRequest(http.MethodPost, "/api/order", strings.NewReader(body))
//	rec := httptest.NewRecorder()
//	srv.Handler().ServeHTTP(rec, req)     // ← 这就是一次完整请求
//
// 注意 ServeHTTP 是同步调用：调用返回时响应已经全部写完，
// 所以可以立刻断言 rec.Code / rec.Body。不需要任何「等待」或「轮询」。
// ─────────────────────────────────────────────────────────────────────────────

func TestCreateOrder(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantSymbol string // 成功时的期望交易对（会被转成大写）
	}{
		{"正常下单", `{"symbol":"btcusdt","side":"buy","qty":0.5}`, http.StatusCreated, "BTCUSDT"},
		{"小写 side 也接受", `{"symbol":"ethusdt","side":"SELL","qty":1}`, http.StatusCreated, "ETHUSDT"},
		{"空 body", ``, http.StatusBadRequest, ""},
		{"非法 JSON", `{"symbol":`, http.StatusBadRequest, ""},
		{"缺 symbol", `{"side":"buy","qty":1}`, http.StatusBadRequest, ""},
		{"数量为 0", `{"symbol":"btcusdt","side":"buy","qty":0}`, http.StatusBadRequest, ""},
		{"数量为负", `{"symbol":"btcusdt","side":"buy","qty":-1}`, http.StatusBadRequest, ""},
		{"非法 side", `{"symbol":"btcusdt","side":"HOLD","qty":1}`, http.StatusBadRequest, ""},
		{"多余的字段被忽略", `{"symbol":"btcusdt","side":"buy","qty":1,"extra":"x"}`, http.StatusCreated, "BTCUSDT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := NewServer() // ← 每条用例一个干净实例，用例之间零污染

			req := httptest.NewRequest(http.MethodPost, "/api/order", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			srv.Handler().ServeHTTP(rec, req)

			// ── 断言 1：状态码 ────────────────────────────────────────────
			if rec.Code != tt.wantStatus {
				t.Fatalf("状态码 = %d, want %d（响应体：%s）", rec.Code, tt.wantStatus, rec.Body.String())
			}

			if tt.wantStatus != http.StatusCreated {
				return // 错误分支不关心响应体结构
			}

			// ── 断言 2：Content-Type ──────────────────────────────────────
			// 注意用 strings.HasPrefix 而不是 == ：真实值里带 charset。
			if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
				t.Errorf("Content-Type = %q, want 以 application/json 开头", ct)
			}

			// ── 断言 3：自定义 header ─────────────────────────────────────
			if id := rec.Header().Get("X-Order-Id"); id == "" {
				t.Errorf("缺少 X-Order-Id 响应头")
			}

			// ── 断言 4：响应体 —— 断言「结构」而不是「字符串」────────────
			// 如果这里写成 strings.Contains(body, `"status":"created"`)，
			// 那么给 Order 加一个字段、或者调一下字段顺序，测试就会碎。
			// 解成结构体再断言，是抗重构的写法。
			var got Order
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("响应体不是合法 JSON: %v（原文：%s）", err, rec.Body.String())
			}
			if got.Symbol != tt.wantSymbol {
				t.Errorf("Symbol = %q, want %q", got.Symbol, tt.wantSymbol)
			}
			if got.Status != "created" {
				t.Errorf("Status = %q, want %q", got.Status, "created")
			}
			if got.ID == "" {
				t.Errorf("ID 不该为空")
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 直测 handler 也能覆盖路由本身。
//
// 这里用 NewRequest + ServeMux 验证「路径参数被正确解析」和「未注册路由 404」。
// 顺带注意：GET /healthz 用的是最简单的 w.Write，不需要任何 JSON。
// ─────────────────────────────────────────────────────────────────────────────

func TestRoutes(t *testing.T) {
	t.Run("路径参数被正确解析", func(t *testing.T) {
		srv := NewServer()
		// 先造一单，拿到 ID
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/order",
			strings.NewReader(`{"symbol":"btcusdt","side":"buy","qty":1}`)))
		if rec.Code != http.StatusCreated {
			t.Fatalf("造数据失败: %d %s", rec.Code, rec.Body.String())
		}
		var created Order
		if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
			t.Fatalf("解析失败: %v", err)
		}

		// 再查它
		rec2 := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/api/order/"+created.ID, nil))

		if rec2.Code != http.StatusOK {
			t.Fatalf("查单状态码 = %d, want 200", rec2.Code)
		}
		var fetched Order
		if err := json.Unmarshal(rec2.Body.Bytes(), &fetched); err != nil {
			t.Fatalf("解析失败: %v", err)
		}
		if fetched.ID != created.ID {
			t.Errorf("查到的订单 ID = %q, want %q", fetched.ID, created.ID)
		}
	})

	t.Run("不存在的订单返回 404", func(t *testing.T) {
		srv := NewServer()
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/order/ord-9999", nil))

		if rec.Code != http.StatusNotFound {
			t.Errorf("状态码 = %d, want 404", rec.Code)
		}
	})

	t.Run("健康检查", func(t *testing.T) {
		srv := NewServer()
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

		if rec.Code != http.StatusOK {
			t.Fatalf("状态码 = %d, want 200", rec.Code)
		}
		if body := rec.Body.String(); body != "ok" {
			t.Errorf("body = %q, want %q", body, "ok")
		}
	})

	t.Run("方法不匹配返回 405", func(t *testing.T) {
		srv := NewServer()
		rec := httptest.NewRecorder()
		// /api/order 只注册了 POST，用 GET 访问应该 405
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/order", nil))

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("状态码 = %d, want 405（Go 1.22 的 ServeMux 会自动处理）", rec.Code)
		}
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// 姿势二：起真服务。
//
// httptest.NewServer 会在一个随机空闲端口上起一个真 HTTP 服务，
// 返回的 srv.URL 形如 http://127.0.0.1:53412。
// 用完必须 srv.Close()，否则端口和 goroutine 会一直挂着 ——
// 所以这里用 defer，或者 t.Cleanup（复习 T-02）。
//
// 什么时候非用它不可？
//   · 测的是「客户端代码」—— 你想验证自己的 HTTP client 拼对了 URL、带对了 header
//   · 测超时、重试、连接复用、TLS
//   · 测中间件链路（比如日志中间件真的把请求记下来了）
//
// 一个必须记住的点：**NewServer 起的是真服务，所以别忘了它也会真的走一遍
// http.Client 的全部行为 —— 包括自动跟随重定向、自动加 Accept-Encoding。**
// ─────────────────────────────────────────────────────────────────────────────

func TestRealServer(t *testing.T) {
	srv := NewServer()
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close) // ← 一定要关；用 t.Cleanup 比 defer 更稳（helper 里也能用）

	// 一次真实的「下单 → 查单」，全程走 TCP
	body := `{"symbol":"solusdt","side":"buy","qty":2.5}`
	resp, err := http.Post(ts.URL+"/api/order", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST 失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("状态码 = %d, want 201（响应体：%s）", resp.StatusCode, raw)
	}

	// 真服务下可以直接读 resp.Header —— 这在直测 handler 里只能读 rec.Header()
	if id := resp.Header.Get("X-Order-Id"); id == "" {
		t.Errorf("真服务下也应该带上 X-Order-Id")
	}

	var created Order
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	if created.Symbol != "SOLUSDT" {
		t.Errorf("Symbol = %q, want SOLUSDT", created.Symbol)
	}

	// 用同一个 client 查单
	getResp, err := http.Get(ts.URL + "/api/order/" + created.ID)
	if err != nil {
		t.Fatalf("GET 失败: %v", err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusOK {
		t.Errorf("查单状态码 = %d, want 200", getResp.StatusCode)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 只有真服务才测得出来的东西：客户端超时。
//
// 直测 handler 时 ServeHTTP 是同步调用，「超时」这个概念根本不存在 ——
// 你没有网络、没有 deadline。要测超时就必须起真服务 + 用带 Timeout 的 client。
//
// ⚠ 思考题 4：这个测试依赖真实时间（50ms 超时 vs 100ms 睡眠），
//   在负载高的 CI 机器上会变得不稳定。生产项目的做法是把时长做成可注入的参数。
// ─────────────────────────────────────────────────────────────────────────────

func TestClientTimeout(t *testing.T) {
	srv := NewServer()
	srv.SetSlow(100 * time.Millisecond) // 服务端故意慢

	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)

	client := &http.Client{Timeout: 50 * time.Millisecond} // 客户端等不了那么久

	_, err := client.Post(ts.URL+"/api/order", "application/json",
		strings.NewReader(`{"symbol":"btcusdt","side":"buy","qty":1}`))

	if err == nil {
		t.Fatal("期望超时错误，实际请求成功了 —— 检查 srv.SetSlow 和 client.Timeout 的时长关系")
	}
	// 用 errors.Is 判断而不是匹配错误字符串（复习 L1-19 错误链）
	if !strings.Contains(err.Error(), "deadline exceeded") && !strings.Contains(err.Error(), "timeout") {
		t.Errorf("期望超时类错误，实际: %v", err)
	}
	t.Logf("拿到预期的超时错误: %v", err)
}
