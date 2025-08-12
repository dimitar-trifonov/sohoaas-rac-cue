package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"sohoaas-backend/internal/types"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
)

// getInputKeys returns the keys of a map[string]interface{} for logging
func getInputKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Helper function to get map keys for debugging
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func getString(m map[string]interface{}, key string) string {
	if val, exists := m[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getMap(m map[string]interface{}, key string) map[string]interface{} {
	if val, exists := m[key]; exists {
		if mapVal, ok := val.(map[string]interface{}); ok {
			return mapVal
		}
	}
	return make(map[string]interface{})
}

// Simplified Intent Analyst Input/Output Types for Genkit compatibility
type IntentAnalystInput struct {
	UserMessage      string `json:"user_message"`
	AvailableServices []string `json:"available_services"`
}

type IntentAnalystOutput struct {
	IsAutomationRequest bool     `json:"is_automation_request"`
	RequiredServices    []string `json:"required_services"`
	CanFulfill          bool     `json:"can_fulfill"`
	MissingInfo         []string `json:"missing_info"`
	NextAction          string   `json:"next_action"`
	Explanation         string   `json:"explanation"`
}

type WorkflowGeneratorInput struct {
	UserInput        string                 `json:"user_input"`
	ValidatedIntent  map[string]interface{} `json:"validated_intent"`
	AvailableServices string                `json:"available_services"`
}

type WorkflowGeneratorOutput struct {
	Version        string                           `json:"version"`
	Name           string                           `json:"name"`
	Description    string                           `json:"description"`
	Steps          []types.WorkflowStep             `json:"steps"`
	UserParameters map[string]types.UserParameter   `json:"user_parameters"`
	Services       map[string]interface{}           `json:"services"`
}

// GenkitService handles LLM interactions using Google Genkit
type GenkitService struct {
	genkit                   *genkit.Genkit
	ctx                      context.Context
	mcpService               *MCPService
	mcpParser                *MCPCatalogParser

	workflowStorage          *WorkflowStorageService
	personalCapabilitiesFlow *core.Flow[map[string]interface{}, map[string]interface{}, struct{}]
	intentGathererFlow       *core.Flow[map[string]interface{}, map[string]interface{}, struct{}]
	intentAnalystFlow        *core.Flow[IntentAnalystInput, IntentAnalystOutput, struct{}]
	workflowGeneratorFlow    *core.Flow[WorkflowGeneratorInput, WorkflowGeneratorOutput, struct{}]
	promptsDir               string
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
		genkit.WithPlugins(&googlegenai.GoogleAI{
			APIKey: apiKey,
		}),
		genkit.WithPromptDir("prompts"),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize Genkit: %v", err))
	}

	// Initialize workflow storage service
	workflowsDir := os.Getenv("WORKFLOWS_DIR")
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

	// Initialize all flows during startup
	service.initializeFlows()

	return service
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
			"user_id": userID,
			"service_catalog": serviceCatalog,
			"user_capabilities": g.buildUserCapabilities(serviceCatalog),
			"available_actions": g.extractAvailableActions(serviceCatalog),
			"examples": g.generateCapabilityExamples(serviceCatalog),
			"status": "ready",
			"oauth_validated": len(oauthTokens) > 0,
			"connected_services": g.getConnectedServiceNames(serviceCatalog),
		}
		
		return output, nil
	})

	// Intent Gatherer Flow
	g.intentGathererFlow = genkit.DefineFlow(g.genkit, "intent-gatherer", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		model := googlegenai.GoogleAIModel(g.genkit, "gemini-1.5-flash")
		if model == nil {
			return nil, fmt.Errorf("model gemini-1.5-flash not found")
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
		model := googlegenai.GoogleAIModel(g.genkit, "gemini-1.5-flash")
		if model == nil {
			return IntentAnalystOutput{}, fmt.Errorf("model gemini-1.5-flash not found")
		}
		
		// Extract user message and available services from simplified typed input
		userMessage := input.UserMessage
		availableServices := input.AvailableServices
		log.Printf("[DEBUG] Intent Analyst: user_message: %s", userMessage)
		log.Printf("[DEBUG] Intent Analyst: available services: %v", availableServices)
		
		// Template data is now directly used in the prompt formatting below
		
		// Load Genkit dotprompt with YAML front matter handling
		intentPrompt, err := g.loadPrompt("intent_analyst")
		if err != nil {
			return IntentAnalystOutput{}, fmt.Errorf("failed to load intent analyst prompt: %v", err)
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
			"user_input":         input.UserMessage,  // Fixed: match prompt schema field name
			"available_services": input.AvailableServices,  // Pass array directly to match schema
			"rac_context":        racContext, // Inject RaC agent specification as string
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
		model := googlegenai.GoogleAIModel(g.genkit, "gemini-1.5-flash")
		if model == nil {
			return WorkflowGeneratorOutput{}, fmt.Errorf("model gemini-1.5-flash not found")
		}
		
		// Load Genkit dotprompt with YAML front matter handling
		workflowPrompt, err := g.loadPrompt("workflow_generator")
		if err != nil {
			return WorkflowGeneratorOutput{}, fmt.Errorf("failed to load workflow generator prompt: %w", err)
		}
		
		// Get live MCP service catalog for input data
		mcpServices, err := g.mcpService.GetServiceCatalog()
		if err != nil {
			return WorkflowGeneratorOutput{}, fmt.Errorf("failed to get MCP service catalog: %w", err)
		}
		
		// Use unified parser to build available services section
		availableServices, err := g.mcpParser.BuildAvailableServicesSection(mcpServices)
		if err != nil {
			return WorkflowGeneratorOutput{}, fmt.Errorf("failed to build available services section: %w", err)
		}

		// Load RaC context from canonical RaC specification
		racContextPath := "rac/agents/workflow_generator.cue"
		racContextBytes, err := os.ReadFile(racContextPath)
		if err != nil {
			return WorkflowGeneratorOutput{}, fmt.Errorf("failed to load RaC context: %w", err)
		}
		racContext := string(racContextBytes)
		
		// Prepare input data for Genkit dotprompt execution
		inputData := map[string]interface{}{
			"user_input":         input.UserInput,
			"validated_intent":   input.ValidatedIntent,
			"available_services": availableServices,
			"rac_context":        racContext, // Inject actual RaC context as string
		}
		
		// Execute prompt with input data (Genkit handles templating)
		// Use the same pattern as Intent Analyst for consistency
		aiPrompt, ok := workflowPrompt.(*ai.Prompt)
		if !ok {
			return WorkflowGeneratorOutput{}, fmt.Errorf("loaded prompt is not *ai.Prompt type")
		}
		
		resp, err := aiPrompt.Execute(ctx, ai.WithInput(inputData))
		
		log.Printf("[GenkitService] Using Genkit dotprompt execution for workflow generator")
		
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

// ExecutePersonalCapabilitiesAgent executes the Personal Capabilities Agent
func (g *GenkitService) ExecutePersonalCapabilitiesAgent(input map[string]interface{}) (*types.AgentResponse, error) {
	// Execute the pre-defined flow (uses inline prompts for now)
	result, err := g.personalCapabilitiesFlow.Run(g.ctx, input)
	if err != nil {
		return &types.AgentResponse{
			AgentID: "personal_capabilities",
			Error:   err.Error(),
		}, nil
	}
	
	return &types.AgentResponse{
		AgentID: "personal_capabilities",
		Output:  result,
	}, nil
}

// ExecuteIntentGathererAgent executes the Intent Gatherer Agent
func (g *GenkitService) ExecuteIntentGathererAgent(input map[string]interface{}) (*types.AgentResponse, error) {
	// Execute the pre-defined flow (uses inline prompts for now)
	result, err := g.intentGathererFlow.Run(g.ctx, input)
	if err != nil {
		return &types.AgentResponse{
			AgentID: "intent_gatherer",
			Error:   err.Error(),
		}, nil
	}
	
	return &types.AgentResponse{
		AgentID: "intent_gatherer",
		Output:  result,
	}, nil
}

// ExecuteIntentAnalystAgent executes the Intent Analyst Agent
func (g *GenkitService) ExecuteIntentAnalystAgent(input map[string]interface{}) (*types.AgentResponse, error) {
	// Convert map[string]interface{} input to simplified IntentAnalystInput
	typedInput := IntentAnalystInput{}
	
	// Extract user message from workflow_intent (proper Go type handling)
	if workflowIntent, ok := input["workflow_intent"].(*types.WorkflowIntent); ok {
		typedInput.UserMessage = workflowIntent.UserMessage
	}
	
	// Extract available services from service_schemas (proper Go type handling)
	availableServices := []string{}
	if serviceSchemas, ok := input["service_schemas"].(map[string]types.ServiceSchema); ok {
		for serviceName := range serviceSchemas {
			availableServices = append(availableServices, serviceName)
		}
	}
	typedInput.AvailableServices = availableServices
	
	log.Printf("[DEBUG] ExecuteIntentAnalystAgent: Simplified input: %+v", typedInput)
	
	// Execute the pre-defined flow with typed input
	result, err := g.intentAnalystFlow.Run(g.ctx, typedInput)
	if err != nil {
		return &types.AgentResponse{
			AgentID: "intent_analyst",
			Error:   err.Error(),
		}, nil
	}
	
	// Convert typed output back to map[string]interface{} for compatibility
	outputMap := map[string]interface{}{
		"is_automation_request": result.IsAutomationRequest,
		"required_services":     result.RequiredServices,
		"can_fulfill":           result.CanFulfill,
		"missing_info":          result.MissingInfo,
		"next_action":           result.NextAction,
	}
	
	return &types.AgentResponse{
		AgentID: "intent_analyst",
		Output:  outputMap,
	}, nil
}

// ExecuteWorkflowGeneratorAgent executes the Workflow Generator Agent with JSON → CUE conversion
func (g *GenkitService) ExecuteWorkflowGeneratorAgent(input map[string]interface{}) (*types.AgentResponse, error) {
	log.Printf("[GenkitService] === EXECUTING WORKFLOW GENERATOR AGENT ===")
	log.Printf("[GenkitService] Input keys: %+v", getInputKeys(input))
	if userID, exists := input["user_id"]; exists {
		log.Printf("[GenkitService] User ID: %v", userID)
	}
	if validatedIntent, exists := input["validated_intent"]; exists {
		log.Printf("[GenkitService] Validated intent: %+v", validatedIntent)
	}
	
	// Convert map input to typed struct
	workflowInput := WorkflowGeneratorInput{
		UserInput:        getString(input, "user_input"),
		ValidatedIntent:  getMap(input, "validated_intent"),
		AvailableServices: getString(input, "available_services"),
	}
	
	// Execute the pre-defined flow to get JSON workflow
	result, err := g.workflowGeneratorFlow.Run(g.ctx, workflowInput)
	if err != nil {
		return &types.AgentResponse{
			AgentID: "workflow_generator",
			Error:   err.Error(),
		}, nil
	}
	
	log.Printf("[GenkitService] LLM flow completed successfully")
	log.Printf("[GenkitService] Workflow Generator result: %+v", result)
	
	// Extract workflow JSON from LLM result and convert to CUE (following RaC specification)
	var cueContent string
	
	log.Printf("[GenkitService] Extracting workflow JSON from LLM result (RaC-compliant)")
	log.Printf("[GenkitService] Result version: %s, name: %s", result.Version, result.Name)
	
	// RaC Specification: LLM should return structured JSON, system converts to CUE
	// Convert typed struct to map for JSON processing
	var workflowJSON map[string]interface{}
	
	// Method 1: Convert typed struct to JSON map
	if jsonBytes, err := json.Marshal(result); err == nil {
		if err := json.Unmarshal(jsonBytes, &workflowJSON); err == nil {
			// Validate that this looks like a workflow JSON (has required fields)
			if g.isValidWorkflowJSON(workflowJSON) {
				cueContent = g.convertJSONToCUE(workflowJSON)
				log.Printf("[GenkitService] Generated CUE from LLM JSON workflow (%d characters)", len(cueContent))
			}
		}
	}
	
	// No fallback methods - LLM must generate valid JSON workflow according to schema
	
	if cueContent == "" {
		return &types.AgentResponse{
			AgentID: "workflow_generator",
			Error:   "no valid workflow generated by LLM",
		}, nil
	}
	
	// Extract user ID from input
	userID := "authenticated_user" // Default fallback
	if uid, exists := input["user_id"]; exists {
		if uidStr, ok := uid.(string); ok {
			userID = uidStr
		}
	}
	
	// Extract workflow name from the generated CUE or generate one
	workflowName := g.extractWorkflowName(input, cueContent)
	
	// Save CUE file to disk with organized artifacts
	log.Printf("[GenkitService] Attempting to save workflow file for user %s, name: %s", userID, workflowName)
	log.Printf("[GenkitService] CUE content length: %d characters", len(cueContent))
	workflowFile, saveErr := g.workflowStorage.SaveWorkflow(userID, workflowName, cueContent)
	if saveErr != nil {
		// Log error but don't fail the response
		log.Printf("[GenkitService] ERROR: Failed to save workflow file: %v", saveErr)
	} else {
		log.Printf("[GenkitService] SUCCESS: Workflow file saved successfully")
		log.Printf("[GenkitService] File ID: %s, Filename: %s, Path: %s", workflowFile.ID, workflowFile.Filename, workflowFile.Path)
		
		// Extract workflow ID for artifact saving
		workflowID := strings.TrimPrefix(workflowFile.ID, userID+"_")
		
		// Convert typed struct to map for artifact saving
		resultMap := make(map[string]interface{})
		if jsonBytes, err := json.Marshal(result); err == nil {
			json.Unmarshal(jsonBytes, &resultMap)
		}
		
		// Save generation artifacts
		g.saveWorkflowArtifacts(userID, workflowID, input, resultMap)
		
		// Create final output map with workflow file information
		outputMap := resultMap
		outputMap["workflow_file"] = map[string]interface{}{
			"id":       workflowFile.ID,
			"filename": workflowFile.Filename,
			"path":     workflowFile.Path,
			"saved_at": workflowFile.CreatedAt,
		}
		outputMap["workflow_cue"] = cueContent
		outputMap["original_cue"] = cueContent
		
		return &types.AgentResponse{
			AgentID: "workflow_generator",
			Output:  outputMap,
		}, nil
	}
	
	// Convert typed struct to map for final output
	resultMap := make(map[string]interface{})
	if jsonBytes, err := json.Marshal(result); err == nil {
		json.Unmarshal(jsonBytes, &resultMap)
	}
	
	// Update result with the generated CUE content
	resultMap["workflow_cue"] = cueContent
	resultMap["original_cue"] = cueContent
	
	return &types.AgentResponse{
		AgentID: "workflow_generator",
		Output:  resultMap,
	}, nil
}

// extractWorkflowName extracts a meaningful workflow name from input or CUE content
func (g *GenkitService) extractWorkflowName(input map[string]interface{}, cueContent string) string {
	// Try to extract from input first
	if message, exists := input["message"]; exists {
		if msgStr, ok := message.(string); ok {
			// Clean and shorten the message for filename
			name := strings.ToLower(msgStr)
			// Replace spaces and special characters with underscores
			re := regexp.MustCompile(`[^a-z0-9]+`)
			name = re.ReplaceAllString(name, "_")
			// Limit length
			if len(name) > 30 {
				name = name[:30]
			}
			// Remove trailing underscores
			name = strings.Trim(name, "_")
			if name != "" {
				return name
			}
		}
	}
	
	// Try to extract from validated_intent
	if intent, exists := input["validated_intent"]; exists {
		if intentMap, ok := intent.(map[string]interface{}); ok {
			if pattern, exists := intentMap["workflow_pattern"]; exists {
				if patternStr, ok := pattern.(string); ok {
					name := strings.ToLower(patternStr)
					re := regexp.MustCompile(`[^a-z0-9]+`)
					name = re.ReplaceAllString(name, "_")
					if len(name) > 30 {
						name = name[:30]
					}
					name = strings.Trim(name, "_")
					if name != "" {
						return name
					}
				}
			}
		}
	}
	
	// Fallback to generic name
	return "workflow"
}

// saveWorkflowArtifacts saves all generation artifacts to the workflow folder
func (g *GenkitService) saveWorkflowArtifacts(userID string, workflowID string, input map[string]interface{}, result map[string]interface{}) {
	log.Printf("[GenkitService] Saving workflow artifacts for workflow %s", workflowID)
	
	// Save input prompt/request
	if inputJSON, err := json.MarshalIndent(input, "", "  "); err == nil {
		if saveErr := g.workflowStorage.SavePrompt(userID, workflowID, "user_input", string(inputJSON)); saveErr != nil {
			log.Printf("[GenkitService] WARNING: Failed to save user input: %v", saveErr)
		} else {
			log.Printf("[GenkitService] Saved user input prompt")
		}
	}
	
	// Save LLM response
	if resultJSON, err := json.MarshalIndent(result, "", "  "); err == nil {
		if saveErr := g.workflowStorage.SaveResponse(userID, workflowID, "llm_response", string(resultJSON)); saveErr != nil {
			log.Printf("[GenkitService] WARNING: Failed to save LLM response: %v", saveErr)
		} else {
			log.Printf("[GenkitService] Saved LLM response")
		}
	}
	
	// Save generation metadata
	metadata := map[string]interface{}{
		"generated_at": time.Now().Format(time.RFC3339),
		"agent_id": "workflow_generator",
		"model": "gemini-1.5-flash",
		"user_id": userID,
		"workflow_id": workflowID,
		"generation_source": "genkit_service",
	}
	
	if metadataJSON, err := json.MarshalIndent(metadata, "", "  "); err == nil {
		if saveErr := g.workflowStorage.SaveWorkflowArtifact(userID, workflowID, "metadata", "generation.json", string(metadataJSON)); saveErr != nil {
			log.Printf("[GenkitService] WARNING: Failed to save generation metadata: %v", saveErr)
		} else {
			log.Printf("[GenkitService] Saved generation metadata")
		}
	}
	
	log.Printf("[GenkitService] Workflow artifacts saving completed for %s", workflowID)
}

// isValidWorkflowJSON validates that the JSON contains required workflow fields
func (g *GenkitService) isValidWorkflowJSON(workflowJSON map[string]interface{}) bool {
	// Check for required fields according to workflow_json_schema.json
	requiredFields := []string{"version", "name", "description", "steps", "user_parameters", "services"}
	
	for _, field := range requiredFields {
		if _, exists := workflowJSON[field]; !exists {
			log.Printf("[GenkitService] Missing required field '%s' in workflow JSON", field)
			return false
		}
	}
	
	// Validate steps is an array
	if steps, ok := workflowJSON["steps"]; ok {
		if stepsArray, ok := steps.([]interface{}); ok {
			if len(stepsArray) == 0 {
				log.Printf("[GenkitService] Workflow JSON has empty steps array")
				return false
			}
		} else {
			log.Printf("[GenkitService] Workflow JSON steps field is not an array")
			return false
		}
	}
	
	log.Printf("[GenkitService] Workflow JSON validation passed")
	return true
}



// buildUserCapabilities creates structured user capabilities from service catalog (using unified parser)
func (g *GenkitService) buildUserCapabilities(serviceCatalog map[string]interface{}) map[string]interface{} {
	// Use unified MCPCatalogParser to eliminate duplicate logic
	serviceCapabilities, err := g.mcpParser.BuildServiceCapabilities(serviceCatalog, true)
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
	
	// Package declaration and imports
	cueBuilder.WriteString("package workflows\n\n")
	cueBuilder.WriteString("import \"../schemas.cue\"\n\n")
	
	// Main workflow definition
	cueBuilder.WriteString("workflow: #DeterministicWorkflow & {\n")
	cueBuilder.WriteString(fmt.Sprintf("\tversion: \"1.0.0\"\n"))
	cueBuilder.WriteString(fmt.Sprintf("\tname: %q\n", workflowName))
	cueBuilder.WriteString(fmt.Sprintf("\tdescription: %q\n", description))
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
	
	paramBuilder.WriteString("\t\t}")
	return paramBuilder.String()
}

// convertJSONServiceBindingsToCUE converts JSON service bindings to CUE format
func (g *GenkitService) convertJSONServiceBindingsToCUE(workflowJSON map[string]interface{}) string {
	var servicesBuilder strings.Builder
	servicesBuilder.WriteString("\tservices: {\n")
	
	if servicesData, exists := workflowJSON["service_bindings"]; exists {
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

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
