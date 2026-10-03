# Go Study: Go 语言实战学习仓库

<p align="center">
  <strong>中文</strong> |
  <a href="README.md">English</a>
</p>

欢迎来到 **Go Study**，这是一个精心设计的实战示例集合，旨在加速你的 Go 语言学习之旅。本仓库中的每个模块都是一个独立的 Go 项目（各自有自己的 `go.mod`），包含详细的代码、注释以及专属的 `README.md` 文档。

## 🗺️ 学习路线

按阶段推进，前一阶段是后一阶段的基础：

| 阶段 | 目标 | 模块 |
| :--- | :--- | :--- |
| ① 语言基础 | 语法、类型系统、指针语义 | [`go-fundamentals`](./go-fundamentals/)、[`go-pointers`](./go-pointers/)、[`go-data-structures`](./go-data-structures/) |
| ② 并发编程 | goroutine 编排与同步原语 | [`go-concurrency`](./go-concurrency/) |
| ③ 标准库 | 手写 HTTP 服务与日志，理解框架底层 | [`stdlib-http`](./stdlib-http/)、[`stdlib-logger`](./stdlib-logger/) |
| ④ 算法与数据结构 | 常见排序与复杂度直觉 | [`go-algorithms`](./go-algorithms/) |
| ⑤ 存储与缓存 | SQL/NoSQL/缓存的落地方式 | [`go-mysql-example`](./go-mysql-example/)、[`go-redis-example`](./go-redis-example/)、[`go-mongodb-example`](./go-mongodb-example/) |
| ⑥ Web 框架 | 从单文件到分层工程脚手架 | [`gin-framework-example`](./gin-framework-example/)、[`echo-framework-example`](./echo-framework-example/)、[`go-jwt-example`](./go-jwt-example/) |
| ⑦ 消息与异步 | 削峰、解耦、定时与幂等 | [`Asynq`](./Asynq/)、NSQ 集成（见 gin 模块） |
| ⑧ 微服务与扩展 | 服务拆分、代码生成、第三方 API | [`go-zero-example`](./go-zero-example/)、[`go-telegram-bot-example`](./go-telegram-bot-example/)、[`go-web3-example`](./go-web3-example/) |

## 🚀 模块总览

### 核心语言特性

| 模块 | 描述 |
| :--- | :--- |
| [`go-fundamentals`](./go-fundamentals/) | 涵盖 Go 语言的基石：函数、包和接口。 |
| [`go-pointers`](./go-pointers/) | 深入讲解指针，以及“值传递”与“指针传递”的根本区别。 |
| [`go-concurrency`](./go-concurrency/) | 探索 Go 强大的并发原语：Goroutines, Channels, Select, Mutexes, 和 WaitGroups。 |
| [`go-data-structures`](./go-data-structures/) | 演示 Go 的内置数据结构（切片、映射、结构体）以及如何实现一个集合 (Set)。 |
| [`go-algorithms`](./go-algorithms/) | 基础算法实现：冒泡排序与快速排序。 |

### 标准库实战

| 模块 | 描述 |
| :--- | :--- |
| [`stdlib-http`](./stdlib-http/) | 使用 `net/http` 包构建 HTTP 客户端和服务器的实用示例（含进阶用法）。 |
| [`stdlib-logger`](./stdlib-logger/) | 使用标准 `log` 包进行有效日志记录的指南。 |

### Web 框架与工程化

| 模块 | 描述 |
| :--- | :--- |
| [`gin-framework-example`](./gin-framework-example/) | Gin 框架的完整项目脚手架：分层架构、环境化配置、JWT 中间件、GORM、Redis、验证码、日志轮转、NSQ 生产/消费、交叉编译。 |
| [`echo-framework-example`](./echo-framework-example/) | Echo 框架的现代化目录结构示例（`cmd` + `internal`），内置 Logger/Recover 中间件。 |
| [`go-zero-example`](./go-zero-example/) | go-zero 微服务三件套：`greet` API、`user` RPC、`user-api`（演示 API 调用 gRPC 与 goctl 代码生成）。 |
| [`go-jwt-example`](./go-jwt-example/) | 学习如何创建和验证用于无状态身份验证的 JSON Web Tokens (JWT)。 |

### 消息队列与异步任务

| 模块 | 描述 |
| :--- | :--- |
| [`Asynq`](./Asynq/) | 基于 Redis 的异步任务队列：优先级队列、延迟任务、唯一约束、批量投递、订单超时取消、Cron 调度、中间件、幂等与优雅退出。 |
| NSQ（位于 [`gin-framework-example`](./gin-framework-example/)） | `src/app/handler/nsq.go` 生产消息，`src/cmd/nsq-consumer` 独立消费者进程。 |

### 数据存储

| 模块 | 描述 |
| :--- | :--- |
| [`go-mysql-example`](./go-mysql-example/) | 使用标准 `database/sql` 包安全高效地操作 MySQL。 |
| [`go-mongodb-example`](./go-mongodb-example/) | 使用官方驱动对 MongoDB 数据库执行 CRUD 操作。 |
| [`go-redis-example`](./go-redis-example/) | 使用 `go-redis` 与 Redis 交互以实现缓存等功能。 |

### 扩展实战

| 模块 | 描述 |
| :--- | :--- |
| [`go-telegram-bot-example`](./go-telegram-bot-example/) | Telegram Bot 消息发送与 Excel 报表生成。 |
| [`go-web3-example`](./go-web3-example/) | 通过以太坊 RPC 查询链上信息。 |

## ⚙️ 环境与运行方式

- **Go 版本**：建议 `go 1.25+`（各模块 `go.mod` 声明为 1.18 ~ 1.25.11）。
- **工作区**：根目录的 [`go.work`](./go.work) 聚合了全部模块，因此在仓库根目录就能跨模块编译，例如
  `go build ./go-zero-example/user-api`。`go.work` 建议提交，`go.work.sum` 已被忽略。
- **统一入口**：根目录 [`Makefile`](./Makefile) 按模块遍历执行（多模块仓库无法用一条 `go build ./...` 覆盖）。

```sh
make list                 # 列出所有模块
make build                # 编译全部模块
make vet                  # 静态检查全部模块
make test                 # 运行全部模块的测试
make tidy                 # 依次 go mod tidy
make run M=./Asynq                       # 运行入口在模块根目录的示例
make run M=./gin-framework-example/src P=./cmd   # 入口在子目录时指定 P
```

也可以沿用传统方式：`cd <模块目录> && go run .`（gin 模块为 `go run src/cmd/main.go -env=dev`）。

## 🌐 依赖服务与端口一览

| 服务 | 地址 | 涉及模块 |
| :--- | :--- | :--- |
| Redis | `127.0.0.1:6379` | `go-redis-example`、`Asynq`、`gin-framework-example` |
| MySQL | `127.0.0.1:3306`（`go-mysql-example`）/ `127.0.0.1:3308`（gin dev 配置） | `go-mysql-example`、`gin-framework-example` |
| MongoDB | `mongodb://localhost:27017` | `go-mongodb-example` |
| etcd | `127.0.0.1:2379` | `go-zero-example/user`（服务注册；API 侧默认直连，无需 etcd） |
| nsqd / nsqlookupd | `127.0.0.1:4150` / `127.0.0.1:4161` | `gin-framework-example`（NSQ 生产与消费） |
| HTTP 监听端口 | `8080`~`8085`（stdlib-http、gin、user.rpc）；`8888`（greet-api）；`8889`（user-api）；`1323`（echo） | 对应模块 |

> `stdlib-http/client` 与 `stdlib-logger` 的示例需要对应的服务端先起来。

## 🔑 环境变量

需要密钥的模块统一从环境变量读取，不要硬编码到代码里：

| 变量 | 用途 | 模块 |
| :--- | :--- | :--- |
| `TG_KEYS1` | Telegram Bot Token | `go-telegram-bot-example` |
| `CHAT_ID1` | Telegram 聊天 ID | `go-telegram-bot-example` |
| `ETH_RPC_URL` | 以太坊 RPC 节点地址 | `go-web3-example` |

```sh
export TG_KEYS1="<你的 bot token>"
export CHAT_ID1="<你的 chat id>"
export ETH_RPC_URL="https://your-eth-rpc-endpoint"
```

## 💡 仓库约定

- 一个目录 = 一个独立 Go module，模块路径统一为 `go-study/<目录名>`（`gin-framework-example/src`、`user-api` 为历史遗留）。
- 模块文档以中文为主；`gin-framework-example` 与根文档提供中英双语。
- 新增模块后记得同步更新：`go.work`、根 `README.md` / `README.zh.md` 的模块表。
- 二进制产物统一输出到根目录 `bin/`（已 gitignore）。

编程愉快！ ✨
