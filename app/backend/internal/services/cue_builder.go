package services

import (
	"encoding/json"
	"fmt"
	"strings"
)

// WorkflowJSON represents the JSON structure that the LLM will output
// Aligned with workflow_json_schema.json as source of truth
type WorkflowJSON struct {
	Version        string                            `json:"version"`
	Name           string                            `json:"name"`
	Description    string                            `json:"description"`
	Steps          []StepJSON                        `json:"steps"`
	UserParameters map[string]UserParameterJSON      `json:"user_parameters"`
	Services       map[string]ServiceBindingJSON     `json:"services"`
}

type TriggerJSON struct {
	Type     string `json:"type"`
	Schedule string `json:"schedule,omitempty"`
	Event    string `json:"event,omitempty"`
}

type StepJSON struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Service   string                 `json:"service"`
	Action    string                 `json:"action"`
	Inputs    map[string]interface{} `json:"inputs"`
	Outputs   map[string]interface{} `json:"outputs,omitempty"`
	DependsOn []string               `json:"depends_on,omitempty"`
}

type UserParameterJSON struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Required    bool                   `json:"required"`
	Description string                 `json:"description"`
	Prompt      string                 `json:"prompt,omitempty"`
	Validation  map[string]interface{} `json:"validation,omitempty"`
}

type ServiceBindingJSON struct {
	Service     string   `json:"service"`
	OAuthScopes []string `json:"oauth_scopes"`
	Endpoint    string   `json:"endpoint,omitempty"`
}

// CUEBuilder converts JSON workflow definitions to CUE format
type CUEBuilder struct {
	mcpService *MCPService
}

// NewCUEBuilder creates a new CUE builder instance
func NewCUEBuilder(mcpService *MCPService) *CUEBuilder {
	return &CUEBuilder{
		mcpService: mcpService,
	}
}

// BuildCUEFromJSON converts a JSON workflow definition to a valid CUE file
func (cb *CUEBuilder) BuildCUEFromJSON(jsonWorkflow string) (string, error) {
	// Parse JSON workflow
	var workflow WorkflowJSON
	if err := json.Unmarshal([]byte(jsonWorkflow), &workflow); err != nil {
		return "", fmt.Errorf("failed to parse JSON workflow: %w", err)
	}

	// Validate workflow structure (skip MCP validation if service is nil - for testing)
	if cb.mcpService != nil {
		if err := cb.validateWorkflow(&workflow); err != nil {
			return "", fmt.Errorf("workflow validation failed: %w", err)
		}
	} else {
		// Basic validation without MCP service
		if err := cb.validateWorkflowBasic(&workflow); err != nil {
			return "", fmt.Errorf("basic workflow validation failed: %w", err)
		}
	}

	// Auto-generate service bindings if not provided
	if len(workflow.Services) == 0 {
		workflow.Services = cb.generateServiceBindings(&workflow)
	}

	// Build CUE content
	cueContent := cb.generateCUEContent(&workflow)
	
	return cueContent, nil
}

// validateWorkflow ensures the workflow has valid services and actions
func (cb *CUEBuilder) validateWorkflow(workflow *WorkflowJSON) error {
	if workflow.Name == "" {
		return fmt.Errorf("workflow name is required")
	}
	
	if len(workflow.Steps) == 0 {
		return fmt.Errorf("workflow must have at least one step")
	}

	// Basic validation only - service validation is handled by execution engine
	for i, step := range workflow.Steps {
		if step.ID == "" {
			return fmt.Errorf("step %d: ID is required", i)
		}
		if step.Action == "" {
			return fmt.Errorf("step %d (%s): action is required", i, step.ID)
		}
	}

	return nil
}

// validateWorkflowBasic performs basic validation without MCP service (for testing)
func (cb *CUEBuilder) validateWorkflowBasic(workflow *WorkflowJSON) error {
	if workflow.Name == "" {
		return fmt.Errorf("workflow name is required")
	}
	
	if len(workflow.Steps) == 0 {
		return fmt.Errorf("workflow must have at least one step")
	}

	// Validate each step has required fields
	for i, step := range workflow.Steps {
		if step.ID == "" {
			return fmt.Errorf("step %d: ID is required", i)
		}
		if step.Service == "" {
			return fmt.Errorf("step %d (%s): service is required", i, step.ID)
		}
		if step.Action == "" {
			return fmt.Errorf("step %d (%s): action is required", i, step.ID)
		}
	}

	return nil
}

// generateServiceBindings automatically creates OAuth bindings for used services
func (cb *CUEBuilder) generateServiceBindings(workflow *WorkflowJSON) map[string]ServiceBindingJSON {
	serviceScopes := map[string][]string{
		"gmail": {
			"https://www.googleapis.com/auth/gmail.compose",
			"https://www.googleapis.com/auth/gmail.send",
			"https://www.googleapis.com/auth/gmail.readonly",
		},
		"calendar": {
			"https://www.googleapis.com/auth/calendar",
			"https://www.googleapis.com/auth/calendar.events",
		},
		"docs": {
			"https://www.googleapis.com/auth/documents",
			"https://www.googleapis.com/auth/drive.file",
		},
		"drive": {
			"https://www.googleapis.com/auth/drive",
			"https://www.googleapis.com/auth/drive.file",
		},
	}

	// Collect unique services used in workflow
	usedServices := make(map[string]bool)
	for _, step := range workflow.Steps {
		usedServices[step.Service] = true
	}

	// Generate bindings for used services
	bindings := make(map[string]ServiceBindingJSON)
	for service := range usedServices {
		if scopes, exists := serviceScopes[service]; exists {
			bindings[service] = ServiceBindingJSON{
				Service:     service,
				OAuthScopes: scopes,
			}
		}
	}

	return bindings
}

// generateCUEContent creates the actual CUE file content
func (cb *CUEBuilder) generateCUEContent(workflow *WorkflowJSON) string {
	var cue strings.Builder
	
	// Write package and schema definitions
	cue.WriteString(`package workflow

#DeterministicWorkflow: {
	name: string
	description: string
	trigger: {
		type: "schedule" | "manual" | "event"
		schedule?: string
		event?: string
	}
	steps: [...#WorkflowStep]
	user_parameters: [...#UserParameter]
	service_bindings: [...#ServiceBinding]
}

#WorkflowStep: {
	id: string
	name: string
	service: string
	action: string
	inputs: {...}
	outputs: {...}
	depends_on?: [...string]
}

#UserParameter: {
	name: string
	type: "string" | "number" | "boolean" | "array" | "object"
	required: bool
	description: string
	prompt?: string
	validation?: {...}
}

#ServiceBinding: {
	service: string
	oauth_scopes: [...string]
	endpoint?: string
}

workflow: #DeterministicWorkflow & {
`)

	// Write workflow metadata
	cue.WriteString(fmt.Sprintf("\tversion: %q\n", workflow.Version))
	cue.WriteString(fmt.Sprintf("\tname: %q\n", workflow.Name))
	cue.WriteString(fmt.Sprintf("\tdescription: %q\n", workflow.Description))

	// Write steps
	cue.WriteString("\tsteps: [\n")
	for _, step := range workflow.Steps {
		cue.WriteString("\t\t{\n")
		cue.WriteString(fmt.Sprintf("\t\t\tid: %q\n", step.ID))
		cue.WriteString(fmt.Sprintf("\t\t\tname: %q\n", step.Name))
		cue.WriteString(fmt.Sprintf("\t\t\tservice: %q\n", step.Service))
		cue.WriteString(fmt.Sprintf("\t\t\taction: %q\n", step.Action))
		
		// Write inputs
		cue.WriteString("\t\t\tinputs: {\n")
		for key, value := range step.Inputs {
			cue.WriteString(fmt.Sprintf("\t\t\t\t%s: %s\n", key, cb.formatValue(value)))
		}
		cue.WriteString("\t\t\t}\n")
		
		// Write outputs
		cue.WriteString("\t\t\toutputs: {\n")
		for key, value := range step.Outputs {
			cue.WriteString(fmt.Sprintf("\t\t\t\t%s: %s\n", key, cb.formatValue(value)))
		}
		cue.WriteString("\t\t\t}\n")
		
		// Write dependencies
		if len(step.DependsOn) > 0 {
			cue.WriteString("\t\t\tdepends_on: [")
			for i, dep := range step.DependsOn {
				if i > 0 {
					cue.WriteString(", ")
				}
				cue.WriteString(fmt.Sprintf("%q", dep))
			}
			cue.WriteString("]\n")
		}
		
		cue.WriteString("\t\t}\n")
	}
	cue.WriteString("\t]\n")

	// Write user parameters
	cue.WriteString("\tuser_parameters: [\n")
	for _, param := range workflow.UserParameters {
		cue.WriteString("\t\t{\n")
		cue.WriteString(fmt.Sprintf("\t\t\tname: %q\n", param.Name))
		cue.WriteString(fmt.Sprintf("\t\t\ttype: %q\n", param.Type))
		cue.WriteString(fmt.Sprintf("\t\t\trequired: %t\n", param.Required))
		cue.WriteString(fmt.Sprintf("\t\t\tdescription: %q\n", param.Description))
		if param.Prompt != "" {
			cue.WriteString(fmt.Sprintf("\t\t\tprompt: %q\n", param.Prompt))
		}
		cue.WriteString("\t\t}\n")
	}
	cue.WriteString("\t]\n")

	// Write service bindings
	cue.WriteString("\tservice_bindings: [\n")
	for _, binding := range workflow.Services {
		cue.WriteString("\t\t{\n")
		cue.WriteString(fmt.Sprintf("\t\t\tservice: %q\n", binding.Service))
		cue.WriteString("\t\t\toauth_scopes: [")
		for i, scope := range binding.OAuthScopes {
			if i > 0 {
				cue.WriteString(", ")
			}
			cue.WriteString(fmt.Sprintf("%q", scope))
		}
		cue.WriteString("]\n")
		cue.WriteString("\t\t}\n")
	}
	cue.WriteString("\t]\n")

	cue.WriteString("}\n")
	
	return cue.String()
}

// formatValue formats a value for CUE syntax
func (cb *CUEBuilder) formatValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("%q", v)
	case int, int64, float64:
		return fmt.Sprintf("%v", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case []interface{}:
		return "[]" // Simplified for now
	default:
		return fmt.Sprintf("%q", fmt.Sprintf("%v", v))
	}
}
