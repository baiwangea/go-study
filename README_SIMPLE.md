# Go Study：直观学习模式总指南

> **从语法到并发到框架，再到交易机器人，每个文件独立运行！**

## 🎉 重构完成状态

**所有核心模块已完成直观化重构**：
- ✅ **01-go-fundamentals** (11 关) - Go 语法基础
- ✅ **02-go-pointers** (7 关) - 指针与引用语义
- ✅ **03-go-concurrency** (13 关) - 并发模型
- ✅ **04-goframe** (16 关) - GoFrame 框架
- 🎯 **05-trading-bot** - 交易机器人终点项目

## 🎯 新的学习方式

### ✅ 现在（直观模式）
```bash
cd 02-go-pointers
go run 11_value_vs_pointer.go    # 直接运行任何文件
```

**特点**：
- 每个文件有完整的 `main()` 函数
- 清晰的顶部注释说明
- 格式化的输出结果
- 零框架依赖

### ❌ 原来（框架模式）
```bash
go run . 1    # 需要理解 level 框架
```

**问题**：
- 需要理解 Level 结构体
- 代码层层嵌套
- 认知负担重

## 📚 完整学习路径

### 阶段 1：语法基础（2-3 周）

#### L1 Part 1: Go 基础 (01-go-fundamentals)
**时间**：3-5 天 | **关卡**：11 关

```bash
cd 01-go-fundamentals

# 函数 (6 关)
go run functions/01_basic_func.go
go run functions/02_multi_return.go
go run functions/03_variadic.go
go run functions/04_closure.go
go run functions/05_recursion.go
go run functions/06_defer.go

# 包管理 (2 关)
go run packages/01_export.go
go run packages/02_import.go

# 接口 (3 关)
go run interfaces/01_define.go
go run interfaces/02_implicit_impl.go
go run interfaces/03_polymorphism.go
```

**学习重点**：
- 多返回值 error 模式
- 闭包捕获变量
- 首字母大小写 = 可见性
- 接口隐式实现

**查看详细指南**：`cat 01-go-fundamentals/README_SIMPLE.md`

---

#### L1 Part 2: 指针与陷阱 (02-go-pointers)
**时间**：2-3 天 | **关卡**：7 关

```bash
cd 02-go-pointers

# 按顺序运行
go run 11_value_vs_pointer.go      # 值传递 vs 指针
go run 12_struct_copy.go           # 结构体拷贝
go run 13_receiver.go              # 方法接收者
go run 14_slice_map_shared.go      # 切片共享与失联
go run 15_nil_traps.go             # nil 陷阱 ⚠️
go run 16_escape.go                # 逃逸分析
go run 17_php_vs_go_interface.go   # PHP vs Go
```

**学习重点**：
- Go 的第一原则：一切都是拷贝
- 切片 = (指针, len, cap)
- **接口装 nil 指针 ≠ nil**（最大的坑）
- 方法接收者选择规则

**查看详细指南**：`cat 02-go-pointers/README_SIMPLE.md`

---

### 阶段 2：并发模型（2-3 周）

#### L2: 并发 (03-go-concurrency)
**时间**：2-3 周 | **关卡**：13 关（核心）+ 8 关（扩展）

```bash
cd 03-go-concurrency

# 核心路径（必学）
go run goroutines/01_launch.go        # goroutine 启动
go run goroutines/02_waitgroup.go     # 等待完成
go run channels/01_unbuffered.go      # 无缓冲 channel
go run channels/02_buffered.go        # 缓冲 channel
go run channels/03_close_range.go     # close 和 range
go run channels/04_deadlock.go        # 死锁示例
go run selects/01_multiplex.go        # select 多路复用
go run selects/02_timeout.go          # 超时控制 ⭐
go run selects/03_nonblock.go         # 非阻塞
go run selects/04_ticker.go           # 定时器
go run contextpkg/01_cancel.go        # Context 取消 ⭐
go run patterns/01_worker_pool.go     # Worker Pool ⭐
go run mutexes/01_race.go             # 竞态检测 ⭐

# 扩展路径（选学）
go run patterns/02_pipeline.go        # 流水线
go run synclib/01_once.go             # sync.Once
go run synclib/02_rwmutex.go          # 读写锁
go run synclib/03_pool.go             # sync.Pool
go run synclib/04_syncmap.go          # sync.Map
go run xlib/01_errgroup.go            # errgroup
go run xlib/02_rate_limit.go          # 限流
go run xlib/03_singleflight.go        # 防击穿
```

**学习重点**：
- goroutine 启动与等待
- channel 通信模式
- **select 超时控制**（生产必备）
- **Context 取消传播**（现代 Go 核心）
- **Worker Pool 并发限流**
- **竞态检测**：`go run -race`

**查看详细指南**：`cat 03-go-concurrency/README_SIMPLE.md`

---

### 阶段 3：Web 框架（3-4 周）

#### L3: GoFrame (04-goframe)
**时间**：3-4 周 | **关卡**：16 关

```bash
cd 04-goframe

# 第一组：核心框架（无外部依赖）
go run 01_hello_route.go          # HTTP 服务
go run 02_params_binding.go       # 参数绑定
go run 03_middleware.go           # 中间件
go run 04_config.go               # 配置
go run 05_logging.go              # 日志
go run 06_error_handling.go       # 错误处理

# 第二组：任务与鉴权（无外部依赖）
go run 10_cron_task.go            # 定时任务
go run 11_async_queue.go          # 异步队列
go run 12_jwt_auth.go             # JWT 鉴权

# 第三组：数据层（需要 MySQL 和 Redis）
export BOT_DB_LINK='mysql:root:pass@tcp(127.0.0.1:3308)/go_study'
go run 07_db_mysql.go             # MySQL
go run 08_orm_model.go            # ORM 事务
go run 09_cache_redis.go          # Redis

# 第四组：集成（需要 Token）
export TG_BOT_TOKEN='your_token'
go run 13_telegram_notify.go      # Telegram
go run 14_web3_onchain_read.go    # Web3

# 第五组：工程实践（无外部依赖）
go run 15_project_layout.go       # 分层架构
go run 16_deploy_build.go         # 交叉编译
```

**学习重点**：
- GoFrame 核心 API：g.Server、g.Cfg、g.DB、g.Redis
- 参数绑定与校验
- 中间件模式
- ORM 与事务
- 定时任务与异步队列
- JWT 鉴权
- 交叉编译

**查看详细指南**：`cat 04-goframe/README_SIMPLE.md`

---

### 阶段 4：实战项目（4-5 周）

#### L4: 交易机器人 (05-trading-bot)
**时间**：4-5 周 | **终点项目**

```bash
cd 05-trading-bot

# 运行完整链路
go run .                          # Mock 数据
go run -race .                    # 检测竞态

# 真实行情
go run . -feed=bitget             # REST 轮询
go run . -feed=bitget-ws          # WebSocket 推送

# 落库与幂等
export BOT_DB_LINK='...'
export BOT_REDIS_ADDR='127.0.0.1:6379'
go run . -ticks=16
```

**查看详细说明**：`cat 05-trading-bot/README.md`

---

## 📖 各模块学习指南

每个模块都有详细的 `README_SIMPLE.md`：

| 模块 | 指南位置 | 内容 |
|------|---------|------|
| Go 基础 | `01-go-fundamentals/README_SIMPLE.md` | 函数、包、接口 |
| 指针 | `02-go-pointers/README_SIMPLE.md` | 值/指针、切片、nil 陷阱 |
| 并发 | `03-go-concurrency/README_SIMPLE.md` | goroutine、channel、select、context |
| 框架 | `04-goframe/README_SIMPLE.md` | Web、数据库、任务、鉴权 |
| 机器人 | `05-trading-bot/README.md` | 完整项目 |

---

## 💡 学习建议

### 每天的学习节奏（1 小时）
| 时间 | 动作 |
|------|------|
| 0-5 min | 看昨日笔记，确认今天要学哪一关 |
| 5-15 min | 运行文件，看输出 |
| 15-35 min | **按思考题改代码**（必须动手） |
| 35-50 min | 合上代码，用一句话总结 |
| 50-60 min | 写 3 行笔记 |

### 学习节奏
- **周一~周五**：每天 1-2 关
- **周六**：合成一个小工具（周作业）
- **周日**：复习或休息

### 完成时间（每天 1 小时）
- L1：3 周（18 关）
- L2：3 周（13 关）
- L3：3-4 周（16 关）
- L4：4-5 周（终点项目）
- **总计**：11-13 周

---

## 🔧 通用工具

### 检测竞态
```bash
go run -race 文件名.go
```

### 查看 goroutine 数
```go
import "runtime"
fmt.Println("Goroutines:", runtime.NumGoroutine())
```

### 逃逸分析
```bash
go build -gcflags='-m' .
```

---

## 🎯 里程碑验收

### L1 末（语法基础）
能手写：
- 返回 `(结果, error)` 的函数
- 接口存不同实现
- 指针修改生效
- 解释 nil 陷阱

### L2 末（并发模型）
能做到：
- 30 秒内默写 Worker Pool
- 用 Context 给任意函数加取消
- `go run -race` 无警告
- 定位死锁原因

### L3 末（Web 框架）
能独立搭出：
- g.Cfg + glog + gerror
- g.DB 落订单
- g.Redis 去重限频
- gcron 轮询
- JWT 鉴权接口
- 交叉编译 Linux 静态二进制

### L4 末（交易机器人）
能做到：
- 本地跑通「行情 → 信号 → 风控 → 下单 → 通知」
- `go run -race` 无告警
- 说清幂等与冷却为什么缺一不可

---

## 📊 重构效果对比

| 维度 | 原框架模式 | 新直观模式 |
|------|-----------|-----------|
| **学习曲线** | 需要先理解框架 | 直接看代码 |
| **认知负担** | Level、Run、闭包 | 只有 main() |
| **运行方式** | `go run . 1` | `go run 文件名.go` |
| **调试** | 要找到 Run 里的代码 | 直接在 main() 里 |
| **修改** | 改完不确定能否单独运行 | 改完立即运行验证 |
| **分享** | 要解释怎么跑这个框架 | 发一个文件就能跑 |

---

## ⚠️ 常见问题

### Q: //go:build level 是什么？
A: 构建标签，允许同一目录有多个 `main()` 函数。单独运行文件时会忽略。

### Q: 可以删掉原来的 main.go 和 level/ 吗？
A: 可以！新模式完全不需要它们，保留也不影响学习。

### Q: 思考题的答案在哪里？
A: 没有标准答案！**动手改代码验证**才是正确方式。

### Q: 我改坏了代码怎么恢复？
A: `git checkout 文件名` 或重新 clone 项目。

---

## 🚀 快速开始

```bash
# 克隆项目
cd /Users/cluo/code/go/src/go-study

# 从第一关开始
cd 01-go-fundamentals
go run functions/01_basic_func.go

# 查看模块指南
cat README_SIMPLE.md
```

---

## 🎓 学习原则

1. **别只看，必须改**：思考题没动手改过，不算通关
2. **改出错误是好事**：先读懂编译错误信息
3. **用一句话总结**：说不清 = 没过，明天重来
4. **并发代码必须 -race**：检测数据竞争
5. **不懂就问，不要硬猜**：看代码 > 看文档 > 问人

---

💡 **核心思想**：学习 Go 应该专注于语言本身，而不是理解学习框架！

🎉 **开始你的 Go 学习之旅吧！**
