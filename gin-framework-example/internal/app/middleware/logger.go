package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()

		// before request

		c.Next()

		// after request
		latency := time.Since(t)
		log.Printf("Request: %s %s - %v", c.Request.Method, c.Request.URL.Path, latency)

		// access the status we are sending
		status := c.Writer.Status()
		log.Printf("Response status: %d", status)
	}
}
