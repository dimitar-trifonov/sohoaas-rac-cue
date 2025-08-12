# SOHOAAS Docker Deployment Guide

This guide covers deploying SOHOAAS using Docker containers with proper environment variable templating and nginx proxy configuration.

## Architecture Overview

SOHOAAS uses a 3-service Docker architecture:

1. **MCP Service** (port 8080) - OAuth2 authentication and Google Workspace API proxy
2. **SOHOAAS Backend** (port 8081) - 5-agent workflow automation engine
3. **SOHOAAS Frontend** (port 3000) - React app with nginx proxy for unified routing

## Prerequisites

- Docker and Docker Compose installed
- Google OAuth2 credentials configured
- OpenAI API key

## Environment Configuration

1. Copy the environment template:
```bash
cp .env.template .env
```

2. Fill in your actual values in `.env`:
```bash
# Google OAuth2 Configuration
GOOGLE_CLIENT_ID=your_actual_google_client_id
GOOGLE_CLIENT_SECRET=your_actual_google_client_secret
OAUTH_REDIRECT_URL=http://localhost:3000/api/auth/callback

# OpenAI API Configuration
OPENAI_API_KEY=your_actual_openai_api_key
```

## Docker Deployment

### Quick Start
```bash
# Build and start all services
docker-compose up --build

# Or run in background
docker-compose up --build -d
```

### Individual Service Management
```bash
# Build specific service
docker-compose build mcp-service
docker-compose build sohoaas-backend
docker-compose build sohoaas-frontend

# Start specific service
docker-compose up mcp-service
docker-compose up sohoaas-backend
docker-compose up sohoaas-frontend

# View logs
docker-compose logs -f mcp-service
docker-compose logs -f sohoaas-backend
docker-compose logs -f sohoaas-frontend
```

### Service Health Checks
```bash
# Check service status
docker-compose ps

# Test health endpoints
curl http://localhost:8080/health  # MCP Service
curl http://localhost:8081/health  # SOHOAAS Backend
curl http://localhost:3000/health  # SOHOAAS Frontend
```

## Environment Variable Templating

### Nginx Configuration
The frontend uses `nginx.conf.template` with environment variable substitution:
- `${NGINX_PORT}` - nginx listening port
- `${MCP_SERVICE_URL}` - MCP service URL for OAuth2 proxy
- `${BACKEND_SERVICE_URL}` - Backend service URL for API proxy

### Frontend API Service
The React app uses Vite environment variables:
- `VITE_PROXY_URL` - nginx proxy URL (for unified routing)
- `VITE_MCP_URL` - direct MCP URL (fallback)
- `VITE_BACKEND_URL` - direct backend URL (fallback)

### Runtime Variable Substitution
The frontend Dockerfile uses `envsubst` to replace template variables:
```bash
envsubst < /etc/nginx/templates/default.conf.template > /etc/nginx/conf.d/default.conf
```

## Network Architecture

```
User Browser
     ↓
SOHOAAS Frontend (nginx:3000)
     ├── /api/auth/* → MCP Service (8080)
     ├── /api/v1/* → SOHOAAS Backend (8081)
     └── /* → React App (static files)
```

### OAuth2 Flow
1. User clicks login → Frontend opens `/api/auth/login`
2. Nginx proxies to MCP Service OAuth2 endpoint
3. Google OAuth2 redirect → MCP Service handles callback
4. Frontend polls for auth completion
5. Authenticated requests include Bearer token

## Development vs Production

### Development (Local)
```bash
# Frontend development server
cd app/frontend
npm run dev  # Runs on port 5173 with HMR

# Backend development
cd app/backend
go run main.go  # Runs on port 8081

# MCP service
cd mcp/server
npm run dev  # Runs on port 8080
```

### Production (Docker)
```bash
# All services containerized
docker-compose up --build
# Frontend: nginx serving built React app on port 3000
# Backend: Go binary on port 8081
# MCP: Node.js service on port 8080
```

## Troubleshooting

### OAuth2 Issues
1. Check Google Cloud Console redirect URIs match `OAUTH_REDIRECT_URL`
2. Verify MCP service can reach Google OAuth2 endpoints
3. Check nginx proxy configuration for `/api/auth/*` routes

### Service Communication
1. Verify Docker network connectivity: `docker network inspect sohoaas-network`
2. Check service health: `docker-compose ps`
3. Review logs: `docker-compose logs -f [service-name]`

### Environment Variables
1. Verify `.env` file exists and has correct values
2. Check Docker container environment: `docker exec -it sohoaas-frontend env`
3. Verify nginx template substitution: `docker exec -it sohoaas-frontend cat /etc/nginx/conf.d/default.conf`

## Scaling and Production Deployment

### Kubernetes Deployment
The Docker setup can be adapted for Kubernetes:
1. Convert docker-compose services to Kubernetes Deployments
2. Use ConfigMaps for environment variables
3. Use Secrets for sensitive data (OAuth2 credentials, API keys)
4. Add Ingress controller for external access

### Cloud Deployment
For cloud deployment (GCP Cloud Run, AWS ECS, etc.):
1. Build and push images to container registry
2. Configure environment variables in cloud service
3. Set up load balancer for frontend service
4. Configure OAuth2 redirect URLs for production domain

## Security Considerations

1. **Environment Variables**: Never commit `.env` files with real credentials
2. **OAuth2 Redirect**: Ensure redirect URLs match exactly in Google Cloud Console
3. **CORS**: nginx configuration includes proper CORS headers
4. **Security Headers**: X-Frame-Options, X-Content-Type-Options, X-XSS-Protection
5. **Network Isolation**: Services communicate via Docker network, not exposed ports

## Monitoring and Logging

### Health Checks
All services include health check endpoints:
- MCP Service: `GET /health`
- SOHOAAS Backend: `GET /health`
- SOHOAAS Frontend: `GET /health`

### Logging
```bash
# View all logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f sohoaas-frontend
docker-compose logs -f sohoaas-backend
docker-compose logs -f mcp-service

# Follow logs with timestamps
docker-compose logs -f -t
```

This Docker deployment provides a production-ready SOHOAAS environment with proper service orchestration, environment variable templating, and unified routing through nginx proxy.
