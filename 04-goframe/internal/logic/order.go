// Package logic 是业务层：不出现 *ghttp.Request，也不出现 SQL，只处理业务规则。
// 这样才能被 HTTP、定时任务、命令行工具复用，也才能写单元测试。
package logic

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go-study/goframe/internal/model"
)

var (
	mu     sync.Mutex
	seq    int
	maxUSD = 500.0 // 单笔上限，实际应从 g.Cfg() 读取
)

// SetMaxPosition 让测试与上层配置可以调整风控阈值。
func SetMaxPosition(usd float64) {
	mu.Lock()
	defer mu.Unlock()
	maxUSD = usd
}

// Place 执行下单业务：风控校验 → 生成订单号 → 返回订单状态。
func Place(ctx context.Context, in *model.PlaceOrder) (*model.Order, error) {
	if in.Qty*63000 > maxUSD {
		return nil, fmt.Errorf("超出单笔限额 %.0f USD", maxUSD)
	}
	if in.Symbol == "DELISTED" {
		return nil, errors.New("交易对已下架")
	}

	mu.Lock()
	seq++
	id := fmt.Sprintf("ORD-%04d", seq)
	mu.Unlock()

	return &model.Order{OrderID: id, Symbol: in.Symbol, Side: in.Side, Qty: in.Qty, Status: "submitted"}, nil
}
