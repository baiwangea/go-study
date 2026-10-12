# 08-async-concurrent：直观学习模式

> **async 并发工具库 - 从 5 行代码到生产级并发系统**

## 🎯 这个模块是什么？

**async** 是一个强大的 Go 并发工具库，让你用**链式 API**轻松实现复杂的并发场景。

**一句话**：手写协程池要 50+ 行代码，async 只需 5 行！

---

## 📚 学习内容（12 关）

### 基础：协程池与任务（必学）

| 关卡 | 文件 | 内容 | 难度 |
|------|------|------|------|
| A-01 | `01_pool_basic.go` | 5 行代码搭建协程池 | ⭐ |
| A-02 | `02_pool_advanced.go` | 协程池高级配置 | ⭐⭐ |
| A-03 | `03_task_group.go` | 任务组并发执行 | ⭐⭐ |
| A-04 | `04_retry.go` | 重试机制 | ⭐⭐ |

### 进阶：限流与数据处理

| 关卡 | 文件 | 内容 | 难度 |
|------|------|------|------|
| A-05 | `05_rate_limiter.go` | 限流保护 | ⭐⭐ |
| A-06 | `06_map_reduce.go` | Map/Reduce 并行 | ⭐⭐⭐ |
| A-07 | `07_pipeline.go` | 管道编排 | ⭐⭐⭐ |
| A-08 | `08_sharding.go` | 水平分片 | ⭐⭐⭐ |

### 实战：真实业务场景

| 关卡 | 文件 | 内容 | 难度 |
|------|------|------|------|
| A-09 | `09_bounded_runner.go` | BoundedRunner | ⭐⭐⭐ |
| A-10 | `10_real_crawler.go` | 爬虫系统 | ⭐⭐⭐⭐ |
| A-11 | `11_real_batch_process.go` | 批量处理 | ⭐⭐⭐⭐ |
| A-12 | `12_real_data_pipeline.go` | 数据管道 | ⭐⭐⭐⭐ |

---

## 🚀 快速开始

```bash
# 1. 进入目录
cd 08-async-concurrent

# 2. 安装依赖
go get github.com/chichengyu/async@latest

# 3. 从第一关开始
go run 01_pool_basic.go
```

---

## 💡 为什么学这个？

### 对比：手写 vs async

**场景**：10 个 worker 并发处理 100 个任务

#### 手写协程池（50+ 行）
```go
jobs := make(chan int, 100)
results := make(chan Result, 100)
var wg sync.WaitGroup

// 启动 10 个 worker
for w := 0; w < 10; w++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        for job := range jobs {
            result := process(job)
            results <- result
        }
    }()
}

// 提交任务
for i := 0; i < 100; i++ {
    jobs <- i
}
close(jobs)

// 等待完成
wg.Wait()
close(results)

// 收集结果
for result := range results {
    // ...
}
```

#### async（5 行）
```go
pool := async.NewPool[int]().
    WithWorkers(10).
    WithHandler(process)

for i := 0; i < 100; i++ {
    pool.Submit(i)
}
pool.Wait()
```

**省略了什么？**
- ✓ 创建 channel
- ✓ 创建 WaitGroup
- ✓ 启动 goroutine
- ✓ defer wg.Done()
- ✓ close channel
- ✓ 收集结果

**代码量**：50 行 → 5 行，减少 **90%**！

---

## 🎯 核心场景速查

### 1. API 批量调用
```go
// 并发调用 100 个 API，限流 20 req/s
pool := async.NewPool[string]().
    WithWorkers(10).
    WithRateLimit(20, time.Second).
    WithHandler(callAPI)

for _, url := range urls {
    pool.Submit(url)
}
pool.Wait()
```
→ 学习：`A-01` + `A-05`

---

### 2. 数据批量处理
```go
// 并发处理 1000 个文件
results := async.Map(files, func(file string) error {
    return processFile(file)
}, async.WithConcurrency(20))
```
→ 学习：`A-06`

---

### 3. 带重试的 API 调用
```go
// 最多重试 3 次，指数退避
result, err := async.Retry(func() (Data, error) {
    return callAPI()
}, async.WithMaxRetries(3),
   async.WithBackoff(time.Second))
```
→ 学习：`A-04`

---

### 4. 多任务并发执行
```go
// 并发执行 3 个任务，任一失败则全部失败
err := async.NewGroup().
    Add(task1).
    Add(task2).
    Add(task3).
    Wait()
```
→ 学习：`A-03`

---

### 5. 数据处理管道
```go
// Extract -> Transform -> Load
pipeline := async.NewPipeline().
    AddStage("extract", extractData).
    AddStage("transform", transformData).
    AddStage("load", loadToDB)

pipeline.Run(sources)
```
→ 学习：`A-07`

---

## 🎓 学习路径建议

### 第 1 天：基础协程池（2-3 小时）
- A-01: 5 行代码协程池
- A-02: 协程池高级配置
- A-03: 任务组

**目标**：能用 async 替换手写协程池

---

### 第 2 天：重试与限流（2-3 小时）
- A-04: 重试机制
- A-05: 限流器

**目标**：保护 API 调用

---

### 第 3 天：数据处理（2-3 小时）
- A-06: Map/Reduce
- A-07: 管道编排
- A-08: 水平分片

**目标**：并行处理海量数据

---

### 第 4-5 天：实战项目（4-6 小时）
- A-09: BoundedRunner
- A-10: 爬虫系统
- A-11: 批量处理
- A-12: 数据管道

**目标**：构建完整并发系统

---

## 📊 性能提升

| 场景 | 串行耗时 | async 并发（10 workers） | 提升 |
|------|---------|-------------------------|------|
| 100 个 API 调用（100ms/个） | 10 秒 | 1 秒 | 10x |
| 1000 个文件处理（50ms/个） | 50 秒 | 5 秒 | 10x |
| 10000 条数据库写入（10ms/条） | 100 秒 | 10 秒 | 10x |

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

### Q: 适合什么场景？
A:
- ✓ API 批量调用
- ✓ 数据批量处理
- ✓ 爬虫系统
- ✓ ETL 管道
- ✓ 任何需要并发的场景

---

## 🔗 相关学习

### 前置知识
- **L2-02**: WaitGroup（了解底层原理）
- **L2-03~L2-05**: Channel（了解通信机制）
- **L2-12**: Worker Pool（对比手写实现）

### 对比学习
| 标准库 | async 等价 | 优势 |
|--------|-----------|------|
| goroutine + channel + wg | `Pool` | 简洁 |
| golang.org/x/sync/errgroup | `Group` | 零依赖 |
| golang.org/x/time/rate | `RateLimiter` | 零依赖 |
| 手写 Map/Reduce | `Map/Reduce` | 声明式 |

---

## 🎯 通关标准

### 基础篇
能独立写出：
- ✓ 5 行代码协程池
- ✓ 带重试的 API 调用
- ✓ 限流的并发任务

### 进阶篇
能独立实现：
- ✓ Map/Reduce 数据处理
- ✓ 多阶段管道
- ✓ 水平分片处理

### 实战篇
能独立构建：
- ✓ 完整爬虫系统
- ✓ 批处理引擎
- ✓ 实时数据管道

---

## 📦 安装

```bash
go get github.com/chichengyu/async@latest
```

---

## 🚀 开始学习

```bash
cd 08-async-concurrent
go run 01_pool_basic.go
```

---

💡 **一句话总结**：async 让 Go 并发编程从 50 行变成 5 行，从繁琐变优雅！
