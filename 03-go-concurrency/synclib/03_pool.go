//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 03-go-concurrency && go run synclib/03_pool.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
)

/*
═══════════════════════════════════════════════════════════════════════
L2-16：sync.Pool —— 对象复用
═══════════════════════════════════════════════════════════════════════

【核心概念】
    sync.Pool 是一份「临时对象仓库」：Get 取一个、Put 还一个，用来减少
    短命对象带来的分配和 GC 压力。它并发安全，但**不是缓存**：
    池里的对象随时可能被 GC 清掉，Get 也可能返回 nil 或带脏数据的旧对象。

【你会看到】
    ✓ 同一个对象被反复复用，New 的调用次数不增长
    ✓ 复用出来的 buffer 带着上一轮的残留数据 —— 必须自己 Reset
    ✓ 两次 runtime.GC() 之后池被清空，对象要重新创建
    ✓ 自己写的 channel 池不会被 GC 清掉，但容量和重置都得自己管

【思考题】（动手改代码验证，别空想）
    1. 示例 2 里如果忘了 Reset，把「残留数据」换成一段用户隐私字段，
       会出什么事故？
    2. 示例 3 只调一次 runtime.GC()，池里的对象还在吗？
       （提示：Pool 有一层 victim 缓存，要两次 GC 才清干净）
    3. 示例 4 的 channel 池如果改成「池空时阻塞等待」，
       什么情况下会导致死锁？
    4. 为什么数据库连接池不能用 sync.Pool 实现？

【通关标准】
    能说清 sync.Pool「复用对象」和「缓存对象」的区别，并写出
    「取出必重置」的正确用法

【前置】L2-15              【标签】sync.Pool · GC · 对象复用 · 重置
═══════════════════════════════════════════════════════════════════════
*/

// chanPool 是「自己写的对象池」：用带缓冲的 channel 装对象。
// 与 sync.Pool 的三个关键区别：
//  1. 容量固定、有上限 —— 既不会无限膨胀，也不会因为 GC 被清空；
//  2. Get 永远不返回 nil（池空就新建，或者你可以改成阻塞等待）；
//  3. 「归还前重置」和「不要重复归还」都得自己保证。
type chanPool struct {
	ch chan *bytes.Buffer
}

// newChanPool 预热 n 个 buffer，之后池里最多同时存在 n 个。
func newChanPool(n int) *chanPool {
	p := &chanPool{ch: make(chan *bytes.Buffer, n)}
	for i := 0; i < n; i++ {
		p.ch <- new(bytes.Buffer)
	}
	return p
}

// Get 取一个对象：池里有就复用，没有就新建（不阻塞）。
func (p *chanPool) Get() *bytes.Buffer {
	select {
	case b := <-p.ch:
		return b
	default:
		return new(bytes.Buffer)
	}
}

// Put 归还对象：池满就丢弃，避免归还方被阻塞（这是「有上限」的代价）。
func (p *chanPool) Put(b *bytes.Buffer) {
	b.Reset()
	select {
	case p.ch <- b:
	default:
	}
}

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L2-16 sync.Pool")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// ── 示例 1：对象复用，New 的调用次数不增长 ──────────────────────────
	fmt.Println("📝 示例 1：复用：同一个对象反复取还，New 只调用 4 次")
	fmt.Println(strings.Repeat("-", 60))
	var created int64
	pool := &sync.Pool{New: func() any {
		atomic.AddInt64(&created, 1)
		return new(bytes.Buffer)
	}}
	// 预热：先把 4 个都取出来（池是空的 → New 调用 4 次），再一起还回去
	objs := make([]any, 4)
	for i := range objs {
		objs[i] = pool.Get()
	}
	for _, o := range objs {
		pool.Put(o)
	}
	fmt.Printf("  预热后 New 调用次数 = %d\n", atomic.LoadInt64(&created))
	// 再取还一轮：这次池里有货，直接从池里拿，不该再触发 New
	for i := range objs {
		objs[i] = pool.Get()
	}
	for _, o := range objs {
		pool.Put(o)
	}
	fmt.Printf("  再复用一轮后 New 调用次数 = %d（没涨 → 对象被复用了，没有重新分配）\n",
		atomic.LoadInt64(&created))

	// ── 示例 2：Get 返回的可能不是零值 ──────────────────────────────────
	fmt.Println("📌 示例 2：Get 拿到的可能是「上一轮用过的」对象，必须自己重置")
	fmt.Println(strings.Repeat("-", 60))
	dirty := &sync.Pool{New: func() any { return new(bytes.Buffer) }}
	buf := dirty.Get().(*bytes.Buffer)
	buf.WriteString("上一轮请求的残留数据")
	dirty.Put(buf) // ← 故意不 Reset
	reused := dirty.Get().(*bytes.Buffer)
	fmt.Printf("  再次 Get 到的 buffer 内容 = %q（脏数据！不是零值）\n", reused.String())

	reused.Reset() // 正确姿势：取出后立刻重置，或归还前先重置
	reused.WriteString("干净数据")
	dirty.Put(reused)
	again := dirty.Get().(*bytes.Buffer)
	fmt.Printf("  Reset 之后再取 = %q\n", again.String())
	fmt.Println("  习惯上在 Put 之前 Reset 更好：脏数据不留在池里，别人拿到也是干净的")

	// ── 示例 3：GC 会清空池（它不是缓存）────────────────────────────────
	fmt.Println("🔍 示例 3：两次 runtime.GC() 之后，池里的对象被清掉了")
	fmt.Println(strings.Repeat("-", 60))
	var made int64
	p2 := &sync.Pool{New: func() any {
		atomic.AddInt64(&made, 1)
		return new(bytes.Buffer)
	}}
	// 同样先把 3 个对象都取出来再一起还回去，让池里真的有 3 个
	held := make([]any, 3)
	for i := range held {
		held[i] = p2.Get()
	}
	for _, o := range held {
		p2.Put(o)
	}
	before := atomic.LoadInt64(&made)
	runtime.GC()
	runtime.GC() // Pool 有一层 victim 缓存，所以要两次 GC 才彻底清空
	for i := range held {
		held[i] = p2.Get() // 池已空 → 只能重新创建
	}
	for _, o := range held {
		p2.Put(o)
	}
	after := atomic.LoadInt64(&made)
	fmt.Printf("  GC 前 New 调用次数 = %d，两次 GC 后再取 3 个 = %d（又新建了 %d 个）\n",
		before, after, after-before)
	fmt.Println("  ➡️  sync.Pool 只承诺「复用」不承诺「保住」：对象随时可能被回收，")
	fmt.Println("    所以它不能当缓存用 —— 要缓存就用 map + 锁 / LRU")

	// ── 示例 4：自己写的对象池 ──────────────────────────────────────────
	fmt.Println("🧪 示例 4：自己写的 channel 池：有上限、不被 GC 清掉")
	fmt.Println(strings.Repeat("-", 60))
	cp := newChanPool(3) // 预热 3 个，池里就有 3 个常驻对象
	runtime.GC()
	runtime.GC()
	kept := 0
	for i := 0; i < 3; i++ {
		select {
		case <-cp.ch:
			kept++
		default:
		}
	}
	fmt.Printf("  两次 GC 之后，手写池里还留着 %d 个对象（同一时刻 sync.Pool 已经空了）\n", kept)
	fmt.Println("  代价：容量写死（池满时只能丢弃或阻塞）、要自己防重复归还、要自己 Reset")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ Pool 复用对象以降低分配和 GC 压力；New 是「池空时的兜底工厂」")
	fmt.Println("  ✅ Get 可能返回 nil（没设 New）或带脏数据的旧对象 → 取出必须重置")
	fmt.Println("  ✅ 池随时可能被 GC 清空 → 它是复用工具，不是缓存，不能靠它保数据")
	fmt.Println("  ✅ 该用：高频创建、生命周期极短、创建代价明显的临时对象")
	fmt.Println("    （bytes.Buffer、[]byte 缓冲、JSON 编解码器、protobuf 结构体）")
	fmt.Println("  ✅ 不该用：需要持久缓存（用 map+锁/LRU）、需要严格上限或长连接")
	fmt.Println("    （用带 channel 的对象池 / database/sql 的连接池）")
}
