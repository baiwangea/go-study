# 08-async-concurrent：async 并发工具库实战

> **基于 Go 泛型的全能并发工具库，从 5 行代码搭建协程池到生产级配置**

## 🎯 模块介绍

**async** 是一个链式 API 设计的并发工具库，覆盖 Go 后端并发编程的全部高频场景：
- ✓ 协程池（Worker Pool）
- ✓ 任务组（Task Group）
- ✓ Map/Reduce
- ✓ 重试机制
- ✓ 限流器
- ✓ 管道编排
- ✓ 水平分片
- ✓ BoundedRunner

**特点**：
- 零外部依赖
- 链式 API 优雅易用
- 基于 Go 泛型
- 从简单到复杂逐步学习

**GitHub**: https://github.com/chichengyu/async

---

## 📚 学习顺序（12 关）

### 基础篇（A-01 ~ A-04）
| 编号 | 文件 | 核心概念 | 运行命令 |
|------|------|----------|----------|
| **A-01** | `01_pool_basic.go` | 5 行搭建协程池 | `go run 01_pool_basic.go` |
| **A-02** | `01_pool_advanced.go` | 协程池高级配置 | `go run 01_pool_advanced.go` |
| **A-03** | `03_task_group.go` | 任务组并发执行 | `go run 03_task_group.go` |
| **A-04** | `04_retry.go` | 重试机制 | `go run 04_retry.go` |

### 进阶篇（A-05 ~ A-08）
| 编号 | 文件 | 核心概念 | 运行命令 |
|------|------|----------|----------|
| **A-05** | `05_rate_limiter.go` | 限流控制 | `go run 05_rate_limiter.go` |
| **A-06** | `06_map_reduce.go` | Map/Reduce 模式 | `go run 06_map_reduce.go` |
| **A-07** | `07_pipeline.go` | 管道编排 | `go run 07_pipeline.go` |
| **A-08** | `08_sharding.go` | 水平分片 | `go run 08_sharding.go` |

### 实战篇（A-09 ~ A-12）
| 编号 | 文件 | 核心概念 | 运行命令 |
|------|------|----------|----------|
| **A-09** | `09_bounded_runner.go` | BoundedRunner | `go run 09_bounded_runner.go` |
| **A-10** | `10_real_crawler.go` | 实战：爬虫系统 | `go run 10_real_crawler.go` |
| **A-11** | `11_real_batch_process.go` | 实战：批量处理 | `go run 11_real_batch_process.go` |
| **A-12** | `12_real_data_pipeline.go` | 实战：数据管道 | `go run 12_real_data_pipeline.go` |

---

## 🚀 快速开始

```bash
# 1. 进入目录
cd 08-async-concurrent

# 2. 从最简单的开始（5 行代码协程池）
go run 01_pool_basic.go

# 3. 按顺序学习每一关
go run 02_pool_advanced.go
go run 03_task_group.go
# ... 依此类推

# 4. 查看学习指南
cat README.md
```

---

## 💡 核心特性速览

### 1. 协程池（5 行代码）
```go
import "github.com/chichengyu/async"

pool := async.NewPool[int]().
    WithWorkers(10).         // 10 个 worker
    WithHandler(func(task int) error {
        fmt.Printf("处理任务 %d\n", task)
        return nil
    })

pool.Submit(1)
pool.Submit(2)
pool.Wait()
```

### 2. 任务组
```go
group := async.NewGroup().
    Add(func() error { return doTask1() }).
    Add(func() error { return doTask2() }).
    Add(func() error { return doTask3() })

if err := group.Wait(); err != nil {
    log.Fatal(err)
}
```

### 3. Map/Reduce
```go
results := async.Map(items, func(item Item) Result {
    return processItem(item)
}, async.WithConcurrency(10))

total := async.Reduce(results, func(acc, val int) int {
    return acc + val
}, 0)
```

### 4. 重试机制
```go
result, err := async.Retry(func() (string, error) {
    return callAPI()
}, async.WithMaxRetries(3), 
   async.WithBackoff(time.Second))
```

### 5. 限流器
```go
limiter := async.NewRateLimiter(10, time.Second)  // 10 req/s

limiter.Wait()
doRequest()
```

---

## 🎯 学习目标

### 基础篇目标
- ✓ 5 行代码搭建生产可用的协程池
- ✓ 理解 Worker、Task、Handler 概念
- ✓ 掌握任务组的并发执行
- ✓ 实现带重试的 API 调用

### 进阶篇目标
- ✓ 用限流器保护 API
- ✓ Map/Reduce 并行处理数据
- ✓ 管道模式组织数据流
- ✓ 水平分片处理海量数据

### 实战篇目标
- ✓ 构建高性能爬虫系统
- ✓ 批量处理百万级数据
- ✓ 搭建实时数据处理管道

---

## 📊 与标准库对比

| 场景 | 标准库实现 | async 实现 | 优势 |
|------|-----------|-----------|------|
| 协程池 | 手写 channel + goroutine（50+ 行） | 5 行链式调用 | 简洁 |
| 任务组 | errgroup（需导入） | 内置 Group | 零依赖 |
| Map/Reduce | 手写循环 + channel | 一行搞定 | 优雅 |
| 重试 | 手写循环 + sleep | 声明式配置 | 可读 |
| 限流 | golang.org/x/time/rate | 内置 | 零依赖 |

---

## 🔧 环境要求

- Go 1.18+ （支持泛型）
- 无外部依赖

---

## 🎓 通关标准

### 基础篇
能独立写出：
- ✓ 5 行代码的协程池
- ✓ 带错误处理的任务组
- ✓ 3 次重试的 API 调用

### 进阶篇
能独立实现：
- ✓ 限流的 HTTP 客户端
- ✓ 并发处理 1000 个文件
- ✓ 多阶段数据处理管道

### 实战篇
能独立构建：
- ✓ 完整的爬虫系统
- ✓ 批量数据处理系统
- ✓ 实时数据管道

---

## 🆚 与其他方案对比

### vs 手写协程池
```go
// 手写（50+ 行）
jobs := make(chan int, 100)
results := make(chan Result, 100)
var wg sync.WaitGroup

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
// ... 更多代码

// async（5 行）
pool := async.NewPool[int]().
    WithWorkers(10).
    WithHandler(process)
pool.Submit(1)
```

### vs golang.org/x/sync/errgroup
```go
// errgroup
g, ctx := errgroup.WithContext(ctx)
g.Go(func() error { return task1() })
g.Go(func() error { return task2() })
err := g.Wait()

// async（更简洁）
err := async.NewGroup().
    Add(task1).
    Add(task2).
    Wait()
```

---

## 💼 实际应用场景

### 1. API 批量调用（带限流）
```go
pool := async.NewPool[string]().
    WithWorkers(10).
    WithRateLimit(20, time.Second).  // 20 req/s
    WithRetry(3).
    WithHandler(callAPI)

for _, endpoint := range endpoints {
    pool.Submit(endpoint)
}
```

### 2. 数据 ETL 管道
```go
// Extract -> Transform -> Load
pipeline := async.NewPipeline().
    AddStage("extract", extractData).
    AddStage("transform", transformData).
    AddStage("load", loadToDatabase).
    WithConcurrency(10)

pipeline.Run(dataSources)
```

### 3. 分布式任务处理
```go
// 水平分片处理百万数据
shards := async.Shard(data, 10)  // 分 10 片

async.Map(shards, func(shard []Item) error {
    return processShard(shard)
}, async.WithConcurrency(10))
```

---

## 🚀 下一步

完成本模块后，你将掌握：
1. 高性能并发编程
2. 生产级协程池设计
3. 复杂数据处理管道
4. 分布式任务处理

可以应用到：
- 高并发 Web 服务
- 数据处理系统
- 爬虫系统
- 实时计算引擎

---

## 📖 参考资源

- **GitHub**: https://github.com/chichengyu/async
- **文档**: 查看项目 README
- **示例**: 本目录的 12 个关卡文件

---

💡 **提示**：从 A-01 开始，每关 10-15 分钟，一天 2-3 关，4-5 天完成全部！
