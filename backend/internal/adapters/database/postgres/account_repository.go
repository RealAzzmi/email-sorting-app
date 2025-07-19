package postgres

import (
	"context"
	"fmt"

	"github.com/email-sorting-app/internal/domain/entities"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"
)

type AccountRepository struct {
	db *pgxpool.Pool
}

func NewAccountRepository(db *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) GetAll(ctx context.Context) ([]entities.Account, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, email, name, created_at, updated_at 
		FROM accounts 
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query accounts: %w", err)
	}
	defer rows.Close()

	var accounts []entities.Account
	for rows.Next() {
		var account entities.Account
		err := rows.Scan(&account.ID, &account.Email, &account.Name, &account.CreatedAt, &account.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (r *AccountRepository) GetByID(ctx context.Context, id int64) (*entities.Account, error) {
	var account entities.Account
	err := r.db.QueryRow(ctx, `
		SELECT id, email, name, access_token, refresh_token, token_expiry, created_at, updated_at 
		FROM accounts WHERE id = $1
	`, id).Scan(
		&account.ID, &account.Email, &account.Name, &account.AccessToken, 
		&account.RefreshToken, &account.TokenExpiry, &account.CreatedAt, &account.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("account not found")
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return &account, nil
}

func (r *AccountRepository) GetByEmail(ctx context.Context, email string) (*entities.Account, error) {
	var account entities.Account
	err := r.db.QueryRow(ctx, `
		SELECT id, email, name, access_token, refresh_token, token_expiry, created_at, updated_at 
		FROM accounts WHERE email = $1
	`, email).Scan(
		&account.ID, &account.Email, &account.Name, &account.AccessToken, 
		&account.RefreshToken, &account.TokenExpiry, &account.CreatedAt, &account.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("account not found")
		}
		return nil, fmt.Errorf("failed to get account by email: %w", err)
	}

	return &account, nil
}

func (r *AccountRepository) Create(ctx context.Context, email, name string, token *oauth2.Token) (*entities.Account, error) {
	var account entities.Account
	err := r.db.QueryRow(ctx, `
		INSERT INTO accounts (email, name, access_token, refresh_token, token_expiry, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, email, name, created_at, updated_at
	`, email, name, token.AccessToken, token.RefreshToken, token.Expiry).Scan(
		&account.ID, &account.Email, &account.Name, &account.CreatedAt, &account.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return &account, nil
}

func (r *AccountRepository) Update(ctx context.Context, account *entities.Account) error {
	_, err := r.db.Exec(ctx, `
		UPDATE accounts 
		SET access_token = $1, refresh_token = $2, token_expiry = $3, updated_at = NOW()
		WHERE id = $4
	`, account.AccessToken, account.RefreshToken, account.TokenExpiry, account.ID)
	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	return nil
}

func (r *AccountRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, "DELETE FROM accounts WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	return nil
}