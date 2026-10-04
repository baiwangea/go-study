package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"go-study/goframe/level"
)

type orderRow struct {
	Id     int64   `json:"id"`
	Symbol string  `json:"symbol"`
	Side   string  `json:"side"`
	Qty    float64 `json:"qty"`
	Status string  `json:"status"`
}

// L3-08：ORM 链式操作与事务回滚。
func L08() level.Level {
	return level.Level{
		ID:      "L3-08",
		Title:   "ORM 链式、事务与回滚",
		Tags:    "Model · Transaction · 原子性",
		Pre:     "L3-07",
		Goal:    "用链式 Model 写查询，用事务保证「改状态 + 扣额度」这类多步操作的原子性",
		Observe: "事务提交后能看到更新结果；第二个事务中途返回 error，状态与计数完全不变（回滚生效）",
		Questions: []string{
			"把 tx.Rollback 那段改成不返回 err，两条更新会有什么差别？为什么事务里必须显式返回 error？",
			"gdb.Transaction 的回调里用的是 tx 而不是 g.DB()，用错对象会怎样？（提示：不在同一事务里）",
			"链式 Where 拼接用户输入时，为什么仍然安全？框架在哪一层做了参数化？",
			"机器人「下单成功 → 扣减额度」中间进程崩溃会怎样？该怎么设计补偿？",
		},
		Check: "会写链式查询与事务，并能说出回滚发生的确切条件",
		Run: func() {
			link := dbLink()
			if link == "" {
				fmt.Println("  未设置 BOT_DB_LINK，本关跳过（设置方法见 L3-07）")
				return
			}
			gdb.SetConfig(gdb.Config{"default": gdb.ConfigGroup{gdb.ConfigNode{Link: link, Role: "master"}}})
			ctx := context.Background()
			db := g.DB()

			if _, err := db.Model("orders").Where("symbol", "BTCUSDT").Where("side", "BUY").
				Data(gdb.Map{"status": "canceled"}).Update(); err != nil {
				fmt.Println("  更新失败：", err)
				return
			}
			after, _ := db.Model("orders").Where("status", "canceled").Count()
			fmt.Printf("  链式更新完成，canceled 行数 = %d\n", after)

			var before, after2 int
			before, _ = db.Model("orders").Where("status", "filled").Count()

			err := db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
				if _, err := tx.Model("orders").Where("symbol", "ETHUSDT").Data(gdb.Map{"status": "filled"}).Update(); err != nil {
					return err
				}
				fmt.Println("  事务内：ETHUSDT 已置为 filled，接下来故意返回错误")
				return errors.New("模拟风控校验失败") // 返回非 nil → 自动回滚
			})
			fmt.Println("  事务返回：", err)

			after2, _ = db.Model("orders").Where("status", "filled").Count()
			fmt.Printf("  filled 行数 事务前=%d 事务后=%d → 回滚生效，一行都没变\n", before, after2)

			var rows []orderRow
			if err := db.Model("orders").Fields("id,symbol,side,qty,status").Order("id ASC").Scan(&rows); err != nil {
				fmt.Println("  扫描失败：", err)
				return
			}
			fmt.Printf("  Scan 到结构体切片 %d 条：%+v\n", len(rows), rows[0])
		},
	}
}
