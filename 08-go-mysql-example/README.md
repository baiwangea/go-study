# Go 操作 MySQL 示例 (`database/sql`)

本项目使用 Go 语言内置的 `database/sql` 包和流行的 `github.com/go-sql-driver/mysql` 驱动，演示了如何在 Go 应用程序中安全、高效地与 MySQL 数据库进行交互。

## 先决条件

1.  **MySQL 服务器**: 确保你的本地或网络中有一个正在运行的 MySQL 服务器。
2.  **数据库**: 在你的 MySQL 服务器中创建一个数据库，名为 `testdb`。
    ```sql
    CREATE DATABASE testdb;
    ```
3.  **配置 DSN**: 打开 `main.go` 文件，找到 `dsn` 变量，**将其中的 `root:password` 替换为你自己的 MySQL 用户名和密码**。

## 如何运行示例

1.  **安装依赖**: 在项目目录下，运行 `go mod tidy` 来下载 `mysql` 驱动。

    ```sh
    go mod tidy
    ```

2.  **运行程序**: 直接运行 `main.go` 文件。

    ```sh
    go run main.go
    ```

程序将连接到 MySQL，自动创建 `users` 表（如果不存在），然后按顺序执行一系列 CRUD（创建、读取、更新、删除）操作示例。

---

## 核心概念与示例详解

### `database/sql` 包

`database/sql` 是 Go 语言操作数据库的标准接口。它本身不提供任何特定数据库的驱动，而是定义了一套通用的、抽象的接口。你需要通过导入第三方驱动（如 `go-sql-driver/mysql`）来与具体的数据库进行通信。

*   **匿名导入**: 我们使用 `_ "github.com/go-sql-driver/mysql"` 的方式导入驱动。`_` 表示我们只需要执行该包的 `init()` 函数，让它自行向 `database/sql` 注册自己，而我们不会在代码中直接使用这个包的任何导出成员。
*   **连接池**: `sql.Open` 返回的 `*sql.DB` 对象并不是一个单一的数据库连接，而是一个**连接池**。它会为你管理底层的多个连接，处理连接的复用和生命周期，非常高效。你应该在程序启动时创建一次，然后在不同的函数间共享这个 `*sql.DB` 对象。

### 连接数据库

*   **DSN (Data Source Name)**: 这是一个格式化的字符串，包含了连接数据库所需的所有信息（用户名、密码、主机、端口、数据库名等）。
*   **`sql.Open("mysql", dsn)`**: 打开一个到数据库的连接池。注意，`Open` 不会立即建立连接，只是验证参数。
*   **`db.Ping()`**: 这是验证 DSN 是否正确、数据库是否可达的最佳方式。它会尝试与数据库建立一个连接并立即关闭。

### `createTable()` - 执行无返回值的查询

*   **`db.Exec(query)`**: 用于执行不返回数据行的 SQL 语句，例如 `CREATE TABLE`, `INSERT`, `UPDATE`, `DELETE`。它返回一个 `sql.Result` 对象，从中可以获取 `LastInsertId`（最后插入行的 ID）或 `RowsAffected`（受影响的行数）。

### `createExample()` & `updateExample()` & `deleteExample()` - 预处理语句

*   **`db.Prepare(query)`**: 这是**防止 SQL 注入攻击**的最重要方法。它将 SQL 查询语句发送到数据库进行预编译，然后你可以安全地将用户输入作为参数传递给 `stmt.Exec()`。
*   **`stmt.Exec(args...)`**: 执行预处理语句，并将 `args` 安全地传递给查询中的 `?` 占位符。
*   **`result.LastInsertId()`**: 在执行 `INSERT` 语句后，用它来获取新创建行的自增 ID。
*   **`result.RowsAffected()`**: 在执行 `UPDATE` 或 `DELETE` 后，用它来获取受影响的行数。

### `readExample()` - 执行数据查询

*   **`db.QueryRow(query, args...)`**: 用于查询**单行**数据。它返回一个 `*sql.Row` 对象。你需要紧接着调用 `.Scan()` 方法将查询到的列数据映射到 Go 变量中。
    *   **`sql.ErrNoRows`**: 如果 `QueryRow` 没有找到任何行，`.Scan()` 会返回这个特殊的错误。你应该总是检查这个错误，以优雅地处理“未找到”的情况。
*   **`db.Query(query, args...)`**: 用于查询**多行**数据。它返回一个 `*sql.Rows` 对象，你需要像遍历迭代器一样处理它。
    *   **`for rows.Next() { ... }`**: 使用这个循环来遍历结果集中的每一行。
    *   **`rows.Scan(...)`**: 在循环内部，使用 `Scan` 将当前行的数据读入 Go 变量。
    *   **`defer rows.Close()`**: 查询结束后，必须调用 `rows.Close()` 来释放数据库连接。使用 `defer` 是确保它被执行的最佳方式。
