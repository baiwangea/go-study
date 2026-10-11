# L3 · GoFrame 工程主线（直观学习模式）

语法与并发打完（L1、L2），**从这里开始只学一个框架：GoFrame**。
原先散落成独立目录的 mysql / redis / 日志 / jwt / telegram / web3 / 任务队列，
全部变成本框架里的关卡 —— 因为它们本来就要在一个工程里协同工作。

## 怎么用

**一个文件 = 一关 = 一个完整的 `main()`。没有 `main.go`、没有 `level/` 运行时、没有关卡注册表，直接运行那个文件。**

```sh
cd 04-goframe

go run 01_hello_route.go        # L3-01 第一个 GoFrame 服务
go run 12_jwt_auth.go           # L3-12 手写 JWT（可直接跳关）
go run 16_deploy_build.go       # L3-16 交叉编译
```

在仓库根目录也可以：

```sh
make levels M=./04-goframe                       # 列出本模块所有关卡文件
make level  F=01_hello_route.go                  # 跑指定关卡（M 可省略）
```

每关文件的头部注释就是「关卡卡」：**核心概念 / 你会看到 / 思考题 / 通关标准 / 前置 / 标签**。
**思考题动手改过代码，才算通关**，可以进入下一关。

## 关卡表（16 关，已全部实现）

| 编号 | 文件（`go run <文件>`） | 主题 | 前置 | 外部依赖 |
| :--- | :--- | :--- | :--- | :--- |
| L3-01 | `01_hello_route.go` | `g.Server` 实例、路由注册、优雅关闭 | L2-11 | 无（随机端口起临时服务） |
| L3-02 | `02_params_binding.go` | `r.Parse` 绑定 + `v` 校验 + 中文文案 | L3-01 | 无 |
| L3-03 | `03_middleware.go` | 标准出入参、`g.Meta`、统一响应中间件 | L3-02 | 无 |
| L3-04 | `04_config.go` | `g.Cfg` 与 `manifest/config/config.yaml` | L3-01 | 无 |
| L3-05 | `05_logging.go` | `glog` 分级、字段、trace id、按天落盘 | L3-04 | 无 |
| L3-06 | `06_error_handling.go` | `gerror` 错误码、Wrap、堆栈 | L1-02, L3-05 | 无 |
| L3-07 | `07_db_mysql.go` | `g.DB()` 建表、插入、条件查询、聚合 | L3-04 | `BOT_DB_LINK`（MySQL） |
| L3-08 | `08_orm_model.go` | ORM 链式、事务与回滚 | L3-07 | `BOT_DB_LINK`（MySQL） |
| L3-09 | `09_cache_redis.go` | `g.Redis()` 缓存、INCR 限频、SetNX 幂等 | L2-09, L3-04 | 本地 Redis（db 9） |
| L3-10 | `10_cron_task.go` | `gcron` 秒级定时、命名任务、延迟一次性任务 | L2-10 | 无 |
| L3-11 | `11_async_queue.go` | 异步任务：重试退避、幂等去重、优雅停止 | L2-12 | 无 |
| L3-12 | `12_jwt_auth.go` | 手写 HS256 JWT + 鉴权中间件 | L3-03 | 无 |
| L3-13 | `13_telegram_notify.go` | TG 通知：无效 token 的真实失败路径 + 降级 | L3-02, L3-05 | `TG_BOT_TOKEN`+`TG_CHAT_ID`（可选） |
| L3-14 | `14_web3_onchain_read.go` | 链上只读：JSON-RPC + calldata 拼装 | L3-02, L3-03 | `ETH_RPC_URL`（不设则用公共节点） |
| L3-15 | `15_project_layout.go` | `controller / logic / model` 分层与依赖方向 | L3-02, L3-03 | 无 |
| L3-16 | `16_deploy_build.go` | 交叉编译、静态二进制、部署产物 | L3-15 | `DEPLOY_DEMO=1` 才真实编译 |

> 编号顺序即推荐学习顺序。
> L3-07/08 需要环境变量 `BOT_DB_LINK`、L3-09 需要本地 Redis（配置用 db 9 隔离）、
> L3-13 需要 `TG_BOT_TOKEN` + `TG_CHAT_ID`、L3-14 可选 `ETH_RPC_URL`（不设则用公共节点）；
> **关卡一律不内置账号密码**，缺依赖时明确跳过并给出设置方法，不会 panic、不会卡住。

## 目录结构

```
04-goframe/
├── README.md
├── go.mod / go.sum
├── 01_hello_route.go … 16_deploy_build.go   # ← 一个文件一关，每关自带 func main()
├── demoserver/demoserver.go                 # 5 个 HTTP 关卡共用的「起临时服务」辅助包
├── manifest/config/config.yaml              # GoFrame 自动加载的配置文件
├── internal/                                # L3-15 的三层骨架（正常可编译包，勿改）
│   ├── model/                               # 数据结构 + 校验规则
│   ├── logic/                               # 业务规则（不出现 HTTP、不出现 SQL）
│   └── controller/                          # 协议翻译：标准出入参 handler
└── runtime/                                 # 日志与编译产物（已 gitignore）
```

## 每个文件首行的 `//go:build level`，为什么必须留着

Go 规定同一个目录下 `package main` 只能有一个 `func main()`。16 个文件都自带 `main()`，
不加标签就会编译失败：

```
./02_params_binding.go:53:6: main redeclared in this block
        ./01_hello_route.go:57:6: other declaration of main
```

加了 `//go:build level` 之后：

- `go run 01_hello_route.go` —— **显式指定文件名时会绕过构建约束**，照样能单独跑；
- `go build ./...` / `go vet ./...` —— 自动跳过这些关卡文件，冲突消失，
  但仍会正常编译 `demoserver/` 与 `internal/` 这两个真实包。

> **不要换成 `//go:build ignore`。** 实测 `ignore` 会让 `go mod tidy` 认为模块里
> 没有任何代码，从而**清空 go.mod 的 require 列表**（04-goframe 会因此丢掉
> goframe / mysql / redis 依赖）。自定义标签 `level` 没有这个问题。

## `demoserver/` 包是干什么的

L3-01 / L3-02 / L3-03 / L3-12 / L3-15 五个关卡都要**起一个真实 HTTP 服务**来演示，
而「一个文件 = 一关 = 完整 `main()`」的约定意味着每个关卡文件都是各自独立的
`package main`，无法共享同目录下的私有函数。与其把这 35 行复制五遍，
不如抽成一个真正可导入的包 `go-study/goframe/demoserver`：

- `demoserver.FreePort()`：向内核要一个空闲端口，避免关卡之间、以及与本机其它服务冲突；
- `demoserver.WithServer(setup, fn)`：起一个随机端口的临时服务 → 执行 `fn` 发请求 → 跑完优雅关闭。

关卡文件里写 `demoserver.WithServer(...)` 即可。

## 两个附带的注意点

**1. `go run .` 不再可用。** 关卡文件全被标签排除后，模块根目录在常规构建下
「匹配不到任何包」，`go run .` 会报 `build constraints exclude all Go files`。
这不是坏了，而是本模块的设计：**跑单个文件，而不是跑整个目录**。

```sh
go run 01_hello_route.go   # ✅ 正确用法
go run .                   # ❌ 本模块没有「整个目录」这个入口
```

**2. 根目录的 `make build` / `make vet` 已经能正确识别这种情况**（见 `Makefile` 的 `NOPKG` 变量）。
真正的编译/静态检查失败依然会 `FAIL` 并中断。

## 实测要点（跑起来真的能看到）

| 关卡 | 真实观察到的框架行为 |
| :--- | :--- |
| L3-02 | `{"side":"HOLD"}` → 400 + `方向只能是 BUY 或 SELL`；未写自定义文案的规则退回英文默认值 |
| L3-03 | 业务错误仍返回 **HTTP 200**，信息在 body 的 `code/message`（`51` = Validation Failed） |
| L3-04 | 读不存在的 key **不报错**、返回空值 —— 配置缺失会静默兜底 |
| L3-05 | `ERRO` 级别自动附堆栈，同一链路共享 trace id；`SetLevel(LEVEL_ERRO)` 后 `Info` 消失 |
| L3-06 | `WrapCode` 后同时拿到对外提示、内部原因、堆栈、错误码 |
| L3-07 | 建表 + 插入 3 行，`filled=2`、SUM 聚合得 `13055.89`；全程没有手写 `sql.Open/Close` |
| L3-08 | 链式更新生效（canceled=1），但事务内故意返回 error 后 `filled` 前后都=1（回滚真实生效） |
| L3-09 | TTL 1s 后读到空串；INCR 连发 5 次计数=5；`SetNX` 首次 true、后两次 false（幂等生效） |
| L3-10 | 每秒任务跑到第 2 次被 `Remove`，再等 1.1s 计数不变；`AddOnce` 恰好 1 次 |
| L3-11 | 4 次投递 2 个 ID → 只执行 2 次；`SUB-2` 按 **100ms、200ms** 退避后第 3 次成功 |
| L3-12 | 无 token 401、篡改签名 401、过期 401、合法 token 200 并回显 `hello trader01` |
| L3-13 | 无效 token 被正确判为失败：`Telegram 返回失败 http=401 ok=false desc=Unauthorized` |
| L3-14 | 真实链上数据：区块号、以太坊基金会余额、transfer calldata 完整拼出 |
| L3-15 | 同一 `/api/order`：正常 `code:0 ORD-0001`；超限额与下架单各 `code:50`；参数非法 `code:51` |
| L3-16 | `CGO_ENABLED=0 GOOS=linux GOARCH=amd64` 产出静态 ELF（`file` 校验通过） |

## 这轮踩到并写进关卡的两个坑

1. **文件名不能叫 `16_deploy_linux.go`**：Go 把 `*_GOOS.go` / `*_GOARCH.go` 后缀当作**隐式构建约束**，
   在 macOS 上整个文件不参与编译。已改名 `16_deploy_build.go`，并写进 L3-16 思考题。
2. **`time.Duration` 用 `%d` 打印出来是纳秒数**：重试延迟一度显示成 `100000000ms`，实际是 100ms；
   要用 `%v`。见 L3-11。

## 交叉编译验证

```sh
make build-linux                        # 根目录执行，产物 04-goframe/runtime/dist/bot-linux-amd64
DEPLOY_DEMO=1 go run 16_deploy_build.go # 关卡内自己编译并校验 ELF 文件头（默认跳过以保速度）
```

## 通关标准（本阶段）

能独立搭出一个 GoFrame 服务：配置驱动 + 结构化日志 + 统一错误码 + `g.DB` 落订单 +
`g.Redis` 去重限频 + `gcron` 轮询 + JWT 鉴权接口，交叉编译成静态二进制后在 VPS 上以 systemd 常驻。

## 下一站

L3-01 ~ L3-16 全部通关后进入 **L4 交易机器人**（`05-trading-bot`）：
把本框架的路由 / 配置 / 日志 / 错误码 / `g.DB` / `g.Redis` / `gcron` / 鉴权 / 通知 / 链上交互，
串成「行情 → 信号 → 风控 → 下单 → 通知 → 上线」一条真实链路。
