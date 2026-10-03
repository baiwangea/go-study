# Go 操作 MongoDB 示例 (`mongo-go-driver`)

本项目使用官方的 Go 语言驱动 `go.mongodb.org/mongo-driver`，演示如何在 Go 应用程序中与 MongoDB 数据库进行连接，并执行完整的 CRUD（创建、读取、更新、删除）操作。

## 先决条件

在运行本示例之前，请确保你的本地计算机上已经安装并运行了 MongoDB 服务器。默认情况下，程序会尝试连接 `mongodb://localhost:27017`。

## 如何运行示例

1.  **安装依赖**: 在项目目录下，运行 `go mod tidy` 来下载 `mongo-go-driver`。

    ```sh
    go mod tidy
    ```

2.  **运行程序**: 直接运行 `main.go` 文件。

    ```sh
    go run main.go
    ```

程序将连接到 MongoDB，清空 `testdb` 数据库下的 `users` 集合，然后按顺序执行一系列 CRUD 操作示例，最后关闭连接。

---

## 示例详解

### 连接到 MongoDB

*   **`options.Client().ApplyURI(...)`**: 这是连接到 MongoDB 的标准方式，通过一个连接字符串来指定主机和端口。
*   **`context.WithTimeout(...)`**: 在执行数据库操作时，设置一个带有超时的 `context` 是一个非常好的实践。这可以防止因为网络问题或数据库无响应而导致程序永久阻塞。
*   **`mongo.Connect(...)`**: 使用指定的选项和上下文建立连接。
*   **`client.Ping(...)`**: 在执行任何操作之前，先 `Ping` 一下数据库，以验证连接是否真正建立成功。

### `createExample()` - 创建文档

*   **BSON**: MongoDB 在后台使用 BSON（Binary JSON）格式存储数据。`go.mongodb.org/mongo-driver/bson` 包提供了在 Go 结构体和 BSON 之间进行转换的工具。我们在 `User` 结构体中使用的 `bson:"..."` 标签就是用来控制这个映射关系的。
*   **`collection.InsertOne(ctx, document)`**: 插入单个文档。`document` 可以是一个 `bson.D` 对象或是一个定义了 `bson` 标签的 Go 结构体。
*   **`collection.InsertMany(ctx, documents)`**: 插入多个文档。`documents` 参数必须是一个 `[]interface{}` 类型的切片。

### `readExample()` - 读取文档

*   **`bson.D`**: 这是 `mongo-go-driver` 中用来构建查询条件（过滤器）的一种有序的键值对类型。使用 `bson.D` 可以确保字段的顺序。
*   **`collection.FindOne(ctx, filter).Decode(&result)`**: 根据 `filter` 查询单个文档，并使用 `.Decode()` 方法将结果解码到一个 Go 结构体变量中。
*   **`collection.Find(ctx, filter)`**: 查询多个匹配的文档，它会返回一个 `Cursor`（游标）。
*   **遍历游标**: 你需要使用一个 `for cur.Next(ctx)` 循环来遍历查询结果。在循环内部，使用 `cur.Decode(&elem)` 来获取每一条文档。处理完游标后，记得使用 `defer cur.Close(ctx)` 来关闭它，以释放资源。

### `updateExample()` - 更新文档

*   **更新操作符 (`$set`)**: MongoDB 使用特殊的更新操作符来修改文档中的字段。`$set` 是最常用的一个，用于设置或更新一个字段的值。
*   **`collection.UpdateOne(ctx, filter, update)`**: 根据 `filter` 找到第一个匹配的文档，然后根据 `update` 指令对其进行修改。它返回一个 `UpdateResult` 对象，其中包含了匹配和修改的文档数量等信息。

### `deleteExample()` - 删除文档

*   **`collection.DeleteOne(ctx, filter)`**: 根据 `filter` 找到并删除第一个匹配的文档。
*   **`collection.DeleteMany(ctx, filter)`**: 如果你想删除所有匹配的文档，可以使用 `DeleteMany`。
