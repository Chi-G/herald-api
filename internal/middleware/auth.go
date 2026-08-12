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

// APIKeyAuth validates the `Authorization: Bearer hrld_live_xxx` header against
// api_keys.key_hash, and sets tenant_id in the Gin context for downstream handlers.
//
// This is a "middleware factory": it takes the repo as a dependency and returns
// the actual gin.HandlerFunc. This is how you inject dependencies into Gin
// middleware without relying on globals.
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

		// Stash tenant_id in the context — every handler downstream reads this,
		// it's the single source of truth for "which tenant is this request for."
		c.Set("tenant_id", apiKey.TenantID)

		c.Next()
	}
}

// TenantIDFromContext is a small helper so handlers don't repeat the type assertion.
func TenantIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get("tenant_id")
	if !exists {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}
