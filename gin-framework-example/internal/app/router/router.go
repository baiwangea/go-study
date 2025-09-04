package router

import (
	"gin-framework-example/internal/app/handler"
	"gin-framework-example/internal/app/middleware"

	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine) {
	r.Use(middleware.Logger())

	r.GET("/ping", handler.Ping)

	userRoutes := r.Group("/users")
	{
		userRoutes.POST("", handler.CreateUser)
		userRoutes.GET("/:id", handler.GetUser)
		userRoutes.PUT("/:id", handler.UpdateUser)
		userRoutes.DELETE("/:id", handler.DeleteUser)
	}
}
