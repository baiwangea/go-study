package channels

import "go-study/go-concurrency/level"

// Levels 返回「channel」主题的关卡：01_unbuffered / 02_buffered / 03_close_range / 04_deadlock。
func Levels() []level.Level {
	return []level.Level{
		L03(),
		L04(),
		L05(),
		L06(),
	}
}
