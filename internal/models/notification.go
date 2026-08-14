package models

import (
	"time"

	"github.com/google/uuid"
)

type NotificationChannel string

const (
	ChannelEmail NotificationChannel = "email"
	ChannelSMS   NotificationChannel = "sms"
	ChannelPush  NotificationChannel = "push"
)

type NotificationStatus string

const (
	StatusPending   NotificationStatus = "pending"
	StatusQueued    NotificationStatus = "queued"
	StatusSending   NotificationStatus = "sending"
	StatusSent      NotificationStatus = "sent"
	StatusFailed    NotificationStatus = "failed"
	StatusRetrying  NotificationStatus = "retrying"
	StatusCancelled NotificationStatus = "cancelled"
)

type Notification struct {
	ID        uuid.UUID              `json:"id" db:"id"`
	TenantID  uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	Channel   NotificationChannel    `json:"channel" db:"channel"`
	Priority  string                 `json:"priority" db:"priority"`
	Status    NotificationStatus     `json:"status" db:"status"`

	Recipient string                 `json:"recipient" db:"recipient"`
	Subject   *string                `json:"subject,omitempty" db:"subject"`
	Body      string                 `json:"body" db:"body"`
	Metadata  map[string]interface{} `json:"metadata" db:"metadata"`

	ScheduledAt *time.Time `json:"scheduled_at,omitempty" db:"scheduled_at"`
	SentAt      *time.Time `json:"sent_at,omitempty" db:"sent_at"`
	FailedAt    *time.Time `json:"failed_at,omitempty" db:"failed_at"`

	AttemptCount int `json:"attempt_count" db:"attempt_count"`
	MaxAttempts  int `json:"max_attempts" db:"max_attempts"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}


type CreateNotificationRequest struct {
	Channel   NotificationChannel    `json:"channel" binding:"required,oneof=email sms push"`
	Recipient string                 `json:"recipient" binding:"required"`
	Subject   *string                `json:"subject"`
	Body      string                 `json:"body" binding:"required"`
	Priority  string                 `json:"priority" binding:"omitempty,oneof=low normal high urgent"`
	Metadata  map[string]interface{} `json:"metadata"`
	ScheduledAt *time.Time           `json:"scheduled_at"`
}
