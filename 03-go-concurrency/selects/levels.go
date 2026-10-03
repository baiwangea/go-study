package selects

import "go-study/go-concurrency/level"

// Levels 返回「select」主题的关卡：01_multiplex / 02_timeout / 03_nonblock / 04_ticker。
func Levels() []level.Level {
	return []level.Level{
		L07(),
		L08(),
		L09(),
		L10(),
	}
}
