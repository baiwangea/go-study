package packages

import "go-study/go-fundamentals/level"

// Levels 返回「包」主题的关卡：01_export / 02_import。
func Levels() []level.Level {
	return []level.Level{
		L06(),
		L07(),
	}
}
