package router

import (
	"go-study/echo-framework-example/internal/app/handler"

	"github.com/labstack/echo/v4"
)

// InitRouter initializes the routes for the application.
func InitRouter(e *echo.Echo) {
	// Simple health check route
	e.GET("/ping", handler.Ping)

	// User routes - CRUD
	e.GET("/users/:id", handler.GetUser)
	e.GET("/users/muilt", handler.GetMuiltUser)
	e.POST("/users", handler.CreateUser)
	e.PUT("/users/:id", handler.UpdateUser)
	e.DELETE("/users/:id", handler.DeleteUser)
}
