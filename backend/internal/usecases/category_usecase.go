package usecases

import (
	"context"
	"fmt"

	"github.com/email-sorting-app/internal/domain/entities"
	"github.com/email-sorting-app/internal/domain/repositories"
)

type CategoryUsecase struct {
	categoryRepo repositories.CategoryRepository
}

func NewCategoryUsecase(categoryRepo repositories.CategoryRepository) *CategoryUsecase {
	return &CategoryUsecase{
		categoryRepo: categoryRepo,
	}
}

func (u *CategoryUsecase) GetAccountCategories(ctx context.Context, accountID int64) ([]entities.Category, error) {
	categories, err := u.categoryRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	return categories, nil
}