# T · 测试与工程实践（横切模块）

前面四个模块教你「怎么写 Go」，这一个教你「**怎么证明你写对了**」。

这是全仓库唯一一个**没有 `_test.go` 会显得荒谬**的模块 —— 在它出现之前，
这个仓库有 46 个关卡、0 个测试文件。补上它之后，你才有资格说自己在写工程代码，
而不只是在写练习题。

## 怎么用

**一个目录 = 一关。** 和其他模块的「一个文件 = 一关」不一样，原因见下。

```sh
cd 06-go-testing

go test -v ./01_table_driven/       # 跑一关
go test -v ./...                    # 跑全部
go test -race ./...                 # 并发相关的关卡建议带 -race
```

在仓库根目录也可以：

```sh
make levels M=./06-go-testing                    # 列出本模块所有关卡
make level  F=01_table_driven                    # 跑指定关卡（M 可省略）
```

## 关卡表

| 编号 | 目录 | 主题 | 前置 |
| :--- | :--- | :--- | :--- |
| T-01 | `01_table_driven/` | 表驱动测试 · `t.Errorf` vs `t.Fatalf` · `t.Helper` | L1-02 |
| T-02 | `02_subtests/` | 子测试 `t.Run` · `t.Parallel` · `t.Cleanup` · `t.Setenv` | T-01 |
| T-03 | `03_httptest/` | `httptest` 直测 handler vs 起真服务 · 客户端超时 | L1-08, T-02 |
| T-04 | `04_benchmark/` | 基准测试 · `b.N` · `-benchmem` · 优化前后对比 | T-01, L1-23 |
| T-05 | `05_pprof/` | pprof 抓热点 · `-list` 定位到行 · flat vs cum | T-04 |
| T-06 | `06_fuzz/` | 模糊测试 · 不变量 · 崩溃复现自动落盘 | T-01 |
| T-07 | `07_stub_mock/` | 接口打桩 · 依赖注入 · spy · 为什么 Go 不需要 mock 框架 | L1-08, L2-11 |
| T-08 | `08_tooling/` | 覆盖率 · `go vet` · staticcheck · `-race` · CI | T-01 ~ T-07 |

编号顺序即推荐学习顺序。**建议在学完 `03-go-concurrency`（L2）之后、进入 `04-goframe`（L3）之前插入学习**，
也可以随时回补 —— 它和前后模块没有硬依赖，只有 T-03 用到了 HTTP、T-07 用到了 context。

## 目录结构

```
06-go-testing/
├── go.mod                    # module go-study/testing
├── README.md
├── 01_table_driven/
│   ├── calc.go               # 包注释 = 关卡卡
│   └── calc_test.go
├── 02_subtests/
│   ├── risk.go
│   └── risk_test.go
├── ...
├── 06_fuzz/
│   └── testdata/fuzz/FuzzParseSymbol/   # ← fuzz 自动落盘的最小复现用例，要提交进版本库
└── 08_tooling/
    ├── main.go               # 打印工具链命令 + CI 配置
    ├── money.go              # 故意只被部分覆盖，用来看覆盖率盲区
    └── money_test.go
```

## 为什么本模块是「一个目录 = 一关」

其他模块的关卡是「一个文件 = 一关 = 一个 `func main()`」，用 `go run` 跑。
本模块做不到，因为**测试天然是「一个包 + 一个测试文件」**：

- 测试文件必须和被测代码**在同一个包**里（`package calc` + `calc_test.go`），
  这样才能访问包内未导出的符号 —— 这是 Go 的刻意设计，不是限制；
- 单文件 `func main()` 装不下「被测代码 + 测试代码」两件事；
- 而且本模块的「运行」就是 `go test`，不是 `go run`。

所以这里的约定是：**一个目录 = 一关 = 一个完整的包（实现 + 测试）**，
关卡卡写在实现文件的包注释里（`go doc` 也能看到）。

## 实测记录

几处真实跑出来的数字和发现，都写进了对应关卡的注释：

| 关卡 | 发现 |
| :--- | :--- |
| T-04 | 8 个元素时 `+=` 只比 `strings.Builder` 慢 1.8 倍；到 10000 个元素变成 **慢 663 倍、多分配 1226 倍内存**。基准的输入规模会骗人 |
| T-04 | 「优化」把 `%v` 换成 `strconv.FormatFloat('f',-1,64)`，**基准显示快了 1.6 倍，但正确性测试立刻红了** —— 极小值下两者产生不同的订单键，会导致幂等去重失效、重复下单 |
| T-05 | `pprof -list=classifySlow` 精确指向第 138 行 `regexp.MustCompile`：440ms 里有 440ms 在这一行 |
| T-06 | 模糊器在本关开发过程中**现场找出第三个 bug**（输入 `"0//"` → 返回 quote 为 `"/"`），最小复现用例已自动落盘到 `testdata/` |
| T-08 | 覆盖率 17.4%，但真正有意义的是「FormatQty 100% / IsMakerOnly 0%」这个对比 —— 覆盖率是盲区指标，不是质量指标 |

## 通关标准

能说清这三句话：

1. **表驱动 + 子测试**是 Go 测试的默认姿势，失败信息必须能定位到「哪个用例、拿到什么、期望什么」；
2. **先测量再优化**：benchmark 定快慢、pprof 定位置，而且优化必须配正确性测试兜底；
3. **覆盖率是盲区指标**：要看的是「哪些分支是红的」，不是那个百分比。

## 下一站

[L3 · GoFrame 主线](../04-goframe/)。到这里你已经会写测试了，
`04-goframe` 里的中间件、错误码、JWT 鉴权都是天然适合写测试的地方 ——
学的时候可以顺手给它们补上 `_test.go`。
