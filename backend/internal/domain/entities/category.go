package entities

import "time"

type Category struct {
	ID          int64     `json:"id"`
	AccountID   int64     `json:"account_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}