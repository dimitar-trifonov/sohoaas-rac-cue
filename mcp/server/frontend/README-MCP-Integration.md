# RAC Service Proxies - Frontend with MCP Integration

This frontend now supports both REST API and MCP (Model Context Protocol) interfaces for interacting with Google Workspace services.

## Features

### REST API Interface (Original)
- **Service Proxy Dashboard**: View available providers and services
- **Proxy Tester**: Test REST API endpoints with custom payloads
- **OAuth2 Authentication**: Secure authentication with Google Workspace

### MCP Protocol Interface (New)
- **WebSocket Connection**: Real-time MCP protocol communication
- **Resource Discovery**: Browse available MCP resources
- **Tool Execution**: Execute MCP tools with JSON arguments
- **Connection Logging**: Monitor MCP protocol messages

## Usage

### 1. Start the Backend Server
```bash
cd ../backend
go run .
```

The server will start on `http://localhost:8080` with both REST and MCP endpoints.

### 2. Start the Frontend Development Server
```bash
npm run dev
```

The frontend will be available at `http://localhost:5173`.

### 3. Authentication
1. Click "Get Auth URL" in the top-right corner
2. Follow the OAuth2 flow to authenticate with Google
3. Return to the dashboard once authenticated

### 4. Using the REST API Interface
- Select the "REST API Interface" tab
- Choose a provider (e.g., "workspace")
- View available services and test API endpoints

### 5. Using the MCP Protocol Interface
- Select the "MCP Protocol Interface" tab
- Click "Connect to MCP Server" (requires authentication)
- Browse available resources and tools
- Execute tools with JSON arguments

## MCP Tool Examples

### Send Email via Gmail
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

### Share Document
```json
{
  "file_id": "1234567890abcdef",
  "email": "collaborator@example.com",
  "role": "writer"
}
```

### Create Calendar Event
```json
{
  "title": "MCP Integration Meeting",
  "description": "Discuss MCP protocol implementation",
  "start_time": "2024-01-15T10:00:00Z",
  "end_time": "2024-01-15T11:00:00Z"
}
```

## Available Endpoints

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

## Architecture

The frontend uses a tabbed interface to switch between:

1. **REST API Interface**: Traditional HTTP REST API calls
2. **MCP Protocol Interface**: WebSocket-based MCP protocol communication

Both interfaces share the same authentication system and can access the same Google Workspace services through different protocols.

## Development

### Building for Production
```bash
npm run build
```

### Type Checking
```bash
npm run type-check
```

### Linting
```bash
npm run lint
```

## Environment Variables

The backend requires these environment variables:
- `GOOGLE_CLIENT_ID` - Google OAuth2 client ID
- `GOOGLE_CLIENT_SECRET` - Google OAuth2 client secret
- `OAUTH_REDIRECT_URL` - OAuth2 redirect URL
- `PORT` - Server port (default: 8080)

## Next Steps

1. **Genkit Integration**: Add Genkit orchestration flows
2. **Workflow Monitoring**: Add real-time workflow status tracking
3. **Enhanced Error Handling**: Improve error messages and recovery
4. **Testing**: Add comprehensive unit and integration tests
