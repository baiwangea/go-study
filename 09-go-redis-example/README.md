# Go 操作 Redis 示例 (`go-redis`)

本项目使用流行的 `github.com/go-redis/redis` 库，演示如何在 Go 应用程序中与 Redis 数据库进行交互。

## 先决条件

在运行本示例之前，请确保你的本地计算机上已经安装并运行了 Redis 服务器。默认情况下，程序会尝试连接 `localhost:6379`。

## 如何运行示例

1.  **安装依赖**: 在项目目录下，运行 `go mod tidy` 来下载 `go-redis` 库。

    ```sh
    go mod tidy
    ```

2.  **运行程序**: 直接运行 `main.go` 文件。

    ```sh
    go run main.go
    ```

程序将连接到 Redis，并按顺序执行一系列操作示例。

---

## 示例详解

### `stringExample()` - 字符串操作

这是 Redis 最基本的数据类型。这个示例展示了：

*   **`rdb.Set(ctx, key, value, expiration)`**: 设置一个键值对。`expiration` 设置为 `0` 表示永不过期。
*   **`rdb.Get(ctx, key)`**: 根据键获取对应的值。
*   **处理 `redis.Nil`**: 当 `Get` 一个不存在的键时，`go-redis` 会返回一个特殊的错误 `redis.Nil`。代码中演示了如何正确地检查和处理这种情况，而不是将其视为一个程序异常。

### `hashExample()` - 哈希操作

哈希（Hash）非常适合用来存储对象。你可以把一个对象的多个字段存储在同一个哈希键中，从而方便地管理整个对象。

*   **`rdb.HSet(ctx, key, values)`**: 将一个 Go 的 `map` 直接设置为一个哈希键的值。`go-redis` 会自动处理 `map` 到哈希字段的转换。
*   **`rdb.HGet(ctx, key, field)`**: 获取哈希键中单个字段的值。
*   **`rdb.HGetAll(ctx, key)`**: 获取哈希键中所有的字段和值，返回一个 `map[string]string`。

### `listExample()` - 列表操作

列表（List）是一个字符串元素的集合，按照插入顺序排序。你可以把它用作队列或栈。

*   **`rdb.LPush(ctx, key, values...)`**: 将一个或多个值推入列表的左侧（头部）。
*   **`rdb.RPush(ctx, key, values...)`**: 将一个或多个值推入列表的右侧（尾部）。
*   **`rdb.LRange(ctx, key, start, stop)`**: 获取列表中指定范围的元素。`0` 是第一个元素，`-1` 是最后一个元素。

### `expireExample()` - 过期时间

这个示例演示了如何为键设置一个自动过期时间（Time-To-Live, TTL）。这对于缓存场景非常有用。

*   在 `rdb.Set` 中，将最后一个参数设置为一个大于 `0` 的 `time.Duration`，例如 `5*time.Second`。
*   程序会先获取一次键的值，然后等待超过过期时间（6秒），再次尝试获取。第二次获取时，由于键已经自动被 Redis 删除，程序会收到 `redis.Nil` 错误，从而验证了过期功能。
