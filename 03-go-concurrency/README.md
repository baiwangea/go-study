# L2 · Go 并发模型（关卡制）

13 关打通 Go 并发的五件事：**goroutine、channel、select、context、锁**。
终点能力：写一个**能限流、能超时、能取消、不泄漏**的并发骨架 —— 这正是交易机器人拉行情、跑策略、下单元的数据通路。

## 怎么用

```sh
cd 03-go-concurrency

go run . list        # 关卡目录（编号 / 前置 / 目标）
go run . 3           # 只跑第 3 关
go run . 3-6         # 跑一组
go run .             # 全跑（复习用）
go run -race . 13    # L2-13 专用：看官方 DATA RACE 报告
```

## 关卡表

| 编号 | 文件 | 主题 | 前置 | 关键坑 |
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
├── main.go          # 关卡导航：list / 单关 / 区间 / 全跑
├── level/level.go   # 关卡运行时（与 01-go-fundamentals 同一份，零依赖）
├── goroutines/      # L2-01 ~ L2-02
├── channels/        # L2-03 ~ L2-06
├── selects/         # L2-07 ~ L2-10
├── contextpkg/      # L2-11
├── patterns/        # L2-12
└── mutexes/         # L2-13
```

## 本阶段结束你应该能做到

1. 30 秒内默写出 worker pool 骨架（`jobs → close → N workers → results → close → 收敛`）；
2. 给任意后台协程加「可取消」能力，并用 `runtime.NumGoroutine()` 证明没泄漏；
3. 看到 `fatal error: all goroutines are asleep` 时，能立刻定位是**发送无接收**、**接收无发送**还是**忘 close**；
4. 用 `go run -race` 抓出自己写出的竞态，并说清 Mutex / atomic / channel 的选型依据。

对应机器人里程碑（见根目录 `LEARNING_PATH.md`）：**能并发订阅多个交易对，行情不阻塞下单，且能优雅停机**。
