package postgres

import (
	"context"
	"fmt"

	"github.com/email-sorting-app/internal/domain/entities"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepository struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) GetByAccountID(ctx context.Context, accountID int64) ([]entities.Category, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, account_id, name, description, created_at, updated_at 
		FROM categories 
		WHERE account_id = $1 
		ORDER BY name
	`, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []entities.Category
	for rows.Next() {
		var category entities.Category
		err := rows.Scan(&category.ID, &category.AccountID, &category.Name, &category.Description, &category.CreatedAt, &category.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func (r *CategoryRepository) GetByName(ctx context.Context, accountID int64, name string) (*entities.Category, error) {
	var category entities.Category
	err := r.db.QueryRow(ctx, `
		SELECT id, account_id, name, description, created_at, updated_at 
		FROM categories WHERE account_id = $1 AND name = $2
	`, accountID, name).Scan(
		&category.ID, &category.AccountID, &category.Name, &category.Description, &category.CreatedAt, &category.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get category by name: %w", err)
	}

	return &category, nil
}

func (r *CategoryRepository) Create(ctx context.Context, category *entities.Category) (*entities.Category, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO categories (account_id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`, category.AccountID, category.Name, category.Description).Scan(
		&category.ID, &category.CreatedAt, &category.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return category, nil
}

func (r *CategoryRepository) GetOrCreate(ctx context.Context, accountID int64, name string) (*entities.Category, error) {
	// First try to get existing category
	category, err := r.GetByName(ctx, accountID, name)
	if err != nil {
		return nil, err
	}

	if category != nil {
		return category, nil
	}

	// Create new category if it doesn't exist
	newCategory := &entities.Category{
		AccountID: accountID,
		Name:      name,
	}

	return r.Create(ctx, newCategory)
}

func (r *CategoryRepository) Delete(ctx context.Context, categoryID int64) error {
	_, err := r.db.Exec(ctx, "DELETE FROM categories WHERE id = $1", categoryID)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}
