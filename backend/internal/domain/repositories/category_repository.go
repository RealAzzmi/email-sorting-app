package repositories

import (
	"context"

	"github.com/email-sorting-app/internal/domain/entities"
)

type CategoryRepository interface {
	GetByAccountID(ctx context.Context, accountID int64) ([]entities.Category, error)
	GetByName(ctx context.Context, accountID int64, name string) (*entities.Category, error)
	Create(ctx context.Context, category *entities.Category) (*entities.Category, error)
	Delete(ctx context.Context, categoryID int64) error
	GetOrCreate(ctx context.Context, accountID int64, name string) (*entities.Category, error)
}
