package usecases

import (
	"context"
	"fmt"

	"github.com/email-sorting-app/internal/domain/entities"
	"github.com/email-sorting-app/internal/domain/repositories"
)

type EmailUsecase struct {
	emailRepo    repositories.EmailRepository
	accountRepo  repositories.AccountRepository
	gmailService repositories.GmailService
}

func NewEmailUsecase(
	emailRepo repositories.EmailRepository,
	accountRepo repositories.AccountRepository,
	gmailService repositories.GmailService,
) *EmailUsecase {
	return &EmailUsecase{
		emailRepo:    emailRepo,
		accountRepo:  accountRepo,
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

	// Delete existing emails for this account
	err = u.emailRepo.DeleteByAccountID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("failed to delete existing emails: %w", err)
	}

	// Sync all emails from Gmail
	err = u.syncAllEmailsFromGmail(ctx, account)
	if err != nil {
		return fmt.Errorf("failed to sync emails from Gmail: %w", err)
	}

	return nil
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

func (u *EmailUsecase) syncAllEmailsFromGmail(ctx context.Context, account *entities.Account) error {
	token := account.ToOAuth2Token()

	// Get all messages from Gmail using pagination
	gmailMessages, err := u.gmailService.ListAllMessages(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to list all Gmail messages: %w", err)
	}

	var emailsToCreate []entities.Email

	for _, gmailMsg := range gmailMessages {
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

	if len(emailsToCreate) > 0 {
		err = u.emailRepo.BulkCreate(ctx, emailsToCreate)
		if err != nil {
			return fmt.Errorf("failed to bulk create emails: %w", err)
		}
	}

	return nil
}

func (u *EmailUsecase) GetEmailByID(ctx context.Context, id int64) (*entities.Email, error) {
	return u.emailRepo.GetByID(ctx, id)
}