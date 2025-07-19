package usecases

import (
	"context"

	"github.com/email-sorting-app/internal/domain/entities"
	"github.com/email-sorting-app/internal/domain/repositories"
)

type AccountUsecase struct {
	accountRepo repositories.AccountRepository
}

func NewAccountUsecase(accountRepo repositories.AccountRepository) *AccountUsecase {
	return &AccountUsecase{
		accountRepo: accountRepo,
	}
}

func (u *AccountUsecase) GetAllAccounts(ctx context.Context) ([]entities.Account, error) {
	return u.accountRepo.GetAll(ctx)
}

func (u *AccountUsecase) GetAccountByID(ctx context.Context, id int64) (*entities.Account, error) {
	return u.accountRepo.GetByID(ctx, id)
}

func (u *AccountUsecase) DeleteAccount(ctx context.Context, id int64) error {
	return u.accountRepo.Delete(ctx, id)
}