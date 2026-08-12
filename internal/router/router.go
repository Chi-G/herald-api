package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "herald/docs"
	"herald/internal/handlers"
	"herald/internal/middleware"
	"herald/internal/repository"
)

type Dependencies struct {
	NotificationHandler *handlers.NotificationHandler
	WebhookHandler      *handlers.WebhookHandler
	HealthHandler       *handlers.HealthHandler
	APIKeyRepo          *repository.APIKeyRepository
}

func New(deps Dependencies) *gin.Engine {
	r := gin.New()
	r.Use(middleware.CORS(), middleware.Logger(), middleware.Recovery())


	r.GET("/health", deps.HealthHandler.Check)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))


	v1 := r.Group("/api/v1")
	v1.Use(middleware.APIKeyAuth(deps.APIKeyRepo))
	v1.Use(middleware.RateLimit()) // per-tenant token bucket, reads tenant_id set by auth middleware
	{
		notifications := v1.Group("/notifications")
		{
			notifications.POST("", deps.NotificationHandler.Create)
			notifications.GET("", deps.NotificationHandler.List)
			notifications.GET("/:id", deps.NotificationHandler.Get)
		}

		webhooks := v1.Group("/webhooks")
		{
			webhooks.POST("", deps.WebhookHandler.Create)
			webhooks.GET("", deps.WebhookHandler.List)
			webhooks.DELETE("/:id", deps.WebhookHandler.Delete)
		}
	}

	return r
}
