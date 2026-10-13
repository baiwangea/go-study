//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 03-go-concurrency && go run synclib/04_syncmap.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

/*
═══════════════════════════════════════════════════════════════════════
L2-17：sync.Map —— 并发安全的 map
═══════════════════════════════════════════════════════════════════════

【核心概念】
    普通 map 并发读写会被 runtime 主动杀掉（fatal error，不可 recover）。
    sync.Map 用「只读副本 + 原子指针」实现免锁读，适合「读多写少、
    key 相对稳定」的场景 —— 它不是通用 map 的替代品。

【你会看到】
    ✓ 普通 map 并发写 → fatal error: concurrent map writes（子进程里复现）
    ✓ sync.Map 的 Load / Store / LoadOrStore / Delete / Range 全套用法
    ✓ 8 写 4 读并发跑完 4 万条，程序安然无恙
    ✓ Load + Store 不是原子的 → 不能拿它当计数器

【思考题】（动手改代码验证，别空想）
    1. 示例 1 的子进程如果换成「一个 goroutine 读、一个写」，
       报错信息会变成什么？（提示：concurrent map read and map write）
    2. 为什么 fatal error 用 recover 兜不住？和 panic 有什么本质区别？
    3. 示例 3 里把 8 个写者改成 1 个写者 + 7 个读者，还会崩吗？
       那为什么工程上仍然推荐用 sync.Map？
    4. sync.Map 没有 len()，那「统计在线连接数」该怎么写？

【通关标准】
    能说清 sync.Map 的适用边界（读多写少 / key 稳定），
    并知道它「读-改-写不原子」「没有 len」这两个硬伤

【前置】L2-16              【标签】sync.Map · concurrent map writes · 选型
═══════════════════════════════════════════════════════════════════════
*/

// crashChild 是「子进程模式」的入口：它故意让多个 goroutine 同时写同一个
// 普通 map。fatal error 不可 recover，会直接杀掉进程，所以必须放到子进程里跑，
// 否则本关自己的退出码就不是 0 了。
func crashChild() {
	m := make(map[int]int)
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 200000; i++ {
				m[g*1000000+i] = i // ← 多个 goroutine 同时写同一个 map
			}
		}(g)
	}
	wg.Wait()
	fmt.Println("  竟然没崩？那就再加点并发量试试")
}

// runCrashInChild 重新执行本进程（进入 __crash 模式），把子进程的
// fatal error 抓回来展示。这是唯一能「不让本关自己崩溃」的观察办法。
func runCrashInChild() {
	cmd := exec.Command(os.Args[0], "__crash")
	out, err := cmd.CombinedOutput()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	}
	fmt.Printf("  子进程退出码 = %d\n", code)
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "fatal error") {
			fmt.Printf("  子进程输出：%s\n", strings.TrimSpace(line))
			break
		}
	}
	fmt.Println("  ➡️  fatal error 由 runtime 直接抛给操作系统，不是 panic → recover 兜不住")
}

// runReads 并发跑一批读操作（8 个 goroutine 分摊），
// 用来粗略对比两种容器的读路径开销。
func runReads(reads int, do func(i int)) {
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := g; i < reads; i += 8 {
				do(i)
			}
		}(g)
	}
	wg.Wait()
}

func main() {
	// 子进程模式：只跑会崩的那段，不打印横幅
	if len(os.Args) > 1 && os.Args[1] == "__crash" {
		crashChild()
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L2-17 sync.Map")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// ── 示例 1：普通 map 并发写 → fatal error ───────────────────────────
	fmt.Println("📝 示例 1：普通 map 并发写：不是「可能出错」，而是必崩")
	fmt.Println(strings.Repeat("-", 60))
	runCrashInChild()
	fmt.Println("  结论：普通 map 只能在「单 goroutine 写」的前提下用，否则必须加锁")

	// ── 示例 2：sync.Map 基本操作 ───────────────────────────────────────
	fmt.Println("📌 示例 2：Load / Store / LoadOrStore / Delete / Range")
	fmt.Println(strings.Repeat("-", 60))
	var sm sync.Map
	sm.Store("BTCUSDT", 68000.0)
	sm.Store("ETHUSDT", 3500.0)
	if v, ok := sm.Load("BTCUSDT"); ok {
		fmt.Printf("  Load(\"BTCUSDT\")      = %v, ok=%v\n", v, ok)
	}
	if _, ok := sm.Load("SOLUSDT"); !ok {
		fmt.Println("  Load(\"SOLUSDT\")      → ok=false（不存在的 key 不 panic，靠 ok 判断）")
	}
	// LoadOrStore：key 不存在才写入；返回「实际留在 map 里的值」+「是否已存在」
	actual, loaded := sm.LoadOrStore("BTCUSDT", 99999.0)
	fmt.Printf("  LoadOrStore(BTCUSDT) → 返回 %v, loaded=%v（已存在 → 保留旧值，写入被忽略）\n", actual, loaded)
	actual2, loaded2 := sm.LoadOrStore("SOLUSDT", 150.0)
	fmt.Printf("  LoadOrStore(SOLUSDT) → 返回 %v, loaded=%v（不存在 → 真的写进去了）\n", actual2, loaded2)
	sm.Delete("ETHUSDT")
	// Range 的顺序不确定，这里排个序保证输出稳定
	items := []string{}
	sm.Range(func(k, v any) bool {
		items = append(items, fmt.Sprintf("%v=%v", k, v))
		return true // 返回 false 会提前结束遍历
	})
	sort.Strings(items)
	fmt.Printf("  Delete 掉 ETHUSDT 后 Range：%v\n", items)
	fmt.Println("  注意：sync.Map 没有 len()；Range 也不是一致性快照（并发写时可能看到部分更新）")

	// ── 示例 3：并发读写不会崩 ──────────────────────────────────────────
	fmt.Println("🔍 示例 3：8 个写者 + 4 个读者并发跑，程序安然无恙")
	fmt.Println(strings.Repeat("-", 60))
	var sm2 sync.Map
	const writers, each = 8, 5000
	var wg sync.WaitGroup
	for g := 0; g < writers; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < each; i++ {
				sm2.Store(fmt.Sprintf("k-%d-%d", g, i), i)
			}
		}(g)
	}
	for r := 0; r < 4; r++ {
		wg.Add(1)
		go func(r int) {
			defer wg.Done()
			for i := 0; i < each; i++ {
				sm2.Load(fmt.Sprintf("k-%d-%d", r, i)) // 读不存在的 key 同样安全
			}
		}(r)
	}
	wg.Wait()
	total := 0
	sm2.Range(func(k, v any) bool { total++; return true })
	fmt.Printf("  8 个写者 × %d 条 + 4 个读者并发 → 最终 %d 条，没崩\n", each, total)

	// ── 示例 4：Load+Store 不是原子的 ───────────────────────────────────
	fmt.Println("🧪 示例 4：sync.Map 的 Load+Store 不是原子操作，不能当计数器")
	fmt.Println(strings.Repeat("-", 60))
	// 把两个 goroutine 的交错「手工摆出来」，看清丢更新的全过程
	var sm3 sync.Map
	sm3.Store("n", 0)
	a, _ := sm3.Load("n")     // 协程 A 读到 0
	b, _ := sm3.Load("n")     // 协程 B 也读到 0
	sm3.Store("n", a.(int)+1) // A 写回 1
	sm3.Store("n", b.(int)+1) // B 写回 1 ← 把 A 的结果覆盖掉了
	n, _ := sm3.Load("n")
	fmt.Printf("  两次「自增」之后 n = %v（期望 2，实际 1 —— 丢了一次更新）\n", n)

	// 正确写法：存 *int64 指针，用 atomic 累加（所有 goroutine 共享同一个计数器）
	var sm4 sync.Map
	sm4.Store("n", new(int64))
	v, _ := sm4.Load("n")
	counter := v.(*int64)
	var wg2 sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			for i := 0; i < 1000; i++ {
				atomic.AddInt64(counter, 1)
			}
		}()
	}
	wg2.Wait()
	fmt.Printf("  正确写法（*int64 + atomic）8×1000 次自增 = %d\n", atomic.LoadInt64(counter))
	fmt.Println("  ➡️  sync.Map 只保证「单个操作」原子；复合逻辑（读-改-写）还得自己配 atomic 或锁")

	// ── 示例 5：和 RWMutex + map 的对比 ─────────────────────────────────
	fmt.Println("⚙️ 示例 5：sync.Map vs RWMutex + map")
	fmt.Println(strings.Repeat("-", 60))
	const keys, reads = 64, 200000
	var sm5 sync.Map
	for i := 0; i < keys; i++ {
		sm5.Store(i, i)
	}
	start := time.Now()
	runReads(reads, func(i int) { sm5.Load(i % keys) })
	smTime := time.Since(start)

	var mu sync.RWMutex
	m := make(map[int]int, keys)
	for i := 0; i < keys; i++ {
		m[i] = i
	}
	start = time.Now()
	runReads(reads, func(i int) {
		mu.RLock()
		_ = m[i%keys]
		mu.RUnlock()
	})
	rwTime := time.Since(start)

	fmt.Printf("  读多写少（%d 个固定 key，%d 次读，全部命中）：\n", keys, reads)
	fmt.Printf("    sync.Map      ：%v\n", smTime.Round(time.Microsecond))
	fmt.Printf("    RWMutex + map ：%v\n", rwTime.Round(time.Microsecond))
	fmt.Println("  这组数字只作参考（真正的对比要用 testing.B 的 benchmark）：")
	fmt.Println("  sync.Map 的读走「只读副本 + atomic」，不用抢锁，key 稳定时更省；")
	fmt.Println("  但 key 频繁增删时要维护 read/dirty 两份表、内存只增不减，")
	fmt.Println("  而且没有 len()、值是 any 要类型断言 —— 所以别把它当通用 map 用")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ 普通 map 并发写是 fatal error（不可 recover），必须加锁或换 sync.Map")
	fmt.Println("  ✅ sync.Map 五个操作：Load / Store / LoadOrStore / Delete / Range")
	fmt.Println("  ✅ Load+Store 不是原子的；计数器要存 *int64 配 atomic，或直接用锁")
	fmt.Println("  ✅ 该用：读远多于写、key 相对稳定（注册表、连接元数据、配置快照）")
	fmt.Println("  ✅ 不该用：当通用 map（没有 len、类型是 any、写多时更慢更费内存）")
	fmt.Println("    —— 写多或要遍历/统计，就用 map + RWMutex，语义更直白")
}
