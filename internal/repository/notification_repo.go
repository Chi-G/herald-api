package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"herald/internal/models"
)

type NotificationRepository struct {
	db *pgxpool.Pool
}

func NewNotificationRepository(db *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, n *models.Notification) error {
	metadataBytes, err := json.Marshal(n.Metadata)
	if err != nil {
		return fmt.Errorf("marshal notification metadata: %w", err)
	}

	query := `
		INSERT INTO notifications (
			tenant_id, channel, priority, status, recipient, subject, body, metadata, scheduled_at, max_attempts
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
		RETURNING id, created_at, updated_at
	`
	err = r.db.QueryRow(
		ctx, query,
		n.TenantID, n.Channel, n.Priority, n.Status, n.Recipient,
		n.Subject, n.Body, metadataBytes, n.ScheduledAt, n.MaxAttempts,
	).Scan(&n.ID, &n.CreatedAt, &n.UpdatedAt)

	if err != nil {
		return fmt.Errorf("create notification: %w", err)
	}
	return nil
}

func (r *NotificationRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Notification, error) {
	query := `
		SELECT id, tenant_id, channel, priority, status, recipient, subject, body,
		       metadata, scheduled_at, sent_at, failed_at, attempt_count, max_attempts,
		       created_at, updated_at
		FROM notifications
		WHERE id = $1
	`
	return r.scanRow(r.db.QueryRow(ctx, query, id))
}

func (r *NotificationRepository) FindByIDAndTenant(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*models.Notification, error) {
	query := `
		SELECT id, tenant_id, channel, priority, status, recipient, subject, body,
		       metadata, scheduled_at, sent_at, failed_at, attempt_count, max_attempts,
		       created_at, updated_at
		FROM notifications
		WHERE id = $1 AND tenant_id = $2
	`
	return r.scanRow(r.db.QueryRow(ctx, query, id, tenantID))
}

func (r *NotificationRepository) ListForTenant(ctx context.Context, tenantID uuid.UUID, status string, limit int, offset int) ([]*models.Notification, error) {
	query := `
		SELECT id, tenant_id, channel, priority, status, recipient, subject, body,
		       metadata, scheduled_at, sent_at, failed_at, attempt_count, max_attempts,
		       created_at, updated_at
		FROM notifications
		WHERE tenant_id = $1 AND ($2 = '' OR status::text = $2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.Query(ctx, query, tenantID, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list notifications for tenant: %w", err)
	}
	defer rows.Close()

	notifications := make([]*models.Notification, 0)
	for rows.Next() {
		n, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, nil
}

func (r *NotificationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.NotificationStatus) error {
	now := time.Now()
	var sentAt, failedAt *time.Time

	if status == models.StatusSent {
		sentAt = &now
	} else if status == models.StatusFailed {
		failedAt = &now
	}

	query := `
		UPDATE notifications
		SET status = $1,
		    sent_at = COALESCE($2, sent_at),
		    failed_at = COALESCE($3, failed_at),
		    updated_at = now()
		WHERE id = $4
	`
	_, err := r.db.Exec(ctx, query, status, sentAt, failedAt, id)
	if err != nil {
		return fmt.Errorf("update notification status: %w", err)
	}
	return nil
}

func (r *NotificationRepository) RecordAttempt(ctx context.Context, notificationID uuid.UUID, attemptNumber int, attemptErr error, provider string) error {
	statusStr := "success"
	var errMsg *string
	if attemptErr != nil {
		statusStr = "failed"
		msg := attemptErr.Error()
		errMsg = &msg
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin attempt tx: %w", err)
	}
	defer tx.Rollback(ctx)

	insertQuery := `
		INSERT INTO notification_attempts (notification_id, attempt_number, status, provider, error_message)
		VALUES ($1, $2, $3, $4, $5)
	`
	if _, err := tx.Exec(ctx, insertQuery, notificationID, attemptNumber, statusStr, provider, errMsg); err != nil {
		return fmt.Errorf("insert notification attempt: %w", err)
	}

	updateQuery := `UPDATE notifications SET attempt_count = attempt_count + 1 WHERE id = $1`
	if _, err := tx.Exec(ctx, updateQuery, notificationID); err != nil {
		return fmt.Errorf("increment attempt count: %w", err)
	}

	return tx.Commit(ctx)
}

type rowScanner interface {
	Scan(dest ...any) error
}

func (r *NotificationRepository) scanRow(s rowScanner) (*models.Notification, error) {
	var n models.Notification
	var metadataBytes []byte

	err := s.Scan(
		&n.ID, &n.TenantID, &n.Channel, &n.Priority, &n.Status,
		&n.Recipient, &n.Subject, &n.Body, &metadataBytes,
		&n.ScheduledAt, &n.SentAt, &n.FailedAt,
		&n.AttemptCount, &n.MaxAttempts, &n.CreatedAt, &n.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scan notification row: %w", err)
	}

	if len(metadataBytes) > 0 {
		if err := json.Unmarshal(metadataBytes, &n.Metadata); err != nil {
			n.Metadata = make(map[string]interface{})
		}
	}
	return &n, nil
}
