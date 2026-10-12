//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 03-go-concurrency && go run synclib/02_rwmutex.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

/*
═══════════════════════════════════════════════════════════════════════
L2-15：sync.RWMutex —— 读写锁
═══════════════════════════════════════════════════════════════════════

【核心概念】
    RWMutex 把「读」和「写」分开：多个读者可以同时待在临界区里，写者独占。
    前提是「读远多于写」—— 只有这种负载下它才划算；写密集时它没有任何优势。

【你会看到】
    ✓ 8 个读锁能同时在临界区里；换成写锁就只剩 1 个
    ✓ 读多写少时 RWMutex 吞吐明显高于 Mutex；写密集时反而略慢
    ✓ 有写者在排队时，新读者也会被挡住（这就是「防写饥饿」）
    ✓ 持有读锁期间再取读锁，一旦有写者排队 → 永久阻塞

【思考题】（动手改代码验证，别空想）
    1. 把示例 1 的 RLock 换成 Lock，maxInflight 会变成几？
    2. 把 readWork 的循环次数从 2000 改成 1（临界区变得极短），
       示例 3 里两者的耗时关系还一样吗？为什么？
    3. 示例 5 里，如果第二层 RLock 在写者排队「之前」就完成了，还会死锁吗？
    4. RWMutex 比 Mutex 多了哪些开销？为什么它不适合写多的场景？

【通关标准】
    能判断一份共享状态该用 Mutex、RWMutex 还是 atomic / channel，
    并说清「递归读锁」和「写者优先」这两个坑

【前置】L2-13              【标签】RWMutex · 读写锁 · 写饥饿 · 死锁
═══════════════════════════════════════════════════════════════════════
*/

// ── 辅助函数 ─────────────────────────────────────────────────────────────

// atomicMax 用 CAS 把 *dst 更新为 max(*dst, v)。
// 这里用它统计「同时在临界区里的最大数量」，避免再引入一把锁把演示搞乱。
func atomicMax(dst *int64, v int64) {
	for {
		old := atomic.LoadInt64(dst)
		if v <= old || atomic.CompareAndSwapInt64(dst, old, v) {
			return
		}
	}
}

// readWork 模拟一次「真实的读操作」：对一小片数据做纯 CPU 计算。
// 之所以要让读操作有点耗时，是因为读锁的价值在于让多个读者真正并行；
// 如果临界区只有一条指令，RWMutex 读锁自身的原子计数开销
// 反而可能盖过并行带来的收益。
func readWork(seed int) int {
	sum := seed
	for i := 0; i < 2000; i++ {
		sum += i % 7
	}
	return sum
}

// benchSink 用来接住基准计算的结果，避免编译器把整个循环优化掉。
var benchSink int64

// benchRead 跑「读多写少」负载：goroutines 个协程各自做 iters 次
// 「加锁 → 读一次共享数据并计算 → 解锁」，返回总耗时。
// 注意这里是「真读」：共享变量 shared 全程只读，计算结果写在各协程的
// 本地变量里 —— 所以 RWMutex.RLock 才能让多个读者安全地并行。
// useRW=true 走 RWMutex.RLock，否则走 Mutex.Lock（连读者也被串行化）。
func benchRead(useRW bool, goroutines, iters int) time.Duration {
	var mu sync.Mutex
	var rw sync.RWMutex
	shared := 12345 // 只读的共享数据
	start := time.Now()
	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			local := 0
			for i := 0; i < iters; i++ {
				if useRW {
					rw.RLock()
					local += readWork(shared)
					rw.RUnlock()
				} else {
					mu.Lock()
					local += readWork(shared)
					mu.Unlock()
				}
			}
			atomic.StoreInt64(&benchSink, int64(local))
		}()
	}
	wg.Wait()
	return time.Since(start)
}

// benchWrite 跑「写密集」负载：临界区里只有一次 v++（真写，所以必须排他）。
// 用来对比写路径的开销：RWMutex.Lock 还要额外维护读者计数和等待队列。
func benchWrite(useRW bool, goroutines, iters int) time.Duration {
	var mu sync.Mutex
	var rw sync.RWMutex
	var v int
	start := time.Now()
	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				if useRW {
					rw.Lock()
					v++
					rw.Unlock()
				} else {
					mu.Lock()
					v++
					mu.Unlock()
				}
			}
		}()
	}
	wg.Wait()
	return time.Since(start)
}

func main() {
	fmt.Println("\n=== L2-15 RWMutex ===")

	// ── 示例 1：读锁可以并发 ────────────────────────────────────────────
	fmt.Println("\n【示例 1】读锁可并发：8 个读者同时待在临界区里")
	var rw sync.RWMutex
	var inflight, maxInflight int64
	const readers = 8
	var wg sync.WaitGroup
	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rw.RLock()
			defer rw.RUnlock()
			atomicMax(&maxInflight, atomic.AddInt64(&inflight, 1))
			time.Sleep(50 * time.Millisecond) // 让 8 个读者真正重叠
			atomic.AddInt64(&inflight, -1)
		}()
	}
	wg.Wait()
	fmt.Printf("  同时在读的最大数量 = %d（期望 8 → 读锁是共享的）\n", atomic.LoadInt64(&maxInflight))

	// ── 示例 2：写锁互斥 ────────────────────────────────────────────────
	fmt.Println("\n【示例 2】写锁互斥：同样 8 个 goroutine，同时在临界区的只有 1 个")
	inflight, maxInflight = 0, 0
	start := time.Now()
	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rw.Lock()
			defer rw.Unlock()
			atomicMax(&maxInflight, atomic.AddInt64(&inflight, 1))
			time.Sleep(50 * time.Millisecond)
			atomic.AddInt64(&inflight, -1)
		}()
	}
	wg.Wait()
	fmt.Printf("  同时在写的最大数量 = %d（期望 1 → 写锁是排他的）\n", atomic.LoadInt64(&maxInflight))
	fmt.Printf("  8 个写者各睡 50ms，串行下来实测耗时 %v（≈ 8×50ms，完全排他）\n",
		time.Since(start).Round(time.Millisecond))

	// ── 示例 3：吞吐对比 —— 读多写少 vs 写密集 ──────────────────────────
	fmt.Println("\n【示例 3】吞吐对比：RWMutex 只在「读多写少」时划算")
	const goroutines, iters = 8, 500
	muRead := benchRead(false, goroutines, iters)
	rwRead := benchRead(true, goroutines, iters)
	ratio := 0.0
	if rwRead > 0 {
		ratio = float64(muRead) / float64(rwRead)
	}
	fmt.Println("  ① 读多写少（临界区里做 ~2000 次循环的读计算）")
	fmt.Printf("     Mutex   ：%v\n", muRead.Round(time.Microsecond))
	fmt.Printf("     RWMutex ：%v  ← 读者并行，Mutex 耗时 / RWMutex 耗时 = %.1f\n",
		rwRead.Round(time.Microsecond), ratio)

	muWrite := benchWrite(false, goroutines, iters*50)
	rwWrite := benchWrite(true, goroutines, iters*50)
	fmt.Println("  ② 写密集（临界区里只有一次 v++）")
	fmt.Printf("     Mutex   ：%v\n", muWrite.Round(time.Microsecond))
	fmt.Printf("     RWMutex ：%v  ← 同样是排他的，写密集没有任何并行收益\n", rwWrite.Round(time.Microsecond))
	fmt.Println("     两者谁快只取决于实现细节（若实测 RWMutex 反而更快，别惊讶：")
	fmt.Println("     Mutex 长时间争抢会进入饥饿模式、把锁直接交接给等待者，调度开销更大）")
	fmt.Println("     所以写多就用更简单的 Mutex，RWMutex 的价值只在「读能并行」这一点上")

	// ── 示例 4：写者优先（防写饥饿）─────────────────────────────────────
	fmt.Println("\n【示例 4】有写者在排队时，新读者也会被挡住")
	var rw2 sync.RWMutex
	rw2.RLock() // 主协程先占住读锁
	writerDone := make(chan struct{})
	go func() {
		rw2.Lock() // 写者被读锁挡住，进入等待队列
		rw2.Unlock()
		close(writerDone)
	}()
	time.Sleep(50 * time.Millisecond) // 给写者时间排进队列
	gotRead := make(chan struct{})
	go func() {
		rw2.RLock() // 新读者：因为已经有写者在等，它会被挡住
		close(gotRead)
		rw2.RUnlock()
	}()
	select {
	case <-gotRead:
		fmt.Println("  新读者立刻拿到读锁 → 写者会被无限拖延（写饥饿）")
	case <-time.After(150 * time.Millisecond):
		fmt.Println("  新读者被挡住 → 排队的写者优先，RWMutex 用这个策略避免写饥饿")
	}
	rw2.RUnlock() // 释放读锁，写者和新读者依次通过
	<-writerDone
	time.Sleep(20 * time.Millisecond)

	// ── 示例 5：递归读锁会死锁 ──────────────────────────────────────────
	fmt.Println("\n【示例 5】持有读锁期间再取读锁 —— 有写者排队时永久阻塞")
	var rw3 sync.RWMutex
	gotFirst := make(chan struct{})
	gotSecond := make(chan struct{})
	trySecond := make(chan struct{}) // 主协程用它通知「写者已经排队了」
	go func() {
		rw3.RLock() // 第一层读锁：拿到
		close(gotFirst)
		<-trySecond // 等主协程把写者放进等待队列
		rw3.RLock() // 第二层读锁（递归）：此时有写者在排队 → 永久阻塞
		close(gotSecond)
		rw3.RUnlock()
		rw3.RUnlock()
	}()
	<-gotFirst
	go func() {
		rw3.Lock() // 写者排队（被第一层读锁挡住）
		rw3.Unlock()
	}()
	time.Sleep(60 * time.Millisecond) // 给写者时间排进队列
	close(trySecond)                  // 现在才去尝试第二层读锁
	select {
	case <-gotSecond:
		fmt.Println("  第二层 RLock 拿到了（没死锁）")
	case <-time.After(150 * time.Millisecond):
		fmt.Println("  第二层 RLock 永久阻塞 → 递归读锁 + 写者排队 = 死锁")
	}
	fmt.Println("  ↑ 这个 goroutine 会永久阻塞（程序退出时由进程回收）；")
	fmt.Println("    所以 RWMutex 的读锁不能嵌套 —— 要嵌套就把读锁提到最外层")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n【关键总结】")
	fmt.Println("  • RLock 共享、Lock 排他；写者排队时新读者也会被挡住（防写饥饿）")
	fmt.Println("  • 读锁不能嵌套：持读锁期间再 RLock，遇到排队的写者就会死锁")
	fmt.Println("  • 只有「读远多于写 + 临界区有一定长度」时 RWMutex 才比 Mutex 快")
	fmt.Println("  • 该用：配置/行情快照这类「大量读、偶尔整体替换」的共享状态")
	fmt.Println("  • 不该用：写多（用 Mutex）、单个计数器（用 atomic）、需要传递所有权（用 channel）")
}
