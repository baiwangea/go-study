/*
═══════════════════════════════════════════════════════════════════════
T-03：测 HTTP handler（net/http/httptest）
═══════════════════════════════════════════════════════════════════════

【核心概念】

	测 HTTP 有两种完全不同的姿势，必须分清：

	    ① 直测 handler：httptest.NewRecorder() + httptest.NewRequest()
	       不起网络、不占端口、毫秒级。**绝大多数用例都该用这个。**

	    ② 起真服务：httptest.NewServer(handler)  →  拿到一个真实的 http://127.0.0.1:port
	       真走 TCP。用来测「客户端代码」、测中间件链路、测超时/重试。

	对照其他语言：
	    PHPUnit : 只能起真服务或依赖 Symfony 的 KernelBrowser，测一次几百毫秒
	    JUnit   : Spring 的 MockMvc 就是「直测 handler」的思路，但要一堆注解和上下文
	    Go      : 标准库自带两种，零依赖。httptest.NewRequest 造请求、
	              httptest.NewRecorder 收响应 —— 两行代码搞定一个完整请求。

	直测 handler 的本质：http.Handler 就是一个
	    ServeHTTP(w http.ResponseWriter, r *http.Request)
	所以只要给它一个「假 w」和一个「假 r」，就能在不碰网络的情况下跑完整个处理逻辑。
	ResponseRecorder 就是那个假 w：它把 WriteHeader / Write 的结果都记下来供你断言。

	⚠ 一个必须记住的细节：**直测 handler 不会经过真实网络，
	所以它测不出「超时」「连接被拒」「TLS」这类问题。**
	这也是为什么需要第二种姿势 —— 两种都要会。

【运行方式】

	cd 06-go-testing && go test -v ./03_httptest/

	    go test -v -run 'TestCreateOrder/非法JSON' ./03_httptest/   # 只跑一条子测试
	    go test -race ./03_httptest/                                # 真起服务，顺便查竞态

【你会看到】

	✓ 直测 handler：一次请求全程不占端口，7 条用例不到 1 毫秒
	✓ 断言四件套：状态码、Content-Type、响应体 JSON、以及 handler 写的 header
	✓ httptest.NewServer 起真服务，用 http.Client 走完一次真实的「下单 → 查单」
	✓ 真服务版本能测出「客户端超时」—— 这是直测 handler 做不到的

【思考题】（动手改代码验证，别空想）

 1. 把 TestCreateOrder 里的 httptest.NewRecorder 换成
    httptest.NewServer(srv) + http.Post(...)，同样的用例要改多少行？
    再想：如果每个用例都起真服务，100 条用例要多少毫秒？
    这就是「为什么绝大多数用例应该直测 handler」。

 2. 给 handleCreate 加一个「读取请求体失败」的分支（比如把
    io.ReadAll 换成 json.NewDecoder(r.Body).Decode 后故意传畸形 JSON），
    用直测 handler 能不能测出来？起真服务能不能？
    想一想：什么类型的 bug **只有真服务才测得出来**？

 3. 现在 handleCreate 的错误响应是手写的 http.Error。
    把它改成统一返回 JSON（{"error": "..."}），
    然后想清楚：这个改动会影响多少条测试？
    如果测试断言的是完整 JSON 字符串，改动成本是多少？
    （提示：断言「结构」而不是「字符串」，是让测试不被无关改动打碎的关键。）

 4. TestClientTimeout 里的超时设成了 50ms。
    把 handleSlow 的 sleep 改成 10ms 再跑一次，测试会怎样失败？
    这种「依赖真实时间的测试」在 CI 上最容易变成 flaky。
    有什么办法能既不依赖真实时间、又测到超时行为？
    （提示：用 context 注入一个可控的时钟，或者把 sleep 时长做成参数。）

【通关标准】

	能说清「直测 handler」和「起真服务」各自的适用场景，
	能默写 httptest.NewRecorder + httptest.NewRequest 的骨架，
	并且知道「断言响应结构而不是完整字符串」。

【前置】L1-08（接口）、T-02（子测试）   【标签】testing · httptest · handler · 超时
═══════════════════════════════════════════════════════════════════════
*/
package orderapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// 被测代码：一个极简的下单 API。
// 用标准库 net/http 而不是 GoFrame，是因为本关的重点是「怎么测」，
// 不该让框架的启动开销和注解干扰注意力。
// ─────────────────────────────────────────────────────────────────────────────

// CreateOrderReq 是下单请求体。
type CreateOrderReq struct {
	Symbol string  `json:"symbol"`
	Side   string  `json:"side"`
	Qty    float64 `json:"qty"`
}

// Order 是下单成功后返回的订单。
type Order struct {
	ID     string  `json:"id"`
	Symbol string  `json:"symbol"`
	Side   string  `json:"side"`
	Qty    float64 `json:"qty"`
	Status string  `json:"status"`
}

// Server 是订单服务。用 sync.Mutex 保护内存里的订单表
// （真实项目里这里会是数据库；测试里用内存实现是为了让本关零依赖）。
type Server struct {
	mu     sync.Mutex
	orders map[string]Order
	seq    int
	slow   time.Duration // 仅用于演示「客户端超时」用例
}

// NewServer 构造一个空的服务实例。
func NewServer() *Server {
	return &Server{orders: map[string]Order{}}
}

// Handler 返回挂好路由的 http.Handler。
//
// 用 Go 1.22+ 的增强版 ServeMux：「METHOD /path」和「/path/{id}」都是内建语法，
// 不再需要 gorilla/mux 或 httprouter。
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/order", s.handleCreate)
	mux.HandleFunc("GET /api/order/{id}", s.handleGet)
	mux.HandleFunc("GET /healthz", s.handleHealth)
	return mux
}

// SetSlow 让下单接口故意变慢，用于演示客户端超时。
func (s *Server) SetSlow(d time.Duration) { s.slow = d }

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	if s.slow > 0 {
		time.Sleep(s.slow)
	}

	var req CreateOrderReq
	// ⚠ 这里刻意用了 json.NewDecoder().Decode 而不是 io.ReadAll + Unmarshal：
	//   前者能处理流式请求体，且对多余字段更宽容。
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// 错误响应目前是纯文本 —— 思考题 3 会让你改成统一 JSON 结构。
		http.Error(w, "请求体不是合法 JSON", http.StatusBadRequest)
		return
	}

	if req.Symbol == "" || req.Qty <= 0 {
		http.Error(w, "symbol 不能为空且 qty 必须大于 0", http.StatusBadRequest)
		return
	}
	if side := strings.ToUpper(req.Side); side != "BUY" && side != "SELL" {
		http.Error(w, "side 只能是 BUY 或 SELL", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.seq++
	o := Order{
		ID:     fmt.Sprintf("ord-%04d", s.seq),
		Symbol: strings.ToUpper(req.Symbol),
		Side:   strings.ToUpper(req.Side),
		Qty:    req.Qty,
		Status: "created",
	}
	s.orders[o.ID] = o
	s.mu.Unlock()

	// 先写 Header 再写状态码再写 Body —— 顺序反了 header 就丢
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Order-Id", o.ID)
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(o)
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id") // ← Go 1.22+ 的路由参数，替代了 mux.Vars(r)["id"]

	s.mu.Lock()
	o, ok := s.orders[id]
	s.mu.Unlock()

	if !ok {
		http.Error(w, "订单不存在", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(o)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
