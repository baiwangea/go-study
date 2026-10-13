# L1 · 指针与引用语义（直观学习模式）

L1 语言地基分两个目录：[`01-go-fundamentals`](../01-go-fundamentals/)（接口、错误、字符串、泛型、组合）+
本目录（**L1-11~17**）。这 7 关专治「从 PHP/Java/Python 过来最容易被绊倒」的四件事：
**一切都是拷贝、nil 的真面目、切片/map 的共享与失联、接口靠方法而不是 implements**。

## 怎么用

**一个文件 = 一关 = 一个完整的 `main()`。不需要导航框架，直接运行那个文件。**

```sh
cd 02-go-pointers

go run 11_value_vs_pointer.go     # L1-11 值传递 vs 指针传递
go run 15_nil_traps.go            # L1-15 nil 陷阱（可直接跳关）
go run 17_php_vs_go_interface.go  # L1-17 PHP implements vs Go 隐式实现
```

在仓库根目录也可以：

```sh
make levels M=./02-go-pointers                          # 列出本模块所有关卡文件
make level  F=11_value_vs_pointer.go                    # 跑指定关卡（M 可省略）
```

每关文件的头部注释就是「关卡卡」：**核心概念 / 你会看到 / 思考题 / 通关标准 / 前置**。
**思考题动手改过代码，才算通关**，可以进入下一关。

## 关卡表

| 编号 | 文件（`go run <文件>`） | 主题 | 前置 | 关键坑 |
| :--- | :--- | :--- | :--- | :--- |
| L1-11 | `11_value_vs_pointer.go` | 值传递 vs 指针传递（`&` 与 `*`） | L1-01 | 改副本不生效 |
| L1-12 | `12_struct_copy.go` | 结构体整体拷贝、何时改用指针 | L1-11 | 赋值即复制全部字段 |
| L1-13 | `13_receiver.go` | 值接收者 vs 指针接收者 | L1-12 | 值接收者自增永远返回 1 |
| L1-14 | `14_slice_map_shared.go` | 切片三元组、`append` 扩容失联、map 引用语义 | L1-12 | 未扩容时两个切片互相写脏 |
| L1-15 | `15_nil_traps.go` | nil map/slice/指针/接口 四种 nil 的行为 | L1-14 | 接口装 nil 指针后 `err != nil` |
| L1-16 | `16_escape.go` | 取地址与逃逸分析（为什么没有悬垂指针） | L1-11 | 担心「返回局部变量地址」是多余的 |
| L1-17 | `17_php_vs_go_interface.go` | PHP `implements` vs Go 隐式实现 | L1-08, L1-13 | 找 `implements` 关键字、想用 `extends` 复用 |

## 目录结构

```
02-go-pointers/
├── go.mod                        # module go-study/go-pointers
├── README.md
├── 11_value_vs_pointer.go        # ← 一个文件一关，每关自带 func main()
├── 12_struct_copy.go
├── 13_receiver.go
├── 14_slice_map_shared.go
├── 15_nil_traps.go
├── 16_escape.go
└── 17_php_vs_go_interface.go
```

**没有 `main.go`、没有 `level/` 运行时、没有 `levels.go` 注册表、没有占位文件。**
目录里除了关卡文件，只有 README 和 go.mod。

## 每个文件首行的 `//go:build level`，为什么必须留着

Go 规定同一个目录下 `package main` 只能有一个 `func main()`。7 个文件都自带 `main()`，
不加标签就会编译失败：

```
./12_struct_copy.go:57:6: main redeclared in this block
        ./11_value_vs_pointer.go:43:6: other declaration of main
```

加了 `//go:build level` 之后：

- `go run 11_value_vs_pointer.go` —— **显式指定文件名时会绕过构建约束**，照样能单独跑；
- `go build ./...` / `go vet ./...` —— 自动跳过这些文件，冲突消失。

> **不要换成 `//go:build ignore`。** 实测 `ignore` 会让 `go mod tidy` 认为模块里
> 没有任何代码，从而**清空 go.mod 的 require 列表**（04-goframe 会因此丢掉
> goframe / mysql / redis 依赖）。自定义标签 `level` 没有这个问题。

## 两个附带的注意点

**1. `go run .` 不再可用。** 关卡文件全被标签排除后，模块根目录在常规构建下
「匹配不到任何包」，`go run .` 会报 `build constraints exclude all Go files`。
这不是坏了，而是本模块的设计：**跑单个文件，而不是跑整个目录**。

```sh
go run 11_value_vs_pointer.go   # ✅ 正确用法
go run .                        # ❌ 本模块没有「整个目录」这个入口
```

**2. 根目录的 `make build` / `make vet` 已经能正确识别这种情况。**
`go build ./...` 和 `go vet ./...` 在「匹配不到包」时都会以退出码 1 结束，
所以 `Makefile` 里对这两个目标加了显式放行（见 `NOPKG` 变量），
输出会显示 `OK (本模块没有可编译的命令行程序)`。这不是掩盖错误 ——
真正的编译/静态检查失败依然会 `FAIL` 并中断。

## 实测要点（跑起来真的能看到）

- L1-13：`IncValue()` 连调三次都是 `1 1 1`，`IncPtr()` 是 `1 2 3`
- L1-14：`append` 未扩容时 `s1` 与 `s3` 互相影响；扩容后 `cap 3 → 8`，改 `big[0]` 不再影响 `s1`
- L1-15：`badDo(false) != nil` 为 **true**（坑），`goodDo(false)` 才是 false；
  且这个 bug **标准 `go vet` 抓不到**，要靠 staticcheck 的 nilness 分析
- L1-16：`go build -o /dev/null -gcflags='-m' 16_escape.go` 能看到 `moved to heap: n`
- L1-17：拿仓库根目录 `demo/index.php`（PHP 版）与 `demo/main.go`（Go 版）做真对照 ——
  同一段逻辑，PHP 要写 `class Cat extends Animal implements Loggable`，Go 只需要有方法；
  输出里 `Speaker` 先后接住 Dog 与 Cat，而 `NotASpeaker` 不满足接口（真实代码里是编译期报错）

## 通关标准

能说清这三句话：① Go 里所有赋值与传参都是拷贝，想改原件就传指针；
② 切片是 (指针, len, cap) 三元组，共享底层数组但可能在 `append` 后失联；
③ 接口值 = (类型, 值)，装了 nil 指针的接口不是 nil。

## 下一站

[L2 · Go 并发模型](../03-go-concurrency/)（13 关）。其中 context 取消、
worker pool 是后面 [L3 · GoFrame 主线](../04-goframe/) 的直接前置。
