# Go Study: Syntax → Concurrency → GoFrame, on the Way to a Web3 Trading Bot

<p align="center">
  <a href="README.zh.md">中文</a> |
  <strong>English</strong>
</p>

A Go practice repo in **direct learning mode**: **one file = one level = a complete
`package main` with `func main()`**. No navigation framework — just `go run` that file.
Directory names *are* the learning order. Written for someone with years of other-language
experience whose goal is to build their own **Web3 trading bot**.

**88 levels**, from language fundamentals through testing and engineering practice to a full
Web3 trading-bot pipeline.

```sh
cd 02-go-pointers
go run 11_value_vs_pointer.go     # run one level, that's it
```

Convenience entry points from the repo root (`M=` is optional — it searches every module;
`F=` accepts a prefix and the `.go` suffix is optional):

```sh
make level F=11_value_vs_pointer.go     # run this level (file level → go run)
make level F=11_value_vs_pointer        # .go optional
make level F=15_nil                     # prefix match (runs if unique, else lists candidates)
make level F=01_table_driven            # directory level (06-go-testing) → go test -v
make levels M=./02-go-pointers          # list every level in a module
```

## Seven Modules

| Stage | Directory | Content | Levels |
| :--- | :--- | :--- | :--- |
| **L1 Foundations** | [`01-go-fundamentals`](./01-go-fundamentals/) | interfaces, error chains, custom errors, panic/recover, string/rune, generics, embedding | ✅ 12 |
| | [`02-go-pointers`](./02-go-pointers/) | value copies, receivers, slice/map sharing, nil traps, escape analysis, PHP `implements` compared | ✅ 7 |
| **L2 Concurrency** | [`03-go-concurrency`](./03-go-concurrency/) | goroutine, channel, select, context, Mutex/atomic, worker pool, sync.Once/RWMutex/Pool/Map, pipeline, errgroup, rate limiting, singleflight, semaphore, atomic.Value | ✅ 23 |
| **L3 GoFrame mainline** | [`04-goframe`](./04-goframe/) | routing / binding+validation / middleware / config / logging / error codes / g.DB / ORM+tx / Redis / gcron / async queue / JWT / TG / on-chain reads / layering / cross-compile | ✅ 16 |
| **Async helpers** | [`08-async-concurrent`](./08-async-concurrent/) | `async.Map` / ForEach / Group / Retry / RateLimiter / Pipeline / BoundedRunner / crawler | ✅ 8 |
| **Cross-cutting · Engineering** | [`06-go-testing`](./06-go-testing/) | table-driven tests, subtests, httptest, benchmarks, pprof, fuzzing, interface stubs, coverage & CI | ✅ 8 |
| **Cross-cutting · Web3 & trading** | [`07-web3-trading`](./07-web3-trading/) | **decimal money precision, keys/addresses/keystore, HD wallets, ABI encode/decode, event parsing, EIP-1559 signing, order state machine, WS reconnect + gap fill, Prometheus, ethclient, nonce, receipts, ERC20 approve, slippage/deadline** | ✅ 16 |
| **L4 Capstone** | [`05-trading-bot`](./05-trading-bot/) | market (REST polling + WS push) → signal → risk → order (timeout/retry/idempotent) → **order state machine** → persist → notify; **all money in decimal, table-driven tests, backtest mode** | ✅ verified with live data |

MySQL, Redis, logging, JWT, Telegram, Web3 and task queues **no longer live in their own directories** —
they are levels inside `04-goframe` (see the mapping table in [`04-goframe/README.md`](./04-goframe/README.md)).
`05-trading-bot` wires those capabilities into one real pipeline and is the only module shaped like a
normal multi-file project.

The two cross-cutting modules sort last by number, but **not by study order**:

- `06-go-testing` — study it **after L2 (`03-go-concurrency`), before L3**;
- `07-web3-trading` — study it **after L3 (`04-goframe`), before L4**; it is the parts-level
  prerequisite for `05-trading-bot`.

Neither has a hard dependency on the others, so you can slot them in whenever.

## How to Study

1. Enter a stage directory, look at its `README.md` level table, pick a level.
2. `go run <that-file>.go`, then read the **core concept / what you'll see** lines in its header comment.
3. Open the file and **change code to answer the exercises** (producing a compile error is a good sign).
4. Close the file and explain the level in one sentence (can't do it = not passed, retry tomorrow).
5. Write 3 lines of notes: what it solved / how it differs from PHP/Java / which part of the bot needs it.

**A level only counts as passed once you have actually edited the code for the exercises.**

## Common Commands

```sh
make build / make vet   # compile / vet every module
make test               # run tests
make race               # run tests with the race detector (5-10x slower, separate CI step)
make levels M=./02-go-pointers   # list levels
make level  F=11_value_vs_pointer.go   # run one level (M is optional)
make run M=./04-goframe # run one module (P=./cmd for a subdir entry)
make clean              # remove bin/
```

The root [`go.work`](./go.work) aggregates the modules and the [`Makefile`](./Makefile) loops over them
(a multi-module repo cannot be covered by a single `go build ./...`). CI lives in
[`.github/workflows/ci.yml`](./.github/workflows/ci.yml) and runs build / vet / test / race / gofmt on every push.

## Level Shapes (two of them — don't mix them up)

**Shape A · file levels** (`01` / `02` / `03` / `04` / `07` / `08`)
- Name `NN_topic.go`; **one file = one level = one complete `func main()`**, run with `go run that-file.go`
- The **first line must be `//go:build level`** — never `ignore`:
  - several `func main()` in one directory would otherwise fail with `main redeclared`;
  - `go run <file>.go` names files explicitly and therefore **bypasses build constraints**, so the file still runs alone;
  - `go build ./...` / `go vet ./...` skip these files, so the conflict disappears;
  - using `ignore` makes `go mod tidy` **wipe the require list** (04-goframe would lose goframe/mysql/redis).
- The header comment *is* the level card: **core concept / what you'll see / exercises / pass criteria /
  prerequisites+tags**. The old `level.Level` runtime, `main.go` and `levels.go` are gone.
- Side effect: if a module has no ordinary buildable package, `go run .` reports
  `build constraints exclude all Go files`. That is by design — run a single file instead.
  `make build` / `make vet` recognise this case explicitly and do not report a false failure.
- Directory renames do **not** change Go module paths (`go-study/go-concurrency` lives in 03-go-concurrency,
  `go-study/goframe` lives in 04-goframe), so historical imports keep working.
- Exceptions: `04-goframe/internal/` and `04-goframe/demoserver/` are **ordinary buildable packages** and carry no tag.

**Shape B · directory levels** (`06-go-testing`, 8 levels)
- **One directory = one level = one complete package** (implementation file + `_test.go`), run with
  `go test -v ./NN_name/`
- Why not shape A: **a test is inherently "a package plus a test file"** — the test file must live in the
  same package as the code under test (to reach unexported symbols), and a single-file `func main()`
  cannot hold both. Besides, this level's "run" command *is* `go test`.
- The level card lives in the package doc comment of the implementation file (visible via `go doc`)
- `06-go-testing/06_fuzz/testdata/fuzz/` holds the minimised reproducers the fuzzer wrote out
  automatically — **commit them**

`make levels` / `make level` **auto-detect both shapes**: file levels go through `go run`,
directory levels through `go test -v`.

## Required Services & Environment

| Service | Address | Needed from |
| :--- | :--- | :--- |
| MySQL | `127.0.0.1:3308`, db `go_study` | L3-07, L3-08 (supplied via the `BOT_DB_LINK` env var; no password in git) |
| Redis | `127.0.0.1:6379` (levels use db 9) | L3-09 |
| Ethereum RPC | `ETH_RPC_URL` | L3-14, `07-web3-trading` W-12 (falls back to public nodes; skips gracefully when unreachable) |
| Telegram | `TG_BOT_TOKEN` | L3-13, real sends in `05-trading-bot` |
| Bitget market data | `https://api.bitget.com` (REST) / `wss://ws.bitget.com/v2/ws/public` (push), both public | `05-trading-bot -feed=bitget` / `-feed=bitget-ws` |

All of L1 and L2, the service-free L3 levels (only L3-07/08/09 need MySQL/Redis, L3-13 needs a TG token),
and the `05-trading-bot` mock mode run **without any external service** — clone and `go run`.
When a dependency is missing the level skips explicitly and tells you how to set it up; it never panics
or hangs. Live quotes, MySQL persistence and cross-process Redis idempotency have all been verified.
