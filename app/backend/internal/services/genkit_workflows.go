package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"sohoaas-backend/internal/types"
)

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

	// Convert map input to typed struct with safe defaults
	userIntent := getString(input, "user_intent")
	validatedIntentMap := getMap(input, "validated_intent")
	availableServices := getString(input, "available_services")

	// Convert validated_intent map to typed struct
	validatedIntent := ValidatedIntent{
		IsAutomationRequest: getBool(validatedIntentMap, "is_automation_request"),
		RequiredServices:    getStringSlice(validatedIntentMap, "required_services"),
		CanFulfill:          getBool(validatedIntentMap, "can_fulfill"),
		MissingInfo:         getStringSlice(validatedIntentMap, "missing_info"),
		NextAction:          getStringFromMap(validatedIntentMap, "next_action"),
		Explanation:         getStringFromMap(validatedIntentMap, "explanation"),
		Confidence:          getFloat64(validatedIntentMap, "confidence"),
		WorkflowPattern:     getStringFromMap(validatedIntentMap, "workflow_pattern"),
	}

	// Validate required user input
	if userIntent == "" {
		return nil, fmt.Errorf("user_intent is required for workflow generation")
	}

	log.Printf("[GenkitService] User input: %s", userIntent)

	if len(validatedIntentMap) == 0 {
		validatedIntent = ValidatedIntent{
			IsAutomationRequest: false,
			RequiredServices:    []string{},
			CanFulfill:          false,
			MissingInfo:         []string{},
			NextAction:          "clarify_request",
		}
		log.Printf("[GenkitService] WARNING: Empty validated_intent, using default")
	}

	if availableServices == "" {
		availableServices = "No services available"
		log.Printf("[GenkitService] WARNING: Empty available_services, using default")
	}

	// Load focused RaC context from workflow-prompt.cue (streamlined for LLM)
	racBasePath := os.Getenv("RAC_CONTEXT_PATH")
	if racBasePath == "" {
		racBasePath = "rac" // Default fallback
	}
	log.Printf("[GenkitService] === RAC CONTEXT LOADING ===")
	log.Printf("[GenkitService] RAC_CONTEXT_PATH env: %s", os.Getenv("RAC_CONTEXT_PATH"))
	log.Printf("[GenkitService] Using RaC base path: %s", racBasePath)

	// Load workflow-prompt.cue (focused schema for LLM workflow generation)
	racContextPath := fmt.Sprintf("%s/agents/prompts/workflow-prompt.cue", racBasePath)
	log.Printf("[GenkitService] Loading focused workflow context from: %s", racContextPath)
	racContextBytes, err := os.ReadFile(racContextPath)
	if err != nil {
		log.Printf("[GenkitService] ERROR: Failed to load workflow prompt context from %s: %v", racContextPath, err)
		return nil, fmt.Errorf("failed to load RaC workflow prompt context from %s: %w", racContextPath, err)
	}
	log.Printf("[GenkitService] SUCCESS: Loaded focused workflow context (%d bytes)", len(racContextBytes))

	// Use focused workflow-prompt.cue directly (no combination needed)
	racContext := string(racContextBytes)
	log.Printf("[GenkitService] SUCCESS: Using focused RaC workflow context (%d total bytes)", len(racContext))

	workflowInput := WorkflowGeneratorInput{
		UserIntent:        userIntent,
		ValidatedIntent:   validatedIntent,
		AvailableServices: availableServices,
		RacContext:        racContext,
	}
	log.Printf("[GenkitService] === WORKFLOW INPUT PREPARED ===")
	log.Printf("[GenkitService] UserIntent length: %d chars", len(userIntent))
	log.Printf("[GenkitService] AvailableServices length: %d chars", len(availableServices))
	log.Printf("[GenkitService] RacContext length: %d chars", len(racContext))
	log.Printf("[GenkitService] ValidatedIntent services: %+v", validatedIntent.RequiredServices)

	// Additional validation for completely empty original input
	allEmpty := getString(input, "user_intent") == "" &&
		len(getMap(input, "validated_intent")) == 0 &&
		getString(input, "available_services") == ""

	if allEmpty {
		log.Printf("[GenkitService] WARNING: All inputs empty, returning minimal fallback workflow")
		// Return minimal valid workflow instead of calling LLM
		return &types.AgentResponse{
			AgentID: "workflow_generator",
			Output: map[string]interface{}{
				"version":         "1.0",
				"name":            "empty_input_workflow",
				"description":     "Generated from empty input - no user intent provided",
				"steps":           []types.WorkflowStep{},
				"user_parameters": make(map[string]types.UserParameter),
				"services":        make(map[string]interface{}),
				"workflow_cue":    "// Empty workflow - no user input provided\nworkflow: {}\n",
				"original_cue":    "// Empty workflow - no user input provided\nworkflow: {}\n",
			},
		}, nil
	}

	// Execute the pre-defined flow to get JSON workflow with error recovery
	log.Printf("[GenkitService] === EXECUTING WORKFLOW GENERATOR FLOW ===")
	log.Printf("[GenkitService] About to call workflowGeneratorFlow.Run() with RaC context")
	log.Printf("[GenkitService] Flow context: %+v", g.ctx != nil)
	result, err := g.workflowGeneratorFlow.Run(g.ctx, workflowInput)
	if err != nil {
		log.Printf("[GenkitService] Processing workflow generation for user input: %s", userIntent)
		// Check if this is a JSON parsing error from Genkit framework
		if strings.Contains(err.Error(), "not valid JSON") || strings.Contains(err.Error(), "unexpected end of JSON input") {
			log.Printf("[GenkitService] Detected JSON parsing error, returning structured error response")
			return &types.AgentResponse{
				AgentID: "workflow_generator",
				Output: map[string]interface{}{
					"version":         "1.0",
					"name":            "json_parse_error_workflow",
					"description":     "Workflow generation failed due to JSON parsing error",
					"steps":           []types.WorkflowStep{},
					"user_parameters": make(map[string]types.UserParameter),
					"services":        make(map[string]interface{}),
					"workflow_cue":    "// JSON parsing error occurred\nworkflow: { error: \"JSON parsing failed\" }\n",
					"original_cue":    "// JSON parsing error occurred\nworkflow: { error: \"JSON parsing failed\" }\n",
					"error_details":   err.Error(),
				},
			}, nil
		}
		return &types.AgentResponse{
			AgentID: "workflow_generator",
			Error:   err.Error(),
		}, nil
	}

	log.Printf("[=== GenkitService] LLM flow completed successfully")
	log.Printf("[GenkitService] Workflow Generator result: %+v", result)

	// Extract workflow JSON from LLM result and convert to CUE (following RaC specification)
	var cueContent string

	log.Printf("[GenkitService] Extracting workflow JSON from LLM result (RaC-compliant)")
	log.Printf("[GenkitService] Result version: %s, name: %s", result.Version, result.Name)

	// RaC Specification: LLM should return structured JSON, system converts to CUE
	// Convert typed struct to map for JSON processing
	var workflowJSON map[string]interface{}

	// Method 1: Convert typed struct to JSON map with markdown cleaning
	if jsonBytes, err := json.Marshal(result); err == nil {
		// Clean any markdown formatting from the JSON before parsing
		cleanedJSON := cleanMarkdownFromJSON(string(jsonBytes))
		log.Printf("[GenkitService] Cleaned JSON length: %d characters", len(cleanedJSON))

		if err := json.Unmarshal([]byte(cleanedJSON), &workflowJSON); err == nil {
			// Validate that this looks like a workflow JSON (has required fields)
			if g.isValidWorkflowJSON(workflowJSON) {
				cueContent = g.convertJSONToCUE(workflowJSON)
				// Sanitize CUE content to remove illegal characters
				cueContent = g.sanitizeCUEContent(cueContent)
				log.Printf("[GenkitService] Generated CUE from LLM JSON workflow (%d characters)", len(cueContent))
			} else {
				log.Printf("[GenkitService] WARNING: Generated JSON does not look like valid workflow")
			}
		} else {
			log.Printf("[GenkitService] WARNING: Failed to unmarshal cleaned JSON: %v", err)
			log.Printf("[GenkitService] Cleaned JSON content: %s", cleanedJSON)
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

		// Save essential LLM artifacts using WorkflowStorageService (no duplication)
		workflowID := strings.TrimPrefix(workflowFile.ID, userID+"_")
		log.Printf("[GenkitService] Saving LLM artifacts for workflow %s", workflowID)

		// Save original LLM input (prompt) with RaC context
		if inputJSON, err := json.MarshalIndent(workflowInput, "", "  "); err == nil {
			if saveErr := g.workflowStorage.SavePrompt(userID, workflowID, "user_intent", string(inputJSON)); saveErr != nil {
				log.Printf("[GenkitService] ERROR: Failed to save user input: %v", saveErr)
			} else {
				log.Printf("[GenkitService] SUCCESS: Saved user input prompt")
			}
		}

		// Save LLM response (raw output)
		if resultJSON, err := json.MarshalIndent(result, "", "  "); err == nil {
			if saveErr := g.workflowStorage.SaveResponse(userID, workflowID, "llm_response", string(resultJSON)); saveErr != nil {
				log.Printf("[GenkitService] ERROR: Failed to save LLM response: %v", saveErr)
			} else {
				log.Printf("[GenkitService] SUCCESS: Saved LLM response")
			}
		}

		// Save original JSON workflow (convert struct to JSON)
		if jsonContent, err := json.MarshalIndent(result, "", "  "); err == nil {
			if saveErr := g.workflowStorage.SaveWorkflowArtifact(userID, workflowID, ".", "workflow.json", string(jsonContent)); saveErr != nil {
				log.Printf("[GenkitService] ERROR: Failed to save workflow.json: %v", saveErr)
			} else {
				log.Printf("[GenkitService] SUCCESS: Saved workflow.json")
			}
		}

		// Convert typed struct to map for output
		outputMap := make(map[string]interface{})
		if jsonBytes, err := json.Marshal(result); err == nil {
			json.Unmarshal(jsonBytes, &outputMap)
		}
		
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

// isValidWorkflowJSON validates that the workflow JSON has required fields
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
