/*
═══════════════════════════════════════════════════════════════════════
T-06：模糊测试（fuzzing）—— 让机器替你找边界
═══════════════════════════════════════════════════════════════════════

【核心概念】

	表驱动测试只能覆盖「你想得到的输入」。fuzzing 覆盖「你想不到的」。

	    func FuzzParseSymbol(f *testing.F) {
	        f.Add("BTCUSDT")                       // 种子语料：给模糊器一个起点
	        f.Fuzz(func(t *testing.T, raw string) { // ← raw 由模糊器生成
	            // 这里放「不变量」断言，而不是「具体期望值」
	        })
	    }

	Go 1.18 起 fuzzing 是标准库的一部分。它的工作方式：
	  ① 从种子语料出发，做变异（改字节、插字符、删字符、拼接）
	  ② 执行目标函数，如果**崩溃**或**断言失败**，就把输入写进 testdata/fuzz/FuzzXxx/
	  ③ 从语料库里挑选「能覆盖新代码路径」的输入继续变异（覆盖率引导）
	  ④ 找到的失败输入会被永久保存，之后每次 go test 都会重放它

	⚠ 三个关键认知：
	  ① **fuzz 目标和普通测试的断言方式完全不同。**
	     普通测试说「输入 X 应该得到 Y」；fuzz 说「对**任何**输入，
	     都不该 panic、且不变量 P 必须成立」。写不出不变量就用不了 fuzzing。
	  ② **fuzzing 只跑一段时间，不是穷举。** 默认 -fuzz 跑 10 秒左右就停，
	     它保证的是「找到过就是真 bug」，不保证「没找到就是没 bug」。
	  ③ **崩溃（panic）是最容易发现的**，因为模糊器一撞就停。
	     不变量违反需要你先写出正确的断言，价值更高但也更难写。

	对照其他语言：
	    PHP  : 没有内建 fuzzing，要靠第三方库，且解释型语言「崩溃」的形态不统一
	    Java : Jazzer 做得很好，但要 JVM 启动、配置复杂
	    Go   : 内建 `go test -fuzz`，零依赖，找到的用例自动落盘并进版本库

【运行方式】

	cd 06-go-testing

	# 普通跑：只重放语料库里已有的用例（秒级，CI 里就跑这个）
	go test -v ./06_fuzz/

	# 真模糊：持续生成新输入（默认 10s，可以 -fuzztime 加长）
	go test -run=^$ -fuzz=FuzzParseSymbol -fuzztime=20s ./06_fuzz/
	go test -run=^$ -fuzz=FuzzParseSymbol -fuzztime=100000x ./06_fuzz/   # 按次数

	# 看语料库里都存了什么
	ls 06_fuzz/testdata/fuzz/FuzzParseSymbol/

【你会看到】

	本关的 ParseSymbol 是**修好之后**的版本，所以 fuzz 跑不出问题 —— 这是刻意的：
	go test ./... 必须一直是绿的。真正的 bug 被「钉」在测试文件里当展品。

	go test -v ./06_fuzz/ 会看到：

	    --- PASS: TestParseSymbol
	    --- PASS: TestFuzzFoundTheseTwoBugs/bug1_一个空格导致越界panic
	    --- PASS: TestFuzzFoundTheseTwoBugs/bug2_USDT_返回空base不报错
	    --- PASS: FuzzParseSymbol
	    --- PASS: FuzzNormalize

	模糊器找到的三个 bug（都在 1 秒内找到）：

	    bug 1 —— panic: index out of range
	        输入：" "（一个空格）
	        原因：先判 raw == "" 再 TrimSpace，TrimSpace 之后 raw 变成空串，
	              紧接着 raw[0] 直接越界。**判断顺序错了。**

	    bug 2 —— 不变量违反（不崩溃，但结果错）
	        输入："USDT"
	        原因：无分隔符分支用 HasSuffix 匹配到计价单位后，
	              raw[:len(raw)-len(q)] 得到空串，却没报错。
	              下游拿到 base="" 会去查一个不存在的币种。

	    bug 3 —— 不变量违反（写本关时 fuzz 现场找到的）
	        输入："0//"
	        原因：带分隔符分支只检查了「两侧非空」，没校验 quote 是不是已知计价单位，
	              于是返回了 ("0", "/")。
	        ⚠ 注意它是在 bug 1、bug 2 都修好**之后**才被发现的 ——
	          前两个 bug 让模糊器早早失败退出，根本没机会探索到这条路径。
	          **修一个，它继续找下一个**，这就是 fuzzing 的工作节奏。

	找到 bug 3 之后，模糊器自动把最小复现用例写进了仓库：

	    06_fuzz/testdata/fuzz/FuzzParseSymbol/2117ecd131c4ac8b
	    内容：go test fuzz v1
	          string("0//")

	这个文件**要提交进版本库** —— 之后每次 go test 都会重放它，
	保证 bug 3 永远不会悄悄回归。

【思考题】（动手改代码验证，别空想）

 1. 把 ParseSymbol 改回有 bug 的版本（测试文件里 TestFuzzFoundTheseTwoBugs
    已经把那两行原样写出来了，照抄即可），然后跑：
    go test -run=^$ -fuzz=FuzzParseSymbol -fuzztime=20s ./06_fuzz/
    看它多快能撞出 panic。找到之后：
    cat 06_fuzz/testdata/fuzz/FuzzParseSymbol/*
    这就是模糊器替你写的「最小复现用例」—— 它会永久留在仓库里。

 2. **为什么 bug 1 用表驱动测试很难发现？**
    想清楚：你写表的时候会不会想到「一个空格」这个输入？
    再想：bug 2 呢？这个更难，因为「USDT」看起来像个正常输入。
    这就是 fuzzing 的核心价值：它不受你想象力的限制。

 3. 给 FuzzParseSymbol 加一条新不变量：
    「解析成功后，base + quote 必须等于去空格后的原输入」。
    然后**故意把 ParseSymbol 的 TrimSpace 删掉**再跑 fuzz，
    看它能不能找到违反这条不变量的输入。
    （提示：输入 " BTCUSDT " 会同时暴露 TrimSpace 缺失和往返不一致。）

 4. fuzzing 的局限在哪？
    想三个 FuzzParseSymbol 永远发现不了的问题类型。
    （提示：性能退化、并发竞态、依赖外部状态的逻辑、业务语义错误 ——
    fuzz 只验证你写出来的不变量。）

【通关标准】

	能写出一个带 f.Fuzz 的模糊目标并解释 f.Add 种子的作用，
	能说清「不变量」和「具体期望值」的区别，
	并且知道「fuzz 找到的用例会自动落盘、要提交进版本库」。

【前置】T-01（表驱动测试）   【标签】fuzzing · 不变量 · 边界 · 崩溃复现
═══════════════════════════════════════════════════════════════════════
*/
package symbol

import (
	"errors"
	"fmt"
	"strings"
)

// ─────────────────────────────────────────────────────────────────────────────
// 被测代码：交易对解析。
//
// 机器人启动时要解析订阅列表，"BTC/USDT" 和 "BTCUSDT" 两种写法都要支持，
// 结果会用来查币种精度、最小下单量等元数据。
// 解析错了 → 查不到元数据 → 用默认精度下单 → 被交易所拒单。
//
// ⚠ 下面这版是**修好之后**的。两个原始 bug 的成因写在注释里，
//   可运行的「展品」在 symbol_test.go 的 TestFuzzFoundTheseTwoBugs。
// ─────────────────────────────────────────────────────────────────────────────

// knownQuotes 是支持的计价单位。
//
// 顺序有意义：必须先匹配更长的，否则 "USDC" 会被 "USDT" 之类的短后缀先吃掉。
// 这里恰好没有互为后缀的组合，但真实项目里加币种时要留意这点。
var knownQuotes = []string{"USDT", "USDC", "BUSD", "BTC", "ETH"}

// ErrEmptySymbol 表示输入为空。
var ErrEmptySymbol = errors.New("交易对为空")

// ParseSymbol 把交易对解析成 (基础币种, 计价单位)。
//
// 支持两种写法：
//
//	"BTC/USDT" → ("BTC", "USDT")     带分隔符
//	"BTCUSDT"  → ("BTC", "USDT")     无分隔符，按已知计价单位后缀匹配
//
// 解析失败返回 error，绝不会返回空的 base ——
// 这条保证是模糊测试逼出来的（见下面 bug 2 的说明）。
func ParseSymbol(raw string) (base, quote string, err error) {
	// ── 修复点 1：先 TrimSpace，再判空 ────────────────────────────────
	//
	// 原始 bug：顺序反了。
	//     if raw == "" { return ..., ErrEmptySymbol }   // ← 先判空
	//     raw = strings.TrimSpace(raw)                  // ← 这里可能变成 ""
	//     if raw[0] >= '0' && raw[0] <= '9' { ... }     // ← panic: index out of range
	//
	// 输入 " "（单个空格）就会炸。模糊器在 1 秒内就找到了它。
	// 教训：**任何「先判空、后做可能改变长度的变换」的代码都要重新审视。**
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", ErrEmptySymbol
	}

	// ── 分支 1：带分隔符 ─────────────────────────────────────────────
	if i := strings.IndexByte(raw, '/'); i >= 0 {
		base, quote = raw[:i], raw[i+1:]
		if base == "" || quote == "" {
			return "", "", fmt.Errorf("交易对 %q 的分隔符两侧不能为空", raw)
		}

		// ── 修复点 3：计价单位也要校验 ───────────────────────────────
		//
		// 这是**写本关的时候 fuzz 现场找出来的第三个 bug**，不是我事后编的：
		//
		//     输入 "0//"  →  ParseSymbol 返回 ("0", "/")   ← quote 是个斜杠！
		//
		// 原因：分支 1 只检查了「两侧非空」，却没检查 quote 是不是已知计价单位。
		// 它和 bug 2 是同一类 —— 不 panic、不报错，只是结果毫无意义。
		//
		// 而且这个 bug 是**在修好 bug 1 和 bug 2 之后**才被发现的：
		// 前两个 bug 让模糊器早早失败退出，根本没机会探索到这条路径。
		// 这是 fuzzing 的典型工作方式：**修一个，它继续找下一个。**
		if !isKnownQuote(quote) {
			return "", "", fmt.Errorf("不认识的计价单位: %q", quote)
		}
		return base, quote, nil
	}

	// ── 分支 2：无分隔符，按后缀匹配计价单位 ─────────────────────────
	for _, q := range knownQuotes {
		if !strings.HasSuffix(raw, q) {
			continue
		}
		b := raw[:len(raw)-len(q)]

		// ── 修复点 2：base 不能为空 ─────────────────────────────────
		//
		// 原始 bug：这里直接 return b, q, nil，没有检查 b。
		// 输入 "USDT" 会返回 ("", "USDT", nil) —— 不崩溃、不报错，
		// 但 base 是空串，下游拿它去查币种元数据必然查不到。
		//
		// 这个 bug 比 bug 1 更危险：**崩溃会被监控告警抓住，
		// 静默的错误结果会一路流到下单环节才爆。**
		if b == "" {
			return "", "", fmt.Errorf("交易对 %q 只有计价单位，缺少基础币种", raw)
		}
		return b, q, nil
	}

	return "", "", fmt.Errorf("无法识别计价单位: %q", raw)
}

// isKnownQuote 判断 s 是否是已知的计价单位。
func isKnownQuote(s string) bool {
	for _, q := range knownQuotes {
		if s == q {
			return true
		}
	}
	return false
}

// Normalize 把交易对统一成 "BASE/QUOTE" 的大写形式。
//
// 用于把用户输入、配置文件、交易所返回的三种写法归一。
func Normalize(raw string) (string, error) {
	base, quote, err := ParseSymbol(raw)
	if err != nil {
		return "", err
	}
	return strings.ToUpper(base) + "/" + strings.ToUpper(quote), nil
}
