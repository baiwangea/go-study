// ── L1 阶段 · 指针与引用语义（关卡导航）────────────────────────
//
// 前置说明：完成 01-go-fundamentals 的 L1-01~L1-10。
// 本模块目标：6 关（L1-11 ~ L1-16）搞定 Go 最反直觉的三件事 ——
//
//	一切都是拷贝、nil 的真面目、切片/map 的共享与失联。
//
// 面向的读者：写过 PHP/Java/Python，习惯「对象传进去随便改」的人。
// 使用契约：一关一文件（11_xxx.go），每关只看 目标/观察/思考/通关 四行，
//
//	思考题动手改过代码再进下一关。
//
// 关卡分布：
//
//	L1-11 值传递与指针传递        L1-12 结构体拷贝与指针
//	L1-13 方法接收者选择          L1-14 切片共享与 map 引用
//	L1-15 nil 陷阱与接口 nil      L1-16 逃逸分析
package main

import (
	"os"

	"go-study/go-pointers/level"
)

func main() {
	levels := []level.Level{
		L11(), // 11_value_vs_pointer.go
		L12(), // 12_struct_copy.go
		L13(), // 13_receiver.go
		L14(), // 14_slice_map_shared.go
		L15(), // 15_nil_traps.go
		L16(), // 16_escape.go
	}

	level.Play("L1 指针与引用语义", levels, os.Args[1:])
}
