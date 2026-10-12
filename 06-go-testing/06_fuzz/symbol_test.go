// 本文件是 T-06 的模糊测试。
//
// 结构：
//   TestParseSymbol           —— 常规表驱动回归测试（T-01 的复习）
//   TestFuzzFoundTheseTwoBugs —— 把 fuzz 当年找到的两个 bug 钉成展品
//   FuzzParseSymbol           —— 真正的模糊目标，断言「不变量」
//
// 先读 symbol.go 顶部的关卡卡，再回来看代码。

package symbol

import (
	"strings"
	"testing"
)

// ─────────────────────────────────────────────────────────────────────────────
// 常规回归测试：把「已知输入 → 已知输出」固定住。
//
// 注意表里那两条标了「fuzz 发现」的用例 —— 它们本来不在我的表里。
// 这就是 fuzzing 和表驱动的关系：**fuzz 负责发现，表负责守住。**
// 模糊器找到的每一个 bug，修完之后都应该变成表里的一行，
// 这样它永远不会悄悄回归。
// ─────────────────────────────────────────────────────────────────────────────

func TestParseSymbol(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		wantBase  string
		wantQuote string
		wantErr   bool
	}{
		{"带分隔符", "BTC/USDT", "BTC", "USDT", false},
		{"无分隔符", "BTCUSDT", "BTC", "USDT", false},
		{"长后缀优先", "ETHUSDC", "ETH", "USDC", false},
		{"数字开头的合约", "1000SHIBUSDT", "1000SHIB", "USDT", false},
		{"前后空格", "  BTCUSDT  ", "BTC", "USDT", false},

		// ── 下面三条是 fuzz 找出来的，不是我想出来的 ─────────────────────
		{"fuzz发现·只有一个空格", " ", "", "", true},
		{"fuzz发现·只有计价单位", "USDT", "", "", true},
		{"fuzz发现·多余的分隔符", "0//", "", "", true},

		{"空字符串", "", "", "", true},
		{"分隔符在开头", "/USDT", "", "", true},
		{"分隔符在结尾", "BTC/", "", "", true},
		{"不认识的计价单位", "BTCEUR", "", "", true},
		{"不认识的计价单位（带分隔符）", "BTC/EUR", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base, quote, err := ParseSymbol(tt.raw)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseSymbol(%q) 期望报错，实际返回 (%q, %q)", tt.raw, base, quote)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseSymbol(%q) 意外报错: %v", tt.raw, err)
			}
			if base != tt.wantBase || quote != tt.wantQuote {
				t.Errorf("ParseSymbol(%q) = (%q, %q), want (%q, %q)",
					tt.raw, base, quote, tt.wantBase, tt.wantQuote)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 展品：把 fuzz 当年找到的两个 bug 原样复现。
//
// 为什么要把「错的代码」写在测试里？
//   因为它是最好的教材。你可以在这一步亲手跑一遍 bug 版本，
//   看它怎么炸、模糊器怎么把最小复现用例写进 testdata/。
//   如果只写在注释里，你永远感受不到那个「原来这么简单就能炸」的冲击。
//
// 这两条断言都是「PASS」—— 它们在证明「bug 确实存在」。
// 也就是说：**测试通过 = bug 被成功复现。** 这有点反直觉，但很常见。
// ─────────────────────────────────────────────────────────────────────────────

func TestFuzzFoundTheseThreeBugs(t *testing.T) {
	t.Run("bug1_一个空格导致越界panic", func(t *testing.T) {
		// 原始实现的顺序：先判空 → 再 TrimSpace → 再访问 raw[0]
		parseSymbolBug1 := func(raw string) (string, string, error) {
			if raw == "" { // ← 先判空
				return "", "", ErrEmptySymbol
			}
			raw = strings.TrimSpace(raw)        // ← 这里 " " 变成了 ""
			if raw[0] >= '0' && raw[0] <= '9' { // ← panic: index out of range [0] with length 0
				return "数字合约", "USDT", nil
			}
			return "other", "USDT", nil
		}

		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("期望 panic，实际没有 —— 这个展品已经失效，请更新注释")
			}
			t.Logf("✓ 成功复现 panic：%v", r)
			t.Logf("  输入是 %q（一个空格）—— 判断顺序错了", " ")
		}()

		_, _, _ = parseSymbolBug1(" ")
	})

	t.Run("bug2_USDT_返回空base不报错", func(t *testing.T) {
		// 原始实现：后缀匹配成功后直接返回，没检查 base 是否为空
		parseSymbolBug2 := func(raw string) (string, string, error) {
			raw = strings.TrimSpace(raw)
			for _, q := range knownQuotes {
				if strings.HasSuffix(raw, q) {
					return raw[:len(raw)-len(q)], q, nil // ← base 可能是 ""
				}
			}
			return "", "", ErrEmptySymbol
		}

		base, quote, err := parseSymbolBug2("USDT")

		if err != nil {
			t.Fatal("期望「不报错但结果错」，实际报错了 —— 这个展品已经失效")
		}
		if base != "" {
			t.Fatalf("期望 base 为空串，实际 %q —— 这个展品已经失效", base)
		}
		t.Logf("✓ 成功复现静默错误：ParseSymbol(%q) = (%q, %q), err = nil", "USDT", base, quote)
		t.Logf("  base 是空串却不报错 —— 下游拿它去查币种元数据必然查不到")
		t.Logf("  ⚠ 这个 bug 比 panic 更危险：崩溃会被告警抓住，静默错值会流到下单环节才爆")

		// 对比修好之后的版本：同样的输入必须报错
		if _, _, err := ParseSymbol("USDT"); err == nil {
			t.Error("修好之后的 ParseSymbol(\"USDT\") 必须报错")
		}
	})

	t.Run("bug3_分隔符分支没校验计价单位", func(t *testing.T) {
		// 原始实现：带分隔符的分支只检查了「两侧非空」
		parseSymbolBug3 := func(raw string) (string, string, error) {
			raw = strings.TrimSpace(raw)
			if i := strings.IndexByte(raw, '/'); i >= 0 {
				base, quote := raw[:i], raw[i+1:]
				if base == "" || quote == "" {
					return "", "", ErrEmptySymbol
				}
				return base, quote, nil // ← quote 是什么都照收，哪怕是个斜杠
			}
			return "", "", ErrEmptySymbol
		}

		base, quote, err := parseSymbolBug3("0//")
		if err != nil {
			t.Fatal("期望「不报错但结果错」，实际报错了 —— 这个展品已经失效")
		}
		if quote != "/" {
			t.Fatalf("期望 quote 是 %q，实际 %q —— 这个展品已经失效", "/", quote)
		}
		t.Logf("✓ 成功复现静默错误：ParseSymbol(%q) = (%q, %q), err = nil", "0//", base, quote)
		t.Logf("  quote 是个斜杠 —— 完全不可能是计价单位")

		// 对比修好之后的版本
		if _, _, err := ParseSymbol("0//"); err == nil {
			t.Error("修好之后的 ParseSymbol(\"0//\") 必须报错")
		}

		// 这个用例是模糊器自己写进仓库的，可以去看看长什么样：
		t.Logf("  最小复现用例已落盘：testdata/fuzz/FuzzParseSymbol/2117ecd131c4ac8b")
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// 真正的模糊目标。
//
// 和普通测试最大的区别：**这里断言的是「不变量」，不是「具体期望值」。**
//
//	普通测试：ParseSymbol("BTCUSDT") == ("BTC", "USDT")   ← 只能测你想得到的输入
//	模糊目标：对任意输入 raw，都必须满足下面三条不变量        ← 机器替你找输入
//
// 三条不变量：
//   ① 绝不 panic —— 模糊器最擅长发现这个
//   ② 报错时不能返回非空的 base/quote（避免调用方拿到半成品）
//   ③ 成功时 base 和 quote 都必须非空，且拼回去能覆盖原输入
//
// ⚠ 关于 f.Add 种子：种子语料不是「必须覆盖的用例」，而是给模糊器的**起点**。
//   它会从这些字符串出发做变异。种子越贴近真实输入，找到问题越快。
// ─────────────────────────────────────────────────────────────────────────────

func FuzzParseSymbol(f *testing.F) {
	// ── 种子语料 ──────────────────────────────────────────────────────
	f.Add("BTCUSDT")
	f.Add("BTC/USDT")
	f.Add("ETHUSDC")
	f.Add("1000SHIBUSDT")
	f.Add("USDT") // ← bug 2 的触发输入
	f.Add("0//")  // ← bug 3 的触发输入
	f.Add(" ")    // ← bug 1 的触发输入
	f.Add("/")
	f.Add("")

	f.Fuzz(func(t *testing.T, raw string) {
		base, quote, err := ParseSymbol(raw)

		// ── 不变量 ②：报错时不能返回半成品 ────────────────────────────
		if err != nil {
			if base != "" || quote != "" {
				t.Fatalf("ParseSymbol(%q) 返回了错误却又带出了结果 (%q, %q)", raw, base, quote)
			}
			return
		}

		// ── 不变量 ③：成功时 base/quote 必须非空 ──────────────────────
		if base == "" {
			t.Fatalf("ParseSymbol(%q) 成功了但 base 为空（quote=%q）", raw, quote)
		}
		if quote == "" {
			t.Fatalf("ParseSymbol(%q) 成功了但 quote 为空（base=%q）", raw, base)
		}

		// ── 不变量 ③（续）：拼回去要能覆盖原输入 ──────────────────────
		// 去掉空格和分隔符之后，base+quote 应该等于原输入。
		// 这条能抓住「解析时丢了字符」或「多算了字符」的 bug。
		stripped := strings.ReplaceAll(strings.TrimSpace(raw), "/", "")
		if base+quote != stripped {
			t.Fatalf("ParseSymbol(%q) = (%q, %q)，拼回去是 %q，与原输入 %q 不一致",
				raw, base, quote, base+quote, stripped)
		}
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Normalize 的往返性质，同样适合用不变量表达。
// ─────────────────────────────────────────────────────────────────────────────

func FuzzNormalize(f *testing.F) {
	f.Add("btcusdt")
	f.Add("eth/usdc")
	f.Add("  BTCUSDT  ")
	f.Add("USDT")

	f.Fuzz(func(t *testing.T, raw string) {
		norm, err := Normalize(raw)
		if err != nil {
			return // 解析失败是合法结果，不在这里断言
		}

		// 不变量：归一化结果必须能被重新解析，且两次结果一致（幂等）
		base1, quote1, err1 := ParseSymbol(norm)
		if err1 != nil {
			t.Fatalf("Normalize(%q) = %q，但这个结果自己解析不了: %v", raw, norm, err1)
		}

		norm2, err2 := Normalize(norm)
		if err2 != nil {
			t.Fatalf("Normalize 不幂等：Normalize(%q) 失败: %v", norm, err2)
		}
		if norm2 != norm {
			t.Fatalf("Normalize 不幂等：%q → %q → %q", raw, norm, norm2)
		}

		// 归一化后必须是大写
		if norm != strings.ToUpper(norm) {
			t.Fatalf("Normalize(%q) = %q，不是全大写", raw, norm)
		}
		_ = base1
		_ = quote1
	})
}
