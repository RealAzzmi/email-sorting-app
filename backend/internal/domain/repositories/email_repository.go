package repositories

import (
	"context"

	"github.com/email-sorting-app/internal/domain/entities"
)

type PaginationParams struct {
	Page     int
	PageSize int
}

type PaginatedEmails struct {
	Emails     []entities.Email `json:"emails"`
	TotalCount int64            `json:"total_count"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

type EmailRepository interface {
	GetByAccountID(ctx context.Context, accountID int64) ([]entities.Email, error)
	GetByAccountIDPaginated(ctx context.Context, accountID int64, params PaginationParams) (*PaginatedEmails, error)
	GetByID(ctx context.Context, id int64) (*entities.Email, error)
	Create(ctx context.Context, email *entities.Email) (*entities.Email, error)
	Update(ctx context.Context, email *entities.Email) error
	Delete(ctx context.Context, id int64) error
	DeleteByAccountID(ctx context.Context, accountID int64) error
	ExistsByGmailMessageID(ctx context.Context, accountID int64, gmailMessageID string) (bool, error)
	BulkCreate(ctx context.Context, emails []entities.Email) error
	UpdateCategoryByGmailMessageID(ctx context.Context, accountID int64, gmailMessageID string, categoryID *int64) error
	GetByCategoryIDPaginated(ctx context.Context, accountID, categoryID int64, params PaginationParams) (*PaginatedEmails, error)
}