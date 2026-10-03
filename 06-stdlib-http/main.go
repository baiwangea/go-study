// ── L4 阶段 · 标准库 HTTP（关卡导航）──────────────────────────
//
// 前置说明：完成 L1（接口部分尤其重要）与 L2（context 取消）。
// 本模块目标：8 关把 net/http 的客户端与服务端写对 ——
//
//	超时、错误分类、Body 生命周期、中间件、取消传播、输入防护。
//
// 对应机器人能力：调 CEX REST 接口、写控制后台 API。
// 使用契约：一关一文件，每关只看 目标/观察/思考/通关；
//
//	所有关卡用 httptest 本地起服务，不依赖外网、不会阻塞。
//
// 关卡分布：
//
//	L4-01 ~ L4-04  client/  GET 与 Body、两层超时、JSON 与金额精度、状态码 vs error
//	L4-05 ~ L4-08  server/  http.Server 与优雅关闭、中间件接口、context 取消、输入防护
package main

import (
	"os"

	"go-study/stdlib-http/client"
	"go-study/stdlib-http/level"
	"go-study/stdlib-http/server"
)

func main() {
	var levels []level.Level
	levels = append(levels, client.Levels()...)
	levels = append(levels, server.Levels()...)

	level.Play("L4 标准库 HTTP", levels, os.Args[1:])
}
