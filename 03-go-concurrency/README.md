# L2 · Go 并发模型（直观学习模式）

13 关打通 Go 并发的五件事：**goroutine、channel、select、context、锁**。
终点能力：写一个**能限流、能超时、能取消、不泄漏**的并发骨架 —— 这正是交易机器人拉行情、跑策略、下单元的数据通路。

## 怎么用

**一个文件 = 一关 = 一个完整的 `main()`。不需要导航框架，直接运行那个文件。**

```sh
cd 03-go-concurrency

go run goroutines/01_launch.go        # L2-01 go 语句与调度
go run channels/01_unbuffered.go      # L2-03 无缓冲 channel（可直接跳关）
go run patterns/01_worker_pool.go     # L2-12 worker pool 骨架

go run -race mutexes/01_race.go       # L2-13：看官方 DATA RACE 报告
```

在仓库根目录也可以：

```sh
make levels M=./03-go-concurrency                # 列出本模块所有关卡文件
make level  F=channels/01_unbuffered.go          # 跑指定关卡（M 可省略）
```

每关文件的头部注释就是「关卡卡」：**核心概念 / 你会看到 / 思考题 / 通关标准 / 前置**。
**思考题动手改过代码，才算通关**，可以进入下一关。

## 关卡表

| 编号 | 文件（`go run <文件>`） | 主题 | 前置 | 关键坑 |
| :--- | :--- | :--- | :--- | :--- |
| L2-01 | `goroutines/01_launch.go` | `go` 语句与调度不确定性 | L1-04 | 主协程不等，后台直接被抛弃 |
| L2-02 | `goroutines/02_waitgroup.go` | WaitGroup 的 Add/Done/Wait 位置 | L2-01 | `Add` 写进 goroutine 里 → 竞态 |
| L2-03 | `channels/01_unbuffered.go` | 无缓冲 channel = 同步会合 | L2-02 | 无人接收的发送会一直堵 |
| L2-04 | `channels/02_buffered.go` | 缓冲、`len/cap`、背压 | L2-03 | 队列满 → 生产者被拖慢 |
| L2-05 | `channels/03_close_range.go` | close / range / 逗号 ok | L2-04 | 忘记 close → 永久阻塞；接收方关 → panic |
| L2-06 | `channels/04_deadlock.go` | 死锁、goroutine 泄漏、nil channel | L2-05 | fatal error 不可 recover |
| L2-07 | `selects/01_multiplex.go` | select 多路复用与随机公平 | L2-06 | 多分支就绪时别假设顺序 |
| L2-08 | `selects/02_timeout.go` | `time.After` 超时快速失败 | L2-07 | 超时返回 ≠ 后端没执行 |
| L2-09 | `selects/03_nonblock.go` | default 非阻塞与丢弃策略 | L2-08 | 不可丢的数据（订单回报）不能这么写 |
| L2-10 | `selects/04_ticker.go` | Ticker 轮询与 `Stop` | L2-09 | 不 Stop → timer 泄漏 |
| L2-11 | `contextpkg/01_cancel.go` | context 取消与协程回收 | L2-10 | 忘记 `cancel()` → 下游永远收不到令 |
| L2-12 | `patterns/01_worker_pool.go` | **Worker Pool 并发限流** | L2-02, L2-05 | 先 `close(jobs)` 再 `Wait` |
| L2-13 | `mutexes/01_race.go` | data race、Mutex 与 atomic 选型 | L2-12 | 无锁自增会丢计数 |

> `contextpkg` 目录名不叫 `context`，是为了不遮挡标准库 `context`。

## 目录结构

```
03-go-concurrency/
├── go.mod
├── README.md
├── goroutines/01_launch.go            # L2-01  ← 一个文件一关，每关自带 func main()
├── goroutines/02_waitgroup.go         # L2-02
├── channels/01_unbuffered.go          # L2-03
├── channels/02_buffered.go            # L2-04
├── channels/03_close_range.go         # L2-05
├── channels/04_deadlock.go            # L2-06
├── selects/01_multiplex.go            # L2-07
├── selects/02_timeout.go              # L2-08
├── selects/03_nonblock.go             # L2-09
├── selects/04_ticker.go               # L2-10
├── contextpkg/01_cancel.go            # L2-11
├── patterns/01_worker_pool.go         # L2-12
└── mutexes/01_race.go                 # L2-13
```

**没有 `main.go`、没有 `level/` 运行时、没有 `levels.go` 注册表、没有占位文件。**
目录里除了关卡文件和 go.mod，只有这份 README。

## 每个文件首行的 `//go:build level`，为什么必须留着

Go 规定同一个目录下 `package main` 只能有一个 `func main()`。本模块里 `channels/`、
`selects/` 等目录各自有多个自带 `main()` 的文件，不加标签就会编译失败：

```
./channels/02_buffered.go:31:6: main redeclared in this block
        ./channels/01_unbuffered.go:31:6: other declaration of main
```

加了 `//go:build level` 之后：

- `go run channels/01_unbuffered.go` —— **显式指定文件名时会绕过构建约束**，照样能单独跑；
- `go build ./...` / `go vet ./...` —— 自动跳过这些文件，冲突消失。

> **不要换成 `//go:build ignore`。** 实测 `ignore` 会让 `go mod tidy` 认为模块里
> 没有任何代码，从而**清空 go.mod 的 require 列表**。自定义标签 `level` 没有这个问题。

## 两个附带的注意点

**1. `go run .` 不再可用。** 关卡文件全被标签排除后，模块根目录在常规构建下
「匹配不到任何包」，`go run .` 会报 `build constraints exclude all Go files`。
这不是坏了，而是本模块的设计：**跑单个文件，而不是跑整个目录**。

```sh
go run channels/01_unbuffered.go   # ✅ 正确用法
go run .                           # ❌ 本模块没有「整个目录」这个入口
```

**2. 根目录的 `make build` / `make vet` 已经能正确识别这种情况。**
`go build ./...` 和 `go vet ./...` 在「匹配不到包」时都会以退出码 1 结束，
所以 `Makefile` 里对这两个目标加了显式放行（见 `NOPKG` 变量），
输出会显示 `OK (本模块没有可编译的命令行程序)`。这不是掩盖错误 ——
真正的编译/静态检查失败依然会 `FAIL` 并中断。

## 本阶段结束你应该能做到

1. 30 秒内默写出 worker pool 骨架（`jobs → close → N workers → results → close → 收敛`）；
2. 给任意后台协程加「可取消」能力，并用 `runtime.NumGoroutine()` 证明没泄漏；
3. 看到 `fatal error: all goroutines are asleep` 时，能立刻定位是**发送无接收**、**接收无发送**还是**忘 close**；
4. 用 `go run -race mutexes/01_race.go` 抓出自己写出的竞态，并说清 Mutex / atomic / channel 的选型依据。

对应机器人里程碑：**能并发订阅多个交易对，行情不阻塞下单，且能优雅停机**。

## 下一站

[L3 · GoFrame 主线](../04-goframe/)。本阶段的 L2-08（超时）、L2-11（context 取消）、
L2-12（worker pool）会在那边反复用到 —— 框架里的限频重试、优雅停机、并发拉行情都靠它们。
