# SOHOAAS Backend

SOHOAAS (Small Office/Home Office Automation as a Service) backend implementation with 4-agent architecture managed by Agent Manager.

## Architecture

### Core Agents
- **Personal Capabilities Agent** - MCP service discovery and capability mapping
- **Intent Gatherer Agent** - Multi-turn workflow discovery conversations  
- **Intent Analyst Agent** - Intent validation and parameter extraction
- **Workflow Generator Agent** - Deterministic workflow CUE file generation

### Agent Manager
- Deterministic orchestration without LLM decisions
- Event-driven agent coordination
- REST API endpoints for frontend integration

## Technology Stack
- **Backend**: Golang + Gin web framework
- **LLM Integration**: Google Genkit with Gemini models
- **Authentication**: OAuth2 token validation via MCP service
- **API Protocol**: REST JSON endpoints

## Environment Setup

1. Copy environment configuration:
```bash
cp .env.example .env
```

2. Configure required environment variables:
```env
PORT=8080
OPENAI_API_KEY=your_google_genai_api_key_here
MCP_BASE_URL=http://localhost:3002
MCP_AUTH_ENDPOINT=/api/auth/validate
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
ENVIRONMENT=development
```

## Running the Backend

1. Install dependencies:
```bash
go mod tidy
```

2. Run the server:
```bash
go run main.go
```

The server will start on port 8080 (or the port specified in PORT environment variable).

## API Endpoints

### Public Endpoints
- `GET /health` - Health check

### Protected Endpoints (require OAuth2 token)
- `GET /api/v1/agents` - List all available agents
- `GET /api/v1/capabilities` - Get user's personal automation capabilities
- `POST /api/v1/workflow/discover` - Start workflow discovery conversation
- `POST /api/v1/workflow/continue` - Continue workflow discovery conversation
- `POST /api/v1/intent/analyze` - Analyze and validate workflow intent
- `POST /api/v1/workflow/generate` - Generate deterministic workflow from validated intent
- `POST /api/v1/workflow/execute` - Execute generated workflow
- `GET /api/v1/services` - Get user's connected MCP services

## Authentication

All protected endpoints require an `Authorization: Bearer <token>` header. Tokens are validated against the MCP service configured in `MCP_BASE_URL`.

## Agent Flow

1. **User Authentication** → MCP OAuth2 validation
2. **Personal Capabilities** → Discover user's connected services and automation capabilities
3. **Workflow Discovery** → Multi-turn conversation to identify automation patterns
4. **Intent Analysis** → Validate and extract parameters from discovered workflow intent
5. **Workflow Generation** → Create deterministic CUE workflow specification
6. **Workflow Execution** → Execute workflow via MCP services

## Development

The backend follows a clean architecture pattern:
- `internal/api/` - REST API handlers and routes
- `internal/config/` - Configuration management
- `internal/manager/` - Agent Manager orchestration
- `internal/middleware/` - Authentication and CORS middleware
- `internal/services/` - External service integrations (Genkit, MCP)
- `internal/types/` - Type definitions and data structures
- `prompts/` - LLM prompt templates for each agent

## Integration with Frontend

The backend provides JSON REST APIs designed to be consumed by a React-based frontend through nginx proxy with MCP OAuth2 authentication.
