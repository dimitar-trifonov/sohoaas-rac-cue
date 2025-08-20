package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// TokenManager handles secure storage and management of OAuth2 tokens
type TokenManager struct {
	tokens    map[string]*UserTokens // userID -> tokens
	mutex     sync.RWMutex
	config    *oauth2.Config
}

// UserTokens stores OAuth2 tokens for a user
type UserTokens struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	Expiry       time.Time `json:"expiry"`
	UserID       string    `json:"user_id"`
	Email        string    `json:"email"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// NewTokenManager creates a new token manager
func NewTokenManager() *TokenManager {
	// Configure OAuth2 for Google Workspace
	config := &oauth2.Config{
		ClientID:     "", // Will be set from environment
		ClientSecret: "", // Will be set from environment
		RedirectURL:  "postmessage", // For Firebase Auth
		Scopes: []string{
			"https://www.googleapis.com/auth/gmail.modify",
			"https://www.googleapis.com/auth/documents",
			"https://www.googleapis.com/auth/drive",
			"https://www.googleapis.com/auth/calendar",
		},
		Endpoint: google.Endpoint,
	}

	return &TokenManager{
		tokens: make(map[string]*UserTokens),
		config: config,
	}
}

// StoreGoogleToken stores a Google OAuth2 access token for a user
func (tm *TokenManager) StoreGoogleToken(userID, email, accessToken string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	// Validate token by making a test API call
	if err := tm.validateGoogleToken(accessToken); err != nil {
		return fmt.Errorf("invalid Google token: %v", err)
	}

	// Store token with metadata
	tm.tokens[userID] = &UserTokens{
		AccessToken:  accessToken,
		RefreshToken: "", // Firebase doesn't provide refresh tokens directly
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(1 * time.Hour), // Google tokens typically expire in 1 hour
		UserID:       userID,
		Email:        email,
		UpdatedAt:    time.Now(),
	}

	log.Printf("[TokenManager] Stored Google token for user %s (%s)", userID, email)
	return nil
}

// GetGoogleToken retrieves a valid Google OAuth2 token for a user
func (tm *TokenManager) GetGoogleToken(userID string) (string, error) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	userTokens, exists := tm.tokens[userID]
	if !exists {
		return "", fmt.Errorf("no Google token found for user %s", userID)
	}

	// Check if token is expired
	if time.Now().After(userTokens.Expiry) {
		return "", fmt.Errorf("Google token expired for user %s", userID)
	}

	return userTokens.AccessToken, nil
}

// RefreshGoogleToken attempts to refresh an expired Google token
func (tm *TokenManager) RefreshGoogleToken(userID string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	userTokens, exists := tm.tokens[userID]
	if !exists {
		return fmt.Errorf("no tokens found for user %s", userID)
	}

	if userTokens.RefreshToken == "" {
		return fmt.Errorf("no refresh token available for user %s", userID)
	}

	// Use OAuth2 config to refresh token
	token := &oauth2.Token{
		AccessToken:  userTokens.AccessToken,
		RefreshToken: userTokens.RefreshToken,
		TokenType:    userTokens.TokenType,
		Expiry:       userTokens.Expiry,
	}

	ctx := context.Background()
	tokenSource := tm.config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return fmt.Errorf("failed to refresh token: %v", err)
	}

	// Update stored token
	userTokens.AccessToken = newToken.AccessToken
	userTokens.Expiry = newToken.Expiry
	userTokens.UpdatedAt = time.Now()

	log.Printf("[TokenManager] Refreshed Google token for user %s", userID)
	return nil
}

// ValidateUserToken ensures the user owns the provided token
func (tm *TokenManager) ValidateUserToken(userID, providedToken string) error {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	userTokens, exists := tm.tokens[userID]
	if !exists {
		return fmt.Errorf("no tokens found for user %s", userID)
	}

	if userTokens.AccessToken != providedToken {
		return fmt.Errorf("token mismatch for user %s", userID)
	}

	return nil
}

// CleanupExpiredTokens removes expired tokens from memory
func (tm *TokenManager) CleanupExpiredTokens() {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	now := time.Now()
	for userID, tokens := range tm.tokens {
		if now.After(tokens.Expiry.Add(24 * time.Hour)) { // Keep for 24h after expiry
			delete(tm.tokens, userID)
			log.Printf("[TokenManager] Cleaned up expired token for user %s", userID)
		}
	}
}

// GetTokenInfo returns token metadata without exposing the actual token
func (tm *TokenManager) GetTokenInfo(userID string) (*TokenInfo, error) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	userTokens, exists := tm.tokens[userID]
	if !exists {
		return nil, fmt.Errorf("no tokens found for user %s", userID)
	}

	return &TokenInfo{
		UserID:    userTokens.UserID,
		Email:     userTokens.Email,
		TokenType: userTokens.TokenType,
		Expiry:    userTokens.Expiry,
		IsExpired: time.Now().After(userTokens.Expiry),
		UpdatedAt: userTokens.UpdatedAt,
	}, nil
}

// TokenInfo provides token metadata without exposing sensitive data
type TokenInfo struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	TokenType string    `json:"token_type"`
	Expiry    time.Time `json:"expiry"`
	IsExpired bool      `json:"is_expired"`
	UpdatedAt time.Time `json:"updated_at"`
}

// validateGoogleToken validates a Google OAuth2 token by making a test API call
func (tm *TokenManager) validateGoogleToken(token string) error {
	// Make a simple API call to validate the token
	ctx := context.Background()
	client := tm.config.Client(ctx, &oauth2.Token{AccessToken: token})
	
	// Test with Google OAuth2 userinfo endpoint
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return fmt.Errorf("token validation failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("token validation failed with status %d", resp.StatusCode)
	}

	return nil
}

// StartCleanupRoutine starts a background routine to clean up expired tokens
func (tm *TokenManager) StartCleanupRoutine() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			tm.CleanupExpiredTokens()
		}
	}()
}
