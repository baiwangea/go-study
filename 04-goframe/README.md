# L3 · GoFrame 工程主线（关卡制）

语法与并发打完（L1、L2），**从这里开始只学一个框架：GoFrame**。
原先散落成独立目录的 mysql / redis / logger / jwt / telegram / web3 / 任务队列，
全部变成本框架里的关卡 —— 因为它们本来就要在一个工程里协同工作。

## 怎么用

```sh
cd 04-goframe

go run . list    # 看已实现的关卡目录
go run . 1       # 只跑第 1 关
go run . 1-6     # 跑一组
go run .         # 全跑
```

前 6 关**不依赖任何外部服务**（随机端口起服务、跑完自动关闭），有 MySQL/Redis/RPC 之后需要本地服务。

## 关卡规划（16 关）

| 编号 | 文件 | 主题 | 吸收自（原独立目录） | 状态 |
| :--- | :--- | :--- | :--- | :--- |
| L3-01 | `01_hello_route.go` | Server 实例、路由注册、优雅关闭 | `stdlib-http` 服务端关 | ✅ 已实现 |
| L3-02 | `02_params_binding.go` | `r.Parse` 绑定 + `v` 校验 + 中文文案 | `gin` 参数处理 | ✅ 已实现 |
| L3-03 | `03_middleware.go` | 标准出入参、`g.Meta`、统一响应中间件 | `stdlib-http` 中间件关 | ✅ 已实现 |
| L3-04 | `04_config.go` | `g.Cfg` 与 `manifest/config/config.yaml` | 各框架的配置目录 | ✅ 已实现 |
| L3-05 | `05_logging.go` | `glog` 分级、字段、trace id、按天落盘 | `stdlib-logger` | ✅ 已实现 |
| L3-06 | `06_error_handling.go` | `gerror` 错误码、Wrap、堆栈 | 手写 error 处理 | ✅ 已实现 |
| L3-07 | `07_db_mysql.go` | `g.DB()` CRUD、参数化、分页 | `go-mysql-example` | ⬜ 需 MySQL |
| L3-08 | `08_orm_model.go` | ORM 链式、事务、模型与 DAO 生成 | `gin` 里的 GORM 层 | ⬜ 需 MySQL |
| L3-09 | `09_cache_redis.go` | `g.Redis()`、`gcache` 本地缓存与降级 | `go-redis-example` | ⬜ 需 Redis |
| L3-10 | `10_cron_task.go` | `gcron` 定时轮询行情 / 生成日报 | `Asynq` 的 cron 部分 | ⬜ |
| L3-11 | `11_async_queue.go` | 异步任务与重试、幂等（框架内落地） | `Asynq` | ⬜ |
| L3-12 | `12_jwt_auth.go` | `ggjwt` 签发校验 + 鉴权中间件 | `go-jwt-example` | ⬜ |
| L3-13 | `13_telegram_notify.go` | TG 通知成交与异常 | `go-telegram-bot-example` | ⬜ 需 token |
| L3-14 | `14_web3_onchain_read.go` | `ethclient` 链上只读（区块/余额/事件） | `go-web3-example` | ⬜ 需 RPC |
| L3-15 | `15_project_layout.go` | 工程分层、`gf gen` 代码生成、依赖注入 | `gin`/`go-zero` 脚手架 | ⬜ |
| L3-16 | `16_deploy_linux.go` | `gf pack`、交叉编译、systemd、健康检查 | 原散落在 README | ⬜ |

> ⬜ 的关卡需要先起对应服务（MySQL / Redis / ETH RPC）。下一轮按你的节奏逐关补齐。

## 目录结构

```
04-goframe/
├── main.go                      # 关卡导航
├── demo_server.go               # 起临时服务的公共辅助（随机端口 + 优雅关闭）
├── level/level.go               # 关卡运行时（与其它模块同一份，零依赖）
├── manifest/config/config.yaml  # GoFrame 自动加载的配置文件
├── runtime/log/                 # glog 落盘目录（L3-05 生成）
└── NN_主题.go                   # 一关一文件
```

## 已实现 6 关的实测要点（跑起来能看到的）

- L3-02：`{"side":"HOLD"}` → 400 + `方向只能是 BUY 或 SELL`（`v` tag 里 `#` 后的自定义文案生效；未提供的规则会退回英文默认文案）
- L3-03：业务错误也返回 **HTTP 200**，错误信息在 body 的 `code/message` 里（`code:51` = Validation Failed）—— 这是统一响应中间件的取舍，思考题让你判断它合不合理
- L3-04：读不存在的 key **不报错**，返回空值 —— 配置缺失会静默兜底，必须自己判空
- L3-05：`ERRO` 级别自动附调用堆栈，同一次调用链共享 trace id；`SetLevel(glog.LEVEL_ERRO)` 后 `Info` 直接消失
- L3-06：`gerror.WrapCode` 之后能同时拿到对外提示、内部原因、堆栈与错误码

## 通关标准（本阶段）

能独立搭出一个 GoFrame 服务：配置驱动 + 结构化日志 + 统一错误码 + MySQL/Redis 落数据 +
定时任务拉行情 + JWT 鉴权的控制接口，并用 `gf` 交叉编译成 Linux 二进制部署到 VPS。
