// ── L3 阶段 · GoFrame 工程主线（关卡导航）──────────────────────
//
// 前置说明：完成 L1（语法与指针）与 L2（并发，尤其 L2-11 context 取消）。
// 本模块目标：从第一个 HTTP 服务开始，一路做到「配置 / 日志 / 错误码 / MySQL /
//
//	Redis / 定时任务 / 鉴权 / 通知 / 链上交互」—— 原先散落在各目录的
//	mysql、redis、jwt、logger、telegram、web3 示例，全部作为本框架下的关卡。
//
// 使用契约：一关一文件（NN_主题.go），每关只看 目标/观察/思考/通关 四行；
//
//	前 6 关不依赖任何外部服务（MySQL/Redis/RPC 需要服务的关卡见 README 规划）。
//
// 已实现：L3-01 路由 / L3-02 参数绑定校验 / L3-03 中间件 / L3-04 配置 / L3-05 日志 / L3-06 错误码
// 待实现：L3-07 ~ L3-16（g.DB、ORM、Redis、gcron、异步任务、JWT、TG 通知、Web3、工程结构、部署）
package main

import (
	"os"

	"go-study/goframe/level"
)

func main() {
	levels := []level.Level{
		L01(), // 01_hello_route.go
		L02(), // 02_params_binding.go
		L03(), // 03_middleware.go
		L04(), // 04_config.go
		L05(), // 05_logging.go
		L06(), // 06_error_handling.go
		L10(), // 10_cron_task.go
		L11(), // 11_async_queue.go
		L15(), // 15_project_layout.go
		L16(), // 16_deploy_linux.go
	}

	level.Play("L3 GoFrame 主线", levels, os.Args[1:])
}
