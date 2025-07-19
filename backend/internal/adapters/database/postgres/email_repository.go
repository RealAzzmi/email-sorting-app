package postgres

import (
	"context"
	"fmt"
	"math"
	"strings"

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
		SELECT id, account_id, gmail_message_id, sender, subject, body, 
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
			&email.ID, &email.AccountID, &email.GmailMessageID,
			&email.Sender, &email.Subject, &email.Body, &email.AISummary,
			&email.ReceivedAt, &email.IsArchivedInGmail, &email.UnsubscribeLink,
			&email.CreatedAt, &email.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan email: %w", err)
		}
		
		// Load categories for this email
		categoryIDs, err := r.GetEmailCategories(ctx, email.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get email categories: %w", err)
		}
		email.CategoryIDs = categoryIDs
		
		emails = append(emails, email)
	}

	return emails, nil
}

func (r *EmailRepository) GetByID(ctx context.Context, id int64) (*entities.Email, error) {
	var email entities.Email
	err := r.db.QueryRow(ctx, `
		SELECT id, account_id, gmail_message_id, sender, subject, body, 
		       ai_summary, received_at, is_archived_in_gmail, unsubscribe_link, created_at, updated_at
		FROM emails WHERE id = $1
	`, id).Scan(
		&email.ID, &email.AccountID, &email.GmailMessageID,
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

	// Load categories for this email
	categoryIDs, err := r.GetEmailCategories(ctx, email.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get email categories: %w", err)
	}
	email.CategoryIDs = categoryIDs

	return &email, nil
}

func (r *EmailRepository) Create(ctx context.Context, email *entities.Email) (*entities.Email, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO emails (account_id, gmail_message_id, sender, subject, body, received_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`, email.AccountID, email.GmailMessageID, email.Sender, email.Subject, email.Body, email.ReceivedAt).Scan(
		&email.ID, &email.CreatedAt, &email.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create email: %w", err)
	}

	// Add email to categories if any
	if len(email.CategoryIDs) > 0 {
		for _, categoryID := range email.CategoryIDs {
			_, err = tx.Exec(ctx, `
				INSERT INTO email_categories (email_id, category_id, created_at)
				VALUES ($1, $2, NOW())
			`, email.ID, categoryID)
			if err != nil {
				return nil, fmt.Errorf("failed to add email to category: %w", err)
			}
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return email, nil
}

func (r *EmailRepository) Update(ctx context.Context, email *entities.Email) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		UPDATE emails 
		SET sender = $1, subject = $2, body = $3, ai_summary = $4, 
		    is_archived_in_gmail = $5, unsubscribe_link = $6, updated_at = NOW()
		WHERE id = $7
	`, email.Sender, email.Subject, email.Body, email.AISummary,
		email.IsArchivedInGmail, email.UnsubscribeLink, email.ID)
	if err != nil {
		return fmt.Errorf("failed to update email: %w", err)
	}

	// Update categories - remove all existing and add new ones
	_, err = tx.Exec(ctx, "DELETE FROM email_categories WHERE email_id = $1", email.ID)
	if err != nil {
		return fmt.Errorf("failed to remove existing categories: %w", err)
	}

	for _, categoryID := range email.CategoryIDs {
		_, err = tx.Exec(ctx, `
			INSERT INTO email_categories (email_id, category_id, created_at)
			VALUES ($1, $2, NOW())
		`, email.ID, categoryID)
		if err != nil {
			return fmt.Errorf("failed to add email to category: %w", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
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

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for i, email := range emails {
		var emailID int64
		err := tx.QueryRow(ctx, `
			INSERT INTO emails (account_id, gmail_message_id, sender, subject, body, received_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
			RETURNING id
		`, email.AccountID, email.GmailMessageID, email.Sender, email.Subject, email.Body, email.ReceivedAt).Scan(&emailID)
		if err != nil {
			return fmt.Errorf("failed to insert email at index %d: %w", i, err)
		}

		// Add email to categories
		for _, categoryID := range email.CategoryIDs {
			_, err = tx.Exec(ctx, `
				INSERT INTO email_categories (email_id, category_id, created_at)
				VALUES ($1, $2, NOW())
			`, emailID, categoryID)
			if err != nil {
				return fmt.Errorf("failed to add email to category: %w", err)
			}
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
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
		SELECT id, account_id, gmail_message_id, sender, subject, body, 
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
			&email.ID, &email.AccountID, &email.GmailMessageID,
			&email.Sender, &email.Subject, &email.Body, &email.AISummary,
			&email.ReceivedAt, &email.IsArchivedInGmail, &email.UnsubscribeLink,
			&email.CreatedAt, &email.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan email: %w", err)
		}
		
		// Load categories for this email
		categoryIDs, err := r.GetEmailCategories(ctx, email.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get email categories: %w", err)
		}
		email.CategoryIDs = categoryIDs
		
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

func (r *EmailRepository) UpdateCategoriesByGmailMessageID(ctx context.Context, accountID int64, gmailMessageID string, categoryIDs []int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get email ID
	var emailID int64
	err = tx.QueryRow(ctx, `
		SELECT id FROM emails 
		WHERE account_id = $1 AND gmail_message_id = $2
	`, accountID, gmailMessageID).Scan(&emailID)
	if err != nil {
		return fmt.Errorf("failed to get email ID: %w", err)
	}

	// Remove existing categories
	_, err = tx.Exec(ctx, "DELETE FROM email_categories WHERE email_id = $1", emailID)
	if err != nil {
		return fmt.Errorf("failed to remove existing categories: %w", err)
	}

	// Add new categories
	for _, categoryID := range categoryIDs {
		_, err = tx.Exec(ctx, `
			INSERT INTO email_categories (email_id, category_id, created_at)
			VALUES ($1, $2, NOW())
		`, emailID, categoryID)
		if err != nil {
			return fmt.Errorf("failed to add email to category: %w", err)
		}
	}

	// Update email timestamp
	_, err = tx.Exec(ctx, "UPDATE emails SET updated_at = NOW() WHERE id = $1", emailID)
	if err != nil {
		return fmt.Errorf("failed to update email timestamp: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *EmailRepository) GetByCategoryIDPaginated(ctx context.Context, accountID, categoryID int64, params repositories.PaginationParams) (*repositories.PaginatedEmails, error) {
	// Get total count
	var totalCount int64
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM emails e
		INNER JOIN email_categories ec ON e.id = ec.email_id
		WHERE e.account_id = $1 AND ec.category_id = $2
	`, accountID, categoryID).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Calculate offset
	offset := (params.Page - 1) * params.PageSize
	
	// Get paginated emails
	rows, err := r.db.Query(ctx, `
		SELECT e.id, e.account_id, e.gmail_message_id, e.sender, e.subject, e.body, 
		       e.ai_summary, e.received_at, e.is_archived_in_gmail, e.unsubscribe_link, e.created_at, e.updated_at
		FROM emails e
		INNER JOIN email_categories ec ON e.id = ec.email_id
		WHERE e.account_id = $1 AND ec.category_id = $2
		ORDER BY e.received_at DESC
		LIMIT $3 OFFSET $4
	`, accountID, categoryID, params.PageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query paginated emails by category: %w", err)
	}
	defer rows.Close()

	var emails []entities.Email
	for rows.Next() {
		var email entities.Email
		err := rows.Scan(
			&email.ID, &email.AccountID, &email.GmailMessageID,
			&email.Sender, &email.Subject, &email.Body, &email.AISummary,
			&email.ReceivedAt, &email.IsArchivedInGmail, &email.UnsubscribeLink,
			&email.CreatedAt, &email.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan email: %w", err)
		}
		
		// Load categories for this email
		categoryIDs, err := r.GetEmailCategories(ctx, email.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get email categories: %w", err)
		}
		email.CategoryIDs = categoryIDs
		
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

func (r *EmailRepository) GetEmailCategories(ctx context.Context, emailID int64) ([]int64, error) {
	rows, err := r.db.Query(ctx, `
		SELECT category_id FROM email_categories WHERE email_id = $1
	`, emailID)
	if err != nil {
		return nil, fmt.Errorf("failed to query email categories: %w", err)
	}
	defer rows.Close()

	var categoryIDs []int64
	for rows.Next() {
		var categoryID int64
		err := rows.Scan(&categoryID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category ID: %w", err)
		}
		categoryIDs = append(categoryIDs, categoryID)
	}

	return categoryIDs, nil
}

func (r *EmailRepository) AddEmailToCategories(ctx context.Context, emailID int64, categoryIDs []int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, categoryID := range categoryIDs {
		_, err = tx.Exec(ctx, `
			INSERT INTO email_categories (email_id, category_id, created_at)
			VALUES ($1, $2, NOW())
			ON CONFLICT (email_id, category_id) DO NOTHING
		`, emailID, categoryID)
		if err != nil {
			return fmt.Errorf("failed to add email to category: %w", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *EmailRepository) RemoveEmailFromCategories(ctx context.Context, emailID int64, categoryIDs []int64) error {
	if len(categoryIDs) == 0 {
		return nil
	}

	// Build the IN clause for the category IDs
	args := make([]interface{}, len(categoryIDs)+1)
	args[0] = emailID
	placeholders := make([]string, len(categoryIDs))
	for i, categoryID := range categoryIDs {
		args[i+1] = categoryID
		placeholders[i] = fmt.Sprintf("$%d", i+2)
	}

	query := fmt.Sprintf(`
		DELETE FROM email_categories 
		WHERE email_id = $1 AND category_id IN (%s)
	`, strings.Join(placeholders, ", "))

	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to remove email from categories: %w", err)
	}

	return nil
}