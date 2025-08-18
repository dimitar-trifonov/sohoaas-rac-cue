package types

// MCPServiceCatalog represents the complete MCP service catalog structure
// Matches actual MCP server response: providers.workspace[serviceName]
type MCPServiceCatalog struct {
	Providers MCPProviders `json:"providers"`
}

// MCPProviders represents the providers section of MCP catalog
type MCPProviders struct {
	Workspace MCPWorkspaceProvider `json:"workspace"`
}

// MCPWorkspaceProvider represents the workspace provider structure
type MCPWorkspaceProvider struct {
	Description string                              `json:"description"`
	DisplayName string                              `json:"display_name"`
	Services    map[string]MCPServiceDefinition     `json:"services"`
}

// MCPServiceDefinition represents a complete service definition in MCP catalog
// Matches actual MCP response structure with description, display_name, functions
type MCPServiceDefinition struct {
	Description  string                        `json:"description"`
	DisplayName  string                        `json:"display_name"`
	Functions    map[string]MCPFunctionSchema  `json:"functions"`
}

// MCPFunctionSchema represents the schema definition for an MCP function
// Matches actual MCP response: name, display_name, description, example_payload, required_fields
// Extended with response schemas for complete service contract
type MCPFunctionSchema struct {
	Name           string                 `json:"name"`
	DisplayName    string                 `json:"display_name"`
	Description    string                 `json:"description"`
	ExamplePayload map[string]interface{} `json:"example_payload"`
	RequiredFields []string               `json:"required_fields"`
	// Response schema information for workflow generation
	OutputSchema   *MCPResponseSchema     `json:"output_schema,omitempty"`
	ErrorSchema    *MCPResponseSchema     `json:"error_schema,omitempty"`
}

// MCPParameterSchema represents parameter schema for MCP functions
type MCPParameterSchema struct {
	Type       string                           `json:"type"`
	Properties map[string]MCPParameterProperty  `json:"properties,omitempty"`
	Required   []string                         `json:"required,omitempty"`
}

// MCPParameterProperty represents individual parameter properties
type MCPParameterProperty struct {
	Type        string      `json:"type"`
	Description string      `json:"description,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
	Format      string      `json:"format,omitempty"`
}

// MCPResponseSchema represents response schema for MCP function outputs and errors
type MCPResponseSchema struct {
	Type        string                           `json:"type"`
	Description string                           `json:"description,omitempty"`
	Properties  map[string]MCPParameterProperty  `json:"properties,omitempty"`
	Required    []string                         `json:"required,omitempty"`
}

// MCPOAuthConfig represents OAuth2 configuration for MCP services
type MCPOAuthConfig struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	Scopes       []string `json:"scopes"`
	RedirectURI  string   `json:"redirect_uri"`
	AuthURL      string   `json:"auth_url,omitempty"`
	TokenURL     string   `json:"token_url,omitempty"`
}

// MCPServiceMetadata represents additional service metadata
type MCPServiceMetadata struct {
	Version     string            `json:"version,omitempty"`
	Enabled     bool              `json:"enabled"`
	Tags        []string          `json:"tags,omitempty"`
	Endpoints   map[string]string `json:"endpoints,omitempty"`
	RateLimit   *MCPRateLimit     `json:"rate_limit,omitempty"`
}

// MCPRateLimit represents rate limiting configuration
type MCPRateLimit struct {
	RequestsPerMinute int `json:"requests_per_minute"`
	BurstLimit        int `json:"burst_limit"`
}

// WorkflowStepValidation represents a workflow step for validation purposes
type WorkflowStepValidation struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Action      string                 `json:"action"`
	Parameters  map[string]interface{} `json:"parameters"`
	DependsOn   []string               `json:"depends_on,omitempty"`
	Description string                 `json:"description,omitempty"`
}

// ServiceBindingValidation represents service binding configuration for validation
type ServiceBindingValidation struct {
	ServiceName string          `json:"service_name"`
	AuthType    string          `json:"auth_type"`
	OAuthConfig *MCPOAuthConfig `json:"oauth_config,omitempty"`
	Endpoint    string          `json:"endpoint,omitempty"`
	Enabled     bool            `json:"enabled"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// MCPValidationResult represents the result of MCP validation
type MCPValidationResult struct {
	IsValid            bool                        `json:"is_valid"`
	FunctionValidation *FunctionValidationResult   `json:"function_validation,omitempty"`
	ParameterValidation *ParameterValidationResult `json:"parameter_validation,omitempty"`
	BindingValidation  *BindingValidationResult    `json:"binding_validation,omitempty"`
	ValidationErrors   []MCPValidationError        `json:"validation_errors"`
	ValidationWarnings []MCPValidationWarning      `json:"validation_warnings"`
}

// FunctionValidationResult represents function validation results
type FunctionValidationResult struct {
	ValidatedSteps   []StepFunctionValidation `json:"validated_steps"`
	TotalSteps       int                      `json:"total_steps"`
	ValidSteps       int                      `json:"valid_steps"`
	InvalidSteps     int                      `json:"invalid_steps"`
}

// StepFunctionValidation represents validation result for a single step's function
type StepFunctionValidation struct {
	StepID           string `json:"step_id"`
	Action           string `json:"action"`
	ServiceName      string `json:"service_name"`
	FunctionName     string `json:"function_name"`
	FunctionExists   bool   `json:"function_exists"`
	ServiceAvailable bool   `json:"service_available"`
}

// ParameterValidationResult represents parameter validation results
type ParameterValidationResult struct {
	ValidatedSteps           []StepParameterValidation `json:"validated_steps"`
	TotalParametersChecked   int                       `json:"total_parameters_checked"`
	ValidParameters          int                       `json:"valid_parameters"`
	InvalidParameters        int                       `json:"invalid_parameters"`
}

// StepParameterValidation represents validation result for a single step's parameters
type StepParameterValidation struct {
	StepID                   string   `json:"step_id"`
	Action                   string   `json:"action"`
	RequiredParamsPresent    bool     `json:"required_params_present"`
	ParamTypesValid          bool     `json:"param_types_valid"`
	ParameterReferencesValid bool     `json:"parameter_references_valid"`
	MissingRequiredParams    []string `json:"missing_required_params,omitempty"`
	InvalidParamTypes        []string `json:"invalid_param_types,omitempty"`
	InvalidParameterRefs     []string `json:"invalid_parameter_refs,omitempty"`
}

// BindingValidationResult represents service binding validation results
type BindingValidationResult struct {
	ValidatedBindings    []ServiceBindingValidationResult `json:"validated_bindings"`
	TotalBindings        int                              `json:"total_bindings"`
	ValidBindings        int                              `json:"valid_bindings"`
	InvalidBindings      int                              `json:"invalid_bindings"`
	MissingBindings      []string                         `json:"missing_bindings,omitempty"`
}

// ServiceBindingValidationResult represents validation result for a single service binding
type ServiceBindingValidationResult struct {
	ServiceName           string   `json:"service_name"`
	EndpointReachable     bool     `json:"endpoint_reachable"`
	AuthConfigValid       bool     `json:"auth_config_valid"`
	ScopesValid           bool     `json:"scopes_valid"`
	RequiredScopesPresent bool     `json:"required_scopes_present"`
	MissingScopes         []string `json:"missing_scopes,omitempty"`
}

// MCPValidationError represents a validation error with context
type MCPValidationError struct {
	Type        string `json:"type"`
	StepID      string `json:"step_id,omitempty"`
	ServiceName string `json:"service_name,omitempty"`
	Action      string `json:"action,omitempty"`
	Parameter   string `json:"parameter,omitempty"`
	Message     string `json:"message"`
	Details     string `json:"details,omitempty"`
	Severity    string `json:"severity"` // "error", "warning", "info"
}

// MCPValidationWarning represents a validation warning
type MCPValidationWarning struct {
	Type        string `json:"type"`
	StepID      string `json:"step_id,omitempty"`
	ServiceName string `json:"service_name,omitempty"`
	Message     string `json:"message"`
	Suggestion  string `json:"suggestion,omitempty"`
}

// ParameterReference represents a parsed parameter reference
type ParameterReference struct {
	OriginalValue string                 `json:"original_value"`
	Type          ParameterReferenceType `json:"type"`
	Source        string                 `json:"source"`        // "user", "steps", "computed", "env"
	Path          []string               `json:"path"`          // e.g., ["user", "param"] or ["steps", "step_id", "outputs", "field"]
	IsValid       bool                   `json:"is_valid"`
}

// ParameterReferenceType represents the type of parameter reference
type ParameterReferenceType string

const (
	ParamRefUser     ParameterReferenceType = "user"
	ParamRefStep     ParameterReferenceType = "step"
	ParamRefComputed ParameterReferenceType = "computed"
	ParamRefEnv      ParameterReferenceType = "env"
	ParamRefInvalid  ParameterReferenceType = "invalid"
)

// GoogleWorkspaceScopes defines required OAuth scopes for Google Workspace services
var GoogleWorkspaceScopes = map[string][]string{
	"gmail": {
		"https://www.googleapis.com/auth/gmail.send",
		"https://www.googleapis.com/auth/gmail.readonly",
	},
	"docs": {
		"https://www.googleapis.com/auth/documents",
	},
	"drive": {
		"https://www.googleapis.com/auth/drive.file",
	},
	"calendar": {
		"https://www.googleapis.com/auth/calendar",
	},
	"sheets": {
		"https://www.googleapis.com/auth/spreadsheets",
	},
}
