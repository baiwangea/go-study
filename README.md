# Go Study: Syntax → Concurrency → GoFrame, on the Way to a Web3 Trading Bot

<p align="center">
  <a href="README.zh.md">中文</a> |
  <strong>English</strong>
</p>

A level-based Go practice repository: **one concept = one small file = one level**, and the directory
names *are* the learning order. Written for someone with years of other-language experience whose goal
is to build their own **Web3 trading bot**.

> **👉 The path lives in one file: [LEARNING_PATH.md](./LEARNING_PATH.md)**
> Four stages, which levels you can skip, how to spend one hour a day, and bot capabilities mapped back to levels.

```sh
make path                              # print the stage roadmap
cd 01-go-fundamentals && go run . list # list levels of the current stage
cd 01-go-fundamentals && go run . 3    # run level 3 only
```

## Four Directories

| Directory | Stage | Content | Levels |
| :--- | :--- | :--- | :--- |
| [`01-go-fundamentals`](./01-go-fundamentals/) | L1 | functions, packages, interfaces | ✅ 10 |
| [`02-go-pointers`](./02-go-pointers/) | L1 | value copies, receivers, slice/map sharing, nil traps, escape analysis | ✅ 6 |
| [`03-go-concurrency`](./03-go-concurrency/) | L2 | goroutine, WaitGroup, channel, select, context, mutex/atomic | ✅ 13 |
| [`04-goframe`](./04-goframe/) | L3 | **GoFrame mainline**: routing / binding+validation / middleware / config / logging / error codes → MySQL, Redis, cron, JWT, Telegram, Web3, deploy | 🟡 6/16 |

MySQL, Redis, logging, JWT, Telegram, Web3 and task queues **no longer live in their own directories** —
they are levels inside `04-goframe` (see the mapping table in [`04-goframe/README.md`](./04-goframe/README.md)).
Planned next: `05-trading-bot` (the target project).

## Common Commands

```sh
make build / make vet   # compile / vet every module (binaries land in bin/)
make test               # run tests
make clean              # remove bin/
make run M=./04-goframe # run one module
```

The root [`go.work`](./go.work) aggregates the modules and [`Makefile`](./Makefile) loops over them
(a multi-module repo cannot be covered by a single `go build ./...`).

## Level Runtime & Conventions

- File name `NN_topic.go`, one level per file, **30~50 lines**
- Each level returns a `level.Level`: `ID / Title / Tags / Pre / Goal / Observe / Questions / Check / Run`
- `level/level.go` is a dependency-free runtime copied into each module; `main.go` is navigation only
  (`list` / single level / range / all)
- Directory renames do **not** change Go module paths (`go-study/go-concurrency`, `go-study/stdlib-http`),
  so historical imports keep working

## Required Services & Environment

| Service | Address | Needed from |
| :--- | :--- | :--- |
| MySQL | `127.0.0.1:3306` | `04-goframe` L3-07 |
| Redis | `127.0.0.1:6379` | L3-09, L3-11 |
| Ethereum RPC | `ETH_RPC_URL` | L3-14 |
| Telegram | `TG_BOT_TOKEN`, `TG_CHAT_ID` | L3-13 |

Everything in L1, L2 and the first six L3 levels (L3-01 ~ L3-06) runs **without any external service** —
clone and `go run`.
