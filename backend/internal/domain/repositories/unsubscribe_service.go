package repositories

import (
	"context"

	"github.com/email-sorting-app/internal/domain/entities"
)

type UnsubscribeResult struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	ErrorType    string `json:"error_type,omitempty"`
	RequiresAuth bool   `json:"requires_auth,omitempty"`
}

type UnsubscribeService interface {
	// UnsubscribeFromEmail attempts to unsubscribe from a single email using its unsubscribe link
	UnsubscribeFromEmail(ctx context.Context, email *entities.Email) (*UnsubscribeResult, error)
	
	// BulkUnsubscribe attempts to unsubscribe from multiple emails
	BulkUnsubscribe(ctx context.Context, emails []*entities.Email) (map[int64]*UnsubscribeResult, error)
	
	// ValidateUnsubscribeLink checks if an unsubscribe link is valid and accessible
	ValidateUnsubscribeLink(ctx context.Context, link string) (bool, error)
}