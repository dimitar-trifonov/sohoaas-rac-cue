package types

// ServiceSchema defines the complete schema for an MCP service
type ServiceSchema struct {
	ServiceName string                    `json:"service_name"`
	Actions     map[string]ActionSchema   `json:"actions"`
	Status      string                    `json:"status"` // "connected", "disconnected", "error"
}

// ActionSchema defines the schema for a specific service action
type ActionSchema struct {
	ActionName     string        `json:"action_name"`
	RequiredFields []FieldSchema `json:"required_fields"`
	OptionalFields []FieldSchema `json:"optional_fields"`
	Description    string        `json:"description"`
}

// FieldSchema defines the schema for a service action field
type FieldSchema struct {
	FieldName        string   `json:"field_name"`
	FieldType        string   `json:"field_type"`        // "string", "email", "date", "array", "object"
	PlaceholderType  string   `json:"placeholder_type"`  // "USER_INPUT", "SYSTEM", "TEMPLATE", "RUNTIME"
	ValidationRules  []string `json:"validation_rules"`  // e.g., ["email_format", "required", "max_length:100"]
	Description      string   `json:"description"`
	ExampleValue     string   `json:"example_value"`
}

// ServiceCatalog holds all available service schemas for a user
type ServiceCatalog struct {
	UserID   string                    `json:"user_id"`
	Services map[string]ServiceSchema  `json:"services"`
	LastUpdated string                 `json:"last_updated"`
}

// NOTE: Static service catalogs removed - use live MCP catalog via MCPService.GetServiceCatalog()
// All service schemas should be retrieved dynamically from the MCP service
