// Package store 负责「把状态留在进程之外」：订单落库 + 幂等去重。
//
// 设计要点（对应前面关卡）：
//   - 接口定义在使用方（executor.Deduper），实现在这里 —— 隐式实现，换实现不改调用方（L1-09）
//   - MySQL 走 GoFrame 的 gdb，与 L3-07/08 同一套 API
//   - Redis SetNX 与 L3-09 同一个思路；两者都提供内存兜底实现，本地无依赖也能跑
package store

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
)

// Record 是一条订单落库记录。
type Record struct {
	OrderID  string
	Symbol   string
	Side     string
	Qty      float64
	Status   string
	Attempts int
	Ts       time.Time
}

// Store 是持久化契约：只要实现了 Save，机器人不在乎后面是 MySQL 还是切片。
type Store interface {
	Save(ctx context.Context, r Record) error
	Name() string
}

// ---- 内存实现：无数据库时的兜底，也让单元测试可以直接跑 ----

type memoryStore struct {
	mu   sync.Mutex
	rows []Record
}

func NewMemoryStore() Store { return &memoryStore{} }

func (m *memoryStore) Save(_ context.Context, r Record) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rows = append(m.rows, r)
	return nil
}

func (m *memoryStore) Name() string { return "memory（进程退出即丢）" }

// ---- MySQL 实现：连接串只从环境变量来，不进 Git ----

// LinkFromEnv 返回 BOT_DB_LINK；未配置时返回空串，由调用方决定降级策略。
func LinkFromEnv() string { return os.Getenv("BOT_DB_LINK") }

type mysqlStore struct{ db gdb.DB }

// NewMySQL 配置 gdb 并自建 bot_orders 表（表前缀与 L3-07 的演示表隔开）。
func NewMySQL(ctx context.Context, link string) (Store, error) {
	gdb.SetConfig(gdb.Config{"default": gdb.ConfigGroup{
		gdb.ConfigNode{Link: link, Role: "master"},
	}})
	db := g.DB()
	if _, err := db.Exec(ctx, `CREATE TABLE IF NOT EXISTS bot_orders (
		id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
		order_id VARCHAR(40) NOT NULL,
		symbol VARCHAR(20) NOT NULL,
		side VARCHAR(4) NOT NULL,
		qty DECIMAL(20,8) NOT NULL,
		status VARCHAR(16) NOT NULL,
		attempts INT NOT NULL DEFAULT 0,
		ts DATETIME(3) NULL,
		UNIQUE KEY uk_order_id (order_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`); err != nil {
		return nil, fmt.Errorf("建表失败: %w", err)
	}
	return &mysqlStore{db: db}, nil
}

func (m *mysqlStore) Save(ctx context.Context, r Record) error {
	// InsertIgnore：唯一索引 order_id 兜住重复落库，这是幂等的最后一道防线
	_, err := m.db.Model("bot_orders").Ctx(ctx).InsertIgnore(r)
	if err != nil {
		return fmt.Errorf("写入 bot_orders 失败: %w", err)
	}
	return nil
}

func (m *mysqlStore) Name() string { return "MySQL(g.DB) 表 bot_orders" }

// Count 只用于验证：把落库条数读回来。
func (m *mysqlStore) Count(ctx context.Context) (int, error) {
	return m.db.Model("bot_orders").Ctx(ctx).Count()
}

// Counter 是可选能力：实现它的 Store 才能被 main 用来回读条数（行为接口断言，L1-10）。
type Counter interface {
	Count(ctx context.Context) (int, error)
}

// ---- 幂等去重：实现 executor.Deduper ----

type memoryDeduper struct {
	mu   sync.Mutex
	seen map[string]bool
}

// NewMemoryDeduper 单进程内有效，重启即失效。
func NewMemoryDeduper() *memoryDeduper { return &memoryDeduper{seen: map[string]bool{}} }

func (m *memoryDeduper) MarkOnce(_ context.Context, key string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.seen[key] {
		return false, nil
	}
	m.seen[key] = true
	return true, nil
}

func (m *memoryDeduper) Name() string { return "内存 SetNX（重启失效）" }
