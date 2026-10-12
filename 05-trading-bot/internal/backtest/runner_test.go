// 回测交易所与 Runner 的测试。
//
// Runner 的断言落在**管线不变量**上（信号 = 风控拦截 + 下单、成交 = 下单 − 失败、
// 落库 = 下单），而不是具体数字 —— 数字由 pnl_test 钉住，这里钉的是「装配没接错」。

package backtest

import (
	"context"
	"strings"
	"testing"

	"github.com/shopspring/decimal"

	"go-study/trading-bot/internal/executor"
	"go-study/trading-bot/internal/market"
)

// 回测交易所必须是 executor 认的那个 Exchange。
var _ executor.Exchange = (*Exchange)(nil)

// 滑点要算在「更差」的方向上：买入上滑、卖出下滑；为 0 时按原价成交。
func TestExchangeSlippage(t *testing.T) {
	tests := []struct {
		name   string
		slipBP string
		side   string
		price  string
		want   string
	}{
		{name: "无滑点买入", slipBP: "0", side: sideBuy, price: "100", want: "100"},
		{name: "无滑点卖出", slipBP: "0", side: sideSell, price: "100", want: "100"},
		{name: "1bp 滑点买入上滑", slipBP: "1", side: sideBuy, price: "100", want: "100.01"},
		{name: "1bp 滑点卖出下滑", slipBP: "1", side: sideSell, price: "100", want: "99.99"},
		{name: "10bp 滑点买入", slipBP: "10", side: sideBuy, price: "63000", want: "63063"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exch := NewExchange(decimal.RequireFromString(tt.slipBP))
			exch.SetPrice(decimal.RequireFromString(tt.price))

			o := &executor.Order{ID: "X", Symbol: "BTCUSDT", Side: tt.side, Qty: decimal.NewFromInt(1)}
			if err := exch.Submit(context.Background(), o); err != nil {
				t.Fatalf("回测交易所不该失败，实际: %v", err)
			}
			if len(exch.Fills) != 1 {
				t.Fatalf("成交笔数 = %d, want 1", len(exch.Fills))
			}
			if got := exch.Fills[0].Price.String(); got != tt.want {
				t.Errorf("成交价 = %s, want %s", got, tt.want)
			}
			if exch.Fills[0].OrderID != "X" || exch.Fills[0].Side != tt.side {
				t.Errorf("成交记录字段不对: %+v", exch.Fills[0])
			}
		})
	}
}

// Runner：把一段小序列喂进管线，检查各项计数自洽。
func TestRunInvariants(t *testing.T) {
	for _, p := range AllPatterns() {
		t.Run(string(p), func(t *testing.T) {
			qs, err := Series(p, 200, 7)
			if err != nil {
				t.Fatalf("Series: %v", err)
			}
			res, err := Run(context.Background(), DefaultConfig(qs, p))
			if err != nil {
				t.Fatalf("Run: %v", err)
			}

			if res.Quotes != len(qs) {
				t.Errorf("行情笔数 = %d, want %d", res.Quotes, len(qs))
			}
			if res.Signals != res.Denied+res.Orders {
				t.Errorf("信号 %d != 风控拦截 %d + 下单 %d", res.Signals, res.Denied, res.Orders)
			}
			if res.Orders != res.Filled+res.Failed {
				t.Errorf("下单 %d != 成交 %d + 失败 %d", res.Orders, res.Filled, res.Failed)
			}
			// 回测交易所永不失败，所以失败恒为 0
			if res.Failed != 0 {
				t.Errorf("回测交易所不该有失败，实际 %d 笔", res.Failed)
			}
			if res.Persisted != res.Orders {
				t.Errorf("落库 %d 条 != 下单 %d 笔（成功下单都要落库）", res.Persisted, res.Orders)
			}
			if res.Stats.Fills != res.Filled {
				t.Errorf("统计里的成交笔数 %d != 成交 %d", res.Stats.Fills, res.Filled)
			}
			if res.Stats.Wins > res.Stats.ClosedTrades {
				t.Errorf("盈利笔数 %d 不该超过平仓笔数 %d", res.Stats.Wins, res.Stats.ClosedTrades)
			}
		})
	}
}

// 同一个输入跑两次，报告必须一字不差 —— 回测结论不可复现就等于没有回测。
func TestRunDeterministic(t *testing.T) {
	qs, err := Series(PatternVReversal, 300, 7)
	if err != nil {
		t.Fatalf("Series: %v", err)
	}
	first, err := Run(context.Background(), DefaultConfig(qs, PatternVReversal))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	second, err := Run(context.Background(), DefaultConfig(qs, PatternVReversal))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if first.Report() != second.Report() {
		t.Errorf("两次回测报告不一致：\n%s\n---\n%s", first.Report(), second.Report())
	}
}

// 入参不合法时明确报错，而不是 panic 或跑出一个空报告。
func TestRunErrors(t *testing.T) {
	t.Run("空序列", func(t *testing.T) {
		_, err := Run(context.Background(), DefaultConfig(nil, PatternRange))
		if err == nil || !strings.Contains(err.Error(), "行情序列为空") {
			t.Errorf("空序列应报错，实际 err=%v", err)
		}
	})

	t.Run("均线窗口非正", func(t *testing.T) {
		qs, _ := Series(PatternRange, 50, 7)
		cfg := DefaultConfig(qs, PatternRange)
		cfg.Window = 0
		if _, err := Run(context.Background(), cfg); err == nil || !strings.Contains(err.Error(), "均线窗口") {
			t.Errorf("窗口非正应报错，实际 err=%v", err)
		}
	})

	t.Run("价格字符串无法解析", func(t *testing.T) {
		bad := []market.Quote{
			{Symbol: "BTCUSDT", Price: 100, PriceStr: "100"},
			{Symbol: "BTCUSDT", Price: 101, PriceStr: "101"},
			{Symbol: "BTCUSDT", Price: 102, PriceStr: "102"},
			{Symbol: "BTCUSDT", Price: 103, PriceStr: "103"},
			{Symbol: "BTCUSDT", Price: 104, PriceStr: "104"},
			{Symbol: "BTCUSDT", Price: 105, PriceStr: "105"},
			{Symbol: "BTCUSDT", Price: 200, PriceStr: "不是数字"},
		}
		cfg := DefaultConfig(bad, PatternRange)
		cfg.Threshold = 0.001
		if _, err := Run(context.Background(), cfg); err == nil || !strings.Contains(err.Error(), "无法解析") {
			t.Errorf("坏价格应报错，实际 err=%v", err)
		}
	})
}

// 报告要一眼看懂：形态名、各项计数、金额保留 2 位小数、胜率 1 位小数。
func TestReport(t *testing.T) {
	qs, err := Series(PatternRange, 500, 7)
	if err != nil {
		t.Fatalf("Series: %v", err)
	}
	res, err := Run(context.Background(), DefaultConfig(qs, PatternRange))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	report := res.Report()

	for _, want := range []string{
		"回测报告",
		"形态：震荡",
		"行情 500 笔",
		"手续费 5 bp/边",
		"信号", "风控拦截", "下单", "成交", "失败",
		"毛利", "手续费", "净利", "胜率",
		"最大回撤", "期末持仓", "USDT", "BTC",
	} {
		if !strings.Contains(report, want) {
			t.Errorf("报告缺少 %q：\n%s", want, report)
		}
	}

	// 金额必须 2 位小数、胜率 1 位小数 —— 直接拿统计值算一遍对得上
	for _, want := range []string{
		res.Stats.Gross.StringFixed(2),
		res.Stats.Fee.StringFixed(2),
		res.Stats.Net.StringFixed(2),
		res.Stats.MaxDrawdown.StringFixed(2),
		res.Stats.WinRate.StringFixed(1),
	} {
		if !strings.Contains(report, want) {
			t.Errorf("报告缺少数值 %q：\n%s", want, report)
		}
	}
}

// 没有成交时也要能出报告（除零不能崩），期末持仓为 0。
func TestReportNoFills(t *testing.T) {
	// 阈值设得极高，任何信号都出不来
	qs, _ := Series(PatternRange, 100, 7)
	cfg := DefaultConfig(qs, PatternRange)
	cfg.Threshold = 10

	res, err := Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Signals != 0 || res.Orders != 0 {
		t.Fatalf("阈值极高时不该有信号，实际 signals=%d orders=%d", res.Signals, res.Orders)
	}
	if !res.Stats.OpenQty.IsZero() || !res.Stats.Net.IsZero() {
		t.Errorf("无成交时持仓/净利应为 0，实际 %s / %s", res.Stats.OpenQty, res.Stats.Net)
	}
	if !strings.Contains(res.Report(), "期末持仓 0 BTC") {
		t.Errorf("无成交的报告不对：\n%s", res.Report())
	}
}

// 风控确实在拦：把名义金额上限压到单笔之下，所有信号都该被拦下、一单不下。
func TestRunRiskGuardBlocks(t *testing.T) {
	qs, _ := Series(PatternUp, 300, 7)
	cfg := DefaultConfig(qs, PatternUp)
	cfg.MaxNotional = decimal.NewFromInt(1) // 0.002 × 63000 远超 1 USDT

	res, err := Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Signals == 0 {
		t.Fatal("这段行情应该有信号，否则这条用例没验证到风控")
	}
	if res.Orders != 0 || res.Denied != res.Signals {
		t.Errorf("名义金额超限时应全部拦下：signals=%d denied=%d orders=%d",
			res.Signals, res.Denied, res.Orders)
	}
}

// 报告里的持仓单位取交易对的基础币：非 USDT 计价、以及只有 USDT 的退化情况都不能空着。
func TestReportSymbolBase(t *testing.T) {
	tests := []struct {
		symbol string
		want   string
	}{
		{"BTCUSDT", "期末持仓 0 BTC"},
		{"ETHBTC", "期末持仓 0 ETHBTC"},
		{"USDT", "期末持仓 0 USDT"}, // TrimSuffix 后为空 → 回落到原符号
	}
	for _, tt := range tests {
		r := Result{Pattern: PatternCSV, Symbol: tt.symbol, FeeBP: decimal.NewFromInt(5)}
		if !strings.Contains(r.Report(), tt.want) {
			t.Errorf("symbol=%q 的报告缺少 %q：\n%s", tt.symbol, tt.want, r.Report())
		}
	}
}
