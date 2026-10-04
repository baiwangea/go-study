# L1 · Go 语言基础（关卡制）

**L1 阶段跨两个目录**：本目录放 L1-01~10（函数、包、接口），
[`02-go-pointers`](../02-go-pointers/) 放 L1-11~16（值/指针拷贝、nil 陷阱、切片共享、逃逸）。
关卡编号连续，按编号学就是正确顺序。

本目录把地基概念拆成 10 个小关卡，一关一个文件、一次只看一个知识点，
跑一关、改一关、验一关，不用一口气读完整个模块。

## 怎么用

```sh
cd 01-go-fundamentals

go run . list       # 先看关卡目录（编号 / 前置 / 目标）
go run . 1          # 只跑第 1 关
go run . 3-5        # 跑第 3 到第 5 关
go run .            # 全跑一遍（复习时用，初学不建议）
```

或在仓库根目录：`make run M=./01-go-fundamentals`（等价于 `go run .`）。

## 通关契约（每关只看这四行）

运行任一关卡，会先打出这张「关卡卡」：

| 字段 | 含义 |
| :--- | :--- |
| **目标** | 这一关要搞定的一件事，别贪多 |
| **观察** | 运行结果里应该看到什么，用来确认自己看懂了 |
| **思考** | 2~3 个问题，**必须动手改代码验证**，不是空想 |
| **通关** | 过关标准：能做到什么才算真的会了 |

> 规矩：思考题没动手改过，就不要进下一关。改坏了编译错误也别急着回滚，先读懂报错信息。

## 关卡表（L1-01 ~ L1-10）

| 编号 | 文件 | 主题 | 前置 |
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
├── main.go              # 关卡导航：list / 单关 / 区间 / 全跑
├── level/level.go       # 关卡运行时（Level 结构 + 卡片打印），全仓库通用模板
├── functions/           # L1-01 ~ L1-05，一关一文件 + levels.go 注册表
├── packages/            # L1-06 ~ L1-07
│   └── helper/          # 被导入的示例包（演示导出与非导出）
└── interfaces/          # L1-08 ~ L1-10
```

## 想自己加一关？照抄这个模板

```go
package functions

import (
	"fmt"

	"go-study/go-fundamentals/level"
)

// L11 你的第 11 关。
func L11() level.Level {
	return level.Level{
		ID:        "L1-11",
		Title:     "关卡标题",
		Tags:      "关键词 · 关键词",
		Pre:       "L1-10",              // 前置关卡编号
		Goal:      "一句话说清这关搞定什么",
		Observe:   "运行后应该看到什么",
		Questions: []string{"思考题 1？", "思考题 2？"},
		Check:     "能做到什么算通关",
		Run: func() {
			fmt.Println("  演示代码，控制在 10 行内")
		},
	}
}
```

然后在同目录的 `levels.go` 里把 `L11()` 追加进 `Levels()`，`main.go` 无需改动。

## 下一站

L1-01~10 通关后，先把 [L1 · 指针与引用语义](../02-go-pointers/)（L1-11~16）打完 ——
写过 PHP/Java/Python 的人最容易在这里翻车；之后进入 [L2 · Go 并发模型](../03-go-concurrency/)（13 关），
再到 [L3 · GoFrame 主线](../04-goframe/)。
