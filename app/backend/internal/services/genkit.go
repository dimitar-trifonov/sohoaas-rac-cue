package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"sohoaas-backend/internal/types"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/compat_oai/openai"
)

// GenkitService handles LLM interactions using Google Genkit
type GenkitService struct {
	ctx                      context.Context
	genkit                   *genkit.Genkit
	mcpService               *MCPService
	mcpParser                *MCPCatalogParser
	workflowStorage          *WorkflowStorageService
	personalCapabilitiesFlow *core.Flow[map[string]interface{}, map[string]interface{}, struct{}]
	intentGathererFlow       *core.Flow[map[string]interface{}, map[string]interface{}, struct{}]
	intentAnalystFlow        *core.Flow[IntentAnalystInput, IntentAnalystOutput, struct{}]
	workflowGeneratorFlow    *core.Flow[WorkflowGeneratorInput, WorkflowGeneratorOutput, struct{}]
	promptsDir               string
	// Pre-loaded prompts to avoid re-registration
	intentAnalystPrompt      interface{}
	workflowGeneratorPrompt  interface{}
}

// loadPrompt loads a Genkit dotprompt file with proper YAML front matter handling
// Returns the loaded prompt interface that can be executed
func (g *GenkitService) loadPrompt(promptName string) (interface{}, error) {
	promptPath := fmt.Sprintf("./prompts/%s.prompt", promptName)
	prompt, err := genkit.LoadPrompt(g.genkit, promptPath, promptName)
	if err != nil {
		return nil, fmt.Errorf("failed to load prompt %s: %v", promptName, err)
	}
	if prompt == nil {
		return nil, fmt.Errorf("empty prompt loaded for %s", promptName)
	}
	return prompt, nil
}

// NewGenkitService creates a new Genkit service instance
func NewGenkitService(apiKey string, mcpService *MCPService) *GenkitService {
	ctx := context.Background()

	// Initialize Genkit with Google GenAI plugin and prompt directory
	// Reflection port is configured via GENKIT_REFLECTION_PORT environment variable
	g, err := genkit.Init(ctx,
		genkit.WithPlugins(&openai.OpenAI{
			APIKey: apiKey,
		}),
		genkit.WithPromptDir("prompts"),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize Genkit: %v", err))
	}

	// Initialize workflow storage service
	// Use unified ARTIFACT_OUTPUT_DIR for all artifact storage
	workflowsDir := os.Getenv("ARTIFACT_OUTPUT_DIR")
	if workflowsDir == "" {
	}
	if workflowsDir == "" {
		workflowsDir = "./generated_workflows"
	}
	workflowStorage := NewWorkflowStorageService(workflowsDir)

	service := &GenkitService{
		ctx:             ctx,
		genkit:          g,
		mcpService:      mcpService,
		mcpParser:       NewMCPCatalogParser(),
		workflowStorage: workflowStorage,
		promptsDir:      "./prompts",
	}

	// Pre-load prompts to avoid re-registration during flow execution
	service.preloadPrompts()

	// Initialize all flows during startup
	service.initializeFlows()

	return service
}

// preloadPrompts loads all prompts once during initialization to avoid re-registration
func (g *GenkitService) preloadPrompts() {
	var err error
	
	// Load intent analyst prompt
	g.intentAnalystPrompt, err = g.loadPrompt("intent_analyst")
	if err != nil {
		log.Printf("Warning: Failed to preload intent_analyst prompt: %v", err)
	}
	
	// Load workflow generator prompt
	g.workflowGeneratorPrompt, err = g.loadPrompt("workflow_generator")
	if err != nil {
		log.Printf("Warning: Failed to preload workflow_generator prompt: %v", err)
	}
}

// initializeFlows creates all Genkit flows during service initialization
func (g *GenkitService) initializeFlows() {
	// Personal Capabilities Flow - Enhanced for OAuth2 + MCP Service Catalog Integration
	g.personalCapabilitiesFlow = genkit.DefineFlow(g.genkit, "personal-capabilities", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		// Extract OAuth2 token and user info
		userID, _ := input["user_id"].(string)
		oauthTokens, _ := input["oauth_tokens"].(map[string]interface{})

		// Query MCP service catalog using centralized MCPService
		mcpServices, err := g.mcpService.GetServiceCatalog()
		if err != nil {
			return nil, fmt.Errorf("failed to query MCP service catalog: %w", err)
		}

		// Check if user has Google OAuth2 token
		hasGoogleAuth := false
		if oauthTokens != nil {
			if _, exists := oauthTokens["google"]; exists {
				hasGoogleAuth = true
			}
		}

		// Build service capabilities using centralized parser
		serviceCatalog, err := g.mcpParser.BuildServiceCapabilities(mcpServices, hasGoogleAuth)
		if err != nil {
			return nil, fmt.Errorf("failed to build service capabilities: %w", err)
		}

		// Return structured capabilities instead of LLM analysis
		output := map[string]interface{}{
			"user_id":            userID,
			"service_catalog":    serviceCatalog,
			"user_capabilities":  g.buildUserCapabilities(serviceCatalog),
			"available_actions":  g.extractAvailableActions(serviceCatalog),
			"examples":           g.generateCapabilityExamples(serviceCatalog),
			"status":             "ready",
			"oauth_validated":    len(oauthTokens) > 0,
			"connected_services": g.getConnectedServiceNames(serviceCatalog),
		}

		return output, nil
	})

	// Intent Gatherer Flow
	g.intentGathererFlow = genkit.DefineFlow(g.genkit, "intent-gatherer", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		model := genkit.LookupModel(g.genkit, "openai", "gpt-4o-mini")
		if model == nil {
			return nil, fmt.Errorf("model openai/gpt-4o-mini not found")
		}

		resp, err := model.Generate(ctx, &ai.ModelRequest{
			Messages: []*ai.Message{
				{
					Content: []*ai.Part{
						ai.NewTextPart(fmt.Sprintf("Execute intent gathering with input: %v", input)),
					},
					Role: ai.RoleUser,
				},
			},
		}, nil)

		if err != nil {
			return nil, fmt.Errorf("failed to generate response: %w", err)
		}

		var output map[string]interface{}
		responseText := resp.Text()
		if err := json.Unmarshal([]byte(responseText), &output); err != nil {
			output = map[string]interface{}{
				"intent_summary": responseText,
			}
		}

		return output, nil
	})

	// Intent Analyst Flow with proper types
	g.intentAnalystFlow = genkit.DefineFlow(g.genkit, "intent-analyst", func(ctx context.Context, input IntentAnalystInput) (IntentAnalystOutput, error) {
		model := genkit.LookupModel(g.genkit, "openai", "gpt-4o-mini")
		if model == nil {
			return IntentAnalystOutput{}, fmt.Errorf("model openai/gpt-4o-mini not found")
		}

		// Extract user message and available services from simplified typed input
		userMessage := input.UserMessage
		availableServices := input.AvailableServices
		log.Printf("[DEBUG] Intent Analyst: user_message: %s", userMessage)
		log.Printf("[DEBUG] Intent Analyst: available services: %v", availableServices)

		// Template data is now directly used in the prompt formatting below

		// Use pre-loaded prompt to avoid re-registration
		intentPrompt := g.intentAnalystPrompt
		if intentPrompt == nil {
			return IntentAnalystOutput{}, fmt.Errorf("intent analyst prompt not loaded")
		}

		// Load RaC context for Intent Analyst agent
		racContextPath := "rac/agents/intent_analyst.cue"
		racContextBytes, err := os.ReadFile(racContextPath)
		if err != nil {
			log.Printf("[DEBUG] Intent Analyst: Failed to load RaC context from %s: %v", racContextPath, err)
			racContextBytes = []byte("// RaC context not available")
		}
		racContext := string(racContextBytes)

		// Execute prompt with input data (Genkit handles templating)
		inputData := map[string]interface{}{
			"user_intent":        input.UserMessage,       // Fixed: match prompt schema field name
			"available_services": input.AvailableServices, // Pass array directly to match schema
			"rac_context":        racContext,              // Inject RaC agent specification as string
		}

		// Execute prompt directly (Genkit *ai.Prompt has Execute method)
		aiPrompt, ok := intentPrompt.(*ai.Prompt)
		if !ok {
			return IntentAnalystOutput{}, fmt.Errorf("loaded prompt is not *ai.Prompt type")
		}

		resp, err := aiPrompt.Execute(ctx, ai.WithInput(inputData))
		if err != nil {
			return IntentAnalystOutput{}, fmt.Errorf("failed to generate response: %w", err)
		}

		log.Printf("[DEBUG] Intent Analyst: Using Genkit dotprompt execution")

		// Parse the response into the typed output
		responseText := resp.Text()
		log.Printf("[DEBUG] Intent Analyst: LLM response: %s", responseText)

		var output IntentAnalystOutput
		if err := json.Unmarshal([]byte(responseText), &output); err != nil {
			// Try to extract JSON from response if it's wrapped in markdown
			jsonStart := strings.Index(responseText, "{")
			jsonEnd := strings.LastIndex(responseText, "}") + 1
			if jsonStart >= 0 && jsonEnd > jsonStart {
				jsonStr := responseText[jsonStart:jsonEnd]
				if err := json.Unmarshal([]byte(jsonStr), &output); err != nil {
					log.Printf("[DEBUG] Intent Analyst: Failed to parse JSON, using fallback: %v", err)
					// Fallback to basic analysis
					output = IntentAnalystOutput{
						IsAutomationRequest: false,
						RequiredServices:    []string{},
						CanFulfill:          false,
						MissingInfo:         []string{},
						NextAction:          "need_clarification",
						Explanation:         "Failed to parse LLM response, using fallback analysis",
					}
				} else {
					// Normalize null arrays to empty arrays after successful parsing
					if output.RequiredServices == nil {
						output.RequiredServices = []string{}
					}
					if output.MissingInfo == nil {
						output.MissingInfo = []string{}
					}
				}
			} else {
				log.Printf("[DEBUG] Intent Analyst: No JSON found in response, using fallback")
				output = IntentAnalystOutput{
					IsAutomationRequest: false,
					RequiredServices:    []string{},
					CanFulfill:          false,
					MissingInfo:         []string{},
					NextAction:          "need_clarification",
					Explanation:         "No valid JSON found in LLM response, using fallback analysis",
				}
			}
		} else {
			// Normalize null arrays to empty arrays after successful parsing
			if output.RequiredServices == nil {
				output.RequiredServices = []string{}
			}
			if output.MissingInfo == nil {
				output.MissingInfo = []string{}
			}
		}

		log.Printf("[DEBUG] Intent Analyst: Parsed output: %+v", output)
		return output, nil
	})

	// Workflow Generator Flow
	g.workflowGeneratorFlow = genkit.DefineFlow(g.genkit, "workflow-generator", func(ctx context.Context, input WorkflowGeneratorInput) (WorkflowGeneratorOutput, error) {
		log.Printf("[GenkitService] === WORKFLOW GENERATOR FLOW STARTED ===")
		log.Printf("[GenkitService] Flow received RaC context: %d bytes", len(input.RacContext))
		log.Printf("[GenkitService] Flow received AvailableServices: %d bytes", len(input.AvailableServices))

		model := genkit.LookupModel(g.genkit, "openai", "gpt-4o-mini")
		if model == nil {
			return WorkflowGeneratorOutput{}, fmt.Errorf("model openai/gpt-4o-mini not found")
		}

		// Use pre-loaded prompt to avoid re-registration
		workflowPrompt := g.workflowGeneratorPrompt
		if workflowPrompt == nil {
			return WorkflowGeneratorOutput{}, fmt.Errorf("workflow generator prompt not loaded")
		}

		// Use the input data as-is - this will be the WorkflowGeneratorInput struct
		// passed from ExecuteWorkflowGeneratorAgent with RaC context included
		log.Printf("[GenkitService] Using WorkflowGeneratorInput struct with RaC context")

		// Execute prompt with input data (Genkit handles templating)
		// Use the same pattern as Intent Analyst for consistency
		aiPrompt, ok := workflowPrompt.(*ai.Prompt)
		if !ok {
			return WorkflowGeneratorOutput{}, fmt.Errorf("loaded prompt is not *ai.Prompt type")
		}
		log.Printf("[=== DEBUG ===] Workflow Generator input: %v", input)
		resp, err := aiPrompt.Execute(ctx, ai.WithInput(input))

		log.Printf("[GenkitService] Using flow-based execution with RaC context for workflow generator")

		if err != nil {
			return WorkflowGeneratorOutput{}, fmt.Errorf("failed to generate response: %w", err)
		}

		var output WorkflowGeneratorOutput
		responseText := resp.Text()
		log.Printf("[DEBUG] Workflow Generator: LLM response: %s", responseText)

		if err := json.Unmarshal([]byte(responseText), &output); err != nil {
			// Try to extract JSON from response if it's wrapped in markdown or has extra text
			jsonStart := strings.Index(responseText, "{")
			jsonEnd := strings.LastIndex(responseText, "}") + 1
			if jsonStart >= 0 && jsonEnd > jsonStart {
				jsonStr := responseText[jsonStart:jsonEnd]
				log.Printf("[DEBUG] Workflow Generator: Extracted JSON: %s", jsonStr)
				if err := json.Unmarshal([]byte(jsonStr), &output); err != nil {
					log.Printf("[DEBUG] Workflow Generator: Failed to parse extracted JSON, using fallback: %v", err)
					// Fallback for unparseable JSON - return minimal valid structure
					output = WorkflowGeneratorOutput{
						Version:        "1.0",
						Name:           "fallback_workflow",
						Description:    "Generated from unparseable LLM response",
						Steps:          []types.WorkflowStep{},
						UserParameters: make(map[string]types.UserParameter),
						Services:       make(map[string]interface{}),
					}
				}
			} else {
				log.Printf("[DEBUG] Workflow Generator: No JSON found in response, using fallback")
				// Fallback for unparseable JSON - return minimal valid structure
				output = WorkflowGeneratorOutput{
					Version:        "1.0",
					Name:           "fallback_workflow",
					Description:    "No valid JSON found in LLM response",
					Steps:          []types.WorkflowStep{},
					UserParameters: make(map[string]types.UserParameter),
					Services:       make(map[string]interface{}),
				}
			}
		}

		// Normalize null arrays to empty arrays after successful parsing
		if output.Steps == nil {
			output.Steps = []types.WorkflowStep{}
		}
		if output.UserParameters == nil {
			output.UserParameters = make(map[string]types.UserParameter)
		}
		if output.Services == nil {
			output.Services = make(map[string]interface{})
		}

		log.Printf("[DEBUG] Workflow Generator: Parsed output: %+v", output)
		return output, nil
	})
}

// buildUserCapabilities creates structured user capabilities from service catalog (using unified parser)
func (g *GenkitService) buildUserCapabilities(serviceCatalog map[string]interface{}) map[string]interface{} {
	// Use unified MCPCatalogParser to eliminate duplicate logic
	serviceCapabilities, err := g.mcpParser.BuildServiceCapabilities(serviceCatalog, false)
	if err != nil {
		// Fallback to empty capabilities if parsing fails
		return make(map[string]interface{})
	}

	return serviceCapabilities
}

// extractAvailableActions extracts all available actions from connected services
func (g *GenkitService) extractAvailableActions(serviceCatalog map[string]interface{}) []string {
	var actions []string

	for _, serviceData := range serviceCatalog {
		if serviceMap, ok := serviceData.(map[string]interface{}); ok {
			if actionsList, ok := serviceMap["actions"].([]interface{}); ok {
				for _, action := range actionsList {
					if actionStr, ok := action.(string); ok {
						actions = append(actions, actionStr)
					}
				}
			}
		}
	}

	return actions
}

// generateCapabilityExamples creates example automation scenarios
func (g *GenkitService) generateCapabilityExamples(serviceCatalog map[string]interface{}) []string {
	examples := []string{}

	// Generate examples dynamically based on available services from MCP catalog
	serviceNames := g.getConnectedServiceNames(serviceCatalog)

	// Generate generic examples for any available services
	if len(serviceNames) > 0 {
		examples = append(examples,
			"Automate your daily workflows",
			"Create recurring task automations",
			"Set up notification systems",
			"Build document processing workflows")
	} else {
		examples = append(examples, "Connect Google Workspace to see automation examples")
	}

	return examples
}

// getConnectedServiceNames extracts service names from service catalog (using unified parser)
func (g *GenkitService) getConnectedServiceNames(serviceCatalog map[string]interface{}) []string {
	// Use unified MCPCatalogParser to eliminate duplicate logic
	serviceNames, err := g.mcpParser.ExtractServiceNames(serviceCatalog)
	if err != nil {
		// Fallback to empty list if parsing fails
		return []string{}
	}
	return serviceNames
}

// convertJSONToCUE converts JSON workflow to CUE format following deterministic_workflow.cue schema
func (g *GenkitService) convertJSONToCUE(workflowJSON map[string]interface{}) string {
	log.Printf("[GenkitService] Converting JSON workflow to CUE format")

	// Extract basic workflow information
	workflowName := g.extractStringField(workflowJSON, "workflow_name", "Generated Workflow")
	description := g.extractStringField(workflowJSON, "description", "Auto-generated workflow")

	// Build CUE structure
	var cueBuilder strings.Builder

	// Package declaration
	cueBuilder.WriteString("package workflow\n\n")

	// Embed schema content directly instead of import (more robust)
	schemaContent, err := g.loadSchemaContent()
	if err != nil {
		log.Printf("[GenkitService] Warning: Failed to load schema content: %v", err)
		// Fallback to import if schema loading fails
		cueBuilder.WriteString("import \"../../rac/schemas.cue\"\n\n")
	} else {
		// Embed schema directly
		cueBuilder.WriteString("// === EMBEDDED RAC SCHEMAS ===\n")
		cueBuilder.WriteString(schemaContent)
		cueBuilder.WriteString("\n// === END EMBEDDED SCHEMAS ===\n\n")
	}

	// Main workflow definition
	cueBuilder.WriteString("workflow: #DeterministicWorkflow & {\n")
	cueBuilder.WriteString(fmt.Sprintf("\tversion: \"1.0.0\"\n"))
	cueBuilder.WriteString(fmt.Sprintf("\tname: %q\n", workflowName))
	cueBuilder.WriteString(fmt.Sprintf("\tdescription: %q\n", description))
	
	// Add original_intent if present
	if originalIntent, exists := workflowJSON["original_intent"]; exists && originalIntent != nil {
		if intentStr, ok := originalIntent.(string); ok && intentStr != "" {
			cueBuilder.WriteString(fmt.Sprintf("\toriginal_intent: %q\n", intentStr))
		}
	}
	cueBuilder.WriteString("\n")

	// Convert steps
	stepsStr := g.convertJSONStepsToCUE(workflowJSON)
	cueBuilder.WriteString(stepsStr)
	cueBuilder.WriteString("\n")

	// Convert user parameters
	userParamsStr := g.convertJSONUserParametersToCUE(workflowJSON)
	cueBuilder.WriteString(userParamsStr)
	cueBuilder.WriteString("\n")

	// Convert service bindings
	serviceBindingsStr := g.convertJSONServiceBindingsToCUE(workflowJSON)
	cueBuilder.WriteString(serviceBindingsStr)
	cueBuilder.WriteString("\n")

	// Add execution config
	cueBuilder.WriteString("\texecution_config: {\n")
	cueBuilder.WriteString("\t\tmode: \"sequential\"\n")
	cueBuilder.WriteString("\t\ttimeout: \"5m\"\n")
	cueBuilder.WriteString("\t\tenvironment: \"development\"\n")
	cueBuilder.WriteString("\t}\n")

	// Close workflow definition
	cueBuilder.WriteString("}\n")

	result := cueBuilder.String()
	log.Printf("[GenkitService] Generated CUE workflow (%d characters)", len(result))
	return result
}

// extractStringField safely extracts a string field from JSON with default fallback
func (g *GenkitService) extractStringField(data map[string]interface{}, field, defaultValue string) string {
	if value, exists := data[field]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

// convertJSONStepsToCUE converts JSON steps array to CUE format
func (g *GenkitService) convertJSONStepsToCUE(workflowJSON map[string]interface{}) string {
	var stepsBuilder strings.Builder
	stepsBuilder.WriteString("\tsteps: [\n")

	if stepsData, exists := workflowJSON["steps"]; exists {
		if stepsArray, ok := stepsData.([]interface{}); ok {
			for i, stepData := range stepsArray {
				if stepMap, ok := stepData.(map[string]interface{}); ok {
					stepCUE := g.convertSingleStepToCUE(stepMap, i)
					stepsBuilder.WriteString(stepCUE)
					if i < len(stepsArray)-1 {
						stepsBuilder.WriteString(",")
					}
					stepsBuilder.WriteString("\n")
				}
			}
		}
	}

	stepsBuilder.WriteString("\t]")
	return stepsBuilder.String()
}

// convertSingleStepToCUE converts a single step from JSON to CUE format
func (g *GenkitService) convertSingleStepToCUE(stepData map[string]interface{}, index int) string {
	var stepBuilder strings.Builder

	// Extract step fields
	stepID := g.extractStringField(stepData, "id", fmt.Sprintf("step_%d", index+1))
	stepName := g.extractStringField(stepData, "name", fmt.Sprintf("Step %d", index+1))
	action := g.extractStringField(stepData, "action", "unknown.action")
	description := g.extractStringField(stepData, "description", "")

	stepBuilder.WriteString("\t\t{\n")
	stepBuilder.WriteString(fmt.Sprintf("\t\t\tid: %q\n", stepID))
	stepBuilder.WriteString(fmt.Sprintf("\t\t\tname: %q\n", stepName))
	stepBuilder.WriteString(fmt.Sprintf("\t\t\taction: %q\n", action))

	if description != "" {
		stepBuilder.WriteString(fmt.Sprintf("\t\t\tdescription: %q\n", description))
	}

	// Convert parameters
	if paramsData, exists := stepData["parameters"]; exists {
		if paramsMap, ok := paramsData.(map[string]interface{}); ok {
			stepBuilder.WriteString("\t\t\tparameters: {\n")
			for key, value := range paramsMap {
				stepBuilder.WriteString(fmt.Sprintf("\t\t\t\t%s: %s\n", key, g.formatCUEValue(value)))
			}
			stepBuilder.WriteString("\t\t\t}\n")
		}
	}

	// Convert dependencies
	if depsData, exists := stepData["depends_on"]; exists {
		if depsArray, ok := depsData.([]interface{}); ok && len(depsArray) > 0 {
			stepBuilder.WriteString("\t\t\tdepends_on: [")
			for i, dep := range depsArray {
				if depStr, ok := dep.(string); ok {
					stepBuilder.WriteString(fmt.Sprintf("%q", depStr))
					if i < len(depsArray)-1 {
						stepBuilder.WriteString(", ")
					}
				}
			}
			stepBuilder.WriteString("]\n")
		}
	}

	// Add timeout if specified
	if timeout := g.extractStringField(stepData, "timeout", ""); timeout != "" {
		stepBuilder.WriteString(fmt.Sprintf("\t\t\ttimeout: %q\n", timeout))
	}

	stepBuilder.WriteString("\t\t}")
	return stepBuilder.String()
}

// convertJSONUserParametersToCUE converts JSON user parameters to CUE format
func (g *GenkitService) convertJSONUserParametersToCUE(workflowJSON map[string]interface{}) string {
	var paramsBuilder strings.Builder
	paramsBuilder.WriteString("\tuser_parameters: {\n")

	if paramsData, exists := workflowJSON["user_parameters"]; exists {
		if paramsMap, ok := paramsData.(map[string]interface{}); ok {
			for paramName, paramData := range paramsMap {
				if paramMap, ok := paramData.(map[string]interface{}); ok {
					paramCUE := g.convertSingleUserParameterToCUE(paramMap)
					paramsBuilder.WriteString(fmt.Sprintf("\t\t%s: %s\n", paramName, paramCUE))
				}
			}
		}
	}

	paramsBuilder.WriteString("\t}")
	return paramsBuilder.String()
}

// convertSingleUserParameterToCUE converts a single user parameter from JSON to CUE format
func (g *GenkitService) convertSingleUserParameterToCUE(paramData map[string]interface{}) string {
	var paramBuilder strings.Builder

	paramBuilder.WriteString("{\n")

	// Required fields
	paramType := g.extractStringField(paramData, "type", "string")
	prompt := g.extractStringField(paramData, "prompt", "Enter value")

	paramBuilder.WriteString(fmt.Sprintf("\t\t\ttype: %q\n", paramType))
	paramBuilder.WriteString(fmt.Sprintf("\t\t\tprompt: %q\n", prompt))

	// Required flag
	if requiredData, exists := paramData["required"]; exists {
		if required, ok := requiredData.(bool); ok {
			paramBuilder.WriteString(fmt.Sprintf("\t\t\trequired: %t\n", required))
		}
	} else {
		paramBuilder.WriteString("\t\t\trequired: true\n")
	}

	// Optional fields
	if description := g.extractStringField(paramData, "description", ""); description != "" {
		paramBuilder.WriteString(fmt.Sprintf("\t\t\tdescription: %q\n", description))
	}

	if validation := g.extractStringField(paramData, "validation", ""); validation != "" {
		paramBuilder.WriteString(fmt.Sprintf("\t\t\tvalidation: %q\n", validation))
	}

	if placeholder := g.extractStringField(paramData, "placeholder", ""); placeholder != "" {
		paramBuilder.WriteString(fmt.Sprintf("\t\t\tplaceholder: %q\n", placeholder))
	}

	// Handle default field - this was missing!
	if defaultValue, exists := paramData["default"]; exists && defaultValue != nil {
		switch v := defaultValue.(type) {
		case string:
			paramBuilder.WriteString(fmt.Sprintf("\t\t\tdefault: %q\n", v))
		case bool:
			paramBuilder.WriteString(fmt.Sprintf("\t\t\tdefault: %t\n", v))
		case float64:
			paramBuilder.WriteString(fmt.Sprintf("\t\t\tdefault: %g\n", v))
		case int:
			paramBuilder.WriteString(fmt.Sprintf("\t\t\tdefault: %d\n", v))
		default:
			// For other types, convert to string
			paramBuilder.WriteString(fmt.Sprintf("\t\t\tdefault: %q\n", fmt.Sprintf("%v", v)))
		}
	}

	paramBuilder.WriteString("\t\t}")
	return paramBuilder.String()
}

// convertJSONServiceBindingsToCUE converts JSON service bindings to CUE format
func (g *GenkitService) convertJSONServiceBindingsToCUE(workflowJSON map[string]interface{}) string {
	var servicesBuilder strings.Builder
	servicesBuilder.WriteString("\tservice_bindings: {\n")

	if servicesData, exists := workflowJSON["services"]; exists {
		if servicesArray, ok := servicesData.([]interface{}); ok {
			for _, serviceData := range servicesArray {
				if serviceMap, ok := serviceData.(map[string]interface{}); ok {
					serviceName := g.extractStringField(serviceMap, "service", "unknown")
					serviceCUE := g.convertSingleServiceBindingToCUE(serviceMap)
					servicesBuilder.WriteString(fmt.Sprintf("\t\t%s: %s\n", serviceName, serviceCUE))
				}
			}
		}
	}

	servicesBuilder.WriteString("\t}")
	return servicesBuilder.String()
}

// convertSingleServiceBindingToCUE converts a single service binding from JSON to CUE format
func (g *GenkitService) convertSingleServiceBindingToCUE(serviceData map[string]interface{}) string {
	var serviceBuilder strings.Builder

	serviceBuilder.WriteString("{\n")
	serviceBuilder.WriteString("\t\t\ttype: \"mcp_service\"\n")
	serviceBuilder.WriteString("\t\t\tprovider: \"workspace\"\n")

	// Auth configuration
	serviceBuilder.WriteString("\t\t\tauth: {\n")
	serviceBuilder.WriteString("\t\t\t\ttype: \"oauth2\"\n")

	// OAuth scopes
	if scopesData, exists := serviceData["oauth_scopes"]; exists {
		if scopesArray, ok := scopesData.([]interface{}); ok {
			serviceBuilder.WriteString("\t\t\t\tscopes: [")
			for i, scope := range scopesArray {
				if scopeStr, ok := scope.(string); ok {
					serviceBuilder.WriteString(fmt.Sprintf("%q", scopeStr))
					if i < len(scopesArray)-1 {
						serviceBuilder.WriteString(", ")
					}
				}
			}
			serviceBuilder.WriteString("]\n")
		}
	}

	serviceBuilder.WriteString("\t\t\t}\n")
	serviceBuilder.WriteString("\t\t}")
	return serviceBuilder.String()
}

// sanitizeCUEContent removes illegal characters and formatting from CUE content
func (g *GenkitService) sanitizeCUEContent(cueContent string) string {
	// Remove markdown code block markers first
	sanitized := strings.ReplaceAll(cueContent, "```cue", "")
	sanitized = strings.ReplaceAll(sanitized, "```", "")

	// Remove backticks that cause CUE parsing errors - replace with double quotes
	sanitized = strings.ReplaceAll(sanitized, "`", `"`)

	// Remove unicode backticks
	sanitized = strings.ReplaceAll(sanitized, "\u0060", `"`)

	// Fix any triple quotes that might cause multiline string issues
	sanitized = strings.ReplaceAll(sanitized, `'''`, `"""`)
	sanitized = strings.ReplaceAll(sanitized, `""""`, `"""`)

	return sanitized
}

// formatCUEValue formats a JSON value for CUE syntax
func (g *GenkitService) formatCUEValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		// Check if it's a parameter reference
		if strings.HasPrefix(v, "${user.") || strings.HasPrefix(v, "${steps.") {
			return fmt.Sprintf("\"${%s}\"", strings.TrimPrefix(strings.TrimSuffix(v, "}"), "${"))
		}
		return fmt.Sprintf("%q", v)
	case float64:
		return fmt.Sprintf("%.0f", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%q", fmt.Sprintf("%v", v))
	}
}



// loadSchemaContent loads the RaC schema content for embedding in generated CUE files
func (g *GenkitService) loadSchemaContent() (string, error) {
	// Get RaC context path from environment or use default
	racContextPath := os.Getenv("RAC_CONTEXT_PATH")
	if racContextPath == "" {
		racContextPath = "../../rac" // Default relative path
	}

	schemaPath := filepath.Join(racContextPath, "schemas.cue")

	// Read schema file content
	content, err := os.ReadFile(schemaPath)
	if err != nil {
		return "", fmt.Errorf("failed to read schema file %s: %w", schemaPath, err)
	}

	// Convert to string and remove package declaration (we'll use our own)
	schemaContent := string(content)

	// Remove the package declaration line since we're embedding
	lines := strings.Split(schemaContent, "\n")
	var filteredLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "package ") {
			filteredLines = append(filteredLines, line)
		}
	}

	return strings.Join(filteredLines, "\n"), nil
}
