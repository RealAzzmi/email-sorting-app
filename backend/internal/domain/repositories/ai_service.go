package repositories

import (
	"context"

	"github.com/email-sorting-app/internal/domain/entities"
)

type UnsubscribeAction struct {
	Action    string                 `json:"action"`     // "click", "fill", "select", "wait"
	Selector  string                 `json:"selector"`   // CSS selector or element description
	Value     string                 `json:"value"`      // Text to fill or option to select
	Reasoning string                 `json:"reasoning"`  // Why this action is needed
	MetaData  map[string]interface{} `json:"metadata"`   // Additional action-specific data
}

type UnsubscribePageAnalysis struct {
	IsUnsubscribePage bool                  `json:"is_unsubscribe_page"`
	RequiresAuth      bool                  `json:"requires_auth"`
	Actions           []UnsubscribeAction   `json:"actions"`
	SuccessIndicators []string              `json:"success_indicators"`
	ErrorIndicators   []string              `json:"error_indicators"`
	Reasoning         string                `json:"reasoning"`
}

type AIService interface {
	SummarizeEmail(ctx context.Context, email *entities.Email) (string, error)
	CategorizeEmail(ctx context.Context, email *entities.Email, categories []entities.Category) ([]int64, error)
	AnalyzeUnsubscribePage(ctx context.Context, pageContent, pageURL string) (*UnsubscribePageAnalysis, error)
	ExtractUnsubscribeLink(ctx context.Context, headers, body string) (string, error)
}