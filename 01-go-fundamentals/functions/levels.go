package functions

import "go-study/go-fundamentals/level"

// Levels 返回本主题的关卡，顺序即通关顺序。
// 每关一个文件：01_basic_func / 02_multi_return / 03_variadic / 04_closure / 05_recursion
func Levels() []level.Level {
	return []level.Level{
		L01(),
		L02(),
		L03(),
		L04(),
		L05(),
	}
}
