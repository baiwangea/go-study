# 08-async-concurrent：async 并发工具库实战

> **链式 API 把「手写协程池 50 行」压成几行。** 一个文件 = 一关 = 完整 `main()`。

本目录 **8 关**，对应仓库里实际存在的文件。控制台用 🚀 / 📝 / 💡 分区。
更口语的路线图见 [`README_SIMPLE.md`](./README_SIMPLE.md)。

**GitHub**: https://github.com/chichengyu/async

## 怎么跑

```sh
cd 08-async-concurrent

go run 01_map_basic.go       # A-01 Map 并发数据处理
go run 08_real_crawler.go    # A-08 爬虫综合实战
```

仓库根目录：

```sh
make levels M=./08-async-concurrent
make level  F=01_map_basic.go
```

## 关卡表（8 关）

| 编号 | 文件 | 主题 |
| :--- | :--- | :--- |
| A-01 | `01_map_basic.go` | Map 并发处理，收集结果 |
| A-02 | `02_foreach.go` | ForEach 并发遍历，不收集结果 |
| A-03 | `03_group.go` | Group 任务组 |
| A-04 | `04_retry.go` | Retry 重试 |
| A-05 | `05_rate_limiter.go` | RateLimiter 限流 |
| A-06 | `06_pipeline.go` | Pipeline 管道 |
| A-07 | `07_bounded_runner.go` | BoundedRunner 并发上限 |
| A-08 | `08_real_crawler.go` | 爬虫：限流 + 重试 + 并发管控 |

## 目录里有什么

```
08-async-concurrent/
├── 01_map_basic.go
├── 02_foreach.go
├── 03_group.go
├── 04_retry.go
├── 05_rate_limiter.go
├── 06_pipeline.go
├── 07_bounded_runner.go
├── 08_real_crawler.go
├── README.md
└── README_SIMPLE.md
```

每个文件首行 `//go:build level`，同一目录多个 `main()` 才不会冲突。
`go run 文件名.go` 显式指定文件时仍能跑。

## 和标准库的关系

| 场景 | 标准库 | 本模块关卡 |
|------|--------|-----------|
| 批量并发处理 | 手写 goroutine + channel | A-01 Map / A-02 ForEach |
| 一组任务全成或全败 | `x/sync/errgroup` | A-03 Group |
| 重试 | 手写循环 + sleep | A-04 |
| 限流 | `x/time/rate` | A-05 |
| 多阶段流水线 | 自己串 channel | A-06 |
| 总并发上限 | 自己做 worker pool | A-07 |

前置建议：先打完 [`03-go-concurrency`](../03-go-concurrency/) 的 WaitGroup、channel、worker pool。

## 下一站

学完 A-08 后，可以把这些模式对到 [`05-trading-bot`](../05-trading-bot/) 的拉行情 / 下单 / 通知。
