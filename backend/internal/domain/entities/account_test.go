package entities

import (
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestAccount_ToOAuth2Token(t *testing.T) {
	expiry := time.Now().Add(time.Hour)
	account := &Account{
		AccessToken:  "access-token-123",
		RefreshToken: "refresh-token-456",
		TokenExpiry:  expiry,
	}

	token := account.ToOAuth2Token()

	if token.AccessToken != "access-token-123" {
		t.Errorf("Expected AccessToken to be 'access-token-123', got '%s'", token.AccessToken)
	}

	if token.RefreshToken != "refresh-token-456" {
		t.Errorf("Expected RefreshToken to be 'refresh-token-456', got '%s'", token.RefreshToken)
	}

	if !token.Expiry.Equal(expiry) {
		t.Errorf("Expected Expiry to be %v, got %v", expiry, token.Expiry)
	}
}

func TestAccount_UpdateTokens(t *testing.T) {
	account := &Account{
		AccessToken:  "old-access-token",
		RefreshToken: "old-refresh-token",
		TokenExpiry:  time.Now().Add(-time.Hour),
		UpdatedAt:    time.Now().Add(-time.Hour),
	}

	newExpiry := time.Now().Add(time.Hour)
	newToken := &oauth2.Token{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
		Expiry:       newExpiry,
	}

	account.UpdateTokens(newToken)

	if account.AccessToken != "new-access-token" {
		t.Errorf("Expected AccessToken to be 'new-access-token', got '%s'", account.AccessToken)
	}

	if account.RefreshToken != "new-refresh-token" {
		t.Errorf("Expected RefreshToken to be 'new-refresh-token', got '%s'", account.RefreshToken)
	}

	if !account.TokenExpiry.Equal(newExpiry) {
		t.Errorf("Expected TokenExpiry to be %v, got %v", newExpiry, account.TokenExpiry)
	}

	if account.UpdatedAt.Before(time.Now().Add(-time.Second)) {
		t.Error("Expected UpdatedAt to be updated to current time")
	}
}