package postgres

import (
	"context"
	"fmt"
	"math"

	"github.com/email-sorting-app/internal/domain/entities"
	"github.com/email-sorting-app/internal/domain/repositories"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EmailRepository struct {
	db *pgxpool.Pool
}

func NewEmailRepository(db *pgxpool.Pool) *EmailRepository {
	return &EmailRepository{db: db}
}

func (r *EmailRepository) GetByAccountID(ctx context.Context, accountID int64) ([]entities.Email, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, account_id, category_id, gmail_message_id, sender, subject, body, 
		       ai_summary, received_at, is_archived_in_gmail, unsubscribe_link, created_at, updated_at
		FROM emails 
		WHERE account_id = $1 
		ORDER BY received_at DESC
	`, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to query emails: %w", err)
	}
	defer rows.Close()

	var emails []entities.Email
	for rows.Next() {
		var email entities.Email
		err := rows.Scan(
			&email.ID, &email.AccountID, &email.CategoryID, &email.GmailMessageID,
			&email.Sender, &email.Subject, &email.Body, &email.AISummary,
			&email.ReceivedAt, &email.IsArchivedInGmail, &email.UnsubscribeLink,
			&email.CreatedAt, &email.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan email: %w", err)
		}
		emails = append(emails, email)
	}

	return emails, nil
}

func (r *EmailRepository) GetByID(ctx context.Context, id int64) (*entities.Email, error) {
	var email entities.Email
	err := r.db.QueryRow(ctx, `
		SELECT id, account_id, category_id, gmail_message_id, sender, subject, body, 
		       ai_summary, received_at, is_archived_in_gmail, unsubscribe_link, created_at, updated_at
		FROM emails WHERE id = $1
	`, id).Scan(
		&email.ID, &email.AccountID, &email.CategoryID, &email.GmailMessageID,
		&email.Sender, &email.Subject, &email.Body, &email.AISummary,
		&email.ReceivedAt, &email.IsArchivedInGmail, &email.UnsubscribeLink,
		&email.CreatedAt, &email.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("email not found")
		}
		return nil, fmt.Errorf("failed to get email: %w", err)
	}

	return &email, nil
}

func (r *EmailRepository) Create(ctx context.Context, email *entities.Email) (*entities.Email, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO emails (account_id, category_id, gmail_message_id, sender, subject, body, received_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`, email.AccountID, email.CategoryID, email.GmailMessageID, email.Sender, email.Subject, email.Body, email.ReceivedAt).Scan(
		&email.ID, &email.CreatedAt, &email.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create email: %w", err)
	}

	return email, nil
}

func (r *EmailRepository) Update(ctx context.Context, email *entities.Email) error {
	_, err := r.db.Exec(ctx, `
		UPDATE emails 
		SET category_id = $1, sender = $2, subject = $3, body = $4, ai_summary = $5, 
		    is_archived_in_gmail = $6, unsubscribe_link = $7, updated_at = NOW()
		WHERE id = $8
	`, email.CategoryID, email.Sender, email.Subject, email.Body, email.AISummary,
		email.IsArchivedInGmail, email.UnsubscribeLink, email.ID)
	if err != nil {
		return fmt.Errorf("failed to update email: %w", err)
	}

	return nil
}

func (r *EmailRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, "DELETE FROM emails WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete email: %w", err)
	}

	return nil
}

func (r *EmailRepository) ExistsByGmailMessageID(ctx context.Context, accountID int64, gmailMessageID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM emails WHERE account_id = $1 AND gmail_message_id = $2)
	`, accountID, gmailMessageID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return exists, nil
}

func (r *EmailRepository) BulkCreate(ctx context.Context, emails []entities.Email) error {
	if len(emails) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, email := range emails {
		batch.Queue(`
			INSERT INTO emails (account_id, category_id, gmail_message_id, sender, subject, body, received_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		`, email.AccountID, email.CategoryID, email.GmailMessageID, email.Sender, email.Subject, email.Body, email.ReceivedAt)
	}

	results := r.db.SendBatch(ctx, batch)
	defer results.Close()

	for i := 0; i < len(emails); i++ {
		_, err := results.Exec()
		if err != nil {
			return fmt.Errorf("failed to execute batch insert at index %d: %w", i, err)
		}
	}

	return nil
}

func (r *EmailRepository) GetByAccountIDPaginated(ctx context.Context, accountID int64, params repositories.PaginationParams) (*repositories.PaginatedEmails, error) {
	// Get total count
	var totalCount int64
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM emails WHERE account_id = $1
	`, accountID).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Calculate offset
	offset := (params.Page - 1) * params.PageSize
	
	// Get paginated emails
	rows, err := r.db.Query(ctx, `
		SELECT id, account_id, category_id, gmail_message_id, sender, subject, body, 
		       ai_summary, received_at, is_archived_in_gmail, unsubscribe_link, created_at, updated_at
		FROM emails 
		WHERE account_id = $1 
		ORDER BY received_at DESC
		LIMIT $2 OFFSET $3
	`, accountID, params.PageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query paginated emails: %w", err)
	}
	defer rows.Close()

	var emails []entities.Email
	for rows.Next() {
		var email entities.Email
		err := rows.Scan(
			&email.ID, &email.AccountID, &email.CategoryID, &email.GmailMessageID,
			&email.Sender, &email.Subject, &email.Body, &email.AISummary,
			&email.ReceivedAt, &email.IsArchivedInGmail, &email.UnsubscribeLink,
			&email.CreatedAt, &email.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan email: %w", err)
		}
		emails = append(emails, email)
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(params.PageSize)))

	return &repositories.PaginatedEmails{
		Emails:     emails,
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *EmailRepository) DeleteByAccountID(ctx context.Context, accountID int64) error {
	_, err := r.db.Exec(ctx, "DELETE FROM emails WHERE account_id = $1", accountID)
	if err != nil {
		return fmt.Errorf("failed to delete emails by account ID: %w", err)
	}

	return nil
}

func (r *EmailRepository) UpdateCategoryByGmailMessageID(ctx context.Context, accountID int64, gmailMessageID string, categoryID *int64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE emails 
		SET category_id = $1, updated_at = NOW()
		WHERE account_id = $2 AND gmail_message_id = $3
	`, categoryID, accountID, gmailMessageID)
	if err != nil {
		return fmt.Errorf("failed to update email category by Gmail message ID: %w", err)
	}

	return nil
}