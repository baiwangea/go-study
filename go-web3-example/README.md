# Go 与 Web3 交互示例 (go-ethereum)

本项目使用官方的 Go 语言库 `github.com/ethereum/go-ethereum` (通常称为 `geth`)，演示了如何在 Go 应用程序中与以太坊区块链进行交互。

## 先决条件

**以太坊 RPC 端点**: 要运行此示例，你需要一个以太坊节点的 RPC (Remote Procedure Call) 端点 URL。你可以通过以下方式获取：

1.  **第三方服务 (推荐)**: 在 [Infura](https://infura.io/) 或 [Alchemy](https://www.alchemy.com/) 等网站上注册一个免费账户，创建一个新的应用，你将获得一个免费的以太坊主网 RPC 端点 URL。
2.  **本地节点**: 如果你正在本地运行一个以太坊节点 (如 `geth` 或 `ganache`)，你可以使用其本地 RPC 地址 (通常是 `http://127.0.0.1:8545`)。

获取 URL 后，请**务必替换 `main.go` 文件中 `rpcEndpoint` 变量的值**。

## 如何运行示例

1.  **安装依赖**: 在项目目录下，运行 `go mod tidy` 来下载 `go-ethereum` 及其依赖项。

    ```sh
    go mod tidy
    ```

2.  **运行程序**: 直接运行 `main.go` 文件。

    ```sh
    go run main.go
    ```

程序将连接到你配置的以太坊节点，并按顺序执行一系列交互示例。

---

## 示例详解

### 连接到以太坊客户端

*   **`ethclient.Dial(rpcEndpoint)`**: 这是连接到以太坊节点的入口点。它接收一个 RPC 端点 URL，并返回一个 `*ethclient.Client` 实例，后续的所有操作都将通过这个客户端实例进行。

### `queryLatestBlock()` - 查询最新区块

*   **`client.HeaderByNumber(ctx, nil)`**: `nil` 参数表示我们想要获取最新的区块头。区块头包含了区块号、时间戳、矿工地址等元数据。

### `queryAccountBalance()` - 查询账户余额

*   **`common.HexToAddress(hexString)`**: 将一个十六进制的字符串地址转换为以太坊的 `Address` 类型。
*   **`client.BalanceAt(ctx, address, nil)`**: 查询指定地址在最新区块 (`nil`) 上的余额。 
*   **单位转换**: 以太坊的余额单位是 `Wei`，这是最小的单位。1 Ether = 10^18 Wei。代码中演示了如何使用 `math/big` 包将 `Wei` 安全地转换为更易读的 `ETH`。

### `generateNewWallet()` - 生成新钱包

*   **`crypto.GenerateKey()`**: 生成一个新的 ECDSA (椭圆曲线数字签名算法) 私钥。这是你的钱包的核心和凭证。
*   **`crypto.PubkeyToAddress(...)`**: 从公钥派生出公开的以太坊地址。地址是公钥的 Keccak-256 哈希的最后20个字节。
*   **安全警告**: 示例中打印了十六进制格式的私钥，这**绝对不能在生产环境中使用**。私钥必须被安全地存储和管理，一旦泄露，钱包中的所有资产都将面临风险。

### `interactWithContract()` - 与智能合约交互 (占位符)

这个函数是一个占位符，展示了与智能合约交互的基本思路。在实际应用中，你会使用 `abigen` 工具从合约的 ABI (Application Binary Interface) 生成 Go 代码，从而可以像调用 Go 函数一样调用智能合约的函数。
