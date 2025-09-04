package main

import (
	"go-study/echo-framework-example/internal/app/router"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// CustomValidator holds the validator instance.
type CustomValidator struct {
	validator *validator.Validate
}

// Validate is the implementation of the echo.Validator interface.
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func main() {
	// Echo instance
	e := echo.New()

	// Register the custom validator
	e.Validator = &CustomValidator{validator: validator.New()}

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Initialize routes
	router.InitRouter(e)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}
