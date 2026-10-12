// Package backtest 用一段**确定的价格序列**跑同一条管线
// （strategy → risk → executor → store），最后输出一份盈亏报告。
//
// 它的价值在于：没有任何网络、没有任何密钥，也能评估策略。
// 内置四种形态 + 外部 CSV，固定种子保证同一 (形态, 笔数, 种子) 永远得到同一段行情 ——
// 回测结果不可复现，就等于没有回测。
package backtest

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"go-study/trading-bot/internal/market"
)

// Pattern 是内置的价格形态。
type Pattern string

const (
	PatternUp        Pattern = "up"    // 单边上涨
	PatternDown      Pattern = "down"  // 单边下跌
	PatternRange     Pattern = "range" // 震荡（区间来回）
	PatternVReversal Pattern = "v"     // V 型反转
	PatternCSV       Pattern = "csv"   // 外部 CSV（由 LoadCSV 提供序列）
)

// DefaultSymbol 是内置序列的交易对。
const DefaultSymbol = "BTCUSDT"

// BarInterval 是内置序列相邻两根 K 线的间隔。
//
// 回测里风控用的是**行情时间**（Quote.Ts）而不是墙钟时间：回测跑得再快也不代表行情快，
// 否则 25ms 的冷却期会把所有信号都拦掉，结论完全失真。
const BarInterval = time.Minute

const (
	startPrice  = 63000.0 // 起始价，贴近 BTC 的量级，名义金额才有意义
	priceDigits = 2       // 序列价格保留 2 位小数，模拟交易所返回的原始字符串
)

// AllPatterns 返回内置形态（不含 csv）。
func AllPatterns() []Pattern {
	return []Pattern{PatternUp, PatternDown, PatternRange, PatternVReversal}
}

// PatternName 返回形态的中文名，报告里用。
func PatternName(p Pattern) string {
	switch p {
	case PatternUp:
		return "单边上涨"
	case PatternDown:
		return "单边下跌"
	case PatternRange:
		return "震荡"
	case PatternVReversal:
		return "V 型反转"
	case PatternCSV:
		return "外部 CSV"
	default:
		return string(p)
	}
}

// ParsePattern 把 -pattern 的字符串转成 Pattern，非法值给出可选列表。
func ParsePattern(s string) (Pattern, error) {
	v := Pattern(strings.ToLower(strings.TrimSpace(s)))
	switch v {
	case PatternUp, PatternDown, PatternRange, PatternVReversal:
		return v, nil
	case PatternCSV:
		return "", fmt.Errorf("形态 %q 请改用 -csv=<文件路径> 指定", s)
	default:
		names := make([]string, 0, len(AllPatterns()))
		for _, p := range AllPatterns() {
			names = append(names, string(p))
		}
		return "", fmt.Errorf("未知形态 %q（可选：%s）", s, strings.Join(names, " / "))
	}
}

// Series 生成 n 根 K 线。固定 seed，结果可复现。
//
// 四种形态都是「趋势骨架 + 往复摆动」的合成，而不是直线：
// 直线行情下均线交叉只会朝一个方向连续触发，策略会朝一个方向无限加仓，
// 回测数字完全失真（这也是真实回测里必须用带回调的行情的原因）。
//
//   - up   ：500 根内累计 +30%，叠加 3%/40 根的摆动；
//   - down ：500 根内累计 −30%，叠加同样的摆动；
//   - range：无趋势，只有 4%/60 根的摆动，价格始终被拉回区间；
//   - v    ：前半段跌到 −30%、后半段涨回 0，叠加同样的摆动。
func Series(p Pattern, n int, seed int64) ([]market.Quote, error) {
	if n <= 0 {
		return nil, fmt.Errorf("行情笔数必须为正，实际 %d", n)
	}
	if _, err := ParsePattern(string(p)); err != nil {
		return nil, err
	}

	rnd := rand.New(rand.NewSource(seed))
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	out := make([]market.Quote, 0, n)

	for i := 0; i < n; i++ {
		var rel float64
		switch p {
		case PatternUp:
			rel = 0.30*float64(i)/float64(n) + oscillation(i)
		case PatternDown:
			rel = -0.30*float64(i)/float64(n) + oscillation(i)
		case PatternRange:
			rel = 0.04 * math.Sin(2*math.Pi*float64(i)/60)
		case PatternVReversal:
			rel = vShape(i, n) + oscillation(i)
		}
		price := startPrice * (1 + rel + rnd.NormFloat64()*0.0004)
		out = append(out, newQuote(DefaultSymbol, price, base.Add(time.Duration(i)*BarInterval)))
	}
	return out, nil
}

// oscillation 是叠加在趋势上的往复摆动：振幅 3%、周期 40 根。
//
// 振幅必须**明显大于趋势每根的贡献**（30%/500 根 = 0.06%/根），否则趋势项会把摆动
// 的一个方向顶到阈值以下，均线交叉又退化成单方向触发。3% 对应的相对偏离约
// 0.03×2π/40 ≈ 0.47%，两个方向都稳稳越过 0.2% 的阈值。
func oscillation(i int) float64 {
	return 0.03 * math.Sin(2*math.Pi*float64(i)/40)
}

// vShape 是 V 型骨架：前半段线性跌到 −30%，后半段线性涨回 0。
func vShape(i, n int) float64 {
	half := float64(n) / 2
	x := float64(i)
	if x < half {
		return -0.30 * x / half
	}
	return -0.30 + 0.30*(x-half)/half
}

// newQuote 把 float64 价格落成「字符串 + float64」两种表示，与 market.Quote 的约定一致：
// 字符串给下单转 decimal 用（不把浮点误差带进金额），float64 给策略算均线用。
func newQuote(symbol string, price float64, ts time.Time) market.Quote {
	s := strconv.FormatFloat(price, 'f', priceDigits, 64)
	d := decimal.RequireFromString(s) // s 由 FormatFloat 生成，必定可解析
	return market.Quote{Symbol: symbol, Price: d.InexactFloat64(), PriceStr: s, Ts: ts}
}

// LoadCSV 读外部行情 CSV。要求两列：timestamp,price。
//
// timestamp 支持 RFC3339（2024-01-01T00:00:00Z）与 Unix 秒（1704067200）；
// 首行若是表头（timestamp/time/ts/date/datetime、price/close/last）自动跳过。
// 任何一行有问题都**明确报错**（带上行号），绝不静默跳过 —— 静默跳过会让人
// 拿着半截数据得出错误结论。
func LoadCSV(path string) ([]market.Quote, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开行情 CSV 失败: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1 // 列数自己校验，好给出「缺列」的明确错误
	var out []market.Quote

	for line := 1; ; line++ {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("读取 CSV 第 %d 行失败: %w", line, err)
		}
		if line == 1 && isHeader(rec) {
			continue
		}
		if len(rec) < 2 {
			return nil, fmt.Errorf("CSV 第 %d 行列数不足：需要 timestamp,price 两列，实际 %d 列", line, len(rec))
		}

		ts, err := parseTimestamp(strings.TrimSpace(rec[0]))
		if err != nil {
			return nil, fmt.Errorf("CSV 第 %d 行时间戳 %q 非法（期望 RFC3339 或 Unix 秒）: %w", line, rec[0], err)
		}
		priceStr := strings.TrimSpace(rec[1])
		price, err := decimal.NewFromString(priceStr)
		if err != nil {
			return nil, fmt.Errorf("CSV 第 %d 行价格 %q 不是数字: %w", line, priceStr, err)
		}
		if !price.IsPositive() {
			return nil, fmt.Errorf("CSV 第 %d 行价格 %s 必须为正数", line, price)
		}

		out = append(out, market.Quote{
			Symbol:   DefaultSymbol,
			Price:    price.InexactFloat64(),
			PriceStr: price.String(), // 原样保留十进制表示，不让浮点过一遍
			Ts:       ts,
		})
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("行情 CSV %s 没有任何有效数据行", path)
	}
	return out, nil
}

// isHeader 判断首行是不是表头：只要有一列是常见的列名就算。
func isHeader(rec []string) bool {
	if len(rec) < 2 {
		return false
	}
	return isHeaderName(rec[0]) || isHeaderName(rec[1])
}

func isHeaderName(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "timestamp", "time", "ts", "date", "datetime", "price", "close", "last":
		return true
	}
	return false
}

// parseTimestamp 支持 RFC3339 与 Unix 秒。
func parseTimestamp(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	if sec, err := strconv.ParseInt(s, 10, 64); err == nil {
		return time.Unix(sec, 0).UTC(), nil
	}
	return time.Time{}, fmt.Errorf("无法解析 %q", s)
}
