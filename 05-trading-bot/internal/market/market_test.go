// Package market 的测试：
//
//   - RandomWalk：固定种子可复现、ctx 取消后输出 channel 一定会被 close、
//     产出的价格为正 —— 这三条对应 L2-05（只由生产者 close）与可复现性；
//   - parseWSTicker：纯解析函数，直接喂 JSON 断言；
//   - Bitget.FetchTicker：用 httptest 起本地假交易所，钉住「HTTP 状态错误 vs
//     业务 code 错误」的分流（README 里 L3 的重点）。
//
// 全部只用标准库，不碰网络。

package market

import (
	"context"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// 可复现性：同一个种子两次跑出完全一样的价格序列。
//
// 这是 mock 行情能拿来当测试输入的前提（也是 README「固定随机种子，输出可复现」那条）。
// 注意每个 RandomWalk 必须自持 rand 源 —— 共用同一个 *rand.Rand 会 DATA RACE（L2-13）。
// ─────────────────────────────────────────────────────────────────────────────

func TestRandomWalkReproducible(t *testing.T) {
	collect := func(seed int64) []float64 {
		f := RandomWalk{
			Symbol: "BTCUSDT",
			Start:  63000,
			Step:   time.Millisecond,
			Vol:    0.006,
			Rnd:    rand.New(rand.NewSource(seed)), // 每个源各自持有种子
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var prices []float64
		for q := range f.Subscribe(ctx, 12) {
			prices = append(prices, q.Price)
		}
		return prices
	}

	a := collect(7)
	b := collect(7)
	c := collect(8)

	if len(a) != 12 {
		t.Fatalf("收到 %d 笔行情, want 12", len(a))
	}
	if !slices.Equal(a, b) {
		t.Errorf("同种子(7)两次结果不一致：\n  a=%v\n  b=%v", a, b)
	}
	if slices.Equal(a, c) {
		t.Errorf("不同种子(7 vs 8)不该产出完全相同的序列：%v", a)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ctx 取消后，Subscribe 返回的 channel **一定会被 close**（L2-05：只有生产者 close）。
//
// Step 设成 1 小时，让行情协程确定地卡在 select 上，然后 cancel ——
// 若没有 defer close(out)，这里会一直阻塞到测试超时。
// ─────────────────────────────────────────────────────────────────────────────

func TestSubscribeClosesOnCancel(t *testing.T) {
	f := RandomWalk{
		Symbol: "BTCUSDT",
		Start:  100,
		Step:   time.Hour, // 永远等不到第一根
		Vol:    0.006,
		Rnd:    rand.New(rand.NewSource(1)),
	}

	ctx, cancel := context.WithCancel(context.Background())
	out := f.Subscribe(ctx, 1000)
	cancel() // 立刻取消：协程应退出并关闭 channel

	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("取消后不该再收到行情")
		}
		// ok == false 说明 channel 已被生产者关闭 —— 正是我们要的行为
	case <-time.After(2 * time.Second):
		t.Fatal("ctx 取消后输出 channel 没有被 close（生产者 close 契约被破坏）")
	}
}

// 产出的价格必须为正、Symbol 原样透传、Ts 非零，且 PriceStr 与 Price 必须一致。
//
// PriceStr 是**下单时**用来转 decimal 的原始字符串，它和 Price 必须是同一个价格 ——
// 两者不一致就意味着「策略看到的价」和「下单用的价」是两个数，那是最难查的一类 bug。
func TestRandomWalkQuotesSane(t *testing.T) {
	f := RandomWalk{
		Symbol: "ETHUSDT",
		Start:  3000,
		Step:   time.Millisecond,
		Vol:    0.05, // 用较大波动去压价格，确认不会翻负
		Rnd:    rand.New(rand.NewSource(3)),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var n int
	for q := range f.Subscribe(ctx, 30) {
		n++
		if q.Price <= 0 {
			t.Errorf("第 %d 笔价格 = %v, 必须为正", n, q.Price)
		}
		if q.Symbol != "ETHUSDT" {
			t.Errorf("第 %d 笔 Symbol = %q, want ETHUSDT", n, q.Symbol)
		}
		if q.Ts.IsZero() {
			t.Errorf("第 %d 笔 Ts 为零值", n)
		}
		if q.PriceStr == "" {
			t.Fatalf("第 %d 笔 PriceStr 为空 —— 下游下单没东西可转 decimal", n)
		}
		back, err := strconv.ParseFloat(q.PriceStr, 64)
		if err != nil {
			t.Fatalf("第 %d 笔 PriceStr=%q 不是数字: %v", n, q.PriceStr, err)
		}
		if back != q.Price {
			t.Errorf("第 %d 笔 PriceStr=%q 解析回 %v, 与 Price=%v 不一致", n, q.PriceStr, back, q.Price)
		}
	}
	if n != 30 {
		t.Errorf("收到 %d 笔行情, want 30", n)
	}
}

// Rnd 为 nil 时必须兜底，不能 panic。
//
// 裸着解引用 *rand.Rand 会把整条管线带走（而且是在 goroutine 里 panic，堆栈很难看）。
// 兜底种子是固定的，所以「忘了设 Rnd」时依然可复现。
func TestRandomWalkNilRndDoesNotPanic(t *testing.T) {
	f := RandomWalk{Symbol: "BTCUSDT", Start: 100, Step: time.Millisecond} // Rnd 故意留 nil

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var got []float64
	for q := range f.Subscribe(ctx, 5) {
		got = append(got, q.Price)
	}
	if len(got) != 5 {
		t.Fatalf("收到 %d 笔行情, want 5（Rnd 为 nil 也要能正常产出）", len(got))
	}
	for i, p := range got {
		if p <= 0 {
			t.Errorf("第 %d 笔价格 = %v, 必须为正", i+1, p)
		}
	}

	// 兜底种子固定 → 两次运行结果一致
	f2 := RandomWalk{Symbol: "BTCUSDT", Start: 100, Step: time.Millisecond}
	var got2 []float64
	for q := range f2.Subscribe(ctx, 5) {
		got2 = append(got2, q.Price)
	}
	if !slices.Equal(got, got2) {
		t.Errorf("Rnd 为 nil 的兜底种子应固定（可复现）：\n  a=%v\n  b=%v", got, got2)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// parseWSTicker：WebSocket 推送的纯解析逻辑（回执/pong 等非行情消息必须返回 false）。
// 表驱动，不依赖网络。
// ─────────────────────────────────────────────────────────────────────────────

func TestParseWSTicker(t *testing.T) {
	tests := []struct {
		name         string
		raw          string
		fallback     string // 调用方传的 symbol，仅在推送里没有 instId 时兜底
		wantOK       bool
		wantSymbol   string
		wantPrice    float64
		wantPriceStr string // 原始价格字符串必须原样带出来（下游靠它转 decimal 下单）
	}{
		{
			name:         "snapshot 正常行情",
			raw:          `{"action":"snapshot","arg":{"instId":"BTCUSDT"},"data":[{"instId":"BTCUSDT","lastPr":"84800.63","ts":"1791081880964"}]}`,
			wantOK:       true,
			wantSymbol:   "BTCUSDT",
			wantPrice:    84800.63,
			wantPriceStr: "84800.63",
		},
		{
			name:         "update 增量行情同样被接受",
			raw:          `{"action":"update","arg":{"instId":"ETHUSDT"},"data":[{"instId":"ETHUSDT","lastPr":"3000.5","ts":"1791081880964"}]}`,
			wantOK:       true,
			wantSymbol:   "ETHUSDT",
			wantPrice:    3000.5,
			wantPriceStr: "3000.5",
		},
		{
			name:   "订阅回执不是行情",
			raw:    `{"event":"subscribe","arg":{"instType":"SPOT","channel":"ticker","instId":"BTCUSDT"}}`,
			wantOK: false,
		},
		{
			name:   "pong 不是行情",
			raw:    `{"event":"pong"}`,
			wantOK: false,
		},
		{
			name:   "非法 JSON",
			raw:    `{not json`,
			wantOK: false,
		},
		{
			name:   "data 为空",
			raw:    `{"action":"snapshot","arg":{"instId":"BTCUSDT"},"data":[]}`,
			wantOK: false,
		},
		{
			name:   "lastPr 为空",
			raw:    `{"action":"snapshot","arg":{"instId":"BTCUSDT"},"data":[{"instId":"BTCUSDT","lastPr":""}]}`,
			wantOK: false,
		},
		{
			name:   "lastPr 不是数字",
			raw:    `{"action":"snapshot","arg":{"instId":"BTCUSDT"},"data":[{"instId":"BTCUSDT","lastPr":"abc"}]}`,
			wantOK: false,
		},
		{
			// data 里没给 instId，回退到 arg.instId
			name:         "symbol 回退到 arg.instId",
			raw:          `{"action":"snapshot","arg":{"instId":"SOLUSDT"},"data":[{"lastPr":"150.25"}]}`,
			wantOK:       true,
			wantSymbol:   "SOLUSDT",
			wantPrice:    150.25,
			wantPriceStr: "150.25",
		},
		{
			// 推送里完全没有 symbol，用调用方传入的兜底
			name:         "symbol 回退到调用方传参",
			raw:          `{"action":"snapshot","data":[{"lastPr":"150.25"}]}`,
			fallback:     "DOGEUSDT",
			wantOK:       true,
			wantSymbol:   "DOGEUSDT",
			wantPrice:    150.25,
			wantPriceStr: "150.25",
		},
		{
			// 交易所给的高精度字符串必须原样保留，不能被 float64 往返“洗掉”尾数
			name:         "高精度价格字符串原样保留",
			raw:          `{"action":"snapshot","arg":{"instId":"BTCUSDT"},"data":[{"lastPr":"84800.63000000001"}]}`,
			wantOK:       true,
			wantSymbol:   "BTCUSDT",
			wantPrice:    84800.63000000001,
			wantPriceStr: "84800.63000000001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			q, ok := parseWSTicker([]byte(tt.raw), tt.fallback)
			if ok != tt.wantOK {
				t.Fatalf("parseWSTicker ok = %t, want %t（quote=%+v）", ok, tt.wantOK, q)
			}
			if !tt.wantOK {
				return
			}
			if q.Symbol != tt.wantSymbol {
				t.Errorf("Symbol = %q, want %q", q.Symbol, tt.wantSymbol)
			}
			if q.Price != tt.wantPrice {
				t.Errorf("Price = %v, want %v", q.Price, tt.wantPrice)
			}
			if q.PriceStr != tt.wantPriceStr {
				t.Errorf("PriceStr = %q, want %q（原始字符串必须原样带出）", q.PriceStr, tt.wantPriceStr)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// FetchTicker：用 httptest 起本地假交易所，钉住错误分类。
//
// README 里强调的三点：
//   1. 价格在响应里是 string，解析成 float64 只用于算信号；
//   2. 「网络层失败」（可重试）与「HTTP 状态 / 业务 code 错误」（重试没意义）必须分开；
//   3. HTTP 200 但 code != "00000" 也算失败。
// 这些都是金额/价格从 float64 换 decimal 时最容易被改坏的地方。
// ─────────────────────────────────────────────────────────────────────────────

func TestFetchTicker(t *testing.T) {
	tests := []struct {
		name         string
		status       int
		body         string
		wantErr      string // 错误信息里必须包含的子串
		wantSymbol   string
		wantPrice    float64
		wantPriceStr string // 原始 lastPr 字符串必须原样透传（下单转 decimal 用）
	}{
		{
			name:         "正常返回",
			status:       http.StatusOK,
			body:         `{"code":"00000","msg":"success","data":[{"symbol":"BTCUSDT","lastPr":"84800.63","ts":"1791081880964"}]}`,
			wantSymbol:   "BTCUSDT",
			wantPrice:    84800.63,
			wantPriceStr: "84800.63",
		},
		{
			name:    "HTTP 状态非 200",
			status:  http.StatusInternalServerError,
			body:    `oops`,
			wantErr: "HTTP 状态异常",
		},
		{
			name:    "业务 code 非 00000",
			status:  http.StatusOK,
			body:    `{"code":"40000","msg":"param error","data":[]}`,
			wantErr: "code=40000",
		},
		{
			name:    "data 为空",
			status:  http.StatusOK,
			body:    `{"code":"00000","msg":"success","data":[]}`,
			wantErr: "无数据",
		},
		{
			name:    "lastPr 不是数字",
			status:  http.StatusOK,
			body:    `{"code":"00000","msg":"success","data":[{"symbol":"BTCUSDT","lastPr":"N/A","ts":"1"}]}`,
			wantErr: "不是数字",
		},
		{
			name:    "JSON 非法",
			status:  http.StatusOK,
			body:    `{not json`,
			wantErr: "JSON 解析失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Query().Get("symbol"); got != "BTCUSDT" {
					t.Errorf("请求 symbol 参数 = %q, want BTCUSDT", got)
				}
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			b := &Bitget{Host: srv.URL, Symbol: "BTCUSDT", Timeout: time.Second}
			q, err := b.FetchTicker(context.Background())

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("期望报错（含 %q），实际成功，quote=%+v", tt.wantErr, q)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("错误 = %q, 期望包含 %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("意外报错: %v", err)
			}
			if q.Symbol != tt.wantSymbol {
				t.Errorf("Symbol = %q, want %q", q.Symbol, tt.wantSymbol)
			}
			if q.Price != tt.wantPrice {
				t.Errorf("Price = %v, want %v", q.Price, tt.wantPrice)
			}
			if q.PriceStr != tt.wantPriceStr {
				t.Errorf("PriceStr = %q, want %q（原始 lastPr 必须原样透传）", q.PriceStr, tt.wantPriceStr)
			}
		})
	}
}

// 连不上（服务已关）属于「网络层失败」，与业务错误分开 —— 上层据此决定能否重试。
func TestFetchTickerNetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close() // 立刻关掉，制造连接失败

	b := &Bitget{Host: url, Symbol: "BTCUSDT", Timeout: 500 * time.Millisecond}
	_, err := b.FetchTicker(context.Background())
	if err == nil {
		t.Fatal("服务已关闭时应报网络错误")
	}
	if !strings.Contains(err.Error(), "网络层失败") {
		t.Errorf("错误 = %q, 期望被归类为「网络层失败」（可重试）", err)
	}
}
