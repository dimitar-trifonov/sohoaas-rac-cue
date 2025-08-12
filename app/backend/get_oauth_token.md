# Getting Google OAuth2 Tokens for SOHOAAS Testing

## Quick Setup Guide

### 1. **Create Google Cloud Project**
```bash
# Go to: https://console.cloud.google.com/
# Create new project or select existing one
# Enable APIs: Gmail API, Calendar API, Drive API, Docs API
```

### 2. **Create OAuth2 Credentials**
```bash
# In Google Cloud Console:
# APIs & Services → Credentials → Create Credentials → OAuth 2.0 Client IDs
# Application type: Desktop application
# Name: SOHOAAS Testing
```

### 3. **Get Token Using Google OAuth2 Playground**
```bash
# Go to: https://developers.google.com/oauthplayground/
# Click gear icon → Use your own OAuth credentials
# Enter your Client ID and Client Secret
# Select scopes:
#   - https://www.googleapis.com/auth/gmail.compose
#   - https://www.googleapis.com/auth/gmail.send
#   - https://www.googleapis.com/auth/gmail.readonly
#   - https://www.googleapis.com/auth/calendar
#   - https://www.googleapis.com/auth/documents
#   - https://www.googleapis.com/auth/drive.file
# Authorize APIs → Exchange authorization code for tokens
# Copy the Access Token
```

### 4. **Alternative: Using curl**
```bash
# Step 1: Get authorization code
echo "Visit this URL in your browser:"
echo "https://accounts.google.com/o/oauth2/auth?client_id=YOUR_CLIENT_ID&redirect_uri=urn:ietf:wg:oauth:2.0:oob&scope=https://www.googleapis.com/auth/gmail.compose%20https://www.googleapis.com/auth/gmail.send%20https://www.googleapis.com/auth/calendar%20https://www.googleapis.com/auth/documents%20https://www.googleapis.com/auth/drive.file&response_type=code&access_type=offline"

# Step 2: Exchange code for token
curl -X POST https://oauth2.googleapis.com/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=YOUR_CLIENT_ID" \
  -d "client_secret=YOUR_CLIENT_SECRET" \
  -d "code=AUTHORIZATION_CODE_FROM_STEP_1" \
  -d "grant_type=authorization_code" \
  -d "redirect_uri=urn:ietf:wg:oauth:2.0:oob"
```

### 5. **Test with SOHOAAS**
```bash
# Export the token
export OAUTH_TOKEN="ya29.your_actual_access_token_here"

# Run the OAuth testing script
./test_real_oauth.sh
```

## Required Scopes for SOHOAAS

- **Gmail**: `https://www.googleapis.com/auth/gmail.compose`, `https://www.googleapis.com/auth/gmail.send`
- **Calendar**: `https://www.googleapis.com/auth/calendar`
- **Docs**: `https://www.googleapis.com/auth/documents`
- **Drive**: `https://www.googleapis.com/auth/drive.file`

## Security Notes

- Access tokens expire (usually 1 hour)
- Use refresh tokens for long-term access
- Never commit tokens to version control
- Tokens have full access to your Google account - use test accounts

## Troubleshooting

- **Invalid credentials**: Check client ID/secret
- **Insufficient permissions**: Verify all required scopes are selected
- **Token expired**: Get a new token from OAuth playground
- **API not enabled**: Enable required APIs in Google Cloud Console
