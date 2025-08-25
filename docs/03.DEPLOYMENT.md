# SOHOAAS Deployment Guide

Comprehensive deployment guide covering both Docker containerization and cloud deployment options.

## Table of Contents
- [Docker Deployment](#docker-deployment)
- [Cloud Run Deployment](#cloud-run-deployment)
- [Environment Configuration](#environment-configuration)
- [Security Considerations](#security-considerations)

## Docker Deployment

### Architecture Overview
SOHOAAS uses a 3-service Docker architecture:
1. **MCP Service** (port 8080) - OAuth2 authentication and Google Workspace API proxy
2. **SOHOAAS Backend** (port 8081) - 4-agent workflow automation engine
3. **SOHOAAS Frontend** (port 3000) - React app with nginx proxy for unified routing

### Quick Start
```bash
# Copy environment template
cp .env.template .env

# Fill in your actual values in .env
# Build and start all services
docker-compose up --build

# Or run in background
docker-compose up --build -d
```

### Environment Configuration for Docker
```bash
# Google OAuth2 Configuration
GOOGLE_CLIENT_ID=your_actual_google_client_id
GOOGLE_CLIENT_SECRET=your_actual_google_client_secret
OAUTH_REDIRECT_URL=http://localhost:3000/api/auth/callback

# OpenAI API Configuration
OPENAI_API_KEY=your_actual_openai_api_key
```

### Network Architecture
```
User Browser
     ↓
SOHOAAS Frontend (nginx:3000)
     ├── /api/auth/* → MCP Service (8080)
     ├── /api/v1/* → SOHOAAS Backend (8081)
     └── /* → React App (static files)
```

## 🚀 Production Deployment (Google Cloud Run)

### Prerequisites
- Google Cloud Project with billing enabled
- Cloud Run API enabled
- Container Registry or Artifact Registry access
- Firebase project for Firestore
- OpenAI API key

### Docker Service Management
```bash
# Individual service management
docker-compose build mcp-service
docker-compose build sohoaas-backend
docker-compose build sohoaas-frontend

# View logs
docker-compose logs -f mcp-service
docker-compose logs -f sohoaas-backend
docker-compose logs -f sohoaas-frontend

# Health checks
curl http://localhost:8080/health  # MCP Service
curl http://localhost:8081/health  # SOHOAAS Backend
curl http://localhost:3000/health  # SOHOAAS Frontend
```

## Cloud Run Deployment

### Environment Variables for Cloud Run
```bash
# Required for production
export GOOGLE_CLOUD_PROJECT="your-project-id"
export FIREBASE_PROJECT_ID="your-firebase-project"
export OPENAI_API_KEY="your-openai-key"
export MCP_BASE_URL="https://your-mcp-server.run.app"

# OAuth2 Configuration
export GOOGLE_CLIENT_ID="your-oauth-client-id"
export GOOGLE_CLIENT_SECRET="your-oauth-client-secret"
export OAUTH_REDIRECT_URI="https://your-app.run.app/auth/callback"

# Genkit Configuration
export GENKIT_REFLECTION_PORT="3101"
export GENKIT_ENVIRONMENT="prod"
```

### Build and Deploy
```bash
# Build the container
docker build -t gcr.io/$GOOGLE_CLOUD_PROJECT/sohoaas-backend ./app/backend

# Push to Container Registry
docker push gcr.io/$GOOGLE_CLOUD_PROJECT/sohoaas-backend

# Deploy to Cloud Run
gcloud run deploy sohoaas-backend \
  --image gcr.io/$GOOGLE_CLOUD_PROJECT/sohoaas-backend \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --port 8080 \
  --memory 1Gi \
  --cpu 1 \
  --concurrency 80 \
  --max-instances 10 \
  --set-env-vars GOOGLE_CLOUD_PROJECT=$GOOGLE_CLOUD_PROJECT,FIREBASE_PROJECT_ID=$FIREBASE_PROJECT_ID,OPENAI_API_KEY=$OPENAI_API_KEY,MCP_BASE_URL=$MCP_BASE_URL
```

### Service Configuration
Based on RaC specifications:

```yaml
# cloud-run-service.yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: sohoaas-backend
  annotations:
    run.googleapis.com/ingress: all
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/maxScale: "10"
        autoscaling.knative.dev/minScale: "0"
        run.googleapis.com/cpu-throttling: "false"
        run.googleapis.com/memory: "1Gi"
    spec:
      containerConcurrency: 80
      containers:
      - image: gcr.io/PROJECT_ID/sohoaas-backend
        ports:
        - containerPort: 8080
        env:
        - name: GENKIT_REFLECTION_PORT
          value: "3101"
        resources:
          limits:
            cpu: "1"
            memory: "1Gi"
```

## 🏠 Local Development Setup

### Prerequisites
- Go 1.21+
- Node.js 18+
- Docker (for MCP servers)
- Firebase CLI
- CUE CLI

### Backend Setup
```bash
# Clone repository
git clone <repository-url>
cd sohoaas

# Install Go dependencies
cd app/backend
go mod download

# Set up environment variables
cp .env.example .env
# Edit .env with your API keys

# Run backend
go run main.go
```

### Environment Variables (.env)
```bash
# Development configuration
OPENAI_API_KEY=your-openai-key
FIREBASE_PROJECT_ID=your-firebase-project
GOOGLE_CLIENT_ID=your-oauth-client-id
GOOGLE_CLIENT_SECRET=your-oauth-client-secret

# Local MCP server
MCP_BASE_URL=http://localhost:8080

# Genkit development
GENKIT_ENVIRONMENT=dev
GENKIT_REFLECTION_PORT=3101
```

### Frontend Setup
```bash
# Install frontend dependencies
cd app/frontend
npm install

# Start development server
npm run dev
```

### MCP Server Setup
```bash
# Start MCP servers (if running locally)
cd mcp/server
docker-compose up -d
```

## 🔧 Development Workflow

### 1. RaC-First Development
- Always update RaC specifications first (`rac/system.cue`)
- Validate with CUE: `cue vet rac/system.cue`
- Generate code from RaC specifications

### 2. Agent Development
```bash
# Test individual agents
cd app/backend
go test ./internal/agents/intent_gatherer -v
go test ./internal/agents/intent_analyst -v
go test ./internal/agents/workflow_generator -v
go test ./internal/agents/workflow_validator -v
```

### 3. Service Development  
```bash
# Test deterministic services
go test ./internal/services/personal_capabilities -v
go test ./internal/services/cue_generator -v
go test ./internal/services/workflow_executor -v
go test ./internal/services/agent_manager -v
```

### 4. Integration Testing
```bash
# Full system integration test
go test ./internal/integration -v

# MCP integration test
go test ./internal/mcp -v
```

## Security Considerations

### Docker Security
1. **Environment Variables**: Never commit `.env` files with real credentials
2. **OAuth2 Redirect**: Ensure redirect URLs match exactly in Google Cloud Console
3. **CORS**: nginx configuration includes proper CORS headers
4. **Security Headers**: X-Frame-Options, X-Content-Type-Options, X-XSS-Protection
5. **Network Isolation**: Services communicate via Docker network, not exposed ports

### Cloud Run Security
1. **IAM Roles**: Use least-privilege service accounts
2. **Secrets Management**: Store sensitive data in Google Secret Manager
3. **VPC Connector**: Use VPC for internal service communication
4. **Authentication**: Enable Cloud Run authentication for production
5. **HTTPS Only**: Enforce HTTPS for all external traffic

## Monitoring and Logging

### Health Checks
All services include health check endpoints:
- MCP Service: `GET /health`
- SOHOAAS Backend: `GET /health`
- SOHOAAS Frontend: `GET /health`

### Docker Logging
```bash
# View all logs
docker-compose logs -f

# View specific service logs with timestamps
docker-compose logs -f -t sohoaas-backend
```

### Cloud Run Monitoring
```bash
# View Cloud Run logs
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=sohoaas-backend"

# Monitor metrics
gcloud monitoring metrics list --filter="resource.type=cloud_run_revision"
```

## 📊 Monitoring and Observability

### Genkit Reflection Server
- **URL**: `http://localhost:3101` (dev) or `https://your-app.run.app:3101` (prod)
- **Features**: LLM call tracing, flow visualization, performance metrics

### Firebase Console
- **Firestore**: Monitor state storage and event logs
- **Authentication**: Track OAuth2 flows
- **Functions**: Monitor any Firebase functions

### Cloud Run Metrics
- **CPU/Memory Usage**: Monitor resource consumption
- **Request Latency**: Track response times
- **Error Rates**: Monitor failed requests
- **Concurrency**: Track concurrent request handling

### Agent Logging
Based on `rac/observability/agent_logging_schema.cue`:

```json
{
  "agent_id": "intent_gatherer",
  "session_id": "session-123",
  "event_type": "workflow_intent_discovered",
  "timestamp": "2024-01-15T10:30:00Z",
  "llm_provider": "openai",
  "token_usage": {
    "prompt_tokens": 150,
    "completion_tokens": 75,
    "total_tokens": 225
  },
  "execution_time_ms": 1250,
  "status": "success"
}
```

## 🔒 Security Configuration

### OAuth2 Setup
1. **Google Cloud Console**: Create OAuth2 credentials
2. **Authorized Origins**: Add your domain(s)
3. **Redirect URIs**: Configure callback URLs
4. **Scopes**: gmail, docs, calendar, drive

### API Key Management
- Store in Google Secret Manager (production)
- Use environment variables (development)
- Rotate keys regularly
- Monitor usage and quotas

### Network Security
- **Cloud Run**: Configure ingress rules
- **Firestore**: Set up security rules
- **MCP Servers**: Secure API endpoints
- **CORS**: Configure allowed origins

## 🧪 Testing in Production

### Health Checks
```bash
# Backend health
curl https://your-app.run.app/health

# Agent Manager status
curl https://your-app.run.app/api/agents/status

# MCP connectivity
curl https://your-app.run.app/api/mcp/health
```

### Load Testing
```bash
# Install artillery
npm install -g artillery

# Run load test
artillery run load-test.yml
```

### Rollback Strategy
```bash
# Rollback to previous revision
gcloud run services update-traffic sohoaas-backend --to-revisions=PREVIOUS=100

# Zero-downtime deployment
gcloud run services update-traffic sohoaas-backend --to-revisions=LATEST=50,PREVIOUS=50
```

## 📋 Troubleshooting

### Common Issues
1. **Agent timeout**: Increase Cloud Run timeout
2. **Memory limit**: Scale up memory allocation
3. **OAuth errors**: Verify redirect URIs
4. **MCP connectivity**: Check network policies
5. **CUE validation**: Verify schema compliance

### Debug Commands
```bash
# View logs
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=sohoaas-backend"

# Check agent status
curl -H "Authorization: Bearer $(gcloud auth print-access-token)" https://your-app.run.app/api/debug/agents

# Validate RaC specifications
cue vet rac/system.cue rac/agents/*.cue rac/services/*.cue
```
