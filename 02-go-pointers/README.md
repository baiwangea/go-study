# L1 · 指针与引用语义（关卡制）

L1 语言地基分两个目录：[`01-go-fundamentals`](../01-go-fundamentals/)（L1-01~10）+
本目录（**L1-11~16**）。这 6 关专治「从 PHP/Java/Python 过来最容易被绊倒」的三件事：
**一切都是拷贝、nil 的真面目、切片/map 的共享与失联**。

## 怎么用

```sh
cd 02-go-pointers

go run . list      # 关卡目录（编号 / 前置 / 目标）
go run . 5         # 只跑第 5 关（nil 陷阱）
go run . 1-6       # 按顺序全跑
go run .           # 同上
```

一关一个文件（`11_xxx.go` 起编），每关 30~50 行，只看 **目标 / 观察 / 思考 / 通关** 四行。

## 关卡表

| 编号 | 文件 | 主题 | 前置 | 关键坑 |
| :--- | :--- | :--- | :--- | :--- |
| L1-11 | `11_value_vs_pointer.go` | 值传递 vs 指针传递（`&` 与 `*`） | L1-01 | 改副本不生效 |
| L1-12 | `12_struct_copy.go` | 结构体整体拷贝、何时改用指针 | L1-11 | 赋值即复制全部字段 |
| L1-13 | `13_receiver.go` | 值接收者 vs 指针接收者 | L1-12 | 值接收者自增永远返回 1 |
| L1-14 | `14_slice_map_shared.go` | 切片三元组、`append` 扩容失联、map 引用语义 | L1-12 | 未扩容时两个切片互相写脏 |
| L1-15 | `15_nil_traps.go` | nil map/slice/指针/接口 四种 nil 的行为 | L1-14 | 接口装 nil 指针后 `err != nil` |
| L1-16 | `16_escape.go` | 取地址与逃逸分析（为什么没有悬垂指针） | L1-11 | 担心「返回局部变量地址」是多余的 |

## 目录结构

```
02-go-pointers/
├── main.go          # 关卡导航（list / 单关 / 区间 / 全跑）
├── level/level.go   # 关卡运行时（与 01 同一份，零依赖）
└── NN_主题.go       # 一关一文件，全部在模块根目录
```

关卡少（6 关）时不必再建主题子目录，直接放模块根目录，`main.go` 里按顺序登记。

## 实测要点（跑起来真的能看到）

- L1-13：`IncValue()` 连调三次都是 `1 1 1`，`IncPtr()` 是 `1 2 3`
- L1-14：`append` 未扩容时 `s1` 与 `s3` 互相影响；扩容后 `cap 3 → 8`，改 `big[0]` 不再影响 `s1`
- L1-15：`badDo(false) != nil` 为 **true**（坑），`goodDo(false)` 才是 false；
  且这个 bug **标准 `go vet` 抓不到**，要靠 staticcheck 的 nilness 分析
- L1-16：`go build -gcflags='-m' .` 能看到 `moved to heap: n`，证明返回局部变量地址安全

## 通关标准

能说清这三句话：① Go 里所有赋值与传参都是拷贝，想改原件就传指针；
② 切片是 (指针, len, cap) 三元组，共享底层数组但可能在 `append` 后失联；
③ 接口值 = (类型, 值)，装了 nil 指针的接口不是 nil。

## 下一站

[L2 · Go 并发模型](../03-go-concurrency/)（13 关）。其中 L2-11 的 context 取消、
L2-12 的 worker pool 是后面 [L3 · GoFrame 主线](../04-goframe/) 的直接前置。
