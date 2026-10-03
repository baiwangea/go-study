package contextpkg

import "go-study/go-concurrency/level"

// Levels 返回「context」主题的关卡：01_cancel。
func Levels() []level.Level {
	return []level.Level{L11()}
}
