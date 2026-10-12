//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 07-web3-trading && go run 10_ws_reconnect.go
// 依赖：github.com/gorilla/websocket（本关需要，见 07-web3-trading/go.mod）
// 说明：假交易所在**本进程内**用 httptest 起，全程离线可跑，不需要网络。
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

/*
═══════════════════════════════════════════════════════════════════════
W-10：WebSocket 断线重连与心跳
═══════════════════════════════════════════════════════════════════════

【核心概念】
    WebSocket 是「一条长连接」，不是「一次请求」。交易所的行情和私有订单推送
    都走它。长连接的生命周期里，**断线是常态而不是异常**：
      · 交易所每 24 小时强制断开一次（Binance 就是 24h）；
      · 网络抖动、NAT 超时、中间设备悄悄掐掉空闲连接；
      · 服务端发布/重启、限频踢人。
    所以「连上就能一直收」的想法在生产里活不过一天。你必须实现三件事：

    ① 心跳（heartbeat）：客户端定时发 ping，服务端回 pong；
       连续 N 次收不到 pong 就判定连接已死，主动重连。
       —— 没有心跳，半开的连接（对端没了但你这边不知道）会让你安静地收不到数据。

    ② 指数退避（exponential backoff）：重连间隔 100ms → 200ms → 400ms → … 封顶，
       并加随机抖动（jitter）。
       —— 不加退避：交易所宕机时你的机器人会每秒重连几百次，直接被封 IP；
       —— 不加抖动：所有客户端会在同一毫秒一起重连，把服务端打成「惊群」。

    ③ 断线补数（gap fill）：重连后先用 REST 把「断开期间漏掉的 tick」拉回来，
       再切回 WS。
       —— 这是最容易被忽略、后果最严重的一步。行情在你断线的 800ms 里继续走，
          不补数你的 K 线就永久缺一段，指标算错、信号误判、回测对不上实盘。

【你会看到】
    ✓ 本进程内用 httptest 起一个假交易所：WS 推行情 + REST 补数接口
    ✓ 服务端每推送 4 条就主动断开（模拟交易所踢人），客户端自动爬起来
    ✓ 退避时长实测：100ms → 200ms → 400ms 这个量级，带 ±20% 抖动
    ✓ 每次重连后先 REST 补数（真的补回若干条缺口），再切回 WS 继续读
    ✓ 客户端 ping / 服务端 pong，pong 用来续读超时
    ✓ ctx 取消后 goroutine 数回落到基线（不泄漏、不卡死）

【思考题】（动手改代码验证，别空想）
    1. 把 nextBackoff 里的抖动去掉（固定 100/200/400ms），再跑一次。
       输出变了吗？如果全世界的机器人都这么写，交易所宕机重启的瞬间会发生什么？
    2. 把 pongWait 从 1s 改成 10ms，会发生什么？反过来改成 10s 呢？
       心跳间隔和超时时间之间应该是什么关系？（提示：ping 至少要比超时快几倍）
    3. 注释掉 session() 里那段 fillGap 调用，再跑。输出的「序号连续性校验」
       会变成什么样？如果这是实盘，你会在什么指标上先看出问题？
    4. 补数为什么要**先补再读 WS**？把顺序反过来（先读 WS 再补数）会怎样？
       （提示：WS 一读，lastSeq 就被推高了，缺口就再也补不回来）
    5. 如果补数期间漏掉的 tick 太多（比如断线 10 分钟），REST 接口该怎么做？
       （提示：分页/时间窗拉取 + 限制最大补数条数，超限就重建整个本地行情）
    6. 现在 killAfter 是 4，服务端断开得很规律。真实交易所的断开没有规律，
       你会怎么测这条重连路径？（提示：注入故障，比如随机概率断开/半开连接）

【通关标准】
    能默写出「心跳 + 指数退避 + 断线补数 + ctx 优雅退出」的重连骨架，
    并说清「为什么断线补数不能省」和「为什么退避必须加抖动」。

【前置】L2-13（context）· L2-18（pipeline 与泄漏）· L3-11（重试与幂等）
【标签】WebSocket · 心跳 · 指数退避 · 抖动 · 断线补数 · 优雅退出
═══════════════════════════════════════════════════════════════════════
*/

// ─────────────────────────────────────────────────────────────────────────────
// 可调参数：都是为教学选的值。真实项目里这几个数要贴着交易所的规则来定
// （Binance 要求 3 分钟内至少发一次 ping；24h 强制断开）。
// ─────────────────────────────────────────────────────────────────────────────
const (
	tickInterval = 25 * time.Millisecond  // 假交易所的行情推送间隔
	killAfter    = 4                      // 服务端每个连接推送几条后主动断开（模拟踢人）
	targetLive   = 13                     // 客户端目标：WS 收到 13 条 live tick 就收工（13 = 3×4 + 1，正好逼出第 4 次连接，把 100/200/400ms 三段退避都走一遍）
	pingPeriod   = 40 * time.Millisecond  // 客户端心跳间隔
	pongWait     = 1 * time.Second        // 多久收不到 pong 就判定连接已死
	backoffBase  = 100 * time.Millisecond // 退避基数：100 → 200 → 400 → …
	backoffMax   = 800 * time.Millisecond // 退避上限：再涨就没必要了，不如直接告警
)

// Tick 是一条行情。
//
// ⚠ Price 用 float64 只是为了演示 WS/JSON 编解码方便 ——
// 真正参与下单计算的金额/数量绝不能用 float64，见 W-01~W-03 的 decimal 方案。
type Tick struct {
	Seq   int64   `json:"seq"`
	Price float64 `json:"price"`
}

// ─────────────────────────────────────────────────────────────────────────────
// 一、假交易所：WS 推送 + REST 补数
// ─────────────────────────────────────────────────────────────────────────────

// hub 是假交易所的核心：它自己「产生行情」，并把行情广播给所有连着的客户端。
//
// 关键设计：**行情不依赖客户端是否存在**。没有客户端连着时，行情照样生成、
// 照样写进 history。这正是「断线会丢 tick」的根因 ——
// 断开期间行情继续走，只是没人收；所以重连后必须靠 REST 把这段补回来。
type hub struct {
	mu      sync.Mutex
	history []Tick               // 最近若干条行情，供 REST 补数
	clients map[*wsConn]struct{} // 当前连着的客户端
	seq     int64                // 行情序号，全局单调递增
	price   float64              // 当前价格，随机游走
	rng     *rand.Rand

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// wsConn 是一条已升级的连接 + 它的发送队列。
// 发送走 channel 而不是直接 Write，是为了让「广播」和「写连接」解耦：
// 广播方永远不会因为某个客户端网络慢而被阻塞。
type wsConn struct {
	conn *websocket.Conn
	send chan Tick
	done chan struct{} // 由读泵在连接结束时 close，通知写循环收工
}

// newHub 建一个假交易所。行情价格从 63000（BTC 的整数关口）开始随机游走。
func newHub(parent context.Context) *hub {
	ctx, cancel := context.WithCancel(parent)
	return &hub{
		clients: map[*wsConn]struct{}{},
		seq:     0,
		price:   63000,
		rng:     rand.New(rand.NewPCG(20241012, 1001)), // 固定种子：让教学输出可复现
		ctx:     ctx,
		cancel:  cancel,
	}
}

// run 是行情生产者：每 tickInterval 生成一条行情，写历史 + 广播给所有客户端。
func (h *hub) run() {
	defer h.wg.Done()
	t := time.NewTicker(tickInterval)
	defer t.Stop()
	for {
		select {
		case <-h.ctx.Done():
			return
		case <-t.C:
			tick := h.nextTick()
			for _, c := range h.snapshotClients() {
				select {
				case c.send <- tick:
				default:
					// 客户端太慢（队列满）：真实系统这里必须做决策 ——
					// 丢弃这条（丢行情）还是踢掉这个连接（让它重连补数）。
					// 我们这里静默丢弃，因为重连补数本来就能兜住。
				}
			}
		}
	}
}

// nextTick 生成下一条行情并写入历史，返回它。
func (h *hub) nextTick() Tick {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.seq++
	h.price *= 1 + (h.rng.Float64()-0.5)*0.004 // ±0.2% 随机游走
	tick := Tick{Seq: h.seq, Price: math.Round(h.price*100) / 100}
	h.history = append(h.history, tick)
	if len(h.history) > 4000 { // 只留最近 4000 条，真实服务是保留 N 天的 K 线
		h.history = h.history[len(h.history)-4000:]
	}
	return tick
}

// snapshotClients 复制一份当前客户端列表，让广播时不持有锁。
func (h *hub) snapshotClients() []*wsConn {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]*wsConn, 0, len(h.clients))
	for c := range h.clients {
		out = append(out, c)
	}
	return out
}

// since 返回序号 > seq 的全部历史行情 —— REST 补数接口的数据源。
func (h *hub) since(seq int64) []Tick {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := []Tick{}
	for _, t := range h.history {
		if t.Seq > seq {
			out = append(out, t)
		}
	}
	return out
}

// clientCount 返回当前连接数，只用于打印。
func (h *hub) clientCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.clients)
}

// removeClient 把连接从广播列表里摘掉。注意这里**不 close(c.done)**：
// done 由读泵负责 close，两处都关会 panic（close 已关闭的 channel）。
func (h *hub) removeClient(c *wsConn) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
}

// upgrader 把普通 HTTP 请求升级成 WebSocket。
// CheckOrigin 返回 true 只是教学方便：真实服务必须校验 Origin，否则会被跨站滥用。
var upgrader = websocket.Upgrader{
	CheckOrigin: func(*http.Request) bool { return true },
}

// serveWS 处理一条 WS 连接：注册 → 推送 → 推满 killAfter 条后主动断开。
func (h *hub) serveWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("    [服务端] 升级失败：%v\n", err)
		return
	}
	c := &wsConn{conn: conn, send: make(chan Tick, 64), done: make(chan struct{})}

	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
	fmt.Printf("    [服务端] 新连接接入（当前连接数=%d）\n", h.clientCount())

	// 读泵：必须有 goroutine 在读，gorilla 才会处理 ping/pong/close 这些控制帧。
	// 客户端的 ping 就是靠它触发默认 PingHandler，自动回一个 pong。
	go func() {
		defer close(c.done)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	// 写循环：从 send 队列取行情写出去，写满 killAfter 条就主动断开 ——
	// 模拟交易所踢人（真实里是 24h 强制断开、或限频踢人）。
	idle := time.NewTimer(3 * time.Second)
	defer idle.Stop()
	sent := 0
	for {
		select {
		case <-h.ctx.Done():
			conn.Close()
			h.removeClient(c)
			return
		case <-c.done:
			h.removeClient(c)
			return
		case tick := <-c.send:
			_ = conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
			if err := conn.WriteJSON(tick); err != nil {
				conn.Close()
				h.removeClient(c)
				return
			}
			sent++
			if sent >= killAfter {
				fmt.Printf("    [服务端] 已推送 %d 条，现在**主动断开**这个连接（模拟交易所踢人）\n", sent)
				conn.Close()
				h.removeClient(c)
				return
			}
		case <-idle.C:
			conn.Close()
			h.removeClient(c)
			return
		}
	}
}

// serveHistory 是 REST 补数接口：GET /history?since=N → 返回序号 > N 的行情。
//
// 真实对应物：Binance 的 GET /api/v3/klines?symbol=BTCUSDT&startTime=...，
// 或 /api/v3/aggTrades?fromId=N。
func (h *hub) serveHistory(w http.ResponseWriter, r *http.Request) {
	since, _ := strconv.ParseInt(r.URL.Query().Get("since"), 10, 64)
	ticks := h.since(since)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ticks)
}

// ─────────────────────────────────────────────────────────────────────────────
// 二、行情订阅客户端：心跳 + 指数退避 + 断线补数
// ─────────────────────────────────────────────────────────────────────────────

// feedClient 是「行情订阅客户端」。它对外只做一件事：
// 把行情可靠地喂给上层 —— 断了也要自己爬起来，还要把断线期间的缺口补上。
type feedClient struct {
	wsURL   string
	histURL string
	rng     *rand.Rand // 固定种子的抖动源，让教学输出可复现

	lastSeq   int64  // 已收到的最大序号：既是补数起点，也是去重依据
	liveTicks []Tick // 通过 WS 收到的
	gapTicks  []Tick // 通过 REST 补回来的

	sessions    int // WS 连接次数
	reconnects  int // 重连次数
	sessionLive int // 本次连接收到的 live tick 数（打印用）
	backoffs    []time.Duration
	pongs       int // 收到多少次 pong
}

func newFeedClient(wsURL, histURL string) *feedClient {
	return &feedClient{
		wsURL:   wsURL,
		histURL: histURL,
		rng:     rand.New(rand.NewPCG(7, 7)),
	}
}

// run 是重连主循环：连 → 断了就退避 → 再连 → 直到收够或 ctx 取消。
func (c *feedClient) run(ctx context.Context) error {
	for {
		done, err := c.session(ctx)
		if err != nil {
			fmt.Printf("  [客户端] 第 %d 次连接结束：%v（本次收到 %d 条 live tick）\n",
				c.sessions, err, c.sessionLive)
		}
		if done {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// 指数退避 + 抖动
		c.reconnects++
		d := c.nextBackoff(c.reconnects)
		c.backoffs = append(c.backoffs, d)
		fmt.Printf("  [客户端] 第 %d 次重连：退避 %v（基数 %v 指数增长、上限 %v，含 ±20%% 抖动）后重试...\n",
			c.reconnects, d.Round(time.Millisecond), backoffBase, backoffMax)
		select {
		case <-time.After(d):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// nextBackoff 算第 attempt 次重连该等多久：
//
//	基数 100ms 按 2 的幂增长 → 100 / 200 / 400 / 800（封顶），
//	再乘一个 [0.8, 1.2) 的随机因子（抖动）。
//
// 为什么要抖动？如果一万个客户端都用固定的 100/200/400ms，
// 交易所重启的瞬间它们会在同一毫秒一起重连 —— 刚起来的服务立刻被打垮，
// 这叫「惊群（thundering herd）」。抖动把重连时刻打散开。
func (c *feedClient) nextBackoff(attempt int) time.Duration {
	base := backoffBase << (attempt - 1) // 100ms << 0/1/2/3…
	if base > backoffMax || base <= 0 {  // <=0 防左移溢出
		base = backoffMax
	}
	jitter := 0.8 + 0.4*c.rng.Float64() // [0.8, 1.2)
	return time.Duration(float64(base) * jitter)
}

// session 建立一次连接，读到断线（或收够）为止。
// 返回 done=true 表示「目标已达成，可以收工了」。
func (c *feedClient) session(ctx context.Context) (done bool, err error) {
	c.sessions++
	got := 0
	defer func() { c.sessionLive = got }() // 无论怎么退出，都记下本次收了几条

	// ① 拨号。必须带超时：没有超时的拨号会在网络黑洞里挂很久。
	dialCtx, cancelDial := context.WithTimeout(ctx, 3*time.Second)
	conn, _, err := websocket.DefaultDialer.DialContext(dialCtx, c.wsURL, nil)
	cancelDial()
	if err != nil {
		return false, fmt.Errorf("拨号失败: %w", err)
	}
	defer conn.Close()
	fmt.Printf("  [客户端] 第 %d 次连接建立：%s\n", c.sessions, c.wsURL)

	// ② 断线补数：重连后**先**用 REST 把断开期间的缺口拉回来，再切回 WS。
	//    顺序绝不能反 —— 一旦开始读 WS，lastSeq 就被推高了，缺口再也补不回来。
	if c.lastSeq > 0 {
		since := c.lastSeq
		gap, gerr := c.fillGap(ctx)
		switch {
		case gerr != nil:
			fmt.Printf("  [客户端] ⚠ REST 补数失败：%v（继续用 WS，但这段行情就永久丢了）\n", gerr)
		case len(gap) == 0:
			fmt.Println("  [客户端] REST 补数：没有缺口（断开时间短于推送间隔）")
		default:
			c.gapTicks = append(c.gapTicks, gap...)
			c.lastSeq = gap[len(gap)-1].Seq
			fmt.Printf("  [客户端] REST 补数：since=%d → 补回 %d 条缺口（seq %d..%d，价格 %g → %g）\n",
				since, len(gap), gap[0].Seq, gap[len(gap)-1].Seq, gap[0].Price, gap[len(gap)-1].Price)
			fmt.Println("             ↑ 少了这一步，断线期间的 tick 就永久丢了：K 线缺段、指标跳空")
		}
	}

	// ③ 心跳：读超时 + pong 续期。
	//    客户端定时发 ping，服务端的默认 PingHandler 自动回 pong；
	//    我们每收到一个 pong 就把读超时往后推。收不到 pong → 读超时 → 报错 → 重连。
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		c.pongs++
		if c.pongs == 1 {
			fmt.Println("  [客户端] 收到第一个 pong：心跳链路正常（客户端 ping / 服务端自动回 pong）")
		}
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	pingDone := make(chan struct{})
	var pingWg sync.WaitGroup
	pingWg.Add(1)
	go func() {
		defer pingWg.Done()
		t := time.NewTicker(pingPeriod)
		defer t.Stop()
		for {
			select {
			case <-pingDone:
				return
			case <-ctx.Done():
				return
			case <-t.C:
				// WriteControl 可以和别的写并发调用（gorilla 明确保证这点），
				// 而且**必须带 deadline**：没有 deadline 的写在对端假死时会永久阻塞。
				if err := conn.WriteControl(websocket.PingMessage, []byte("hb"),
					time.Now().Add(time.Second)); err != nil {
					return
				}
			}
		}
	}()
	// 收尾顺序：先停心跳协程，再关连接。
	// 关连接是必须的 —— 只有它能让下面阻塞的 ReadJSON 立刻返回，否则读循环收不回来。
	defer func() {
		close(pingDone)
		pingWg.Wait()
		conn.Close()
	}()

	// ④ 读行情
	for {
		var t Tick
		if err := conn.ReadJSON(&t); err != nil {
			return false, fmt.Errorf("连接断开: %w", err)
		}
		if t.Seq <= c.lastSeq {
			continue // 去重：补数已经拿到的 tick，WS 可能又推一遍
		}
		c.lastSeq = t.Seq
		c.liveTicks = append(c.liveTicks, t)
		got++
		if len(c.liveTicks) == 1 {
			fmt.Printf("  [客户端] 首条行情：seq=%d price=%g（后续同类行情不再逐条打印）\n", t.Seq, t.Price)
		}
		if len(c.liveTicks) >= targetLive {
			fmt.Printf("  [客户端] 累计 live tick 达到目标 %d → 主动收工\n", targetLive)
			return true, nil
		}
	}
}

// fillGap 用 REST 把「lastSeq 之后」的行情补回来。
// 真实场景里这是 GET /api/v3/klines?symbol=BTCUSDT&startTime=... 之类。
func (c *feedClient) fillGap(ctx context.Context) ([]Tick, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	url := c.histURL + "?since=" + strconv.FormatInt(c.lastSeq, 10)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ticks []Tick
	if err := json.Unmarshal(body, &ticks); err != nil {
		return nil, err
	}
	// 去重：WS 与 REST 的时间窗口必然重叠，同一个 seq 两边都可能拿到。
	out := make([]Tick, 0, len(ticks))
	for _, t := range ticks {
		if t.Seq > c.lastSeq {
			out = append(out, t)
		}
	}
	return out, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// 三、收尾与校验用的小工具
// ─────────────────────────────────────────────────────────────────────────────

// seqCheck 把 WS 收到的和 REST 补回的 tick 合起来，检查序号有没有空洞。
// 这是「到底有没有丢 tick」最直接的验证方式 —— 比肉眼看日志可靠得多。
func seqCheck(c *feedClient) string {
	all := make([]Tick, 0, len(c.liveTicks)+len(c.gapTicks))
	all = append(all, c.liveTicks...)
	all = append(all, c.gapTicks...)
	sort.Slice(all, func(i, j int) bool { return all[i].Seq < all[j].Seq })

	dups, holes := 0, 0
	for i := 1; i < len(all); i++ {
		switch {
		case all[i].Seq == all[i-1].Seq:
			dups++
		case all[i].Seq != all[i-1].Seq+1:
			holes++
		}
	}
	verdict := "无丢包 ✓"
	if holes > 0 {
		verdict = "有丢包 ✗"
	}
	return fmt.Sprintf("合并 %d 条（seq %d..%d），重复 %d 条，空洞 %d 处 → %s",
		len(all), all[0].Seq, all[len(all)-1].Seq, dups, holes, verdict)
}

// waitGoroutines 轮询等 goroutine 数回落到 target 以下，最多等 timeout，
// 返回最终观察到的数量 —— 用它来证明「收尾之后一个协程都没留」。
func waitGoroutines(target int, timeout time.Duration) int {
	deadline := time.Now().Add(timeout)
	for {
		n := runtime.NumGoroutine()
		if n <= target || time.Now().After(deadline) {
			return n
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func main() {
	fmt.Println("\n=== W-10 WebSocket 断线重连与心跳 ===")

	base := runtime.NumGoroutine() // 记录基线，最后用它证明没泄漏
	rootCtx, cancelRoot := context.WithCancel(context.Background())
	defer cancelRoot()

	// ── 示例 1：本进程内起一个「假交易所」────────────────────────────────
	fmt.Println("\n【示例 1】本进程内起一个「假交易所」：WS 推行情 + REST 补数")
	h := newHub(rootCtx)
	h.wg.Add(1)
	go h.run()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", h.serveWS)
	mux.HandleFunc("/history", h.serveHistory)
	srv := httptest.NewServer(mux)

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	histURL := srv.URL + "/history"
	fmt.Printf("  ✓ 假交易所已启动：%s\n", srv.URL)
	fmt.Printf("    · WS 行情   ：%s（每 %v 一条 tick，连满 %d 条就主动断开）\n",
		wsURL, tickInterval, killAfter)
	fmt.Printf("    · REST 补数 ：%s?since=N（返回序号 > N 的历史 tick）\n", histURL)
	fmt.Println("  ↑ 真实对应物：wss://stream.binance.com:9443/ws/btcusdt@trade")
	fmt.Println("    和 https://api.binance.com/api/v3/klines?symbol=BTCUSDT&startTime=...")
	fmt.Println("  ↑ 注意：行情是 hub 自己产生的，跟有没有客户端无关 —— 这正是断线会丢 tick 的原因")

	// ── 示例 2：客户端连、断、退避、补数 ────────────────────────────────
	fmt.Println("\n【示例 2】客户端：连接 → 心跳 → 被踢 → 指数退避重连 → REST 补数 → 继续")
	// 兜底超时：哪怕重连逻辑写错，20 秒后也一定会退出，绝不卡死。
	clientCtx, cancelClient := context.WithTimeout(rootCtx, 20*time.Second)

	c := newFeedClient(wsURL, histURL)
	start := time.Now()
	if err := c.run(clientCtx); err != nil && clientCtx.Err() == nil {
		fmt.Printf("  [客户端] 异常退出：%v\n", err)
	}
	elapsed := time.Since(start)

	// ── 示例 3：运行结果 ────────────────────────────────────────────────
	fmt.Println("\n【示例 3】运行结果")
	parts := make([]string, 0, len(c.backoffs))
	for _, d := range c.backoffs {
		parts = append(parts, d.Round(time.Millisecond).String())
	}
	fmt.Printf("  WS 连接次数      : %d 次（服务端每连满 %d 条就踢一次，所以必然要重连）\n",
		c.sessions, killAfter)
	fmt.Printf("  重连次数         : %d 次\n", c.reconnects)
	fmt.Printf("  实际退避时长     : %s\n", strings.Join(parts, " → "))
	fmt.Printf("  退避规律         : 基数 %v 指数增长，上限 %v，再乘 ±20%% 抖动\n", backoffBase, backoffMax)
	fmt.Printf("  WS 收到的 tick   : %d 条\n", len(c.liveTicks))
	fmt.Printf("  REST 补回的 tick : %d 条 ← 没有它，这些 tick 就永久丢了\n", len(c.gapTicks))
	fmt.Printf("  收到 pong        : %d 次（客户端每 %v 发一次 ping）\n", c.pongs, pingPeriod)
	fmt.Printf("  总耗时           : %v\n", elapsed.Round(time.Millisecond))
	fmt.Printf("  序号连续性校验   : %s\n", seqCheck(c))

	// ── 示例 4：优雅停止，不留 goroutine ────────────────────────────────
	fmt.Println("\n【示例 4】优雅停止：cancel → 关服务 → goroutine 回落到基线")
	cancelClient()
	cancelRoot()
	srv.Close()
	h.wg.Wait()
	settled := waitGoroutines(base, 3*time.Second)
	fmt.Printf("  goroutine：启动前 %d → 收尾后 %d\n", base, settled)
	fmt.Println("  ↑ 每个阻塞点都绑了退出信号：ping 协程靠 close(pingDone) 退出，")
	fmt.Println("    读循环靠 conn.Close() 唤醒，hub 靠 cancel() 收工 —— 一个协程都不留。")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n【关键总结】")
	fmt.Println("  • 长连接里「断线是常态」：心跳 + 指数退避 + 断线补数，三件套缺一不可")
	fmt.Println("  • 心跳：客户端定时 ping，服务端自动回 pong；收不到 pong 就判定连接已死")
	fmt.Println("  • 退避：100 → 200 → 400ms 封顶，**必须加抖动**，否则会惊群打垮刚重启的服务端")
	fmt.Println("  • 补数：重连后先 REST 拉缺口、再切回 WS；顺序反了缺口就补不回来了")
	fmt.Println("  • 去重：WS 与 REST 的时间窗必然重叠，靠 seq 去重，别重复处理同一根 K 线")
	fmt.Println("  • 优雅退出：所有阻塞的收发都绑 ctx / deadline / close(chan)，不留协程")
	fmt.Println("  • 该用：行情订阅、私有订单推送 —— 任何需要「持续在线」的场景")
	fmt.Println("  • 不该用：一次性的 REST 查询（用普通 HTTP 客户端 + 超时重试就够了）")
}
