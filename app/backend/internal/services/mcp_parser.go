package services

import (
	"fmt"
)

// MCPCatalogParser provides centralized MCP service catalog parsing utilities
type MCPCatalogParser struct{}

// NewMCPCatalogParser creates a new MCP catalog parser
func NewMCPCatalogParser() *MCPCatalogParser {
	return &MCPCatalogParser{}
}

// ParseServicesFromCatalog extracts services from MCP catalog structure
// Expected structure: {"providers": {"workspace": {"services": {...}}}}
func (p *MCPCatalogParser) ParseServicesFromCatalog(mcpCatalog map[string]interface{}) (map[string]interface{}, error) {
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
func (p *MCPCatalogParser) BuildUserCapabilities(mcpCatalog map[string]interface{}, connectedServices []string) ([]map[string]interface{}, error) {
	servicesData, err := p.ParseServicesFromCatalog(mcpCatalog)
	if err != nil {
		return nil, err
	}
	
	var capabilities []map[string]interface{}
	
	for _, serviceName := range connectedServices {
		if serviceData, exists := servicesData[serviceName]; exists {
			if serviceMap, ok := serviceData.(map[string]interface{}); ok {
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
func (p *MCPCatalogParser) BuildServiceCapabilities(mcpCatalog map[string]interface{}, hasGoogleAuth bool) (map[string]interface{}, error) {
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

// ValidateWorkflowServices validates workflow services against MCP catalog
func (p *MCPCatalogParser) ValidateWorkflowServices(mcpCatalog map[string]interface{}, workflow *ParsedWorkflow) error {
	servicesData, err := p.ParseServicesFromCatalog(mcpCatalog)
	if err != nil {
		return err
	}
	
	for i, step := range workflow.Steps {
		// Validate service exists in MCP catalog
		serviceData, exists := servicesData[step.Service]
		if !exists {
			return fmt.Errorf("unknown service '%s' in step %d (%s) - service not found in MCP catalog", step.Service, i, step.ID)
		}
		
		// Validate action exists for service in MCP catalog
		serviceMap, ok := serviceData.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid service data for '%s' in MCP catalog", step.Service)
		}
		
		functions, ok := serviceMap["functions"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("no functions found for service '%s' in MCP catalog", step.Service)
		}
		
		_, actionExists := functions[step.Action]
		if !actionExists {
			return fmt.Errorf("unknown action '%s' for service '%s' in step %d (%s) - action not found in MCP catalog", step.Action, step.Service, i, step.ID)
		}
	}
	
	return nil
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

// BuildAvailableServicesSection creates the available services section for prompts (eliminates duplicate logic)
func (p *MCPCatalogParser) BuildAvailableServicesSection(mcpCatalog map[string]interface{}) (string, error) {
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
