// 本文件是 T-02 的测试代码，重点是三件事：
//
//	① t.Run 子测试：让「一张表」变成「一堆可以单独跑、单独失败的测试」
//	② t.Parallel：让互不影响的用例真正并发执行
//	③ t.Cleanup：比 defer 更适合「测试级清理」的钩子
//
// 先读一遍 risk.go 顶部的关卡卡，再回来看代码。

package risk

import (
	"os"
	"slices"
	"testing"
)

// ─────────────────────────────────────────────────────────────────────────────
// 表驱动 + 子测试：Go 里最常用的组合拳。
//
// 和 T-01 的裸循环相比只多了两行（t.Run 包一层 + t.Parallel），
// 换来的能力：
//   · go test -run 'TestCheck/名义金额' 可以只跑一条
//   · 某条用例 t.Fatalf 只终止它自己
//   · 每条用例可以并行
//
// 注意闭包捕获：这里直接用了 tt，没有写 `tt := tt`。
// Go 1.22 起 for-range 的迭代变量是**每次迭代新建**的，
// 所以 t.Parallel() 把闭包推迟执行也不会拿到错误的 tt。
// 如果你看到老项目里那句 tt := tt，那是在给 1.21 及以前打补丁，现在已经多余。
// ─────────────────────────────────────────────────────────────────────────────

func TestCheck(t *testing.T) {
	guard := Guard{
		MaxNotional: 10000,
		MinQty:      0.001,
		Allowed:     map[string]bool{"BTCUSDT": true, "ETHUSDT": true},
	}

	tests := []struct {
		name    string
		order   Order
		wantErr bool
	}{
		{"正常通过", Order{"BTCUSDT", 0.5, 3000}, false},
		{"白名单内的另一个币", Order{"ETHUSDT", 1, 3000}, false},
		{"交易对不在白名单", Order{"DOGEUSDT", 1, 100}, true},
		{"数量低于最小限制", Order{"BTCUSDT", 0.0001, 5}, true},
		{"超过名义金额上限", Order{"BTCUSDT", 1, 20000}, true},
		{"恰好等于上限（边界）", Order{"BTCUSDT", 1, 10000}, false},
		{"恰好等于最小数量（边界）", Order{"BTCUSDT", 0.001, 50}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // ← 每条用例独立并行；删掉它对比 -v 输出的顺序

			err := guard.Check(tt.order)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Check(%+v) 期望报错，实际通过", tt.order)
				}
				return
			}
			if err != nil {
				t.Fatalf("Check(%+v) 意外报错: %v", tt.order, err)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 演示「子测试里 t.Fatalf 不会连坐」。
//
// 这个测试默认**跳过**，因为它会故意失败一条子测试（教学用途，
// 不能让 go test ./... 一直是红的）。想亲眼看隔离效果就加环境变量：
//
//	DEMO=1 go test -v -run TestFatalfIsolation ./02_subtests/
//
// 你会看到：子测试 A 报 FAIL 之后，B、C 依然照常执行并 PASS。
// 把 t.Fatalf 换成父测试里的直接调用，整个测试会立刻终止，B、C 一行都不跑。
// ─────────────────────────────────────────────────────────────────────────────

func TestFatalfIsolation(t *testing.T) {
	if os.Getenv("DEMO") == "" {
		t.Skip("加 DEMO=1 运行：这个用例会故意失败一条子测试，用来演示隔离效果")
	}

	cases := []struct {
		name string
		ok   bool
	}{
		{"A·这条故意失败", false},
		{"B·照常执行", true},
		{"C·照常执行", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if !c.ok {
				t.Fatalf("故意失败 —— 看后面两条子测试是否还在跑")
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// t.Setenv：环境变量隔离。
//
// 和 t.Parallel 互斥 —— t.Setenv 会修改进程级环境变量，
// 并行跑会互相污染，所以 Go 直接禁止（同一个测试里先 t.Parallel 再 t.Setenv 会 panic）。
//
// 对照：
//   PHPUnit : putenv('X=1') 之后要记得在 tearDown 里还原，忘了就污染其他测试
//   JUnit   : 需要 System.setProperty + @AfterEach，或者引入 System Stubs 之类的库
//   Go      : t.Setenv 自动记录旧值、测试结束时自动还原，一行搞定
// ─────────────────────────────────────────────────────────────────────────────

func TestGuardFromEnv(t *testing.T) {
	t.Run("默认值", func(t *testing.T) {
		g, err := GuardFromEnv()
		if err != nil {
			t.Fatalf("默认配置不该报错: %v", err)
		}
		if g.MaxNotional != 10000 {
			t.Errorf("默认 MaxNotional = %v, want 10000", g.MaxNotional)
		}
		if !g.Allowed["BTCUSDT"] {
			t.Errorf("默认白名单应该包含 BTCUSDT，实际 %v", g.Allowed)
		}
	})

	t.Run("覆盖上限与白名单", func(t *testing.T) {
		t.Setenv("RISK_MAX_NOTIONAL", "250")
		t.Setenv("RISK_ALLOWED", "BTCUSDT, ETHUSDT ,SOLUSDT")

		g, err := GuardFromEnv()
		if err != nil {
			t.Fatalf("解析失败: %v", err)
		}
		if g.MaxNotional != 250 {
			t.Errorf("MaxNotional = %v, want 250", g.MaxNotional)
		}
		// 注意 " ETHUSDT " 前后的空格必须被 TrimSpace 吃掉
		for _, s := range []string{"BTCUSDT", "ETHUSDT", "SOLUSDT"} {
			if !g.Allowed[s] {
				t.Errorf("白名单缺少 %s，实际 %v", s, g.Allowed)
			}
		}
	})

	t.Run("非法数字要报错", func(t *testing.T) {
		t.Setenv("RISK_MAX_NOTIONAL", "一万")
		if _, err := GuardFromEnv(); err == nil {
			t.Fatal("RISK_MAX_NOTIONAL=一万 应该报错，实际通过了")
		}
	})

	// 上面三个子测试各自 Setenv 互不影响，也不需要任何手动还原。
	// 这里断言环境已经还原回默认值 —— 这就是 t.Setenv 的价值。
	t.Run("子测试结束后环境自动还原", func(t *testing.T) {
		if v := os.Getenv("RISK_MAX_NOTIONAL"); v != "" {
			t.Errorf("RISK_MAX_NOTIONAL 应该已还原为空，实际 %q", v)
		}
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// t.Cleanup：测试级清理钩子。
//
// 两个关键语义，都和 defer 不一样：
//
//	① 逆序执行（LIFO）—— 和 defer 一样，后注册的先跑。
//	② **执行时机是「测试结束时」，不是「注册它的函数返回时」。**
//	   这是它和 defer 最大的区别：写在 helper 里的 t.Cleanup
//	   会在**调用方测试结束**时才跑；写在 helper 里的 defer
//	   在 helper 一返回就跑掉了（对「关掉一个测试级资源」来说太早）。
//
// 断言写法有个小技巧：cleanup 是 LIFO 的，所以
// **先注册的断言最后执行** —— 把断言第一个注册进去，就能检查到全部清理结果。
// ─────────────────────────────────────────────────────────────────────────────

func TestCleanupLIFO(t *testing.T) {
	var order []string

	// 第一个注册 → 最后一个执行 → 此时 order 已收集完毕
	t.Cleanup(func() {
		// 注册顺序：assert(1) → helper(cleanup)(2) → A(3) → B(4) → C(5)
		// 执行顺序是 LIFO：C → B → A → helper(cleanup) → assert
		// 而 helper 里的 defer 早在 openResource 返回时就跑掉了，排在最前面。
		want := []string{
			"helper(opened)",
			"helper(defer)",                          // ← 立刻执行（helper 返回时）
			"C(cleanup)", "B(cleanup)", "A(cleanup)", // ← 测试结束时逆序执行
			"helper(cleanup)", // ← 同样在测试结束时，但比 A/B/C 早注册，所以最后跑
		}
		if !slices.Equal(order, want) {
			t.Errorf("清理顺序 = %v\n           want = %v", order, want)
		}
	})

	openResource(t, "helper", &order) // 里面的 defer 会立刻执行
	for _, name := range []string{"A", "B", "C"} {
		t.Cleanup(func() { order = append(order, name+"(cleanup)") })
	}
}

// openResource 演示 helper 里 defer 与 t.Cleanup 的语义差异。
func openResource(t *testing.T, name string, order *[]string) {
	t.Helper()

	// 注册在**测试结束时**执行
	t.Cleanup(func() { *order = append(*order, name+"(cleanup)") })

	// 注册在**本函数返回时**执行 —— 比 cleanup 早得多
	defer func() { *order = append(*order, name+"(defer)") }()

	*order = append(*order, name+"(opened)")
}
