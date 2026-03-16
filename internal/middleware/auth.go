package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const sessionTTL = 24 * time.Hour

// Auth validates X-Session-Token header against Redis with sliding TTL.
func Auth(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-Session-Token")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing session token"})
			return
		}

		key := fmt.Sprintf("sess:%s", token)
		userID, err := rdb.GetEx(context.Background(), key, sessionTTL).Result()
		if err == redis.Nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired session"})
			return
		}
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "session lookup failed"})
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
