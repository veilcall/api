package middleware

import "github.com/gin-gonic/gin"

// NoIPLogging must be the first middleware registered.
// It zeroes out all identifying connection metadata before any handler runs.
func NoIPLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.RemoteAddr = ""
		c.Request.Header.Del("X-Forwarded-For")
		c.Request.Header.Del("X-Real-IP")
		c.Request.Header.Del("X-Original-IP")
		c.Request.Header.Del("CF-Connecting-IP")
		c.Request.Header.Del("True-Client-IP")
		c.Next()
	}
}
