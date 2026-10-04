// ── L3 阶段 · GoFrame 工程主线（关卡导航）──────────────────────
//
// 前置说明：完成 L1（语法与指针）与 L2（并发，尤其 L2-11 context 取消）。
// 本模块目标：从第一个 HTTP 服务开始，一路做到配置 / 日志 / 错误码 / MySQL / Redis /
// 定时任务 / 异步重试 / 鉴权 / 分层 / 部署。原先散落在各目录的 mysql、redis、jwt、
// logger、telegram、web3 示例，全部成为本框架下的关卡。
//
// 使用契约：一关一文件（NN_主题.go），每关只看 目标/观察/思考/通关 四行。
//
// 已实现：L3-01 ~ L3-12（共 12 关），加 L3-15 分层、L3-16 交叉编译，合计 14 关。
// 待实现：L3-13 TG 通知、L3-14 链上只读。
// L3-07/08 需要环境变量 BOT_DB_LINK，L3-09 需要本地 Redis；未配置时自动跳过，不报错。
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
		L07(), // 07_db_mysql.go       需 BOT_DB_LINK
		L08(), // 08_orm_model.go      需 BOT_DB_LINK
		L09(), // 09_cache_redis.go    需本地 Redis
		L10(), // 10_cron_task.go
		L11(), // 11_async_queue.go
		L12(), // 12_jwt_auth.go
		L15(), // 15_project_layout.go
		L16(), // 16_deploy_build.go（注意：文件名不能叫 *_linux.go）
	}

	level.Play("L3 GoFrame 主线", levels, os.Args[1:])
}
