# L3 · GoFrame 工程主线（关卡制）

语法与并发打完（L1、L2），**从这里开始只学一个框架：GoFrame**。
原先散落成独立目录的 mysql / redis / 日志 / jwt / telegram / web3 / 任务队列，
全部变成本框架里的关卡 —— 因为它们本来就要在一个工程里协同工作。

## 怎么用

```sh
cd 04-goframe

go run . list    # 看关卡目录
go run . 1       # 只跑第 1 关
go run . 7-10    # 跑一组（按 list 里的序号，不是关卡号）
go run .         # 全跑
```

L3-01~06、L3-10~12、L3-15~16 **不依赖任何外部服务**（随机端口起服务、跑完自动关闭）；
仅 L3-07/08/09 需要本地 MySQL / Redis。

## 关卡表（16 关，已实现 14）

| 编号 | 文件 | 主题 | 这关替掉原先哪块知识 | 状态 |
| :--- | :--- | :--- | :--- | :--- |
| L3-01 | `01_hello_route.go` | Server 实例、路由注册、优雅关闭 | 标准库 net/http 服务端要点 | ✅ 已实现 |
| L3-02 | `02_params_binding.go` | `r.Parse` 绑定 + `v` 校验 + 中文文案 | 手写参数解析与校验 | ✅ 已实现 |
| L3-03 | `03_middleware.go` | 标准出入参、`g.Meta`、统一响应中间件 | net/http 中间件要点 | ✅ 已实现 |
| L3-04 | `04_config.go` | `g.Cfg` 与 `manifest/config/config.yaml` | 各框架的配置文件套路 | ✅ 已实现 |
| L3-05 | `05_logging.go` | `glog` 分级、字段、trace id、按天落盘 | 标准库 log 与日志组织 | ✅ 已实现 |
| L3-06 | `06_error_handling.go` | `gerror` 错误码、Wrap、堆栈 | 手写 error 分类 | ✅ 已实现 |
| L3-10 | `10_cron_task.go` | `gcron` 秒级定时、命名任务、延迟一次性任务 | 任务队列的定时部分 | ✅ 已实现 |
| L3-11 | `11_async_queue.go` | 异步任务：重试退避、幂等去重、优雅停止 | Redis 任务队列（优先级重试幂等） | ✅ 已实现 |
| L3-15 | `15_project_layout.go` | `controller / logic / model` 分层与依赖方向 | 框架工程分层经验 | ✅ 已实现 |
| L3-16 | `16_deploy_build.go` | 交叉编译、静态二进制、部署产物 | 散落在各框架 README 的部署说明 | ✅ 已实现 |
| L3-07 | `07_db_mysql.go` | `g.DB()` 建表、插入、条件查询、聚合 | `database/sql` 裸写 MySQL | ✅ 已实现 |
| L3-08 | `08_orm_model.go` | ORM 链式、事务与回滚 | 框架内 ORM 与分层 | ✅ 已实现 |
| L3-09 | `09_cache_redis.go` | `g.Redis()` 缓存、INCR 限频、SetNX 幂等 | `go-redis` 客户端 | ✅ 已实现 |
| L3-12 | `12_jwt_auth.go` | 手写 HS256 JWT + 鉴权中间件 | 手写 JWT 签发校验 | ✅ 已实现 |
| L3-13 | `13_telegram_notify.go` | TG 推送成交与异常 | TG Bot 通知与报表 | ⬜ 需 token |
| L3-14 | `14_web3_onchain_read.go` | `ethclient` 链上只读（区块/余额/事件） | ethclient 链上只读 | ⬜ 需 RPC |

> 编号顺序即推荐学习顺序（表中把已实现的排在前面便于查看）。
> L3-07/08 需要环境变量 `BOT_DB_LINK`（关卡不内置账号密码），L3-09 需要本地 Redis（配置用 db 9 隔离）；
> 未配置时这三关会**明确跳过并给出设置方法**，不会报错。

## 目录结构

```
04-goframe/
├── main.go                      # 关卡导航：list / 单关 / 区间 / 全跑
├── demo_server.go               # 起临时服务的公共辅助（随机端口 + 优雅关闭）
├── level/level.go               # 关卡运行时（与其它模块同一份，零依赖）
├── manifest/config/config.yaml  # GoFrame 自动加载的配置文件
├── internal/                    # L3-15 的三层骨架
│   ├── model/                   # 数据结构 + 校验规则
│   ├── logic/                   # 业务规则（不出现 HTTP、不出现 SQL）
│   └── controller/              # 协议翻译：标准出入参 handler
├── runtime/                     # 日志与编译产物（已 gitignore）
└── NN_主题.go                   # 一关一文件
```

## 实测要点（都是跑出来的）

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
| L3-15 | 同一 `/api/order`：正常 `code:0 ORD-0001`；超限额与下架单各 `code:50`；参数非法 `code:51` |
| L3-16 | `CGO_ENABLED=0 GOOS=linux GOARCH=amd64` 产出 **19M 静态 ELF**（`file` 校验通过） |

## 这轮踩到并写进关卡的两个坑

1. **文件名不能叫 `16_deploy_linux.go`**：Go 把 `*_GOOS.go` / `*_GOARCH.go` 后缀当作**隐式构建约束**，
   在 macOS 上整个文件不参与编译，报错 `undefined: L16`。已改名 `16_deploy_build.go`，并写进 L3-16 思考题。
2. **`time.Duration` 用 `%d` 打印出来是纳秒数**：重试延迟一度显示成 `100000000ms`，实际是 100ms；
   要用 `%v`。见 L3-11。

## 交叉编译验证

```sh
make build-linux                      # 根目录执行，产物 04-goframe/runtime/dist/bot-linux-amd64
DEPLOY_DEMO=1 go run . 10             # 关卡内自己编译并校验 ELF 文件头（默认跳过以保速度）
```

## 通关标准（本阶段）

能独立搭出一个 GoFrame 服务：配置驱动 + 结构化日志 + 统一错误码 + `g.DB` 落订单 +
`g.Redis` 去重限频 + `gcron` 轮询 + JWT 鉴权接口，交叉编译成静态二进制后在 VPS 上以 systemd 常驻。

## 下一站

L3-01 ~ L3-16 全部通关后进入 **L4 交易机器人**（待建 `05-trading-bot`）：
把本框架的路由 / 配置 / 日志 / 错误码 / `g.DB` / `g.Redis` / `gcron` / 鉴权 / 通知 / 链上交互，
串成「行情 → 信号 → 风控 → 下单 → 通知 → 上线」一条真实链路。
