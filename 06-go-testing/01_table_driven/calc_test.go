// 本文件是 T-01 的测试代码。文件名必须以 _test.go 结尾，go test 才会编译它。
//
// ⚠ 三条和 PHP/Java 很不一样的规则：
//  1. 测试文件必须和被测代码**在同一个包**里（这里都是 package calc），
//     所以测试可以访问包内未导出的符号 —— 这是 Go 刻意的设计，
//     不需要 PHP 的 ReflectionClass 或 Java 的 package-private 技巧。
//  2. 测试函数必须是 `func TestXxx(t *testing.T)`，Xxx 首字母大写，
//     参数类型严格是 *testing.T。没有注解、没有继承 TestCase。
//  3. `_test.go` 文件不参与正常构建，`go build ./...` 不会把它打进产物。

package calc

import (
	"math"
	"testing"
)

// ─────────────────────────────────────────────────────────────────────────────
// 第一步：写一个断言辅助函数（Go 里叫 helper）。
//
// 这不是「框架」，只是一个普通函数。Go 社区的标准做法就是这样：
// 自己写几行 helper，而不是引入一个断言库。
// ─────────────────────────────────────────────────────────────────────────────

// assertNear 比较两个 float64 是否足够接近。
//
// 为什么不用 got == want？
//
//	因为 IEEE-754 浮点数不能精确表示 0.1、0.2 这类十进制小数。
//	PHPUnit 的 assertEquals($delta, ...)、JUnit 的 assertEquals(expected, actual, delta)
//	存在的原因完全一样，只是 Go 不给你，你自己写。
//
// 三个设计细节，每个都对应一个真实教训：
//  1. t.Helper() —— 声明「我是辅助函数」。加上它之后，测试失败时
//     报告的行号会指向**调用 assertNear 的那一行**，而不是本函数内部。
//     删掉它试试：行号会永远指向下面 t.Errorf 这一行，等于没说。
//  2. name 参数 —— 表驱动测试里必须知道是哪个用例挂了。
//  3. 打印 diff —— 只看 got/want 有时看不出偏差量级，附上差值更快定位。
func assertNear(t *testing.T, name string, got, want float64) {
	t.Helper() // ← 关键：让失败行号上溯到调用者

	const eps = 1e-9
	if math.Abs(got-want) > eps {
		t.Errorf("%s: got %.10f, want %.10f (diff=%.3e)", name, got, want, got-want)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// TestPnL：最经典的表驱动写法。
//
// 对照一下三种语言写同一批用例的代码量：
//
//   PHPUnit：每个用例一个 public function testXxx()，5 个用例 = 5 个方法，
//            每个方法里重复一遍 $this->assertEquals(...)。
//   pytest ：@pytest.mark.parametrize("side,entry,exit,qty,want", [...])
//            已经非常接近表驱动，但表写在装饰器里、类型信息丢失。
//   Go     ：表就是普通的结构体切片 —— 编译器会检查每个字段的类型和数量，
//            少了字段、写错类型都会**编译失败**，而不是运行时才发现。
//
// 最后这点是 Go 表驱动的核心优势：**表本身是被编译器检查的**。
// 用 map[string]struct{...} 也可以，但 map 遍历顺序随机，
// 失败顺序不稳定，调试体验差；所以 Go 社区推荐用切片。
// ─────────────────────────────────────────────────────────────────────────────

func TestPnL(t *testing.T) {
	tests := []struct {
		name  string  // 用例名：失败信息里靠它定位
		side  Side    // 输入
		entry float64 // 输入
		exit  float64 // 输入
		qty   float64 // 输入
		want  float64 // 期望输出
	}{
		{"买入·上涨盈利", SideBuy, 100, 110, 2, 20},
		{"买入·下跌亏损", SideBuy, 100, 90, 2, -20},
		{"卖出·下跌盈利", SideSell, 100, 90, 2, 20},
		{"卖出·上涨亏损", SideSell, 100, 110, 2, -20},
		{"小数数量", SideBuy, 30000, 30500, 0.5, 250},
		{"零数量", SideBuy, 30000, 30500, 0, 0},
	}

	for _, tt := range tests {
		// ── 反面写法（思考题 1 会让你亲手试一次）────────────────────────────
		// 如果这里写成：
		//     if got != tt.want { t.Fatalf("%s: got %v, want %v", tt.name, got, tt.want) }
		// 那么「买入·下跌亏损」之后的 4 个用例**一行都不会执行**，
		// 因为 t.Fatalf 会立刻终止 TestPnL 这个函数。
		// 在用例循环里，永远用 t.Errorf。
		// ────────────────────────────────────────────────────────────────
		got := PnL(tt.side, tt.entry, tt.exit, tt.qty)
		assertNear(t, tt.name, got, tt.want)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// TestCommission：演示「浮点数不要用 == 比较」。
//
// 200 USDT 名义 * 5 bp / 10000 = 0.1，
// 但 0.1 在 float64 里是 0.1000000000000000055511151231257827...
// 所以必须走 assertNear（带 epsilon），不能用 ==。
//
// PHPUnit 里等价写法是 assertEquals(0.1, $fee, 0.000000001)，
// JUnit 是 assertEquals(0.1, fee, 1e-9)。
// ─────────────────────────────────────────────────────────────────────────────

func TestCommission(t *testing.T) {
	tests := []struct {
		name     string
		notional float64
		rateBps  float64
		want     float64
	}{
		{"1000 USDT @ 5bp", 1000, 5, 0.5},
		{"200 USDT @ 5bp", 200, 5, 0.1}, // ← 0.1 需要 epsilon 比较
		{"0 费率（VIP 或活动）", 1000, 0, 0},
		{"100 USDT @ 10bp（taker）", 100, 10, 0.1},
		{"大额 1e6 @ 10bp", 1e6, 10, 1000},
	}

	for _, tt := range tests {
		got := Commission(tt.notional, tt.rateBps)
		assertNear(t, tt.name, got, tt.want)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// TestRoundQty：stepSize 对齐。
//
// ⚠ 注意表里的 step 都是 0.5 或 1 —— 它们是 2 的负幂，二进制可精确表示。
// 换成 0.1 就会踩到 calc.go 注释里写的浮点坑（思考题 3）。
//
// 这是真实世界的教训：**很多「只在生产环境偶发」的 bug，
// 根源是测试数据挑得太干净**。写用例时要主动挑「不整齐」的输入。
// ─────────────────────────────────────────────────────────────────────────────

func TestRoundQty(t *testing.T) {
	tests := []struct {
		name string
		qty  float64
		step float64
		want float64
	}{
		{"整数步长向下取整", 2.7, 1, 2},
		{"半步长向下取整", 2.7, 0.5, 2.5},
		{"已经对齐", 3, 0.5, 3},
		{"零数量", 0, 0.5, 0},
		{"恰好一个步长", 0.5, 0.5, 0.5},
	}

	for _, tt := range tests {
		got := RoundQty(tt.qty, tt.step)
		assertNear(t, tt.name, got, tt.want)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// TestParseSide：一个测试同时覆盖正常路径和错误分支。
//
// 这就是 Go 里「测试错误分支」的标准姿势 —— 在表里加一个 wantErr bool 列，
// 而不是像 JUnit 那样写 @Test(expected = IllegalArgumentException.class)
// 或者像 PHPUnit 那样 expectException()。
//
// 为什么 Go 的写法更好？因为 expectException 只能断言「抛了某种异常」，
// 表驱动的 wantErr 可以继续断言「返回值也不能是垃圾」，
// 而且正常用例和错误用例在同一张表里，一眼就能看出覆盖率。
// ─────────────────────────────────────────────────────────────────────────────

func TestParseSide(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    Side
		wantErr bool
	}{
		{"小写 buy", "buy", SideBuy, false},
		{"大写 SELL", "SELL", SideSell, false},
		{"混合大小写 Buy", "Buy", SideBuy, false},
		{"空字符串", "", 0, true},
		{"不认识的指令", "HOLD", 0, true},
		// ⚠ 这一条记录的是「当前行为」，不是「正确行为」。
		//   思考题 4 会让你给 ParseSide 加 TrimSpace —— 那时要把 wantErr 改成 false。
		{"带尾随空格（当前会失败）", "BUY ", 0, true},
	}

	for _, tt := range tests {
		got, err := ParseSide(tt.raw)

		if tt.wantErr {
			if err == nil {
				// 失败信息里带上 %q 的输入和实际返回值 —— 排障时这两样缺一不可
				t.Errorf("%s: ParseSide(%q) 期望报错，实际返回 %v", tt.name, tt.raw, got)
			}
			continue // 错误分支不关心 got
		}

		if err != nil {
			t.Errorf("%s: ParseSide(%q) 意外报错: %v", tt.name, tt.raw, err)
			continue
		}
		if got != tt.want {
			t.Errorf("%s: ParseSide(%q) = %v, want %v", tt.name, tt.raw, got, tt.want)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// TestNetPnL：演示 t.Fatalf 的**正确**用途。
//
// t.Fatalf 不是「更严厉的 Errorf」，它有明确的适用场景：
// **前置条件不满足，继续往下跑没有任何意义。**
//
// 典型例子就是这里的 ParseSide 失败 —— 如果连方向都解析不出来，
// 后面的盈亏计算必然是一堆无意义的数字。
// 这时候用 Fatalf 立刻退出，比让测试跑完再报 20 个莫名其妙的错误清楚得多。
//
// 对照：PHPUnit 的 $this->fail() / JUnit 的 fail() 在语义上是「主动让测试失败」，
// Go 没有这个概念 —— 你不需要主动失败，你只需要 t.Errorf。
// ─────────────────────────────────────────────────────────────────────────────

func TestNetPnL(t *testing.T) {
	side, err := ParseSide("BUY")
	if err != nil {
		// 前置条件失败 → Fatalf，不再继续
		t.Fatalf("前置条件失败，无法继续测试 NetPnL: %v", err)
	}

	// 买入 2 个 @100，涨到 110 卖出，双边各 5bp 手续费：
	//   毛利 = (110-100)*2 = 20
	//   开仓费 = 100*2*5/10000 = 0.1
	//   平仓费 = 110*2*5/10000 = 0.11
	//   净利 = 20 - 0.1 - 0.11 = 19.79
	got := NetPnL(side, 100, 110, 2, 5)
	assertNear(t, "双边手续费后的净盈亏", got, 19.79)
}

// ─────────────────────────────────────────────────────────────────────────────
// TestSideString：一个容易被忽略的点 —— 测试代码也会影响可观测性。
//
// 如果 Side 没有实现 String()，上面所有失败信息里的 Side 都会打印成 0/1。
// 这类「测试基建」的小事，在 PHP/Java 里由框架的 toString()/__toString() 兜底，
// 在 Go 里要靠你自己实现。
// ─────────────────────────────────────────────────────────────────────────────

func TestSideString(t *testing.T) {
	tests := []struct {
		side Side
		want string
	}{
		{SideBuy, "BUY"},
		{SideSell, "SELL"},
		{Side(99), "Side(99)"}, // 未定义值也要能安全打印，不能 panic
	}

	for _, tt := range tests {
		if got := tt.side.String(); got != tt.want {
			t.Errorf("Side(%d).String() = %q, want %q", int(tt.side), got, tt.want)
		}
	}
}
