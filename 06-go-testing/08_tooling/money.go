/*
═══════════════════════════════════════════════════════════════════════
T-08：覆盖率与质量工具链（cover / vet / staticcheck / race / CI）
═══════════════════════════════════════════════════════════════════════

【核心概念】

	前面七关解决「怎么写测试」。这一关解决「怎么知道测试够不够、代码干不干净」。

	五个工具，各管一段：

	    go test -cover            覆盖率：哪些代码没被测到
	    go vet                    标准库自带的静态检查（编译器抓不到的错）
	    staticcheck               更严格的第三方静态检查（要单独装）
	    go test -race             数据竞态检测（并发代码必开）
	    GitHub Actions            把上面四件事搬进 CI，每次 push 自动跑

	⚠ 关于覆盖率，必须先破除三个误解：

	  ① **覆盖率不是质量指标，是「盲区指标」。**
	     100% 覆盖率的代码可以一行断言都没有 —— 它只说明「这行被执行过」，
	     不说明「结果被检查过」。反过来，80% 覆盖率也可能是很扎实的测试。

	  ② **不要为了凑覆盖率写测试。** 为了把 78% 抬到 90% 而去测 getter/setter，
	     是纯粹的浪费。**要看的是「哪些关键分支没被覆盖」**，
	     尤其是错误分支和边界分支。

	  ③ **看覆盖率要用 HTML 视图，不要只看那个百分比。**
	     `go tool cover -html` 会把没覆盖的行标红，一眼就能看出
	     「哦，原来所有错误分支我都没测」。

【运行方式】

	cd 06-go-testing

	# 1. 覆盖率（本关的 money.go 故意只被部分覆盖）
	go test -cover ./08_tooling/
	go test -coverprofile=cover.out ./08_tooling/
	go tool cover -func=cover.out          # 按函数列出覆盖率
	go tool cover -html=cover.out          # 浏览器打开，未覆盖的行标红 ← 最有用

	# 2. 静态检查
	go vet ./...
	# staticcheck 需要单独安装，装不上不影响本关
	# go install honnef.co/go/tools/cmd/staticcheck@latest
	# staticcheck ./...

	# 3. 竞态检测（并发代码必开，慢 5~10 倍但值得）
	go test -race ./...

	# 4. 看本关打印的 CI 配置
	go run ./08_tooling/

【你会看到】

	本机实测（go1.27）：

	    $ go test -cover ./08_tooling/
	    ok  	go-study/testing/08_tooling	0.4s	coverage: 17.4% of statements

	    $ go tool cover -func=cover.out
	    go-study/testing/08_tooling/main.go:66:     main         0.0%
	    go-study/testing/08_tooling/main.go:122:    splitLines   0.0%
	    go-study/testing/08_tooling/money.go:120:   FormatQty    100.0%
	    go-study/testing/08_tooling/money.go:144:   FeeTier      50.0%
	    go-study/testing/08_tooling/money.go:162:   IsMakerOnly  0.0%   ← 盲区
	    go-study/testing/08_tooling/money.go:167:   DescribeTier 0.0%   ← 盲区
	    total:                                      (statements) 17.4%

	这张表本身就是本关最好的教材，注意三个细节：

	  ① **FormatQty 是 100%，IsMakerOnly 是 0%。**
	     这很典型：写了测试的函数往往被测得很全，
	     而「看起来简单到不用测」的函数反而一行都没覆盖。
	     真实项目里的盲区几乎全是这么产生的。

	  ② **main 和 splitLines 都是 0%，把总覆盖率拉低了。**
	     这是 Go 覆盖率的一个已知问题：**可执行的 main 包会被算进分母**，
	     而它本来就没什么好测的。
	     真实项目里的处理办法：把业务逻辑拆到独立包里（`internal/xxx`），
	     main 包只留「解析参数 + 调业务 + 打印」这几行 ——
	     那样覆盖率数字才有意义。本关故意没拆，就是为了让你看到这个现象。

	  ③ **FeeTier 是 50%**：5 个分支只测了 2 个。
	     用 -html 看会看到后三个 case 是红的。

	所以「17.4%」这个数字本身没有意义。
	有意义的是：**FormatQty 已经测透了，而 IsMakerOnly / DescribeTier 一行都没测。**

【思考题】（动手改代码验证，别空想）

 1. 跑 go tool cover -html=cover.out，看哪几行是红的。
    然后给 money_test.go 补测试，把 IsMakerOnly 从 0% 抬起来。
    注意：**先想清楚这个函数有哪些边界，再写用例** ——
    不要为了凑数字写一个只调用一次的「覆盖率测试」。

 2. 故意在 money.go 里写一个 `go vet` 能抓到但编译器抓不到的错，比如：
    fmt.Printf("qty = %d\n", 0.5)     // %d 配 float64
    跑 go vet ./08_tooling/ 看它报什么。
    然后想一想：**为什么编译器不报这个错？**（提示：%d 的参数是 any）

 3. 覆盖率有没有可能「虚高」？
    写一个测试，它调用了 FormatQty 但**不断言任何东西**，
    然后看覆盖率变化。这说明「覆盖率」和「测试质量」是什么关系？

 4. 看 main() 打印的 CI 配置。想清楚：
    为什么 CI 里 `make test` 和 `go test -race` 要分开跑？
    如果合成一条命令，会有什么代价？（提示：-race 让测试慢 5~10 倍）

 5. 生产项目里，覆盖率阈值应该卡在多少？为什么？
    提示：没有标准答案，但「卡 100%」几乎一定是错的 ——
    它会逼着团队写一堆没有断言的测试来交差。

【通关标准】

	能完整跑通「cover → coverprofile → cover -html」这条链路并读懂红色标注，
	能说清「覆盖率是盲区指标而不是质量指标」，
	并且知道 CI 里该跑哪几条命令。

【前置】T-01 ~ T-07   【标签】覆盖率 · go vet · staticcheck · race · CI
═══════════════════════════════════════════════════════════════════════
*/
package main

import (
	"fmt"
	"strings"
)

// ─────────────────────────────────────────────────────────────────────────────
// 被测代码：交易对数量格式化与费率档位。
//
// ⚠ 这两个函数**故意没有被完整测试**（money_test.go 只覆盖了一部分），
//   目的是让你亲眼看到覆盖率工具怎么指出盲区。
//   不要把它们改成「全都被测到」—— 那样这一关就没东西可看了。
// ─────────────────────────────────────────────────────────────────────────────

// FormatQty 把下单数量格式化成交易所要求的字符串。
//
// 规则：按 stepSize 决定小数位数，去掉末尾多余的 0。
//
//	stepSize=0.001, qty=1.5    → "1.500"
//	stepSize=1,     qty=3      → "3"
//	stepSize=0.01,  qty=0.005  → 小于一个步长，报错
func FormatQty(qty, stepSize float64) (string, error) {
	if stepSize <= 0 {
		return "", fmt.Errorf("stepSize 必须大于 0，实际 %v", stepSize)
	}
	if qty < stepSize {
		return "", fmt.Errorf("数量 %v 小于最小步长 %v", qty, stepSize)
	}

	// 从 stepSize 推出小数位数：0.001 → 3，1 → 0
	digits := 0
	for s := stepSize; s < 1 && digits < 8; s *= 10 {
		digits++
	}

	s := fmt.Sprintf("%.*f", digits, qty)
	return s, nil
}

// FeeTier 按 30 天交易量返回费率档位（单位：基点 bp，1bp = 万分之一）。
//
//	< 10 万 USDT   → 10 bp（普通用户）
//	< 100 万 USDT  → 8 bp
//	< 1000 万 USDT → 5 bp
//	>= 1000 万     → 2 bp（VIP）
func FeeTier(volume30d float64) int {
	switch {
	case volume30d < 0:
		return 10 // 异常值按最高费率处理，宁可多收
	case volume30d < 100_000:
		return 10
	case volume30d < 1_000_000:
		return 8
	case volume30d < 10_000_000:
		return 5
	default:
		return 2
	}
}

// IsMakerOnly 判断该档位是否只能挂单（maker）。
//
// 高等级用户走 maker 费率更低，但必须挂单成交，吃单会被拒。
func IsMakerOnly(tier int) bool {
	return tier <= 5
}

// DescribeTier 生成一句人类可读的描述，用于通知和日志。
func DescribeTier(tier int) string {
	var name string
	switch {
	case tier <= 2:
		name = "VIP"
	case tier <= 5:
		name = "高级"
	case tier <= 8:
		name = "标准"
	default:
		name = "基础"
	}

	var b strings.Builder
	b.WriteString(name)
	b.WriteString("（")
	b.WriteString(fmt.Sprintf("%d bp", tier))
	b.WriteString("）")
	return b.String()
}
