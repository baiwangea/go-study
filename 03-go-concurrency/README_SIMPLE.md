# 03-go-concurrency：直观学习模式

> **每个文件独立运行，学习 Go 并发的核心概念！**

## 🎯 新的学习方式

### ✅ 现在（直观模式）
```bash
cd 03-go-concurrency
go run goroutines/01_launch.go    # 直接运行，立即看效果
```

每个文件：
- ✓ 完整的 `package main` + `main()` 函数
- ✓ 清晰的注释说明
- ✓ 实际可运行的并发示例
- ✓ 常见陷阱和最佳实践

## 📚 学习顺序（13 关）

### 基础并发（L2-01 ~ L2-02）
| 编号 | 文件 | 核心概念 | 运行命令 |
|------|------|----------|----------|
| **L2-01** | `goroutines/01_launch.go` | 启动 goroutine | `go run goroutines/01_launch.go` |
| **L2-02** | `goroutines/02_waitgroup.go` | WaitGroup 等待 | `go run goroutines/02_waitgroup.go` |

### Channel 通信（L2-03 ~ L2-06）
| 编号 | 文件 | 核心概念 | 运行命令 |
|------|------|----------|----------|
| **L2-03** | `channels/01_unbuffered.go` | 无缓冲 channel | `go run channels/01_unbuffered.go` |
| **L2-04** | `channels/02_buffered.go` | 缓冲 channel | `go run channels/02_buffered.go` |
| **L2-05** | `channels/03_close_range.go` | close 和 range | `go run channels/03_close_range.go` |
| **L2-06** | `channels/04_deadlock.go` | 死锁与泄漏 | `go run channels/04_deadlock.go` |

### Select 多路复用（L2-07 ~ L2-10）
| 编号 | 文件 | 核心概念 | 运行命令 |
|------|------|----------|----------|
| **L2-07** | `selects/01_multiplex.go` | select 基础 | `go run selects/01_multiplex.go` |
| **L2-08** | `selects/02_timeout.go` | 超时控制 | `go run selects/02_timeout.go` |
| **L2-09** | `selects/03_nonblock.go` | 非阻塞操作 | `go run selects/03_nonblock.go` |
| **L2-10** | `selects/04_ticker.go` | 定时器 | `go run selects/04_ticker.go` |

### 高级模式（L2-11 ~ L2-13）
| 编号 | 文件 | 核心概念 | 运行命令 |
|------|------|----------|----------|
| **L2-11** | `contextpkg/01_cancel.go` | Context 取消 | `go run contextpkg/01_cancel.go` |
| **L2-12** | `patterns/01_worker_pool.go` | Worker Pool | `go run patterns/01_worker_pool.go` |
| **L2-13** | `mutexes/01_race.go` | 竞态与锁 | `go run mutexes/01_race.go` |

### 扩展模式
| 文件 | 核心概念 | 运行命令 |
|------|----------|----------|
| `patterns/02_pipeline.go` | 流水线模式 | `go run patterns/02_pipeline.go` |
| `synclib/01_once.go` | sync.Once | `go run synclib/01_once.go` |
| `synclib/02_rwmutex.go` | 读写锁 | `go run synclib/02_rwmutex.go` |
| `synclib/03_pool.go` | sync.Pool | `go run synclib/03_pool.go` |
| `synclib/04_syncmap.go` | sync.Map | `go run synclib/04_syncmap.go` |
| `xlib/01_errgroup.go` | errgroup | `go run xlib/01_errgroup.go` |
| `xlib/02_rate_limit.go` | 限流 | `go run xlib/02_rate_limit.go` |
| `xlib/03_singleflight.go` | 防击穿 | `go run xlib/03_singleflight.go` |

## 🚀 快速开始

```bash
# 1. 进入目录
cd 03-go-concurrency

# 2. 按顺序学习（从 goroutine 开始）
go run goroutines/01_launch.go
go run goroutines/02_waitgroup.go
go run channels/01_unbuffered.go
# ... 依此类推

# 3. 检测竞态
go run -race mutexes/01_race.go
```

## 💡 学习建议

### 每关的学习步骤
1. **先运行**：观察并发行为
2. **看代码**：理解同步机制
3. **做实验**：
   - 去掉 WaitGroup 看会怎样
   - 忘记 close channel 会怎样
   - 不用 Mutex 会有竞态吗
4. **用 -race**：`go run -race` 检测数据竞争

### 重点关卡
- **L2-02**：WaitGroup（必须掌握）
- **L2-05**：close 和 range（常见坑）
- **L2-08**：超时控制（生产必备）
- **L2-11**：Context 取消（现代 Go 核心）
- **L2-12**：Worker Pool（并发限流模式）
- **L2-13**：竞态检测（调试必备）

## 🎓 学完这 13 关你会掌握

1. **Goroutine**：启动、等待、不泄漏
2. **Channel**：无缓冲、缓冲、close、range
3. **Select**：多路复用、超时、非阻塞
4. **Context**：取消传播、超时控制
5. **同步**：Mutex、RWMutex、atomic
6. **模式**：Worker Pool、Pipeline、errgroup

## ⚠️ 常见陷阱

1. **忘记 WaitGroup.Wait()**
   - 主 goroutine 退出 → 整个程序结束
   - 后台 goroutine 被直接抛弃

2. **忘记 close channel**
   - range 会永久阻塞
   - 导致 goroutine 泄漏

3. **在接收方 close**
   - channel 只能由发送方 close
   - 接收方 close 会 panic

4. **nil channel**
   - 读写 nil channel 永久阻塞
   - select 中可以用来禁用某个分支

5. **循环中启动 goroutine**
   - 闭包捕获循环变量
   - 旧版本必须 `i := i`

6. **数据竞争**
   - 多个 goroutine 同时读写同一变量
   - 用 `go run -race` 检测

## 🔧 调试技巧

### 检测死锁
```bash
go run channels/04_deadlock.go
# 会输出：fatal error: all goroutines are asleep - deadlock!
```

### 检测竞态
```bash
go run -race mutexes/01_race.go
# 会输出详细的 DATA RACE 报告
```

### 检测泄漏
```go
import "runtime"
fmt.Println("Goroutines:", runtime.NumGoroutine())
```

## 🎯 通关标准（整个模块）

能做到：
1. **30 秒内默写** Worker Pool 骨架
2. **给任意函数加取消能力**（用 Context）
3. **定位死锁原因**：发送无接收 / 接收无发送 / 忘 close
4. **用 -race 抓竞态**并说清修复方案

## 📊 对应交易机器人能力

| 并发能力 | 对应关卡 | 机器人场景 |
|---------|---------|-----------|
| 并发订阅多交易对 | L2-02, L2-12 | Worker Pool 拉行情 |
| 行情不阻塞下单 | L2-03, L2-07 | Channel 通信 |
| 超时控制 | L2-08 | 下单超时取消 |
| 优雅停机 | L2-11 | Context 取消传播 |
| 无竞态 | L2-13 | go run -race 验证 |

## 🚀 下一站

完成本模块后，前往：
- **L3 GoFrame 主线**：`../04-goframe/`
- L2-08（超时）、L2-11（context）、L2-12（worker pool）会在 L3 反复用到

---

💡 **提示**：
- 并发代码一定要用 `go run -race` 检测
- 别靠 time.Sleep 等结果，用 WaitGroup / channel
- 主 goroutine 退出 = 程序结束，后台全部丢弃
