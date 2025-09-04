package router

import (
	"go-study/echo-framework-example/internal/app/handler"

	"github.com/labstack/echo/v4"
)

// InitRouter initializes the routes for the application.
func InitRouter(e *echo.Echo) {
	// Simple health check route
	e.GET("/ping", handler.Ping)
}
