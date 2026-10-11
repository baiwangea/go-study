//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 04-goframe && go run 09_cache_redis.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	_ "github.com/gogf/gf/contrib/nosql/redis/v2"
)

/*
═══════════════════════════════════════════════════════════════════════
L3-09：g.Redis 缓存、限频与幂等
═══════════════════════════════════════════════════════════════════════

【核心概念】
    用 Redis 做机器人离不开的三件事：最新价缓存（Set/Get + TTL）、
    滑动窗口限频（INCR + EXPIRE）、幂等去重（SetNX）。

【你会看到】
    ✓ Set/Get 往返成功；TTL 1s 到期后读到空；INCR 累加到 5；同一订单号第二次 SetNX 返回 false 被丢弃

【思考题】（动手改代码验证，别空想）
    1. 为什么限频要用 INCR+EXPIRE 而不是本地计数器？（多实例部署时本地各算各的，形同没限）
    2. SetNX 的 key 该设多长 TTL？太短会留下重复下单窗口，太长会堆内存 —— 你怎么定？
    3. 这里用 db 9 而不是 db 0，为什么？关卡自己 Del 清理又能完全隔离吗？
    4. Redis 挂掉时，行情缓存和幂等分别该怎么降级？（gcache 本地兜底 + 数据库唯一索引）

【通关标准】
    会用 SetNX 做幂等、INCR 做限频，并说清 TTL 与库编号的取舍

【前置】L2-09, L3-04      【标签】TTL · INCR · SetNX · 降级
═══════════════════════════════════════════════════════════════════════
*/

func main() {
	fmt.Println("\n=== L3-09 g.Redis 缓存、限频与幂等 ===")
	ctx := context.Background()
	rdb := g.Redis()
	key := "quote:BTCUSDT"

	// ── 示例 0：Redis 不可用就明确跳过（不 panic、不卡住） ───────────────
	if _, err := rdb.Set(ctx, key, "63000.5"); err != nil {
		fmt.Println("  Redis 不可用（config: redis.default.address），本关跳过：", err)
		return
	}

	// ── 示例 1：最新价缓存 ───────────────────────────────────────────────
	v, _ := rdb.Get(ctx, key)
	fmt.Printf("  Set/Get 最新价：%s = %s\n", key, v.String())

	// ── 示例 2：TTL 到期后读到空 ─────────────────────────────────────────
	short := "quote:ttl-demo"
	_ = rdb.SetEX(ctx, short, "temp", 1)
	before, _ := rdb.Get(ctx, short)
	time.Sleep(1200 * time.Millisecond)
	after, _ := rdb.Get(ctx, short)
	fmt.Printf("  TTL 1s：到期前=%q，1.2s 后=%q（空即已过期）\n", before.String(), after.String())

	// ── 示例 3：INCR + EXPIRE 做滑动窗口限频 ─────────────────────────────
	rl := "ratelimit:order"
	_, _ = rdb.Del(ctx, rl)
	var n int64
	for i := 0; i < 5; i++ {
		n, _ = rdb.Incr(ctx, rl)
		if n == 1 {
			_, _ = rdb.Expire(ctx, rl, 10) // 10 秒窗口
		}
	}
	fmt.Printf("  连发 5 次请求后计数 = %d，窗口 10s 内超过 3 次就应拒绝下单\n", n)
	_, _ = rdb.Del(ctx, rl)

	// ── 示例 4：SetNX 做幂等去重 ─────────────────────────────────────────
	for i := 1; i <= 3; i++ {
		ok, err := rdb.SetNX(ctx, "processed:ORD-777", "1")
		if err != nil {
			fmt.Println("  SetNX 失败：", err)
			break
		}
		verdict := "已处理过 → 丢弃（幂等生效）"
		if ok {
			verdict = "首次出现 → 放行"
		}
		fmt.Printf("  第 %d 次处理 ORD-777：SetNX=%-5v %s\n", i, ok, verdict)
	}

	// ── 清理：本关写入的 key 全部删掉 ────────────────────────────────────
	_, _ = rdb.Del(ctx, key, "processed:ORD-777", short)
	fmt.Println("  已清理本关写入的 key（用的 db 9，与业务库隔离）")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n【关键总结】")
	fmt.Println("  • 限频用 INCR+EXPIRE 而非本地计数器：多实例部署时本地各算各的形同没限")
	fmt.Println("  • 幂等用 SetNX：首次 true 放行、再次 false 丢弃")
	fmt.Println("  • Redis 挂掉时缓存可降级到本地 gcache，幂等兜底靠数据库唯一索引")
}
