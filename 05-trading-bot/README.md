# L4 · 交易机器人骨架（端到端管线）

把前面所有阶段的能力串成一条**能跑、能测、能停**的链路：

```
market.Feed ──▶ strategy.Strategy ──▶ risk.Guard ──▶ executor.Engine ──▶ notify.Notifier
  行情源            信号                守门           提交/重试/幂等         通知
```

每层只依赖下一层的**接口**，所以：换 CEX/链上行情、加第二个策略、把 mock 交易所换成真交易所、
把控制台通知换成 Telegram，都不需要动其它层 —— 这就是 L1-09 隐式接口 + L2 并发的实际价值。

## 运行

```sh
cd 05-trading-bot
go run .          # 跑一轮完整链路（mock 行情，无需网络与密钥）
go run -race .    # 顺带验证管线没有竞态

# 真实行情（Bitget 现货 tickers，公开接口无需 API Key）
go run . -feed=bitget -ticks=40 -interval=300ms -threshold=0.00000001

# 推送式行情（WebSocket，自带心跳与断线重连）
go run . -feed=bitget-ws -ticks=70 -threshold=0.00000001

# 状态落在进程之外（MySQL 落库 + Redis 跨进程幂等）
export BOT_DB_LINK='mysql:root:<密码>@tcp(127.0.0.1:3308)/go_study?loc=Local&parseTime=true'
export BOT_REDIS_ADDR='127.0.0.1:6379'
go run . -ticks=16

# 回测模式：零网络、零密钥，用确定的价格序列跑同一条管线并出报告
go run . -mode=backtest -pattern=range
go run . -mode=backtest -pattern=v -bars=1000 -fee-bp=10
go run . -mode=backtest -csv=./data/btc.csv      # 列：timestamp,price
```

固定随机种子，mock 输出可复现。mock 一轮实测：

```
收到行情 16 笔｜产生信号 7 个｜下单成功 4 笔｜落库 5 条
风控拦截 2 次：冷却期内重复信号
幂等演示：首次 err=<nil>（accepted/submitted）；同一订单号再提 err=重复提交 ORD-IDEM，已忽略（duplicate/created，状态不动）
状态机演示：submitted --部分成交(0.001)--> partial --全部成交--> filled --重复回调--> filled（幂等）
协程数：启动前 3 → 结束后 3（cancel 生效，行情协程已退出）
go run -race .   → 0 个 DATA RACE
```

这几行输出正好对应八个验收点：**行情不丢、信号可产出、风控真拦截、重试后成交、幂等生效、
状态机幂等、协程不泄漏、无竞态**。

## 接真实行情（Bitget 现货 tickers）

行情源只是 `Feed` 接口的一个实现，换源不动下游：

| 实现 | 文件 | 数据来源 |
| :--- | :--- | :--- |
| `RandomWalk` | `internal/market/market.go` | 本地随机游走（零依赖） |
| `Bitget` | `internal/market/bitget.go` | 现货 tickers **REST 轮询** |
| `BitgetWS` | `internal/market/bitget_ws.go` | 现货 ticker **WebSocket 推送**（心跳 + 断线重连） |

## 状态落地：MySQL 存订单、Redis 管幂等

两个能力都是**接口定义在使用方、实现可替换**，不配环境变量就自动降级到内存版，照样能跑：

| 接口 | 定义在 | 实现 | 开启方式 |
| :--- | :--- | :--- | :--- |
| `executor.Deduper` | 使用方（executor） | 内存 map / Redis SetNX(db 9, TTL 1h) / 库唯一索引兼做兜底 | `BOT_REDIS_ADDR` |
| `store.Store` | `internal/store` | memory / `g.DB()` 写 `go_study.bot_orders` | `BOT_DB_LINK` |

配好后连跑两次（订单号相同）的实测结果：

```
第一次：幂等：Redis SetNX（跨进程有效）｜存储：MySQL(g.DB) 表 bot_orders
        收到行情 16 笔｜下单成功 3 笔｜落库 4 条 → 回读 bot_orders 累计行数 = 4
第二次（新进程）：✗ 下单失败：重复提交 ORD-SELL-9，已忽略（状态 duplicate）
        收到行情 16 笔｜下单成功 2 笔｜落库 4 条 → 回读 bot_orders 累计行数 = 6
```

这就是“重启也不会重复下单”的真实含义：**幂等状态不在进程里**。同时 `bot_orders.order_id` 建了唯一索引
+ `InsertIgnore`，等于**两层防线**（Redis 快、数据库硬）。

## 金额精度：为什么这里全用 decimal

**钱和数量一律 `decimal.Decimal`，绝不 `float64`。** 这不是洁癖，是这一行的基本要求。

改动来自 `07-web3-trading` 的 W-01~W-03：`0.1+0.2` 在 float64 下是 `0.30000000000000004`，
数量按 stepSize 对齐时会算错，交易所直接拒单（`0.3/0.1` 在 float64 里是 `2.9999999999999996`，floor 之后少下一档）。

**界线画得很清楚，不要无脑全局替换：**

| 用 `decimal` | 用 `float64` |
| :--- | :--- |
| `executor.Order.Qty` / `FilledQty` | `strategy.Signal.Strength`（信号强度） |
| `risk.Guard.MaxNotional`、名义金额 | `MACross` 的均线价格、`Threshold`、`Peak()` |
| `store.Record.Qty`、落库数量字段 | `RandomWalk.Start/Vol`（随机游走参数） |
| `BOT_MAX_NOTIONAL` 等金额环境变量 | `executor.Mock.FailRate/SlowRate`（概率） |

判断标准一句话：**「这是钱/数量，还是统计量/概率？」** 前者用 decimal，后者 float64 完全够用。

### `Quote.PriceStr`：从源头避开浮点误差

交易所返回的价格**本来就是字符串**：`{"lastPr":"84800.63"}` —— 人家自己就在避开浮点误差。
所以 `market.Quote` 多了一个字段：

```go
type Quote struct {
    Symbol   string
    Price    float64  // 给策略算均线用（统计量，误差无所谓）
    PriceStr string   // 交易所原始字符串 —— 下单时用它转 decimal
}
```

`Bitget` 填 `t.LastPr`、`BitgetWS` 填推送里的 `lastPr`、`RandomWalk` 用 `strconv.FormatFloat(...,-1,64)`。
**下游下单走 `decimal.NewFromString(q.PriceStr)`，绝不用 `decimal.NewFromFloat(q.Price)`** ——
后者会把 float64 的误差原样带进来，等于白改。

另外 `order.AlignQtyDown(qty, step)` 用 `qty.Sub(qty.Mod(step))` 做步长对齐，
**避开除法**（除法才是精度损失的来源）。

## 订单状态机

`internal/order` 是一个独立的状态机包，状态流转：

```
created ──▶ submitted ──▶ partial ──▶ filled     （终态）
   │            │            │
   │            ├────────────┴──────▶ canceled   （终态）
   └──▶ rejected                                 （终态）
```

三个必须记住的设计点：

1. **非法跃迁一律拒绝**，返回明确错误。`filled → submitted`、`canceled → filled` 都不行 ——
   终态出不去。`Transition(s, s)` 例外：**同状态视为「无变化」而不是错误**，
   因为交易所回调会重复推送，报错反而会让上层误判。
2. **幂等**：重复投递同一个成交回报不会改变状态，也不会累加数量。
   实测输出里的 `--重复回调--> filled（幂等）` 就是在演示这条。
3. **不变量校验**：`0 <= FilledQty <= Qty`，违反时报错而不是悄悄写脏数据。

接入点：`Engine.Submit` 成功后转 `submitted`、重试耗尽转 `rejected`；
新增 `Engine.ApplyFill(ctx, orderID, filledQty)` 按累计成交量推导 `partial` 还是 `filled`。
`store` 落库时**记录状态**（不再只写终态）。

## 回测模式

`-mode=backtest` 用**同一条管线**（strategy → risk → executor → store）跑一段确定的价格序列，
最后出一份报告。零网络、零密钥，`go run . -mode=backtest` 直接能跑。

```sh
go run . -mode=backtest -pattern=range              # 内置四种形态
go run . -mode=backtest -pattern=v -bars=1000       # 可调行情笔数
go run . -mode=backtest -fee-bp=10 -slippage-bp=2   # 可调手续费与滑点
go run . -mode=backtest -csv=./data/btc.csv         # 或读外部 CSV（timestamp,price）
```

四种内置形态的实测报告（500 笔，手续费 5 bp/边）：

| 形态 | 信号/拦截/下单/成交 | 毛利 | 手续费 | **净利** | 胜率 | 最大回撤 |
| :--- | :--- | ---: | ---: | ---: | ---: | ---: |
| 单边上涨 `up` | 325/247/78/78 | 64.18 | 5.61 | **+58.58** | 64.7% | 14.07 |
| 单边下跌 `down` | 377/290/87/87 | 95.82 | 4.64 | **+91.18** | 75.7% | 7.77 |
| 震荡 `range` | 336/262/74/74 | -44.14 | 4.66 | **-48.80** | 27.8% | 48.80 |
| V 型反转 `v` | 366/282/84/84 | -52.68 | 4.48 | **-57.16** | 47.6% | 144.85 |

（单位 USDT）

**这张表本身就是这个策略的体检报告**：趋势里能赚钱，震荡里被反复打脸，
V 型反转在拐点处吃最大回撤（144.85 —— 单笔名义金额才 126，等于一把亏掉一倍多）。
这正是回测的意义：**不用真金白银就能知道策略的死法**。

两个统计口径说明：

- **胜率** = 盈利的平仓笔数 / 总平仓笔数。注意**平仓笔数 ≠ 成交笔数**：
  按「持仓 + 开仓均价」记账，同向成交是加仓（只更新均价），反向成交才产生一次平仓。
- **最大回撤** = 逐笔成交后的**已实现净权益曲线**（累计毛利 − 累计手续费）的峰谷差。
  起点权益视为 0，所以「先亏一笔」本身就构成回撤。

全部金额用 `decimal` 计算，展示保留 2 位小数。

## WebSocket 行情（协议是实测出来的）

文档那一页只给了 REST，所以直接握手探协议：

```text
wss://ws.bitget.com/v2/ws/public
订阅 {"op":"subscribe","args":[{"instType":"SPOT","channel":"ticker","instId":"BTCUSDT"}]}
回执 {"event":"subscribe",...}  →  推送 {"action":"snapshot|update","arg":{...},"data":[{...}]}
```

踩到的两个坑（写进了 `bitget_ws.go` 注释）：`instType` 写 `sp` 被拒（code 30016 Param error），
顶层字段写成 `action` 也被拒（30003 INVALID op:null）—— 必须 `op` + `SPOT` + `instId`。

实现里的三件事（都是行情链路的必选项）：

1. **心跳**：20s 发一次 `ping`（服务端要求 30s 内），否则会被踢；
2. **读超时**：每轮 `SetReadDeadline(90s)`，死连接不会把协程永远卡在读上；
3. **断线重连**：读失败 → 上报 → 500ms 起指数退避（封顶 8s）重新握手重订，`Reconnects()` 可观察次数。

实测（70 个真实推送）：

```text
行情 #52 BTCUSDT 价 84859.02 → ma-cross 发出 SELL 信号（强度 2.3568499e-08）
   ✓ 下单成功：ORD-SELL-52（尝试 2 次）      ← 重试分支被真实触发
   ✗ 下单失败：交易所响应超时：context deadline exceeded（尝试 3 次，状态 failed）
收到行情 70 笔｜幂等：Redis SetNX｜存储：MySQL(g.DB) 表 bot_orders
```

一个诚实的提醒：**别用真实行情 + 默认阈值跑几 tick 就认为“策略坏了”**。BTC 现货在 300ms/推送尺度上
均线偏离只有 **e-08 量级**，默认 `0.0004` 永远不会触发；本模块的 `Peak()` 读数就是用来定阈值的。

实测响应（公开接口，无需 API Key）：

```json
{"code":"00000","msg":"success","data":[{"symbol":"BTCUSDT","lastPr":"84800.63",
 "bidPr":"84800.62","askPr":"84800.63","baseVolume":"1038.764845","ts":"1791081880964"}]}
```

实测得到的三个结论，都是 `FetchTicker` 里写明的处理要点：

1. **价格全是 string**（`"lastPr":"84800.63"`）—— 交易所自己就在避开浮点误差；转 float 只用于算信号，**下单金额继续用字符串/定点**（对照 L3-02）。
2. **两层错误分开**：`Do()` 失败（连不上/超时，可重试）≠ `code != "00000"` 或 HTTP 非 200（业务错，盲重试没意义）。实跑 40 轮出现过 1~2 次单次超时：**只丢那个 tick，管线不中断**。
3. **先测波动再定阈值**：`MACross` 会报告实测最大均线偏离。实测 BTC 现货 300ms 轮询下峰值只有 **2.36e-08**，随手拍的 `0.0004`、`1e-07` 都比它大 → 变成「程序跑得欢，信号永远为 0」。

用 `-threshold=0.00000001` 的真实数据一轮：

```
  收到行情 40 笔｜行情拉取失败 0 次｜产生信号 6 个｜下单成功 5 笔｜提交过的唯一订单号 6 个
  协程数：启动前 1 → 结束后 1（cancel 生效，行情协程已退出）
  校准参考：ma-cross 实测最大均线偏离 = 2.3586240264908043e-08（当前阈值 1e-08）
```

真实行情还自然触发了两条以前只能靠 mock 逼出的分支：`✗ 下单失败：交易所响应超时（尝试 3 次，状态 failed）`
与重试后 `✓ 尝试 2 次` 成功。另两个细节：`CloseIdleConnections()` 不调，HTTP keep-alive 会让协程数下不去；
轮询间隔建议 ≥300ms，否则会被限频（`code` 直接报错）。

## 目录结构

```
05-trading-bot/
├── main.go                    # 管线装配 + 汇总统计 + 数据源/存储/回测开关
└── internal/
    ├── market/                # Feed 接口 + RandomWalk + Bitget(REST) + BitgetWS(推送)
    ├── strategy/              # Strategy 接口 + MACross（带 Peak() 阈值校准读数）
    ├── risk/                  # Guard：名义金额上限、信号偏离、同方向冷却（带锁）
    ├── order/                 # 订单状态机：跃迁表 / 幂等 / 不变量校验 / 步长对齐
    ├── executor/              # Engine：限时+退避重试；Deduper 接口（幂等）
    ├── store/                 # Store 接口 + memory / MySQL(g.DB)（按订单号 upsert）
    ├── backtest/              # 回测：价格序列生成 / CSV / 立即成交交易所 / PnL 统计
    └── notify/                # Notifier 接口 + Console / Telegram（真发送 + 降级）
```

全部放在 `internal/` 下：外部模块无法 import，改签名不用考虑兼容（对照 04-goframe 的分层）。

**测试**：每个包都有 `_test.go`，表驱动 + 子测试 + 接口打桩（`executor.Exchange`、`store.Store` 都有假实现）。
`risk` / `strategy` / `notify` / `order` 覆盖率 100%，`executor` 97.1%，`backtest` 98.8%。
`market` 与 `store` 偏低，因为 `bitget_ws.go` 的连接/重连、`mysqlStore`、`RedisDeduper` 都需要真实网络或服务，
那部分交给 `go run -race .` 的端到端验证。

## 已兑现的前面关卡能力

| 能力 | 来源 | 在本模块的位置 |
| :--- | :--- | :--- |
| 接口抽象与多实现 | L1-08 ~ L1-10 | `Feed` / `Strategy` / `Exchange` / `Notifier` 四个接口 |
| 指针接收者才满足接口 | L1-13 | `Mock.Submit` 是指针接收者，所以用 `executor.NewMock(...)` 返回 `*Mock` 才能当 `Exchange` |
| 并发安全不靠运气 | L2-13 | 同一个 `*rand.Rand` 被行情协程与下单逻辑共用 → `-race` 抓出 8 处 DATA RACE；现在每组件自持种子 |
| 行为接口断言 | L1-10 | `feed.(interface{ CloseIdleConnections() })`、`s.(interface{ Peak() float64 })`：能力探测而非类型判断 |
| HTTP 客户端与错误分类 | L3-02 ~ L3-04 | `bitget.go` 的单请求超时、限长读取、HTTP 状态与业务 code 分开处理 |
| channel 只由生产者 close | L2-05 | `market.Subscribe` 里 `defer close(out)` |
| select + 超时 | L2-08 | 单次下单 `context.WithTimeout(120ms)` |
| 非阻塞丢弃 | L2-09 | 行情协程 `select` 的 `ctx.Done()` 分支 |
| context 取消传播 | L2-11 | `cancel()` 后行情协程退出，协程数回到 1 |
| worker pool / 重试幂等 | L2-12, L3-11 | `Engine.Submit` 的退避重试与 `seen` 表 |
| 配置与环境变量 | L3-04 | `BOT_MAX_NOTIONAL`、`TG_BOT_TOKEN` 走环境变量，密码不进 Git |
| 金额精度（decimal） | W-01 ~ W-03 | 全模块金额/数量用 `decimal`；`Quote.PriceStr` 保留交易所原始字符串 |
| 订单状态机 | W-09 | `internal/order`：跃迁表、幂等、不变量 |
| 表驱动测试与打桩 | T-01 ~ T-07 | 每个包都有 `_test.go`；`executor.Exchange` / `store.Store` 都有假实现 |
| 真实 TG 通知 | L3-13 | `notify.Telegram`：真发送 + 401 失败路径 + 超时 + 降级 |

## 本模块跑出来的两个坑（已写成注释）

1. **`*rand.Rand` 不是并发安全的**。最初为了"可复现"把同一个 `rnd` 同时传给行情源和 mock 交易所，
   `go run -race .` 立刻报 8 处 DATA RACE（栈顶是 `math/rand.(*rngSource).Uint64`）。
   现在改成 `NewMock(seed)` 内部各自 `rand.New(rand.NewSource(seed))`，依旧可复现，且 0 告警。
2. **阈值与波动率不匹配会让管线"看起来能跑但其实空转"**。最初 `MACross` 阈值 2%、行情每 tick 只动 0.2%，
   信号数永远为 0，风控/重试/幂等全都没被触发。现在 `Vol=0.006` + 阈值 `0.0004`，7 个信号、两次拦截、重试后成交全部出现。

教训：**每写完一关就跑一次，并去看统计行是不是真的非零**，不然等于没演示。

## 迭代记录：从骨架到能用

最初这一节叫「刻意留下的粗糙处」，列了 5 条待办。现在逐条兑现：

| # | 原始待办 | 状态 |
| :--- | :--- | :--- |
| 1 | 幂等表在内存里 | ✅ `Deduper` 接口 + Redis SetNX（`BOT_REDIS_ADDR` 一开就生效） |
| 2 | 没有落库 | ✅ `store.Store` + `g.DB()`；**并且补上了订单状态机**（不再是只写终态） |
| 3 | 通知没真发 | ✅ `notify.Telegram` 真实现：真 POST + 401 失败路径 + 4s 超时 + 缺 token 降级 |
| 4 | 行情只有 REST 轮询 | ✅ `BitgetWS` 推送式（心跳 + 读超时 + 重连退避） |
| 5 | 没有测试 | ✅ 每个包都有 `_test.go`，`risk`/`strategy`/`notify`/`order` 覆盖率 100% |

后续又补了三件这一行绕不开的事：

6. **金额精度**：全模块金额/数量从 `float64` 换成 `decimal`，并新增 `Quote.PriceStr` 保留交易所原始字符串；
7. **订单状态机**：`internal/order` + `Engine.ApplyFill`，支持 `created→submitted→partial→filled`；
8. **回测模式**：`-mode=backtest`，零网络零密钥就能评估策略。

**还没做**（下一轮抓手）：

- **Prometheus 指标**：`07-web3-trading` 的 W-11 已经教了，但还没接进这条管线；
  交易机器人最终要靠 P99 延迟、拒单率、滑点这三个指标来发现「策略悄悄坏掉」；
- **OpenTelemetry trace**：目前只有 `04-goframe` L3-05 的 trace id，还没接真正的链路追踪；
- **测试网下单**：现在 `executor.Exchange` 只有 mock 和回测两个实现，
  接一个真实的测试网（Bitget demo / Sepolia）是最后一公里；
- **持仓与权益曲线落库**：回测里的 PnL 是内存计算，实盘需要持久化。

教训（三轮重构下来最值钱的一条）：**先写测试，再动刀。**
第一轮只加 `_test.go`、不碰生产代码；第二轮改 decimal + 状态机时，
每改一处就跑一次测试 —— 签名变更导致的编译错误和真正的行为回归被清楚地区分开了。
反过来做（先重构再补测试）的话，你没法判断「测试失败」是因为改错了还是因为本来就写错了。

## 通关标准

能自己改动任一层的接口实现而其它层不动；能在 `go run -race .` 下无警告；
能说清「为什么冷却期拦截、重试、幂等三件事缺一不可」；
能说清「为什么金额必须用 decimal 而信号强度可以用 float64」；
能读懂回测报告里胜率与最大回撤的统计口径。
