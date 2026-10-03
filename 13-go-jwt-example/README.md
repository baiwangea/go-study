# Go 实现 JWT 认证示例

本项目使用 `github.com/golang-jwt/jwt/v4` 库，演示了如何在 Go 应用程序中创建、签发和验证 JSON Web Tokens (JWT)。

## 什么是 JWT？

JWT (JSON Web Token) 是一个开放标准 (RFC 7519)，它定义了一种紧凑且自包含的方式，用于在各方之间安全地传输信息（作为 JSON 对象）。由于其体积小、可通过 URL、POST 参数或 HTTP 头部轻松传输，并且是自包含的（包含了所有需要验证用户的信息），因此被广泛用于 Web 应用的无状态身份验证。

一个 JWT 由三部分组成，由点 (`.`) 分隔：

1.  **Header (头部)**: 包含了 token 的类型（即 JWT）和所使用的签名算法（如 HMAC SHA256 或 RSA）。
2.  **Payload (载荷)**: 包含了 “声明 (Claims)”。声明是关于实体（通常是用户）和其他数据的陈述。有三种类型的声明：注册声明、公共声明和私有声明。
3.  **Signature (签名)**: 用于验证消息在此过程中没有被篡改。它是由编码后的头部、编码后的载荷、一个密钥（secret）以及头部中指定的算法进行签名生成的。

## 如何运行示例

1.  **安装依赖**: 在项目目录下，运行 `go mod tidy` 来下载 `jwt` 库。

    ```sh
    go mod tidy
    ```

2.  **运行程序**: 直接运行 `main.go` 文件。

    ```sh
    go run main.go
    ```

程序将模拟一个完整的流程：为一个用户生成 JWT，然后成功验证它，最后尝试验证一个被篡改的 token 和一个已过期的 token，并展示相应的错误信息。

---

## 示例详解

### `jwtKey` 和 `Claims` 结构体

*   **`jwtKey`**: 这是一个字节切片，用作签名的**密钥**。在真实的生产环境中，这个密钥必须是高度复杂的，并且必须从安全的地方（如环境变量或密钥管理服务）加载，**绝对不能硬编码在代码中**。
*   **`Claims`**: 这是一个我们自定义的结构体，用于定义 JWT 的载荷部分。我们通过嵌入 `jwt.RegisteredClaims` 来包含所有标准的、官方建议的声明（如 `ExpiresAt`, `Issuer` 等），同时我们添加了自己的私有声明 `Username`。

### `createToken()` - 创建并签发 Token

1.  **设置过期时间**: 我们为 token 设置一个明确的过期时间。这是一个非常重要的安全实践，可以防止 token 被无限期使用。
2.  **创建 Claims**: 我们创建 `Claims` 结构体的实例，填入我们的自定义数据（用户名）和标准数据（过期时间）。
3.  **`jwt.NewWithClaims(...)`**: 我们使用指定的签名算法（这里是 `HS256`）和我们创建的 `claims` 来生成一个新的 token 对象。
4.  **`token.SignedString(jwtKey)`**: 这是最关键的一步。我们使用密钥 `jwtKey` 对 token 对象进行签名，生成最终的、可以发送给客户端的、由三部分组成的 JWT 字符串。

### `validateToken()` - 解析并验证 Token

这个函数模拟了服务器端接收到客户端传来的 JWT 字符串后进行验证的过程。

1.  **`jwt.ParseWithClaims(...)`**: 这是用于解析和验证 token 的核心函数。它接收三个参数：
    *   要解析的 token 字符串。
    *   一个空的 `Claims` 结构体指针，用于存放解析出的载荷数据。
    *   一个回调函数，这是**安全验证的关键**。
2.  **回调函数 `func(token *jwt.Token) (interface{}, error)`**: 这个函数在验证 token 签名之前被调用。它的责任是：
    *   **检查签名算法**: 检查 `token.Header["alg"]` 是否是你所期望的算法（例如，`HMAC`）。这可以防止攻击者将算法改为 `none` 来绕过签名验证。
    *   **返回密钥**: 返回用于签名的 `jwtKey`。`ParseWithClaims` 会用这个密钥来计算签名，并与 token 自带的签名进行比对。如果比对失败，则整个验证过程失败。
3.  **错误处理**: `ParseWithClaims` 会返回多种错误。示例代码中演示了如何专门检查 `jwt.ErrSignatureInvalid`（签名无效错误）以及其他类型的错误（如 `token is expired` 过期错误）。
4.  **检查 `tkn.Valid`**: 即使没有返回错误，你也应该检查返回的 `tkn.Valid` 字段，以确保 token 是完全有效的。
