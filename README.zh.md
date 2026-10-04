# Go Study：语法 → 并发 → GoFrame，直达 Web3 交易机器人

<p align="center">
  <strong>中文</strong> |
  <a href="README.md">English</a>
</p>

关卡制 Go 练习仓库：**一个知识点 = 一个小文件 = 一关**，目录名就是学习顺序。
面向「有多年其它语言经验、目标是自己写出 Web3 交易机器人」的学习者。

> **👉 学习路径只看一个文件：[LEARNING_PATH.md](./LEARNING_PATH.md)**
> 四个阶段、可跳过的关、每天 1 小时的排法、机器人能力倒推到关卡的映射，都在里面。

```sh
make path                              # 打印阶段路线
cd 01-go-fundamentals && go run . list # 看当前阶段有哪些关
cd 01-go-fundamentals && go run . 3    # 只跑第 3 关
```

## 四个目录

| 目录 | 阶段 | 内容 | 关卡 |
| :--- | :--- | :--- | :--- |
| [`01-go-fundamentals`](./01-go-fundamentals/) | L1 | 函数、包、接口 | ✅ 10 |
| [`02-go-pointers`](./02-go-pointers/) | L1 | 值拷贝、接收者、切片/map 共享、nil 陷阱、逃逸 | ✅ 6 |
| [`03-go-concurrency`](./03-go-concurrency/) | L2 | goroutine、WaitGroup、channel、select、context、Mutex/atomic | ✅ 13 |
| [`04-goframe`](./04-goframe/) | L3 | **GoFrame 主线**：路由/绑定校验/中间件/配置/日志/错误码/定时/异步任务/分层/交叉编译 → MySQL/Redis/JWT/TG/Web3 | 🟡 10/16 |

MySQL、Redis、日志、JWT、Telegram、Web3、任务队列**不再单独建目录**，全部作为 `04-goframe` 里的关卡
（对照表见 [`04-goframe/README.md`](./04-goframe/README.md)）。待建：`05-trading-bot`（终点项目）。

## 常用命令

```sh
make build / make vet   # 编译 / 检查全部模块（产物进 bin/）
make test               # 跑测试
make clean              # 清 bin/
make run M=./04-goframe # 运行某个模块
```

根目录 [`go.work`](./go.work) 聚合上述模块；[`Makefile`](./Makefile) 负责按模块遍历
（多模块仓库无法用一条 `go build ./...` 覆盖）。

## 关卡运行时与规范

- 文件命名 `NN_主题.go`，一关一文件，**30~50 行**
- 每关返回 `level.Level`：`ID / Title / Tags / Pre / Goal / Observe / Questions / Check / Run`
- 运行时 `level/level.go` 零依赖，新模块复制一份；`main.go` 只做导航（`list` / 单关 / 区间 / 全跑）
- 目录改名后 **Go 模块路径保持原样**（`go-study/go-concurrency` 对应目录 03-go-concurrency、`go-study/goframe` 对应目录 04-goframe），不影响历史 import

## 依赖服务与环境变量

| 服务 | 地址 | 何时需要 |
| :--- | :--- | :--- |
| MySQL | `127.0.0.1:3306` | `04-goframe` L3-07 起 |
| Redis | `127.0.0.1:6379` | L3-09、L3-11 起 |
| 以太坊 RPC | `ETH_RPC_URL` | L3-14 起 |
| Telegram | `TG_BOT_TOKEN`、`TG_CHAT_ID` | L3-13 起 |

已实现的 10 关（L3-01~06、L3-10、L3-11、L3-15、L3-16）与 L1、L2 全部**不依赖任何外部服务**，克隆下来即可 `go run`。
