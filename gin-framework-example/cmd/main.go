package main

import (
	"fmt"
	"log"
	"strconv"

	"gin-framework-example/internal/app/router"
	"gin-framework-example/internal/conf"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := conf.Init(); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	r := gin.Default()

	router.InitRouter(r)

	port := strconv.Itoa(conf.Conf.App.Port)
	fmt.Printf("Server is running on :%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
