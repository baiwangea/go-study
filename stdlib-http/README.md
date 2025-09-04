# Go 标准库 `net/http` 学习笔记

本项目旨在通过一系列循序渐进的示例，深入讲解 Go 语言内置的 `net/http` 标准库的用法。`net/http` 库非常强大，足以用来构建生产级别的 HTTP 客户端和服务器。

## 项目结构

```
/stdlib-http
├── client/                  # HTTP 客户端示例
│   └── client_examples.go
├── server/                  # HTTP 服务器示例
│   └── server_examples.go
├── advanced/                # 更高级的服务器功能示例
│   └── advanced_examples.go
├── main.go                  # 项目主入口，用于运行不同模块的示例
└── go.mod
```

## 如何运行示例

你可以通过 `main.go` 程序来运行不同类别的示例。

1.  **运行客户端示例**:

    ```sh
    go run main.go client
    ```

    这会按顺序执行所有客户端相关的 HTTP 请求示例。

2.  **运行服务器或高级示例**:

    由于服务器示例会启动一个长时间运行的进程，你需要手动选择要运行的示例。

    *   打开 `main.go` 文件。
    *   在 `runServerExamples()` 或 `runAdvancedExamples()` 函数中，**取消注释** 你想运行的那个函数调用。
    *   执行命令:
        ```sh
        # 运行服务器示例
        go run main.go server

        # 运行高级功能示例
        go run main.go advanced
        ```

---

## 示例详解

### 客户端 (`client/`)

1.  **`BasicGet()`**: 展示了如何发起一个最基础的 HTTP GET 请求，并读取和打印响应的状态码与内容。
2.  **`GetWithHeaders()`**: 演示了如何创建一个自定义的 `http.Request`，并为其添加自定义的 HTTP 头部（如 `User-Agent`），然后通过 `http.Client` 发送出去。
3.  **`PostJSON()`**: 演示了如何将一个 Go 的 `map` 结构体序列化为 JSON，然后通过 HTTP POST 请求将其作为请求体发送出去。
4.  **`CustomClient()`**: 展示了如何创建一个自定义的 `http.Client`，并为其设置超时时间（Timeout）。这对于避免程序因等待一个无响应的服务器而永久阻塞至关重要。

### 服务器 (`server/`)

1.  **`BasicServer()`**: 启动一个最简单的 HTTP 服务器。它使用 `http.HandleFunc` 为根路径 `/` 注册了一个处理器函数，所有访问该路径的请求都会得到一个固定的文本响应。
2.  **`ServerWithMux()`**: 演示了 `http.ServeMux`（服务多路复用器）的用法。`ServeMux` 就像一个路由器，可以让你为不同的 URL 路径注册不同的处理器，从而构建出更复杂的 Web 服务。
3.  **`ServerWithMiddleware()`**: 展示了中间件（Middleware）的概念。中间件是一个包装了 `http.Handler` 的函数，它允许你在请求被最终处理之前或之后执行一些通用逻辑，例如记录日志、身份验证、添加通用头部等。这个例子实现了一个简单的日志中间件。

### 高级功能 (`advanced/`)

1.  **`ParseFormParams()`**: 演示了如何在服务器端解析来自客户端的参数。这包括解析 URL 中的查询参数（Query Parameters，如 `/search?q=golang`）和 POST 请求体中的表单数据（Form Data）。
2.  **`FileUpload()`**: 构建了一个完整的 HTML 表单和后端处理器，用于接收客户端上传的文件。它展示了如何从 `multipart/form-data` 请求中解析文件，并将其保存到服务器本地。
3.  **`FileServer()`**: 演示了如何使用 `http.FileServer` 快速搭建一个静态文件服务器。只需几行代码，你就可以将一个目录下的所有文件（如 HTML, CSS, JS, 图片）暴露给外部访问。
