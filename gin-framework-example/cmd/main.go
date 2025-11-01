package main

import (
	"fmt"
	"gin-framework-example/pkg/db"
	"log"
	"strconv"

	"gin-framework-example/internal/app/router"
	"gin-framework-example/internal/conf"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化配置，conf.Init 内部会处理错误
	conf.Init("internal/conf/config.yaml")
	// 初始化数据库，db.Init 内部会处理错误
	db.Init()
	// 初始化 Redis
	db.InitRedis()

	r := gin.Default()

	router.InitRouter(r)

	port := strconv.Itoa(conf.Conf.App.Port)
	fmt.Printf("Server is running on :%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
