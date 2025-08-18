package workspace

import (
	"context"

	"github.com/dimitar-trifonov/sohoaas/service-proxies/workflow"
)

// ProxyRequest represents a unified request to any workspace service
type ProxyRequest struct {
	Function    string                 `json:"function" binding:"required"`     // Function name to call
	Token       string                 `json:"token" binding:"required"`        // OAuth2 token
	Payload     map[string]interface{} `json:"payload" binding:"required"`      // Function parameters
	RequestID   string                 `json:"request_id,omitempty"`            // Optional request tracking ID
	ServiceType string                 `json:"service_type" binding:"required"` // gmail, docs, drive, calendar
}

// Note: ProxyResponse, ProxyError, and ResponseMetadata types have been moved to the workflow package
// for unified cross-provider compatibility. All workspace proxies now use workflow.ProxyResponse.

// ResponseSchema represents the schema for function outputs and errors
type ResponseSchema struct {
	Type        string                           `json:"type"`
	Description string                           `json:"description,omitempty"`
	Properties  map[string]PropertySchema        `json:"properties,omitempty"`
	Required    []string                         `json:"required,omitempty"`
}

// PropertySchema represents individual property schema
type PropertySchema struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// FunctionMetadata contains metadata about a service function
type FunctionMetadata struct {
	Name           string                 `json:"name"`
	DisplayName    string                 `json:"display_name"`
	Description    string                 `json:"description"`
	ExamplePayload map[string]interface{} `json:"example_payload"`
	RequiredFields []string               `json:"required_fields"`
	// Response schema information for workflow generation
	OutputSchema   *ResponseSchema        `json:"output_schema,omitempty"`
	ErrorSchema    *ResponseSchema        `json:"error_schema,omitempty"`
}

// ServiceMetadata contains metadata about a service
type ServiceMetadata struct {
	ServiceType string                       `json:"service_type"`
	DisplayName string                       `json:"display_name"`
	Description string                       `json:"description"`
	Functions   map[string]FunctionMetadata `json:"functions"`
}

// WorkspaceProxy defines the interface that all workspace service proxies must implement
type WorkspaceProxy interface {
	// Execute calls a function on the workspace service with the given payload
	Execute(ctx context.Context, function string, token string, payload map[string]interface{}) (*workflow.ProxyResponse, error)

	// GetSupportedFunctions returns a list of supported function names
	GetSupportedFunctions() []string

	// GetServiceType returns the service type (gmail, docs, drive, calendar)
	GetServiceType() string

	// ValidatePayload validates the payload for a given function
	ValidatePayload(function string, payload map[string]interface{}) error

	// GetServiceMetadata returns metadata about the service and its functions
	GetServiceMetadata() ServiceMetadata

	// GetFunctionMetadata returns metadata for a specific function
	GetFunctionMetadata(function string) (FunctionMetadata, error)
}

// Common error codes
const (
	ErrorCodeInvalidFunction      = "INVALID_FUNCTION"
	ErrorCodeInvalidPayload       = "INVALID_PAYLOAD"
	ErrorCodeAuthenticationFailed = "AUTHENTICATION_FAILED"
	ErrorCodeServiceUnavailable   = "SERVICE_UNAVAILABLE"
	ErrorCodeRateLimited          = "RATE_LIMITED"
	ErrorCodeInternalError        = "INTERNAL_ERROR"
	ErrorCodeNotFound             = "NOT_FOUND"
	ErrorCodePermissionDenied     = "PERMISSION_DENIED"
)

// Service types
const (
	ServiceTypeGmail    = "gmail"
	ServiceTypeDocs     = "docs"
	ServiceTypeDrive    = "drive"
	ServiceTypeCalendar = "calendar"
)

// Gmail function names
const (
	GmailFunctionSendMessage    = "send_message"
	GmailFunctionGetMessage     = "get_message"
	GmailFunctionListMessages   = "list_messages"
	GmailFunctionSearchMessages = "search_messages"
)

// Docs function names
const (
	DocsFunctionCreateDocument = "create_document"
	DocsFunctionGetDocument    = "get_document"
	DocsFunctionInsertText     = "insert_text"
	DocsFunctionUpdateDocument = "update_document"
	DocsFunctionBatchUpdate    = "batch_update"
)

// Drive function names
const (
	DriveFunctionCreateFolder = "create_folder"
	DriveFunctionUploadFile   = "upload_file"
	DriveFunctionGetFile      = "get_file"
	DriveFunctionListFiles    = "list_files"
	DriveFunctionShareFile    = "share_file"
	DriveFunctionMoveFile     = "move_file"
)

// Calendar function names
const (
	CalendarFunctionCreateEvent = "create_event"
	CalendarFunctionGetEvent    = "get_event"
	CalendarFunctionListEvents  = "list_events"
	CalendarFunctionUpdateEvent = "update_event"
	CalendarFunctionDeleteEvent = "delete_event"
)

// Common payload field names
const (
	PayloadFieldTo          = "to"
	PayloadFieldSubject     = "subject"
	PayloadFieldBody        = "body"
	PayloadFieldTitle       = "title"
	PayloadFieldContent     = "content"
	PayloadFieldDocumentID  = "document_id"
	PayloadFieldFileID      = "file_id"
	PayloadFieldFolderID    = "folder_id"
	PayloadFieldParentID    = "parent_id"
	PayloadFieldName        = "name"
	PayloadFieldEmail       = "email"
	PayloadFieldRole        = "role"
	PayloadFieldStartTime   = "start_time"
	PayloadFieldEndTime     = "end_time"
	PayloadFieldAttendees   = "attendees"
	PayloadFieldDescription = "description"
)
