package middleware

import (
	"log"

	"github.com/gin-gonic/gin"

	"herald/pkg/apierror"
)

// Recovery recovers from panics and converts them into standardized JSON 500 responses.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC RECOVERY] %v", err)
				apierror.Internal(c, "an unexpected internal server error occurred")
				c.Abort()
			}
		}()
		c.Next()
	}
}
