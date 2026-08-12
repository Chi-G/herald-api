package handlers

import (
	"github.com/gin-gonic/gin"

	"herald/pkg/apierror"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Check(c *gin.Context) {
	apierror.Success(c, 200, gin.H{
		"status":  "healthy",
		"service": "herald-api",
	})
}
