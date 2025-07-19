package gmail

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/email-sorting-app/internal/domain/entities"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type GmailService struct {
	oauthConfig *oauth2.Config
}

func NewGmailService(oauthConfig *oauth2.Config) *GmailService {
	return &GmailService{
		oauthConfig: oauthConfig,
	}
}

func (s *GmailService) ListMessages(ctx context.Context, token *oauth2.Token, maxResults int64) ([]entities.GmailMessage, error) {
	client := s.oauthConfig.Client(ctx, token)
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}

	messages, err := srv.Users.Messages.List("me").MaxResults(maxResults).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	var gmailMessages []entities.GmailMessage
	for _, message := range messages.Messages {
		msg, err := srv.Users.Messages.Get("me", message.Id).Do()
		if err != nil {
			continue // Skip this email if we can't fetch it
		}

		gmailMsg := entities.GmailMessage{
			ID:         message.Id,
			Sender:     s.getHeaderValue(msg.Payload.Headers, "From"),
			Subject:    s.getHeaderValue(msg.Payload.Headers, "Subject"),
			Body:       s.extractBody(msg.Payload),
			ReceivedAt: time.Unix(msg.InternalDate/1000, 0),
		}

		gmailMessages = append(gmailMessages, gmailMsg)
	}

	return gmailMessages, nil
}

func (s *GmailService) GetMessage(ctx context.Context, token *oauth2.Token, messageID string) (*entities.GmailMessage, error) {
	client := s.oauthConfig.Client(ctx, token)
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}

	msg, err := srv.Users.Messages.Get("me", messageID).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	gmailMsg := &entities.GmailMessage{
		ID:         messageID,
		Sender:     s.getHeaderValue(msg.Payload.Headers, "From"),
		Subject:    s.getHeaderValue(msg.Payload.Headers, "Subject"),
		Body:       s.extractBody(msg.Payload),
		ReceivedAt: time.Unix(msg.InternalDate/1000, 0),
	}

	return gmailMsg, nil
}

func (s *GmailService) ArchiveMessage(ctx context.Context, token *oauth2.Token, messageID string) error {
	client := s.oauthConfig.Client(ctx, token)
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("failed to create Gmail service: %w", err)
	}

	// Remove INBOX label to archive the message
	req := &gmail.ModifyMessageRequest{
		RemoveLabelIds: []string{"INBOX"},
	}

	_, err = srv.Users.Messages.Modify("me", messageID, req).Do()
	if err != nil {
		return fmt.Errorf("failed to archive message: %w", err)
	}

	return nil
}

func (s *GmailService) getHeaderValue(headers []*gmail.MessagePartHeader, name string) string {
	for _, header := range headers {
		if header.Name == name {
			return header.Value
		}
	}
	return ""
}

func (s *GmailService) extractBody(part *gmail.MessagePart) string {
	if part.Body != nil && part.Body.Data != "" {
		data, err := base64.URLEncoding.DecodeString(part.Body.Data)
		if err == nil {
			return string(data)
		}
	}

	for _, subPart := range part.Parts {
		if subPart.MimeType == "text/plain" {
			if subPart.Body != nil && subPart.Body.Data != "" {
				data, err := base64.URLEncoding.DecodeString(subPart.Body.Data)
				if err == nil {
					return string(data)
				}
			}
		}
	}

	return ""
}