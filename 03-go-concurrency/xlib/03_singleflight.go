//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 03-go-concurrency && go run xlib/03_singleflight.go
// 依赖：golang.org/x/sync（本关需要，见 03-go-concurrency/go.mod）
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/singleflight"
)

/*
═══════════════════════════════════════════════════════════════════════
L2-21：singleflight —— 同一时刻的重复请求只执行一次
═══════════════════════════════════════════════════════════════════════

【核心概念】
    singleflight 把「同一时刻、同一个 key」的 N 个请求合并成 1 次真实调用，
    其余请求共享这一次的结果。它解决的是「缓存击穿」：热点 key 的缓存刚过期，
    大量并发请求同时打到下游（DB / 交易所），把下游压垮。

【你会看到】
    ✓ 10 个并发请求同一 key → 下游被调用 10 次（没有 singleflight 时）
    ✓ 加上 singleflight → 下游只被调用 1 次，10 个请求共享同一份结果
    ✓ 不同 key 不会被合并
    ✓ DoChan 不阻塞调用者，可以配 select 做超时，结果也不会丢
    ✓ 手写 Mutex + 缓存为什么挡不住击穿：锁是「排队」，不是「合并」

【思考题】（动手改代码验证，别空想）
    1. 示例 2 里让 10 个 goroutine「错开 150ms 陆续发起」，
       下游会被调用几次？为什么？（提示：singleflight 只合并「同时在飞」的）
    2. singleflight 不缓存结果 —— 那它和缓存是什么关系？
    3. 示例 5 里给缓存检查加「二次检查」并锁住整个 miss→fetch 窗口，
       下游调用会降到几次？代价是什么？
    4. 如果下游查询返回错误，10 个等待者会收到什么？

【通关标准】
    能画出「缓存击穿 → singleflight 合并」的时序，
    并说清 Do 与 DoChan 的取舍、以及它和「加锁 + 缓存」的本质区别

【前置】L2-19              【标签】singleflight · 缓存击穿 · 请求合并 · DoChan
═══════════════════════════════════════════════════════════════════════
*/

// priceTable 模拟交易所行情快照。
var priceTable = map[string]float64{
	"BTCUSDT": 68000,
	"ETHUSDT": 3500,
}

// calls 记录「下游真实被调用的次数」—— 本关所有演示都在盯这个数字。
var calls int64

// fetchPrice 模拟一次「慢且贵」的下游查询：向交易所 REST 拉最新价。
// 真实场景里它是一次 100ms 的网络往返，而且有严格的限频。
func fetchPrice(symbol string, delay time.Duration) (float64, error) {
	atomic.AddInt64(&calls, 1)
	time.Sleep(delay)
	p, ok := priceTable[symbol]
	if !ok {
		return 0, fmt.Errorf("未知交易对 %s", symbol)
	}
	return p, nil
}

func main() {
	fmt.Println("\n=== L2-21 singleflight ===")

	const n = 10

	// ── 示例 1：没有 singleflight 的样子 ────────────────────────────────
	fmt.Println("\n【示例 1】没有 singleflight：10 个并发请求 = 10 次下游调用")
	atomic.StoreInt64(&calls, 0)
	var wg sync.WaitGroup
	start1 := make(chan struct{}) // 发令枪：保证 10 个请求同时出发
	prices1 := make([]float64, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start1
			p, _ := fetchPrice("BTCUSDT", 100*time.Millisecond)
			prices1[i] = p
		}(i)
	}
	begin := time.Now()
	close(start1)
	wg.Wait()
	fmt.Printf("  下游被调用 %d 次，耗时 %v，10 个请求拿到同一个价格 %.0f\n",
		atomic.LoadInt64(&calls), time.Since(begin).Round(time.Millisecond), prices1[0])
	fmt.Println("  ↑ 10 个请求同时砸到交易所 —— 这就是「缓存击穿」的形态")

	// ── 示例 2：singleflight.Do 合并同一 key ────────────────────────────
	fmt.Println("\n【示例 2】singleflight.Do：同一个 key 只真正执行一次")
	atomic.StoreInt64(&calls, 0)
	var g singleflight.Group
	var sharedCount int64
	var wg2 sync.WaitGroup
	start2 := make(chan struct{})
	prices2 := make([]float64, n)
	begin = time.Now()
	for i := 0; i < n; i++ {
		wg2.Add(1)
		go func(i int) {
			defer wg2.Done()
			<-start2
			// Do(key, fn)：第一个到的执行 fn，其余的等它的结果
			v, err, shared := g.Do("BTCUSDT", func() (any, error) {
				return fetchPrice("BTCUSDT", 100*time.Millisecond)
			})
			if err != nil {
				return
			}
			prices2[i] = v.(float64)
			if shared {
				atomic.AddInt64(&sharedCount, 1)
			}
		}(i)
	}
	close(start2)
	wg2.Wait()
	fmt.Printf("  下游被调用 %d 次，耗时 %v\n",
		atomic.LoadInt64(&calls), time.Since(begin).Round(time.Millisecond))
	fmt.Printf("  10 个请求都拿到 %.0f，耗时 %v\n",
		prices2[0], time.Since(begin).Round(time.Millisecond))
	fmt.Printf("  下游被调用 %d 次；shared=true 的有 %d 个\n",
		atomic.LoadInt64(&calls), atomic.LoadInt64(&sharedCount))
	fmt.Println("  shared 的含义是「这份结果被多个调用者共享了」，所以真正执行的那一个也是 true")
	fmt.Println("  ↑ 耗时还是 ~100ms（不是 1 秒），下游只被打了一次 —— 合并成功")

	// ── 示例 3：不同 key 不会被合并 ─────────────────────────────────────
	fmt.Println("\n【示例 3】不同 key 不会被合并：2 个交易对 = 2 次下游调用")
	atomic.StoreInt64(&calls, 0)
	var g2 singleflight.Group
	var wg3 sync.WaitGroup
	start3 := make(chan struct{})
	for _, sym := range []string{"BTCUSDT", "ETHUSDT"} {
		for i := 0; i < 5; i++ { // 每个交易对 5 个并发请求
			wg3.Add(1)
			go func(sym string) {
				defer wg3.Done()
				<-start3
				g2.Do(sym, func() (any, error) { return fetchPrice(sym, 50*time.Millisecond) })
			}(sym)
		}
	}
	close(start3)
	wg3.Wait()
	fmt.Printf("  BTCUSDT ×5 + ETHUSDT ×5 并发 → 下游被调用 %d 次（每个 key 各 1 次）\n",
		atomic.LoadInt64(&calls))
	fmt.Println("  ↑ singleflight 按 key 分组：不同 key 各飞各的，互不影响")

	// ── 示例 4：DoChan 不阻塞调用者 ─────────────────────────────────────
	fmt.Println("\n【示例 4】DoChan：不阻塞调用者，可以配 select 做超时")
	var g3 singleflight.Group
	ch := g3.DoChan("BTCUSDT", func() (any, error) {
		return fetchPrice("BTCUSDT", 100*time.Millisecond)
	})
	select {
	case r := <-ch:
		fmt.Printf("  DoChan 拿到结果 %.0f，shared=%v\n", r.Val.(float64), r.Shared)
	case <-time.After(time.Second):
		fmt.Println("  超时了")
	}

	// 慢下游 + 短超时：DoChan 让调用方先走，后台请求自己跑完
	begin = time.Now()
	ch2 := g3.DoChan("SLOWUSDT", func() (any, error) {
		time.Sleep(300 * time.Millisecond) // 模拟一个特别慢的下游
		return 12345.0, nil
	})
	select {
	case r := <-ch2:
		fmt.Printf("  拿到 %v\n", r.Val)
	case <-time.After(50 * time.Millisecond):
		fmt.Printf("  50ms 就超时返回了（t=%v），但下游请求还在后台继续跑\n",
			time.Since(begin).Round(time.Millisecond))
	}
	r2 := <-ch2 // 结果放在带缓冲的 channel 里，稍后再来取也不会丢
	fmt.Printf("  稍后再从同一个 channel 取：%v（t=%v）—— DoChan 的结果不会丢\n",
		r2.Val, time.Since(begin).Round(time.Millisecond))
	fmt.Println("  ↑ Do 会一直等到有结果；DoChan 返回一个 channel，调用方自己决定等多久")

	// ── 示例 5：和「Mutex + 缓存」的区别 ────────────────────────────────
	fmt.Println("\n【示例 5】手写 Mutex + 缓存：锁只会「排队」，不会「合并」")
	atomic.StoreInt64(&calls, 0)
	var mu sync.Mutex
	cache := map[string]float64{}
	// 只在「读写缓存」时加锁 —— 这是工程里最常见的写法，也正是击穿的来源
	getCache := func(sym string) (float64, bool) {
		mu.Lock()
		defer mu.Unlock()
		v, ok := cache[sym]
		return v, ok
	}
	setCache := func(sym string, v float64) {
		mu.Lock()
		defer mu.Unlock()
		cache[sym] = v
	}
	var wg4 sync.WaitGroup
	start5 := make(chan struct{})
	begin = time.Now()
	for i := 0; i < n; i++ {
		wg4.Add(1)
		go func() {
			defer wg4.Done()
			<-start5
			if _, ok := getCache("BTCUSDT"); ok {
				return // 命中缓存，直接返回
			}
			p, _ := fetchPrice("BTCUSDT", 100*time.Millisecond) // ← 查下游在锁外
			setCache("BTCUSDT", p)
		}()
	}
	close(start5)
	wg4.Wait()
	fmt.Printf("  Mutex + 缓存版：下游被调用 %d 次，耗时 %v\n",
		atomic.LoadInt64(&calls), time.Since(begin).Round(time.Millisecond))
	fmt.Println("  ↑ 10 个请求同时发现「缓存里没有」，于是各自查了一次下游。")
	fmt.Println("    锁保护的是「map 的并发访问」，它管不了「同一时刻大家都 miss」")
	fmt.Println("  如果改成「二次检查 + 锁住整个 miss→fetch 窗口」，调用次数会降到 1 次，")
	fmt.Println("  但代价是所有请求串行排队等锁；singleflight 是「合并等待」——")
	fmt.Println("  大家同时等同一个结果，不排队，还能用 DoChan 配超时")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n【关键总结】")
	fmt.Println("  • singleflight 合并「同一时刻同一个 key」的请求，只放一次真实调用过去")
	fmt.Println("  • 它按 key 分组；不同 key 互不影响；只合并「同时在飞」的，不缓存结果")
	fmt.Println("  • Do 阻塞等结果；DoChan 返回 channel，可配 ctx/超时，结果不会丢")
	fmt.Println("  • 和「Mutex + 缓存」的区别：锁是排队（一个个来），singleflight 是合并")
	fmt.Println("  • 该用：缓存击穿防护（热点行情、热点配置），务必和缓存搭配使用")
	fmt.Println("  • 不该用：需要「每次调用都真实执行」（下单、扣款等有副作用的操作）")
}
