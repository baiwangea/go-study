// 价格序列与 CSV 解析的测试。
//
// 序列部分钉住两件事：**可复现**（同参数同结果）和**形态正确**（涨的就是涨的、
// 震荡的不跑出区间）；CSV 部分把每条错误路径都走一遍，确保报错明确而不是 panic。

package backtest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shopspring/decimal"

	"go-study/trading-bot/internal/market"
)

// 同一个 (形态, 笔数, 种子) 必须永远得到同一段行情 —— 否则回测结论不可复现。
func TestSeriesDeterministic(t *testing.T) {
	for _, p := range AllPatterns() {
		a, err := Series(p, 200, 7)
		if err != nil {
			t.Fatalf("%s: Series 报错: %v", p, err)
		}
		b, err := Series(p, 200, 7)
		if err != nil {
			t.Fatalf("%s: Series 报错: %v", p, err)
		}
		if len(a) != 200 || len(b) != 200 {
			t.Fatalf("%s: 长度 = %d/%d, want 200/200", p, len(a), len(b))
		}
		for i := range a {
			if a[i].PriceStr != b[i].PriceStr || !a[i].Ts.Equal(b[i].Ts) {
				t.Fatalf("%s: 第 %d 根不一致：%q/%v vs %q/%v",
					p, i, a[i].PriceStr, a[i].Ts, b[i].PriceStr, b[i].Ts)
			}
		}
	}

	// 换种子应该得到不同的行情（否则种子参数是摆设）
	a, _ := Series(PatternRange, 100, 1)
	b, _ := Series(PatternRange, 100, 2)
	if a[50].PriceStr == b[50].PriceStr {
		t.Error("换种子后第 50 根价格相同，种子似乎没起作用")
	}
}

// 形态正确性：每根都必须是合法报价，且整体走向符合形态名。
func TestSeriesShapes(t *testing.T) {
	const n = 500
	for _, p := range AllPatterns() {
		qs, err := Series(p, n, 7)
		if err != nil {
			t.Fatalf("%s: %v", p, err)
		}

		// 每根都要能转成 decimal，且 PriceStr 与 Price 一致（下单用的是 PriceStr）
		var min, max market.Quote
		min, max = qs[0], qs[0]
		for i, q := range qs {
			d, err := decimal.NewFromString(q.PriceStr)
			if err != nil {
				t.Fatalf("%s: 第 %d 根价格 %q 无法解析: %v", p, i, q.PriceStr, err)
			}
			if !d.IsPositive() {
				t.Fatalf("%s: 第 %d 根价格 %s 非正", p, i, d)
			}
			if q.Price != d.InexactFloat64() {
				t.Fatalf("%s: 第 %d 根 Price=%v 与 PriceStr=%q 不一致", p, i, q.Price, q.PriceStr)
			}
			if q.Symbol != DefaultSymbol {
				t.Fatalf("%s: 第 %d 根 Symbol = %q", p, i, q.Symbol)
			}
			if i > 0 && !q.Ts.After(qs[i-1].Ts) {
				t.Fatalf("%s: 第 %d 根时间戳没有递增", p, i)
			}
			if q.Price < min.Price {
				min = q
			}
			if q.Price > max.Price {
				max = q
			}
		}

		first, last := qs[0].Price, qs[n-1].Price
		switch p {
		case PatternUp:
			if last <= first*1.15 {
				t.Errorf("单边上涨：末价 %.2f 未明显高于首价 %.2f", last, first)
			}
		case PatternDown:
			if last >= first*0.85 {
				t.Errorf("单边下跌：末价 %.2f 未明显低于首价 %.2f", last, first)
			}
		case PatternRange:
			// 震荡必须始终待在起始价的 ±6% 区间内
			if max.Price > startPrice*1.06 || min.Price < startPrice*0.94 {
				t.Errorf("震荡：价格跑出区间 [%.0f, %.0f]", min.Price, max.Price)
			}
			if last < first*0.9 || last > first*1.1 {
				t.Errorf("震荡：末价 %.2f 与首价 %.2f 偏离过大，不像区间来回", last, first)
			}
		case PatternVReversal:
			// V 型：最低点必须明显低于首尾，且末价从低点反弹回来
			if min.Price > first*0.85 {
				t.Errorf("V 型：最低价 %.2f 未明显低于首价 %.2f", min.Price, first)
			}
			if last <= min.Price*1.15 {
				t.Errorf("V 型：末价 %.2f 未从最低价 %.2f 反弹", last, min.Price)
			}
		}
	}
}

// 参数非法时明确报错，而不是返回半截序列或 panic。
func TestSeriesErrors(t *testing.T) {
	tests := []struct {
		name    string
		pattern Pattern
		n       int
		wantSub string
	}{
		{"笔数为 0", PatternRange, 0, "必须为正"},
		{"笔数为负", PatternRange, -3, "必须为正"},
		{"未知形态", Pattern("sideways"), 10, "未知形态"},
		{"csv 要走 -csv 参数", PatternCSV, 10, "-csv"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Series(tt.pattern, tt.n, 7)
			if err == nil {
				t.Fatal("期望报错，实际 nil")
			}
			if !strings.Contains(err.Error(), tt.wantSub) {
				t.Errorf("错误 = %q, 期望包含 %q", err, tt.wantSub)
			}
		})
	}
}

// ParsePattern：合法值原样返回，csv 与未知值给出明确指引。
func TestParsePattern(t *testing.T) {
	for _, p := range AllPatterns() {
		got, err := ParsePattern(string(p))
		if err != nil || got != p {
			t.Errorf("ParsePattern(%q) = %q, %v; want %q, nil", p, got, err, p)
		}
	}
	if got, err := ParsePattern("  UP  "); err != nil || got != PatternUp {
		t.Errorf("ParsePattern 应忽略大小写与空白，得到 %q, %v", got, err)
	}
	if _, err := ParsePattern("csv"); err == nil || !strings.Contains(err.Error(), "-csv") {
		t.Errorf("csv 形态应提示改用 -csv，实际 err=%v", err)
	}
	if _, err := ParsePattern("nope"); err == nil || !strings.Contains(err.Error(), "未知形态") {
		t.Errorf("未知形态应报错并列出可选值，实际 err=%v", err)
	}
}

// writeCSV 把内容写进临时文件，返回路径。
func writeCSV(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "quotes.csv")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("写临时 CSV 失败: %v", err)
	}
	return path
}

func TestLoadCSV(t *testing.T) {
	t.Run("正常：带表头、Unix 秒", func(t *testing.T) {
		path := writeCSV(t, "timestamp,price\n1704067200,63000.5\n1704067260,63100.25\n")

		qs, err := LoadCSV(path)
		if err != nil {
			t.Fatalf("LoadCSV 报错: %v", err)
		}
		if len(qs) != 2 {
			t.Fatalf("读到 %d 行, want 2（表头应被跳过）", len(qs))
		}
		// 价格字符串原样保留：不能让浮点过一遍
		if qs[0].PriceStr != "63000.5" || qs[1].PriceStr != "63100.25" {
			t.Errorf("价格 = %q / %q, want 63000.5 / 63100.25", qs[0].PriceStr, qs[1].PriceStr)
		}
		if !qs[0].Ts.Before(qs[1].Ts) {
			t.Errorf("时间戳未递增：%v / %v", qs[0].Ts, qs[1].Ts)
		}
	})

	t.Run("正常：无表头、RFC3339", func(t *testing.T) {
		path := writeCSV(t, "2024-01-01T00:00:00Z,100\n2024-01-01T00:01:00Z,101.5\n")

		qs, err := LoadCSV(path)
		if err != nil {
			t.Fatalf("LoadCSV 报错: %v", err)
		}
		if len(qs) != 2 {
			t.Fatalf("读到 %d 行, want 2", len(qs))
		}
		if qs[0].Price != 100 || qs[1].Price != 101.5 {
			t.Errorf("价格 = %v / %v, want 100 / 101.5", qs[0].Price, qs[1].Price)
		}
	})

	t.Run("缺列", func(t *testing.T) {
		path := writeCSV(t, "1704067200\n1704067260,63100\n")
		_, err := LoadCSV(path)
		if err == nil || !strings.Contains(err.Error(), "列数不足") {
			t.Errorf("缺列应报「列数不足」，实际 err=%v", err)
		}
	})

	t.Run("价格非数字", func(t *testing.T) {
		path := writeCSV(t, "1704067200,abc\n")
		_, err := LoadCSV(path)
		if err == nil || !strings.Contains(err.Error(), "不是数字") {
			t.Errorf("非数字价格应报错，实际 err=%v", err)
		}
		if err != nil && !strings.Contains(err.Error(), "第 1 行") {
			t.Errorf("错误应带行号，实际 err=%v", err)
		}
	})

	t.Run("时间戳非法", func(t *testing.T) {
		path := writeCSV(t, "昨天,63000\n")
		_, err := LoadCSV(path)
		if err == nil || !strings.Contains(err.Error(), "时间戳") {
			t.Errorf("非法时间戳应报错，实际 err=%v", err)
		}
	})

	t.Run("价格非正", func(t *testing.T) {
		path := writeCSV(t, "1704067200,0\n")
		_, err := LoadCSV(path)
		if err == nil || !strings.Contains(err.Error(), "必须为正数") {
			t.Errorf("非正价格应报错，实际 err=%v", err)
		}
	})

	t.Run("文件不存在", func(t *testing.T) {
		_, err := LoadCSV(filepath.Join(t.TempDir(), "nope.csv"))
		if err == nil || !strings.Contains(err.Error(), "打开行情 CSV 失败") {
			t.Errorf("文件不存在应报明确错误，实际 err=%v", err)
		}
	})

	t.Run("空文件", func(t *testing.T) {
		_, err := LoadCSV(writeCSV(t, ""))
		if err == nil || !strings.Contains(err.Error(), "没有任何有效数据行") {
			t.Errorf("空文件应报错，实际 err=%v", err)
		}
	})

	t.Run("只有表头", func(t *testing.T) {
		_, err := LoadCSV(writeCSV(t, "timestamp,price\n"))
		if err == nil || !strings.Contains(err.Error(), "没有任何有效数据行") {
			t.Errorf("只有表头应报错，实际 err=%v", err)
		}
	})
}

// 形态中文名：四种内置形态各有名字，csv 与未知值也不能返回空串。
func TestPatternName(t *testing.T) {
	tests := []struct {
		pattern Pattern
		want    string
	}{
		{PatternUp, "单边上涨"},
		{PatternDown, "单边下跌"},
		{PatternRange, "震荡"},
		{PatternVReversal, "V 型反转"},
		{PatternCSV, "外部 CSV"},
		{Pattern("zzz"), "zzz"}, // 未知值原样返回，报告里不会出现空白
	}
	for _, tt := range tests {
		if got := PatternName(tt.pattern); got != tt.want {
			t.Errorf("PatternName(%q) = %q, want %q", tt.pattern, got, tt.want)
		}
	}
}

// CSV 语法错误（引号没闭合）时给出行号，而不是静默截断。
func TestLoadCSVMalformed(t *testing.T) {
	path := writeCSV(t, "1704067200,\"63000\n")
	_, err := LoadCSV(path)
	if err == nil || !strings.Contains(err.Error(), "读取 CSV 第") {
		t.Errorf("CSV 语法错误应报错并带行号，实际 err=%v", err)
	}
}
