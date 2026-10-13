//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 04-goframe && go run 07_db_mysql.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
)

/*
═══════════════════════════════════════════════════════════════════════
L3-07：g.DB 建表与增删改查
═══════════════════════════════════════════════════════════════════════

【核心概念】
    用框架的数据层替代手写 sql.DB：注册驱动 → 配 link → Exec/GetAll/聚合，全部带 ctx。
    连接池由框架维护，不需要自己 sql.Open / Close / Ping。

【你会看到】
    ✓ 未设 BOT_DB_LINK 时本关明确跳过并给出设置方法；
    ✓ 设置后能看到建表、插入 3 行、条件查询与 SUM 聚合结果

【思考题】（动手改代码验证，别空想）
    1. 为什么不写任何 Close/Ping？连接池由谁维护？
    2. 把 Exec 改成 Model().Data(g.Map{...}).Insert()，SQL 由谁拼？注入风险还有吗？
    3. loc=Local&parseTime=true 两个参数各解决什么问题？去掉 parseTime 扫描时间列会怎样？
    4. 机器人里为什么「同一秒的行情」不该每笔都落库？（下一步 Redis 去重）

【通关标准】
    能用 g.DB() 完成建表与四种基本操作，并解释 link 各参数含义

【前置】L3-04（配置）      【标签】gdb · 连接串 · Exec/Query · 聚合
═══════════════════════════════════════════════════════════════════════
*/

// 连接串从环境变量读，密码不进 Git：
//
//	export BOT_DB_LINK='mysql:root:你的密码@tcp(127.0.0.1:3308)/go_study?loc=Local&parseTime=true'
func dbLink() string { return os.Getenv("BOT_DB_LINK") }

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 L3-07 g.DB 建表与增删改查")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// ── 示例 0：先检查依赖，缺 BOT_DB_LINK 就明确跳过（不 panic、不卡住） ─
	link := dbLink()
	if link == "" {
		fmt.Println("  未设置 BOT_DB_LINK，本关跳过。设置示例：")
		fmt.Println("    export BOT_DB_LINK='mysql:root:<密码>@tcp(127.0.0.1:3308)/go_study?loc=Local&parseTime=true'")
		fmt.Println("  （关卡刻意不内置任何账号密码）")
		return
	}

	// ── 示例 1：注册驱动配置，拿到 g.DB() ────────────────────────────────
	gdb.SetConfig(gdb.Config{"default": gdb.ConfigGroup{
		gdb.ConfigNode{Link: link, Role: "master"},
	}})
	ctx := context.Background()
	db := g.DB()

	// ── 示例 2：建表（关卡自建自清，跑完不留脏数据） ─────────────────────
	if _, err := db.Exec(ctx, `CREATE TABLE IF NOT EXISTS orders (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			symbol VARCHAR(20) NOT NULL,
			side VARCHAR(4) NOT NULL,
			qty DECIMAL(20,8) NOT NULL,
			price DECIMAL(20,8) NOT NULL,
			status VARCHAR(16) NOT NULL,
			created_at DATETIME NULL,
			UNIQUE KEY uk_symbol_side (symbol, side, price)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`); err != nil {
		fmt.Println("  建表失败：", err)
		return
	}
	if _, err := db.Exec(ctx, "TRUNCATE TABLE orders"); err != nil {
		fmt.Println("  清空失败：", err)
		return
	}
	fmt.Println("  表 orders 就绪（go_study 库，关卡自建自清）")

	// ── 示例 3：批量插入三行 ─────────────────────────────────────────────
	rows := []g.Map{
		{"symbol": "BTCUSDT", "side": "BUY", "qty": 0.1, "price": 63000.5, "status": "filled", "created_at": gtime.Now()},
		{"symbol": "BTCUSDT", "side": "SELL", "qty": 0.05, "price": 63100.0, "status": "submitted", "created_at": gtime.Now()},
		{"symbol": "ETHUSDT", "side": "BUY", "qty": 1.2, "price": 3000.7, "status": "filled", "created_at": gtime.Now()},
	}
	if _, err := db.Model("orders").Data(rows).Insert(); err != nil {
		fmt.Println("  插入失败：", err)
		return
	}

	// ── 示例 4：条件查询 + 聚合 ──────────────────────────────────────────
	total, _ := db.GetCount(ctx, "SELECT COUNT(*) FROM orders")
	filled, _ := db.Model("orders").Where("status", "filled").Count()
	sum, _ := db.Model("orders").Fields("ROUND(SUM(qty*price),2) AS notional").Value()
	list, _ := db.Model("orders").Fields("symbol, side, qty, price").Order("id ASC").All()

	fmt.Printf("  插入后总数=%d，其中 filled=%d，名义金额合计=%s\n", total, filled, sum)
	fmt.Printf("  明细（%d 行）：%s\n", len(list), list.List())

	// ── 示例 5：单行查询 ─────────────────────────────────────────────────
	one, _ := db.Model("orders").Where("symbol", "ETHUSDT").One()
	fmt.Printf("  单行查询 ETHUSDT → side=%s price=%s\n", one["side"], one["price"])

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ➡️  全程没有手写 sql.Open / Close / 扫描，且每条语句都带 ctx")
	fmt.Println("  ✅ 驱动用 _ 空导入注册，配 link 后用 g.DB() 直接拿连接")
	fmt.Println("  ✅ 连接池由框架维护，不需要自己 Ping/Close")
}
