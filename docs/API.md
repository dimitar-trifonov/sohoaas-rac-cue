# SOHOAAS API Documentation

## Event Schemas and Service Interfaces

Following the RaC source-of-truth specifications from `rac/system.cue`.

## 🔄 Core Event Flow

### Authentication Events
```json
{
  "event": "user_authenticated",
  "data": {
    "user_id": "string",
    "email": "string", 
    "access_token": "string",
    "refresh_token": "string",
    "scopes": ["gmail", "docs", "calendar", "drive"]
  }
}
```

### Capability Discovery Events
```json
{
  "event": "capabilities_discovered", 
  "data": {
    "user_id": "string",
    "mcp_servers": ["gmail", "docs", "calendar", "drive"],
    "available_actions": ["send_email", "create_document", "schedule_meeting"],
    "parameter_schemas": {
      "send_email": {
        "to": {"type": "string", "required": true},
        "subject": {"type": "string", "required": true},
        "body": {"type": "string", "required": true}
      }
    }
  }
}
```

### Workflow Events
```json
{
  "event": "workflow_intent_discovered",
  "data": {
    "session_id": "string",
    "user_id": "string", 
    "discovered_intent": "string",
    "required_parameters": ["param1", "param2"],
    "confidence": "high|medium|low"
  }
}
```

```json
{
  "event": "intent_analysis_complete",
  "data": {
    "user_id": "string",
    "analyzed_intent": "string",
    "extracted_parameters": {...},
    "required_services": ["gmail", "docs"],
    "workflow_pattern": "email_automation|document_creation|meeting_scheduling"
  }
}
```

```json
{
  "event": "workflow_generation_complete", 
  "data": {
    "user_id": "string",
    "workflow_name": "string",
    "json_workflow": {...},
    "user_parameters": {...},
    "service_bindings": {...}
  }
}
```

```json
{
  "event": "cue_conversion_complete",
  "data": {
    "user_id": "string", 
    "workflow_name": "string",
    "cue_workflow": "string",
    "schema_path": "/rac/schemas/deterministic_workflow.cue"
  }
}
```

```json
{
  "event": "validation_complete",
  "data": {
    "user_id": "string",
    "workflow_name": "string", 
    "validation_status": "valid|invalid",
    "validation_errors": [...],
    "validated_workflow": {...}
  }
}
```

```json
{
  "event": "execution_completed",
  "data": {
    "user_id": "string",
    "workflow_name": "string",
    "execution_results": [...],
    "status": "success|partial|failed",
    "summary": "string"
  }
}
```

## 🤖 Agent Service Interfaces

### Intent Gatherer Agent
- **Input**: Natural language user intent
- **Output**: `workflow_intent_discovered` event
- **LLM Integration**: Multi-turn conversation capability
- **State Management**: Session-based discovery tracking

### Intent Analyst Agent  
- **Input**: `workflow_intent_discovered` event
- **Output**: `intent_analysis_complete` event
- **LLM Integration**: Pattern recognition and parameter extraction
- **Capabilities**: 5-parameter simplified analysis for PoC

### Workflow Generator Agent
- **Input**: `intent_analysis_complete` event  
- **Output**: `workflow_generation_complete` event
- **LLM Integration**: JSON workflow generation with MCP tool names
- **Schema Compliance**: Must generate MCP-compatible tool references

### Workflow Validator Agent
- **Input**: `cue_conversion_complete` event
- **Output**: `validation_complete` event  
- **LLM Integration**: CUE syntax and schema validation
- **Validation**: Against `#DeterministicWorkflow` schema

## ⚙️ Deterministic Service Interfaces

### Personal Capabilities Service
- **Input**: `user_authenticated` event
- **Output**: `capabilities_discovered` event
- **Function**: MCP server discovery and parameter schema extraction
- **No LLM**: Pure deterministic service

### CUE Generator Service
- **Input**: `workflow_generation_complete` event
- **Output**: `cue_conversion_complete` event
- **Function**: 5-step JSON→CUE conversion pipeline
- **No LLM**: Deterministic transformation rules

### Workflow Executor Service  
- **Input**: `validation_complete` event
- **Output**: `execution_completed` event
- **Function**: Step-by-step MCP service execution
- **No LLM**: Pure execution engine

### Agent Manager Service
- **Function**: Event routing and orchestration
- **Routing Rules**: Maps events to appropriate agents/services
- **No LLM**: Deterministic event bus

## 🔌 MCP Integration

### Supported MCP Servers
- **Gmail MCP Server**: Email operations
- **Google Docs MCP Server**: Document operations  
- **Google Calendar MCP Server**: Calendar operations
- **Google Drive MCP Server**: File operations

### MCP Tool Format
```json
{
  "tool_name": "gmail_send_email",
  "service_type": "gmail", 
  "input_schema": {
    "type": "object",
    "properties": {
      "to": {"type": "string", "required": true},
      "subject": {"type": "string", "required": true}, 
      "body": {"type": "string", "required": true}
    }
  }
}
```

## 🧪 Testing Interface

All agents and services include comprehensive test cases following RaC specifications:
- Input/output event validation
- Schema compliance testing  
- Error handling verification
- Integration testing with MCP services

## 🔒 Authentication

- **OAuth2 Flow**: Google Workspace authentication
- **Scopes**: gmail, docs, calendar, drive
- **Token Management**: Refresh token handling
- **Security**: Environment variable API keys
