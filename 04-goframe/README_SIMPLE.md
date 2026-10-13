# 04-goframe：直观学习模式

> **每个文件独立运行，学习 GoFrame Web 框架！**

## 🎯 新的学习方式

### ✅ 现在（直观模式）
```bash
cd 04-goframe
go run 01_hello_route.go    # 直接运行，自动起服务
```

每个文件：
- 完整的可运行示例
- 自动端口管理（随机端口）+ 优雅关闭
- 控制台分区：🎯 标题、📝📌🔍 示例、💡 总结

## 📚 学习顺序（16 关）

### 核心框架（L3-01 ~ L3-06）
| 编号 | 文件 | 核心概念 | 依赖 | 运行命令 |
|------|------|----------|------|----------|
| **L3-01** | `01_hello_route.go` | Server 与路由 | 无 | `go run 01_hello_route.go` |
| **L3-02** | `02_params_binding.go` | 参数绑定与校验 | 无 | `go run 02_params_binding.go` |
| **L3-03** | `03_middleware.go` | 中间件 | 无 | `go run 03_middleware.go` |
| **L3-04** | `04_config.go` | 配置管理 | 无 | `go run 04_config.go` |
| **L3-05** | `05_logging.go` | 日志系统 | 无 | `go run 05_logging.go` |
| **L3-06** | `06_error_handling.go` | 错误处理 | 无 | `go run 06_error_handling.go` |

### 数据库与缓存（L3-07 ~ L3-09）
| 编号 | 文件 | 核心概念 | 依赖 | 运行命令 |
|------|------|----------|------|----------|
| **L3-07** | `07_db_mysql.go` | MySQL 基础 | MySQL | 需设置 `BOT_DB_LINK` |
| **L3-08** | `08_orm_model.go` | ORM 与事务 | MySQL | 需设置 `BOT_DB_LINK` |
| **L3-09** | `09_cache_redis.go` | Redis 缓存 | Redis | 本地 Redis |

### 任务与鉴权（L3-10 ~ L3-12）
| 编号 | 文件 | 核心概念 | 依赖 | 运行命令 |
|------|------|----------|------|----------|
| **L3-10** | `10_cron_task.go` | 定时任务 | 无 | `go run 10_cron_task.go` |
| **L3-11** | `11_async_queue.go` | 异步队列 | 无 | `go run 11_async_queue.go` |
| **L3-12** | `12_jwt_auth.go` | JWT 鉴权 | 无 | `go run 12_jwt_auth.go` |

### 通知与集成（L3-13 ~ L3-14）
| 编号 | 文件 | 核心概念 | 依赖 | 运行命令 |
|------|------|----------|------|----------|
| **L3-13** | `13_telegram_notify.go` | Telegram 通知 | TG Token | 需设置 `TG_BOT_TOKEN` |
| **L3-14** | `14_web3_onchain_read.go` | 链上只读 | 可选 | `go run 14_web3_onchain_read.go` |

### 工程实践（L3-15 ~ L3-16）
| 编号 | 文件 | 核心概念 | 依赖 | 运行命令 |
|------|------|----------|------|----------|
| **L3-15** | `15_project_layout.go` | 项目分层 | 无 | `go run 15_project_layout.go` |
| **L3-16** | `16_deploy_build.go` | 交叉编译 | 无 | `go run 16_deploy_build.go` |

## 🚀 快速开始

```bash
# 1. 进入目录
cd 04-goframe

# 2. 先学不需要外部依赖的关卡
go run 01_hello_route.go      # HTTP 服务
go run 02_params_binding.go   # 参数校验
go run 03_middleware.go       # 中间件
go run 04_config.go           # 配置
go run 05_logging.go          # 日志
go run 06_error_handling.go   # 错误处理

# 3. 需要外部依赖的关卡
# MySQL: L3-07, L3-08
export BOT_DB_LINK='mysql:root:password@tcp(127.0.0.1:3308)/go_study'
go run 07_db_mysql.go

# Redis: L3-09
redis-server &
go run 09_cache_redis.go

# Telegram: L3-13
export TG_BOT_TOKEN='your_token'
export TG_CHAT_ID='your_chat_id'
go run 13_telegram_notify.go
```

## 💡 学习建议

### 每关的学习步骤
1. **先运行**：观察服务启动和输出
2. **看代码**：理解 GoFrame 的 API
3. **测试接口**：
   ```bash
   # 另开终端测试
   curl http://localhost:端口/path
   ```
4. **修改实验**：改路由、改参数、改逻辑

### 推荐顺序
**第一阶段**（基础，无依赖）：
- L3-01 ~ L3-06, L3-10 ~ L3-12, L3-15 ~ L3-16

**第二阶段**（需要数据库/Redis）：
- L3-07 ~ L3-09

**第三阶段**（需要外部 Token）：
- L3-13, L3-14

## 🎓 学完这 16 关你会掌握

### Web 开发
1. **HTTP 服务**：路由、参数、中间件
2. **数据校验**：自动绑定、中文错误提示
3. **统一响应**：标准出入参格式

### 数据层
4. **数据库**：g.DB() CRUD、链式查询
5. **ORM**：Model、事务、回滚
6. **缓存**：Redis CRUD、TTL、SetNX

### 基础设施
7. **配置**：g.Cfg() 读取 yaml
8. **日志**：glog 分级、Trace ID
9. **错误**：gerror 错误码、堆栈

### 任务调度
10. **定时任务**：gcron 秒级定时
11. **异步队列**：重试、退避、幂等

### 安全与通知
12. **JWT 鉴权**：签发、校验、中间件
13. **Telegram**：消息推送、错误降级
14. **Web3**：链上只读、JSON-RPC

### 工程实践
15. **分层架构**：controller/logic/model
16. **交叉编译**：Linux 静态二进制

## ⚠️ 常见问题

### Q: 为什么有些文件运行后一直在等待？
A: GoFrame 服务默认会阻塞运行。文件中已经设置了自动关闭，按 Ctrl+C 或等待几秒即可退出。

### Q: 端口被占用怎么办？
A: 学习文件使用随机端口，不会冲突。如果遇到问题，检查是否有其他 GoFrame 服务在运行。

### Q: L3-07/08 报错找不到数据库？
A: 需要先设置环境变量：
```bash
export BOT_DB_LINK='mysql:root:密码@tcp(127.0.0.1:3308)/go_study'
```

### Q: L3-09 Redis 连接失败？
A: 确保 Redis 在本地 6379 端口运行：
```bash
redis-server
```

### Q: 可以跳过需要外部依赖的关卡吗？
A: 可以！先学 L3-01~06, L3-10~12, L3-14~16（无依赖），之后再配环境学其他的。

## 🔧 环境变量

| 变量名 | 用途 | 示例 | 需要的关卡 |
|--------|------|------|-----------|
| `BOT_DB_LINK` | MySQL 连接 | `mysql:root:pass@tcp(127.0.0.1:3308)/go_study` | L3-07, L3-08 |
| `BOT_REDIS_ADDR` | Redis 地址 | `127.0.0.1:6379` | L3-09 |
| `TG_BOT_TOKEN` | Telegram Token | `123456:ABC...` | L3-13 |
| `TG_CHAT_ID` | Telegram Chat ID | `-1001234567890` | L3-13 |
| `ETH_RPC_URL` | 以太坊 RPC | `https://eth.llamarpc.com` | L3-14（可选）|

## 🎯 通关标准（整个模块）

能独立搭出一个 GoFrame 服务：
- ✓ 配置驱动（g.Cfg）
- ✓ 结构化日志（glog）
- ✓ 统一错误码（gerror）
- ✓ 数据库操作（g.DB）
- ✓ Redis 缓存（g.Redis）
- ✓ 定时任务（gcron）
- ✓ JWT 鉴权
- ✓ 交叉编译成 Linux 静态二进制

## 📊 对应交易机器人能力

| GoFrame 能力 | 对应关卡 | 机器人场景 |
|-------------|---------|-----------|
| 路由与接口 | L3-01~03 | 控制后台、健康检查 |
| 配置管理 | L3-04 | 不同环境不同参数 |
| 日志系统 | L3-05~06 | 成交日志、异常追踪 |
| 数据库 | L3-07~08 | 订单/成交/持仓落库 |
| Redis | L3-09 | 行情缓存、限频、幂等 |
| 定时任务 | L3-10 | 定时轮询、日报 |
| 异步队列 | L3-11 | 下单重试、通知重试 |
| JWT | L3-12 | 后台鉴权 |
| 通知 | L3-13 | Telegram 告警 |
| Web3 | L3-14 | 链上数据读取 |
| 分层 | L3-15 | 代码组织 |
| 部署 | L3-16 | VPS 部署 |

## 🚀 下一站

完成本模块后，前往：
- **L4 交易机器人**：`../05-trading-bot/`
- 把 L3 的所有能力串成「行情 → 信号 → 风控 → 下单 → 通知」真实链路

---

💡 **提示**：
- 不需要外部依赖的关卡可以随时运行
- 配置好环境变量后再学数据库和缓存关卡
- 每个关卡都有详细注释，边看代码边运行效果最佳
