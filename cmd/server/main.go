package main

import (
	"context"
	"log"

	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"herald/internal/config"
	"herald/internal/database"
	"herald/internal/handlers"
	"herald/internal/models"
	"herald/internal/repository"
	"herald/internal/router"
	"herald/internal/service"
	"herald/internal/worker"
	"herald/internal/worker/providers"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbPool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Printf("Warning: Failed to connect to Postgres pool: %v (continuing with degraded/mock pool if offline)", err)
	} else {
		defer dbPool.Close()
		log.Println("Successfully connected to PostgreSQL connection pool")
	}

	notificationRepo := repository.NewNotificationRepository(dbPool)
	apiKeyRepo := repository.NewAPIKeyRepository(dbPool)

	emailProvider := providers.NewEmailProvider()
	providerMap := map[models.NotificationChannel]worker.Provider{
		models.ChannelEmail: emailProvider,
	}

	dispatcher := worker.NewDispatcher(notificationRepo, providerMap)
	workerPool := worker.NewPool(cfg.WorkerCount, cfg.JobBuffer, dispatcher)

	workerCtx, cancelWorkers := context.WithCancel(context.Background())
	defer cancelWorkers()

	workerPool.Start(workerCtx)
	log.Printf("Started worker pool with %d workers", cfg.WorkerCount)

	notificationService := service.NewNotificationService(notificationRepo, workerPool)

	notificationHandler := handlers.NewNotificationHandler(notificationService)
	webhookHandler := handlers.NewWebhookHandler()
	healthHandler := handlers.NewHealthHandler()

	r := router.New(router.Dependencies{
		NotificationHandler: notificationHandler,
		WebhookHandler:      webhookHandler,
		HealthHandler:       healthHandler,
		APIKeyRepo:          apiKeyRepo,
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Herald HTTP server starting on port %s...", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Herald server gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server forced shutdown error: %v", err)
	}

	workerPool.Shutdown()
	log.Println("Herald server shut down cleanly.")
}
