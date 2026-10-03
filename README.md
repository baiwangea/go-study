# Go Study: From Syntax to a Web3 Trading Bot

<p align="center">
  <a href="README.zh.md">中文</a> |
  <strong>English</strong>
</p>

A **level-based** Go practice repository: directories are numbered in learning order (`01-` ~ `17-`),
each one a standalone Go module. Written for someone who already knows other languages and wants to
build their own **Web3 trading bot**.

> **👉 The learning path lives in one file: [LEARNING_PATH.md](./LEARNING_PATH.md)**
> It covers: the nine stages and their level counts, which levels you can skip, how to spend 1 hour a day,
> and the 18-level breakdown of the target project.

Don't know where to start? Run these two:

```sh
make path                               # print the stage roadmap
cd 01-go-fundamentals && go run . list  # list the levels of the current stage
```

## Directory = Order

| Directory | Stage | Content | Levels |
| :--- | :--- | :--- | :--- |
| [`01-go-fundamentals`](./01-go-fundamentals/) | L1 | functions, packages, interfaces (10 levels, one file each) | ✅ |
| [`02-go-pointers`](./02-go-pointers/) | L1 | pointers, value vs reference semantics | ⬜ |
| [`03-go-concurrency`](./03-go-concurrency/) | L2 | goroutine, channel, select, Mutex, WaitGroup | ⬜ |
| [`04-go-data-structures`](./04-go-data-structures/) | L3 | slice / map / struct / Set | ⬜ |
| [`05-go-algorithms`](./05-go-algorithms/) | L3 | sorting and basic algorithms | ⬜ |
| [`06-stdlib-http`](./06-stdlib-http/) | L4 | `net/http` client / server / advanced | ⬜ |
| [`07-stdlib-logger`](./07-stdlib-logger/) | L4 | the standard `log` package | ⬜ |
| [`08-go-mysql-example`](./08-go-mysql-example/) | L5 | MySQL via `database/sql` | ⬜ |
| [`09-go-redis-example`](./09-go-redis-example/) | L5 | `go-redis`: cache / dedupe / rate limit | ⬜ |
| [`10-go-mongodb-example`](./10-go-mongodb-example/) | L5 | MongoDB official driver CRUD | ⬜ |
| [`11-gin-framework-example`](./11-gin-framework-example/) | L6 | layered Gin scaffold (config/middleware/GORM/Redis/NSQ/logging) | ⬜ |
| [`12-echo-framework-example`](./12-echo-framework-example/) | L6 | Echo with a `cmd`/`internal` layout | ⬜ |
| [`13-go-jwt-example`](./13-go-jwt-example/) | L6 | JWT issuing and verification | ⬜ |
| [`14-asynq`](./14-asynq/) | L7 | Redis task queue: priority, delay, uniqueness, cron, retry, idempotency, graceful shutdown | ⬜ |
| [`15-go-zero-example`](./15-go-zero-example/) | L7 | go-zero microservices: API + RPC + `goctl` codegen | ⬜ |
| [`16-go-telegram-bot-example`](./16-go-telegram-bot-example/) | L8 | Telegram notifications & reports (the bot's alert channel) | ⬜ |
| [`17-go-web3-example`](./17-go-web3-example/) | L8 | Ethereum RPC queries (starting point for on-chain reads) | ⬜ |

Planned: `18-go-websocket-market` (feeds), `19-trading-bot-core` (strategy/risk/execution), `20-deploy-ops` (shipping it).

## How to Run

- **Go version**: `go 1.25+`; the root [`go.work`](./go.work) aggregates every module and [`Makefile`](./Makefile) loops over them.
- **Level navigation** (the intended way to study):

  ```sh
  go run . list     # level catalog (ID · prerequisites · goal)
  go run . 3        # run level 3 only
  go run . 3-5      # run a group
  go run .          # run everything (revision)
  ```

- Other common commands:

  ```sh
  make path                              # print the learning roadmap
  make build / make vet                  # compile / vet every module (binaries go to bin/)
  make run M=./14-asynq                  # module whose entry is at its root
  make run M=./11-gin-framework-example/src P=./cmd   # entry lives in a subdirectory
  ```

## Required Services & Environment Variables

| Service | Address | Used by |
| :--- | :--- | :--- |
| Redis | `127.0.0.1:6379` | `09-`, `14-`, `11-` |
| MySQL | `127.0.0.1:3306` (`08-`) / `127.0.0.1:3308` (`11-` dev config) | `08-`, `11-` |
| MongoDB | `mongodb://localhost:27017` | `10-` |
| etcd | `127.0.0.1:2379` | `15-go-zero-example` (RPC registration; the API side connects directly) |
| nsqd / nsqlookupd | `127.0.0.1:4150` / `127.0.0.1:4161` | `11-` |
| HTTP ports | `8080~8085`, `8888` (greet), `8889` (user-api), `1323` (echo) | respective modules |

| Variable | Purpose |
| :--- | :--- |
| `TG_KEYS1` / `CHAT_ID1` | Telegram bot token and chat ID (`16-`) |
| `ETH_RPC_URL` | Ethereum RPC endpoint (`17-`) |

## Conventions

- **Directory numbers are the learning order.** When adding a module, take the next number and register it
  in the stage table of `LEARNING_PATH.md`.
- One directory = one standalone Go module. **Module paths keep their original names** (e.g. `go-study/Asynq`,
  `go-study/go-fundamentals`) and do not follow the directory rename — that keeps import paths and history stable.
- Levels follow the spec in `LEARNING_PATH.md` section five: one concept per file, 30~50 lines, each level
  carries Goal / Observe / Questions / Pass, `levels.go` registers them, and `main.go` is navigation only.
- Built binaries go to the root `bin/` (git-ignored); commit `go.work`, not `go.work.sum`.
