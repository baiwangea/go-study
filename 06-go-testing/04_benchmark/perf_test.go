// 本文件是 T-04 的基准测试。
//
// 先读 perf.go 顶部的关卡卡（尤其是四条纪律），再回来看代码。
//
// 一条贯穿全篇的原则：**基准测试本身也要被「测试」**。
// 下面每个 Benchmark 都配了一个普通 Test 断言两种实现结果一致 ——
// 否则你优化完「快了 7 倍」但结果算错了，基准跑得再漂亮也没意义。

package perf

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

// sink 是一个包级变量，专门用来「吃掉」基准测试的计算结果。
//
// 为什么必须这么做？因为 Go 编译器会做死代码消除：
// 如果 `SymbolKeyFast(syms)` 的返回值没人用，且编译器能证明它没有副作用，
// 整个调用会被删掉，基准循环变成一个空循环 —— 跑出「0.3 ns/op」这种假数据。
// 把结果写进包级变量，编译器就无法证明它是死代码了。
//
// 同理，b.SetBytes() / runtime.KeepAlive() 也能达到目的，
// 但「赋值给包级 sink」是最简单、最不容易忘的做法。
var sink string

// ─────────────────────────────────────────────────────────────────────────────
// 子基准（sub-benchmark）：把多个实现放进同一次运行里对比。
//
// 为什么强调「同一次运行」？因为两次独立运行之间，
// CPU 频率、后台进程、散热状态都可能不同，误差轻易超过 10%。
// 而两个实现的真实差距可能只有 20% —— 分两次跑就可能得出错误结论。
//
// b.Run(name, fn) 和 t.Run 的用法完全一样（复习 T-02），
// 生成的基准名是 BenchmarkSymbols/slow 这种形式。
// ─────────────────────────────────────────────────────────────────────────────

func BenchmarkSymbols(b *testing.B) {
	syms := DefaultSymbols

	b.Run("slow-字符串拼接", func(b *testing.B) {
		b.ReportAllocs() // ← 报告分配次数和字节数，比耗时更能说明问题
		for i := 0; i < b.N; i++ {
			sink = SymbolKeySlow(syms)
		}
	})

	b.Run("fast-strings.Builder", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sink = SymbolKeyFast(syms)
		}
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// 第二组：格式化开销。
//
// 注意这里每个循环体只做一次格式化，所以 ns/op 很小（几十到几百纳秒）。
// 这种量级下，**单次调用的差别对业务毫无意义**；
// 只有当这个函数每秒被调用几十万次时，差距才会变成真实的 CPU 占用。
// 这就是「先测量、再优化」的含义：不测量，你根本不知道热点在哪。
// ─────────────────────────────────────────────────────────────────────────────

func BenchmarkOrderKey(b *testing.B) {
	b.Run("sprintf", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sink = OrderKeySprintf("BTCUSDT", "BUY", 0.5)
		}
	})

	b.Run("strconv+Builder", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sink = OrderKeyFast("BTCUSDT", "BUY", 0.5)
		}
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// b.ResetTimer：排除「准备工作」的耗时。
//
// 下面这个基准故意先做一堆准备（构造 10000 个元素的切片），
// 然后调 b.ResetTimer() 把这段耗时从统计里剔除。
//
// 试一下：把 b.ResetTimer() 注释掉再跑，ns/op 会显著变大 ——
// 那多出来的部分全是「造数据」，跟被测函数一点关系都没有。
// 这是新手写基准时最常见的一个错误。
//
// 更重要的是：**这里才看得见「O(n²)」的真面目。**
// 8 个元素时 slow 只比 fast 慢不到 2 倍（因为规模太小，常数项占主导）；
// 到 10000 个元素时，slow 的拷贝量是 1+2+…+10000 次，量级彻底炸开。
// 这就是「小数据量的基准会骗人」的经典案例 —— 基准的输入规模必须贴近真实场景。
// ─────────────────────────────────────────────────────────────────────────────

func BenchmarkWithSetup(b *testing.B) {
	// ── 准备阶段：构造 10000 个交易对 ────────────────────────────────
	syms := make([]string, 10000)
	for i := range syms {
		syms[i] = "SYM" + strings.Repeat("X", i%8)
	}

	b.ResetTimer() // ← 上面那段耗时不参与统计

	b.Run("slow-字符串拼接/10000个", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sink = SymbolKeySlow(syms)
		}
	})

	b.Run("fast-strings.Builder/10000个", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sink = SymbolKeyFast(syms)
		}
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// 正确性测试：优化不能改变行为。
//
// 这一节本身没什么新知识，但它是整关的地基 ——
// **没有正确性保证的「优化」只是把 bug 变得更高效。**
// ─────────────────────────────────────────────────────────────────────────────

func TestImplementationsAgree(t *testing.T) {
	tests := []struct {
		name string
		syms []string
	}{
		{"空列表", nil},
		{"单个", []string{"BTCUSDT"}},
		{"默认八个", DefaultSymbols},
		{"含空串", []string{"", "BTCUSDT", ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slow := SymbolKeySlow(tt.syms)
			fast := SymbolKeyFast(tt.syms)
			if slow != fast {
				t.Errorf("两种实现结果不一致:\n  slow = %q\n  fast = %q", slow, fast)
			}
		})
	}
}

func TestOrderKeyAgree(t *testing.T) {
	cases := []struct {
		symbol, side string
		qty          float64
	}{
		{"BTCUSDT", "BUY", 0.5},
		{"ETHUSDT", "SELL", 1},
		{"SOLUSDT", "BUY", 0.000001}, // ← 就是这个值让第一版实现露了馅
		{"BTCUSDT", "BUY", 12345.6789},
	}

	for _, c := range cases {
		t.Run(c.symbol+"/"+c.side, func(t *testing.T) {
			slow := OrderKeySprintf(c.symbol, c.side, c.qty)
			fast := OrderKeyFast(c.symbol, c.side, c.qty)
			if slow != fast {
				t.Errorf("两种实现结果不一致:\n  sprintf = %q\n  fast    = %q", slow, fast)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// TestFormatTrap：把本关踩到的那个坑钉死在测试里。
//
// 这是 T-04 最值钱的一段代码 —— 它解释了一个「性能优化变成生产事故」的完整链条。
//
// 第一版实现：
//
//	Sprintf 版：fmt.Sprintf("%s:%s:%v", symbol, side, qty)
//	Fast    版：strconv.FormatFloat(qty, 'f', -1, 64)
//
// 对 qty=0.5 两者输出一样，基准显示「快了 2.7 倍」。
// 但 qty=0.000001 时：
//
//	%v                          → "1e-06"      ← 科学计数法
//	FormatFloat(qty,'f',-1,64)  → "0.000001"
//
// 订单键不一样 → Redis 里查不到上一次的记录 → 幂等失效 → **同一笔单提交两遍**。
//
// 关键教训：**基准测试永远发现不了这个 bug。**
// 基准只看快慢，不看结果对不对。能拦住它的只有正确性测试。
// 所以「优化」和「测试」必须绑在一起做 —— 先有测试，再动性能。
// ─────────────────────────────────────────────────────────────────────────────

func TestFormatTrap(t *testing.T) {
	t.Run("%v 对极小值会输出科学计数法", func(t *testing.T) {
		got := fmt.Sprintf("%v", 0.000001)
		if got != "1e-06" {
			t.Fatalf("fmt %%v 的输出变了：got %q，本关的说明需要更新", got)
		}
		t.Logf("%%v                  → %q   ← 科学计数法，不适合做订单键", got)
	})

	t.Run("'f',-1 输出最短十进制", func(t *testing.T) {
		got := strconv.FormatFloat(0.000001, 'f', -1, 64)
		if got != "0.000001" {
			t.Fatalf("FormatFloat 的输出变了：got %q", got)
		}
		t.Logf("FormatFloat('f',-1) → %q", got)
	})

	t.Run("两种写法产生不同的订单键", func(t *testing.T) {
		a := fmt.Sprintf("SOLUSDT:BUY:%v", 0.000001)
		b := "SOLUSDT:BUY:" + strconv.FormatFloat(0.000001, 'f', -1, 64)
		if a == b {
			t.Fatal("两者竟然相等了 —— 那本关的说明需要更新")
		}
		t.Logf("同一个订单，两个不同的键：\n    %q\n    %q", a, b)
		t.Logf("这就是「幂等去重失效」的起点：查不到旧键 → 重复下单")
	})

	t.Run("固定精度才是订单键该用的写法", func(t *testing.T) {
		// 现在两个实现都用固定 8 位小数，对任意输入都一致
		if got, want := OrderKeyFast("SOLUSDT", "BUY", 0.000001), "SOLUSDT:BUY:0.00000100"; got != want {
			t.Errorf("OrderKeyFast = %q, want %q", got, want)
		}
		if OrderKeySprintf("SOLUSDT", "BUY", 0.000001) != OrderKeyFast("SOLUSDT", "BUY", 0.000001) {
			t.Error("固定精度下两种实现必须完全一致")
		}
	})
}
