//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 07-web3-trading && go run 11_metrics.go
// 依赖：github.com/prometheus/client_golang（含 promhttp，见 07-web3-trading/go.mod）
// 说明：/metrics 端点在本进程内用 httptest 起，再真的 GET 一次，全程离线可跑。
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

/*
═══════════════════════════════════════════════════════════════════════
W-11：Prometheus 指标
═══════════════════════════════════════════════════════════════════════

【核心概念】
    交易机器人跑起来之后，你**看不见**它。日志只在出事时才告诉你，
    而且日志是「离散事件」，回答不了「现在的下单延迟是多少」这种问题。
    指标（metrics）回答的正是这类问题：把程序内部的状态变成一串**可聚合的数字**，
    让 Prometheus 定时来抓，让 Grafana 画成曲线，让告警规则去判断「要不要叫你起床」。

    Prometheus 只有四种指标类型，交易机器人常用的就三种：

      Counter（计数器）  —— 只增不减。适合「累计发生了多少次」：
                            累计下单数、累计拒单数、累计成交笔数。
                            ⚠ Add(-1) 会直接 panic —— 它从设计上就不允许减。
                            想知道「近 5 分钟下了多少单」？用 rate() 在查询端算，
                            而不是让 Counter 自己减。

      Gauge（仪表盘）    —— 可增可减，就是「当前值」：
                            当前持仓、当前挂单数、当前未实现盈亏、当前连接数。
                            它没有单调性，所以 rate() 对它没有意义。

      Histogram（直方图）—— 把观测值按**预设的桶**分桶计数，暴露三组序列：
                            _bucket{le="0.1"}  累计落入 ≤0.1 的样本数（le = less or equal）
                            _sum               所有观测值之和
                            _count             样本总数
                            它能算出 P50/P95/P99，代价是「近似值」——
                            因为桶是离散的，你只知道 P99 落在哪个桶里。

    每个指标都可以带**标签（label）**：symbol / side / result。
    标签让一条指标裂变成多条时间序列，这是它强大的地方，也是它最危险的地方：
    标签组合数 = 基数（cardinality），基数爆炸会把 Prometheus 打爆
    （比如拿 orderID 当标签，几百万个订单就是几百万条时间序列）。

【你会看到】
    ✓ 三种指标的定义与桶的设计思路（为什么桶边界要贴着 SLO 选）
    ✓ 模拟 200 笔下单：成交 187 笔 / 拒单 13 笔（拒单率 6.5%）
    ✓ 真的 GET 一次自己的 /metrics 端点，拿到 HTTP 200、87 行、5792 字节真实文本
    ✓ Counter 按 symbol/side/result 裂成 16 条时间序列；Gauge 随买卖增减
    ✓ Histogram 的三组序列 _bucket / _sum / _count（BTCUSDT 40 个样本）
    ✓ 从桶反推 P99 落在 le="1" 桶，而原始数据精确 P99 = 795ms —— 这就是「近似」的代价
    ✓ 实测：全部 200 笔的 P50=40ms、P99=1389ms、平均=81ms（平均把长尾藏起来了）
    ✓ Counter.Add(-1) 真的 panic（已 recover），Gauge.Sub() 正常生效
    ✓ 一个意外收获：Gauge 打印出 2.500000000000001 —— float64 累加的漂移又出现了（W-01）
    ✓ 为什么交易机器人必须有指标：拒单率、P99 延迟、滑点

【思考题】（动手改代码验证，别空想）
    1. 把 orderCount 从 200 改成 20，再跑。从桶反推的 P99 和精确 P99 差距变大了还是变小了？
       为什么说「用 20 个样本谈 P99」本身就是耍流氓？
    2. 把 orderLatency 的 Buckets 改成只有 {0.1, 1}，再跑。
       你还能区分「120ms 的下单」和「900ms 的下单」吗？这说明桶边界该怎么选？
    3. 给 ordersTotal 再加一个标签 `order_id`（值用 ORD-1、ORD-2…），会发生什么？
       算一下：200 笔订单 × 2 种 side × 2 种 result 会裂成多少条时间序列？
       （提示：这就是「标签基数爆炸」，线上事故的常见来源）
    4. 示例 5 里我 recover 了 Counter 的 panic。生产代码里该不该这么做？
       如果不 recover，进程直接挂掉 —— 你觉得哪种更好？
    5. 「当前持仓」为什么必须是 Gauge 而不是 Counter？
       「累计成交量」为什么必须是 Counter 而不是 Gauge？（提示：单调性）
    6. 想监控「滑点」（成交价与下单价之差），你会怎么设计指标？
       用 Histogram 还是 Gauge？标签怎么打？（提示：滑点要的是分布，不是一个瞬时值）
    7. 示例 3 里 Gauge 打印出 2.500000000000001。如果「当前持仓」真的用 float64 存，
       还拿指标反过来当账本对账，会出什么问题？
       （提示：指标是业务状态的**投影**，不是账本本身；账本要用 decimal / 整数最小单位）

【通关标准】
    能说清 Counter / Gauge / Histogram 各自的语义和「不能怎么用」，
    能自己设计一套带标签的指标并用 NewRegistry + promhttp 暴露出去，
    并且知道「标签基数」是这套体系最容易踩的坑。

【前置】W-09（订单状态机）· L3-05（日志与分级）
【标签】Prometheus · Counter · Gauge · Histogram · 标签基数 · P99 · SLO
═══════════════════════════════════════════════════════════════════════
*/

// ─────────────────────────────────────────────────────────────────────────────
// 一、三种指标的定义
//
// 用 NewXxxVec（带标签的版本）而不是 NewXxx：Vec 允许用 WithLabelValues 打标签。
// 名字前缀 trading_ 是约定：同一进程里指标多了以后，前缀就是你的命名空间。
// ─────────────────────────────────────────────────────────────────────────────

var (
	// ordersTotal：累计下单数。Counter —— 只增不减。
	// 标签三个：symbol（交易对）、side（买/卖）、result（成交/拒单）。
	// 组合数 = 5 交易对 × 2 方向 × 2 结果 = 最多 20 条时间序列，完全可接受。
	ordersTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "trading_orders_total",
		Help: "累计下单数，按交易对/方向/结果分组。Counter 只增不减。",
	}, []string{"symbol", "side", "result"})

	// positionQty：当前持仓。Gauge —— 可增可减（买入加、卖出/平仓减）。
	positionQty = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "trading_position_qty",
		Help: "当前持仓数量（可增可减，卖出/平仓会让它下降）。",
	}, []string{"symbol"})

	// orderLatency：下单往返延迟分布（秒）。Histogram —— 给的是分布，不是瞬时值。
	//
	// ★ 桶（Buckets）怎么设计？这是 Histogram 唯一需要动脑的地方：
	//   · 桶边界要贴着你的 SLO（服务目标）放，而不是均匀撒。
	//     交易所下单的目标是 P99 < 250ms，所以 100ms / 250ms / 500ms 附近要密集，
	//     这样「P99 有没有越过 250ms 这条线」才能被准确告警。
	//   · 桶太少 → 分不清 200ms 和 900ms（都落进同一个桶）；
	//     桶太多 → 每条时间序列 ×桶数，基数直接乘上去。
	//   · 最后一个桶必须是 +Inf（库会自动加），否则算不出总数。
	//   · 单位统一用「秒」：Prometheus 的约定，_seconds 结尾，别用毫秒。
	orderLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "trading_order_latency_seconds",
		Help:    "下单请求往返耗时（秒）。桶边界贴着 P99<250ms 的 SLO 设计。",
		Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5},
	}, []string{"symbol"})
)

// ─────────────────────────────────────────────────────────────────────────────
// 二、模拟下单（真实项目里这些数字来自实际的交易所调用）
// ─────────────────────────────────────────────────────────────────────────────

const orderCount = 200

// symbols 是模拟订阅的 5 个交易对。标签的取值一定要**有限且可控** ——
// 这就是为什么 symbol 能当标签，而 orderID 不能。
var symbols = []string{"BTCUSDT", "ETHUSDT", "SOLUSDT", "BNBUSDT", "XRPUSDT"}

// simLatency 模拟一次下单的往返耗时：
//
//	主体落在 20~60ms（正常情况），约 3% 的概率进入长尾（再加 0.5~2s）。
//
// 长尾对应真实世界里的：交易所限频、网络抖动、服务端 GC 停顿、本地 GC 停顿。
// ★ 延迟分布从来不是正态的，它是「大部分很快 + 极少数极慢」——
//
//	所以看平均值毫无意义，必须看 P99。这就是 Histogram 存在的理由。
//
// 用固定种子的 rand/v2（PCG），保证每次运行的数字完全一样，方便对照答案。
func simLatency(rng *rand.Rand) float64 {
	base := 0.02 + rng.Float64()*0.04
	if rng.Float64() < 0.03 {
		base += 0.5 + rng.Float64()*1.5
	}
	return math.Round(base*10000) / 10000
}

// quantile 从**已排序**的原始数据里取分位数（教学用，用来和直方图近似值对比）。
func quantile(sorted []float64, q float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(math.Ceil(float64(len(sorted))*q)) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

// ─────────────────────────────────────────────────────────────────────────────
// 三、解析 /metrics 文本（真实项目里这是 Prometheus 干的活，这里手工看一眼）
// ─────────────────────────────────────────────────────────────────────────────

// fetchMetrics 真的发一次 HTTP GET，把 /metrics 的文本拿回来。
// 这就是 Prometheus 抓取（scrape）时做的事，一模一样。
func fetchMetrics(url string) (string, int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, err
	}
	return string(body), resp.StatusCode, nil
}

// findLine 在 exposition 文本里找第一行以 prefix 开头的行（含标签的样本行）。
func findLine(body, prefix string) string {
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, prefix) {
			return line
		}
	}
	return "(没找到)"
}

// histBucket 是一个直方图的桶：le（less-or-equal）上界 + 累计样本数。
type histBucket struct {
	le  float64
	cum uint64
}

// parseBuckets 从 exposition 文本里挑出「某个指标 + 某个标签值」的全部 _bucket 行。
//
// exposition 的格式长这样（标签按字母序排列，所以 le 排在最前面）：
//
//	trading_order_latency_seconds_bucket{le="0.005",symbol="BTCUSDT"} 0
//	trading_order_latency_seconds_bucket{le="+Inf",symbol="BTCUSDT"} 40
func parseBuckets(body, symbol string) []histBucket {
	out := []histBucket{}
	for _, line := range strings.Split(body, "\n") {
		if !strings.HasPrefix(line, "trading_order_latency_seconds_bucket{") {
			continue
		}
		if !strings.Contains(line, `symbol="`+symbol+`"`) {
			continue
		}
		leStr := between(line, `le="`, `"`)
		valStr := line[strings.LastIndex(line, " ")+1:]
		le, err1 := strconv.ParseFloat(leStr, 64)
		cum, err2 := strconv.ParseUint(valStr, 10, 64)
		if err1 != nil || err2 != nil {
			continue
		}
		out = append(out, histBucket{le: le, cum: cum})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].le < out[j].le })
	return out
}

// between 取 s 中 start 之后、end 之前的子串（第一个匹配）。
func between(s, start, end string) string {
	i := strings.Index(s, start)
	if i < 0 {
		return ""
	}
	rest := s[i+len(start):]
	j := strings.Index(rest, end)
	if j < 0 {
		return ""
	}
	return rest[:j]
}

// bucketOfQuantile 找出「分位点落在哪个桶」：第一个累计数 ≥ total*q 的桶。
// Prometheus 的 histogram_quantile() 就是干这个，再在桶内做线性插值。
func bucketOfQuantile(buckets []histBucket, q float64, total uint64) (histBucket, bool) {
	target := uint64(math.Ceil(float64(total) * q))
	for _, b := range buckets {
		if b.cum >= target {
			return b, true
		}
	}
	return histBucket{}, false
}

// counterDecreaseMsg 演示 Counter 只能加不能减：Add(-1) 会直接 panic。
//
// 生产代码里**不该**这么 recover —— panic 说明你把指标类型选错了
// （比如把「当前持仓」写成了 Counter）。让它炸，比悄悄埋一个错误语义好。
func counterDecreaseMsg() (msg string) {
	defer func() {
		if r := recover(); r != nil {
			msg = fmt.Sprint(r)
		}
	}()
	ordersTotal.WithLabelValues("BTCUSDT", "BUY", "filled").Add(-1)
	return "没有 panic —— 不应该发生"
}

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 W-11 Prometheus 指标")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// ── 示例 1：三种指标的定义 ──────────────────────────────────────────
	fmt.Println("📝 示例 1：三种指标的语义（选错类型 = 语义错误，比崩溃更可怕）")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("  Counter  只增不减：累计下单数 / 累计拒单数 / 累计成交笔数")
	fmt.Println("           ⚠ Add(-1) 直接 panic；想算「近 5 分钟的下单数」用查询端的 rate()")
	fmt.Println("  Gauge    可增可减：当前持仓 / 当前挂单数 / 当前未实现盈亏 / 当前连接数")
	fmt.Println("           ⚠ 它没有单调性，对它用 rate() 没有意义")
	fmt.Println("  Histogram 给的是分布：_bucket（累计分桶）+ _sum（和）+ _count（总数）")
	fmt.Println("           ⚠ 桶边界是离散的，所以分位数只能是**近似值**")
	fmt.Println("\n  本例的桶设计：", []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5}, "（秒）")
	fmt.Println("  ➡️  为什么这么选：交易所下单的 SLO 是 P99 < 250ms，所以 100ms / 250ms / 500ms")
	fmt.Println("    附近要放密集的桶，这样「P99 有没有越线」才能被准确告警；")
	fmt.Println("    均匀撒桶（比如 0.1 / 0.2 / 0.3 …）在真实延迟分布下几乎全是浪费。")

	// ── 示例 2：注册到自己的 Registry，模拟 200 笔下单 ──────────────────
	fmt.Println("📌 示例 2：注册到自己的 Registry，模拟 200 笔下单")
	fmt.Println(strings.Repeat("-", 60))
	// ⚠ 用 NewRegistry() 而不是默认的 prometheus.DefaultRegisterer：
	//   默认 Registry 是全局单例，多个库往里塞同名指标会 panic（duplicate metrics collector）。
	//   自己建一个，边界清晰，测试也好写。
	reg := prometheus.NewRegistry()
	reg.MustRegister(ordersTotal, positionQty, orderLatency)
	fmt.Println("  ✓ 已注册 3 个指标到自己的 Registry（不用全局 DefaultRegisterer，避免同名冲突）")

	rng := rand.New(rand.NewPCG(0xC0FFEE, 2024)) // 固定种子 → 输出可复现
	latencies := []float64{}
	filled, rejected := 0, 0

	for i := 0; i < orderCount; i++ {
		symbol := symbols[i%len(symbols)]
		side := "BUY"
		if i%7 == 0 {
			side = "SELL" // 约 1/7 是卖出，让 Gauge 有机会下降
		}

		lat := simLatency(rng)
		latencies = append(latencies, lat)
		orderLatency.WithLabelValues(symbol).Observe(lat)

		result := "filled"
		// 超时/限频导致的拒单：延迟特别大，或小概率随机拒单
		if lat > 0.5 || rng.Float64() < 0.02 {
			result = "rejected"
			rejected++
		} else {
			filled++
		}
		ordersTotal.WithLabelValues(symbol, side, result).Inc()

		if result == "filled" {
			if side == "BUY" {
				positionQty.WithLabelValues(symbol).Add(0.1) // 买入：持仓增加
			} else {
				positionQty.WithLabelValues(symbol).Sub(0.1) // 卖出：持仓减少 ← Gauge 才能这么干
			}
		}
	}

	sorted := append([]float64{}, latencies...)
	sort.Float64s(sorted)
	fmt.Printf("  ✓ 已记录 %d 笔下单：成交 %d 笔、拒单 %d 笔\n", orderCount, filled, rejected)
	fmt.Printf("  原始数据的延迟分布（不是从指标里读的，是直接用模拟数据算的）：\n")
	fmt.Printf("    P50=%.0fms  P90=%.0fms  P95=%.0fms  P99=%.0fms  Max=%.0fms  平均=%.0fms\n",
		quantile(sorted, 0.50)*1000, quantile(sorted, 0.90)*1000, quantile(sorted, 0.95)*1000,
		quantile(sorted, 0.99)*1000, sorted[len(sorted)-1]*1000, mean(sorted)*1000)
	fmt.Println("  ➡️  注意「平均」和「P99」差了多远 —— 平均值会把长尾藏起来，")
	fmt.Println("    而长尾正是你在实盘里被滑点打脸的地方。监控永远看分位数。")

	// ── 示例 3：真的 GET 一次自己的 /metrics ────────────────────────────
	fmt.Println("🔍 示例 3：真的 GET 一次自己的 /metrics 端点（Prometheus 就是这么抓的）")
	fmt.Println(strings.Repeat("-", 60))
	mux := http.NewServeMux()
	// HandlerFor 把 Registry 渲染成 exposition 文本；传自己的 reg，不用全局的。
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	metricsURL := srv.URL + "/metrics"
	body, status, err := fetchMetrics(metricsURL)
	if err != nil {
		fmt.Printf("  ✗ 抓取失败：%v\n", err)
		return
	}
	lines := strings.Split(strings.TrimRight(body, "\n"), "\n")
	fmt.Printf("  GET %s → HTTP %d，%d 行、%d 字节\n", metricsURL, status, len(lines), len(body))
	fmt.Println("  ↓ 真实的 exposition 文本（截取几个片段）")

	for _, m := range []string{"trading_orders_total", "trading_position_qty", "trading_order_latency_seconds"} {
		for _, line := range lines {
			if strings.HasPrefix(line, "# HELP "+m+" ") || strings.HasPrefix(line, "# TYPE "+m+" ") {
				fmt.Println("    " + line)
			}
		}
	}

	fmt.Println("    ...")
	fmt.Println("  Counter 的样本（注意：只增不减，每个标签组合一条时间序列）：")
	shown := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "trading_orders_total{") {
			fmt.Println("    " + line)
			shown++
			if shown == 4 {
				break
			}
		}
	}
	totalSeries := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "trading_orders_total{") {
			totalSeries++
		}
	}
	fmt.Printf("    ... 共 %d 条 orders_total 时间序列（5 交易对 × 2 方向 × 2 结果，未出现的不占位）\n", totalSeries)

	fmt.Println("  Gauge 的样本（当前持仓，买卖都会改它）：")
	for _, line := range lines {
		if strings.HasPrefix(line, "trading_position_qty{") {
			fmt.Println("    " + line)
		}
	}
	fmt.Println("  ⚠️ 看到 2.500000000000001 了吗？指标的数值也是 float64 ——")
	fmt.Println("     0.1 累加 25 次就会漂成 2.500000000000001（见 W-01）。")
	fmt.Println("     所以「持仓/余额」这类东西在业务代码里必须用 decimal 或整数最小单位，")
	fmt.Println("     指标只是它的一个**投影**，别把指标里的值反过来当成账本。")

	// ── 示例 4：Histogram 的三组序列 ────────────────────────────────────
	fmt.Println("🧪 示例 4：Histogram 给的是三组序列：_bucket / _sum / _count")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("  ↓ 只看 BTCUSDT 这一组标签（其它交易对同理）：")
	buckets := parseBuckets(body, "BTCUSDT")
	for _, b := range buckets {
		le := strconv.FormatFloat(b.le, 'f', -1, 64)
		if math.IsInf(b.le, 1) {
			le = "+Inf"
		}
		fmt.Printf("    trading_order_latency_seconds_bucket{le=\"%s\",symbol=\"BTCUSDT\"} %d\n", le, b.cum)
	}
	fmt.Println("    " + findLine(body, `trading_order_latency_seconds_sum{symbol="BTCUSDT"}`))
	fmt.Println("    " + findLine(body, `trading_order_latency_seconds_count{symbol="BTCUSDT"}`))
	fmt.Println("  ➡️  _bucket 是**累计**的（le=\"0.25\" 里包含了所有 ≤0.25 的样本），")
	fmt.Println("    所以最后那个 +Inf 桶的数值 == _count，这是自检直方图是否完整的小技巧。")

	var count uint64
	if len(buckets) > 0 {
		count = buckets[len(buckets)-1].cum
	}
	// 只看 BTCUSDT 的样本，和直方图是同一批数据，正好对比「精确 vs 近似」
	btc := []float64{}
	for i := 0; i < orderCount; i++ {
		if symbols[i%len(symbols)] == "BTCUSDT" {
			btc = append(btc, latencies[i])
		}
	}
	sort.Float64s(btc)
	exactP99 := quantile(btc, 0.99)
	if b, ok := bucketOfQuantile(buckets, 0.99, count); ok {
		le := strconv.FormatFloat(b.le, 'f', -1, 64)
		if math.IsInf(b.le, 1) {
			le = "+Inf"
		}
		fmt.Printf("\n  BTCUSDT 共 %d 个样本：\n", count)
		fmt.Printf("    原始数据精确 P99 = %.0fms\n", exactP99*1000)
		fmt.Printf("    从桶反推 P99     = 落在 le=\"%s\" 这个桶里（该桶累计 %d 个样本 ≥ %d×0.99）\n",
			le, b.cum, count)
		fmt.Println("    ↑ 直方图只能告诉你「P99 落在哪个桶」，给不出精确值 —— 这就是近似的来源。")
		fmt.Println("      Prometheus 的 histogram_quantile() 会在桶内做线性插值，比这准一些，")
		fmt.Println("      但本质上仍是「用桶宽换存储」。桶越密越准，基数也越大。")
		fmt.Printf("    ⚠️ 只有 %d 个样本时 P99 其实就是最大值 —— 样本太少时谈 P99 是耍流氓，\n", count)
		fmt.Println("       真实项目里至少要有几千个样本，或者把窗口拉长（rate 5m/1h）。")
	}

	// ── 示例 5：Counter 不能减，Gauge 可以 ──────────────────────────────
	fmt.Println("⚙️ 示例 5：Counter 不能减（panic），Gauge 可以（正常生效）")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("  ordersTotal.Add(-1) → panic: %s\n", counterDecreaseMsg())
	fmt.Println("  ➡️  这是库的主动设计：Counter 的语义就是「累计发生过的次数」，减它没有意义。")
	fmt.Println("    要表达「减少」，说明你需要的其实是 Gauge。")

	posBefore := findLine(body, `trading_position_qty{symbol="BTCUSDT"}`)
	positionQty.WithLabelValues("BTCUSDT").Sub(0.1)
	body2, _, err2 := fetchMetrics(metricsURL)
	posAfter := findLine(body2, `trading_position_qty{symbol="BTCUSDT"}`)
	if err2 == nil {
		fmt.Println("\n  Gauge 演示：手动平掉 0.1 个 BTC，重新抓一次 /metrics 看真实变化")
		fmt.Printf("    Sub 之前：%s\n", posBefore)
		fmt.Printf("    Sub 之后：%s\n", posAfter)
		fmt.Println("    ↑ 同一个指标真的变小了 —— Gauge 允许双向变化，Counter 不允许。")
	}

	// ── 示例 6：为什么交易机器人必须有指标 ──────────────────────────────
	fmt.Println("🎯 示例 6：为什么交易机器人必须有指标")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("  ① 拒单率：%d/%d = %.1f%% —— 拒单率突然升高通常意味着限频、参数越界或余额不足\n",
		rejected, orderCount, float64(rejected)/float64(orderCount)*100)
	fmt.Printf("  ② P99 延迟：%.0fms —— 下单慢 1 秒，价格可能已经跑了，滑点直接吃掉利润\n",
		quantile(sorted, 0.99)*1000)
	fmt.Println("  ③ 滑点：成交价 vs 下单价之差 —— 要的是**分布**，所以又是一个 Histogram")
	fmt.Println("     trading_slippage_bps{symbol=...}，桶放在 1/5/10/25/50 bps")
	fmt.Println("  ④ 累计成交/拒单：Counter + rate() 就能得到「每分钟下单速率」，用来对账")
	fmt.Println("  ➡️  指标的价值在于「出问题之前就能看见」：日志是事后追凶，指标是事前告警。")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ Counter 只增不减（Add(-1) panic）；算速率用查询端的 rate()")
	fmt.Println("  ✅ Gauge 可增可减，表示「当前值」；对它用 rate() 没有意义")
	fmt.Println("  ✅ Histogram 给三组序列：_bucket（累计分桶）/ _sum / _count，能算近似分位数")
	fmt.Println("  ✅ 桶边界要贴着 SLO 放；最后一个 +Inf 桶的数值应该等于 _count")
	fmt.Println("  ✅ 标签让指标裂变：symbol/side/result 可以，orderID 绝不可以（基数爆炸）")
	fmt.Println("  ✅ 用 prometheus.NewRegistry() + promhttp.HandlerFor 暴露 /metrics，别用全局 Registry")
	fmt.Println("  ✅ 交易机器人最该盯的三个：拒单率、下单 P99 延迟、滑点分布")
	fmt.Println("  ✅ 该用：任何需要长期观察趋势/告警的东西（延迟、速率、存量、错误率）")
	fmt.Println("  ✅ 不该用：单次请求的细节（那是日志和 trace 的活，塞进指标只会把 Prometheus 撑爆）")
}

// mean 求平均值，只用于对比「平均值 vs 分位数」的差距。
func mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	sum := 0.0
	for _, x := range xs {
		sum += x
	}
	return sum / float64(len(xs))
}
