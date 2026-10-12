# Async Library API Fix Plan

Based on investigation of `github.com/chichengyu/async v0.2.0`, all 8 files use fabricated API signatures that don't match the actual library API.

## Actual API Signatures (from `go doc`)

### 1. Slice Operations (Map/ForEach)
```go
// Correct API
async.Slice(ctx context.Context, items []T) *SliceBuilder[T, T]
// Then chain methods - need to discover actual builder methods
```

### 2. Group
```go
// Correct API
async.Group[T any]() *GroupBuilder[T]
// NOT: async.NewGroup[T](ctx)
```

### 3. Retry
```go
// Correct API
async.Retry[T any](ctx context.Context) *RetryChain[T]
// Returns a chain builder, not accepting function + options directly
```

### 4. RateLimiter
```go
// Correct API
async.Ratelimit(ctx context.Context) *RatelimitBuilder
// Then chain: .RateLimiter().Rate(n).Per(duration).Build()
```

### 5. Pipeline
```go
// Correct API
async.Pipeline[T any](items []T) *PipelineBuilder[T]
// Then need to discover builder methods
```

### 6. BoundedRunner
```go
// Correct API
async.NewBoundedRunnerBuilder() *BoundedRunnerBuilder
// NOT: async.NewBoundedRunner(n)
```

## File-by-File Fix Plan

### File 1: `01_map_basic.go`

**Problems:**
1. Line 53: `async.Map(ctx, numbers, async.IO(), func...)` - `Map` is a type, not a function
2. Line 90: Same issue with `Map`
3. Line 117: Same issue
4. Line 133: Same issue

**Fix:**
Replace all `async.Map(ctx, items, concurrency, fn)` with:
```go
results := async.Slice(ctx, items).
    DefaultWorker().  // or .Worker(n) for specific concurrency
    Map(fn)
```

The `Map` method returns `[]core.Result[R]` which matches the current usage pattern.

**Import needed:** May need `github.com/chichengyu/async/core` for `Result` type.

**Files to modify:** `01_map_basic.go`

**Verify:** `cd /Users/cluo/code/go/src/go-study/08-async-concurrent && go build -tags level 01_map_basic.go`

---

### File 2: `02_foreach.go`

**Problems:**
1. Line 53: `async.ForEach(ctx, users, async.IO(), func...)` - Similar to Map issue
2. Line 72: Same pattern
3. Line 85: Same pattern  
4. Line 97: `async.Map()` used incorrectly

**Fix:**
Replace all `async.ForEach(ctx, items, concurrency, fn)` with:
```go
result := async.Slice(ctx, items).
    DefaultWorker().
    ForEach(fn)
```

**Files to modify:** `02_foreach.go`

**Verify:** `cd /Users/cluo/code/go/src/go-study/08-async-concurrent && go build -tags level 02_foreach.go`

---

### File 3: `03_group.go`

**Problems:**
1. Line 51: `async.NewGroup[any](ctx).Add(...).Wait()` - Wrong constructor
2. Line 79: Same issue
3. Line 110: Same issue with `NewGroup` and `.WaitAndGetResults()`

**Fix:**
Replace `async.NewGroup[T](ctx)` with:
```go
group := async.Group[T]().
    Context(ctx).
    Go(func(ctx context.Context) (T, error) { ... })
```

Then call `.Wait(ctx)` (note: ctx passed to Wait, not constructor).

For results, the pattern should be:
```go
results, err := group.Wait(ctx)
```

**Files to modify:** `03_group.go`

**Verify:** `cd /Users/cluo/code/go/src/go-study/08-async-concurrent && go build -tags level 03_group.go`

---

### File 4: `04_retry.go`

**Problems:**
1. Line 55: `async.Retry(ctx, func, options...)` - Wrong signature
2. Lines 62, 87-89: Undefined option functions like `async.WithMaxRetries`
3. Line 78: Same pattern repeated
4. Line 98: Same pattern
5. Line 120: Same pattern

**Fix:**
Replace all retry calls with chain pattern:
```go
result, err := async.Retry[string](ctx).
    Exponential().
    MaxRetries(3).
    Backoff(100*time.Millisecond, 5*time.Second).
    Execute(func(ctx context.Context) (string, error) {
        // retry logic
    })
```

The chain methods are:
- `.Exponential()` or `.Linear()` - sets backoff strategy
- `.MaxRetries(n)` - sets max retry count
- `.Backoff(min, max time.Duration)` - sets backoff range
- `.Execute(fn)` - executes with retry

**Files to modify:** `04_retry.go`

**Verify:** `cd /Users/cluo/code/go/src/go-study/08-async-concurrent && go build -tags level 04_retry.go`

---

### File 5: `05_rate_limiter.go`

**Problems:**
1. Line 40: `async.NewTokenBucketRateLimiter(rate, period, capacity)` - Wrong constructor
2. Line 80: Same issue
3. Line 101: Same pattern
4. Line 127: `async.NewSlidingWindowRateLimiter(limit, window)` - Wrong constructor
5. Line 151: `async.NewAdaptiveRateLimiter().WithX().Build()` - Partially correct but needs verification

**Fix:**
Replace token bucket creation:
```go
limiter := async.Ratelimit(ctx).
    TokenBucket().
    Rate(5).
    Capacity(10).
    Build()
```

Replace sliding window:
```go
limiter := async.Ratelimit(ctx).
    SlidingWindow().
    Limit(3).
    Window(time.Second).
    Build()
```

Replace adaptive:
```go
limiter := async.Ratelimit(ctx).
    Adaptive().
    MinWorker(1).
    MaxWorker(100).
    Build()
```

Then use `limiter.Wait(ctx)` or `limiter.Allow()` depending on the pattern.

**Files to modify:** `05_rate_limiter.go`

**Verify:** `cd /Users/cluo/code/go/src/go-study/08-async-concurrent && go build -tags level 05_rate_limiter.go`

---

### File 6: `06_pipeline.go`

**Problems:**
1. Line 61: `async.NewPipeline[RawData]().AddStage(...).Run(ctx, data)` - Wrong constructor pattern
2. Line 93: `async.NewPipeline[int]().WithConcurrency(3)` - Wrong pattern
3. Line 113: `async.NewPipeline[int]().WithAutoScale(...)` - Wrong pattern
4. Line 139: Same pattern

**Fix:**
Replace pipeline creation. Based on `go doc`, the correct pattern is:
```go
pipeline := async.Pipeline(rawData).
    Context(ctx).
    Stage("name", concurrency, func(ctx context.Context, item T) (R, error) {
        // process
    }).
    Run()
```

Or use the serial/parallel chain methods. Need to verify exact API from documentation.

**Files to modify:** `06_pipeline.go`

**Verify:** `cd /Users/cluo/code/go/src/go-study/08-async-concurrent && go build -tags level 06_pipeline.go`

---

### File 7: `07_bounded_runner.go`

**Problems:**
1. Line 66: `async.NewBoundedRunner(10)` - Wrong signature
2. Line 117: `async.NewBoundedRunner(100).WithQueueSize(1000).WithBackpressure(true)` - Wrong pattern
3. Line 145: Same issue
4. Line 172: Same issue
5. Line 203: Same issue

**Fix:**
Replace all bounded runner creation:
```go
runner := async.NewBoundedRunnerBuilder().
    MaxWorkers(10).
    QueueSize(1000).
    Backpressure(true).
    Build()
```

Then use `runner.Submit(func() error { ... })` and `runner.Wait()`.

**Files to modify:** `07_bounded_runner.go`

**Verify:** `cd /Users/cluo/code/go/src/go-study/08-async-concurrent && go build -tags level 07_bounded_runner.go`

---

### File 8: `08_real_crawler.go`

**Problems:**
Multiple fabricated APIs used throughout:
1. Line 162: `async.NewTokenBucketRateLimiter(qps, qps*2)` - Wrong signature
2. Line 176: `async.Retry(ctx, func)` with options - Wrong pattern
3. Line 189: `async.StopRetrying()` - Need to verify if this exists
4. Line 212: `async.Map(ctx, pages, async.CPU(), fn)` - Wrong pattern
5. Line 250: `async.ForEach(ctx, products, async.IO(), fn)` - Wrong pattern
6. Line 270: `async.NewPipeline[URL, HTMLPage, []Product, error]()` - Complex wrong pattern
7. Line 290: `async.NewBoundedRunner(100)` - Wrong pattern
8. Line 292: `async.NewGroup[CrawlResult]()` - Wrong pattern

**Fix:**
This file needs comprehensive rewrite using correct API patterns from fixes 1-7:

1. Replace `NewTokenBucketRateLimiter` with `Ratelimit(ctx).TokenBucket().Rate(...).Capacity(...).Build()`
2. Replace `Retry(ctx, func, opts...)` with `Retry[T](ctx).Exponential().MaxRetries(...).Execute(fn)`
3. Replace `Map(ctx, items, concurrency, fn)` with `Slice(ctx, items).DefaultWorker().Map(fn)`
4. Replace `ForEach(ctx, items, concurrency, fn)` with `Slice(ctx, items).DefaultWorker().ForEach(fn)`
5. Replace `NewPipeline` with correct `Pipeline(items).Context(ctx).Stage(...).Run()`
6. Replace `NewBoundedRunner(n)` with `NewBoundedRunnerBuilder().MaxWorkers(n).Build()`
7. Replace `NewGroup[T]()` with `Group[T]().Context(ctx).Go(...).Wait(ctx)`

**Files to modify:** `08_real_crawler.go`

**Verify:** `cd /Users/cluo/code/go/src/go-study/08-async-concurrent && go build -tags level 08_real_crawler.go`

---

## Summary of Required Changes

### API Pattern Corrections

1. **Map/ForEach**: `async.Map(ctx, items, n, fn)` → `async.Slice(ctx, items).DefaultWorker().Map(fn)`
2. **Group**: `async.NewGroup[T](ctx)` → `async.Group[T]().Context(ctx).Go(fn).Wait(ctx)`
3. **Retry**: `async.Retry(ctx, fn, opts...)` → `async.Retry[T](ctx).Exponential().MaxRetries(n).Execute(fn)`
4. **RateLimiter**: `async.NewTokenBucketRateLimiter(...)` → `async.Ratelimit(ctx).TokenBucket().Rate(n).Capacity(n).Build()`
5. **Pipeline**: `async.NewPipeline[T]()` → `async.Pipeline(items).Context(ctx).Stage(...)`
6. **BoundedRunner**: `async.NewBoundedRunner(n)` → `async.NewBoundedRunnerBuilder().MaxWorkers(n).Build()`

### Import Requirements

All files already import `"github.com/chichengyu/async"`. May need to add:
- `"github.com/chichengyu/async/core"` for `Result[T]` type if not exported from main package

### Verification Strategy

After each file fix:
1. Run `go build -tags level <filename>` to check compilation
2. If successful, run the program to verify runtime behavior
3. Check that output matches the educational intent of each example

### Implementation Order

1. Fix `01_map_basic.go` (simplest, establishes Slice pattern)
2. Fix `02_foreach.go` (similar to #1)
3. Fix `03_group.go` (establishes Group pattern)
4. Fix `04_retry.go` (establishes Retry chain pattern)
5. Fix `05_rate_limiter.go` (establishes RateLimiter pattern)
6. Fix `06_pipeline.go` (medium complexity)
7. Fix `07_bounded_runner.go` (establishes BoundedRunner pattern)
8. Fix `08_real_crawler.go` (most complex, uses all patterns)

### Notes

- All files use `//go:build level` tag, so normal `go build` won't compile them
- Must use `go build -tags level` for building
- The library uses fluent/chain API design throughout
- Context is passed at different points: sometimes to constructor, sometimes to Wait/Execute
- Need to verify exact method names by testing compilation after each fix
