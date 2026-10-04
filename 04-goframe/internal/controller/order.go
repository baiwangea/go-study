// Package controller 只做「协议 ↔ 业务」的翻译：解析入参、调用 logic、组织返回。
// 一旦出现 SQL 或 HTTP 细节，就说明分层被破坏了。
package controller

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"go-study/goframe/internal/logic"
	"go-study/goframe/internal/model"
)

type OrderReq struct {
	g.Meta `path:"/api/order" method:"post" tags:"交易" summary:"下单"`
	model.PlaceOrder
}

type OrderRes struct {
	*model.Order
}

type cOrder struct{}

var Order = cOrder{}

// Place 是标准出入参写法：框架负责绑定与校验，这里只调业务层。
func (c cOrder) Place(ctx context.Context, req *OrderReq) (res *OrderRes, err error) {
	order, err := logic.Place(ctx, &req.PlaceOrder)
	if err != nil {
		return nil, err
	}
	return &OrderRes{Order: order}, nil
}
