package config

import (
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Config struct {
	DatabaseURL      string
	GoogleClientID   string
	GoogleSecret     string
	RedirectURL      string
	Port             string
	GeminiAPIKey     string
}

func Load() (*Config, error) {
	config := &Config{
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost/email_sorting_app?sslmode=disable"),
		GoogleClientID: getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleSecret:   getEnv("GOOGLE_CLIENT_SECRET", ""),
		RedirectURL:    getEnv("REDIRECT_URL", "http://localhost:8080/auth/callback"),
		Port:           getEnv("PORT", "8080"),
		GeminiAPIKey:   getEnv("GEMINI_API_KEY", ""),
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) validate() error {
	if c.GoogleClientID == "" {
		return fmt.Errorf("GOOGLE_CLIENT_ID is required")
	}
	if c.GoogleSecret == "" {
		return fmt.Errorf("GOOGLE_CLIENT_SECRET is required")
	}
	if c.GeminiAPIKey == "" {
		return fmt.Errorf("GEMINI_API_KEY is required")
	}
	return nil
}

func (c *Config) OAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     c.GoogleClientID,
		ClientSecret: c.GoogleSecret,
		RedirectURL:  c.RedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/gmail.readonly",
			"https://www.googleapis.com/auth/gmail.modify",
		},
		Endpoint: google.Endpoint,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}