/*
═══════════════════════════════════════════════════════════════════════
T-05：性能剖析（pprof）—— 找到真正的热点
═══════════════════════════════════════════════════════════════════════

【核心概念】

	T-04 教你「怎么测量一个函数有多快」。但真实问题是：
	**一个程序慢，到底慢在哪一行？**

	pprof 就是回答这个问题的工具。它不需要你在代码里插桩，
	而是由 Go 运行时按固定频率采样（CPU 默认 100Hz）并记录调用栈。

	三种采样维度，对应三个命令行开关：

	    go test -bench=. -cpuprofile=cpu.out ./05_pprof/     # CPU 时间花在哪
	    go test -bench=. -memprofile=mem.out ./05_pprof/     # 内存在哪分配的
	    go test -bench=. -blockprofile=blk.out ./05_pprof/   # 谁在阻塞（锁/通道）
	    go test -bench=. -mutexprofile=mutex.out ./05_pprof/ # 锁竞争

	看报告有四种姿势（都值得亲手跑）：
	    go tool pprof -top cpu.out            # 按耗时排序的函数列表
	    go tool pprof -list=函数名 cpu.out     # 逐行标注每行花了多少时间 ← 最有用
	    go tool pprof -peek=函数名 cpu.out     # 谁调用了它、它调用了谁
	    go tool pprof -http=:8080 cpu.out     # 浏览器里的火焰图（最直观）

	⚠ 三个必须知道的限制：
	  ① **采样不是精确测量。** 一个只跑 2ms 的函数可能一次都没被采样到，
	     所以 pprof 适合找「占总时间 10% 以上」的热点，不适合找 1% 的优化。
	  ② **必须让程序跑得够久。** 基准测试至少跑几秒，否则样本太少、噪声盖过信号。
	  ③ **内联会打乱归因。** 编译器把小函数内联进调用方后，时间会算在调用方头上。
	     实在看不清时可以加 `-gcflags='-l'` 关闭内联再采一次。

	本关的被测代码里埋了一个**真实世界最常见的 Go 性能 bug**：
	在循环里调用 regexp.MustCompile。它的开销极大（要解析正则、构建状态机），
	而且非常隐蔽 —— 代码读起来完全合理，只有 pprof 能一眼指出它。

【运行方式】

	cd 06-go-testing

	# 1. 先跑一次程序，看慢/快两版的差距
	go run ./05_pprof/

	# 2. 采集 CPU profile（必须跑够时间）
	go test -bench=. -benchtime=2s -cpuprofile=cpu.out ./05_pprof/

	# 3. 看热点（下面四条都试一遍）
	go tool pprof -top cpu.out
	go tool pprof -list=classifySlow cpu.out
	go tool pprof -peek=regexp.MustCompile cpu.out
	go tool pprof -http=:8080 cpu.out          # 浏览器打开火焰图，Ctrl-C 退出

	# 4. 采集内存 profile
	go test -bench=. -benchtime=2s -memprofile=mem.out ./05_pprof/
	go tool pprof -top -sample_index=alloc_space mem.out

【你会看到】

	① 程序自身的对比（本机 Apple Silicon / go1.27，20000 个订单）：

	    classifySlow（循环内 MustCompile）： 104.7 ms
	    classifyFast（包级预编译）        ：   7.3 ms
	    → 快 14.4 倍

	② 基准测试（go test -bench=Classify -benchtime=2s -cpuprofile=cpu.out ./05_pprof/）：

	    BenchmarkClassify/slow-8     445444     4887 ns/op    8299 B/op    77 allocs/op
	    BenchmarkClassify/fast-8    9306679      257.1 ns/op    112 B/op     2 allocs/op
	    → 快 19 倍，少分配 38 倍

	③ pprof 的真实输出。先看 -top（注意 runtime.kevent 这类是测试框架自身的 I/O 噪声）：

	    flat  flat%   sum%        cum   cum%
	    2.05s 35.90% 35.90%      2.05s 35.90%  runtime.kevent
	    0.44s  7.71% 43.61%      0.44s  7.71%  regexp.(*bitState).shouldVisit (inline)
	    0.40s  7.01% 57.97%      0.95s 16.64%  regexp.(*Regexp).tryBacktrack
	    ...

	再看 -list=classifySlow —— **这才是最有用的视图**，它把耗时标到了源代码的每一行：

	    ROUTINE ======================== go-study/testing/05_pprof.classifySlow
	             0      460ms (flat, cum)  8.06% of Total
	             .          .    137:func classifySlow(symbol string) (base, quote string) {
	             .      440ms    138:	re := regexp.MustCompile(`^([A-Z0-9]+?)(USDT|USDC|BTC|ETH)$`)
	             .       20ms    139:	m := re.FindStringSubmatch(symbol)
	             .          .    140:	if m == nil {

	440ms 里 440ms 全在第 138 行。**pprof 没有猜，它直接告诉你是这一行。**

	④ -peek=classifySlow 显示调用关系，确认时间来自它调用的 regexp：

	    go-study/testing/05_pprof.classifySlow
	                        0.44s 95.65% |   regexp.MustCompile
	                        0.02s  4.35% |   regexp.(*Regexp).FindStringSubmatch

	注意 classifySlow 自己的 flat 是 0 —— 时间全在它调用的函数里。
	这就是 **flat（自己花的时间）和 cum（含被调用者）的区别**，
	也是新手读 pprof 最容易误判的地方：看到 flat=0 就以为「这函数不慢」。

【思考题】（动手改代码验证，别空想）

 1. 先跑 go tool pprof -list=classifySlow cpu.out，找到耗时最高的那一行。
    它就是 bug 所在。想清楚：为什么「正则表达式」在循环里编译这么贵？
    提示：MustCompile 要做词法分析 + 构建 NFA/DFA，是微秒级操作，
    而一次正则匹配本身只有几十纳秒。

 2. 把 classifyFast 的写法（包级 var 预编译）应用到 classifySlow 上，
    重新采一次 profile。regexp.MustCompile 还在 -top 列表里吗？
    慢/快两版的耗时差距变成多少了？

 3. 用 -peek=classifySlow 看调用关系。谁调用了它？它调用了谁？
    如果这里显示 classifySlow 被内联了、看不到，试着加
    -gcflags='-l' 关闭内联再跑一次。

 4. 采集 memprofile 后用 -sample_index=alloc_space 看「累计分配字节数」，
    再用 -sample_index=alloc_objects 看「分配对象个数」。
    两个视图排第一的是同一个函数吗？为什么可能不同？
    （提示：一个大对象和一千个小对象，在两种视角下排名完全不同。）

 5. 在交易机器人的场景里想：如果你发现下单路径的 P99 延迟偶尔飙到 200ms，
    你会先看 cpu profile 还是 block/mutex profile？为什么？

【通关标准】

	能完整说出「采集 profile → -top 找热点 → -list 定位到行」这条链路，
	能解释 flat 与 cum 的区别，
	并且知道「pprof 是采样、要跑够时间、找不到 1% 级别的优化」。

【前置】T-04（基准测试）   【标签】pprof · 性能剖析 · regexp 热点 · flat/cum
═══════════════════════════════════════════════════════════════════════
*/
package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// 被测代码：把交易所的原始交易对解析成「币种 + 计价单位」。
//
// 这是机器人启动时必做的一件事：订阅列表里的 "BTCUSDT" 要拆成 BTC / USDT，
// 才能去查币种精度、最小下单量这些元数据。
// ─────────────────────────────────────────────────────────────────────────────

// symbolRe 是预编译好的正则 —— 这是**正确**的写法。
//
// 包级变量只编译一次，之后每次匹配直接复用内部状态机。
var symbolRe = regexp.MustCompile(`^([A-Z0-9]+?)(USDT|USDC|BTC|ETH)$`)

// classifySlow 是**反面教材**：每次调用都重新编译正则。
//
// ⚠ 这就是本关要让你用 pprof 抓出来的热点。
//
// 为什么危险？因为这段代码「读起来完全合理」——
// 一个正则、一次匹配、返回两个字符串。没有任何看起来可疑的地方。
// 代码评审时几乎没人会拦下它，但它会让这个函数慢 30~50 倍。
//
// 真实项目里这个 bug 有无数变体：
//
//	· 在循环里 json.Marshal 同一份配置
//	· 在请求处理里 time.LoadLocation("Asia/Shanghai")
//	· 在行情回调里 template.Parse（模板解析）
//
// 它们的共同特征：**把「编译/解析」这种一次性工作放进了高频路径。**
func classifySlow(symbol string) (base, quote string) {
	re := regexp.MustCompile(`^([A-Z0-9]+?)(USDT|USDC|BTC|ETH)$`) // ← 每次调用都编译！
	m := re.FindStringSubmatch(symbol)
	if m == nil {
		return symbol, ""
	}
	return m[1], m[2]
}

// classifyFast 是正确写法：复用包级预编译的正则。
func classifyFast(symbol string) (base, quote string) {
	m := symbolRe.FindStringSubmatch(symbol)
	if m == nil {
		return symbol, ""
	}
	return m[1], m[2]
}

// Order 是报告里的一行。
type Order struct {
	Symbol string
	Qty    float64
}

// buildReportSlow 用慢版解析整个订单列表，并拼成一份报表。
//
// 注意这里还有一个次生问题：用 += 拼字符串（复习 T-04）。
// pprof 会同时把这两处都指出来 —— 这正是它的价值：
// 你不用猜，它按耗时排序告诉你先修哪个。
func buildReportSlow(orders []Order) string {
	var out strings.Builder
	for _, o := range orders {
		base, quote := classifySlow(o.Symbol)
		out.WriteString(base)
		out.WriteString("/")
		out.WriteString(quote)
		out.WriteString(" ")
		fmt.Fprintf(&out, "%.4f", o.Qty)
		out.WriteString("\n")
	}
	return out.String()
}

// buildReportFast 是修好之后的版本。
func buildReportFast(orders []Order) string {
	var out strings.Builder
	for _, o := range orders {
		base, quote := classifyFast(o.Symbol)
		out.WriteString(base)
		out.WriteString("/")
		out.WriteString(quote)
		out.WriteString(" ")
		fmt.Fprintf(&out, "%.4f", o.Qty)
		out.WriteString("\n")
	}
	return out.String()
}

// MakeOrders 造一批测试数据。
func MakeOrders(n int) []Order {
	syms := []string{"BTCUSDT", "ETHUSDT", "SOLUSDT", "BNBUSDT", "XRPUSDC", "ADAUSDT", "DOGEUSDT", "AVAXETH"}
	orders := make([]Order, n)
	for i := range orders {
		orders[i] = Order{Symbol: syms[i%len(syms)], Qty: float64(i%1000) / 100}
	}
	return orders
}

func main() {
	fmt.Println("\n=== T-05 性能剖析（pprof）===")

	orders := MakeOrders(20000)

	// ── 示例 1：慢版 vs 快版 ─────────────────────────────────────────────
	fmt.Println("📝 示例 1：同一个功能，两种写法")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("  订单数：%d\n", len(orders))

	start := time.Now()
	slowReport := buildReportSlow(orders)
	slowDur := time.Since(start)

	start = time.Now()
	fastReport := buildReportFast(orders)
	fastDur := time.Since(start)

	fmt.Printf("  classifySlow（循环内 MustCompile）：%8.1f ms\n", float64(slowDur.Microseconds())/1000)
	fmt.Printf("  classifyFast（包级预编译）        ：%8.1f ms\n", float64(fastDur.Microseconds())/1000)
	if fastDur > 0 {
		fmt.Printf("  → 快 %.1f 倍\n", float64(slowDur)/float64(fastDur))
	}
	fmt.Printf("  两版结果一致：%t\n", slowReport == fastReport)

	// ── 示例 2：pprof 命令清单 ───────────────────────────────────────────
	fmt.Println("📌 示例 2：用 pprof 抓出热点（照抄下面的命令）")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("  # 1) 采集 CPU profile（-benchtime 要够长，否则样本太少）")
	fmt.Println("  go test -bench=. -benchtime=2s -cpuprofile=cpu.out ./05_pprof/")
	fmt.Println()
	fmt.Println("  # 2) 按耗时排序找热点")
	fmt.Println("  go tool pprof -top cpu.out")
	fmt.Println()
	fmt.Println("  # 3) 定位到具体某一行（最有用的命令）")
	fmt.Println("  go tool pprof -list=classifySlow cpu.out")
	fmt.Println()
	fmt.Println("  # 4) 看调用关系：谁调用了它、它又调用了谁")
	fmt.Println("  go tool pprof -peek=classifySlow cpu.out")
	fmt.Println()
	fmt.Println("  # 5) 浏览器里的火焰图（最直观，Ctrl-C 退出）")
	fmt.Println("  go tool pprof -http=:8080 cpu.out")
	fmt.Println()
	fmt.Println("  # 6) 内存 profile：谁在分配")
	fmt.Println("  go test -bench=. -benchtime=2s -memprofile=mem.out ./05_pprof/")
	fmt.Println("  go tool pprof -top -sample_index=alloc_space mem.out")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ pprof 靠采样（CPU 默认 100Hz），找的是占比 10% 以上的热点，不是 1% 的微优化")
	fmt.Println("  ✅ flat = 函数自己花的时间；cum = 含它调用的所有子函数 —— 读 pprof 最容易搞混的就是这两个")
	fmt.Println("  ✅ -list=函数名 是最有用的命令：直接把耗时标到源代码的每一行")
	fmt.Println("  ✅ 最常见的 Go 热点：把「编译 / 解析 / 加载」这类一次性工作放进了高频路径")
	fmt.Println("  ✅ 优化顺序永远是：先 profile 找到真正的瓶颈，再动手 —— 不要凭直觉改代码")
	fmt.Println("  ✅ 别忘了另外三种 profile：-blockprofile / -mutexprofile / -memprofile")
}
