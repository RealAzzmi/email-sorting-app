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

func (u *EmailUsecase) GetEmailByID(ctx context.Context, id int64) (*entities.Email, error) {
	return u.emailRepo.GetByID(ctx, id)
}