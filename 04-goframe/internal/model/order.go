// Package model 只放数据结构与校验规则，不依赖 HTTP，也不含业务逻辑。
package model

// PlaceOrder 是下单的出入参模型：v 标签负责参数校验。
type PlaceOrder struct {
	Symbol string  `json:"symbol" v:"required|length:2,20#交易对必填|交易对长度 2-20"`
	Side   string  `json:"side"   v:"required|in:BUY,SELL#方向必填|方向只能是 BUY 或 SELL"`
	Qty    float64 `json:"qty"    v:"required|min:0.000001#数量必填|数量太小"`
}

// Order 是落库/返回用的订单对象。
type Order struct {
	OrderID string  `json:"order_id"`
	Symbol  string  `json:"symbol"`
	Side    string  `json:"side"`
	Qty     float64 `json:"qty"`
	Status  string  `json:"status"`
}
