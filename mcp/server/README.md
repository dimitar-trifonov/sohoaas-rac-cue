# Service Proxies - Multi-Provider Workflow Engine

This module provides a unified interface for interacting with multiple service providers (Google Workspace, Office 365, etc.) through a standardized proxy system. It enables seamless workflow orchestration across different platforms.

## Project Structure

```
service-proxies/
├── backend/          # Go backend service proxies
│   ├── main.go       # Main server application
│   ├── providers/    # Service provider implementations
│   ├── workflow/     # Workflow orchestration engine
│   └── Dockerfile    # Backend container configuration
├── frontend/         # React management interface
│   ├── src/          # React TypeScript source code
│   ├── public/       # Static assets
│   └── Dockerfile    # Frontend container configuration
└── README.md         # This file
```

## Quick Start

### Prerequisites
- Docker and Docker Compose
- Google OAuth2 credentials (see `../credentials/README.md`)

### Running the Service Proxies

#### Option 1: Using the Service Proxy Script (Recommended)

1. **Setup credentials**:
   ```bash
   # Place your Google OAuth2 credentials
   cp path/to/your/credentials.json ../credentials/google-credentials.json
   ```

2. **Start services**:
   ```bash
   # From the service-proxies directory
   ./start-service-proxies.sh
   ```

#### Option 2: Using Docker Compose Directly

```bash
# From the service-proxies directory
docker-compose up --build -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

#### Option 3: From Project Root (Legacy)

```bash
# From the project root
./start-services.sh
```

3. **Access the applications**:
   - **Backend API**: http://localhost:8080
   - **Frontend Management**: http://localhost:3002
   - **API Documentation**: http://localhost:8080/providers
   
   **Note**: Ports 3000 and 3001 are reserved for the main SOHOaaS application.

### Development

**Backend Development**:
```bash
cd backend
go run main.go
```

**Frontend Development**:
```bash
cd frontend
npm install
npm run dev
```

### Optional Services

The docker-compose.yaml includes optional services that can be enabled:

**Redis (for caching)**:
```bash
docker-compose --profile redis up -d
```

**PostgreSQL (for persistent storage)**:
```bash
docker-compose --profile postgres up -d
```

**All services including optional ones**:
```bash
docker-compose --profile redis --profile postgres up -d
```

## Features

### Multi-Provider Workflow Engine
- **Provider-Agnostic**: Orchestrate workflows across multiple service providers
- **Dynamic Registration**: Register service proxies at runtime
- **Dependency Resolution**: Steps can reference data from previous steps using `${step_id.field_name}`
- **Retry Logic**: Built-in exponential backoff for robust execution
- **Timeout Support**: Configurable timeouts per step

### Unified Proxy Interface
- **Standardized Requests/Responses**: All services use the same `ProxyRequest`/`ProxyResponse` format
- **LLM-Friendly**: Simple function names and payload structure for AI orchestration
- **Consistent Error Handling**: Standardized error codes and messages
- **Batch Operations**: Support for parallel execution of multiple requests

### Google Workspace Integration

#### Currently Available Services

**Gmail Proxy** (`gmail`)
- `send_message` - Send emails with attachments support (max 25MB)
- `get_message` - Retrieve specific messages
- `list_messages` - List messages in mailbox with threading
- `search_messages` - Advanced search with labels support

**Google Docs Proxy** (`docs`)
- `create_document` - Create new documents
- `get_document` - Retrieve document content
- `insert_text` - Insert text into documents
- `update_document` - Update document content
- `batch_update` - Perform multiple updates (max 10MB documents)

**Google Drive Proxy** (`drive`)
- `create_folder` - Create folders with organization
- `upload_file` - Upload files (max 5TB)
- `get_file` - Retrieve files with version control
- `list_files` - List files and folders with search
- `share_file` - Share files with permission management
- `move_file` - Move files between folders

#### Planned Services

**Google Calendar Proxy** (`calendar`) - *In Development*
- `create_event` - Create calendar events
- `get_event` - Retrieve event details
- `list_events` - List calendar events
- `update_event` - Update existing events
- `delete_event` - Delete events

## Usage

### Basic Workflow Engine Usage

```go
import "github.com/dimitar-trifonov/service-proxies/workflow"

// Create workflow engine
engine := workflow.NewMultiProviderWorkflowEngine()

// Register service proxies
engine.RegisterServiceProxy("workspace", "gmail", workspaceGmailProxy)
engine.RegisterServiceProxy("workspace", "docs", workspaceDocsProxy)

// Set provider tokens
engine.SetProviderToken("workspace", oauthToken)

// Define workflow steps
steps := []workflow.WorkflowStep{
    {
        ID: "create_doc",
        Provider: "workspace",
        Service: "docs",
        Function: "create_document",
        Payload: map[string]interface{}{
            "title": "Project Proposal",
        },
    },
    {
        ID: "send_email",
        Provider: "workspace", 
        Service: "gmail",
        Function: "send_message",
        Payload: map[string]interface{}{
            "to": "client@example.com",
            "subject": "Your proposal is ready",
            "body": "Document: ${create_doc.document_url}",
        },
        DependsOn: []string{"create_doc"},
    },
}

// Execute workflow
execution, err := engine.ExecuteWorkflow(context.Background(), steps, inputData)
```

### Cross-Provider Workflows

```go
// Example: Create document in Workspace, send notification via Office365
steps := []workflow.WorkflowStep{
    {
        ID: "create_doc",
        Provider: "workspace",
        Service: "docs", 
        Function: "create_document",
        Payload: map[string]interface{}{"title": "Proposal"},
    },
    {
        ID: "notify_team",
        Provider: "office365",
        Service: "outlook",
        Function: "send_message", 
        Payload: map[string]interface{}{
            "to": "team@company.com",
            "subject": "New proposal created",
            "body": "Document URL: ${create_doc.document_url}",
        },
        DependsOn: []string{"create_doc"},
    },
}
```

## Installation

```bash
go get github.com/dimitar-trifonov/service-proxies
```

## Project Structure

This project uses a **multi-module Go structure** with separate `go.mod` files for:
- Root module: Main application and shared dependencies
- `workflow/`: Workflow engine module
- `providers/workspace/`: Google Workspace provider module

This modular approach allows for independent versioning and dependency management of different components.

## Dependencies

- Google API Go Client Libraries
- Gin Web Framework for HTTP handlers
- OAuth2 for authentication

## Running the Application

```bash
# Run the main application
go run main.go
```

The main application provides HTTP endpoints for interacting with the proxy services.

## Contributing

This project is part of the SOHOaaS (Small Office Home Office as a Service) architecture and follows the Requirements-as-Code (RAC) methodology.

## License

MIT License
