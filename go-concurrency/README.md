# Go 并发编程核心概念

本项目通过一系列示例，深入讲解 Go 语言的核心并发原语，包括 Goroutines, Channels, Select, Mutex, 和 WaitGroups。

## 项目结构

```
/go-concurrency
├── goroutines/            # Goroutine 基础示例
├── channels/              # Channel (通道) 的用法示例
├── selects/               # Select 语句的用法示例
├── mutexes/               # Mutex (互斥锁) 的用法示例
├── waitgroups/            # WaitGroup (等待组) 的用法示例
├── main.go                # 项目主入口，调用所有示例
└── go.mod
```

## 如何运行示例

直接运行 `main.go` 即可。程序会自动按顺序执行所有并发相关的演示函数。

```sh
go run main.go
```

---

## 示例详解

### 1. Goroutines (`goroutines/`)

Goroutine 是 Go 语言并发设计的核心。它是由 Go 运行时管理的轻量级线程。你只需在函数调用前加上 `go` 关键字，就可以在一个新的 Goroutine 中执行这个函数。

*   **优点**: 创建成本极低，可以轻松创建成千上万个 Goroutine。
*   **注意**: `main` 函数在一个特殊的 Goroutine 中运行。当 `main` Goroutine 结束时，整个程序会立即退出，而不会等待其他 Goroutine 执行完毕。这就引出了同步的需求。

### 2. WaitGroups (`waitgroups/`)

`sync.WaitGroup` 用于等待一组 Goroutine 执行完毕。它内部维护一个计数器。

*   **`wg.Add(n)`**: 将计数器增加 `n`。
*   **`wg.Done()`**: 将计数器减一。通常在 Goroutine 的末尾通过 `defer` 调用。
*   **`wg.Wait()`**: 阻塞当前的 Goroutine，直到计数器归零。

这是实现“等待所有任务完成”这一常见并发模式的标准方法。

### 3. Mutexes (`mutexes/`)

互斥锁 (`sync.Mutex`) 用于保护共享资源，防止多个 Goroutine 同时访问和修改数据而导致的竞争条件 (Race Condition)。

*   **`mu.Lock()`**: 获取锁。如果锁已经被其他 Goroutine持有，则当前 Goroutine 会被阻塞，直到锁被释放。
*   **`mu.Unlock()`**: 释放锁。
*   **最佳实践**: 总是使用 `defer mu.Unlock()` 来确保锁在函数返回时一定会被释放，即使函数发生了 `panic`。

### 4. Channels (`channels/`)

通道是 Go 的一句名言 “不要通过共享内存来通信，而要通过通信来共享内存” 的核心实践。它是在 Goroutine 之间安全传递数据的管道。

*   **无缓冲通道 `make(chan T)`**: 发送方和接收方必须同时准备好，否则就会阻塞。这是一种强同步机制。
*   **有缓冲通道 `make(chan T, capacity)`**: 发送方可以连续发送 `capacity` 个数据而不会阻塞。当缓冲区满时，发送方才会阻塞。当缓冲区空时，接收方会阻塞。
*   **通道方向**: 你可以在函数参数中指定通道是只读 (`<-chan T`) 还是只写 (`chan<- T`)，这可以增强类型安全。

### 5. Selects (`selects/`)

`select` 语句让一个 Goroutine 可以同时等待多个通道操作。它有点像网络编程中的 `select` 或 `poll`。

*   **阻塞**: `select` 会一直阻塞，直到其中一个 `case` 的通道操作准备就绪。
*   **随机选择**: 如果有多个 `case` 同时就绪，`select` 会随机选择一个执行，以避免饥饿问题。
*   **超时处理**: `select` 与 `time.After` 结合使用，可以非常优雅地实现操作超时。如果其他 `case` 在指定时间内没有就绪，`time.After` 的 `case` 就会被执行。
