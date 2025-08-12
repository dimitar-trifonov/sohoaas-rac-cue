# RAC Service Proxies - Docker Deployment

This Docker Compose setup provides a complete containerized environment for the RAC Service Proxies system with both REST API and MCP Protocol support.

## Architecture

The system consists of:

- **Backend Service** (Go): REST API + MCP WebSocket server
- **Frontend Service** (React): Web interface with REST and MCP client capabilities
- **Redis** (Optional): Caching layer
- **PostgreSQL** (Optional): Persistent storage

## Quick Start

### 1. Environment Setup

Copy the example environment file and configure your Google OAuth2 credentials:

```bash
cp .env.example .env
```

Edit `.env` and set your Google OAuth2 credentials:

```bash
GOOGLE_CLIENT_ID=your_actual_google_client_id
GOOGLE_CLIENT_SECRET=your_actual_google_client_secret
OAUTH_REDIRECT_URL=http://localhost:8080/api/auth/callback
```

### 2. Basic Deployment (Backend + Frontend)

Start the core services:

```bash
docker-compose up -d backend frontend
```

This will start:
- Backend server on `http://localhost:8080`
- Frontend interface on `http://localhost:3002`
- MCP WebSocket endpoint at `ws://localhost:8080/mcp`

### 3. Full Deployment (with Redis and PostgreSQL)

Start all services including optional databases:

```bash
docker-compose --profile redis --profile postgres up -d
```

## Available Services

### Core Services

- **Backend**: `http://localhost:8080`
  - REST API endpoints
  - MCP WebSocket endpoint at `/mcp`
  - Health check at `/health`
  
- **Frontend**: `http://localhost:3002`
  - REST API Interface tab
  - MCP Protocol Interface tab
  - OAuth2 authentication flow

### Optional Services

- **Redis**: `localhost:6379` (with `--profile redis`)
- **PostgreSQL**: `localhost:5432` (with `--profile postgres`)

## API Endpoints

### REST API Endpoints
- `GET /health` - Health check
- `POST /workflow/execute` - Execute workflow
- `GET /api/providers` - List providers
- `GET /api/providers/:provider/services` - List services
- `GET /api/services` - Service metadata
- `GET /api/auth/login` - OAuth login
- `GET /api/auth/callback` - OAuth callback
- `GET /api/auth/token` - Get current token

### MCP Protocol Endpoint
- `GET /mcp` - WebSocket endpoint for MCP protocol

## Usage

### 1. Access the Frontend
Navigate to `http://localhost:3002` in your browser.

### 2. Authenticate
1. Click "Get Auth URL" in the top-right corner
2. Complete the Google OAuth2 flow
3. Return to the dashboard

### 3. Use REST API Interface
- Select the "REST API Interface" tab
- Choose a provider and service
- Test API endpoints with custom payloads

### 4. Use MCP Protocol Interface
- Select the "MCP Protocol Interface" tab
- Click "Connect to MCP Server"
- Browse available resources and tools
- Execute tools with JSON arguments

## MCP Tool Examples

### Send Email
```json
{
  "to": "recipient@example.com",
  "subject": "Test Email from MCP",
  "body": "This is a test email sent via the MCP protocol."
}
```

### Create Document from Template
```json
{
  "template_id": "1234567890abcdef",
  "title": "New Document from MCP",
  "replacements": {
    "{{CLIENT_NAME}}": "John Doe",
    "{{PROJECT_NAME}}": "MCP Integration Project"
  }
}
```

## Container Management

### View Logs
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f backend
docker-compose logs -f frontend
```

### Restart Services
```bash
# Restart all
docker-compose restart

# Restart specific service
docker-compose restart backend
```

### Stop Services
```bash
# Stop all
docker-compose down

# Stop and remove volumes
docker-compose down -v
```

### Update and Rebuild
```bash
# Rebuild and restart
docker-compose up -d --build

# Force rebuild without cache
docker-compose build --no-cache
docker-compose up -d
```

## Health Checks

All services include health checks:

- **Backend**: Checks `/health` endpoint
- **Frontend**: Checks nginx availability
- **Redis**: Checks Redis ping
- **PostgreSQL**: Checks database connectivity

View health status:
```bash
docker-compose ps
```

## Networking

All services communicate through the `rac-service-proxy-network` bridge network:

- Backend ↔ Frontend: Internal communication
- Frontend → Backend: API calls and WebSocket connections
- Optional database connections

## Volumes

Persistent data volumes:
- `rac-service-proxy-redis-data`: Redis data persistence
- `rac-service-proxy-postgres-data`: PostgreSQL data persistence

## Troubleshooting

### Common Issues

1. **OAuth2 Redirect URL Mismatch**
   - Ensure `OAUTH_REDIRECT_URL` matches your Google Cloud Console configuration
   - For Docker deployment, use `http://localhost:8080/api/auth/callback`

2. **WebSocket Connection Failed**
   - Check that backend is running and healthy
   - Verify MCP endpoint is accessible at `ws://localhost:8080/mcp`

3. **Build Failures**
   - Clear Docker cache: `docker system prune -a`
   - Rebuild without cache: `docker-compose build --no-cache`

### Debug Mode

Run services in debug mode:
```bash
# Backend with debug logging
docker-compose run --rm -e GIN_MODE=debug backend

# View detailed logs
docker-compose logs -f --tail=100
```

## Production Deployment

For production deployment:

1. **Update Environment Variables**
   - Set production OAuth2 redirect URLs
   - Configure secure database passwords
   - Set `GIN_MODE=release`

2. **Use External Databases**
   - Configure external Redis and PostgreSQL instances
   - Remove database services from docker-compose.yaml

3. **Add Reverse Proxy**
   - Use nginx or Traefik for SSL termination
   - Configure proper domain names

4. **Security Considerations**
   - Use Docker secrets for sensitive data
   - Enable container security scanning
   - Regular security updates

## Development

For development with live reloading:

```bash
# Mount source code volumes
docker-compose -f docker-compose.yaml -f docker-compose.dev.yaml up -d
```

This enables:
- Live code reloading for backend changes
- Hot module replacement for frontend changes
