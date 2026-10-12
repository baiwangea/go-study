// 本文件是 T-05 的基准与正确性测试。
//
// 基准测试在这里的作用不是「比较快慢」（那是 T-04 的事），
// 而是**为 pprof 提供足够长的运行时间**。
//
// 为什么必须够长？
//   CPU profile 是采样得到的，默认每秒 100 个样本。
//   如果一个基准只跑 0.3 秒，总共只有 30 个样本 —— 分到十几个函数上，
//   每个函数可能只有 0~2 个样本，噪声完全盖过信号，pprof 给出的排名就是随机的。
//   所以采 profile 时一定要加 -benchtime=2s 以上，或者 -count=5。

package main

import (
	"testing"
)

// sinkReport 吃掉返回值，防止编译器把整个计算优化掉（复习 T-04）。
var sinkReport string

var benchOrders = MakeOrders(20000)

// ─────────────────────────────────────────────────────────────────────────────
// 这组基准是给 pprof 用的「靶子」。
//
// 采 profile 的命令：
//     go test -bench=. -benchtime=2s -cpuprofile=cpu.out ./05_pprof/
//     go tool pprof -top cpu.out
//     go tool pprof -list=classifySlow cpu.out
//
// 预期：在 -top 里 regexp.(*Regexp).doExecute 之类的函数会排在最前面，
// 而 -list=classifySlow 会把箭头精确指向 regexp.MustCompile 那一行。
// ─────────────────────────────────────────────────────────────────────────────

func BenchmarkBuildReport(b *testing.B) {
	b.Run("slow-循环内编译正则", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sinkReport = buildReportSlow(benchOrders)
		}
	})

	b.Run("fast-包级预编译", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sinkReport = buildReportFast(benchOrders)
		}
	})
}

// BenchmarkClassify 把焦点缩到单个函数上。
//
// 做 pprof 时有个实用技巧：**先写一个只测热点函数的基准**，
// 这样采样样本会集中在那一个函数上，-list 的输出非常干净。
func BenchmarkClassify(b *testing.B) {
	syms := []string{"BTCUSDT", "ETHUSDT", "SOLUSDT", "XRPUSDC", "AVAXETH"}

	b.Run("slow", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = classifySlow(syms[i%len(syms)])
		}
	})

	b.Run("fast", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = classifyFast(syms[i%len(syms)])
		}
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// 正确性测试：优化不能改变行为（复习 T-04 的教训）。
//
// 这一条在本关尤其重要：classifySlow 和 classifyFast 用的是**同一个正则**，
// 所以结果必须逐字节一致。如果哪天有人「顺手」改了正则，
// 这里的测试会立刻拦住 —— 而不是等到线上发现订阅列表解析错了。
// ─────────────────────────────────────────────────────────────────────────────

func TestClassifyAgree(t *testing.T) {
	tests := []struct {
		symbol    string
		wantBase  string
		wantQuote string
	}{
		{"BTCUSDT", "BTC", "USDT"},
		{"ETHUSDC", "ETH", "USDC"},
		{"SOLBTC", "SOL", "BTC"},
		{"AVAXETH", "AVAX", "ETH"},
		{"1000SHIBUSDT", "1000SHIB", "USDT"},
		{"NOTAPAIR", "NOTAPAIR", ""}, // 解析失败时原样返回
		{"", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.symbol, func(t *testing.T) {
			base, quote := classifySlow(tt.symbol)
			if base != tt.wantBase || quote != tt.wantQuote {
				t.Errorf("classifySlow(%q) = (%q, %q), want (%q, %q)",
					tt.symbol, base, quote, tt.wantBase, tt.wantQuote)
			}

			// 关键断言：快版必须和慢版完全一致
			fBase, fQuote := classifyFast(tt.symbol)
			if fBase != base || fQuote != quote {
				t.Errorf("快慢两版结果不一致：\n  slow = (%q, %q)\n  fast = (%q, %q)",
					base, quote, fBase, fQuote)
			}
		})
	}
}

func TestBuildReportAgree(t *testing.T) {
	orders := MakeOrders(500)
	if got, want := buildReportFast(orders), buildReportSlow(orders); got != want {
		t.Errorf("两版报表不一致（长度 %d vs %d）", len(got), len(want))
	}
}
