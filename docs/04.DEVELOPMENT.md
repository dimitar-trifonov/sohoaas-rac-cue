# SOHOAAS Development Setup

## Local Development Environment

Following the RaC specifications and the collaboration system's "examine_existing_before_creating" principle.

## 🛠️ Prerequisites

### Required Tools
- **Go 1.21+**: Backend development with Genkit
- **Node.js 18+**: Frontend development
- **Docker**: MCP server containers
- **CUE CLI**: RaC specification validation
- **Firebase CLI**: Local Firestore emulation
- **Git**: Version control

### Installation Commands
```bash
# Install Go (if not installed)
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Install Node.js (via nvm)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
nvm install 18
nvm use 18

# Install CUE
go install cuelang.org/go/cmd/cue@latest

# Install Firebase CLI
npm install -g firebase-tools

# Install Docker (Ubuntu/Debian)
sudo apt-get update
sudo apt-get install docker.io docker-compose
sudo usermod -aG docker $USER
```

## 🏗️ Project Setup

### 1. Clone and Initialize
```bash
# Clone repository
git clone <repository-url>
cd sohoaas

# Verify RaC specifications
cue vet rac/system.cue
cue vet rac/agents/*.cue
cue vet rac/services/*.cue

# Check project structure matches RaC
ls -la rac/agents/    # Should show 4 agent files
ls -la rac/services/  # Should show 4 service files
```

### 2. Backend Setup (Golang + Genkit)
```bash
cd app/backend

# Initialize Go module (if not exists)
go mod init sohoaas-backend

# Install dependencies
go mod download

# Verify Genkit integration
go list -m github.com/firebase/genkit/go

# Create environment file
cp .env.example .env
```

### 3. Environment Configuration
```bash
# .env file for local development
OPENAI_API_KEY=your-openai-api-key
FIREBASE_PROJECT_ID=your-firebase-project
GOOGLE_CLIENT_ID=your-google-oauth-client-id
GOOGLE_CLIENT_SECRET=your-google-oauth-client-secret

# MCP Configuration
MCP_BASE_URL=http://localhost:8080

# Genkit Configuration
GENKIT_ENVIRONMENT=dev
GENKIT_REFLECTION_PORT=3101

# Local development
PORT=4001
DEBUG=true
```

### 4. Firebase Setup
```bash
# Login to Firebase
firebase login

# Initialize Firebase in project
firebase init firestore

# Start Firestore emulator
firebase emulators:start --only firestore
```

### 5. MCP Server Setup
```bash
# Start MCP servers for Google Workspace
cd mcp/server
docker-compose up -d

# Verify MCP connectivity
curl http://localhost:8080/health
```

## 🧪 Development Workflow

### RaC-First Development Process
Following the collaboration system's "plan_patch_prove_over_direct_write" principle:

1. **Plan**: Update RaC specifications first
2. **Evidence**: Examine existing code structure
3. **Patch**: Make minimal, targeted changes
4. **Prove**: Validate with tests and schema compliance

### 1. RaC Specification Updates
```bash
# Always start with RaC updates
vim rac/system.cue
vim rac/agents/intent_gatherer.cue
vim rac/services/personal_capabilities.cue

# Validate changes
cue vet rac/system.cue
cue fmt rac/**/*.cue
```

### 2. Backend Development
```bash
cd app/backend

# Run backend with hot reload
go run main.go

# Run with Genkit reflection server
GENKIT_REFLECTION_PORT=3101 go run main.go

# Access Genkit UI at http://localhost:3101
```

### 3. Agent Development
Following the 4-agent architecture:

```bash
# Test Intent Gatherer
go test ./internal/agents/intent_gatherer -v

# Test Intent Analyst  
go test ./internal/agents/intent_analyst -v

# Test Workflow Generator
go test ./internal/agents/workflow_generator -v

# Test Workflow Validator
go test ./internal/agents/workflow_validator -v
```

### 4. Service Development
Following the 4-service architecture:

```bash
# Test Personal Capabilities Service
go test ./internal/services/personal_capabilities -v

# Test CUE Generator Service
go test ./internal/services/cue_generator -v

# Test Workflow Executor Service
go test ./internal/services/workflow_executor -v

# Test Agent Manager Service
go test ./internal/services/agent_manager -v
```

### 5. Integration Testing
```bash
# Full system integration
go test ./internal/integration -v

# Event flow testing
go test ./internal/events -v

# MCP integration testing
go test ./internal/mcp -v
```

## 🔧 Development Tools

### Genkit Development
```bash
# Start Genkit reflection server
genkit start --port 3101

# View flows and traces
open http://localhost:3101

# Debug LLM calls
genkit flow run intentGathererFlow --input '{"user_message": "send email"}'
```

### CUE Development
```bash
# Validate all RaC specifications
cue vet rac/...

# Format CUE files
cue fmt rac/**/*.cue

# Export to JSON for debugging
cue export rac/system.cue --out json

# Validate specific agent
cue vet rac/agents/workflow_generator.cue rac/schemas.cue
```

### Firebase Development
```bash
# Start all emulators
firebase emulators:start

# Firestore UI: http://localhost:4000
# Auth UI: http://localhost:9099

# Clear emulator data
firebase emulators:exec --only firestore "curl -X DELETE http://localhost:8080/emulator/v1/projects/demo-project/databases/(default)/documents"
```

## 📊 Debugging and Monitoring

### Backend Debugging
```bash
# Enable debug logging
DEBUG=true go run main.go

# Profile memory usage
go tool pprof http://localhost:4001/debug/pprof/heap

# Profile CPU usage
go tool pprof http://localhost:4001/debug/pprof/profile
```

### Agent Debugging
Following `rac/observability/agent_logging_schema.cue`:

```bash
# View agent logs
curl http://localhost:4001/api/debug/agents

# Check agent status
curl http://localhost:4001/api/agents/status

# View event flow
curl http://localhost:4001/api/debug/events
```

### Event Flow Debugging
```bash
# Trace event routing
curl -X POST http://localhost:4001/api/debug/trace \
  -H "Content-Type: application/json" \
  -d '{"event": "user_authenticated", "data": {...}}'

# View agent manager state
curl http://localhost:4001/api/debug/agent-manager
```

## 🧪 Testing Strategy

### Unit Testing
```bash
# Test all packages
go test ./...

# Test with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Integration Testing
```bash
# Test full workflow
go test ./internal/integration/workflow_test.go -v

# Test MCP integration
go test ./internal/mcp/integration_test.go -v

# Test event flow
go test ./internal/events/flow_test.go -v
```

### RaC Compliance Testing
```bash
# Validate all specifications
cue vet rac/system.cue rac/agents/*.cue rac/services/*.cue

# Test schema compliance
go test ./internal/validation/rac_test.go -v

# Validate against real MCP responses
go test ./internal/mcp/schema_test.go -v
```

## 🔄 Hot Reload Setup

### Backend Hot Reload
```bash
# Install air for Go hot reload
go install github.com/cosmtrek/air@latest

# Create .air.toml configuration
cat > .air.toml << EOF
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ."
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  kill_delay = "0s"
  log = "build-errors.log"
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = false
EOF

# Start with hot reload
air
```

### Frontend Hot Reload
```bash
cd app/frontend

# Install dependencies
npm install

# Start development server with hot reload
npm run dev

# Access at http://localhost:3000
```

## 📋 Common Development Tasks

### Adding New Agent
1. Create RaC specification: `rac/agents/new_agent.cue`
2. Validate: `cue vet rac/agents/new_agent.cue rac/schemas.cue`
3. Implement: `app/backend/internal/agents/new_agent.go`
4. Add tests: `app/backend/internal/agents/new_agent_test.go`
5. Update agent manager routing

### Adding New Service
1. Create RaC specification: `rac/services/new_service.cue`
2. Validate: `cue vet rac/services/new_service.cue rac/schemas.cue`
3. Implement: `app/backend/internal/services/new_service.go`
4. Add tests: `app/backend/internal/services/new_service_test.go`
5. Update service registry

### Updating Event Schema
1. Update: `rac/schemas.cue`
2. Validate: `cue vet rac/system.cue`
3. Update affected agents/services
4. Run integration tests
5. Update API documentation

## 🚨 Troubleshooting

### Common Issues
1. **CUE validation errors**: Check schema compliance
2. **Genkit connection issues**: Verify API keys and network
3. **MCP server unavailable**: Check Docker containers
4. **Firebase emulator issues**: Clear data and restart
5. **Go module issues**: Run `go mod tidy`

### Debug Commands
```bash
# Check Go environment
go env

# Verify dependencies
go list -m all

# Check CUE installation
cue version

# Test Firebase connection
firebase projects:list

# Verify Docker containers
docker ps
```

This development setup ensures consistency with the RaC source-of-truth approach and enables efficient development of the 4-agent + 4-service PoC architecture.
