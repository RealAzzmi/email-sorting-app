package entities

import (
	"time"

	"golang.org/x/oauth2"
)

type Account struct {
	ID                int64     `json:"id"`
	Email             string    `json:"email"`
	Name              string    `json:"name"`
	AccessToken       string    `json:"-"`
	RefreshToken      string    `json:"-"`
	TokenExpiry       time.Time `json:"-"`
	LastSyncHistoryID *string   `json:"-"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type UserInfo struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (a *Account) ToOAuth2Token() *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  a.AccessToken,
		RefreshToken: a.RefreshToken,
		Expiry:       a.TokenExpiry,
	}
}

func (a *Account) UpdateTokens(token *oauth2.Token) {
	a.AccessToken = token.AccessToken
	a.RefreshToken = token.RefreshToken
	a.TokenExpiry = token.Expiry
	a.UpdatedAt = time.Now()
}
