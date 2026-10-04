package market

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// tickerDTO 对应 Bitget GET /api/v2/spot/market/tickers 的单个元素。
// 注意：**价格与数量全是 string**，交易所这么做就是为了避开浮点误差（对照 L3-02）。
type tickerDTO struct {
	Symbol     string `json:"symbol"`
	LastPr     string `json:"lastPr"`
	BidPr      string `json:"bidPr"`
	AskPr      string `json:"askPr"`
	BaseVolume string `json:"baseVolume"`
	Ts         string `json:"ts"`
	Change24h  string `json:"change24h"`
}

// Bitget 是真实行情源：轮询 REST tickers 接口，按 Every 间隔推送 Quote。
// 它和 RandomWalk 满足同一个 Feed 接口，所以 main 里换源不需要改任何下游代码。
type Bitget struct {
	Host   string        // 默认 https://api.bitget.com
	Symbol string        // 例如 BTCUSDT
	Every  time.Duration // 轮询间隔
	// Timeout 是**单次请求**的超时；没有它，一次网络抖动就能把整条管线挂死（L2-08/L3-03）
	Timeout time.Duration
	// OnError 让上层能把失败打出来/计入指标，而不是静默丢掉 tick
	OnError func(error)

	client *http.Client
}

func (b *Bitget) normalize() {
	if b.Host == "" {
		b.Host = "https://api.bitget.com"
	}
	if b.Symbol == "" {
		b.Symbol = "BTCUSDT"
	}
	if b.Every <= 0 {
		b.Every = time.Second
	}
	if b.Timeout <= 0 {
		b.Timeout = 800 * time.Millisecond
	}
	if b.client == nil {
		b.client = &http.Client{Timeout: b.Timeout}
	}
}

// FetchTicker 拉一次行情，返回解析后的 Quote。这里把「传输失败」与「业务失败」分开处理。
func (b *Bitget) FetchTicker(ctx context.Context) (Quote, error) {
	b.normalize()

	url := b.Host + "/api/v2/spot/market/tickers?symbol=" + b.Symbol
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Quote{}, fmt.Errorf("构造请求失败: %w", err)
	}
	req.Header.Set("User-Agent", "go-study/trading-bot")

	resp, err := b.client.Do(req)
	if err != nil {
		return Quote{}, fmt.Errorf("网络层失败: %w", err) // 连不上 / 超时：通常可重试
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 限长，防止异常大响应打爆内存（L3-08 服务端同款）
	if err != nil {
		return Quote{}, fmt.Errorf("读取响应失败: %w", err)
	}
	if resp.StatusCode != http.StatusOK { // HTTP 状态先判：err 为 nil 不代表成功
		return Quote{}, fmt.Errorf("HTTP 状态异常 %d: %s", resp.StatusCode, firstLine(string(body)))
	}

	var envelope struct {
		Code string      `json:"code"`
		Msg  string      `json:"msg"`
		Data []tickerDTO `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return Quote{}, fmt.Errorf("JSON 解析失败: %w", err)
	}
	if envelope.Code != "00000" { // 业务码：限频、参数错都走这里，重试策略与网络错误不同
		return Quote{}, fmt.Errorf("交易所返回错误 code=%s msg=%s", envelope.Code, envelope.Msg)
	}
	if len(envelope.Data) == 0 {
		return Quote{}, fmt.Errorf("交易对 %s 无数据", b.Symbol)
	}

	t := envelope.Data[0]
	price, err := strconv.ParseFloat(t.LastPr, 64)
	if err != nil {
		return Quote{}, fmt.Errorf("lastPr %q 不是数字: %w", t.LastPr, err)
	}
	ts, err := strconv.ParseInt(t.Ts, 10, 64)
	if err != nil {
		ts = time.Now().UnixMilli()
	}
	return Quote{Symbol: t.Symbol, Price: price, Ts: time.UnixMilli(ts)}, nil
}

// Subscribe 轮询 n 次并把成功拿到的 tick 推给下游；失败不中断，只上报。
func (b *Bitget) Subscribe(ctx context.Context, n int) <-chan Quote {
	b.normalize()
	out := make(chan Quote)

	go func() {
		defer close(out) // 只有生产者能关通道（L2-05）
		interval := time.NewTicker(b.Every)
		defer interval.Stop() // 不 Stop 会泄漏 timer（L2-10）

		for i := 0; i < n; i++ {
			select {
			case <-ctx.Done(): // 取消即退出，不留协程（L2-11）
				return
			case <-interval.C:
			}

			q, err := b.FetchTicker(ctx)
			if err != nil {
				if b.OnError != nil {
					b.OnError(err)
				}
				continue // 丢一个 tick 优于挂住整条管线；真实项目还要退避重试
			}
			select {
			case out <- q:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// CloseIdleConnections 释放 HTTP Keep-Alive 的空闲连接。
// 不调用它，一轮跑完事后 `runtime.NumGoroutine()` 仍会比启动前高——那些连接维持 goroutine 还在。
func (b *Bitget) CloseIdleConnections() {
	if b.client != nil {
		b.client.CloseIdleConnections()
	}
}

func firstLine(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			return s[:i]
		}
	}
	if len(s) > 120 {
		return s[:120]
	}
	return s
}
