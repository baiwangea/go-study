package router

import (
	"gin-framework-example/internal/app/handler"
	"gin-framework-example/internal/app/middleware"
	"gin-framework-example/pkg/util"
	"time"

	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine) {
	r.Use(middleware.Logger())

	r.GET("/ping", handler.Ping)

	// Captcha and auth routes
	r.GET("/captcha", middleware.GenerateCaptcha)
	r.POST("/login", handler.Login)

	auth := r.Group("/auth")
	auth.Use(middleware.JWT())
	{
		auth.POST("/logout", func(c *gin.Context) {
			token := c.GetHeader("Authorization")
			claims, _ := util.ParseToken(token)
			expiration := time.Until(time.Unix(claims.ExpiresAt, 0))
			middleware.BlacklistToken(token, expiration)
			c.JSON(200, gin.H{"message": "Logged out successfully"})
		})
	}

	userRoutes := r.Group("/users")
	//userRoutes.Use(middleware.JWT()) // Protect all user routes
	{
		userRoutes.POST("", middleware.VerifyCaptcha(), handler.CreateUser)
		userRoutes.GET("/:id", handler.GetUser)
		userRoutes.PUT("/:id", handler.UpdateUser)
		userRoutes.DELETE("/:id", handler.DeleteUser)
	}
}
