package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"herald/internal/repository"
	"herald/pkg/apierror"
)

func APIKeyAuth(apiKeyRepo *repository.APIKeyRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			apierror.Unauthorized(c, "missing or malformed Authorization header")
			c.Abort()
			return
		}

		rawKey := strings.TrimPrefix(header, "Bearer ")
		hash := sha256.Sum256([]byte(rawKey))
		keyHash := hex.EncodeToString(hash[:])

		apiKey, err := apiKeyRepo.FindByHash(c.Request.Context(), keyHash)
		if err != nil || apiKey == nil || !apiKey.IsActive {
			apierror.Unauthorized(c, "invalid or revoked API key")
			c.Abort()
			return
		}

		c.Set("tenant_id", apiKey.TenantID)

		c.Next()
	}
}

func TenantIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get("tenant_id")
	if !exists {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}
