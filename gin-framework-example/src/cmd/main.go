package main

import (
	"flag"
	"fmt"
	"gin-framework-example/src/app/router"
	"gin-framework-example/src/pkg/db"
	"gin-framework-example/src/pkg/util"
	"strconv"

	"github.com/gin-gonic/gin"
)

func main() {
	// 解析命令行参数
	env := flag.String("env", "dev", "set the runtime environment (dev, test, prod)")
	flag.Parse()

	// 初始化配置
	if err := util.Init(*env); err != nil {
		panic(err)
	}

	// 初始化日志
	util.InitLogger()
	// 初始化数据库
	db.Init()
	// 初始化 Redis
	db.InitRedis()

	r := gin.Default()

	router.InitRouter(r)

	port := strconv.Itoa(util.Conf.App.Port)
	fmt.Printf("Server is running on :%s\n", port)
	if err := r.Run(":" + port); err != nil {
		util.Error("Failed to run server: %v", err)
	}
}
