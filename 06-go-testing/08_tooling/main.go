// 本文件是 T-08 的「说明书 + CI 配置」打印程序。
//
// 运行：go run ./08_tooling/
//
// 为什么这一关是一个「只打印文字」的程序？
// 因为质量工具的价值不在代码里，而在**命令怎么用、CI 怎么配**。
// 把它做成一段能跑的命令清单 + 一份可以直接抄走的 CI 配置，
// 比写一堆注释更容易落地 —— 你可以直接复制粘贴到终端里跑。

package main

import "fmt"

// ciYAML 是可以直接放进 .github/workflows/ci.yml 的配置。
//
// 内嵌成字符串是为了让本关自包含（不需要你在仓库里翻文件）。
// 仓库根目录已经有一份可用的 .github/workflows/ci.yml，内容与本字符串一致。
//
// 三个设计决定，每个都有理由：
//  1. **每个模块单独跑，不用 go build ./...** —— 这是多模块仓库（go.work）的必然，
//     根目录一条命令覆盖不了全部模块。
//  2. **-race 单独一条命令** —— race 检测让测试慢 5~10 倍，
//     和普通测试混在一起会让整个流水线变慢，分开跑可以并行、也可以只在主分支跑。
//  3. **不依赖任何外部服务** —— 本仓库的关卡在缺 MySQL/Redis 时会明确跳过，
//     所以 CI 里不需要起服务，也就不会有「CI 挂了但其实代码没问题」的假告警。
const ciYAML = `name: CI

on:
  push:
    branches: [main]
  pull_request:

jobs:
  build-and-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.work
          cache: true

      # go.work.sum 被 gitignore 了，CI 里需要重新生成
      - name: 准备 workspace
        run: go work sync

      - name: 编译所有模块
        run: make build

      - name: 静态检查
        run: make vet

      - name: 单元测试
        run: make test

      - name: 竞态检测（单独跑，比普通测试慢 5~10 倍）
        run: make race

      - name: 关卡文件语法检查
        # 关卡文件带 //go:build level 标签，常规 build 会跳过它们，
        # 所以这里用 gofmt 兜一层，避免坏代码混进仓库
        run: test -z "$(gofmt -l $(find . -name '*.go' -not -path './.git/*'))"
`

func main() {
	fmt.Println("\n=== T-08 覆盖率与质量工具链 ===")

	// ── 示例 1：覆盖率 ───────────────────────────────────────────────────
	fmt.Println("\n【示例 1】覆盖率 —— 找出没被测到的代码")
	fmt.Println("  cd 06-go-testing")
	fmt.Println("  go test -cover ./08_tooling/                  # 只看一个总百分比")
	fmt.Println("  go test -coverprofile=cover.out ./08_tooling/ # 生成明细")
	fmt.Println("  go tool cover -func=cover.out                 # 按函数列出百分比")
	fmt.Println("  go tool cover -html=cover.out                 # 浏览器里看红色标注 ← 最有用")

	// ── 示例 2：静态检查 ─────────────────────────────────────────────────
	fmt.Println("\n【示例 2】静态检查 —— 抓编译器不管的错")
	fmt.Println("  go vet ./...                                  # 标准库自带，零依赖")
	fmt.Println("  go install honnef.co/go/tools/cmd/staticcheck@latest")
	fmt.Println("  staticcheck ./...                             # 更严格，抓未使用的赋值、可疑比较等")

	// ── 示例 3：竞态检测 ─────────────────────────────────────────────────
	fmt.Println("\n【示例 3】竞态检测 —— 并发代码必开")
	fmt.Println("  go test -race ./...")
	fmt.Println("  go run -race 03-go-concurrency/mutexes/01_race.go   # 关卡文件也能这样跑")
	fmt.Println("  ⚠ 慢 5~10 倍，但能抓到「只在特定时序下出现」的数据竞态")

	// ── 示例 4：本关的覆盖率现状 ─────────────────────────────────────────
	fmt.Println("\n【示例 4】本关的覆盖率现状（money.go 故意只被部分覆盖）")
	for _, fn := range []struct {
		name string
		pct  string
	}{
		{"FormatQty", "部分覆盖"},
		{"FeeTier", "部分覆盖"},
		{"IsMakerOnly", "0%（完全没测）"},
		{"DescribeTier", "0%（完全没测）"},
	} {
		fmt.Printf("  %-14s %s\n", fn.name, fn.pct)
	}
	fmt.Println("  → 目标不是把这些数字抬到 100%，而是**补上 0% 的那两个函数的边界用例**")

	// ── 示例 5：CI 配置 ─────────────────────────────────────────────────
	fmt.Println("\n【示例 5】CI 配置（可直接复制到 .github/workflows/ci.yml）")
	fmt.Println()
	for _, line := range splitLines(ciYAML) {
		fmt.Println("  " + line)
	}

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n【关键总结】")
	fmt.Println("  • 覆盖率是**盲区指标**，不是质量指标：100% 覆盖 ≠ 有断言")
	fmt.Println("  • 看覆盖率要用 -html 视图，重点看「哪些分支是红的」，不是看那个百分比")
	fmt.Println("  • go vet 抓的是编译器不管的错（格式串类型不匹配、锁被复制、无用赋值…）")
	fmt.Println("  • -race 必开，但单独跑 —— 它慢 5~10 倍")
	fmt.Println("  • CI 的意义是「让回归无处可藏」：每次 push 自动跑 build/vet/test/race")
	fmt.Println("  • 本仓库所有关卡在缺外部依赖时会明确跳过，所以 CI 不需要起 MySQL/Redis")
}

// splitLines 把多行字符串拆成行，去掉末尾空行。
func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	// 去掉末尾空行
	for len(out) > 0 && out[len(out)-1] == "" {
		out = out[:len(out)-1]
	}
	return out
}
