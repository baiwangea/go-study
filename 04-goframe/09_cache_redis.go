package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	_ "github.com/gogf/gf/contrib/nosql/redis/v2"

	"go-study/goframe/level"
)

// L3-09：g.Redis —— 热点缓存、限频计数、幂等去重。
func L09() level.Level {
	return level.Level{
		ID:      "L3-09",
		Title:   "g.Redis 缓存、限频与幂等",
		Tags:    "TTL · INCR · SetNX · 降级",
		Pre:     "L2-09, L3-04",
		Goal:    "用 Redis 做机器人离不开的三件事：最新价缓存、滑动窗口限频、SetNX 幂等去重",
		Observe: "Set/Get 往返成功；TTL 1s 到期后读到空；INCR 累加到 5；同一订单号第二次 SetNX 返回 false 被丢弃",
		Questions: []string{
			"为什么限频要用 INCR+EXPIRE 而不是本地计数器？（多实例部署时本地各算各的，形同没限）",
			"SetNX 的 key 该设多长 TTL？太短会留下重复下单窗口，太长会堆内存 —— 你怎么定？",
			"这里用 db 9 而不是 db 0，为什么？关卡自己 Del 清理又能完全隔离吗？",
			"Redis 挂掉时，行情缓存和幂等分别该怎么降级？（gcache 本地兜底 + 数据库唯一索引）",
		},
		Check: "会用 SetNX 做幂等、INCR 做限频，并说清 TTL 与库编号的取舍",
		Run: func() {
			ctx := context.Background()
			rdb := g.Redis()
			key := "quote:BTCUSDT"

			if _, err := rdb.Set(ctx, key, "63000.5"); err != nil {
				fmt.Println("  Redis 不可用（config: redis.default.address），本关跳过：", err)
				return
			}
			v, _ := rdb.Get(ctx, key)
			fmt.Printf("  Set/Get 最新价：%s = %s\n", key, v.String())

			short := "quote:ttl-demo"
			_ = rdb.SetEX(ctx, short, "temp", 1)
			before, _ := rdb.Get(ctx, short)
			time.Sleep(1200 * time.Millisecond)
			after, _ := rdb.Get(ctx, short)
			fmt.Printf("  TTL 1s：到期前=%q，1.2s 后=%q（空即已过期）\n", before.String(), after.String())

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
			_, _ = rdb.Del(ctx, key, "processed:ORD-777", short)
			fmt.Println("  已清理本关写入的 key（用的 db 9，与业务库隔离）")
		},
	}
}
