package mutexes

import "go-study/go-concurrency/level"

// Levels 返回「共享内存保护」主题的关卡：01_race。
func Levels() []level.Level {
	return []level.Level{L13()}
}
