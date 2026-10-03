package server

import "go-study/stdlib-http/level"

// Levels 返回「HTTP 服务端」主题的关卡：01_minimal / 02_middleware / 03_context / 04_body_guard。
func Levels() []level.Level {
	return []level.Level{L05(), L06(), L07(), L08()}
}
