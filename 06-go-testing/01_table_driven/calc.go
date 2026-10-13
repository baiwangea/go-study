/*
═══════════════════════════════════════════════════════════════════════
T-01：表驱动测试（table-driven tests）
═══════════════════════════════════════════════════════════════════════

【核心概念】

	Go 的 testing 包里没有 PHPUnit 的 assert 家族，也没有 JUnit 的 @Test 注解。
	它只提供一件事：**你觉得不对，就自己调 t.Errorf 报错**。

	    PHPUnit :  $this->assertEquals(20.0, $pnl);          ← 框架替你比对
	    JUnit   :  assertEquals(20.0, pnl);                  ← 框架替你比对
	    pytest  :  assert pnl == 20.0                        ← 语言级 assert
	    Go      :  if pnl != 20.0 { t.Errorf("...") }        ← 你自己比对

	「框架不管断言」这个取舍，直接导致 **表驱动测试** 成了 Go 的主流写法：
	把「输入 → 期望输出」写成结构体切片，用一个 for 循环跑完。
	新增用例只需要往表里加一行，而不是复制一整个 @Test 方法。

	    tests := []struct {
	        name string
	        side Side
	        want float64
	    }{
	        {"买入盈利", SideBuy, 20},
	        {"卖出亏损", SideSell, -20},
	    }
	    for _, tt := range tests {
	        got := PnL(tt.side, ...)
	        if got != tt.want {
	            t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
	        }
	    }

	与之配套的两条纪律（本关的重点）：
	  ① t.Errorf 只记失败、继续跑；t.Fatalf 记失败 + 立刻终止当前测试函数。
	     在「用例循环」里必须用 Errorf，否则第一个失败就把后面的用例全吞了。
	  ② 失败信息必须能独立定位问题：哪个用例、拿到什么、期望什么。
	     `t.Errorf("结果不对")` 这种写法在生产仓库里等于没写测试。

【运行方式】

	cd 06-go-testing && go test -v ./01_table_driven/

	常用变体（都值得亲手跑一次）：
	    go test ./01_table_driven/                 # 只打印 PASS/FAIL 一行
	    go test -v ./01_table_driven/              # -v 打印每个测试名
	    go test -run TestPnL ./01_table_driven/    # 只跑名字匹配的测试（支持正则）
	    go test -count=1 ./01_table_driven/        # 禁用结果缓存，强制重跑

	⚠ 注意：go test 默认缓存成功结果。同一个包不改代码再跑一次，输出是 `(cached)`，
	  并不会真的执行。想强制重跑就加 -count=1。

【你会看到】

	✓ TestPnL 用一张 5 行的表覆盖「多/空 × 盈/亏 × 整数/小数数量」
	✓ 每个失败信息都带 name / got / want，不用猜是哪一行出的问题
	✓ assertNear 里的 t.Helper() 让失败行号指向「调用它的那一行」，而不是断言函数内部
	✓ TestNetPnL 演示了 t.Fatalf 的正确用途：前置条件不满足时立刻终止
	✓ TestParseSide 演示「一个测试里同时覆盖正常路径与错误分支」

【思考题】（动手改代码验证，别空想）
 1. 把 TestPnL 循环里的 assertNear 换成一个直接调 t.Fatalf 的版本，运行
    go test -v -run TestPnL ./01_table_driven/
    观察：第一个失败用例之后的用例还执行了吗？
    提示：t.Fatalf 内部调用 runtime.Goexit()，会立刻结束当前测试函数。
 2. 把 assertNear 里的 t.Helper() 删掉，再故意改坏一个 want 值。
    失败信息里的行号变成了 calc_test.go 的哪一行？为什么这个行号毫无信息量？
    对照 PHPUnit：为什么 PHPUnit 的失败信息里能直接显示调用栈？
 3. RoundQty(0.3, 0.1) 在 calc.go 的实现里返回 0.2 而不是 0.3。
    先写一条 `RoundQty(0.3, 0.1) == 0.3` 的用例看它怎么失败，
    再用 fmt.Println(0.3/0.1) 打印中间值。
    为什么 float64 下 0.3/0.1 是 2.9999999999999996？
    （交易所下单量必须对齐 stepSize，这就是真实爆仓级 bug 的来源。）
    提示：修法是给商加一个极小 epsilon，或改用 decimal/整数最小单位。
 4. ParseSide 的错误信息里没有带上原始输入。
    先给 ParseSide("BUY ")（末尾一个空格）加一条 wantErr=true 的用例（当前表里已有），
    然后把 ParseSide 的报错改成带 %q 的输入，再想清楚：
    要不要顺手 TrimSpace？改了之后表里那条用例该改成什么？
    对照 PHP 的 Exception::getMessage()：为什么错误信息里带输入值很重要？

【通关标准】

	能默写表驱动测试的骨架（[]struct{name; in; want} + for + 断言），
	说清 t.Errorf 与 t.Fatalf 的取舍，
	并且知道「失败信息里必须能看出是哪个用例、拿到了什么、期望什么」。

【前置】L1-19              【标签】testing · 表驱动 · t.Errorf/t.Fatalf · t.Helper
═══════════════════════════════════════════════════════════════════════
*/
package calc

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

// ─────────────────────────────────────────────────────────────────────────────
// 本关的被测代码：一个小而真实的「交易计算」库。
// 每一关都围绕交易机器人的真实零件出题，而不是 Foo/Bar。
// ─────────────────────────────────────────────────────────────────────────────

// Side 表示买卖方向。
//
// 用自定义类型而不是裸 string，是 Go 的常见做法：让编译期就能挡住
// `PnL("buy", ...)` 这种笔误。PHP 8.1 的 enum、Java 的 enum 是同一个思路。
type Side int

const (
	// SideBuy 买入（做多）。
	SideBuy Side = iota
	// SideSell 卖出（做空）。
	SideSell
)

// String 让 Side 在 %v / %s 下打印成可读文本。
//
// 这就是 Go 版的「toString()」。fmt 包会优先找这个方法，
// 所以测试失败信息里打印 Side 时不会出现 `0` / `1` 这种天书。
func (s Side) String() string {
	switch s {
	case SideBuy:
		return "BUY"
	case SideSell:
		return "SELL"
	default:
		return fmt.Sprintf("Side(%d)", int(s))
	}
}

// ParseSide 把交易所返回的字符串转成 Side。
//
// ⚠ 本函数故意留了两个「可优化点」，对应思考题 4：
//  1. 错误信息里没有带上原始输入（raw），排障时不知道到底是什么值炸的；
//  2. 没有 TrimSpace，交易所 JSON 里出现 "BUY " 这种带空白的值会直接失败。
func ParseSide(raw string) (Side, error) {
	switch strings.ToUpper(raw) {
	case "BUY":
		return SideBuy, nil
	case "SELL":
		return SideSell, nil
	default:
		return 0, errors.New("unknown side")
	}
}

// PnL 计算一笔已平仓交易的盈亏（不含手续费）。
//
//	买入：(exit - entry) * qty
//	卖出：(entry - exit) * qty
func PnL(side Side, entry, exit, qty float64) float64 {
	if side == SideBuy {
		return (exit - entry) * qty
	}
	return (entry - exit) * qty
}

// Commission 按基点（basis point）计算手续费：1 bp = 万分之一。
//
// 交易所文档写的都是 bp，比如 Binance 现货 taker 是 10 bp（0.1%）。
func Commission(notional, rateBps float64) float64 {
	return notional * rateBps / 10000
}

// RoundQty 把下单数量向下对齐到交易所的 stepSize。
//
// 例如 stepSize=0.5 时，想下 2.7 只能下 2.5。
//
// ⚠ 本函数是浮点数的经典陷阱现场（思考题 3）：
// math.Floor(qty/step) 这一步在 step 不是 2 的幂时会产生误差 ——
// 0.3/0.1 在 float64 里等于 2.9999999999999996，
// Floor 之后变成 2，最终结果 0.2，比用户想下的 0.3 少了整整 0.1。
// 生产代码里这个误差会变成「下单量小于最小交易量」被交易所直接拒单。
func RoundQty(qty, step float64) float64 {
	return math.Floor(qty/step) * step
}

// NetPnL 把盈亏和双边手续费合并成净盈亏。
func NetPnL(side Side, entry, exit, qty, rateBps float64) float64 {
	gross := PnL(side, entry, exit, qty)
	openFee := Commission(entry*qty, rateBps)
	closeFee := Commission(exit*qty, rateBps)
	return gross - openFee - closeFee
}
