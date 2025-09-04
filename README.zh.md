# Go Study: Go 语言实战学习仓库

<p align="center">
  <strong>中文</strong> |
  <a href="README.md">English</a>
</p>


欢迎来到 **Go Study**，这是一个精心设计的实战示例集合，旨在加速你的 Go 语言学习之旅。本仓库中的每个模块都是一个独立的项目，包含了详细的代码、注释以及专属的 `README.md` 文档。

## 🚀 模块总览

本仓库被划分为多个模块，每个模块都专注于 Go 开发的一个特定领域。

### 核心语言特性

| 模块 | 描述 |
| :--- | :--- |
| [`go-fundamentals`](./go-fundamentals/) | 涵盖 Go 语言的基石：函数、包和接口。 |
| [`go-pointers`](./go-pointers/) | 深入讲解指针，以及“值传递”与“指针传递”的根本区别。 |
| [`go-concurrency`](./go-concurrency/) | 探索 Go 强大的并发原语：Goroutines, Channels, Select, Mutexes, 和 WaitGroups。 |
| [`go-data-structures`](./go-data-structures/) | 演示 Go 的内置数据结构（切片、映射、结构体）以及如何实现一个集合 (Set)。 |

### 标准库实战

| 模块 | 描述 |
| :--- | :--- |
| [`stdlib-http`](./stdlib-http/) | 使用 `net/http` 包构建 HTTP 客户端和服务器的实用示例。 |
| [`stdlib-logger`](./stdlib-logger/) | 使用标准 `log` 包进行有效日志记录的指南。 |

### 后端开发与工具

| 模块 | 描述 |
| :--- | :--- |
| [`gin-framework-example`](./gin-framework-example/) | 一个使用 Gin 框架的现代化 Web 应用程序的完整项目脚手架。 |
| [`go-jwt-example`](./go-jwt-example/) | 学习如何创建和验证用于无状态身份验证的 JSON Web Tokens (JWT)。 |
| [`go-mysql-example`](./go-mysql-example/) | 演示如何使用标准 `database/sql` 包安全高效地操作 MySQL。 |
| [`go-mongodb-example`](./go-mongodb-example/) | 使用官方驱动对 MongoDB 数据库执行 CRUD 操作的指南。 |
| [`go-redis-example`](./go-redis-example/) | 展示如何使用 `go-redis` 与 Redis 交互以实现缓存等功能。 |

## 💡 如何使用

1.  **克隆仓库**:
    ```sh
    git clone <repository-url>
    ```
2.  **进入一个模块**: 每个目录都是一个独立的 Go 项目。
    ```sh
    cd go-redis-example
    ```
3.  **运行示例**: 大多数模块都有一个可以直接运行的 `main.go` 文件。
    ```sh
    go run main.go
    ```
4.  **阅读文档**: 请务必阅读每个模块内的 `README.md` 文件，以获取对概念和代码的详细解释。

编程愉快！ ✨
