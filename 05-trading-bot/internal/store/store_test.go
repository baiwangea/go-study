// Package store 的测试：只覆盖**无外部依赖**的部分 —— 内存 Store、内存 Deduper、
// 以及两个环境变量读取函数。
//
// MySQL(g.DB) 与 Redis 两个实现需要真实服务，单测里不碰（README 里靠 BOT_DB_LINK /
// BOT_REDIS_ADDR 手动验证）。测试文件与被测代码同包，所以能直接看未导出的
// memoryStore.rows 来做「按订单号回读」—— 因为 Store 接口本身只暴露了写，没有读。

package store

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"go-study/trading-bot/internal/order"
)

// newRecord 造一条订单记录。Qty 是 decimal（落库列是 DECIMAL(20,8)），
// State 记录订单生命周期状态（不再只写一个终态字符串）。
func newRecord(id string) Record {
	return Record{
		OrderID:  id,
		Symbol:   "BTCUSDT",
		Side:     "BUY",
		Qty:      decimal.RequireFromString("0.002"),
		State:    order.Submitted,
		Status:   "accepted",
		Attempts: 1,
		Ts:       time.Unix(1_700_000_000, 0),
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 存 + 回读：Save 之后能按订单号找回同一条记录（字段不能丢）。
//
// 逐字段比较而不是 `got != want`：decimal.Decimal 内部是 *big.Int 指针，
// 结构体直接比较比的是指针身份，看着过了其实什么都没验证。
// ─────────────────────────────────────────────────────────────────────────────

func TestMemoryStoreSaveAndReadBack(t *testing.T) {
	// 同包测试：断言到具体实现，才能拿到「回读」这个能力
	ms := NewMemoryStore().(*memoryStore)

	want := newRecord("ORD-READ-1")
	if err := ms.Save(context.Background(), want); err != nil {
		t.Fatalf("Save 报错: %v", err)
	}
	if len(ms.rows) != 1 {
		t.Fatalf("落库条数 = %d, want 1", len(ms.rows))
	}

	got := ms.rows[0]
	if got.OrderID != want.OrderID || got.Symbol != want.Symbol || got.Side != want.Side {
		t.Errorf("回读记录 = %+v, want %+v", got, want)
	}
	if !got.Qty.Equal(want.Qty) {
		t.Errorf("回读 Qty = %s, want %s", got.Qty, want.Qty)
	}
	if got.State != want.State || got.Status != want.Status || got.Attempts != want.Attempts || !got.Ts.Equal(want.Ts) {
		t.Errorf("回读记录 = %+v, want %+v", got, want)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 重复 Save 同一订单号：**语义统一成 upsert（覆盖更新）**。
//
// 这一条是上一轮测试留下的「已知不一致」：MySQL 走唯一索引 + InsertIgnore 会静默丢弃，
// 内存实现是纯 append 会落两条。现在两边都改成 upsert：
//   - 同一订单号只留一行；
//   - 后写的内容覆盖先写的 —— 订单状态会随成交推进（submitted → partial → filled），
//     落库必须能反映最新状态，而不是留下两条自相矛盾的记录。
// ─────────────────────────────────────────────────────────────────────────────

func TestMemoryStoreDuplicateSaveUpserts(t *testing.T) {
	ms := NewMemoryStore().(*memoryStore)

	first := newRecord("ORD-DUP")
	if err := ms.Save(context.Background(), first); err != nil {
		t.Fatalf("第一次 Save 报错: %v", err)
	}

	// 同一订单号，状态推进到 filled、成交量也变了
	updated := first
	updated.State = order.Filled
	updated.Status = "accepted"
	updated.Attempts = 3
	updated.Qty = decimal.RequireFromString("0.003")
	if err := ms.Save(context.Background(), updated); err != nil {
		t.Fatalf("第二次 Save 报错: %v", err)
	}

	if len(ms.rows) != 1 {
		t.Fatalf("重复 Save 后条数 = %d, want 1（upsert：同一订单号只留一行）", len(ms.rows))
	}
	if got := ms.rows[0]; got.State != order.Filled {
		t.Errorf("回读 State = %q, want %q（后写的状态必须覆盖先写的）", got.State, order.Filled)
	}
	if got := ms.rows[0]; !got.Qty.Equal(updated.Qty) {
		t.Errorf("回读 Qty = %s, want %s", got.Qty, updated.Qty)
	}
	if got := ms.rows[0]; got.Attempts != 3 {
		t.Errorf("回读 Attempts = %d, want 3", got.Attempts)
	}

	// 再存一个不同订单号：应该新增一行而不是覆盖
	if err := ms.Save(context.Background(), newRecord("ORD-DUP-2")); err != nil {
		t.Fatalf("Save 另一订单号报错: %v", err)
	}
	if len(ms.rows) != 2 {
		t.Errorf("落库条数 = %d, want 2（不同订单号各自一行）", len(ms.rows))
	}
}

// 列表计数：存 N 个**不同**订单号就是 N 条（upsert 只对同号生效）。
func TestMemoryStoreCount(t *testing.T) {
	ms := NewMemoryStore().(*memoryStore)
	const n = 7
	for i := 0; i < n; i++ {
		if err := ms.Save(context.Background(), newRecord(fmt.Sprintf("ORD-%c", 'A'+i))); err != nil {
			t.Fatalf("Save 第 %d 条报错: %v", i, err)
		}
	}
	if len(ms.rows) != n {
		t.Errorf("落库条数 = %d, want %d", len(ms.rows), n)
	}
	if len(ms.idx) != n {
		t.Errorf("索引条数 = %d, want %d", len(ms.idx), n)
	}
}

func TestMemoryStoreName(t *testing.T) {
	if got := NewMemoryStore().Name(); got == "" {
		t.Error("Name() 不该为空（main 的汇总行靠它说明状态存在哪）")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 并发 Save 不 panic、不丢记录。
//
// 配 `go test -race` 跑：memoryStore 内部用 Mutex 保护 rows/idx，这里就是它的守护测试。
//
// ⚠ 每个 (goroutine, 序号) 组合用**唯一订单号**：upsert 语义下重复订单号会合并成一行，
// 那样「不丢记录」就无从断言了。订单号唯一时，总行数仍必须等于总调用次数。
// ─────────────────────────────────────────────────────────────────────────────

func TestMemoryStoreConcurrentSave(t *testing.T) {
	ms := NewMemoryStore().(*memoryStore)

	const (
		goroutines = 32
		perG       = 20
	)
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(g int) {
			defer wg.Done()
			for i := 0; i < perG; i++ {
				_ = ms.Save(context.Background(), newRecord(fmt.Sprintf("ORD-%d-%d", g, i)))
			}
		}(g)
	}
	wg.Wait()

	if want := goroutines * perG; len(ms.rows) != want {
		t.Errorf("并发落库条数 = %d, want %d（不能丢记录）", len(ms.rows), want)
	}
	if want := goroutines * perG; len(ms.idx) != want {
		t.Errorf("并发落库索引条数 = %d, want %d", len(ms.idx), want)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 内存 Deduper（store 版，满足 executor.Deduper）：第一次 true、第二次 false。
// 与 executor.NewMemoryDeduper 是两个独立类型（README 里说明是刻意的），各自测。
// ─────────────────────────────────────────────────────────────────────────────

func TestMemoryDeduper(t *testing.T) {
	d := NewMemoryDeduper()
	ctx := context.Background()

	first, err := d.MarkOnce(ctx, "ORD-X")
	if err != nil {
		t.Fatalf("首次 MarkOnce 报错: %v", err)
	}
	if !first {
		t.Error("第一次应返回 true")
	}

	second, err := d.MarkOnce(ctx, "ORD-X")
	if err != nil {
		t.Fatalf("第二次 MarkOnce 报错: %v", err)
	}
	if second {
		t.Error("第二次应返回 false")
	}

	if d.Name() == "" {
		t.Error("Name() 不该为空")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 环境变量读取：空 → ""，有值 → 原样返回（main 靠它决定降级到内存实现）。
// 用 t.Setenv，测试结束自动还原，不会污染其它用例。
// ─────────────────────────────────────────────────────────────────────────────

func TestEnvReaders(t *testing.T) {
	tests := []struct {
		name string
		env  string
		val  string
		get  func() string
	}{
		{"BOT_DB_LINK 未设置返回空串", "BOT_DB_LINK", "", LinkFromEnv},
		{"BOT_DB_LINK 有值原样返回", "BOT_DB_LINK", "mysql:root:pw@tcp(127.0.0.1:3308)/go_study", LinkFromEnv},
		{"BOT_REDIS_ADDR 未设置返回空串", "BOT_REDIS_ADDR", "", RedisAddrFromEnv},
		{"BOT_REDIS_ADDR 有值原样返回", "BOT_REDIS_ADDR", "127.0.0.1:6379", RedisAddrFromEnv},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.env, tt.val)
			if got := tt.get(); got != tt.val {
				t.Errorf("%s = %q, want %q", tt.env, got, tt.val)
			}
		})
	}
}
