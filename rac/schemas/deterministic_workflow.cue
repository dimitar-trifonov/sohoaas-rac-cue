package schemas

// Deterministic Workflow Schema for SOHOAAS PoC
// Defines the structure for executable, step-based workflows

#DeterministicWorkflow: {
	version:      string
	name:         string
	description?: string
	original_intent?: string // User's original request for transparency

	// 1. EXECUTION STEPS with services
	steps: [...#WorkflowStep]

	// 2. USER PARAMETERS (collected before execution)
	user_parameters: [string]: #UserParameter

	// 3. SERVICE BINDINGS (MCP connections)
	service_bindings: [string]: #ServiceBinding

	// 4. EXECUTION CONFIGURATION (PoC: Sequential only)
	execution_config?: #ExecutionConfig

	// Optional execution metadata
	execution_order?: [...string] // Computed dependency order
	validation_schema?: {...} // Additional validation rules
}

#WorkflowStep: {
	id:   string // Valid identifier
	name: string

	// MCP ALIGNMENT: action must match MCP tool name exactly
	// Examples: "gmail.send_message", "docs.create_document", "drive.share_file"
	action: string // MCP tool name (e.g., "gmail.send_message")

	// Parameters must align with MCP tool inputSchema
	// Structure: {parameter_name: value | reference | object}
	parameters: [string]: string | int | bool | #ParameterReference | {...}

	// What this step produces for subsequent steps (with schema)
	outputs?: [string]: #StepOutput

	// Explicit dependencies on other steps
	depends_on?: [...string]

	// Optional step metadata
	description?: string
	timeout?:     string // e.g., "30s", "5m", "1h"

	// MCP service metadata (derived from MCP tool definition)
	_mcp_service_type?: string // e.g., "gmail", "docs", "drive", "calendar"
}

#ParameterReference: {
	// User input: ${user.parameter_name}
	// Step output: ${steps.step_id.outputs.output_name}
	// Computed: ${computed.expression}
	pattern: string
}

#UserParameter: {
	type:         "string" | "number" | "boolean" | "array" | "object" | "datetime" | "email"
	required:     bool
	prompt:       string
	description?: string

	// Validation rules
	validation?: "email" | "url" | "phone" | string // regex pattern
	min_length?: int
	max_length?: int

	// For choice parameters
	options?: [...string]

	// Default values
	default?: string | number | bool

	// UI hints
	placeholder?: string
	help_text?:   string
}

#ServiceBinding: {
	type:      "mcp_service" | "api_service" | "webhook"
	provider?: string // e.g., "workspace"

	// Authentication configuration
	auth: #AuthConfig

	// API configuration
	base_url?:    string
	api_version?: string

	// Service-specific configuration
	config?: {...}

	// Rate limiting
	rate_limit?: #RateLimit
}

#StepOutput: {
	type:        "string" | "number" | "boolean" | "array" | "object"
	description: string
}

#ExecutionConfig: {
	mode:         "sequential" // PoC: Sequential execution only
	timeout?:     string
	environment?: "development" | "staging" | "production"
}

#AuthConfig: {
	method: "oauth2" | "api_key" | "basic"
	oauth2?: {
		scopes: [...string]
		token_source: "user" | "service"
	}
	api_key?: {
		header_name:  string
		token_source: "user" | "service"
	}
}

#RateLimit: {
	requests_per_minute?: int
	requests_per_hour?:   int
	burst_limit?:         int
}

// Validation functions for workflow integrity
#ValidateWorkflow: {
	workflow: #DeterministicWorkflow

	// Basic validation - dependencies exist, services defined, parameters valid
	// Note: Full validation would require more complex CUE expressions
	valid: true
}

// Example executable workflow with MCP-aligned configuration
#ExampleExecutableWorkflow: #DeterministicWorkflow & {
	version:     "1.0.0"
	name:        "Document Creation and Email Notification"
	description: "Create a document from template and send notification email"

	steps: [
		{
			id:     "create_document"
			name:   "Create Document from Template"
			action: "docs.create_document" // MCP tool name
			parameters: {
				title: "${user.document_title}"
			}
			outputs: {
				document_id: {
					type:        "string"
					description: "Google Docs document ID"
				}
				document_url: {
					type:        "string"
					description: "Shareable document URL"
				}
			}
			timeout:           "30s"
			_mcp_service_type: "docs"
		},
		{
			id:     "share_document"
			name:   "Share Document with Collaborator"
			action: "drive.share_file" // MCP tool name
			parameters: {
				file_id: "${steps.create_document.outputs.document_id}"
				email:   "${user.collaborator_email}"
				role:    "writer"
			}
			depends_on: ["create_document"]
			outputs: {
				share_url: {
					type:        "string"
					description: "Shared document URL"
				}
			}
			timeout:           "30s"
			_mcp_service_type: "drive"
		},
		{
			id:     "send_notification"
			name:   "Send Email Notification"
			action: "gmail.send_message" // MCP tool name
			parameters: {
				to:      "${user.collaborator_email}"
				subject: "Document Ready: ${user.document_title}"
				body:    "Hi! The document '${user.document_title}' has been created and shared with you. Access it here: ${steps.share_document.outputs.share_url}"
			}
			depends_on: ["share_document"]
			timeout:           "30s"
			_mcp_service_type: "gmail"
		},
	]

	user_parameters: {
		document_title: {
			type:       "string"
			required:   true
			prompt:     "What should the document be titled?"
			max_length: 200
		}
		collaborator_email: {
			type:       "string"
			required:   true
			prompt:     "Email address to share with"
			validation: "email"
		}
	}

	service_bindings: {
		drive: {
			type:     "mcp_service"
			provider: "workspace"
			auth: {
				method: "oauth2"
				oauth2: {
					scopes: ["https://www.googleapis.com/auth/drive.file"]
					token_source: "user"
				}
			}
			rate_limit: {
				requests_per_minute: 100
			}
		}
		docs: {
			type:     "mcp_service"
			provider: "workspace"
			auth: {
				method: "oauth2"
				oauth2: {
					scopes: ["https://www.googleapis.com/auth/documents"]
					token_source: "user"
				}
			}
		}
	}

	execution_config: {
		mode:        "sequential"
		timeout:     "5m"
		environment: "development"
	}
}
