# L1 · Go 语言基础（直观学习模式）

**L1 阶段跨两个目录**：本目录是接口、错误、字符串、泛型、组合（L1-08~10、L1-19~27），
[`02-go-pointers`](../02-go-pointers/) 放 L1-11~17（值/指针拷贝、nil 陷阱、切片共享、逃逸）。
按编号学就是正确顺序。

本目录当前 **12 关**，**一个文件 = 一关 = 一个完整的 `main()`**。
直接 `go run` 那个文件就行。控制台用 🎯 / 📝 / 📌 / 💡 把标题、示例、总结分开，方便扫。

## 怎么用

```sh
cd 01-go-fundamentals

go run interfaces/01_define.go           # L1-08 接口 = 方法集合
go run interfaces/03_polymorphism.go     # L1-10 多态与类型断言
go run errors/01_error_chain.go          # L1-19 错误链
```

在仓库根目录也可以（Makefile 里的 `levels` / `level` 目标对所有模块通用）：

```sh
make levels M=./01-go-fundamentals               # 列出本模块所有关卡文件
make level  F=errors/01_error_chain.go           # 跑指定关卡（M 可省略）
```

每关文件头部的大段注释就是「关卡卡」：**核心概念 / 你会看到 / 思考题 / 通关标准 / 前置 / 标签**。
**思考题动手改过代码，才算通关**，可以进入下一关。

## 关卡表

| 编号 | 文件（`go run <文件>`） | 主题 | 前置 |
| :--- | :--- | :--- | :--- |
| L1-08 | `interfaces/01_define.go` | 接口 = 方法集合（行为契约） | — |
| L1-09 | `interfaces/02_implicit_impl.go` | 隐式实现、值/指针接收者、编译期断言 | L1-08 |
| L1-10 | `interfaces/03_polymorphism.go` | 多态、类型断言、`fmt.Stringer` | L1-09 |
| L1-19 | `errors/01_error_chain.go` | 错误链：`%w` vs `%v`、`errors.Is` / `As` / `Unwrap` | L1-10 |
| L1-20 | `errors/02_custom_error.go` | 自定义错误类型、哨兵错误、错误码模式 | L1-19 |
| L1-21 | `errors/03_panic_recover.go` | `panic`/`recover` 的边界、`panic(nil)`、跨 goroutine | L1-19 |
| L1-22 | `strings/01_bytes_rune.go` | `string`/`[]byte`/`rune` 与 UTF-8、中文截断 bug | — |
| L1-23 | `strings/02_builder_strconv.go` | `strings.Builder`、`strconv`、`strings` 高频函数 | L1-22 |
| L1-24 | `generics/01_type_params.go` | 泛型类型参数 `[T any]`、泛型类型、`comparable` | L1-08 |
| L1-25 | `generics/02_constraints.go` | 自定义约束 `~int \| ~float64`、`cmp.Ordered`、泛型的取舍 | L1-24 |
| L1-26 | `composition/01_embedding.go` | 结构体嵌入、方法提升、嵌入 ≠ 继承 | L1-09 |
| L1-27 | `composition/02_interface_embed.go` | 接口嵌入、小接口组合、`any`、`io.Reader` | L1-08 · L1-10 |

> 编号不连续是历史追加留下的：L1-11~17 在 [`02-go-pointers`](../02-go-pointers/)。
> 本目录里没有 `functions/`、`packages/`。

## 目录结构

```
01-go-fundamentals/
├── go.mod
├── README.md
├── README_SIMPLE.md
├── interfaces/                     # L1-08 ~ L1-10
│   ├── 01_define.go
│   ├── 02_implicit_impl.go
│   └── 03_polymorphism.go
├── errors/                         # L1-19 ~ L1-21
│   ├── 01_error_chain.go
│   ├── 02_custom_error.go
│   └── 03_panic_recover.go
├── strings/                        # L1-22 ~ L1-23
│   ├── 01_bytes_rune.go
│   └── 02_builder_strconv.go
├── generics/                       # L1-24 ~ L1-25
│   ├── 01_type_params.go
│   └── 02_constraints.go
└── composition/                    # L1-26 ~ L1-27
    ├── 01_embedding.go
    └── 02_interface_embed.go
```

**没有 `main.go`、没有 `level/` 运行时、没有 `levels.go` 注册表。**
每个关卡文件都自带 `func main()`，跑哪一关就 `go run` 哪个文件。

## 每个文件首行的 `//go:build level`，为什么必须留着

Go 规定同一个目录下 `package main` 只能有一个 `func main()`。比如 `interfaces/` 里 3 个
文件都自带 `main()`，不加标签就会编译失败：

```
./02_implicit_impl.go: main redeclared in this block
        ./01_define.go: other declaration of main
```

加了 `//go:build level` 之后：

- `go run interfaces/01_define.go` —— **显式指定文件名时会绕过构建约束**，照样能单独跑；
- `go build ./...` / `go vet ./...` —— 自动跳过这些文件，冲突消失。

> **不要换成 `//go:build ignore`。** 实测 `ignore` 会让 `go mod tidy` 认为模块里
> 没有任何代码，从而**清空 go.mod 的 require 列表**。自定义标签 `level` 没有这个问题。

## 两个附带的注意点

**1. `go run .` 不是入口。** 关卡文件全被标签排除后，模块根目录在常规构建下
匹配不到包。设计就是：**跑单个文件，而不是跑整个目录**。

```sh
go run errors/01_error_chain.go   # ✅ 正确用法
go run .                          # ❌ 本模块没有「整个目录」这个入口
```

**2. 控制台分区。** 跑起来会看到：

```
============================================================
🎯 L1-19 错误链
============================================================
📝 示例 1：……
------------------------------------------------------------
💡 关键总结
============================================================
```

标题、示例、总结用不同表情，扫一眼就能跳到下一块。

## 想自己加一关？照抄这个模板

```go
//go:build level

package main

import (
	"fmt"
	"strings"
)

/*
═══════════════════════════════════════════════════════════════════════
L1-XX：关卡标题
═══════════════════════════════════════════════════════════════════════

【核心概念】一句话说清这关搞定什么
【你会看到】✓ 运行后应该看到什么
【思考题】1. 思考题一？  2. 思考题二？
【通关标准】能做到什么算通关
【前置】L1-10              【标签】关键词 · 关键词
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L1-XX 关卡标题")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	fmt.Println("📝 示例 1：……")
	fmt.Println(strings.Repeat("-", 60))
	// 演示代码
}
```

存成 `errors/04_xxx.go` 之类，直接 `go run errors/04_xxx.go` 就能跑，
无需注册、无需改任何别的文件。

## 下一站

本目录 12 关打完后，去 [L1 · 指针与引用语义](../02-go-pointers/)（L1-11~17），
写过 PHP/Java/Python 的人最容易在这里翻车；然后进入
[L2 · Go 并发模型](../03-go-concurrency/)（23 关），再到 [L3 · GoFrame 主线](../04-goframe/)。
