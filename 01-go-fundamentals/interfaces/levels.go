package interfaces

import "go-study/go-fundamentals/level"

// Levels 返回「接口」主题的关卡：01_define / 02_implicit_impl / 03_polymorphism。
func Levels() []level.Level {
	return []level.Level{
		L08(),
		L09(),
		L10(),
	}
}
