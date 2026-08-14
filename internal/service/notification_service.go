package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"

	"herald/internal/models"
	"herald/internal/repository"
	"herald/internal/worker"
)

type NotificationService struct {
	repo       *repository.NotificationRepository
	workerPool *worker.Pool
}

func NewNotificationService(repo *repository.NotificationRepository, pool *worker.Pool) *NotificationService {
	return &NotificationService{
		repo:       repo,
		workerPool: pool,
	}
}

func (s *NotificationService) CreateAndQueue(
	ctx context.Context,
	tenantID uuid.UUID,
	channelStr string,
	recipient string,
	subject *string,
	body string,
	priorityStr string,
	metadata map[string]interface{},
) (*models.Notification, error) {
	if priorityStr == "" {
		priorityStr = "normal"
	}

	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	notification := &models.Notification{
		TenantID:    tenantID,
		Channel:     models.NotificationChannel(channelStr),
		Priority:    priorityStr,
		Status:      models.StatusPending,
		Recipient:   recipient,
		Subject:     subject,
		Body:        body,
		Metadata:    metadata,
		MaxAttempts: 5,
	}

	if err := s.repo.Create(ctx, notification); err != nil {
		return nil, fmt.Errorf("service create notification: %w", err)
	}

	if s.workerPool != nil {
		s.workerPool.Enqueue(worker.Job{NotificationID: notification.ID})
	}

	return notification, nil
}

func (s *NotificationService) GetForTenant(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*models.Notification, error) {
	return s.repo.FindByIDAndTenant(ctx, tenantID, id)
}

func (s *NotificationService) ListForTenant(ctx context.Context, tenantID uuid.UUID, status string, limitStr string, offsetStr string) ([]*models.Notification, error) {
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	return s.repo.ListForTenant(ctx, tenantID, status, limit, offset)
}
