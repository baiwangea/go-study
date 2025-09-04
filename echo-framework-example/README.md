# Echo 框架现代化项目脚手架

本项目提供了一个使用 [Echo](https://echo.labstack.com/) 框架的、具有现代化清晰目录结构的高性能 Web 应用程序脚手架。

Echo 是一个高性能、可扩展、极简的 Go Web 框架。

## 项目结构

```
/echo-framework-example
├── cmd/
│   └── main.go              # 程序主入口
├── internal/                # 内部应用逻辑，此包外不可见
│   ├── app/
│   │   ├── handler/         # HTTP 处理器
│   │   └── router/          # 路由定义
│   └── conf/
│       └── config.go        # 应用配置
├── go.mod                   # Go 模块文件
└── README.md                # 项目说明
```

## 如何运行示例

1.  **安装依赖**: 在项目目录下，运行 `go mod tidy` 来下载 `echo` 框架。

    ```sh
    go mod tidy
    ```

2.  **运行程序**: 直接运行 `main.go` 文件。

    ```sh
    go run cmd/main.go
    ```

3.  **测试**: 服务器启动后，你可以访问 `http://localhost:1323/ping`，你应该会看到 `pong` 的响应。

---

## 核心概念

*   **`echo.New()`**: 创建一个 Echo 框架的实例。
*   **中间件 (Middleware)**: Echo 的中间件是在 HTTP 请求被处理器处理之前或之后执行的函数。我们在 `main.go` 中使用了两个内置的中间件：
    *   `middleware.Logger()`: 记录每个请求的详细信息。
    *   `middleware.Recover()`: 从 `panic` 中恢复，防止服务器崩溃。
*   **路由 (Routing)**: `e.GET("/ping", handler.Ping)` 定义了一个 HTTP GET 路由。当有请求访问 `/ping` 路径时，`handler.Ping` 函数将被调用。
*   **处理器 (Handler)**: 处理器函数接收一个 `echo.Context` 类型的参数，它包含了请求和响应的所有信息。处理器必须返回一个 `error`。
*   **响应 (Response)**: `c.String(http.StatusOK, "pong")` 是一个简单的响应方法，它向客户端发送一个状态码为 `200 OK`、内容为 `pong` 的纯文本响应。
