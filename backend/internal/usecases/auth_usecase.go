package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/email-sorting-app/internal/domain/entities"
	"github.com/email-sorting-app/internal/domain/repositories"
	"golang.org/x/oauth2"
)

type AuthUsecase struct {
	accountRepo repositories.AccountRepository
	oauthConfig *oauth2.Config
}

func NewAuthUsecase(accountRepo repositories.AccountRepository, oauthConfig *oauth2.Config) *AuthUsecase {
	return &AuthUsecase{
		accountRepo: accountRepo,
		oauthConfig: oauthConfig,
	}
}

func (u *AuthUsecase) GetAuthURL() string {
	state := fmt.Sprintf("%d", time.Now().Unix())
	return u.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

func (u *AuthUsecase) HandleCallback(ctx context.Context, code string) (*entities.Account, error) {
	if code == "" {
		return nil, fmt.Errorf("missing authorization code")
	}

	// Exchange code for token
	token, err := u.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Get user info
	userInfo, err := u.getUserInfo(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Check if account exists
	existingAccount, err := u.accountRepo.GetByEmail(ctx, userInfo.Email)
	if err == nil && existingAccount != nil {
		// Update existing account tokens
		existingAccount.UpdateTokens(token)
		err = u.accountRepo.Update(ctx, existingAccount)
		if err != nil {
			return nil, fmt.Errorf("failed to update account: %w", err)
		}
		return existingAccount, nil
	}

	// Create new account
	account, err := u.accountRepo.Create(ctx, userInfo.Email, userInfo.Name, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return account, nil
}

func (u *AuthUsecase) getUserInfo(ctx context.Context, token *oauth2.Token) (*entities.UserInfo, error) {
	client := u.oauthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	var userInfo entities.UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}
