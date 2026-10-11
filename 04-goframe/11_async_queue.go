//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 04-goframe && go run 11_async_queue.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"fmt"
	"sync"
	"time"
)

/*
═══════════════════════════════════════════════════════════════════════
L3-11：异步任务：重试、幂等与优雅停止
═══════════════════════════════════════════════════════════════════════

【核心概念】
    把「下单提交」这类不可靠动作放进带重试与幂等的异步通道，接口本身立刻返回。
    worker pool + 指数退避重试 + map 幂等去重 + close(chan) 优雅停止。

【你会看到】
    ✓ 同一个任务 ID 重复入队只执行一次；失败任务按 100/200ms 退避重试后成功；退出前在途任务被排空

【思考题】（动手改代码验证，别空想）
    1. 幂等表用 map + Mutex 只在本进程有效 —— 两个实例同时跑会怎样？（这正是 Redis SETNX / 数据库唯一索引的价值）
    2. 重试为什么必须配合幂等？没有幂等时「超时后重试」会造成什么后果？
    3. 这里的队列是内存的，进程崩溃任务就丢了。要持久化该引入什么（Redis/MySQL 任务表）？

【通关标准】
    能默写出「提交即返回 + 后台重试 + 幂等去重 + 优雅停止」的最小实现

【前置】L2-12（worker pool）      【标签】worker pool · 指数退避 · 幂等 · 优雅退出
═══════════════════════════════════════════════════════════════════════
*/

type submitTask struct {
	ID      string // 业务唯一 ID，用于幂等
	Attempt int
}

func main() {
	fmt.Println("\n=== L3-11 异步任务：重试、幂等与优雅停止 ===")

	const workers = 2
	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		processed = map[string]bool{}
		results   []string
	)
	done := make(chan struct{})
	tasks := make(chan submitTask, 8)

	// alreadyDone / markDone 是进程内幂等表：已处理过的 ID 直接跳过。
	alreadyDone := func(id string) bool {
		mu.Lock()
		defer mu.Unlock()
		return processed[id]
	}
	markDone := func(id string) {
		mu.Lock()
		processed[id] = true
		mu.Unlock()
	}

	// submit 模拟调用交易所：SUB-2 前两次失败
	submit := func(t submitTask) error {
		if t.ID == "SUB-2" && t.Attempt < 2 {
			return fmt.Errorf("网络抖动（第 %d 次尝试）", t.Attempt+1)
		}
		return nil
	}

	// ── 示例 1：起 worker pool 消费队列 ──────────────────────────────────
	fmt.Println("\n【示例 1】2 个 worker 消费任务，失败按 100/200ms 退避重试")
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for t := range tasks {
				if alreadyDone(t.ID) {
					mu.Lock()
					results = append(results, fmt.Sprintf("worker-%d 跳过重复任务 %s", id, t.ID))
					mu.Unlock()
					continue
				}
				for attempt := 0; attempt < 3; attempt++ {
					task := submitTask{ID: t.ID, Attempt: attempt}
					if err := submit(task); err != nil {
						delay := time.Duration(100*(1<<attempt)) * time.Millisecond
						mu.Lock()
						results = append(results, fmt.Sprintf("worker-%d %s 第 %d 次失败：%v → %v 后重试", id, t.ID, attempt+1, err, delay))
						mu.Unlock()
						time.Sleep(delay) // 真实场景还应尊重交易所的 Retry-After
						continue
					}
					markDone(t.ID)
					mu.Lock()
					results = append(results, fmt.Sprintf("worker-%d %s 提交成功（尝试 %d 次）", id, t.ID, attempt+1))
					mu.Unlock()
					break
				}
			}
			done <- struct{}{}
		}(i + 1)
	}

	// ── 示例 2：投递任务并优雅停止 ───────────────────────────────────────
	// 投递：SUB-1 正常、SUB-2 需重试、SUB-1 重复投递两次
	fmt.Println("\n【示例 2】投递 4 次（2 个 ID），close 后等 worker 自然退出")
	for _, id := range []string{"SUB-1", "SUB-2", "SUB-1", "SUB-1"} {
		tasks <- submitTask{ID: id}
	}
	close(tasks)
	for i := 0; i < workers; i++ {
		<-done // 等 worker 自然退出，而不是靠 sleep 猜时间
	}
	wg.Wait()

	// ── 结果 ─────────────────────────────────────────────────────────────
	fmt.Println("\n【运行结果】")
	for _, r := range results {
		fmt.Println("  " + r)
	}
	fmt.Printf("  实际执行的唯一任务数：%d（4 次投递，2 个 ID）\n", len(processed))

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n【关键总结】")
	fmt.Println("  • 重试必须配合幂等：否则「超时后重试」会造成重复下单")
	fmt.Println("  • 内存队列进程崩溃就丢任务；要持久化需引入 Redis/MySQL 任务表")
	fmt.Println("  • 优雅停止靠 close(chan) + 等 worker 退出，而不是 sleep 猜时间")
}
