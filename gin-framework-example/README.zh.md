[English](./README.md)

# Gin 框架示例

这是一个使用 Gin 框架构建的现代化 Web 应用程序的完整项目脚手架。它包含了多种最佳实践和功能，可帮助您快速入门。

## 特性

*   **基于环境的配置:** 使用命令行标志轻松管理开发、测试和生产环境的配置。
*   **服务层:** 专门的服务层，用于将业务逻辑与处理程序分开。
*   **中间件:** 包括用于请求和响应日志记录的记录器中间件，以及用于身份验证的 JWT 中间件。
*   **数据库集成:** 使用 GORM 进行数据库交互，并包含数据库初始化模块。
*   **Redis 集成:** 包括用于缓存和其他目的的 Redis 客户端。
*   **验证码:** 用于防止机器人攻击的验证码端点。
*   **结构化日志记录:** 带有日志轮换功能的文件式日志记录系统。
*   **用于 API 响应的 DTO:** 使用数据传输对象（DTO）来控制 API 返回的字段。

## 目录结构

```
gin-framework-example/
├── conf/                  # 配置文件
│   ├── config.dev.yaml
│   └── config.prod.yaml
├── src/
│   ├── app/
│   │   ├── handler/
│   │   ├── middleware/
│   │   ├── model/
│   │   ├── response/
│   │   ├── router/
│   │   └── service/
│   ├── cmd/
│   │   └── main.go
│   └── pkg/
│       ├── db/
│       ├── e/
│       └── util/
├── .gitignore
└── README.md
```

## 如何运行

1.  **安装依赖:**

    ```bash
    go mod tidy
    ```

2.  **运行应用:**

    您可以使用 `-env` 标志指定环境。默认环境是 `dev`。

    *   **开发:**

        ```bash
        go run src/cmd/main.go
        ```

        或

        ```bash
        go run src/cmd/main.go -env=dev
        ```

    *   **生产:**

        ```bash
        go run src/cmd/main.go -env=prod
        ```

## 构建生产版本

要为生产环境创建更小的二进制文件，请使用 `-ldflags="-s -w"` 标志来剥离调试信息。

```bash
go build -ldflags="-s -w" -o gin-framework-example src/cmd/main.go
```

## 交叉编译

### 1. Mac 编译 Linux/Windows 64位可执行程序:

```bash
# 编译 Linux
go env -w CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# 编译 Windows
go env -w CGO_ENABLED=0 GOOS=windows GOARCH=amd64
```

### 2. Linux 编译 Mac/Windows 64位可执行程序:

```bash
# 编译 Mac
go env -w CGO_ENABLED=0 GOOS=darwin GOARCH=amd64

# 编译 Windows
go env -w CGO_ENABLED=0 GOOS=windows GOARCH=amd64
```

### 3. Windows 编译 Mac/Linux 64位可执行程序:

```bash
# 编译 Mac
go env -w CGO_ENABLED=0 GOOS=darwin GOARCH=amd64

# 编译 Linux
go env -w CGO_ENABLED=0 GOOS=linux GOARCH=amd64
```

### 执行编译

```bash
go build -o <output_path> src/cmd/main.go
```

### 日常开发流程

*   **Mac 编译 Linux:**

    ```bash
    go env -w CGO_ENABLED=0 GOOS=linux GOARCH=amd64
    ```

*   **切换回 Mac:**

    ```bash
    go env -w CGO_ENABLED=1 GOOS=darwin GOARCH=amd64
    ```
