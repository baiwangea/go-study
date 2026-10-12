//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 03-go-concurrency && go run synclib/01_once.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

/*
═══════════════════════════════════════════════════════════════════════
L2-14：sync.Once —— 只执行一次
═══════════════════════════════════════════════════════════════════════

【核心概念】
    sync.Once 保证「无论多少个 goroutine 并发调用，函数体只执行一次」。
    典型用途是懒加载单例：配置、数据库连接、日志器 —— 谁先到谁初始化，
    其余的人阻塞等待，最后所有人拿到同一个对象。

【你会看到】
    ✓ 100 个 goroutine 抢着初始化，newDBConn 仍然只执行 1 次
    ✓ 后来者不是「直接返回」，而是被 Do 阻塞到初始化完成
    ✓ Do 里 panic 之后，Once 仍被视为「已完成」，函数体不会再执行

【思考题】（动手改代码验证，别空想）
    1. 把示例 1 的 Do 换成手写的 `if !done { ... }`（无锁标志位），
       执行次数还会是 1 吗？加 -race 看看官方报告怎么说的
    2. 示例 2 里把 Do 里的 100ms 改成 1 秒，6 个调用者的返回时刻
       会一起往后推吗？为什么？
    3. 在 once.Do 的函数体里再调用「同一个」once.Do 会怎样？
       （提示：Once 内部有互斥量 → 死锁）
    4. init() 的执行时机是什么？为什么「按需初始化 + 可能失败」的场景
       必须用 sync.Once 而不是 init()？

【通关标准】
    能写出并发安全的懒加载单例，并说清 Once 的「阻塞语义」和
    「panic 之后不再执行」

【前置】L2-13              【标签】sync.Once · 单例 · 懒加载
═══════════════════════════════════════════════════════════════════════
*/

// ── 辅助类型与函数 ───────────────────────────────────────────────────────

// dbConn 模拟「昂贵的一次性资源」：真实项目里对应数据库连接池、
// 配置中心客户端、全局日志器这类「全局唯一、初始化很慢」的对象。
type dbConn struct {
	dsn string
}

// newDBConn 模拟一次 100ms 的慢初始化（建连、握手、鉴权）。
func newDBConn(dsn string) *dbConn {
	time.Sleep(100 * time.Millisecond)
	return &dbConn{dsn: dsn}
}

var (
	connOnce sync.Once // 只保证下面这段初始化逻辑跑一次
	conn     *dbConn   // 懒加载出来的单例
	connInit int64     // 初始化函数体真实执行次数，用来验证「只执行一次」
)

// getDBConn 是本关的核心：并发安全 + 懒加载的单例。
// 任意多个 goroutine 同时调用，newDBConn 只会被执行一次；
// 其余的会阻塞在 Do 上，等第一个初始化完成后拿到同一个 *dbConn。
func getDBConn(dsn string) *dbConn {
	connOnce.Do(func() {
		atomic.AddInt64(&connInit, 1)
		conn = newDBConn(dsn)
	})
	return conn
}

var (
	cfgOnce sync.Once // 演示「懒加载 + 缓存错误」
	cfgVal  string
	cfgErr  error
)

// loadConfig 演示「懒加载 + 缓存错误」的写法：加载只在第一次调用时发生，
// 结果（包括 error）一起被缓存。读环境变量是为了让失败分支可以稳定复现。
// 注意：错误也会被永久缓存 —— 想重试就得换一个新的 sync.Once。
func loadConfig() (string, error) {
	cfgOnce.Do(func() {
		v := os.Getenv("GO_STUDY_CONFIG")
		if v == "" {
			cfgErr = fmt.Errorf("环境变量 GO_STUDY_CONFIG 未设置（模拟配置中心不可达）")
			return
		}
		cfgVal = v
	})
	return cfgVal, cfgErr
}

func main() {
	fmt.Println("\n=== L2-14 sync.Once ===")

	// ── 示例 1：100 个 goroutine 抢着初始化，只跑一次 ────────────────────
	fmt.Println("\n【示例 1】100 个 goroutine 同时抢初始化，函数体只执行 1 次")
	const workers = 100
	got := make([]*dbConn, workers)
	start := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			got[i] = getDBConn("mysql://trader@127.0.0.1:3306/orders")
		}(i)
	}
	wg.Wait()

	samePointer := true
	for _, c := range got {
		if c != got[0] {
			samePointer = false
			break
		}
	}
	fmt.Printf("  newDBConn 执行次数 = %d（期望 1）\n", atomic.LoadInt64(&connInit))
	fmt.Printf("  100 个 goroutine 拿到同一个指针 = %v\n", samePointer)
	fmt.Printf("  总耗时 %v（≈ 一次初始化的 100ms，说明大家是「一起等」，而不是各初始化一次）\n",
		time.Since(start).Round(time.Millisecond))

	// ── 示例 2：Do 的阻塞语义 ───────────────────────────────────────────
	fmt.Println("\n【示例 2】Do 的阻塞语义：后来者会一直等到初始化结束")
	var slowOnce sync.Once
	const callers = 6
	returnedAt := make([]time.Duration, callers)
	start = time.Now()
	var wg2 sync.WaitGroup
	for i := 0; i < callers; i++ {
		wg2.Add(1)
		go func(i int) {
			defer wg2.Done()
			slowOnce.Do(func() {
				time.Sleep(100 * time.Millisecond) // 模拟慢初始化
			})
			returnedAt[i] = time.Since(start)
		}(i)
	}
	wg2.Wait()
	for i, d := range returnedAt {
		fmt.Printf("  第 %d 个调用者返回时刻 t=%v\n", i+1, d.Round(time.Millisecond))
	}
	fmt.Println("  ↑ 没有人是 t≈0 返回的：后来的调用者被 Do 挡住，直到初始化完成才放行")

	// ── 示例 3：Do 里 panic 会怎样 ──────────────────────────────────────
	fmt.Println("\n【示例 3】Do 里 panic：Once 仍然被标记为「已完成」")
	var panicOnce sync.Once
	var ran int64
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("  捕获到 panic：%v（panic 会抛给「第一个」调用者）\n", r)
			}
		}()
		panicOnce.Do(func() {
			atomic.AddInt64(&ran, 1)
			panic("初始化失败：连接被拒绝")
		})
	}()
	// 再调用一次：函数体不会再被执行，坏结果被「永久缓存」
	panicOnce.Do(func() { atomic.AddInt64(&ran, 1) })
	fmt.Printf("  函数体执行次数 = %d（第二次 Do 直接跳过）\n", atomic.LoadInt64(&ran))
	fmt.Println("  ↑ 所以 Do 里初始化失败不要只靠 panic：要先校验再赋值，或用「缓存 error」的写法")

	// ── 示例 4：init() 和 sync.Once 的取舍 ──────────────────────────────
	fmt.Println("\n【示例 4】init() vs sync.Once：懒加载 + 缓存 error")
	cfg, err := loadConfig()
	fmt.Printf("  第一次 loadConfig() → cfg=%q\n", cfg)
	fmt.Printf("                         err=%v\n", err)
	// 模拟运维把配置补上（环境变量此刻才存在）
	os.Setenv("GO_STUDY_CONFIG", "wss://stream.example.com/market")
	cfg2, err2 := loadConfig()
	fmt.Printf("  补上环境变量后再调一次 → cfg=%q\n", cfg2)
	fmt.Printf("                              err=%v\n", err2)
	fmt.Println("  ↑ 仍然是同一个错误：Once 把「失败」也永久缓存了")
	fmt.Println("  init() 是包加载时无条件执行、不能返回 error、不能延迟到真正需要时、")
	fmt.Println("  测试里也换不成假实现 —— 需要「按需 + 可能失败 + 可注入」就用 sync.Once")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n【关键总结】")
	fmt.Println("  • Once.Do 保证函数体只执行一次；并发调用者会阻塞等待，不是直接返回")
	fmt.Println("  • Do 里的 panic 会抛给第一个调用者，但 Once 仍算「已完成」→ 坏结果被缓存")
	fmt.Println("  • 缓存 error 是常用手法，但错误不会自动重试；要重试得换新的 Once")
	fmt.Println("  • Once 不能重置：需要「可重置」用 atomic.Bool + CAS 自己写")
	fmt.Println("  • 该用：全局唯一 + 初始化昂贵 + 需要懒加载（连接池、配置、日志器）")
	fmt.Println("  • 不该用：纯常量（用 init 或直接赋值）、需要按 key 多次初始化（用 map+锁）")
}
