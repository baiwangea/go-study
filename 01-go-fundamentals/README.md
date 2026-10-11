# L1 · Go 语言基础（直观学习模式）

**L1 阶段跨两个目录**：本目录放 L1-01~10（函数、包、接口），
[`02-go-pointers`](../02-go-pointers/) 放 L1-11~17（值/指针拷贝、nil 陷阱、切片共享、逃逸）。
关卡编号连续，按编号学就是正确顺序。

本目录把地基概念拆成 10 个小关卡，**一个文件 = 一关 = 一个完整的 `main()`**。
不需要任何导航框架，直接 `go run` 那个文件就行。

## 怎么用

```sh
cd 01-go-fundamentals

go run functions/01_basic_func.go        # L1-01 函数与返回值
go run packages/02_import.go             # L1-07 四种导入写法（可直接跳关）
go run interfaces/03_polymorphism.go     # L1-10 多态与类型断言
```

在仓库根目录也可以（Makefile 里的 `levels` / `level` 目标对所有模块通用）：

```sh
make levels M=./01-go-fundamentals                                     # 列出本模块所有 .go 文件
make level  M=./01-go-fundamentals F=functions/01_basic_func.go        # 跑指定关卡
```

> 注意：`levels` 用的是 `find *.go`，所以会连 `packages/helper/helper.go`
> （L1-06 的教具包）一起列出来，它本身没有 `main()`，不能单独跑。

每关文件头部的大段注释就是「关卡卡」：**核心概念 / 你会看到 / 思考题 / 通关标准 / 前置 / 标签**。
**思考题动手改过代码，才算通关**，可以进入下一关。

## 关卡表（L1-01 ~ L1-10）

| 编号 | 文件（`go run <文件>`） | 主题 | 前置 |
| :--- | :--- | :--- | :--- |
| L1-01 | `functions/01_basic_func.go` | 函数与返回值、返回类型的位置 | — |
| L1-02 | `functions/02_multi_return.go` | 多返回值与 `(结果, error)` 惯例 | L1-01 |
| L1-03 | `functions/03_variadic.go` | 变长参数 `...int` 与切片展开 `s...` | L1-01 |
| L1-04 | `functions/04_closure.go` | 闭包捕获的是变量本身 | L1-01 |
| L1-05 | `functions/05_recursion.go` | 递归与终止条件、栈溢出 | L1-02 |
| L1-06 | `packages/01_export.go` | 首字母大小写 = 唯一可见性规则 | L1-01 |
| L1-07 | `packages/02_import.go` | 普通/别名/点/空白 四种导入 | L1-06 |
| L1-08 | `interfaces/01_define.go` | 接口 = 方法集合（行为契约） | L1-06 |
| L1-09 | `interfaces/02_implicit_impl.go` | 隐式实现、值/指针接收者、编译期断言 | L1-08 |
| L1-10 | `interfaces/03_polymorphism.go` | 多态、类型断言、`fmt.Stringer` | L1-09 |

## 目录结构

```
01-go-fundamentals/
├── go.mod                          # module go-study/go-fundamentals
├── README.md
├── functions/                      # L1-01 ~ L1-05，一关一文件
│   ├── 01_basic_func.go
│   ├── 02_multi_return.go
│   ├── 03_variadic.go
│   ├── 04_closure.go
│   └── 05_recursion.go
├── packages/                       # L1-06 ~ L1-07
│   ├── 01_export.go
│   ├── 02_import.go
│   └── helper/helper.go            # 被导入的教具包（正常 package helper，非关卡）
└── interfaces/                     # L1-08 ~ L1-10
    ├── 01_define.go
    ├── 02_implicit_impl.go
    └── 03_polymorphism.go
```

**没有 `main.go`、没有 `level/` 运行时、没有 `levels.go` 注册表。**
每个关卡文件都自带 `func main()`，跑哪一关就 `go run` 哪个文件。

## 每个文件首行的 `//go:build level`，为什么必须留着

Go 规定同一个目录下 `package main` 只能有一个 `func main()`。比如 `functions/` 里 5 个
文件都自带 `main()`，不加标签就会编译失败：

```
./02_multi_return.go:41:6: main redeclared in this block
        ./01_basic_func.go:52:6: other declaration of main
```

加了 `//go:build level` 之后：

- `go run functions/01_basic_func.go` —— **显式指定文件名时会绕过构建约束**，照样能单独跑；
- `go build ./...` / `go vet ./...` —— 自动跳过这些文件，冲突消失。

> **不要换成 `//go:build ignore`。** 实测 `ignore` 会让 `go mod tidy` 认为模块里
> 没有任何代码，从而**清空 go.mod 的 require 列表**。自定义标签 `level` 没有这个问题。

## 两个附带的注意点

**1. `go run .` 不再可用。** 关卡文件全被标签排除后，模块根目录在常规构建下
「匹配不到包」，`go run .` 会报 `no Go files in ...`（退出码 1）。
这不是坏了，而是本模块的设计：**跑单个文件，而不是跑整个目录**。

```sh
go run functions/01_basic_func.go   # ✅ 正确用法
go run .                            # ❌ 本模块没有「整个目录」这个入口
```

**2. `go build ./...` / `go vet ./...` 依然正常。** 因为 `packages/helper` 是一个
真实可编译的包（不带构建标签），这两个命令能正常匹配到包并退出 0；被 `level` 标签
排除的关卡文件会被静默跳过。

## 想自己加一关？照抄这个模板

```go
//go:build level

package main

import "fmt"

/*
═══════════════════════════════════════════════════════════════════════
L1-11：关卡标题
═══════════════════════════════════════════════════════════════════════

【核心概念】一句话说清这关搞定什么
【你会看到】✓ 运行后应该看到什么
【思考题】1. 思考题一？  2. 思考题二？
【通关标准】能做到什么算通关
【前置】L1-10              【标签】关键词 · 关键词
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n=== L1-11 关卡标题 ===")
	// 演示代码
}
```

存成 `functions/06_xxx.go` 之类，直接 `go run functions/06_xxx.go` 就能跑，
无需注册、无需改任何别的文件。

## 下一站

L1-01~10 通关后，先把 [L1 · 指针与引用语义](../02-go-pointers/)（L1-11~17）打完 ——
写过 PHP/Java/Python 的人最容易在这里翻车；之后进入 [L2 · Go 并发模型](../03-go-concurrency/)（13 关），
再到 [L3 · GoFrame 主线](../04-goframe/)。
