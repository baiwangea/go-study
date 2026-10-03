# Go Study: A Hands-On Learning Repository

<p align="center">
  <a href="README.zh.md">中文</a> |
  <strong>English</strong>
</p>

Welcome to **Go Study**, a curated collection of practical, hands-on examples designed to accelerate your journey into the Go programming language. Each module is a self-contained Go project (with its own `go.mod`), complete with detailed code, comments, and its own `README.md`.

## 🗺️ Learning Path

Work through the stages in order — each one builds on the previous:

| Stage | Goal | Modules |
| :--- | :--- | :--- |
| ① Language basics | Syntax, type system, pointer semantics | [`go-fundamentals`](./go-fundamentals/), [`go-pointers`](./go-pointers/), [`go-data-structures`](./go-data-structures/) |
| ② Concurrency | Goroutine orchestration and sync primitives | [`go-concurrency`](./go-concurrency/) |
| ③ Standard library | Hand-write HTTP servers and logging to understand what frameworks do | [`stdlib-http`](./stdlib-http/), [`stdlib-logger`](./stdlib-logger/) |
| ④ Algorithms | Common sorts and complexity intuition | [`go-algorithms`](./go-algorithms/) |
| ⑤ Storage & caching | SQL / NoSQL / cache in practice | [`go-mysql-example`](./go-mysql-example/), [`go-redis-example`](./go-redis-example/), [`go-mongodb-example`](./go-mongodb-example/) |
| ⑥ Web frameworks | From a single file to a layered project scaffold | [`gin-framework-example`](./gin-framework-example/), [`echo-framework-example`](./echo-framework-example/), [`go-jwt-example`](./go-jwt-example/) |
| ⑦ Messaging & async | Load leveling, decoupling, scheduling, idempotency | [`Asynq`](./Asynq/), NSQ integration (inside the Gin module) |
| ⑧ Microservices & extras | Service splitting, code generation, third-party APIs | [`go-zero-example`](./go-zero-example/), [`go-telegram-bot-example`](./go-telegram-bot-example/), [`go-web3-example`](./go-web3-example/) |

## 🚀 Modules Overview

### Core Language Features

| Module | Description |
| :--- | :--- |
| [`go-fundamentals`](./go-fundamentals/) | Covers the building blocks of Go: functions, packages, and interfaces. |
| [`go-pointers`](./go-pointers/) | A deep dive into pointers and the difference between pass-by-value and pass-by-pointer. |
| [`go-concurrency`](./go-concurrency/) | Explores Go's powerful concurrency primitives: Goroutines, Channels, Select, Mutexes, and WaitGroups. |
| [`go-data-structures`](./go-data-structures/) | Demonstrates Go's built-in data structures (slices, maps, structs) and how to implement a Set. |
| [`go-algorithms`](./go-algorithms/) | Basic algorithms: bubble sort and quick sort. |

### Standard Library in Action

| Module | Description |
| :--- | :--- |
| [`stdlib-http`](./stdlib-http/) | Practical examples of building HTTP clients and servers using the `net/http` package (including advanced usage). |
| [`stdlib-logger`](./stdlib-logger/) | A guide to using the standard `log` package for effective logging. |

### Web Frameworks & Project Structure

| Module | Description |
| :--- | :--- |
| [`gin-framework-example`](./gin-framework-example/) | A complete Gin scaffold: layered architecture, env-based config, JWT middleware, GORM, Redis, captcha, log rotation, NSQ producer/consumer, cross-compilation. |
| [`echo-framework-example`](./echo-framework-example/) | A modern Echo layout (`cmd` + `internal`) with Logger/Recover middleware. |
| [`go-zero-example`](./go-zero-example/) | Three go-zero services: `greet` API, `user` RPC, and `user-api` (demonstrates API → gRPC calls and `goctl` code generation). |
| [`go-jwt-example`](./go-jwt-example/) | Learn how to create and validate JSON Web Tokens (JWT) for stateless authentication. |

### Message Queues & Async Tasks

| Module | Description |
| :--- | :--- |
| [`Asynq`](./Asynq/) | Redis-backed task queue: priority queues, delayed tasks, unique constraints, batch enqueue, order-timeout cancellation, cron scheduling, middleware, idempotency, graceful shutdown. |
| NSQ (inside [`gin-framework-example`](./gin-framework-example/)) | `src/app/handler/nsq.go` produces messages; `src/cmd/nsq-consumer` is a standalone consumer process. |

### Data Storage

| Module | Description |
| :--- | :--- |
| [`go-mysql-example`](./go-mysql-example/) | Demonstrates safe and efficient MySQL operations using the standard `database/sql` package. |
| [`go-mongodb-example`](./go-mongodb-example/) | A guide to performing CRUD operations on MongoDB with the official driver. |
| [`go-redis-example`](./go-redis-example/) | Shows how to interact with Redis for caching and other use cases using `go-redis`. |

### Applied Examples

| Module | Description |
| :--- | :--- |
| [`go-telegram-bot-example`](./go-telegram-bot-example/) | Telegram bot messaging plus Excel report generation. |
| [`go-web3-example`](./go-web3-example/) | Queries on-chain data through an Ethereum RPC endpoint. |

## ⚙️ Requirements & How to Run

- **Go version**: `go 1.25+` recommended (module `go.mod` files declare 1.18 ~ 1.25.11).
- **Workspace**: the root [`go.work`](./go.work) lists every module, so cross-module builds work from the repo root, e.g.
  `go build ./go-zero-example/user-api`. Commit `go.work`; `go.work.sum` is git-ignored.
- **Unified entry point**: the root [`Makefile`](./Makefile) loops over modules (a multi-module repo cannot be covered by a single `go build ./...`).

```sh
make list                 # list all modules
make build                # compile every module
make vet                  # vet every module
make test                 # run tests of every module
make tidy                 # go mod tidy for each module
make run M=./Asynq                                  # module whose entry is at its root
make run M=./gin-framework-example/src P=./cmd      # entry lives in a subdirectory
```

The classic workflow still works: `cd <module> && go run .` (for the Gin module: `go run src/cmd/main.go -env=dev`).

## 🌐 Required Services & Ports

| Service | Address | Used by |
| :--- | :--- | :--- |
| Redis | `127.0.0.1:6379` | `go-redis-example`, `Asynq`, `gin-framework-example` |
| MySQL | `127.0.0.1:3306` (`go-mysql-example`) / `127.0.0.1:3308` (Gin dev config) | `go-mysql-example`, `gin-framework-example` |
| MongoDB | `mongodb://localhost:27017` | `go-mongodb-example` |
| etcd | `127.0.0.1:2379` | `go-zero-example/user` (service registration; the API side connects directly by default) |
| nsqd / nsqlookupd | `127.0.0.1:4150` / `127.0.0.1:4161` | `gin-framework-example` (NSQ produce/consume) |
| HTTP listeners | `8080`~`8085` (stdlib-http, Gin, user.rpc); `8888` (greet-api); `8889` (user-api); `1323` (Echo) | respective modules |

> The `stdlib-http/client` and `stdlib-logger` examples expect their companion servers to be running first.

## 🔑 Environment Variables

Modules that need secrets read them from the environment — never hard-code them:

| Variable | Purpose | Module |
| :--- | :--- | :--- |
| `TG_KEYS1` | Telegram bot token | `go-telegram-bot-example` |
| `CHAT_ID1` | Telegram chat ID | `go-telegram-bot-example` |
| `ETH_RPC_URL` | Ethereum RPC endpoint | `go-web3-example` |

```sh
export TG_KEYS1="<your bot token>"
export CHAT_ID1="<your chat id>"
export ETH_RPC_URL="https://your-eth-rpc-endpoint"
```

## 💡 Repository Conventions

- One directory = one independent Go module, with module paths named `go-study/<dir>` (`gin-framework-example/src` and `user-api` are legacy exceptions).
- Module docs are primarily in Chinese; the Gin module and the root docs are bilingual.
- When adding a module, update `go.work` and the module tables in `README.md` / `README.zh.md`.
- Built binaries go into the root `bin/` directory (git-ignored).

Happy Coding! ✨
