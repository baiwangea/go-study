/*
═══════════════════════════════════════════════════════════════════════
T-04：基准测试（benchmark）与优化对比
═══════════════════════════════════════════════════════════════════════

【核心概念】

	单元测试回答「对不对」，基准测试回答「快不快、费不费内存」。

	    func BenchmarkXxx(b *testing.B) {
	        for i := 0; i < b.N; i++ {     // ← b.N 由测试框架自动调整
	            _ = 被测函数()
	        }
	    }

	b.N 不是你能定的：go test 先给一个小 N 试跑，测出单次耗时，
	再不断放大 N，直到总耗时达到稳定区间（默认 1 秒）。
	所以基准测试的循环体里**不能有随 N 变化的副作用**，否则结果毫无意义。

	四条必守纪律：
	  ① 循环前的准备工作要用 b.ResetTimer() 排除掉，
	     否则「造数据」的耗时会被算进「被测函数」的耗时里。
	  ② 想知道分配了多少次、多少字节，调 b.ReportAllocs()（或命令行加 -benchmem）。
	     **内存分配次数往往比耗时更能说明问题** —— 每次分配都是一次 GC 压力。
	  ③ 防止编译器把没被使用的计算结果优化掉：把结果赋给一个包级变量，
	     或者用 b.SetBytes / runtime.KeepAlive。Go 的编译器很聪明，
	     `_ = f()` 里 f 若没有副作用，整个调用可能被删掉，基准就白跑了。
	  ④ 用子基准（b.Run）把多个实现放在同一次运行里对比，
	     避免「两次 go test 之间机器负载不同」导致的误差。

	对照其他语言：
	    PHP  : 没有内建基准测试，靠 phpbench 这类第三方工具，且 PHP 的 JIT/OPcache 会干扰
	    Java : JMH 是事实标准，但要处理 JIT 预热、死代码消除，配置相当重
	    Go   : 内建 `go test -bench`，自动调 N、自动排除准备时间，最省心

【运行方式】

	cd 06-go-testing && go test -bench=. -benchmem ./04_benchmark/

	一定要亲手试这几个：
	    go test -bench=. -benchmem ./04_benchmark/        # 全部基准 + 内存分配
	    go test -bench=Builder -benchmem ./04_benchmark/  # 只跑名字匹配的
	    go test -bench=. -benchtime=3s ./04_benchmark/    # 每组至少跑 3 秒，更稳
	    go test -bench=. -count=5 ./04_benchmark/         # 跑 5 轮，看波动有多大

	⚠ -bench=. 会真的消耗 CPU。跑的时候别同时干别的事，否则数字会飘。

【你会看到】

	本机（Apple Silicon / go1.27）实测。**你机器上的绝对值一定不同，重点是倍数关系和分配次数。**

	规模 = 8 个交易对：
	    BenchmarkSymbols/slow-字符串拼接-8        163.8 ns/op     296 B/op     7 allocs/op
	    BenchmarkSymbols/fast-strings.Builder-8    92.0 ns/op     120 B/op     4 allocs/op
	    → 只快 1.8 倍

	    BenchmarkOrderKey/sprintf-8               104.3 ns/op      24 B/op     1 allocs/op
	    BenchmarkOrderKey/strconv+Builder-8        64.6 ns/op      64 B/op     2 allocs/op
	    → 只快 1.6 倍

	规模 = 10000 个交易对：
	    BenchmarkWithSetup/slow-字符串拼接/10000个-8      43 034 063 ns/op   350 008 433 B/op   10023 allocs/op
	    BenchmarkWithSetup/fast-strings.Builder/10000个-8     64 929 ns/op       285 434 B/op      22 allocs/op
	    → 快 **663 倍**，少分配 **1226 倍**内存

	⚠ 注意这两组数字的巨大反差，这是本关第一个反直觉的地方：
	  我原本以为 8 个元素时 slow 就会慢 6~7 倍，**实测只有 1.8 倍**。
	  原因：规模太小时常数项占主导，O(n²) 的二次项还没显形。
	  到 10000 个元素，慢的那份要拷贝 1+2+…+10000 次，量级直接炸开。

	**结论：基准的输入规模必须贴近真实场景，否则你会得出完全错误的优化结论。**
	这也是「先测量、再优化」这条铁律的由来 —— 直觉在这里几乎总是错的。

【思考题】（动手改代码验证，别空想）

 1. 把 SymbolKeySlow 里的 `out += s` 换成 `out = out + s + ","`，再跑基准。
    耗时和 allocs 各变了多少？为什么写法上只多一个逗号，编译出来的分配次数会不同？
    然后想想：**为什么 allocs/op 比 ns/op 更值得盯**？
    （提示：每次分配最终都要 GC 回收，高频路径上这就是延迟毛刺的来源。）

 2. 给 SymbolKeyFast 里的 strings.Builder 加上
    `b.Grow(预估长度)`，再跑一次基准。
    分配次数变了吗？耗时变了吗？—— 这个优化值不值得加，取决于什么？

 3. 把 BenchmarkSymbols 里 `sink = ...` 那行删掉，改成 `_ = SymbolKeyFast(syms)`，
    再跑基准。数字有变化吗？为什么？
    （提示：Go 编译器的死代码消除。这就是纪律 ③ 存在的原因。）

 4. 用 -count=5 跑 5 轮，观察同一组基准的波动范围。
    如果两个实现的差距只有 5%，而波动有 10%，你能得出结论说「A 比 B 快」吗？
    生产项目里该怎么决定「这个优化值不值得做」？

 5. **本关最重要的一题。** 把 OrderKeyFast 里的 'f', 8 改回 'f', -1，
    再跑 go test ./04_benchmark/（注意是测试，不是基准）。
    看 TestOrderKeyAgree 怎么失败的，然后回答：
    · 为什么基准测试完全发现不了这个问题？（提示：基准只看快慢，不看结果对不对）
    · 如果这是生产代码，后果是什么？（提示：订单键用于 Redis 幂等去重）
    · 你觉得「先写测试再优化」和「先优化再补测试」，哪个顺序更容易出事故？

【通关标准】

	能默写基准测试骨架并说清 b.N / b.ResetTimer / b.ReportAllocs 的作用，
	能用子基准做「优化前后对比」并读懂 -benchmem 输出，
	并且知道「先测量、再优化，不要凭直觉改代码」。

【前置】T-01（表驱动测试）、L1-23（strings.Builder）   【标签】benchmark · 性能 · 内存分配
═══════════════════════════════════════════════════════════════════════
*/
package perf

import (
	"fmt"
	"strconv"
	"strings"
)

// ─────────────────────────────────────────────────────────────────────────────
// 被测代码：交易机器人里两个真实的高频操作。
//
// 为什么挑这两个？
//   · 订阅列表、日志行、报表 —— 字符串拼接在机器人里随处可见，而且常在
//     「每个 tick 都跑一遍」的路径上。
//   · 订单键（order key）用于 Redis 幂等去重，下单前必算一次。
//   这两处都是「改一行代码就能快好几倍」的典型场景。
// ─────────────────────────────────────────────────────────────────────────────

// SymbolKeySlow 用 += 拼接交易对列表。
//
// ⚠ 这是**反面教材**，思考题 1 会让你改它。
//
// 为什么慢？Go 的 string 是不可变的。`out += s` 每次都会：
//
//	① 申请一块「旧内容 + 新内容」的新内存
//	② 把旧内容拷过去
//	③ 把 s 拷过去
//
// 所以循环 n 次的拷贝量是 1+2+...+n = O(n²)。
// PHP 的 `.=` 因为有写时复制和原地扩容，通常不会这么惨 ——
// 这就是「从 PHP 过来的人容易在这里踩坑」的原因。
func SymbolKeySlow(syms []string) string {
	out := ""
	for _, s := range syms {
		out += s
	}
	return out
}

// SymbolKeyFast 用 strings.Builder 拼接。
//
// strings.Builder 内部是一个 []byte，Append 走的是「按需翻倍扩容」，
// 摊还成本 O(1)，整体 O(n)。它还有个额外的优化：String() 时
// 用 unsafe 把内部 []byte 直接转成 string，**不再复制一次**。
//
// 为什么不用 bytes.Buffer？因为 Buffer 的 String() 会拷贝一份，
// 而 Builder 不会。Builder 就是为「拼完就要一个 string」这个场景设计的。
func SymbolKeyFast(syms []string) string {
	var b strings.Builder
	for _, s := range syms {
		b.WriteString(s)
	}
	return b.String()
}

// ─────────────────────────────────────────────────────────────────────────────
// 第二组对比：格式化一个订单键。
//
// ⚠ 这一组藏着一个**真实踩到的坑**，是本关最值钱的一段。
//
// 第一版我写的是：
//     Sprintf 版：fmt.Sprintf("%s:%s:%v", symbol, side, qty)
//     Fast    版：strconv.FormatFloat(qty, 'f', -1, 64)
// 两者对 qty=0.5 输出完全一样，基准跑出来「快了 2.7 倍」，看起来很成功。
// 但下面那个正确性测试 TestOrderKeyAgree 立刻红了：
//
//     qty = 0.000001 时
//       %v                          → "SOLUSDT:BUY:1e-06"      ← 科学计数法！
//       FormatFloat(qty,'f',-1,64)  → "SOLUSDT:BUY:0.000001"
//
// 两个订单键不一样。在交易机器人里这意味着什么？
// 订单键是用来做**幂等去重**的（同一笔单只能提交一次）。
// 键不一致 → Redis 里查不到上一次的记录 → 同一笔单被提交两遍 → 重复下单。
//
// 所以这里的正确做法是**显式指定精度**，两边都用固定 8 位小数：
//     Sprintf 版：%.8f
//     Fast    版：strconv.FormatFloat(qty, 'f', 8, 64)
//
// 这就是 T-04 最核心的一课：**优化时最容易引入的不是性能问题，而是行为差异。**
// 没有正确性测试兜底的「优化」，只是把 bug 变得更高效。
// ─────────────────────────────────────────────────────────────────────────────

// OrderKeySprintf 用 fmt.Sprintf 拼订单键。
//
// fmt 包走的是反射 + 接口断言，功能强大但有固定开销：
// 每次调用都要解析格式串、把参数装箱成 []any。
func OrderKeySprintf(symbol, side string, qty float64) string {
	return fmt.Sprintf("%s:%s:%.8f", symbol, side, qty)
}

// OrderKeyFast 用 strconv + Builder 手工拼。
//
// 代价是代码可读性下降 —— 所以**只在真正的高频路径上这么写**。
// 判断标准是：这个函数每秒被调用多少次？
// 每秒 10 次的函数，Sprintf 和手工拼的差别你根本感觉不到。
func OrderKeyFast(symbol, side string, qty float64) string {
	var b strings.Builder
	b.Grow(len(symbol) + len(side) + 32)
	b.WriteString(symbol)
	b.WriteByte(':')
	b.WriteString(side)
	b.WriteByte(':')
	// 注意这里必须是 'f', 8 —— 和上面的 %.8f 完全对应。
	// 写成 'f', -1（最短表示）就会在极小值上和 %v / %.8f 分道扬镳。
	b.WriteString(strconv.FormatFloat(qty, 'f', 8, 64))
	return b.String()
}

// ─────────────────────────────────────────────────────────────────────────────
// 供基准测试使用的数据。
//
// 注意：这些是包级变量，而不是在基准函数里临时构造的。
// 原因见 benchmark 文件里的 b.ResetTimer 说明。
// ─────────────────────────────────────────────────────────────────────────────

// DefaultSymbols 是模拟「订阅 8 个交易对」的列表。
var DefaultSymbols = []string{
	"BTCUSDT", "ETHUSDT", "SOLUSDT", "BNBUSDT",
	"XRPUSDT", "ADAUSDT", "DOGEUSDT", "AVAXUSDT",
}
