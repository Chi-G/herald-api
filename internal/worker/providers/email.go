package providers

import (
	"context"
	"fmt"
	"log"

	"herald/internal/models"
)

type EmailProvider struct {
	// Adapter configuration (e.g. SMTP host/port, API key for Resend/SendGrid) can be added here
}

func NewEmailProvider() *EmailProvider {
	return &EmailProvider{}
}

func (p *EmailProvider) Name() string {
	return "email_smtp_adapter"
}

// Send simulates or dispatches an email notification.
func (p *EmailProvider) Send(ctx context.Context, n *models.Notification) error {
	if n.Recipient == "" {
		return fmt.Errorf("email recipient cannot be empty")
	}

	subject := "No Subject"
	if n.Subject != nil {
		subject = *n.Subject
	}

	// For dev/test verification log the dispatch
	log.Printf("[EmailProvider] Dispatched email to %s | Subject: '%s' | ID: %s", n.Recipient, subject, n.ID)
	return nil
}
