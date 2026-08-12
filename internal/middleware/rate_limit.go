package middleware

import (
	"github.com/gin-gonic/gin"
)

// RateLimit provides per-tenant token bucket rate limiting.
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Passthrough stub for initial setup
		c.Next()
	}
}
