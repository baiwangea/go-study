package goroutines

import "go-study/go-concurrency/level"

// Levels 返回「协程与等待」主题的关卡：01_launch / 02_waitgroup。
func Levels() []level.Level {
	return []level.Level{
		L01(),
		L02(),
	}
}
