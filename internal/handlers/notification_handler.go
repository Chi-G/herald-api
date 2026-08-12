package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"herald/internal/middleware"
	"herald/internal/service"
	"herald/pkg/apierror"
)

type NotificationHandler struct {
	service *service.NotificationService
}

func NewNotificationHandler(s *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: s}
}

// POST /api/v1/notifications
func (h *NotificationHandler) Create(c *gin.Context) {
	tenantID, ok := middleware.TenantIDFromContext(c)
	if !ok {
		apierror.Unauthorized(c, "tenant context missing")
		return
	}

	var req struct {
		Channel     string                 `json:"channel" binding:"required,oneof=email sms push"`
		Recipient   string                 `json:"recipient" binding:"required"`
		Subject     *string                `json:"subject"`
		Body        string                 `json:"body" binding:"required"`
		Priority    string                 `json:"priority" binding:"omitempty,oneof=low normal high urgent"`
		Metadata    map[string]interface{} `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.BadRequest(c, err.Error())
		return
	}

	notification, err := h.service.CreateAndQueue(c.Request.Context(), tenantID, req.Channel, req.Recipient, req.Subject, req.Body, req.Priority, req.Metadata)
	if err != nil {
		apierror.Internal(c, "failed to create notification")
		return
	}

	apierror.Success(c, 201, notification)
}

// GET /api/v1/notifications/:id
func (h *NotificationHandler) Get(c *gin.Context) {
	tenantID, ok := middleware.TenantIDFromContext(c)
	if !ok {
		apierror.Unauthorized(c, "tenant context missing")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apierror.BadRequest(c, "invalid notification id")
		return
	}

	notification, err := h.service.GetForTenant(c.Request.Context(), tenantID, id)
	if err != nil || notification == nil {
		apierror.NotFound(c, "notification not found")
		return
	}

	apierror.Success(c, 200, notification)
}

// GET /api/v1/notifications  (list, paginated, filterable by status)
func (h *NotificationHandler) List(c *gin.Context) {
	tenantID, ok := middleware.TenantIDFromContext(c)
	if !ok {
		apierror.Unauthorized(c, "tenant context missing")
		return
	}

	status := c.Query("status") // optional filter
	limit := c.DefaultQuery("limit", "20")
	offset := c.DefaultQuery("offset", "0")

	notifications, err := h.service.ListForTenant(c.Request.Context(), tenantID, status, limit, offset)
	if err != nil {
		apierror.Internal(c, "failed to list notifications")
		return
	}

	apierror.Success(c, 200, notifications)
}
