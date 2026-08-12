package handlers

import (
	"github.com/gin-gonic/gin"

	"herald/pkg/apierror"
)

type WebhookHandler struct{}

func NewWebhookHandler() *WebhookHandler {
	return &WebhookHandler{}
}

func (h *WebhookHandler) Create(c *gin.Context) {
	apierror.Success(c, 201, gin.H{"message": "webhook creation endpoint ready"})
}

func (h *WebhookHandler) List(c *gin.Context) {
	apierror.Success(c, 200, []gin.H{})
}

func (h *WebhookHandler) Delete(c *gin.Context) {
	apierror.Success(c, 200, gin.H{"message": "webhook deletion endpoint ready"})
}
