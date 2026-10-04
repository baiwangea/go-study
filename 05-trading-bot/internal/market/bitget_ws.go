package market

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

// wsTicker 是 ticker 频道推送的单条数据（只声明用得到的字段）。
type wsTicker struct {
	InstID string `json:"instId"`
	LastPr string `json:"lastPr"`
	BidPr  string `json:"bidPr"`
	AskPr  string `json:"askPr"`
	Ts     string `json:"ts"`
}

// BitgetWS 是推送式行情源：一条长连接收 snapshot/update，带心跳与断线重连。
// 它与 RandomWalk、Bitget(REST) 实现同一个 Feed 接口，下游一行都不用改。
//
// 协议是实测出来的（不是照抄文档）：
//
//	订阅 {"op":"subscribe","args":[{"instType":"SPOT","channel":"ticker","instId":"BTCUSDT"}]}
//	回执 {"event":"subscribe",...} → 推送 {"action":"snapshot|update","arg":{...},"data":[{...}]}
//	instType 写 sp 会被拒（30016），顶层字段写成 action 也会被拒（30003 INVALID op:null）。
type BitgetWS struct {
	URL     string      // 默认 wss://ws.bitget.com/v2/ws/public
	Symbol  string      // instId，例如 BTCUSDT
	OnError func(error) // 断线/解析失败上报，不中断管线

	reconnects int
}

const wsDefaultURL = "wss://ws.bitget.com/v2/ws/public"

func (w *BitgetWS) normalize() {
	if w.URL == "" {
		w.URL = wsDefaultURL
	}
	if w.Symbol == "" {
		w.Symbol = "BTCUSDT"
	}
}

// Reconnects 暴露重连次数，用来观察这条链路的健康度。
func (w *BitgetWS) Reconnects() int { return w.reconnects }

// Subscribe 持续接收推送，直到收满 n 条或 ctx 取消。
// 三件必须做的事：心跳（否则服务端会踢掉你）、读超时（否则协程永远卡在读上）、重连退避。
func (w *BitgetWS) Subscribe(ctx context.Context, n int) <-chan Quote {
	w.normalize()
	out := make(chan Quote)

	go func() {
		defer close(out) // 只有生产者能关通道（L2-05）

		recv, backoff := 0, 500*time.Millisecond
		for recv < n {
			if ctx.Err() != nil {
				return
			}
			got, err := w.consume(ctx, out, n-recv)
			recv += got
			if err == nil || ctx.Err() != nil {
				return // 收满了，或者被取消了
			}
			if w.OnError != nil {
				w.OnError(fmt.Errorf("WS 断线，%v 后重连: %w", backoff, err))
			}
			w.reconnects++

			select {
			case <-time.After(backoff):
				if backoff < 8*time.Second {
					backoff *= 2 // 指数退避：别打死交易所，也别把自己饿死
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// consume 建立一条连接、订阅、把推送转发到 out；返回成功投递条数与失败原因。
func (w *BitgetWS) consume(ctx context.Context, out chan<- Quote, want int) (int, error) {
	dialer := websocket.Dialer{HandshakeTimeout: 5 * time.Second}

	conn, _, err := dialer.DialContext(ctx, w.URL, nil)
	if err != nil {
		return 0, fmt.Errorf("握手失败: %w", err)
	}
	defer conn.Close()

	sub := fmt.Sprintf(`{"op":"subscribe","args":[{"instType":"SPOT","channel":"ticker","instId":"%s"}]}`, w.Symbol)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(sub)); err != nil {
		return 0, fmt.Errorf("订阅失败: %w", err)
	}

	type result struct {
		msg []byte
		err error
	}
	readCh := make(chan result, 1)
	go func() { // 读协程单独存在，主循环才能同时处理心跳与取消
		defer close(readCh)
		for {
			// 心跳 20s 一次 → 90s 读不到任何东西就算死线
			_ = conn.SetReadDeadline(time.Now().Add(90 * time.Second))
			_, msg, err := conn.ReadMessage()
			select {
			case readCh <- result{msg, err}:
			case <-ctx.Done():
				return
			}
			if err != nil {
				return
			}
		}
	}()

	ping := time.NewTicker(20 * time.Second)
	defer ping.Stop()

	sent := 0
	for sent < want {
		select {
		case <-ctx.Done():
			return sent, ctx.Err()

		case <-ping.C: // 服务端要求 30s 内收到 ping
			_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := conn.WriteMessage(websocket.TextMessage, []byte("ping")); err != nil {
				return sent, fmt.Errorf("心跳发送失败: %w", err)
			}

		case r, ok := <-readCh:
			if !ok {
				return sent, fmt.Errorf("读协程已结束")
			}
			if r.err != nil {
				return sent, fmt.Errorf("读取失败: %w", r.err)
			}
			q, matched := parseWSTicker(r.msg, w.Symbol)
			if !matched {
				continue // 订阅回执、pong、无关频道
			}
			select {
			case out <- q:
				sent++
			case <-ctx.Done():
				return sent, ctx.Err()
			}
		}
	}
	return sent, nil
}

// parseWSTicker 解析推送；非行情消息（回执/pong）返回 false。
func parseWSTicker(raw []byte, symbol string) (Quote, bool) {
	var env struct {
		Action string `json:"action"`
		Arg    struct {
			InstID string `json:"instId"`
		} `json:"arg"`
		Data []wsTicker `json:"data"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return Quote{}, false
	}
	if env.Action != "snapshot" && env.Action != "update" {
		return Quote{}, false
	}
	if len(env.Data) == 0 || env.Data[0].LastPr == "" {
		return Quote{}, false
	}
	price, err := strconv.ParseFloat(env.Data[0].LastPr, 64)
	if err != nil {
		return Quote{}, false
	}

	sym := env.Data[0].InstID
	if sym == "" {
		sym = env.Arg.InstID
	}
	if sym == "" {
		sym = symbol
	}
	ts, terr := strconv.ParseInt(env.Data[0].Ts, 10, 64)
	if terr != nil {
		ts = time.Now().UnixMilli()
	}
	return Quote{Symbol: sym, Price: price, Ts: time.UnixMilli(ts)}, true
}
