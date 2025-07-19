package usecases

import (
	"context"
	"fmt"

	"github.com/email-sorting-app/internal/domain/entities"
	"github.com/email-sorting-app/internal/domain/repositories"
)

type EmailUsecase struct {
	emailRepo     repositories.EmailRepository
	accountRepo   repositories.AccountRepository
	categoryRepo  repositories.CategoryRepository
	gmailService  repositories.GmailService
}

func NewEmailUsecase(
	emailRepo repositories.EmailRepository,
	accountRepo repositories.AccountRepository,
	categoryRepo repositories.CategoryRepository,
	gmailService repositories.GmailService,
) *EmailUsecase {
	return &EmailUsecase{
		emailRepo:    emailRepo,
		accountRepo:  accountRepo,
		categoryRepo: categoryRepo,
		gmailService: gmailService,
	}
}

func (u *EmailUsecase) GetAccountEmails(ctx context.Context, accountID int64) ([]entities.Email, error) {
	// First, check if account exists
	account, err := u.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}

	// Get emails from database
	emails, err := u.emailRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get emails from database: %w", err)
	}

	// If no emails in database, sync from Gmail
	if len(emails) == 0 {
		err = u.syncEmailsFromGmail(ctx, account)
		if err != nil {
			return nil, fmt.Errorf("failed to sync emails from Gmail: %w", err)
		}

		// Get emails again after sync
		emails, err = u.emailRepo.GetByAccountID(ctx, accountID)
		if err != nil {
			return nil, fmt.Errorf("failed to get emails after sync: %w", err)
		}
	}

	return emails, nil
}

func (u *EmailUsecase) GetAccountEmailsPaginated(ctx context.Context, accountID int64, params repositories.PaginationParams) (*repositories.PaginatedEmails, error) {
	// First, check if account exists
	_, err := u.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}

	// Get paginated emails from database
	paginatedEmails, err := u.emailRepo.GetByAccountIDPaginated(ctx, accountID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get paginated emails: %w", err)
	}

	return paginatedEmails, nil
}

func (u *EmailUsecase) RefreshAccountEmails(ctx context.Context, accountID int64) error {
	// First, check if account exists
	account, err := u.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("account not found: %w", err)
	}

	// Check if this is the first sync (no history ID)
	if account.LastSyncHistoryID == nil {
		// First time sync - do full sync and get initial history ID
		return u.initialSyncAccountEmails(ctx, account)
	}

	// Use incremental sync
	return u.incrementalSyncAccountEmails(ctx, account)
}

func (u *EmailUsecase) syncEmailsFromGmail(ctx context.Context, account *entities.Account) error {
	token := account.ToOAuth2Token()

	// Get messages from Gmail
	gmailMessages, err := u.gmailService.ListMessages(ctx, token, 50)
	if err != nil {
		return fmt.Errorf("failed to list Gmail messages: %w", err)
	}

	var emailsToCreate []entities.Email

	for _, gmailMsg := range gmailMessages {
		// Check if email already exists
		exists, err := u.emailRepo.ExistsByGmailMessageID(ctx, account.ID, gmailMsg.ID)
		if err != nil {
			continue // Skip on error
		}

		if !exists {
			email := entities.Email{
				AccountID:      account.ID,
				GmailMessageID: gmailMsg.ID,
				Sender:         gmailMsg.Sender,
				Subject:        gmailMsg.Subject,
				Body:           gmailMsg.Body,
				ReceivedAt:     gmailMsg.ReceivedAt,
			}
			emailsToCreate = append(emailsToCreate, email)
		}
	}

	if len(emailsToCreate) > 0 {
		err = u.emailRepo.BulkCreate(ctx, emailsToCreate)
		if err != nil {
			return fmt.Errorf("failed to bulk create emails: %w", err)
		}
	}

	return nil
}

func (u *EmailUsecase) initialSyncAccountEmails(ctx context.Context, account *entities.Account) error {
	token := account.ToOAuth2Token()

	// Delete existing emails for clean start
	err := u.emailRepo.DeleteByAccountID(ctx, account.ID)
	if err != nil {
		return fmt.Errorf("failed to delete existing emails: %w", err)
	}

	// Get all messages from Gmail using pagination
	gmailMessages, err := u.gmailService.ListAllMessages(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to list all Gmail messages: %w", err)
	}

	var emailsToCreate []entities.Email

	// Get all unique label IDs from all messages
	var allLabelIds []string
	labelIdSet := make(map[string]bool)
	for _, gmailMsg := range gmailMessages {
		for _, labelId := range gmailMsg.Labels {
			if !labelIdSet[labelId] {
				labelIdSet[labelId] = true
				allLabelIds = append(allLabelIds, labelId)
			}
		}
	}

	// Resolve label IDs to names once for all messages
	token = account.ToOAuth2Token()
	labelNames, err := u.gmailService.GetLabelNames(ctx, token, allLabelIds)
	if err != nil {
		fmt.Printf("Warning: failed to get label names: %v\n", err)
		// Continue with IDs as fallback
		labelNames = make(map[string]string)
		for _, id := range allLabelIds {
			labelNames[id] = id
		}
	}

	for _, gmailMsg := range gmailMessages {
		// Convert label IDs to names
		var labelNamesForMsg []string
		for _, labelId := range gmailMsg.Labels {
			if name, exists := labelNames[labelId]; exists {
				labelNamesForMsg = append(labelNamesForMsg, name)
			} else {
				labelNamesForMsg = append(labelNamesForMsg, labelId)
			}
		}

		// Debug logging for labels
		fmt.Printf("Processing message %s with label IDs: %v, label names: %v\n", gmailMsg.ID, gmailMsg.Labels, labelNamesForMsg)
		
		// Determine category from Gmail label names (not IDs)
		categoryID, err := u.getCategoryFromLabels(ctx, account.ID, labelNamesForMsg)
		if err != nil {
			// Log error but continue processing
			fmt.Printf("Warning: failed to get category for message %s: %v\n", gmailMsg.ID, err)
		}
		
		if categoryID != nil {
			fmt.Printf("Message %s assigned to category ID: %d\n", gmailMsg.ID, *categoryID)
		} else {
			fmt.Printf("Message %s has no category assigned\n", gmailMsg.ID)
		}

		email := entities.Email{
			AccountID:      account.ID,
			CategoryID:     categoryID,
			GmailMessageID: gmailMsg.ID,
			Sender:         gmailMsg.Sender,
			Subject:        gmailMsg.Subject,
			Body:           gmailMsg.Body,
			ReceivedAt:     gmailMsg.ReceivedAt,
		}
		emailsToCreate = append(emailsToCreate, email)
	}

	if len(emailsToCreate) > 0 {
		err = u.emailRepo.BulkCreate(ctx, emailsToCreate)
		if err != nil {
			return fmt.Errorf("failed to bulk create emails: %w", err)
		}
	}

	// Get and store the current history ID
	historyID, err := u.gmailService.GetCurrentHistoryId(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to get current history ID: %w", err)
	}

	err = u.accountRepo.UpdateLastSyncHistoryID(ctx, account.ID, historyID)
	if err != nil {
		return fmt.Errorf("failed to update history ID: %w", err)
	}

	return nil
}

func (u *EmailUsecase) incrementalSyncAccountEmails(ctx context.Context, account *entities.Account) error {
	token := account.ToOAuth2Token()

	// Get new messages since last sync using History API
	newMessages, latestHistoryID, err := u.gmailService.ListHistory(ctx, token, *account.LastSyncHistoryID)
	if err != nil {
		return fmt.Errorf("failed to get history: %w", err)
	}

	// Get all unique label IDs from new messages
	var allLabelIds []string
	labelIdSet := make(map[string]bool)
	for _, gmailMsg := range newMessages {
		for _, labelId := range gmailMsg.Labels {
			if !labelIdSet[labelId] {
				labelIdSet[labelId] = true
				allLabelIds = append(allLabelIds, labelId)
			}
		}
	}

	// Resolve label IDs to names once for all new messages
	labelNames, err := u.gmailService.GetLabelNames(ctx, token, allLabelIds)
	if err != nil {
		fmt.Printf("Warning: failed to get label names for incremental sync: %v\n", err)
		// Continue with IDs as fallback
		labelNames = make(map[string]string)
		for _, id := range allLabelIds {
			labelNames[id] = id
		}
	}

	// Process new messages
	var emailsToCreate []entities.Email
	for _, gmailMsg := range newMessages {
		// Check if email already exists to avoid duplicates
		exists, err := u.emailRepo.ExistsByGmailMessageID(ctx, account.ID, gmailMsg.ID)
		if err != nil {
			continue // Skip on error
		}

		// Convert label IDs to names
		var labelNamesForMsg []string
		for _, labelId := range gmailMsg.Labels {
			if name, exists := labelNames[labelId]; exists {
				labelNamesForMsg = append(labelNamesForMsg, name)
			} else {
				labelNamesForMsg = append(labelNamesForMsg, labelId)
			}
		}

		// Determine category from Gmail label names (not IDs)
		categoryID, err := u.getCategoryFromLabels(ctx, account.ID, labelNamesForMsg)
		if err != nil {
			// Log error but continue processing
			fmt.Printf("Warning: failed to get category for message %s: %v\n", gmailMsg.ID, err)
		}

		if !exists {
			// Create new email
			email := entities.Email{
				AccountID:      account.ID,
				CategoryID:     categoryID,
				GmailMessageID: gmailMsg.ID,
				Sender:         gmailMsg.Sender,
				Subject:        gmailMsg.Subject,
				Body:           gmailMsg.Body,
				ReceivedAt:     gmailMsg.ReceivedAt,
			}
			emailsToCreate = append(emailsToCreate, email)
		} else {
			// Update existing email's category (label changed)
			fmt.Printf("Updating category for existing email %s\n", gmailMsg.ID)
			err := u.updateEmailCategory(ctx, account.ID, gmailMsg.ID, categoryID)
			if err != nil {
				fmt.Printf("Warning: failed to update category for message %s: %v\n", gmailMsg.ID, err)
			}
		}
	}

	// Bulk create new emails
	if len(emailsToCreate) > 0 {
		err = u.emailRepo.BulkCreate(ctx, emailsToCreate)
		if err != nil {
			return fmt.Errorf("failed to bulk create new emails: %w", err)
		}
	}

	// Update the history ID
	err = u.accountRepo.UpdateLastSyncHistoryID(ctx, account.ID, latestHistoryID)
	if err != nil {
		return fmt.Errorf("failed to update history ID: %w", err)
	}

	return nil
}

func (u *EmailUsecase) GetEmailByID(ctx context.Context, id int64) (*entities.Email, error) {
	return u.emailRepo.GetByID(ctx, id)
}

// getCategoryFromLabels maps Gmail labels to app categories
func (u *EmailUsecase) getCategoryFromLabels(ctx context.Context, accountID int64, labels []string) (*int64, error) {
	fmt.Printf("getCategoryFromLabels called with labels: %v\n", labels)
	
	// Skip if no labels
	if len(labels) == 0 {
		fmt.Printf("No labels found, returning nil\n")
		return nil, nil
	}

	// Priority order for selecting the primary category
	var categoryName string
	
	// First check for custom labels (non-system)
	for _, label := range labels {
		if !u.isSystemLabel(label) {
			categoryName = label
			break
		}
	}
	
	// If no custom labels, use the most meaningful system label
	if categoryName == "" {
		for _, label := range labels {
			switch label {
			case "SENT":
				categoryName = "Sent"
			case "DRAFT":
				categoryName = "Drafts"
			case "SPAM":
				categoryName = "Spam"
			case "TRASH":
				categoryName = "Trash"
			case "STARRED":
				categoryName = "Starred"
			case "IMPORTANT":
				categoryName = "Important"
			case "INBOX":
				categoryName = "Inbox"
			}
			if categoryName != "" {
				break
			}
		}
	}

	// If still no category found, return nil
	if categoryName == "" {
		fmt.Printf("No suitable category name found from labels\n")
		return nil, nil
	}

	fmt.Printf("Selected category name: %s\n", categoryName)

	// Get or create category
	category, err := u.categoryRepo.GetOrCreate(ctx, accountID, categoryName)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create category: %w", err)
	}

	fmt.Printf("Category created/found with ID: %d\n", category.ID)
	return &category.ID, nil
}

// isSystemLabel checks if a label is a Gmail system label or auto-category
func (u *EmailUsecase) isSystemLabel(label string) bool {
	systemLabels := []string{
		"INBOX", "SENT", "DRAFT", "SPAM", "TRASH", 
		"UNREAD", "STARRED", "IMPORTANT", "CHAT",
		// Gmail auto-categories (exclude these as they're not user-created labels)
		"CATEGORY_PERSONAL", "CATEGORY_SOCIAL", "CATEGORY_PROMOTIONS", 
		"CATEGORY_UPDATES", "CATEGORY_FORUMS",
	}
	
	for _, sysLabel := range systemLabels {
		if label == sysLabel {
			return true
		}
	}
	
	return false
}

// updateEmailCategory updates the category of an existing email
func (u *EmailUsecase) updateEmailCategory(ctx context.Context, accountID int64, gmailMessageID string, categoryID *int64) error {
	return u.emailRepo.UpdateCategoryByGmailMessageID(ctx, accountID, gmailMessageID, categoryID)
}