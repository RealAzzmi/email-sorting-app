package gmail

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
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
			Labels:     msg.LabelIds,
			ReceivedAt: time.Unix(msg.InternalDate/1000, 0),
		}

		gmailMessages = append(gmailMessages, gmailMsg)
	}

	return gmailMessages, nil
}

func (s *GmailService) ListAllMessages(ctx context.Context, token *oauth2.Token) ([]entities.GmailMessage, error) {
	client := s.oauthConfig.Client(ctx, token)
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}

	var allMessages []entities.GmailMessage
	pageToken := ""

	for {
		req := srv.Users.Messages.List("me").MaxResults(500) // Gmail's max per request
		if pageToken != "" {
			req = req.PageToken(pageToken)
		}

		messages, err := req.Do()
		if err != nil {
			return nil, fmt.Errorf("failed to list messages: %w", err)
		}

		// Process messages in batches to avoid overwhelming the API
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
				Labels:     msg.LabelIds,
				ReceivedAt: time.Unix(msg.InternalDate/1000, 0),
			}

			allMessages = append(allMessages, gmailMsg)
		}

		// Check if there are more pages
		if messages.NextPageToken == "" {
			break
		}
		pageToken = messages.NextPageToken
	}

	return allMessages, nil
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
		Labels:     msg.LabelIds,
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

func (s *GmailService) GetCurrentHistoryId(ctx context.Context, token *oauth2.Token) (string, error) {
	client := s.oauthConfig.Client(ctx, token)
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return "", fmt.Errorf("failed to create Gmail service: %w", err)
	}

	// Get the user's profile to get the current history ID
	profile, err := srv.Users.GetProfile("me").Do()
	if err != nil {
		return "", fmt.Errorf("failed to get user profile: %w", err)
	}

	return fmt.Sprintf("%d", profile.HistoryId), nil
}

func (s *GmailService) ListHistory(ctx context.Context, token *oauth2.Token, startHistoryId string) ([]entities.GmailMessage, string, error) {
	client := s.oauthConfig.Client(ctx, token)
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, "", fmt.Errorf("failed to create Gmail service: %w", err)
	}

	// Convert string history ID to uint64
	startHistoryIdUint, err := strconv.ParseUint(startHistoryId, 10, 64)
	if err != nil {
		return nil, "", fmt.Errorf("invalid history ID format: %w", err)
	}

	// Get history changes since the given history ID
	historyList, err := srv.Users.History.List("me").StartHistoryId(startHistoryIdUint).Do()
	if err != nil {
		return nil, "", fmt.Errorf("failed to list history: %w", err)
	}

	var newMessages []entities.GmailMessage
	messageIds := make(map[string]bool) // To avoid duplicates

	// Process all history records
	for _, history := range historyList.History {
		// Process messages added
		for _, msg := range history.MessagesAdded {
			if !messageIds[msg.Message.Id] {
				messageIds[msg.Message.Id] = true
				
				// Fetch the full message
				fullMsg, err := srv.Users.Messages.Get("me", msg.Message.Id).Do()
				if err != nil {
					continue // Skip if we can't fetch it
				}

				gmailMsg := entities.GmailMessage{
					ID:         msg.Message.Id,
					Sender:     s.getHeaderValue(fullMsg.Payload.Headers, "From"),
					Subject:    s.getHeaderValue(fullMsg.Payload.Headers, "Subject"),
					Body:       s.extractBody(fullMsg.Payload),
					Labels:     fullMsg.LabelIds,
					ReceivedAt: time.Unix(fullMsg.InternalDate/1000, 0),
				}

				newMessages = append(newMessages, gmailMsg)
			}
		}

		// Process label changes (labels added or removed)
		for _, labelChange := range history.LabelsAdded {
			if !messageIds[labelChange.Message.Id] {
				messageIds[labelChange.Message.Id] = true
				
				// Fetch the full message with updated labels
				fullMsg, err := srv.Users.Messages.Get("me", labelChange.Message.Id).Do()
				if err != nil {
					continue // Skip if we can't fetch it
				}

				gmailMsg := entities.GmailMessage{
					ID:         labelChange.Message.Id,
					Sender:     s.getHeaderValue(fullMsg.Payload.Headers, "From"),
					Subject:    s.getHeaderValue(fullMsg.Payload.Headers, "Subject"),
					Body:       s.extractBody(fullMsg.Payload),
					Labels:     fullMsg.LabelIds,
					ReceivedAt: time.Unix(fullMsg.InternalDate/1000, 0),
				}

				newMessages = append(newMessages, gmailMsg)
			}
		}

		for _, labelChange := range history.LabelsRemoved {
			if !messageIds[labelChange.Message.Id] {
				messageIds[labelChange.Message.Id] = true
				
				// Fetch the full message with updated labels
				fullMsg, err := srv.Users.Messages.Get("me", labelChange.Message.Id).Do()
				if err != nil {
					continue // Skip if we can't fetch it
				}

				gmailMsg := entities.GmailMessage{
					ID:         labelChange.Message.Id,
					Sender:     s.getHeaderValue(fullMsg.Payload.Headers, "From"),
					Subject:    s.getHeaderValue(fullMsg.Payload.Headers, "Subject"),
					Body:       s.extractBody(fullMsg.Payload),
					Labels:     fullMsg.LabelIds,
					ReceivedAt: time.Unix(fullMsg.InternalDate/1000, 0),
				}

				newMessages = append(newMessages, gmailMsg)
			}
		}
	}

	// Return the new messages and the latest history ID
	latestHistoryId := fmt.Sprintf("%d", historyList.HistoryId)
	return newMessages, latestHistoryId, nil
}

func (s *GmailService) GetLabelNames(ctx context.Context, token *oauth2.Token, labelIds []string) (map[string]string, error) {
	if len(labelIds) == 0 {
		return make(map[string]string), nil
	}

	client := s.oauthConfig.Client(ctx, token)
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}

	// Get all labels for this user
	labelsResponse, err := srv.Users.Labels.List("me").Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list labels: %w", err)
	}

	// Create a map of label ID to label name
	labelMap := make(map[string]string)
	for _, label := range labelsResponse.Labels {
		labelMap[label.Id] = label.Name
	}

	// Filter to only requested label IDs
	result := make(map[string]string)
	for _, labelId := range labelIds {
		if name, exists := labelMap[labelId]; exists {
			result[labelId] = name
		} else {
			// Fallback to ID if name not found
			result[labelId] = labelId
		}
	}

	return result, nil
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

func (s *GmailService) DeleteLabel(ctx context.Context, token *oauth2.Token, labelName string) error {
	client := s.oauthConfig.Client(ctx, token)
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("failed to create Gmail service: %w", err)
	}

	// First, get all labels to find the one with the matching name
	labelsResponse, err := srv.Users.Labels.List("me").Do()
	if err != nil {
		return fmt.Errorf("failed to list labels: %w", err)
	}

	var labelID string
	for _, label := range labelsResponse.Labels {
		if label.Name == labelName {
			// Don't allow deletion of system labels
			if label.Type == "system" {
				return fmt.Errorf("cannot delete system label: %s", labelName)
			}
			labelID = label.Id
			break
		}
	}

	if labelID == "" {
		return fmt.Errorf("label not found: %s", labelName)
	}

	// Delete the label
	err = srv.Users.Labels.Delete("me", labelID).Do()
	if err != nil {
		return fmt.Errorf("failed to delete label: %w", err)
	}

	return nil
}

func (s *GmailService) CreateLabel(ctx context.Context, token *oauth2.Token, labelName string) error {
	client := s.oauthConfig.Client(ctx, token)
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("failed to create Gmail service: %w", err)
	}

	// Check if label already exists
	labelsResponse, err := srv.Users.Labels.List("me").Do()
	if err != nil {
		return fmt.Errorf("failed to list labels: %w", err)
	}

	for _, label := range labelsResponse.Labels {
		if label.Name == labelName {
			// Label already exists, no need to create
			return nil
		}
	}

	// Create the new label
	newLabel := &gmail.Label{
		Name:                labelName,
		MessageListVisibility: "show",
		LabelListVisibility:   "labelShow",
	}

	_, err = srv.Users.Labels.Create("me", newLabel).Do()
	if err != nil {
		return fmt.Errorf("failed to create label: %w", err)
	}

	return nil
}

func (s *GmailService) GetAllLabels(ctx context.Context, token *oauth2.Token) ([]entities.GmailLabel, error) {
	client := s.oauthConfig.Client(ctx, token)
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}

	labelsResponse, err := srv.Users.Labels.List("me").Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list labels: %w", err)
	}

	var gmailLabels []entities.GmailLabel
	for _, label := range labelsResponse.Labels {
		gmailLabels = append(gmailLabels, entities.GmailLabel{
			ID:   label.Id,
			Name: label.Name,
			Type: label.Type,
		})
	}

	return gmailLabels, nil
}