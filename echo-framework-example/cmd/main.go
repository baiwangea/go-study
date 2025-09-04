package main

import (
	"go-study/echo-framework-example/internal/app/router"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Initialize routes
	router.InitRouter(e)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}
