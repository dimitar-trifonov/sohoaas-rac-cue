# Local Development with Firebase Auth & Google Cloud Storage

## Overview
Develop and test Firebase Authentication and Google Cloud Storage locally using environment variables, then seamlessly deploy to Cloud Run without code changes.

## Prerequisites
1. Google Cloud Project with billing enabled
2. Firebase project (can be same as GCP project)
3. Service account with appropriate permissions

## Setup Steps

### 1. Create Firebase Project & Service Account

```bash
# Create Firebase project (if not exists)
firebase projects:create sohoaas-demo

# Create service account
gcloud iam service-accounts create sohoaas-dev \
    --description="SOHOAAS development service account" \
    --display-name="SOHOAAS Dev"

# Grant necessary permissions
gcloud projects add-iam-policy-binding PROJECT_ID \
    --member="serviceAccount:sohoaas-dev@PROJECT_ID.iam.gserviceaccount.com" \
    --role="roles/storage.admin"

gcloud projects add-iam-policy-binding PROJECT_ID \
    --member="serviceAccount:sohoaas-dev@PROJECT_ID.iam.gserviceaccount.com" \
    --role="roles/firebase.admin"

# Create and download service account key
gcloud iam service-accounts keys create ./config/sohoaas-dev-key.json \
    --iam-account=sohoaas-dev@PROJECT_ID.iam.gserviceaccount.com
```

### 2. Setup Google Cloud Storage Bucket

```bash
# Create bucket for workflows
gsutil mb gs://sohoaas-workflows-dev

# Set bucket permissions
gsutil iam ch serviceAccount:sohoaas-dev@PROJECT_ID.iam.gserviceaccount.com:objectAdmin gs://sohoaas-workflows-dev
```

### 3. Environment Configuration

Create `.env.local` for development:

```bash
# Firebase Configuration
FIREBASE_PROJECT_ID=sohoaas-demo
FIREBASE_WEB_API_KEY=your_web_api_key
FIREBASE_AUTH_DOMAIN=sohoaas-demo.firebaseapp.com

# Google Cloud Configuration
GOOGLE_CLOUD_PROJECT=sohoaas-demo
GOOGLE_APPLICATION_CREDENTIALS=./config/sohoaas-dev-key.json
GCS_BUCKET_NAME=sohoaas-workflows-dev

# Environment
ENVIRONMENT=development
PORT=8080

# MCP Configuration
MCP_SERVER_URL=http://localhost:3000
```

Create `.env.production` for Cloud Run:

```bash
# Firebase Configuration (same values)
FIREBASE_PROJECT_ID=sohoaas-demo
FIREBASE_WEB_API_KEY=your_web_api_key
FIREBASE_AUTH_DOMAIN=sohoaas-demo.firebaseapp.com

# Google Cloud Configuration (uses Cloud Run service account)
GOOGLE_CLOUD_PROJECT=sohoaas-demo
# GOOGLE_APPLICATION_CREDENTIALS not needed - uses metadata service
GCS_BUCKET_NAME=sohoaas-workflows

# Environment
ENVIRONMENT=production
PORT=8080

# MCP Configuration
MCP_SERVER_URL=https://sohoaas-mcp-run.a.run.app
```

## Code Implementation

### Backend Configuration

```go
// internal/config/config.go
package config

import (
    "os"
    "log"
    "github.com/joho/godotenv"
)

type Config struct {
    Environment           string
    Port                 string
    FirebaseProjectID    string
    FirebaseWebAPIKey    string
    FirebaseAuthDomain   string
    GoogleCloudProject   string
    GoogleCredentialsPath string
    GCSBucketName        string
    MCPServerURL         string
}

func Load() *Config {
    // Load environment-specific .env file
    env := os.Getenv("ENVIRONMENT")
    if env == "" {
        env = "development"
    }
    
    envFile := fmt.Sprintf(".env.%s", env)
    if err := godotenv.Load(envFile); err != nil {
        log.Printf("Warning: Could not load %s: %v", envFile, err)
        // Try loading .env.local for development
        if env == "development" {
            godotenv.Load(".env.local")
        }
    }

    return &Config{
        Environment:           env,
        Port:                 getEnv("PORT", "8080"),
        FirebaseProjectID:    getEnv("FIREBASE_PROJECT_ID", ""),
        FirebaseWebAPIKey:    getEnv("FIREBASE_WEB_API_KEY", ""),
        FirebaseAuthDomain:   getEnv("FIREBASE_AUTH_DOMAIN", ""),
        GoogleCloudProject:   getEnv("GOOGLE_CLOUD_PROJECT", ""),
        GoogleCredentialsPath: getEnv("GOOGLE_APPLICATION_CREDENTIALS", ""),
        GCSBucketName:        getEnv("GCS_BUCKET_NAME", ""),
        MCPServerURL:         getEnv("MCP_SERVER_URL", "http://localhost:3000"),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

### Firebase Auth Service

```go
// internal/services/firebase_auth.go
package services

import (
    "context"
    "firebase.google.com/go/v4"
    "firebase.google.com/go/v4/auth"
    "google.golang.org/api/option"
)

type FirebaseAuthService struct {
    client *auth.Client
    config *config.Config
}

func NewFirebaseAuthService(cfg *config.Config) (*FirebaseAuthService, error) {
    ctx := context.Background()
    
    var app *firebase.App
    var err error
    
    if cfg.Environment == "development" && cfg.GoogleCredentialsPath != "" {
        // Local development with service account key
        opt := option.WithCredentialsFile(cfg.GoogleCredentialsPath)
        app, err = firebase.NewApp(ctx, &firebase.Config{
            ProjectID: cfg.FirebaseProjectID,
        }, opt)
    } else {
        // Production - uses Cloud Run service account
        app, err = firebase.NewApp(ctx, &firebase.Config{
            ProjectID: cfg.FirebaseProjectID,
        })
    }
    
    if err != nil {
        return nil, fmt.Errorf("failed to initialize Firebase app: %v", err)
    }

    client, err := app.Auth(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize Firebase Auth: %v", err)
    }

    return &FirebaseAuthService{
        client: client,
        config: cfg,
    }, nil
}

func (f *FirebaseAuthService) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
    token, err := f.client.VerifyIDToken(ctx, idToken)
    if err != nil {
        return nil, fmt.Errorf("failed to verify ID token: %v", err)
    }
    return token, nil
}

func (f *FirebaseAuthService) GetUser(ctx context.Context, uid string) (*auth.UserRecord, error) {
    user, err := f.client.GetUser(ctx, uid)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %v", err)
    }
    return user, nil
}
```

### Google Cloud Storage Service

```go
// internal/services/storage.go
package services

import (
    "context"
    "cloud.google.com/go/storage"
    "google.golang.org/api/option"
)

type StorageService struct {
    client *storage.Client
    bucket string
    config *config.Config
}

func NewStorageService(cfg *config.Config) (*StorageService, error) {
    ctx := context.Background()
    
    var client *storage.Client
    var err error
    
    if cfg.Environment == "development" && cfg.GoogleCredentialsPath != "" {
        // Local development with service account key
        client, err = storage.NewClient(ctx, option.WithCredentialsFile(cfg.GoogleCredentialsPath))
    } else {
        // Production - uses Cloud Run service account
        client, err = storage.NewClient(ctx)
    }
    
    if err != nil {
        return nil, fmt.Errorf("failed to create storage client: %v", err)
    }

    return &StorageService{
        client: client,
        bucket: cfg.GCSBucketName,
        config: cfg,
    }, nil
}

func (s *StorageService) SaveWorkflow(ctx context.Context, userID, workflowID string, data []byte) error {
    objectName := fmt.Sprintf("users/%s/workflows/%s/workflow.json", userID, workflowID)
    
    obj := s.client.Bucket(s.bucket).Object(objectName)
    writer := obj.NewWriter(ctx)
    defer writer.Close()
    
    if _, err := writer.Write(data); err != nil {
        return fmt.Errorf("failed to write workflow: %v", err)
    }
    
    return nil
}

func (s *StorageService) LoadWorkflow(ctx context.Context, userID, workflowID string) ([]byte, error) {
    objectName := fmt.Sprintf("users/%s/workflows/%s/workflow.json", userID, workflowID)
    
    obj := s.client.Bucket(s.bucket).Object(objectName)
    reader, err := obj.NewReader(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to create reader: %v", err)
    }
    defer reader.Close()
    
    data, err := io.ReadAll(reader)
    if err != nil {
        return nil, fmt.Errorf("failed to read workflow: %v", err)
    }
    
    return data, nil
}
```

### Updated Auth Middleware

```go
// internal/middleware/auth.go
package middleware

import (
    "context"
    "net/http"
    "strings"
    "github.com/gin-gonic/gin"
)

func FirebaseAuthMiddleware(firebaseAuth *services.FirebaseAuthService) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }

        if !strings.HasPrefix(authHeader, "Bearer ") {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
            c.Abort()
            return
        }

        idToken := strings.TrimPrefix(authHeader, "Bearer ")
        
        // Verify Firebase ID token
        token, err := firebaseAuth.VerifyIDToken(context.Background(), idToken)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }

        // Get user info
        user, err := firebaseAuth.GetUser(context.Background(), token.UID)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to get user info"})
            c.Abort()
            return
        }

        // Store user info in context
        c.Set("firebase_uid", token.UID)
        c.Set("user_email", user.Email)
        c.Set("user", user)
        
        c.Next()
    }
}
```

## Local Development Workflow

### 1. Start Development Environment

```bash
# Set environment
export ENVIRONMENT=development

# Start backend
cd app/backend
go run main.go

# Start MCP server
cd mcp/server/backend
go run main.go

# Start frontend
cd app/frontend
npm run dev
```

### 2. Test Firebase Auth Locally

```bash
# Install Firebase CLI
npm install -g firebase-tools

# Login to Firebase
firebase login

# Test Firebase Auth
firebase auth:export users.json --project sohoaas-demo
```

### 3. Test GCS Operations

```bash
# Test bucket access
gsutil ls gs://sohoaas-workflows-dev

# Upload test file
echo "test" | gsutil cp - gs://sohoaas-workflows-dev/test.txt

# Download test file
gsutil cp gs://sohoaas-workflows-dev/test.txt -
```

## Deployment Transition

### 1. Build for Production

```bash
# Build backend
cd app/backend
CGO_ENABLED=0 GOOS=linux go build -o sohoaas-backend main.go

# Build MCP server
cd mcp/server/backend
CGO_ENABLED=0 GOOS=linux go build -o sohoaas-mcp main.go
```

### 2. Deploy to Cloud Run

```bash
# Deploy backend
gcloud run deploy sohoaas-backend \
    --image gcr.io/PROJECT_ID/sohoaas-backend:latest \
    --platform managed \
    --region us-central1 \
    --set-env-vars ENVIRONMENT=production \
    --set-env-vars FIREBASE_PROJECT_ID=sohoaas-demo \
    --set-env-vars GCS_BUCKET_NAME=sohoaas-workflows \
    --allow-unauthenticated

# Deploy MCP server
gcloud run deploy sohoaas-mcp \
    --image gcr.io/PROJECT_ID/sohoaas-mcp:latest \
    --platform managed \
    --region us-central1 \
    --set-env-vars ENVIRONMENT=production \
    --set-env-vars FIREBASE_PROJECT_ID=sohoaas-demo \
    --allow-unauthenticated
```

## Benefits of This Approach

1. **Seamless Transition**: Same code works locally and in production
2. **Environment Isolation**: Separate buckets and configs for dev/prod
3. **No Code Changes**: Only environment variables change between environments
4. **Local Testing**: Full Firebase and GCS functionality available locally
5. **Security**: Service account keys only used locally, Cloud Run uses metadata service
6. **Cost Effective**: Separate dev resources prevent production costs during development

## Security Notes

- Never commit service account keys to version control
- Use separate buckets for development and production
- Implement proper IAM policies for production
- Rotate service account keys regularly
- Use Firebase Auth rules to restrict access in production
