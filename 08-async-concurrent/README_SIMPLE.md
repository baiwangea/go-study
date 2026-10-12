# 08-async-concurrent：直观学习模式

> **async 并发工具库 - 从 5 行代码到生产级并发系统**

## 🎯 这个模块是什么？

**async** 是一个强大的 Go 并发工具库，让你用**链式 API**轻松实现复杂的并发场景。

**一句话**：手写协程池要 50+ 行代码，async 只需 5 行！

**GitHub**: https://github.com/chichengyu/async

---

## 📚 学习内容（8 关 - 已创建）

### 基础篇（A-01 ~ A-03）

| 关卡 | 文件 | 内容 | 场景 |
|------|------|------|------|
| A-01 | `01_map_basic.go` | Map 并发数据处理 | 批量数据处理 |
| A-02 | `02_foreach.go` | ForEach 并发遍历 | 批量通知/写文件 |
| A-03 | `03_group.go` | Group 任务组 | 多任务并发执行 |

### 进阶篇（A-04 ~ A-06）

| 关卡 | 文件 | 内容 | 场景 |
|------|------|------|------|
| A-04 | `04_retry.go` | Retry 重试机制 | 下游 API 保护 |
| A-05 | `05_rate_limiter.go` | RateLimiter 限流 | 防止打爆下游 |
| A-06 | `06_pipeline.go` | Pipeline 管道 | 千万级数据 ETL |

### 实战篇（A-07 ~ A-08）

| 关卡 | 文件 | 内容 | 场景 |
|------|------|------|------|
| A-07 | `07_bounded_runner.go` | BoundedRunner | 千万级并发管控 |
| A-08 | `08_real_crawler.go` | 实战：爬虫系统 | 百万 URL 爬取 |

---

## 🚀 快速开始

```bash
# 1. 进入目录
cd 08-async-concurrent

# 2. 从第一关开始
go run 01_map_basic.go
go run 02_foreach.go
go run 03_group.go
# ... 依此类推
```

---

## 💡 核心场景速查

### 1. 批量数据处理
```go
// Map: 并发处理，收集结果
results := async.Map(ctx, items, async.IO(), func(ctx context.Context, item Item) (Result, error) {
    return process(item)
})

// ForEach: 并发处理，不收集结果
async.ForEach(ctx, items, async.IO(), func(ctx context.Context, item Item) error {
    sendNotification(item)
    return nil
})
```
→ 学习：`A-01` + `A-02`

---

### 2. 下游 API 保护
```go
// Retry + RateLimiter
limiter := async.NewTokenBucketRateLimiter(10, time.Second, 20)

result, err := async.Retry(ctx, func(ctx context.Context) (Data, error) {
    limiter.Wait(ctx)
    return callAPI()
}, async.WithMaxRetries(3))
```
→ 学习：`A-04` + `A-05`

---

### 3. 千万级数据 ETL
```go
// Pipeline + AutoScale
pipeline := async.NewPipeline[Data]().
    WithAutoScale(10, 100, 50).
    AddStage("extract", extractData).
    AddStage("transform", transformData).
    AddStage("load", loadToDB)

results := pipeline.Run(ctx, bigData)
```
→ 学习：`A-06`

---

### 4. 百万级并发管控
```go
// BoundedRunner
runner := async.NewBoundedRunner(1000).
    WithQueueSize(10000).
    WithBackpressure(true)

for _, url := range millionURLs {
    runner.Submit(func() error {
        return crawl(url)
    })
}

runner.Wait()
```
→ 学习：`A-07`

---

## 🎓 学习路径建议

### 第 1 天：数据处理（2 小时）
- A-01: Map 并发处理
- A-02: ForEach 遍历
- A-03: Group 任务组

**目标**：掌握基础并发数据处理

---

### 第 2 天：保护机制（2 小时）
- A-04: Retry 重试
- A-05: RateLimiter 限流

**目标**：保护下游服务

---

### 第 3 天：高级场景（2-3 小时）
- A-06: Pipeline 管道
- A-07: BoundedRunner

**目标**：处理大规模数据

---

### 第 4 天：实战项目（3-4 小时）
- A-08: 完整爬虫系统

**目标**：构建生产级系统

---

## 📊 性能对比

| 场景 | 串行 | async 并发 | 提升 |
|------|------|-----------|------|
| 100 个 API 调用 | 10s | 1s | 10x |
| 1000 个文件处理 | 50s | 5s | 10x |
| 10000 条数据库写 | 100s | 10s | 10x |
| 100 万数据 ETL | 5 小时 | 30 分钟 | 10x |

---

## 🆚 与标准库对比

| 功能 | 标准库 | async | 优势 |
|------|--------|-------|------|
| 协程池 | 50+ 行手写 | 5 行 | 简洁 |
| Map/Reduce | 手写循环 | 1 行 | 优雅 |
| 任务组 | errgroup | 内置 | 零依赖 |
| 重试 | 手写循环 | 声明式 | 可读 |
| 限流 | x/time/rate | 内置 | 零依赖 |

---

## 🔧 适用场景一览表

| 场景 | 推荐组件 | 关卡 |
|------|---------|------|
| Web 服务的并发请求处理 | Map/ForEach | A-01, A-02 |
| 批量数据处理 | Map/MapChain/ForEach | A-01, A-02 |
| 千万级数据 ETL | Pipeline + AutoScale | A-06 |
| 下游 API 保护性调用 | Retry + RateLimiter | A-04, A-05 |
| 实时流式消费 | Pipeline + Streaming | A-06 |
| 千万级 goroutine 并发管控 | BoundedRunner | A-07 |
| 热点 Key 按用户隔离 | ShardedPool/ShardedGroup | 高级 |
| 单 Pool 吞吐量达到瓶颈 | MultiPool（水平分片） | 高级 |

---

## ⚠️ 常见问题

### Q: async 和标准库的区别？
A: 
- 标准库：需要手写 channel、goroutine、WaitGroup
- async：链式 API，5 行搞定

### Q: 需要 Go 版本？
A: Go 1.18+（支持泛型）

### Q: 有外部依赖吗？
A: **零依赖**！只依赖标准库

### Q: async.IO() 和 async.CPU() 的区别？
A:
- IO()：IO 密集型（网络、文件），默认并发度 = CPU 核数 * 2
- CPU()：CPU 密集型（计算、加密），默认并发度 = CPU 核数

---

## 🎯 通关标准

### 基础篇
能独立写出：
- ✓ Map/ForEach 并发处理
- ✓ Group 任务组
- ✓ 理解 IO vs CPU 并发度

### 进阶篇
能独立实现：
- ✓ 带重试的 API 调用
- ✓ 限流保护
- ✓ 多阶段数据管道

### 实战篇
能独立构建：
- ✓ 百万级并发系统
- ✓ 完整爬虫架构

---

## 🔗 相关学习

### 前置知识
- **L2-02**: WaitGroup
- **L2-03~L2-05**: Channel
- **L2-12**: Worker Pool（对比手写）

### 对比学习
- **03-go-concurrency**: 标准库并发
- **05-trading-bot**: 实战应用

---

## 📦 安装

```bash
go get github.com/chichengyu/async@latest
```

---

## 🚀 开始学习

```bash
cd 08-async-concurrent
go run 01_map_basic.go
```

---

💡 **一句话总结**：async 让 Go 并发编程从 50 行变成 5 行，从繁琐变优雅！

🎉 **开始你的高效并发之旅吧！**
