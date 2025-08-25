# SOHOAAS Multi-User Deployment Architecture

## Overview
Transform SOHOAAS from single-user PoC to multi-tenant demo platform for client presentations.

## Current Architecture Limitations
- Single hardcoded user (`mock_user_123`)
- Local file storage in `./generated_workflows/`
- No user session management
- Static OAuth token handling

## Proposed Multi-User Architecture

### 1. Authentication: Firebase Auth + Gmail OAuth2
```
Frontend → Firebase Auth → Gmail OAuth2 → Backend JWT Validation
```

**Implementation**:
- Firebase Authentication with Google provider
- Email allowlist for client access control
- JWT token validation in backend middleware
- User-specific OAuth token storage

### 2. Storage: Google Cloud Storage
```
gs://sohoaas-workflows/
├── users/
│   ├── user1@gmail.com/
│   │   ├── workflows/
│   │   │   ├── 20250818_123456/
│   │   │   │   ├── workflow.json
│   │   │   │   ├── workflow.cue
│   │   │   │   └── metadata/
│   │   │   └── 20250818_134567/
│   │   └── sessions/
│   └── user2@gmail.com/
│       ├── workflows/
│       └── sessions/
```

**Benefits**:
- User isolation and data security
- Scalable storage with GCP integration
- Cost-effective for structured workflow data
- Fine-grained IAM controls

### 3. Deployment: Cloud Run Services

#### Backend Service
```yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: sohoaas-backend
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/maxScale: "10"
        run.googleapis.com/memory: "1Gi"
        run.googleapis.com/cpu: "1000m"
    spec:
      containers:
      - image: gcr.io/PROJECT_ID/sohoaas-backend:latest
        ports:
        - containerPort: 8080
        env:
        - name: FIREBASE_PROJECT_ID
          value: "sohoaas-demo"
        - name: GCS_BUCKET
          value: "sohoaas-workflows"
        - name: MCP_SERVER_URL
          value: "https://sohoaas-mcp-run.a.run.app"
```

#### MCP Service
```yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: sohoaas-mcp
spec:
  template:
    spec:
      containers:
      - image: gcr.io/PROJECT_ID/sohoaas-mcp:latest
        ports:
        - containerPort: 3000
        env:
        - name: FIREBASE_PROJECT_ID
          value: "sohoaas-demo"
```

### 4. User Session Management

#### Session Structure
```go
type UserSession struct {
    UserID       string                 `json:"user_id"`
    Email        string                 `json:"email"`
    FirebaseUID  string                 `json:"firebase_uid"`
    OAuthTokens  map[string]interface{} `json:"oauth_tokens"`
    WorkspaceID  string                 `json:"workspace_id"`
    CreatedAt    time.Time              `json:"created_at"`
    ExpiresAt    time.Time              `json:"expires_at"`
}
```

#### Multi-Tenant Workflow Storage
```go
type WorkflowManager struct {
    gcsClient *storage.Client
    bucket    string
}

func (wm *WorkflowManager) SaveWorkflow(userID string, workflow *Workflow) error {
    path := fmt.Sprintf("users/%s/workflows/%s/workflow.json", userID, workflow.ID)
    // Save to GCS with user isolation
}

func (wm *WorkflowManager) ListUserWorkflows(userID string) ([]*Workflow, error) {
    prefix := fmt.Sprintf("users/%s/workflows/", userID)
    // List user-specific workflows from GCS
}
```

## Implementation Plan

### Phase 1: Firebase Authentication
1. **Setup Firebase Project**
   - Create Firebase project `sohoaas-demo`
   - Enable Authentication with Google provider
   - Configure OAuth2 consent screen

2. **Update Frontend Auth**
   - Replace custom auth with Firebase SDK
   - Implement Gmail OAuth2 flow
   - Add user email allowlist validation

3. **Update Backend Middleware**
   - Replace mock auth with Firebase JWT validation
   - Extract user info from Firebase token
   - Implement user session management

### Phase 2: Multi-User Storage
1. **Setup Google Cloud Storage**
   - Create bucket `sohoaas-workflows`
   - Configure IAM for service accounts
   - Implement user-isolated directory structure

2. **Update Workflow Management**
   - Replace local file storage with GCS
   - Add user-specific workflow paths
   - Implement user data isolation

### Phase 3: Cloud Run Deployment
1. **Containerize Services**
   - Create Dockerfiles for backend and MCP
   - Setup CI/CD with Cloud Build
   - Configure environment variables

2. **Deploy to Cloud Run**
   - Deploy backend service
   - Deploy MCP service
   - Configure service-to-service authentication

### Phase 4: Demo Configuration
1. **Client Access Management**
   - Setup email allowlist in Firebase
   - Create demo user accounts
   - Configure OAuth2 scopes for Google Workspace

2. **Monitoring & Logging**
   - Setup Cloud Logging
   - Configure error monitoring
   - Add usage analytics

## Security Considerations

### User Data Isolation
- Each user's workflows stored in separate GCS directories
- Firebase UID used as primary user identifier
- OAuth tokens encrypted and user-specific

### Access Control
- Email allowlist enforced at Firebase level
- Service-to-service authentication between backend and MCP
- GCS IAM policies for data access control

### OAuth2 Token Management
- User OAuth tokens stored securely per session
- Token refresh handled automatically
- Scoped access to Google Workspace APIs

## Cost Estimation (Monthly)

### Firebase
- Authentication: Free tier (up to 10K users)
- Firestore: $0.18/100K reads (if used for sessions)

### Cloud Run
- Backend: ~$20-50/month (based on usage)
- MCP Service: ~$15-30/month
- Cold starts optimized for demo usage

### Google Cloud Storage
- Workflow storage: ~$5-15/month (based on data volume)
- Network egress: Minimal for demo usage

**Total Estimated Cost**: $40-95/month for demo deployment

## Demo Deployment Checklist

- [ ] Create Firebase project and configure authentication
- [ ] Setup Google Cloud Storage bucket with proper IAM
- [ ] Update frontend to use Firebase Auth
- [ ] Update backend middleware for JWT validation
- [ ] Implement user-specific workflow storage
- [ ] Create Docker containers for services
- [ ] Deploy to Cloud Run with proper configuration
- [ ] Setup client email allowlist
- [ ] Test multi-user scenarios
- [ ] Configure monitoring and logging
