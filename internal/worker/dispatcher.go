package worker

import (
	"context"
	"log"
	"math"
	"time"

	"herald/internal/models"
	"herald/internal/repository"
)

// Provider is implemented by every channel adapter (email, sms, push).
// Any new channel just needs to satisfy this interface — dispatcher.go never
// changes when you add a new provider, per the open/closed principle.
type Provider interface {
	Send(ctx context.Context, n *models.Notification) error
	Name() string
}

type Dispatcher struct {
	notificationRepo *repository.NotificationRepository
	providers        map[models.NotificationChannel]Provider
}

func NewDispatcher(repo *repository.NotificationRepository, providers map[models.NotificationChannel]Provider) *Dispatcher {
	return &Dispatcher{notificationRepo: repo, providers: providers}
}

// Dispatch sends one notification, retrying with exponential backoff on failure.
func (d *Dispatcher) Dispatch(ctx context.Context, job Job) {
	notification, err := d.notificationRepo.FindByID(ctx, job.NotificationID)
	if err != nil || notification == nil {
		log.Printf("dispatch: notification %s not found: %v", job.NotificationID, err)
		return
	}

	provider, ok := d.providers[notification.Channel]
	if !ok {
		log.Printf("dispatch: no provider registered for channel %s", notification.Channel)
		return
	}

	d.notificationRepo.UpdateStatus(ctx, notification.ID, models.StatusSending)

	for attempt := 1; attempt <= notification.MaxAttempts; attempt++ {
		err := provider.Send(ctx, notification)

		d.notificationRepo.RecordAttempt(ctx, notification.ID, attempt, err, provider.Name())

		if err == nil {
			d.notificationRepo.UpdateStatus(ctx, notification.ID, models.StatusSent)
			log.Printf("dispatch: notification %s sent via %s on attempt %d", notification.ID, provider.Name(), attempt)
			return
		}

		log.Printf("dispatch: notification %s failed attempt %d: %v", notification.ID, attempt, err)

		if attempt == notification.MaxAttempts {
			d.notificationRepo.UpdateStatus(ctx, notification.ID, models.StatusFailed)
			return
		}

		d.notificationRepo.UpdateStatus(ctx, notification.ID, models.StatusRetrying)

		// Exponential backoff: 2^attempt seconds (2s, 4s, 8s, 16s...)
		backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return
		}
	}
}
