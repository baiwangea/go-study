package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Ping handles the /ping health check route.
func Ping(c echo.Context) error {
	return c.String(http.StatusOK, "pong")
}
