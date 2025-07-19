package repositories

import (
	"context"

	"github.com/email-sorting-app/internal/domain/entities"
	"golang.org/x/oauth2"
)

type GmailService interface {
	ListMessages(ctx context.Context, token *oauth2.Token, maxResults int64) ([]entities.GmailMessage, error)
	ListAllMessages(ctx context.Context, token *oauth2.Token) ([]entities.GmailMessage, error)
	GetMessage(ctx context.Context, token *oauth2.Token, messageID string) (*entities.GmailMessage, error)
	ArchiveMessage(ctx context.Context, token *oauth2.Token, messageID string) error
	GetCurrentHistoryId(ctx context.Context, token *oauth2.Token) (string, error)
	ListHistory(ctx context.Context, token *oauth2.Token, startHistoryId string) ([]entities.GmailMessage, string, error)
	GetLabelNames(ctx context.Context, token *oauth2.Token, labelIds []string) (map[string]string, error)
}