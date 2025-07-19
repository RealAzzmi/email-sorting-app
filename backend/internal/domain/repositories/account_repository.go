package repositories

import (
	"context"

	"github.com/email-sorting-app/internal/domain/entities"
	"golang.org/x/oauth2"
)

type AccountRepository interface {
	GetAll(ctx context.Context) ([]entities.Account, error)
	GetByID(ctx context.Context, id int64) (*entities.Account, error)
	GetByEmail(ctx context.Context, email string) (*entities.Account, error)
	Create(ctx context.Context, email, name string, token *oauth2.Token) (*entities.Account, error)
	Update(ctx context.Context, account *entities.Account) error
	Delete(ctx context.Context, id int64) error
}