# Go Study: A Hands-On Learning Repository

<p align="center">
  <a href="README.zh.md">中文</a> |
  <strong>English</strong>
</p>


Welcome to **Go Study**, a curated collection of practical, hands-on examples designed to accelerate your journey into the Go programming language. Each module in this repository is a self-contained project, complete with detailed code, explanations, and its own `README.md`.

## 🚀 Modules Overview

This repository is organized into several modules, each focusing on a specific area of Go development.

### Core Language Features

| Module | Description |
| :--- | :--- |
| [`go-fundamentals`](./go-fundamentals/) | Covers the building blocks of Go: functions, packages, and interfaces. |
| [`go-pointers`](./go-pointers/) | A deep dive into pointers and the difference between pass-by-value and pass-by-pointer. |
| [`go-concurrency`](./go-concurrency/) | Explores Go's powerful concurrency primitives: Goroutines, Channels, Select, Mutexes, and WaitGroups. |
| [`go-data-structures`](./go-data-structures/) | Demonstrates Go's built-in data structures (slices, maps, structs) and how to implement a Set. |

### Standard Library in Action

| Module | Description |
| :--- | :--- |
| [`stdlib-http`](./stdlib-http/) | Practical examples of building HTTP clients and servers using the `net/http` package. |
| [`stdlib-logger`](./stdlib-logger/) | A guide to using the standard `log` package for effective logging. |

### Backend Development & Tooling

| Module | Description |
| :--- | :--- |
| [`gin-framework-example`](./gin-framework-example/) | A complete project scaffold for a modern web application using the Gin framework. |
| [`go-jwt-example`](./go-jwt-example/) | Learn how to create and validate JSON Web Tokens (JWT) for stateless authentication. |
| [`go-mysql-example`](./go-mysql-example/) | Demonstrates safe and efficient MySQL operations using the standard `database/sql` package. |
| [`go-mongodb-example`](./go-mongodb-example/) | A guide to performing CRUD operations on a MongoDB database with the official driver. |
| [`go-redis-example`](./go-redis-example/) | Shows how to interact with Redis for caching and other use cases using `go-redis`. |

## 💡 How to Use

1.  **Clone the repository**:
    ```sh
    git clone <repository-url>
    ```
2.  **Navigate to a module**: Each directory is a standalone Go project.
    ```sh
    cd go-redis-example
    ```
3.  **Run the example**: Most modules have a `main.go` file that you can run directly.
    ```sh
    go run main.go
    ```
4.  **Read the documentation**: Be sure to read the `README.md` file inside each module for a detailed explanation of the concepts and code.

Happy Coding! ✨
