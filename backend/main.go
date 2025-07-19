package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type Config struct {
	DatabaseURL      string
	GoogleClientID   string
	GoogleSecret     string
	RedirectURL      string
	Port             string
}

type App struct {
	db     *pgxpool.Pool
	config *Config
	oauth  *oauth2.Config
}

type Account struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	AccessToken  string    `json:"-"`
	RefreshToken string    `json:"-"`
	TokenExpiry  time.Time `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Email struct {
	ID               int64     `json:"id"`
	AccountID        int64     `json:"account_id"`
	CategoryID       *int64    `json:"category_id"`
	GmailMessageID   string    `json:"gmail_message_id"`
	Sender           string    `json:"sender"`
	Subject          string    `json:"subject"`
	Body             string    `json:"body"`
	AISummary        *string   `json:"ai_summary"`
	ReceivedAt       time.Time `json:"received_at"`
	IsArchivedInGmail bool     `json:"is_archived_in_gmail"`
	UnsubscribeLink  *string   `json:"unsubscribe_link"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func loadConfig() *Config {
	return &Config{
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost/email_sorting_app?sslmode=disable"),
		GoogleClientID: getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleSecret:   getEnv("GOOGLE_CLIENT_SECRET", ""),
		RedirectURL:    getEnv("REDIRECT_URL", "http://localhost:8080/auth/callback"),
		Port:           getEnv("PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func initDB(databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

func setupOAuth(config *Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     config.GoogleClientID,
		ClientSecret: config.GoogleSecret,
		RedirectURL:  config.RedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/gmail.readonly",
			"https://www.googleapis.com/auth/gmail.modify",
		},
		Endpoint: google.Endpoint,
	}
}

func (a *App) setupRoutes() *gin.Engine {
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Auth routes
	r.GET("/auth/login", a.handleAuthLogin)
	r.GET("/auth/callback", a.handleAuthCallback)
	r.POST("/auth/logout", a.handleLogout)

	// Account routes
	r.GET("/accounts", a.handleGetAccounts)
	r.GET("/accounts/:id/emails", a.handleGetAccountEmails)
	r.DELETE("/accounts/:id", a.handleDeleteAccount)

	return r
}

func (a *App) handleAuthLogin(c *gin.Context) {
	state := fmt.Sprintf("%d", time.Now().Unix())
	url := a.oauth.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	c.JSON(200, gin.H{"auth_url": url})
}

func (a *App) handleAuthCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(400, gin.H{"error": "Missing authorization code"})
		return
	}

	token, err := a.oauth.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to exchange code for token"})
		return
	}

	// Get user info
	client := a.oauth.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get user info"})
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		c.JSON(500, gin.H{"error": "Failed to decode user info"})
		return
	}

	// Save or update account
	_, err = a.saveAccount(userInfo.Email, userInfo.Name, token)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to save account"})
		return
	}

	// Redirect to frontend dashboard
	c.Redirect(302, "http://localhost:3000/dashboard")
}

func (a *App) handleLogout(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Logged out successfully"})
}

func (a *App) handleGetAccounts(c *gin.Context) {
	rows, err := a.db.Query(context.Background(), `
		SELECT id, email, name, created_at, updated_at 
		FROM accounts 
		ORDER BY created_at DESC
	`)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch accounts"})
		return
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		var account Account
		err := rows.Scan(&account.ID, &account.Email, &account.Name, &account.CreatedAt, &account.UpdatedAt)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to scan account"})
			return
		}
		accounts = append(accounts, account)
	}

	c.JSON(200, gin.H{"accounts": accounts})
}

func (a *App) handleGetAccountEmails(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid account ID"})
		return
	}

	// First get account details to check if it exists and get tokens
	var account Account
	err = a.db.QueryRow(context.Background(), `
		SELECT id, email, name, access_token, refresh_token, token_expiry 
		FROM accounts WHERE id = $1
	`, accountID).Scan(&account.ID, &account.Email, &account.Name, &account.AccessToken, &account.RefreshToken, &account.TokenExpiry)
	if err != nil {
		c.JSON(404, gin.H{"error": "Account not found"})
		return
	}

	// First, try to get emails from database
	emails, err := a.getEmailsFromDB(accountID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch emails from database"})
		return
	}

	// If no emails in database, fetch from Gmail and store them
	if len(emails) == 0 {
		err = a.syncEmailsFromGmail(account)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to sync emails from Gmail"})
			return
		}
		
		// Get emails from database after sync
		emails, err = a.getEmailsFromDB(accountID)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to fetch emails after sync"})
			return
		}
	}

	c.JSON(200, gin.H{"emails": emails})
}

func (a *App) handleDeleteAccount(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid account ID"})
		return
	}

	_, err = a.db.Exec(context.Background(), "DELETE FROM accounts WHERE id = $1", accountID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete account"})
		return
	}

	c.JSON(200, gin.H{"message": "Account deleted successfully"})
}

func (a *App) saveAccount(email, name string, token *oauth2.Token) (*Account, error) {
	var account Account
	
	// Check if account already exists
	err := a.db.QueryRow(context.Background(), `
		SELECT id, email, name, created_at, updated_at 
		FROM accounts WHERE email = $1
	`, email).Scan(&account.ID, &account.Email, &account.Name, &account.CreatedAt, &account.UpdatedAt)

	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}

	if err == pgx.ErrNoRows {
		// Create new account
		err = a.db.QueryRow(context.Background(), `
			INSERT INTO accounts (email, name, access_token, refresh_token, token_expiry, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
			RETURNING id, email, name, created_at, updated_at
		`, email, name, token.AccessToken, token.RefreshToken, token.Expiry).Scan(
			&account.ID, &account.Email, &account.Name, &account.CreatedAt, &account.UpdatedAt)
	} else {
		// Update existing account
		_, err = a.db.Exec(context.Background(), `
			UPDATE accounts 
			SET access_token = $1, refresh_token = $2, token_expiry = $3, updated_at = NOW()
			WHERE id = $4
		`, token.AccessToken, token.RefreshToken, token.Expiry, account.ID)
	}

	return &account, err
}

func (a *App) getEmailsFromDB(accountID int64) ([]Email, error) {
	rows, err := a.db.Query(context.Background(), `
		SELECT id, account_id, category_id, gmail_message_id, sender, subject, body, ai_summary, received_at, is_archived_in_gmail, unsubscribe_link, created_at, updated_at
		FROM emails 
		WHERE account_id = $1 
		ORDER BY received_at DESC
	`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emails []Email
	for rows.Next() {
		var email Email
		err := rows.Scan(&email.ID, &email.AccountID, &email.CategoryID, &email.GmailMessageID, &email.Sender, &email.Subject, &email.Body, &email.AISummary, &email.ReceivedAt, &email.IsArchivedInGmail, &email.UnsubscribeLink, &email.CreatedAt, &email.UpdatedAt)
		if err != nil {
			return nil, err
		}
		emails = append(emails, email)
	}

	return emails, nil
}

func (a *App) syncEmailsFromGmail(account Account) error {
	// Create Gmail service
	token := &oauth2.Token{
		AccessToken:  account.AccessToken,
		RefreshToken: account.RefreshToken,
		Expiry:       account.TokenExpiry,
	}

	client := a.oauth.Client(context.Background(), token)
	srv, err := gmail.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return err
	}

	// Fetch emails from Gmail
	messages, err := srv.Users.Messages.List("me").MaxResults(50).Do()
	if err != nil {
		return err
	}

	for _, message := range messages.Messages {
		msg, err := srv.Users.Messages.Get("me", message.Id).Do()
		if err != nil {
			continue // Skip this email if we can't fetch it
		}

		// Check if email already exists in database
		var existingID int64
		err = a.db.QueryRow(context.Background(), `
			SELECT id FROM emails WHERE account_id = $1 AND gmail_message_id = $2
		`, account.ID, message.Id).Scan(&existingID)
		
		if err == pgx.ErrNoRows {
			// Email doesn't exist, insert it
			_, err = a.db.Exec(context.Background(), `
				INSERT INTO emails (account_id, gmail_message_id, sender, subject, body, received_at, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
			`, account.ID, message.Id, getHeaderValue(msg.Payload.Headers, "From"), getHeaderValue(msg.Payload.Headers, "Subject"), extractBody(msg.Payload), time.Unix(msg.InternalDate/1000, 0))
			if err != nil {
				log.Printf("Failed to insert email %s: %v", message.Id, err)
			}
		} else if err != nil {
			log.Printf("Failed to check existing email %s: %v", message.Id, err)
		}
	}

	return nil
}

func getHeaderValue(headers []*gmail.MessagePartHeader, name string) string {
	for _, header := range headers {
		if header.Name == name {
			return header.Value
		}
	}
	return ""
}

func extractBody(part *gmail.MessagePart) string {
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

func main() {
	config := loadConfig()

	if config.GoogleClientID == "" || config.GoogleSecret == "" {
		log.Fatal("GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET must be set")
	}

	db, err := initDB(config.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	oauth := setupOAuth(config)

	app := &App{
		db:     db,
		config: config,
		oauth:  oauth,
	}

	r := app.setupRoutes()

	fmt.Printf("Server starting on port %s\n", config.Port)
	log.Fatal(r.Run(":" + config.Port))
}