# Go Study：从语法到 Web3 交易机器人

<p align="center">
  <strong>中文</strong> |
  <a href="README.md">English</a>
</p>

一个**关卡制**的 Go 练习仓库：目录按学习顺序编号（`01-` ~ `17-`），每个目录是一个独立 Go module。
面向「有其它语言经验、目标是自己写出 Web3 交易机器人」的学习者。

> **👉 学习路径只看一个文件：[LEARNING_PATH.md](./LEARNING_PATH.md)**
> 里面有：九个阶段与关卡数、哪些关可以跳过、每天 1 小时怎么排、终点项目的 18 个关卡拆解。

不知道从哪开始时，直接跑这两条：

```sh
make path                   # 打印阶段路线表
cd 01-go-fundamentals && go run . list   # 看当前阶段有哪些关
```

## 目录即顺序

| 编号目录 | 阶段 | 内容 | 关卡化 |
| :--- | :--- | :--- | :--- |
| [`01-go-fundamentals`](./01-go-fundamentals/) | L1 | 函数、包、接口（10 关，一关一文件） | ✅ |
| [`02-go-pointers`](./02-go-pointers/) | L1 | 指针、值拷贝与引用语义 | ⬜ |
| [`03-go-concurrency`](./03-go-concurrency/) | L2 | goroutine、channel、select、Mutex、WaitGroup | ⬜ |
| [`04-go-data-structures`](./04-go-data-structures/) | L3 | slice/map/struct/Set | ⬜ |
| [`05-go-algorithms`](./05-go-algorithms/) | L3 | 排序与基础算法 | ⬜ |
| [`06-stdlib-http`](./06-stdlib-http/) | L4 | `net/http` 客户端 / 服务端 / 进阶 | ⬜ |
| [`07-stdlib-logger`](./07-stdlib-logger/) | L4 | 标准 `log` 与日志组织 | ⬜ |
| [`08-go-mysql-example`](./08-go-mysql-example/) | L5 | `database/sql` 操作 MySQL | ⬜ |
| [`09-go-redis-example`](./09-go-redis-example/) | L5 | `go-redis` 缓存 / 去重 / 限流 | ⬜ |
| [`10-go-mongodb-example`](./10-go-mongodb-example/) | L5 | MongoDB 官方驱动 CRUD | ⬜ |
| [`11-gin-framework-example`](./11-gin-framework-example/) | L6 | Gin 分层脚手架（配置/中间件/GORM/Redis/NSQ/日志） | ⬜ |
| [`12-echo-framework-example`](./12-echo-framework-example/) | L6 | Echo + `cmd`/`internal` 目录结构 | ⬜ |
| [`13-go-jwt-example`](./13-go-jwt-example/) | L6 | JWT 签发与校验 | ⬜ |
| [`14-asynq`](./14-asynq/) | L7 | Redis 任务队列：优先级/延迟/唯一约束/定时/重试/幂等/优雅退出 | ⬜ |
| [`15-go-zero-example`](./15-go-zero-example/) | L7 | go-zero 微服务：API + RPC + goctl 代码生成 | ⬜ |
| [`16-go-telegram-bot-example`](./16-go-telegram-bot-example/) | L8 | TG Bot 通知与报表（机器人的告警通道） | ⬜ |
| [`17-go-web3-example`](./17-go-web3-example/) | L8 | 以太坊 RPC 查询（机器人的链上只读起点） | ⬜ |

待建：`18-go-websocket-market`（行情）、`19-trading-bot-core`（策略/风控/下单）、`20-deploy-ops`（上线运维）。

## 运行方式

- **Go 版本**：`go 1.25+`；根目录 [`go.work`](./go.work) 聚合全部模块，[`Makefile`](./Makefile) 按模块遍历。
- **关卡导航**（推荐用法）：

  ```sh
  go run . list     # 关卡目录（编号 · 前置 · 目标）
  go run . 3        # 只跑第 3 关
  go run . 3-5      # 跑一组
  go run .          # 全跑（复习）
  ```

- 其他常用命令：

  ```sh
  make path                              # 打印学习路线
  make build / make vet                  # 编译 / 检查全部模块（产物进 bin/）
  make run M=./14-asynq                  # 运行入口在模块根目录的示例
  make run M=./11-gin-framework-example/src P=./cmd   # 入口在子目录时指定 P
  ```

## 依赖服务与环境变量

| 服务 | 地址 | 相关目录 |
| :--- | :--- | :--- |
| Redis | `127.0.0.1:6379` | `09-`、`14-`、`11-` |
| MySQL | `127.0.0.1:3306`（`08-`）/ `127.0.0.1:3308`（`11-` dev 配置） | `08-`、`11-` |
| MongoDB | `mongodb://localhost:27017` | `10-` |
| etcd | `127.0.0.1:2379` | `15-go-zero-example`（RPC 注册；API 侧默认直连） |
| nsqd / nsqlookupd | `127.0.0.1:4150` / `127.0.0.1:4161` | `11-` |
| HTTP 端口 | `8080~8085`、`8888`（greet）、`8889`（user-api）、`1323`（echo） | 对应目录 |

| 变量 | 用途 |
| :--- | :--- |
| `TG_KEYS1` / `CHAT_ID1` | Telegram Bot Token 与 Chat ID（`16-`） |
| `ETH_RPC_URL` | 以太坊 RPC 节点（`17-`） |

## 仓库约定

- **目录编号 = 学习顺序**；新增模块请分配下一个编号，并在 `LEARNING_PATH.md` 的阶段表里登记。
- 一个目录 = 一个独立 Go module；**模块路径保持原样**（如 `go-study/Asynq`、`go-study/go-fundamentals`），
  不随目录改名，避免大面积改 import。
- 关卡规范：一关一文件 `NN_主题.go`（30~50 行）→ 同目录 `levels.go` 注册 → `main.go` 只做导航；
  `level/level.go` 零依赖，新模块复制一份。详见 [LEARNING_PATH.md 第五节](./LEARNING_PATH.md)。
- 二进制统一输出 `bin/`（已 gitignore）；`go.work` 提交，`go.work.sum` 不提交。
