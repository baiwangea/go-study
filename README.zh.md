# Go Study：语法 → 并发 → GoFrame，直达 Web3 交易机器人

<p align="center">
  <strong>中文</strong> |
  <a href="README.md">English</a>
</p>

**直观学习模式**的 Go 练习仓库：**一个文件 = 一关 = 完整的 `package main` + `func main()`**，
不需要任何导航框架，`go run` 那个文件就能跑。目录名就是学习顺序。
面向「有多年其它语言经验、目标是自己写出 Web3 交易机器人」的学习者。

```sh
cd 02-go-pointers
go run 11_value_vs_pointer.go     # 跑一关，就这么简单
```

在仓库根目录也有便捷入口（`M=` 可省略，会自动在所有模块里找；`F=` 可省略 `.go`，也支持前缀）：

```sh
make level F=11_value_vs_pointer.go     # 跑这一关
make level F=11_value_vs_pointer        # .go 可省略
make level F=15_nil                     # 前缀匹配（唯一时直接跑，多个会列出候选让你写全）
make level F=channels/02_buffered.go    # 带子目录路径也行
make levels M=./02-go-pointers          # 列出该模块所有关卡文件
```

## 四个阶段

| 阶段 | 目录 | 内容 | 关卡 |
| :--- | :--- | :--- | :--- |
| **L1 语言地基** | [`01-go-fundamentals`](./01-go-fundamentals/) | 函数、包、接口 | ✅ 10 |
| | [`02-go-pointers`](./02-go-pointers/) | 值拷贝、接收者、切片/map 共享、nil 陷阱、逃逸、PHP implements 对比 | ✅ 7 |
| **L2 并发模型** | [`03-go-concurrency`](./03-go-concurrency/) | goroutine、WaitGroup、channel、select、context、Mutex/atomic、worker pool | ✅ 13 |
| **L3 GoFrame 主线** | [`04-goframe`](./04-goframe/) | 路由/绑定校验/中间件/配置/日志/错误码/g.DB/ORM 事务/Redis/gcron/异步队列/JWT/TG/链上只读/分层/交叉编译 | ✅ 16 |
| **L4 终点项目** | [`05-trading-bot`](./05-trading-bot/) | 行情（REST 轮询 + WS 推送）→ 信号 → 风控 → 下单（超时/重试/幂等）→ 落库 → 通知 | ✅ 已跑真实数据 |

MySQL、Redis、日志、JWT、Telegram、Web3、任务队列**不再单独建目录**，全部作为 `04-goframe` 里的关卡
（对照表见 [`04-goframe/README.md`](./04-goframe/README.md)）；`05-trading-bot` 把它们接成真实链路，
是本仓库唯一一个「正常多文件项目」形态的模块，不适用单文件关卡约定。

## 怎么学

1. 进阶段目录，先 `ls` 或看该目录的 `README.md` 关卡表，挑一关；
2. `go run <那个文件>.go`，对着输出读它头部注释里的**核心概念 / 你会看到**；
3. 打开文件，**按【思考题】改代码**（改出编译错误最好，先读懂它）；
4. 合上代码，用一句话说清这一关（说不清 = 没过，明天重来）；
5. 写 3 行笔记：解决什么 / 和 PHP/Java 差在哪 / 机器人哪块会用到。

**思考题动手改过代码才算通关。**

## 常用命令

```sh
make build / make vet   # 编译 / 静态检查全部模块
make test               # 跑测试
make levels M=./02-go-pointers   # 列关卡
make level  F=11_value_vs_pointer.go   # 跑单关（M 可省略）
make run M=./04-goframe # 运行某个模块（P=./cmd 指定子目录入口）
make clean              # 清 bin/
```

根目录 [`go.work`](./go.work) 聚合上述模块；[`Makefile`](./Makefile) 负责按模块遍历
（多模块仓库无法用一条 `go build ./...` 覆盖）。

## 单文件关卡约定（直观学习模式）

- 命名 `NN_主题.go`，**一个文件 = 一关 = 一个完整的 `func main()`**，直接 `go run 该文件`
- 文件首行必须是 `//go:build level`，**不能用 `ignore`**：
  - 同一目录下有多个 `func main()`，不加标签会 `main redeclared` 编译失败；
  - `go run <文件名>.go` 显式指定文件时会**绕过构建约束**，所以照样能单独跑；
  - `go build ./...` / `go vet ./...` 会跳过这些文件，冲突消失；
  - 用 `ignore` 会让 `go mod tidy` 清空 go.mod 的 require 列表（04-goframe 会丢 goframe/mysql/redis）。
- 关卡文件头部注释就是「关卡卡」：**核心概念 / 你会看到 / 思考题 / 通关标准 / 前置标签**，
  对应关系写清楚，不再有 `level.Level` 运行时，也不再有 `main.go` / `levels.go`。
- 副作用：模块里若一个常规可编译的包都没有，`go run .` 会报
  `build constraints exclude all Go files` —— 这是设计如此，请跑单个文件。
  `make build` / `make vet` 已显式识别这种情况，不会误报失败。
- 目录改名后 **Go 模块路径保持原样**（`go-study/go-concurrency` 对应目录 03-go-concurrency、
  `go-study/goframe` 对应目录 04-goframe），不影响历史 import。
- 例外：`01-go-fundamentals/packages/helper/`（L1-06 的教具）和 `04-goframe/internal/`、
  `04-goframe/demoserver/` 是**正常的可编译包**，不带标签。

## 依赖服务与环境变量

| 服务 | 地址 | 何时需要 |
| :--- | :--- | :--- |
| MySQL | `127.0.0.1:3308`，库 `go_study` | `04-goframe` L3-07、L3-08（由 BOT_DB_LINK 环境变量提供，密码不进 Git） |
| Redis | `127.0.0.1:6379`（关卡用 db 9） | L3-09 |
| 以太坊 RPC | `ETH_RPC_URL` | L3-14 |
| Telegram | `TG_BOT_TOKEN` | L3-13、`05-trading-bot` 真发送 |
| Bitget 行情 | `https://api.bitget.com`（REST）/ `wss://ws.bitget.com/v2/ws/public`（推送），均公开无需 API Key | `05-trading-bot -feed=bitget` / `-feed=bitget-ws` |

L1、L2 与 L3 里不依赖外部服务的关卡（除 L3-07/08/09 需 MySQL 与 Redis、L3-13 需 TG token），
以及 `05-trading-bot` 的 mock 模式，克隆下来即可 `go run`；缺依赖时关卡会明确跳过并提示怎么设置，
不会 panic 也不会卡住。真实行情与落库/跨进程幂等均已实测跑过。
