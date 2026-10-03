package client

import "go-study/stdlib-http/level"

// Levels 返回「HTTP 客户端」主题的关卡：01_get / 02_timeout / 03_json / 04_status。
func Levels() []level.Level {
	return []level.Level{L01(), L02(), L03(), L04()}
}
