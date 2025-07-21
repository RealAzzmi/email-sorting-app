package usecases

import (
	"context"
	"fmt"

	"github.com/email-sorting-app/internal/domain/entities"
	"github.com/email-sorting-app/internal/domain/repositories"
)

type CategoryUsecase struct {
	categoryRepo repositories.CategoryRepository
	accountRepo  repositories.AccountRepository
	gmailService repositories.GmailService
}

func NewCategoryUsecase(
	categoryRepo repositories.CategoryRepository,
	accountRepo repositories.AccountRepository,
	gmailService repositories.GmailService,
) *CategoryUsecase {
	return &CategoryUsecase{
		categoryRepo: categoryRepo,
		accountRepo:  accountRepo,
		gmailService: gmailService,
	}
}

func (u *CategoryUsecase) GetAccountCategories(ctx context.Context, accountID int64) ([]entities.Category, error) {
	categories, err := u.categoryRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	return categories, nil
}

func (u *CategoryUsecase) CreateCategory(ctx context.Context, category *entities.Category) (*entities.Category, error) {
	// Check if category with same name already exists
	existing, err := u.categoryRepo.GetByName(ctx, category.AccountID, category.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing category: %w", err)
	}

	if existing != nil {
		return nil, fmt.Errorf("category with name '%s' already exists", category.Name)
	}

	// Get account to access OAuth token
	account, err := u.accountRepo.GetByID(ctx, category.AccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// Convert to OAuth2 token
	token := account.ToOAuth2Token()

	// Create the Gmail label first
	err = u.gmailService.CreateLabel(ctx, token, category.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail label: %w", err)
	}

	// Create the category in the database
	createdCategory, err := u.categoryRepo.Create(ctx, category)
	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return createdCategory, nil
}

func (u *CategoryUsecase) DeleteCategory(ctx context.Context, accountID, categoryID int64) error {
	// Check if category belongs to the account and get category details
	categories, err := u.categoryRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("failed to get categories: %w", err)
	}

	var categoryToDelete *entities.Category
	for _, cat := range categories {
		if cat.ID == categoryID {
			categoryToDelete = &cat
			break
		}
	}

	if categoryToDelete == nil {
		return fmt.Errorf("category not found or does not belong to account")
	}

	// Get account to access OAuth token
	account, err := u.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}

	// Convert to OAuth2 token
	token := account.ToOAuth2Token()

	// Check if this is a system label that shouldn't be deleted from Gmail
	systemLabels := map[string]bool{
		// Friendly names
		"Inbox": true, "Sent": true, "Drafts": true, "Spam": true, "Trash": true,
		"Important": true, "Starred": true, "All Mail": true, "Chats": true,
		// Gmail system label IDs
		"INBOX": true, "SENT": true, "DRAFT": true, "SPAM": true, "TRASH": true,
		"IMPORTANT": true, "STARRED": true, "UNREAD": true, "CHAT": true,
		// Star labels
		"YELLOW_STAR": true, "BLUE_STAR": true, "RED_STAR": true, "ORANGE_STAR": true,
		"GREEN_STAR": true, "PURPLE_STAR": true,
		// Category labels
		"CATEGORY_PERSONAL": true, "CATEGORY_SOCIAL": true, "CATEGORY_PROMOTIONS": true,
		"CATEGORY_UPDATES": true, "CATEGORY_FORUMS": true,
	}

	// Only try to delete from Gmail if it's not a system label
	if !systemLabels[categoryToDelete.Name] {
		// Try to delete the corresponding Gmail label
		err = u.gmailService.DeleteLabel(ctx, token, categoryToDelete.Name)
		if err != nil {
			// Log the error but don't fail the operation - the category might not exist in Gmail
			// or might have been deleted already
			fmt.Printf("Warning: failed to delete Gmail label '%s': %v\n", categoryToDelete.Name, err)
		}
	}

	// Delete the category from the database
	err = u.categoryRepo.Delete(ctx, categoryID)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}
