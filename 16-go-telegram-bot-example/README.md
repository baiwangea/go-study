# Go 发送 Telegram Bot 消息示例 (实战升级版)

本项目使用流行的 `github.com/go-telegram-bot-api/telegram-bot-api` 库，演示了如何在 Go 应用程序中通过 Bot 发送各种类型的消息到 Telegram，并集成了真实世界的运维和监控场景。

## 先决条件

在运行本示例之前，你必须获取两样东西：**Bot Token** 和 **Chat ID**，并将它们设置为环境变量。

### 1. 如何获取 Bot Token

1.  在 Telegram 中，搜索并打开与 **`@BotFather`** 的聊天。
2.  发送 `/newbot` 命令。
3.  按照提示为你的新 Bot 设置一个名字 (Name) 和一个用户名 (Username)。用户名必须以 `bot` 结尾，例如 `my_test_bot`。
4.  创建成功后，BotFather 会给你一串 **HTTP API token**。这就是你的 **Bot Token**。请妥善保管它。

### 2. 如何获取 Chat ID

Chat ID 是你希望 Bot 发送消息的目标聊天的唯一标识。

*   **私聊 (最简单)**:
    1.  找到你刚刚创建的 Bot 并向它发送一条消息（例如，发送 `/start`）。
    2.  打开你的浏览器，访问以下 URL (将 `<YOUR_BOT_TOKEN>` 替换为你的真实 Bot Token):
        ```
        https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates
        ```
    3.  你会看到一个 JSON 响应。在 `result` 数组中，找到 `message.chat.id` 字段。这个数字就是你的私聊 Chat ID。

*   **群组**:
    1.  将你的 Bot 添加到目标群组中。
    2.  在群组中发送一条消息，并 **@** 你的 Bot (例如 `@my_test_bot hello`)。
    3.  同样，访问上面的 `getUpdates` URL，在 JSON 响应中找到对应的群组消息，其中的 `message.chat.id` 就是你的群组 Chat ID（它通常是一个负数）。

### 3. 设置环境变量

为了安全和灵活性，Bot Token 和 Chat ID 应该通过环境变量传递给程序。在运行程序之前，请在你的终端中设置它们：

```sh
export BOT_TOKEN="YOUR_TELEGRAM_BOT_TOKEN"
export CHAT_ID="YOUR_TELEGRAM_CHAT_ID" # 例如: export CHAT_ID="123456789"
```

**注意**: 在 Windows 上，你可以使用 `set` 命令，或者通过系统设置来添加环境变量。

## 如何运行示例

1.  **设置环境变量**: 按照上述步骤设置 `BOT_TOKEN` 和 `CHAT_ID` 环境变量。

2.  **安装依赖**: 在项目目录下，运行 `go mod tidy` 来下载所需的库。

    ```sh
    go mod tidy
    ```

3.  **运行程序**: 直接运行 `main.go` 文件。

    ```sh
    go run main.go
    ```

程序将连接到 Telegram API，并向你指定的聊天发送一系列模拟的通知消息。

---

## 核心功能与示例详解

### 1. 从环境变量加载配置

*   `loadConfig()` 函数负责从 `BOT_TOKEN` 和 `CHAT_ID` 环境变量中读取 Bot Token 和 Chat ID。这是一种安全且灵活的配置管理方式，避免了将敏感信息硬编码到代码中。
*   如果环境变量未设置或 `CHAT_ID` 格式不正确，程序将报错并退出，确保 Bot 能够正确初始化。

### 2. 优雅的消息模板

为了提供一致且易读的通知，我们定义了以下辅助函数，它们都使用 **MarkdownV2** 格式化文本，并包含直观的表情符号：

*   **`sendSuccess(bot, chatID, title, details)`**: 发送成功通知 (✅)。
*   **`sendError(bot, chatID, title, details)`**: 发送错误通知 (❌)。
*   **`sendWarning(bot, chatID, title, details)`**: 发送警告通知 (⚠️)。

这些函数使得发送格式统一、重点突出、带有图标的消息变得非常简单。

### 3. 真实世界场景演示

*   **`balanceAlert()` - 余额报警**: 模拟检查账户余额。当余额低于预设阈值时，发送一条 `CRITICAL: Low Balance Alert` 的错误通知。这适用于监控支付账户、云服务余额等。

*   **`serverMonitoringAlert()` - 服务器预警**: 模拟检查服务器的 CPU 和内存使用率。当任一指标超过阈值时，发送一条 `High Resource Usage` 的警告通知。这对于运维监控和及时发现服务器性能问题非常有用。

*   **`logUserAction()` - 操作日志触发**: 模拟记录用户执行的敏感操作。例如，当管理员删除了生产数据库时，发送一条 `Audit Log: Critical Action Performed` 的成功通知到指定的审计频道。这有助于追踪关键操作和安全审计。

### 4. 其他消息类型 (在 `main.go` 中注释掉，可自行启用)

*   **内联键盘 (Inline Keyboard)**: 在消息下方附加可点击的按钮，用于交互式操作。
*   **编辑消息**: 演示如何动态修改已发送的消息，例如将“处理中”更新为“已完成”。
*   **发送文档/文件**: 演示如何发送通用文件，如日志文件或报告。

这些功能在 `main.go` 中以单独的函数形式存在，你可以根据需要取消注释并运行它们。
