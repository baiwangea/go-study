package middleware

import (
	"context"
	"gin-framework-example/internal/app/response"
	"gin-framework-example/pkg/db"
	"gin-framework-example/pkg/e"
	"gin-framework-example/pkg/util"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			response.Result(e.ERROR, "Authorization header is required", nil, c)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims, err := util.ParseToken(token)
		if err != nil {
			response.Result(e.ERROR, "Invalid token", nil, c)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Check if the token is in the Redis blacklist
		val, err := db.Rdb.Get(context.Background(), token).Result()
		if err == nil && val != "" {
			response.Result(e.ERROR, "Token is blacklisted", nil, c)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func BlacklistToken(token string, expiration time.Duration) error {
	return db.Rdb.Set(context.Background(), token, "blacklisted", expiration).Err()
}
