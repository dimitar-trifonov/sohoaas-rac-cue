package types

import "time"

// Agent represents the state and capabilities of an agent
type Agent struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	State       string            `json:"state"`
	Capabilities []string         `json:"capabilities"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Event represents an event in the system
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Target    string                 `json:"target,omitempty"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// User represents an authenticated user
type User struct {
	ID           string                 `json:"id"`
	Email        string                 `json:"email"`
	Name         string                 `json:"name"`
	OAuthTokens  map[string]interface{} `json:"oauth_tokens,omitempty"`
	ConnectedServices []string          `json:"connected_services"`
}

// WorkflowIntent represents a structured workflow intent
type WorkflowIntent struct {
	UserMessage       string                 `json:"user_message"`       // User's natural language request
	WorkflowPattern   string                 `json:"workflow_pattern"`
	TriggerConditions map[string]interface{} `json:"trigger_conditions"`
	ActionSequence    []WorkflowAction       `json:"action_sequence"`
	DataRequirements  []DataRequirement      `json:"data_requirements"`
	UserParameters    []UserParameter        `json:"user_parameters"`
}

// WorkflowAction represents an action in a workflow
type WorkflowAction struct {
	Service     string                 `json:"service"`
	Action      string                 `json:"action"`
	Parameters  map[string]interface{} `json:"parameters"`
	Dependencies []string              `json:"dependencies,omitempty"`
}

// DataRequirement represents data needed for workflow execution
type DataRequirement struct {
	Source      string `json:"source"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

// UserParameter represents a parameter that needs user input
type UserParameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Description string      `json:"description"`
	Prompt      string      `json:"prompt"`
	Default     interface{} `json:"default"`
	Validation  string      `json:"validation,omitempty"`
}

// MCPService represents an MCP service capability
type MCPService struct {
	Service   string              `json:"service"`
	Functions []MCPFunction       `json:"functions"`
	Status    string              `json:"status"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// MCPFunction represents an MCP function with full schema
type MCPFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Required    []string               `json:"required"`
	Returns     map[string]interface{} `json:"returns,omitempty"`
}

// ConversationMessage represents a message in a conversation
type ConversationMessage struct {
	Role      string    `json:"role"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// AgentRequest represents a request to an agent
type AgentRequest struct {
	AgentID string                 `json:"agent_id"`
	Input   map[string]interface{} `json:"input"`
	Context map[string]interface{} `json:"context,omitempty"`
}

// AgentResponse represents a response from an agent
type AgentResponse struct {
	AgentID   string                 `json:"agent_id"`
	Output    map[string]interface{} `json:"output,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// WorkflowFile represents a generated CUE workflow file stored on disk
type WorkflowFile struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"` // 'draft' | 'active' | 'completed' | 'error'
	Filename    string                 `json:"filename"`
	Path        string                 `json:"path"`
	UserID      string                 `json:"user_id"`
	Content     string                 `json:"content"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ParsedData  map[string]interface{} `json:"parsed_data,omitempty"` // Parsed CUE workflow structure
}

// WorkflowExecution represents the execution state of a workflow
type WorkflowExecution struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id"`
	WorkflowCUE string                 `json:"workflow_cue"`
	Status      string                 `json:"status"`
	Steps       []WorkflowStep         `json:"steps"`
	Results     map[string]interface{} `json:"results,omitempty"`
	Error       string                 `json:"error,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// WorkflowStep represents a step in workflow execution
type WorkflowStep struct {
	ID          string                 `json:"id"`
	Service     string                 `json:"service"`
	Action      string                 `json:"action"`
	Parameters  map[string]interface{} `json:"parameters"`
	Status      string                 `json:"status"`
	Output      map[string]interface{} `json:"output,omitempty"`
	Error       string                 `json:"error,omitempty"`
	ExecutedAt  *time.Time             `json:"executed_at,omitempty"`
}
