# Go 标准库 `log` 学习笔记

本项目通过一系列示例，展示了 Go 语言内置 `log` 包的各种功能。`log` 包提供了简单、易用的日志记录功能，是 Go 程序中最常用的日志解决方案。

## 项目结构

```
/stdlib-logger
├── logger_examples.go   # 所有日志功能的示例代码
├── main.go              # 项目主入口，按顺序调用所有示例
├── go.mod
└── app.log              # LogToFile() 示例生成的日志文件
```

## 如何运行示例

直接运行 `main.go` 即可。程序会自动按顺序执行 `logger_examples.go` 文件中定义的所有演示函数。

```sh
go run .
```

**注意**: `LogToFile()` 函数会在项目根目录下创建一个 `app.log` 文件并写入日志。`FatalAndPanic()` 中的示例默认是注释掉的，因为它们会中断程序的正常执行。

---

## 示例详解

### `BasicLogging()`

展示了 `log` 包最基本的三种用法：

*   `log.Println()`: 打印一行日志，会自动在末尾添加换行符。
*   `log.Printf()`: 允许你使用格式化字符串来构造日志消息，用法与 `fmt.Printf` 完全相同。
*   默认情况下，所有日志都会被输出到标准错误流（`os.Stderr`）。

### `LogToFile()`

演示了如何将日志输出重定向到一个文件而不是标准错误。这对于在生产环境中持久化应用日志至关重要。

*   使用 `os.OpenFile` 来创建或打开一个日志文件（`app.log`）。
*   通过 `log.SetOutput()` 函数，将标准 logger 的输出目标设置为我们刚刚打开的文件句柄。
*   此后，所有通过 `log.Println`, `log.Printf` 等函数产生的日志都会被写入该文件。

### `CustomizeLogger()`

展示了如何自定义标准 logger 的输出格式。

*   **`log.SetPrefix()`**: 为每一行日志添加一个固定的前缀，例如 `[MyApp]`，这有助于在聚合多个服务的日志时区分来源。
*   **`log.SetFlags()`**: 控制每条日志前缀包含哪些信息。常用的标志有：
    *   `log.Ldate`: 日期 (e.g., `2023/10/27`)
    *   `log.Ltime`: 时间 (e.g., `10:30:00`)
    *   `log.Lmicroseconds`: 微秒
    *   `log.Lshortfile`: 文件名和行号 (e.g., `main.go:25`)
    *   `log.Llongfile`: 完整文件路径和行号
    *   `log.LstdFlags`: `log.Ldate | log.Ltime` 的默认组合

### `CreateNewLogger()`

除了使用全局的标准 logger，你还可以使用 `log.New()` 创建自己的 logger 实例。这在以下场景中非常有用：

*   希望将不同模块的日志输出到不同的地方（例如，一个 logger 写文件，一个 logger 写 `os.Stdout`）。
*   希望为不同的日志级别（如 INFO, WARN, ERROR）使用不同的前缀。

`log.New()` 接收三个参数：输出目标（`io.Writer`），日志前缀（`string`）和日志标志（`int`）。

### `FatalAndPanic()`

演示了 `log` 包中两个会中断程序执行的函数。

*   **`log.Fatal()`**: 打印日志消息后，会立即调用 `os.Exit(1)` 来终止程序。它通常用于报告一个无法恢复的、致命的错误（例如，数据库连接失败，配置文件缺失）。
*   **`log.Panic()`**: 打印日志消息后，会引发一个 `panic`。与 `Fatal` 不同，`panic` 可以被 `defer` 和 `recover` 捕获，从而允许程序进行一些清理工作或尝试从恐慌中恢复。如果不被 `recover`，`panic` 会导致程序崩溃。
