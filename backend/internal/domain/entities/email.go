package entities

import "time"

type Email struct {
	ID               int64      `json:"id"`
	AccountID        int64      `json:"account_id"`
	CategoryIDs      []int64    `json:"category_ids"`
	Categories       []Category `json:"categories,omitempty"`
	GmailMessageID   string     `json:"gmail_message_id"`
	Sender           string     `json:"sender"`
	Subject          string     `json:"subject"`
	Body             string     `json:"body"`
	AISummary        *string    `json:"ai_summary"`
	ReceivedAt       time.Time  `json:"received_at"`
	IsArchivedInGmail bool      `json:"is_archived_in_gmail"`
	UnsubscribeLink  *string    `json:"unsubscribe_link"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type EmailCategory struct {
	ID         int64     `json:"id"`
	EmailID    int64     `json:"email_id"`
	CategoryID int64     `json:"category_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type GmailMessage struct {
	ID              string
	Sender          string
	Subject         string
	Body            string
	Headers         map[string]string
	Labels          []string
	UnsubscribeLink *string
	ReceivedAt      time.Time
}

type GmailLabel struct {
	ID   string
	Name string
	Type string
}