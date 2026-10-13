# W · Web3 与交易专项（横切模块）

前面六个模块教你「会写 Go」「会测 Go」，这一个教你「**用 Go 做交易**」。

它是 `05-trading-bot`（终点项目）的能力前置：那边是**一个完整项目**，
这边是**一个个能单独跑通的最小零件** —— 金额精度、密钥、ABI、签名、状态机、
行情连接、指标。学完这一模块再看 `05-trading-bot`，每一层你都知道它在干什么。

## 怎么用

**一个文件 = 一关 = 一个完整的 `main()`**，和其他模块一样：

```sh
cd 07-web3-trading

go run 01_float64_trap.go        # W-01 从「钱不能用 float64」开始
go run 08_sign_tx.go             # W-08 离线签名，不连网也能跑
go run 12_rpc_readonly.go        # W-12 读链上，连不上会自动降级
```

在仓库根目录也可以：

```sh
make levels M=./07-web3-trading            # 列出本模块所有关卡
make level  F=08_sign_tx                   # 跑指定关卡（M 可省略、可省略 .go）
```

每关文件的头部注释就是「关卡卡」：**核心概念 / 你会看到 / 思考题 / 通关标准 / 前置标签**。
**思考题动手改过代码，才算通关。**

## 关卡表

| 编号 | 文件 | 主题 | 前置 |
| :--- | :--- | :--- | :--- |
| W-01 | `01_float64_trap.go` | **float64 为什么不能表示钱**：0.1+0.2、累加误差、ULP、大数丢位、对齐 stepSize 后被拒单 | — |
| W-02 | `02_decimal_basics.go` | decimal 基本用法：`NewFromString` vs `NewFromFloat`、比较、互转 | W-01 |
| W-03 | `03_decimal_rounding.go` | 精度对齐与舍入：九种 RoundingMode、银行家舍入、对账差额 | W-02 |
| W-04 | `04_keys_address.go` | 私钥 → 公钥 → 地址、EIP-55 checksum、keystore 加解密 | L1-22 |
| W-05 | `05_hd_wallet.go` | BIP-39 助记词与 BIP-44 派生 `m/44'/60'/0'/0/i`、硬化与非硬化 | W-04 |
| W-06 | `06_abi_encode.go` | ABI 编解码：函数选择器、`Pack`/`Unpack`、动态类型的头尾结构 | L1-22 |
| W-07 | `07_abi_events.go` | 事件与日志：`topic0`、`indexed` 参数、解析 ERC20 Transfer | W-06 |
| W-08 | `08_sign_tx.go` | EIP-1559 交易构造、离线签名、从签名恢复发送者、链 ID 防重放 | W-04 |
| W-09 | `09_order_state_machine.go` | 订单状态机：合法跃迁表、幂等、部分成交、不变量校验 | L2-13 |
| W-10 | `10_ws_reconnect.go` | WS 断线重连、指数退避加抖动、心跳、**断线补数** | L2-11 |
| W-11 | `11_metrics.go` | Prometheus：Counter/Gauge/Histogram、标签、`/metrics` 端点 | L3-05 |
| W-12 | `12_rpc_readonly.go` | ethclient 只读：chainID/余额/nonce/gas，**连不上自动降级** | W-03, L3-14 |

编号顺序即推荐学习顺序。**建议在学完 `04-goframe`（L3）之后、进入 `05-trading-bot`（L4）之前学习** ——
W-09/W-10/W-11 都需要 L2 的并发基础，W-12 需要 L3 的错误处理与配置习惯。

## 目录结构

```
07-web3-trading/
├── go.mod      # module go-study/web3-trading
├── README.md
├── 01_float64_trap.go
├── ...
└── 12_rpc_readonly.go
```

## 离线也能跑（很重要）

这一模块**大部分关卡完全离线**，不需要任何 RPC 节点：

| 关卡 | 是否需要网络 |
| :--- | :--- |
| W-01 ~ W-09、W-11 | ❌ 不需要（纯本地计算 / 进程内假服务） |
| W-10 | ❌ 不需要（用 `httptest` 在进程内起一个假交易所 WS 服务） |
| W-12 | ⚠️ 需要（读真实链上数据），**连不上会明确跳过并打印配置方法，不会卡死** |

W-12 默认走几个公开节点，也支持环境变量覆盖：

```sh
ETH_RPC_URL=https://你的节点 go run 12_rpc_readonly.go
```

## 依赖

`go.mod` 里有四个外部依赖，都是这一领域的标准件：

| 依赖 | 用途 |
| :--- | :--- |
| `github.com/ethereum/go-ethereum` | 密钥、地址、ABI、交易类型、签名、ethclient |
| `github.com/shopspring/decimal` | 定点小数（金额与数量） |
| `github.com/tyler-smith/go-bip39` / `go-bip32` | 助记词与 HD 派生 |
| `github.com/gorilla/websocket` | 行情 WS 连接（与 `05-trading-bot` 一致） |
| `github.com/prometheus/client_golang` | 指标 |

## 实测记录

关卡里写的数字都是真跑出来的，几处值得单独拎出来：

| 关卡 | 实测 |
| :--- | :--- |
| W-01 | `0.1+0.2` 与 `0.3` 只差 **1 个 ULP**，但 `==` 判 false；循环加 1000 次 `0.1` 得到 `99.99999999999859312538`；**常量 `0.1+0.2 == 0.3` 竟然是 true** —— 因为 Go 常量是任意精度的，编译期折叠算准了，只有落到变量才会走 float64 |
| W-02 | `NewFromFloat(0.1+0.7)` 会**把浮点误差一起带进来**（`0.7999999999999999`），乘 1e18 后差 100 wei |
| W-03 | `1.005` 在 float64 里其实是 `1.0049999999999999`，`math.Round(x*100)/100` 给 `1.00`，decimal 给 `1.01` |
| W-05 | 用公开测试助记词派生的前三个地址与 Hardhat / Anvil 默认账户完全一致 —— 可以用它对照任何以太坊工具链 |
| W-08 | 用 chainID=1 签名、用 chainID=56 的 signer 恢复 → `invalid chain id for signer`，**这就是防重放** |
| W-10 | 假交易所每 4 条断开一次，客户端实测重连 3 次、退避 `87ms→229ms→398ms`，补回 28 条缺口消息，合并后 **seq 1..41 无空洞** |
| W-12 | 真实读到 chainId=1、区块 26,173,244；余额用 float64 表示会**差 182 wei** |

## 通关标准

能说清这四句话：

1. **金额和数量永远用 decimal（或整数最小单位），永远不用 float64** —— 并且能说出为什么；
2. **私钥 → 公钥 → 地址**这条链每一步用了什么算法，以及 keystore 为什么必须加密存储；
3. **ABI 是合约的调用约定**：选择器 4 字节 + 参数，动态类型靠偏移量寻址；
4. **交易机器人的三个必备件**：订单状态机（含幂等）、能补数的行情连接、能看 P99 的指标。

## 下一站

[L4 · 交易机器人终点项目](../05-trading-bot/)。把这一模块的零件接起来 ——
金额用 decimal、行情走 WS 重连、订单进状态机、指标暴露给 Prometheus。
