package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
	"sohoaas-backend/internal/types"
)

// FirebaseAuthService handles Firebase authentication and JWT validation
type FirebaseAuthService struct {
	client *auth.Client
	ctx    context.Context
}

// NewFirebaseAuthService creates a new Firebase authentication service
func NewFirebaseAuthService() (*FirebaseAuthService, error) {
	ctx := context.Background()
	
	// Build service account credentials from environment variables
	credentials := map[string]interface{}{
		"type":                        "service_account",
		"project_id":                 os.Getenv("FIREBASE_PROJECT_ID"),
		"private_key_id":             os.Getenv("FIREBASE_PRIVATE_KEY_ID"),
		"private_key":                os.Getenv("FIREBASE_PRIVATE_KEY"),
		"client_email":               os.Getenv("FIREBASE_CLIENT_EMAIL"),
		"client_id":                  os.Getenv("FIREBASE_CLIENT_ID"),
		"auth_uri":                   os.Getenv("FIREBASE_AUTH_URI"),
		"token_uri":                  os.Getenv("FIREBASE_TOKEN_URI"),
		"auth_provider_x509_cert_url": os.Getenv("FIREBASE_AUTH_PROVIDER_X509_CERT_URL"),
		"client_x509_cert_url":       os.Getenv("FIREBASE_CLIENT_X509_CERT_URL"),
		"universe_domain":            os.Getenv("FIREBASE_UNIVERSE_DOMAIN"),
	}
	
	// Validate required environment variables
	if credentials["project_id"] == "" || credentials["private_key"] == "" || credentials["client_email"] == "" {
		return nil, fmt.Errorf("missing required Firebase environment variables: FIREBASE_PROJECT_ID, FIREBASE_PRIVATE_KEY, FIREBASE_CLIENT_EMAIL")
	}
	
	// Convert credentials to JSON
	credentialsJSON, err := json.Marshal(credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Firebase credentials: %v", err)
	}
	
	// Initialize Firebase Admin SDK with credentials JSON
	opt := option.WithCredentialsJSON(credentialsJSON)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %v", err)
	}

	// Get Auth client
	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Firebase Auth client: %v", err)
	}

	log.Println("Firebase Auth service initialized successfully")
	
	return &FirebaseAuthService{
		client: client,
		ctx:    ctx,
	}, nil
}

// ValidateIDToken validates a Firebase ID token and returns user information
func (f *FirebaseAuthService) ValidateIDToken(idToken string) (*types.User, error) {
	// Verify the ID token
	token, err := f.client.VerifyIDToken(f.ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %v", err)
	}

	// Get user record for additional information
	userRecord, err := f.client.GetUser(f.ctx, token.UID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user record: %v", err)
	}

	// Server-side email allowlist validation
	if !f.IsEmailAllowedServerSide(userRecord.Email) {
		return nil, fmt.Errorf("email %s is not in the allowed list", userRecord.Email)
	}

	// Create SOHOAAS user from Firebase user
	user := &types.User{
		ID:    token.UID,
		Email: userRecord.Email,
		Name:  userRecord.DisplayName,
		OAuthTokens: map[string]interface{}{
			"google": map[string]interface{}{
				"access_token": idToken,
				"token_type":   "Bearer",
			},
		},
		ConnectedServices: []string{"gmail", "calendar", "docs", "drive"},
	}

	return user, nil
}

// IsEmailAllowed checks if an email is in the allowed list (legacy method)
func (f *FirebaseAuthService) IsEmailAllowed(email string, allowedEmails []string) bool {
	if len(allowedEmails) == 0 {
		return true // No restrictions if allowlist is empty
	}
	
	for _, allowedEmail := range allowedEmails {
		if email == allowedEmail {
			return true
		}
	}
	
	return false
}

// IsEmailAllowedServerSide checks if user exists in Firebase Console
func (f *FirebaseAuthService) IsEmailAllowedServerSide(email string) bool {
	// Check if ALLOWED_EMAILS environment variable is set (fallback to old method)
	allowedEmailsEnv := os.Getenv("ALLOWED_EMAILS")
	if allowedEmailsEnv != "" {
		// Use environment variable method if configured
		allowedEmails := strings.Split(allowedEmailsEnv, ",")
		for _, allowedEmail := range allowedEmails {
			allowedEmail = strings.TrimSpace(allowedEmail)
			if email == allowedEmail {
				return true
			}
		}
		log.Printf("Email %s not in environment allowlist: %s", email, allowedEmailsEnv)
		return false
	}
	
	// Firebase Console approach: Check if user exists in Firebase project
	// This allows dynamic user management through Firebase Console
	userRecord, err := f.client.GetUserByEmail(f.ctx, email)
	if err != nil {
		// User doesn't exist in Firebase project = not allowed
		log.Printf("Email %s not found in Firebase Console users: %v", email, err)
		return false
	}
	
	// Additional check: ensure user is not disabled
	if userRecord.Disabled {
		log.Printf("Email %s is disabled in Firebase Console", email)
		return false
	}
	
	log.Printf("Email %s validated via Firebase Console user management", email)
	return true
}

// GetUserByUID retrieves user information by Firebase UID
func (f *FirebaseAuthService) GetUserByUID(uid string) (*types.User, error) {
	userRecord, err := f.client.GetUser(f.ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by UID: %v", err)
	}

	user := &types.User{
		ID:    userRecord.UID,
		Email: userRecord.Email,
		Name:  userRecord.DisplayName,
		OAuthTokens: map[string]interface{}{
			"google": map[string]interface{}{
				"token_type": "Bearer",
			},
		},
		ConnectedServices: []string{"gmail", "calendar", "docs", "drive"},
	}

	return user, nil
}
