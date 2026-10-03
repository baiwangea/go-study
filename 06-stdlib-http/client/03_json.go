package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"

	"go-study/stdlib-http/level"
)

type OrderReq struct {
	Symbol string `json:"symbol"`
	Side   string `json:"side"`
	Qty    string `json:"qty"`           // 数字字符串：交易所惯例，避免浮点误差
	Tag    string `json:"tag,omitempty"` // 空值时整个字段消失
	Secret string `json:"-"`             // 永不序列化出去
}

// L4-03：JSON 编解码与金额精度。
func L03() level.Level {
	return level.Level{
		ID:      "L4-03",
		Title:   "struct tag 与金额精度",
		Tags:    "encoding/json · tag · big.Int · 浮点陷阱",
		Pre:     "L4-01",
		Goal:    "掌握 tag 的三个必备写法，并理解为什么金额绝不能用 float64",
		Observe: "omitempty 让字段消失、`json:\"-\"` 屏蔽敏感字段；0.1+0.2 的浮点误差清晰可见",
		Questions: []string{
			"把 tag 改成没有 `json:` 标签的字段，序列化结果变成什么？（提示：必须首字母大写才导出）",
			"用 float64 存 1 ETH = 1e18 wei，和 big.Int 对比打印，找出精度丢失发生在哪一步。",
			"反序列化时字段类型不匹配（对方把 qty 发成数字 1）会报什么错？如何做容错解析？",
		},
		Check: "会用 json.Marshal 输出请求体、Unmarshal 到结构体，并解释金额用 string/big.Int 的原因",
		Run: func() {
			body := OrderReq{Symbol: "BTCUSDT", Side: "BUY", Qty: "0.1", Secret: "my-api-secret"}
			raw, _ := json.Marshal(body)
			fmt.Println("  请求体：", string(raw)) // tag 与 Secret 不出现

			var back map[string]any
			json.Unmarshal(raw, &back)
			fmt.Printf("  解成 map 会丢字段顺序与类型信息：%v\n", back)

			f := 0.1
			fmt.Printf("  float64 累加：0.1+0.2 = %.20f ← 这就是为什么交易所传字符串\n", f+f)
			wei, _ := new(big.Int).SetString("1000000000000000000", 10)
			fmt.Printf("  big.Int 精确表示 1 ETH = %s wei\n", wei)

			enc, _ := json.Marshal(map[string]any{"qty": "0.1"})
			fmt.Println("  POST 请求体写法：bytes.NewReader(", string(enc), ") 见下一关的服务端")
			_ = bytes.NewReader
		},
	}
}
