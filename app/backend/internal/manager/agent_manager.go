package manager

import (
	"fmt"
	"log"
	"sync"
	"time"

	"sohoaas-backend/internal/services"
	"sohoaas-backend/internal/types"
)

// AgentManager coordinates all agents and manages workflow state
type AgentManager struct {
	genkitService  *services.GenkitService
	mcpService     *services.MCPService
	mcpParser      *services.MCPCatalogParser

	serviceCatalog types.ServiceCatalog
	agents         map[string]*types.Agent
	mu             sync.RWMutex
}

// NewAgentManager creates a new Agent Manager instance
func NewAgentManager(genkitService *services.GenkitService, mcpService *services.MCPService) *AgentManager {
	am := &AgentManager{
		genkitService:  genkitService,
		mcpService:     mcpService,
		mcpParser:      services.NewMCPCatalogParser(),
		serviceCatalog: types.ServiceCatalog{Services: make(map[string]types.ServiceSchema)},
		agents:         make(map[string]*types.Agent),
	}
	
	// Load service catalog from MCP (single source of truth)
	am.loadServiceCatalogFromMCP()
	
	// Initialize all agents
	am.initializeAgents()
	
	return am
}

// loadServiceCatalogFromMCP loads the service catalog from MCP service (single source of truth)
func (am *AgentManager) loadServiceCatalogFromMCP() {
	log.Printf("[AgentManager] Loading service catalog from MCP...")
	
	// Get MCP catalog
	mcpCatalog, err := am.mcpService.GetServiceCatalog()
	if err != nil {
		log.Printf("[AgentManager] Warning: Failed to load MCP catalog: %v", err)
		return
	}
	
	// Parse services from MCP catalog
	servicesData, err := am.mcpParser.ParseServicesFromCatalog(mcpCatalog)
	if err != nil {
		log.Printf("[AgentManager] Warning: Failed to parse MCP services: %v", err)
		return
	}
	
	// Convert to ServiceSchema format
	serviceSchemas := make(map[string]types.ServiceSchema)
	for serviceName, serviceData := range servicesData {
		if serviceMap, ok := serviceData.(map[string]interface{}); ok {
			schema := types.ServiceSchema{
				ServiceName: serviceName,
				Actions:     make(map[string]types.ActionSchema),
				Status:      "available",
			}
			
			// Extract functions as actions with complete parameter schemas
			if functions, ok := serviceMap["functions"].(map[string]interface{}); ok {
				for actionName, functionData := range functions {
					actionSchema := types.ActionSchema{
						ActionName:     actionName,
						Description:    fmt.Sprintf("%s action for %s", actionName, serviceName),
						RequiredFields: []types.FieldSchema{},
						OptionalFields: []types.FieldSchema{},
					}
					
					// Extract parameter schemas from function definition
					if functionMap, ok := functionData.(map[string]interface{}); ok {
						actionSchema = am.extractParameterSchemas(actionSchema, functionMap)
					}
					
					schema.Actions[actionName] = actionSchema
				}
			}
			
			serviceSchemas[serviceName] = schema
		}
	}
	
	// Store in Agent Manager (thread-safe)
	am.mu.Lock()
	am.serviceCatalog.Services = serviceSchemas
	am.mu.Unlock()
	
	log.Printf("[AgentManager] ✅ Loaded %d services from MCP as single source of truth", len(serviceSchemas))
}

// extractParameterSchemas extracts parameter schemas from MCP function definition
func (am *AgentManager) extractParameterSchemas(actionSchema types.ActionSchema, functionMap map[string]interface{}) types.ActionSchema {
	// Extract description if available
	if description, ok := functionMap["description"].(string); ok && description != "" {
		actionSchema.Description = description
	}
	
	// Extract required fields from MCP structure
	if requiredFields, ok := functionMap["required_fields"].([]interface{}); ok {
		for _, field := range requiredFields {
			if fieldName, ok := field.(string); ok {
				fieldSchema := types.FieldSchema{
					FieldName:       fieldName,
					FieldType:       "string", // default, will infer from example
					PlaceholderType: "USER_INPUT",
					ValidationRules: []string{"required"},
					Description:     fmt.Sprintf("Required parameter: %s", fieldName),
				}
				
				// Try to infer type from example_payload
				if examplePayload, ok := functionMap["example_payload"].(map[string]interface{}); ok {
					if exampleValue, exists := examplePayload[fieldName]; exists {
						fieldSchema.FieldType = am.inferFieldType(exampleValue)
						fieldSchema.ExampleValue = fmt.Sprintf("%v", exampleValue)
					}
				}
				
				actionSchema.RequiredFields = append(actionSchema.RequiredFields, fieldSchema)
			}
		}
	}
	
	return actionSchema
}

// inferFieldType infers field type from example value
func (am *AgentManager) inferFieldType(value interface{}) string {
	switch value.(type) {
	case string:
		return "string"
	case int, int32, int64:
		return "integer"
	case float32, float64:
		return "number"
	case bool:
		return "boolean"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return "string" // default fallback
	}
}



// initializeAgents sets up all 4 core agents
func (am *AgentManager) initializeAgents() {
	agents := []*types.Agent{
		{
			ID:   "personal_capabilities",
			Name: "Personal Capabilities Agent",
			State: "ready",
			Capabilities: []string{
				"mcp_service_discovery",
				"capability_mapping",
				"schema_discovery",
			},
		},
		{
			ID:   "intent_gatherer",
			Name: "Intent Gatherer Agent",
			State: "ready",
			Capabilities: []string{
				"workflow_discovery",
				"multi_turn_conversation",
				"pattern_identification",
			},
		},
		{
			ID:   "intent_analyst",
			Name: "Intent Analyst Agent",
			State: "ready",
			Capabilities: []string{
				"intent_validation",
				"parameter_extraction",
				"data_requirement_analysis",
			},
		},
		{
			ID:   "workflow_generator",
			Name: "Workflow Generator Agent",
			State: "ready",
			Capabilities: []string{
				"deterministic_workflow_generation",
				"cue_file_creation",
				"service_binding",
			},
		},
	}
	
	am.mu.Lock()
	for _, agent := range agents {
		am.agents[agent.ID] = agent
	}
	am.mu.Unlock()
	
	log.Printf("Initialized %d agents", len(agents))
}

// ProcessUserMessage processes a user message through the agent pipeline
func (am *AgentManager) ProcessUserMessage(userID, message string, conversationHistory []types.ConversationMessage, user *types.User) (*types.AgentResponse, error) {
	// Prepare input for Intent Gatherer
	input := map[string]interface{}{
		"user_message":         message,
		"conversation_history": conversationHistory,
		"discovery_phase":      "pattern", // Start with pattern discovery
		"collected_intent":     map[string]interface{}{},
	}

	return am.genkitService.ExecuteIntentGathererAgent(input)
}

// GetPersonalCapabilities retrieves user's personal capabilities
func (am *AgentManager) GetPersonalCapabilities(userID string, user *types.User) (*types.AgentResponse, error) {
	input := map[string]interface{}{
		"user_id":            userID,
		"oauth_tokens":       user.OAuthTokens,
		"connected_services": user.ConnectedServices,
	}
	
	// Always add MCP service catalog information for OAuth2 integration
	catalog, err := am.mcpService.GetServiceCatalog()
	if err != nil {
		log.Printf("[AgentManager] Failed to get MCP catalog: %v", err)
		input["mcp_servers"] = []map[string]interface{}{}
	} else {
		capabilities, err := am.mcpParser.BuildUserCapabilities(catalog, user.ConnectedServices)
		if err != nil {
			log.Printf("[AgentManager] Failed to build capabilities: %v", err)
			input["mcp_servers"] = []map[string]interface{}{}
		} else {
			input["mcp_servers"] = capabilities
		}
	}
	input["service_schemas"] = am.GetServiceSchemas()
	
	return am.genkitService.ExecutePersonalCapabilitiesAgent(input)
}

// GetServiceCatalog returns the current service catalog
func (am *AgentManager) GetServiceCatalog() types.ServiceCatalog {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.serviceCatalog
}

// GetServiceSchemas returns the service schemas for agent coordination
func (am *AgentManager) GetServiceSchemas() map[string]types.ServiceSchema {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.serviceCatalog.Services
}

// ValidateServices checks if services exist in the catalog
// Returns (allExist bool, missingServices []string)
func (am *AgentManager) ValidateServices(serviceNames ...string) (bool, []string) {
	am.mu.RLock()
	defer am.mu.RUnlock()
	
	var missingServices []string
	for _, serviceName := range serviceNames {
		if _, exists := am.serviceCatalog.Services[serviceName]; !exists {
			missingServices = append(missingServices, serviceName)
		}
	}
	
	return len(missingServices) == 0, missingServices
}



// AnalyzeIntent analyzes and validates a workflow intent
func (am *AgentManager) AnalyzeIntent(userID string, workflowIntent *types.WorkflowIntent, user *types.User) (*types.AgentResponse, error) {
	start := time.Now()
	log.Printf("[AgentManager] Starting intent analysis for user %s", userID)
	
	// Validate input parameters
	if workflowIntent == nil {
		log.Printf("[AgentManager] ERROR: workflow intent is nil for user %s", userID)
		return nil, fmt.Errorf("workflow intent cannot be nil")
	}
	
	if user == nil {
		log.Printf("[AgentManager] ERROR: user is nil for user %s", userID)
		return nil, fmt.Errorf("user cannot be nil")
	}
	
	// Build input with service catalog from Agent Manager
	// The Agent Manager provides the available Google Workspace services regardless of user.ConnectedServices
	catalog, err := am.mcpService.GetServiceCatalog()
	if err != nil {
		log.Printf("[AgentManager] Failed to get MCP catalog: %v", err)
		return nil, fmt.Errorf("failed to get service catalog: %w", err)
	}
	userCapabilities, err := am.mcpParser.BuildUserCapabilities(catalog, user.ConnectedServices)
	if err != nil {
		log.Printf("[AgentManager] Failed to build capabilities: %v", err)
		return nil, fmt.Errorf("failed to build user capabilities: %w", err)
	}
	serviceSchemas := am.GetServiceSchemas()
	
	log.Printf("[AgentManager] Providing Intent Analysis with %d service schemas available for user %s", 
		len(serviceSchemas), userID)
	
	input := map[string]interface{}{
		"user_id":           userID,
		"workflow_intent":   workflowIntent,
		"user_capabilities": userCapabilities,
		"service_schemas":   serviceSchemas,
		"oauth_tokens":      user.OAuthTokens,
	}
	
	// Execute Intent Analyst Agent
	response, err := am.genkitService.ExecuteIntentAnalystAgent(input)
	
	duration := time.Since(start)
	if err != nil {
		log.Printf("[AgentManager] ERROR: Intent analysis failed for user %s after %v: %v", userID, duration, err)
		return nil, fmt.Errorf("intent analysis failed: %w", err)
	}
	
	if response.Error != "" {
		log.Printf("[AgentManager] WARNING: Intent analysis returned error for user %s: %s", userID, response.Error)
	} else {
		log.Printf("[AgentManager] SUCCESS: Intent analysis completed for user %s in %v", userID, duration)
	}
	
	return response, nil
}

// GenerateWorkflow generates a deterministic workflow from validated intent
func (am *AgentManager) GenerateWorkflow(userID string, validatedIntent map[string]interface{}, user *types.User) (*types.AgentResponse, error) {
	start := time.Now()
	log.Printf("[AgentManager] Starting workflow generation for user %s", userID)
	
	// Validate input parameters
	if validatedIntent == nil {
		log.Printf("[AgentManager] ERROR: validated intent is nil for user %s", userID)
		return nil, fmt.Errorf("validated intent cannot be nil")
	}
	
	if user == nil {
		log.Printf("[AgentManager] ERROR: user is nil for user %s", userID)
		return nil, fmt.Errorf("user cannot be nil")
	}
	
	// Validate required services from intent
	requiredServices, ok := validatedIntent["required_services"].([]string)
	if !ok {
		log.Printf("[AgentManager] WARNING: no required_services found in validated intent for user %s", userID)
	} else {
		// Validate all required services are available
		allAvailable, missingServices := am.ValidateServices(requiredServices...)
		if !allAvailable {
			log.Printf("[AgentManager] ERROR: Missing required services for user %s: %v", userID, missingServices)
			return &types.AgentResponse{
				AgentID: "workflow_generator",
				Error:   fmt.Sprintf("Required services not available: %v", missingServices),
			}, nil
		}
		log.Printf("[AgentManager] All required services available for user %s: %v", userID, requiredServices)
	}
	
	// Build input with comprehensive data
	catalog, err := am.mcpService.GetServiceCatalog()
	if err != nil {
		log.Printf("[AgentManager] Failed to get MCP catalog: %v", err)
		return nil, fmt.Errorf("failed to get service catalog: %w", err)
	}
	userCapabilities, err := am.mcpParser.BuildUserCapabilities(catalog, user.ConnectedServices)
	if err != nil {
		log.Printf("[AgentManager] Failed to build capabilities: %v", err)
		return nil, fmt.Errorf("failed to build user capabilities: %w", err)
	}
	serviceSchemas := am.GetServiceSchemas()
	
	input := map[string]interface{}{
		"user_id":          userID,
		"validated_intent": validatedIntent,
		"user_capabilities": userCapabilities,
		"service_schemas":  serviceSchemas,
		"oauth_tokens":     user.OAuthTokens,
	}
	
	// Execute Workflow Generator Agent
	response, err := am.genkitService.ExecuteWorkflowGeneratorAgent(input)
	
	duration := time.Since(start)
	if err != nil {
		log.Printf("[AgentManager] ERROR: Workflow generation failed for user %s after %v: %v", userID, duration, err)
		return nil, fmt.Errorf("workflow generation failed: %w", err)
	}
	
	if response.Error != "" {
		log.Printf("[AgentManager] WARNING: Workflow generation returned error for user %s: %s", userID, response.Error)
	} else {
		log.Printf("[AgentManager] SUCCESS: Workflow generation completed for user %s in %v", userID, duration)
		
		// Log workflow details if available
		if response.Output != nil {
			if workflowName, exists := response.Output["workflow_name"].(string); exists {
				log.Printf("[AgentManager] Generated workflow: %s for user %s", workflowName, userID)
			}
		}
	}
	
	return response, nil
}



// GetAgents returns all registered agents
func (am *AgentManager) GetAgents() map[string]*types.Agent {
	am.mu.RLock()
	defer am.mu.RUnlock()
	
	agents := make(map[string]*types.Agent)
	for id, agent := range am.agents {
		agents[id] = agent
	}
	
	return agents
}

// GetAgent returns a specific agent by ID
func (am *AgentManager) GetAgent(agentID string) (*types.Agent, bool) {
	am.mu.RLock()
	defer am.mu.RUnlock()
	
	agent, exists := am.agents[agentID]
	return agent, exists
}

// Shutdown gracefully shuts down the Agent Manager
func (am *AgentManager) Shutdown() {
	log.Println("Agent Manager shutdown complete")
}
