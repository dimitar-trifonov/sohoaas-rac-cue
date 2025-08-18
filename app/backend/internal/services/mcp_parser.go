package services

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"sohoaas-backend/internal/types"
)

// MCPCatalogParser provides centralized MCP service catalog parsing utilities
type MCPCatalogParser struct{}

// NewMCPCatalogParser creates a new MCP catalog parser
func NewMCPCatalogParser() *MCPCatalogParser {
	return &MCPCatalogParser{}
}

// ParseServicesFromCatalog extracts services from MCP catalog structure
// Now works directly with strongly-typed catalog - no adapters needed
func (p *MCPCatalogParser) ParseServicesFromCatalog(mcpCatalog interface{}) (map[string]interface{}, error) {
	// Handle strongly-typed catalog (preferred)
	if catalog, ok := mcpCatalog.(*types.MCPServiceCatalog); ok {
		// Direct access - no conversion needed since types match MCP response
		servicesData := make(map[string]interface{})
		for serviceName, serviceDefinition := range catalog.Providers.Workspace.Services {
			servicesData[serviceName] = serviceDefinition
		}
		return servicesData, nil
	}
	
	// Handle legacy map format for backward compatibility
	if catalogMap, ok := mcpCatalog.(map[string]interface{}); ok {
		return p.parseServicesFromMap(catalogMap)
	}
	
	return nil, fmt.Errorf("unsupported catalog type: %T", mcpCatalog)
}

// parseServicesFromMap handles legacy map format (for backward compatibility)
func (p *MCPCatalogParser) parseServicesFromMap(mcpCatalog map[string]interface{}) (map[string]interface{}, error) {
	// Parse MCP services structure: providers → workspace → services
	providersWrapper, ok := mcpCatalog["providers"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid MCP catalog: missing 'providers' key")
	}
	
	workspaceWrapper, ok := providersWrapper["workspace"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid MCP catalog: missing 'workspace' key under 'providers'")
	}
	
	servicesData, ok := workspaceWrapper["services"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid MCP catalog: missing 'services' key under 'workspace'")
	}
	
	return servicesData, nil
}

// BuildUserCapabilities creates user capabilities from MCP services and connected services
// Supports both legacy map[string]interface{} and new *types.MCPServiceCatalog
func (p *MCPCatalogParser) BuildUserCapabilities(mcpCatalog interface{}, connectedServices []string) ([]map[string]interface{}, error) {
	// Handle strongly-typed catalog directly for better performance
	if catalog, ok := mcpCatalog.(*types.MCPServiceCatalog); ok {
		var capabilities []map[string]interface{}
		
		for _, serviceName := range connectedServices {
			if serviceDefinition, exists := catalog.Providers.Workspace.Services[serviceName]; exists {
				var actions []string
				for functionName := range serviceDefinition.Functions {
					// Build full action name as service.function for consistency
					actions = append(actions, serviceName+"."+functionName)
				}
				
				capabilities = append(capabilities, map[string]interface{}{
					"service": serviceName,
					"actions": actions,
					"status":  "connected",
				})
			}
		}
		
		return capabilities, nil
	}
	
	// Handle legacy map format for backward compatibility
	servicesData, err := p.ParseServicesFromCatalog(mcpCatalog)
	if err != nil {
		return nil, err
	}
	
	var capabilities []map[string]interface{}
	
	for _, serviceName := range connectedServices {
		if service, exists := servicesData[serviceName]; exists {
			if serviceMap, ok := service.(map[string]interface{}); ok {
				var actions []string
				if functions, ok := serviceMap["functions"].(map[string]interface{}); ok {
					for actionName := range functions {
						actions = append(actions, actionName)
					}
				}
				
				capabilities = append(capabilities, map[string]interface{}{
					"service": serviceName,
					"actions": actions,
					"status":  "connected",
				})
			}
		}
	}
	
	return capabilities, nil
}

// BuildServiceCapabilities creates service catalog for Personal Capabilities Agent
// Supports both legacy map[string]interface{} and new *types.MCPServiceCatalog
func (p *MCPCatalogParser) BuildServiceCapabilities(mcpCatalog interface{}, hasGoogleAuth bool) (map[string]interface{}, error) {
	serviceCatalog := make(map[string]interface{})
	
	if !hasGoogleAuth {
		return serviceCatalog, nil
	}
	
	servicesData, err := p.ParseServicesFromCatalog(mcpCatalog)
	if err != nil {
		return serviceCatalog, err
	}
	
	for serviceName, serviceData := range servicesData {
		if serviceMap, ok := serviceData.(map[string]interface{}); ok {
			// Extract functions/actions from the service
			var actions []string
			if functions, ok := serviceMap["functions"].(map[string]interface{}); ok {
				for actionName := range functions {
					actions = append(actions, actionName)
				}
			}
			
			serviceCatalog[serviceName] = map[string]interface{}{
				"name":         serviceName,
				"display_name": serviceMap["display_name"],
				"description":  serviceMap["description"],
				"functions":    serviceMap["functions"],
				"actions":      actions,
				"status":       "connected",
				"auth_type":    "oauth2",
			}
		}
	}
	
	return serviceCatalog, nil
}



// ExtractServiceNames extracts all service names from MCP catalog (eliminates duplicate logic)
func (p *MCPCatalogParser) ExtractServiceNames(mcpCatalog map[string]interface{}) ([]string, error) {
	servicesData, err := p.ParseServicesFromCatalog(mcpCatalog)
	if err != nil {
		return nil, err
	}
	
	var serviceNames []string
	for serviceName := range servicesData {
		serviceNames = append(serviceNames, serviceName)
	}
	
	return serviceNames, nil
}

// FormatServicesForPrompt formats service list for LLM prompts (eliminates duplicate logic)
func (p *MCPCatalogParser) FormatServicesForPrompt(mcpCatalog map[string]interface{}) (string, error) {
	serviceNames, err := p.ExtractServiceNames(mcpCatalog)
	if err != nil {
		return "", err
	}
	
	if len(serviceNames) == 0 {
		return "No services available", nil
	}
	
	// Create formatted string for prompts
	result := ""
	for i, serviceName := range serviceNames {
		if i > 0 {
			result += ", "
		}
		result += serviceName
	}
	
	return result, nil
}

// BuildAvailableServicesSection creates the available services section for LLM prompts
// Supports both legacy map[string]interface{} and new *types.MCPServiceCatalog
func (p *MCPCatalogParser) BuildAvailableServicesSection(mcpCatalog interface{}) (string, error) {
	servicesData, err := p.ParseServicesFromCatalog(mcpCatalog)
	if err != nil {
		return "", err
	}
	
	section := "AVAILABLE MCP SERVICES (ONLY use these real services):\n"
	
	for serviceName, serviceData := range servicesData {
		if serviceMap, ok := serviceData.(map[string]interface{}); ok {
			var functions []string
			if functionsData, exists := serviceMap["functions"].(map[string]interface{}); exists {
				for functionName := range functionsData {
					functions = append(functions, functionName)
				}
			}
			section += fmt.Sprintf("- %s: %v\n", serviceName, functions)
		}
	}
	
	return section, nil
}

// ParseMCPCatalog converts map[string]interface{} to strongly-typed MCPServiceCatalog
func (p *MCPCatalogParser) ParseMCPCatalog(mcpCatalog map[string]interface{}) (*types.MCPServiceCatalog, error) {
	// Convert to JSON and back to get proper typing
	jsonData, err := json.Marshal(mcpCatalog)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal MCP catalog: %w", err)
	}

	var catalog types.MCPServiceCatalog
	if err := json.Unmarshal(jsonData, &catalog); err != nil {
		return nil, fmt.Errorf("failed to unmarshal MCP catalog: %w", err)
	}

	return &catalog, nil
}

// ParseWorkflowSteps converts map[string]interface{} steps to strongly-typed WorkflowStepValidation
func (p *MCPCatalogParser) ParseWorkflowSteps(steps []map[string]interface{}) ([]types.WorkflowStepValidation, error) {
	jsonData, err := json.Marshal(steps)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal workflow steps: %w", err)
	}

	var typedSteps []types.WorkflowStepValidation
	if err := json.Unmarshal(jsonData, &typedSteps); err != nil {
		return nil, fmt.Errorf("failed to unmarshal workflow steps: %w", err)
	}

	return typedSteps, nil
}

// ParseServiceBindings converts map[string]interface{} bindings to strongly-typed ServiceBindingValidation
func (p *MCPCatalogParser) ParseServiceBindings(bindings map[string]interface{}) (map[string]types.ServiceBindingValidation, error) {
	jsonData, err := json.Marshal(bindings)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal service bindings: %w", err)
	}

	var typedBindings map[string]types.ServiceBindingValidation
	if err := json.Unmarshal(jsonData, &typedBindings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal service bindings: %w", err)
	}

	return typedBindings, nil
}


// ValidateMCPFunctionsTyped validates workflow steps using strongly-typed structures
func (p *MCPCatalogParser) ValidateMCPFunctionsTyped(catalog *types.MCPServiceCatalog, steps []types.WorkflowStepValidation) (bool, []string) {
	var errors []string
	services := catalog.Providers.Workspace.Services
	
	for i, step := range steps {
		if step.Action == "" {
			errors = append(errors, fmt.Sprintf("Step %d (%s): missing action field", i, step.ID))
			continue
		}

		// Parse service.function format (e.g., "gmail.send_message")
		parts := strings.Split(step.Action, ".")
		if len(parts) < 2 {
			errors = append(errors, fmt.Sprintf("Step %d (%s): invalid action format '%s' - should be 'service.function'", i, step.ID, step.Action))
			continue
		}

		serviceName := parts[0]
		functionName := parts[1] // Just the function name part

		// Check if service exists
		serviceDefinition, exists := services[serviceName]
		if !exists {
			errors = append(errors, fmt.Sprintf("Step %d (%s): service '%s' not found in MCP catalog", i, step.ID, serviceName))
			continue
		}

		// Check if function exists in service
		if _, functionExists := serviceDefinition.Functions[functionName]; !functionExists {
			errors = append(errors, fmt.Sprintf("Step %d (%s): function '%s' not found in service '%s'", i, step.ID, functionName, serviceName))
		}
	}

	return len(errors) == 0, errors
}


// ValidateMCPParametersTyped validates step parameters using strongly-typed structures
func (p *MCPCatalogParser) ValidateMCPParametersTyped(catalog *types.MCPServiceCatalog, steps []types.WorkflowStepValidation) (bool, []string) {
	var errors []string
	services := catalog.Providers.Workspace.Services

	for i, step := range steps {
		if step.Action == "" {
			continue // Skip validation if no action (handled by function validation)
		}

		// Parse service name
		parts := strings.Split(step.Action, ".")
		if len(parts) < 2 {
			continue // Skip if invalid format (handled by function validation)
		}

		serviceName := parts[0]
		functionName := parts[1] // Just the function name part
		
		// Get service definition
		serviceDefinition, exists := services[serviceName]
		if !exists {
			continue // Skip if service doesn't exist (handled by function validation)
		}

		// Get function schema
		functionSchema, exists := serviceDefinition.Functions[functionName]
		if !exists {
			continue // Skip if function doesn't exist (handled by function validation)
		}

		// Validate required parameters
		for _, requiredParam := range functionSchema.RequiredFields {
			if _, hasParam := step.Parameters[requiredParam]; !hasParam {
				errors = append(errors, fmt.Sprintf("Step %d (%s): missing required parameter '%s' for function '%s'", i, step.ID, requiredParam, step.Action))
			}
		}

		// Validate parameter references (${user.param}, ${steps.id.outputs.field}, etc.)
		for paramName, paramValue := range step.Parameters {
			if paramStr, ok := paramValue.(string); ok {
				if strings.Contains(paramStr, "${") {
					if !p.isValidParameterReference(paramStr) {
						errors = append(errors, fmt.Sprintf("Step %d (%s): invalid parameter reference '%s' in parameter '%s'", i, step.ID, paramStr, paramName))
					}
				}
			}
		}
	}

	return len(errors) == 0, errors
}


// ValidateServiceBindingsTyped validates service bindings using strongly-typed structures
func (p *MCPCatalogParser) ValidateServiceBindingsTyped(catalog *types.MCPServiceCatalog, serviceBindings map[string]types.ServiceBindingValidation, steps []types.WorkflowStepValidation) (bool, []string) {
	var errors []string
	services := catalog.Providers.Workspace.Services

	// Extract required services from workflow steps
	requiredServices := make(map[string]bool)
	for _, step := range steps {
		if parts := strings.Split(step.Action, "."); len(parts) >= 2 {
			requiredServices[parts[0]] = true
		}
	}

	// Check that all required services have bindings
	for serviceName := range requiredServices {
		if _, exists := serviceBindings[serviceName]; !exists {
			errors = append(errors, fmt.Sprintf("Missing service binding for required service: %s", serviceName))
		}
	}

	// Validate each service binding
	for serviceName, binding := range serviceBindings {
		// Check if service exists in MCP catalog
		if _, exists := services[serviceName]; !exists {
			errors = append(errors, fmt.Sprintf("Service binding '%s' not found in MCP catalog", serviceName))
			continue
		}

		// Validate OAuth configuration for Google services
		if p.isGoogleService(serviceName) {
			if binding.AuthType != "oauth2" {
				errors = append(errors, fmt.Sprintf("Service '%s' requires OAuth2 authentication", serviceName))
			}

			if binding.OAuthConfig != nil {
				requiredScopes := types.GoogleWorkspaceScopes[serviceName]
				if missing := p.findMissingScopesTyped(binding.OAuthConfig.Scopes, requiredScopes); len(missing) > 0 {
					errors = append(errors, fmt.Sprintf("Service '%s' missing required OAuth scopes: %v", serviceName, missing))
				}
			} else {
				errors = append(errors, fmt.Sprintf("Service '%s' missing OAuth configuration", serviceName))
			}
		}
	}

	return len(errors) == 0, errors
}

// Helper methods for validation

// isValidParameterReference validates parameter reference syntax
func (p *MCPCatalogParser) isValidParameterReference(paramRef string) bool {
	// Valid patterns: ${user.param}, ${steps.step_id.outputs.field}, ${computed.expr}, ${ENV_VAR}
	validPatterns := []string{
		`^\$\{user\.[a-zA-Z_][a-zA-Z0-9_]*\}$`,                             // ${user.param}
		`^\$\{steps\.[a-zA-Z_][a-zA-Z0-9_]*\.outputs\.[a-zA-Z_][a-zA-Z0-9_]*\}$`, // ${steps.step_id.outputs.field}
		`^\$\{computed\.[a-zA-Z0-9_.]+\}$`,                                 // ${computed.expr}
		`^\$\{[A-Z_][A-Z0-9_]*\}$`,                                         // ${ENV_VAR}
	}

	for _, pattern := range validPatterns {
		if matched, _ := regexp.MatchString(pattern, paramRef); matched {
			return true
		}
	}

	return false
}

// isGoogleService checks if a service is a Google Workspace service
func (p *MCPCatalogParser) isGoogleService(serviceName string) bool {
	googleServices := []string{"gmail", "docs", "drive", "calendar", "sheets"}
	for _, service := range googleServices {
		if serviceName == service {
			return true
		}
	}
	return false
}

// getRequiredScopes returns required OAuth scopes for a service (deprecated - use types.GoogleWorkspaceScopes)
func (p *MCPCatalogParser) getRequiredScopes(serviceName string) []string {
	if scopes, exists := types.GoogleWorkspaceScopes[serviceName]; exists {
		return scopes
	}
	return []string{}
}

// findMissingScopes finds missing required scopes (deprecated - use findMissingScopesTyped)
func (p *MCPCatalogParser) findMissingScopes(providedScopes []interface{}, requiredScopes []string) []string {
	var missing []string
	
	// Convert provided scopes to string slice
	providedMap := make(map[string]bool)
	for _, scope := range providedScopes {
		if scopeStr, ok := scope.(string); ok {
			providedMap[scopeStr] = true
		}
	}
	
	// Find missing required scopes
	for _, required := range requiredScopes {
		if !providedMap[required] {
			missing = append(missing, required)
		}
	}
	
	return missing
}

// findMissingScopesTyped finds missing required scopes using strongly-typed structures
func (p *MCPCatalogParser) findMissingScopesTyped(providedScopes []string, requiredScopes []string) []string {
	var missing []string
	
	// Convert provided scopes to map for efficient lookup
	providedMap := make(map[string]bool)
	for _, scope := range providedScopes {
		providedMap[scope] = true
	}
	
	// Find missing required scopes
	for _, required := range requiredScopes {
		if !providedMap[required] {
			missing = append(missing, required)
		}
	}
	
	return missing
}

// ParseParameterReference parses a parameter reference string into a structured format
func (p *MCPCatalogParser) ParseParameterReference(paramRef string) types.ParameterReference {
	result := types.ParameterReference{
		OriginalValue: paramRef,
		Type:          types.ParamRefInvalid,
		IsValid:       false,
	}

	// Check if it's a parameter reference format
	if !strings.HasPrefix(paramRef, "${") || !strings.HasSuffix(paramRef, "}") {
		return result
	}

	// Extract content between ${ and }
	content := paramRef[2 : len(paramRef)-1]
	if content == "" {
		return result
	}

	// Split by dots to get path components
	parts := strings.Split(content, ".")
	if len(parts) == 0 {
		return result
	}

	result.Source = parts[0]
	result.Path = parts

	// Determine type and validate format
	switch parts[0] {
	case "user":
		if len(parts) == 2 && p.isValidIdentifier(parts[1]) {
			result.Type = types.ParamRefUser
			result.IsValid = true
		}
	case "steps":
		if len(parts) >= 4 && parts[2] == "outputs" && p.isValidIdentifier(parts[1]) && parts[3] != "" && p.isValidIdentifier(parts[3]) {
			result.Type = types.ParamRefStep
			result.IsValid = true
		}
	case "computed":
		if len(parts) >= 2 {
			result.Type = types.ParamRefComputed
			result.IsValid = true
		}
	default:
		// Check if it's an environment variable (all caps with underscores)
		if len(parts) == 1 && p.isValidEnvVar(parts[0]) {
			result.Type = types.ParamRefEnv
			result.IsValid = true
		}
	}

	return result
}

// isValidIdentifier checks if a string is a valid identifier
func (p *MCPCatalogParser) isValidIdentifier(id string) bool {
	if len(id) == 0 {
		return false
	}
	
	// Must start with letter or underscore
	first := id[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}
	
	// Rest must be letters, digits, or underscores
	for _, char := range id[1:] {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}
	
	return true
}

// isValidEnvVar checks if a string is a valid environment variable name
func (p *MCPCatalogParser) isValidEnvVar(name string) bool {
	if len(name) == 0 {
		return false
	}
	
	// Must start with letter or underscore
	first := name[0]
	if !((first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}
	
	// Rest must be uppercase letters, digits, or underscores
	for _, char := range name[1:] {
		if !((char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}
	
	return true
}


