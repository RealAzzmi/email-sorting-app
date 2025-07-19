package repositories

import (
	"context"

	"github.com/email-sorting-app/internal/domain/entities"
)

type EmailRepository interface {
	GetByAccountID(ctx context.Context, accountID int64) ([]entities.Email, error)
	GetByID(ctx context.Context, id int64) (*entities.Email, error)
	Create(ctx context.Context, email *entities.Email) (*entities.Email, error)
	Update(ctx context.Context, email *entities.Email) error
	Delete(ctx context.Context, id int64) error
	ExistsByGmailMessageID(ctx context.Context, accountID int64, gmailMessageID string) (bool, error)
	BulkCreate(ctx context.Context, emails []entities.Email) error
}