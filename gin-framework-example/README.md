[中文](./README.zh.md)

# Gin Framework Example

This project is a complete and modern scaffold for a web application built with the Gin framework in Go. It includes a variety of best practices and features to help you get started quickly.

## Features

*   **Environment-Based Configuration:** Easily manage configurations for development, testing, and production environments using a command-line flag.
*   **Service Layer:** A dedicated service layer to separate business logic from the handlers.
*   **Middleware:** Includes a logger middleware for request and response logging, and a JWT middleware for authentication.
*   **Database Integration:** Uses GORM for database interactions and includes a database initialization module.
*   **Redis Integration:** Includes a Redis client for caching and other purposes.
*   **Captcha:** A captcha endpoint to protect against bots.
*   **Structured Logging:** A file-based logging system with log rotation.
*   **DTOs for API Responses:** Uses Data Transfer Objects (DTOs) to control the fields returned by the API.

## Directory Structure

```
gin-framework-example/
├── conf/                  # Configuration files
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

## How to Run

1.  **Install dependencies:**

    ```bash
    go mod tidy
    ```

2.  **Run the application:**

    You can specify the environment using the `-env` flag. The default environment is `dev`.

    *   **Development:**

        ```bash
        go run src/cmd/main.go
        ```

        or

        ```bash
        go run src/cmd/main.go -env=dev
        ```

    *   **Production:**

        ```bash
        go run src/cmd/main.go -env=prod
        ```

## Building for Production

To create a smaller binary for production, use the `-ldflags="-s -w"` flags to strip debugging information.

```bash
go build -ldflags="-s -w" -o gin-framework-example src/cmd/main.go
```

## Cross-Compilation

### 1. Mac to Linux/Windows (64-bit):

```bash
# To Linux
go env -w CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# To Windows
go env -w CGO_ENABLED=0 GOOS=windows GOARCH=amd64
```

### 2. Linux to Mac/Windows (64-bit):

```bash
# To Mac
go env -w CGO_ENABLED=0 GOOS=darwin GOARCH=amd64

# To Windows
go env -w CGO_ENABLED=0 GOOS=windows GOARCH=amd64
```

### 3. Windows to Mac/Linux (64-bit):

```bash
# To Mac
go env -w CGO_ENABLED=0 GOOS=darwin GOARCH=amd64

# To Linux
go env -w CGO_ENABLED=0 GOOS=linux GOARCH=amd64
```

### Build Command

```bash
go build -o <output_path> src/cmd/main.go
```

### Daily Development Workflow

*   **Mac to Linux:**

    ```bash
    go env -w CGO_ENABLED=0 GOOS=linux GOARCH=amd64
    ```

*   **Switch back to Mac:**

    ```bash
    go env -w CGO_ENABLED=1 GOOS=darwin GOARCH=amd64
    ```
