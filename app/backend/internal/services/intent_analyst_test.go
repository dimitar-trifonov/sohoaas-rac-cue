package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"sohoaas-backend/internal/types"
)

// TestIntentAnalystWithRealLLM tests the Intent Analyst agent with actual LLM calls
func TestIntentAnalystWithRealLLM(t *testing.T) {
	// Skip if no API key available
	if os.Getenv("GOOGLE_API_KEY") == "" {
		t.Skip("Skipping Intent Analyst LLM integration test - GOOGLE_API_KEY not set")
	}

	testCases := []struct {
		name           string
		userMessage    string
		expectedIntent bool
		expectedServices []string
		expectedCanFulfill bool
		expectedAction string
	}{
		{
			name:           "Gmail Automation Request",
			userMessage:    "Send an email to john@example.com about the project update",
			expectedIntent: true,
			expectedServices: []string{"gmail"},
			expectedCanFulfill: true,
			expectedAction: "generate_workflow",
		},
		{
			name:           "Complex Multi-Service Request",
			userMessage:    "When I click Run, take the oldest Gmail message from bojidar@investclub.bg, create a Google Doc from it in a Drive folder Email-Automation/Bojidar/{{YYYY‑MM‑DD}} if it does not exist yet.",
			expectedIntent: true,
			expectedServices: []string{"gmail", "docs", "drive"},
			expectedCanFulfill: true,
			expectedAction: "generate_workflow",
		},
		{
			name:           "Non-Automation Request",
			userMessage:    "What's the weather like today?",
			expectedIntent: false,
			expectedServices: []string{},
			expectedCanFulfill: false,
			expectedAction: "unsupported_request",
		},
		{
			name:           "Incomplete Request",
			userMessage:    "Send an email to someone",
			expectedIntent: true,
			expectedServices: []string{"gmail"},
			expectedCanFulfill: false,
			expectedAction: "need_clarification",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("=== TESTING INTENT ANALYST: %s ===", tc.name)
			t.Logf("User Message: %s", tc.userMessage)

			// Initialize real Genkit service
			genkitService, err := initializeGenkitServiceForTest(t)
			if err != nil {
				t.Fatalf("Failed to initialize Genkit service: %v", err)
			}

			// Prepare input for Intent Analyst
			workflowIntent := &types.WorkflowIntent{
				UserMessage: tc.userMessage,
			}

			// Mock service schemas (Google Workspace services)
			serviceSchemas := map[string]types.ServiceSchema{
				"gmail": {
					ServiceName: "gmail",
					Status:      "connected",
					Actions: map[string]types.ActionSchema{
						"send_email": {
							ActionName:  "send_email",
							Description: "Send an email message",
							RequiredFields: []types.FieldSchema{
								{FieldName: "to", FieldType: "string", Description: "Recipient email address"},
								{FieldName: "subject", FieldType: "string", Description: "Email subject"},
								{FieldName: "body", FieldType: "string", Description: "Email body content"},
							},
						},
					},
				},
				"docs": {
					ServiceName: "docs",
					Status:      "connected",
					Actions: map[string]types.ActionSchema{
						"create_document": {
							ActionName:  "create_document",
							Description: "Create a new Google Doc",
							RequiredFields: []types.FieldSchema{
								{FieldName: "title", FieldType: "string", Description: "Document title"},
							},
						},
					},
				},
				"drive": {
					ServiceName: "drive",
					Status:      "connected",
					Actions: map[string]types.ActionSchema{
						"create_folder": {
							ActionName:  "create_folder",
							Description: "Create a new folder in Drive",
							RequiredFields: []types.FieldSchema{
								{FieldName: "name", FieldType: "string", Description: "Folder name"},
							},
						},
					},
				},
				"calendar": {
					ServiceName: "calendar",
					Status:      "connected",
					Actions: map[string]types.ActionSchema{
						"create_event": {
							ActionName:  "create_event",
							Description: "Create a calendar event",
							RequiredFields: []types.FieldSchema{
								{FieldName: "title", FieldType: "string", Description: "Event title"},
							},
						},
					},
				},
			}

			input := map[string]interface{}{
				"workflow_intent": workflowIntent,
				"service_schemas": serviceSchemas,
			}

			t.Logf("=== MAKING REAL LLM CALL TO INTENT ANALYST ===")

			// Execute Intent Analyst with real LLM
			response, err := genkitService.ExecuteIntentAnalystAgent(input)
			if err != nil {
				t.Logf("Intent Analyst LLM call failed: %v", err)
				if response != nil && response.Output != nil {
					t.Logf("LLM raw output: %+v", response.Output)
				}
				t.Fatalf("Intent Analyst returned error: %v", err)
			}

			if response.Error != "" {
				t.Fatalf("Intent Analyst returned error: %s", response.Error)
			}

			// Validate response structure
			validateIntentAnalystResponse(t, response.Output)

			// Extract and validate specific fields
			output := response.Output
			isAutomationRequest, _ := output["is_automation_request"].(bool)
			requiredServices, _ := output["required_services"].([]interface{})
			canFulfill, _ := output["can_fulfill"].(bool)
			nextAction, _ := output["next_action"].(string)

			// Convert required services to string slice
			var serviceStrings []string
			for _, service := range requiredServices {
				if s, ok := service.(string); ok {
					serviceStrings = append(serviceStrings, s)
				}
			}

			// Log results
			t.Logf("=== INTENT ANALYST RESULTS ===")
			t.Logf("Is Automation Request: %t", isAutomationRequest)
			t.Logf("Required Services: %v", serviceStrings)
			t.Logf("Can Fulfill: %t", canFulfill)
			t.Logf("Next Action: %s", nextAction)

			// Validate against expectations
			if isAutomationRequest != tc.expectedIntent {
				t.Errorf("Expected is_automation_request=%t, got %t", tc.expectedIntent, isAutomationRequest)
			}

			if canFulfill != tc.expectedCanFulfill {
				t.Errorf("Expected can_fulfill=%t, got %t", tc.expectedCanFulfill, canFulfill)
			}

			if nextAction != tc.expectedAction {
				t.Errorf("Expected next_action=%s, got %s", tc.expectedAction, nextAction)
			}

			// Validate required services (if automation request)
			if tc.expectedIntent && len(tc.expectedServices) > 0 {
				if !containsAllServices(serviceStrings, tc.expectedServices) {
					t.Errorf("Expected services %v not all found in %v", tc.expectedServices, serviceStrings)
				}
			}

			// Generate test files for debugging
			err = generateIntentAnalystTestFiles(t, tc.name, input, response.Output)
			if err != nil {
				t.Logf("Warning: Failed to generate test files: %v", err)
			}
		})
	}
}

// TestIntentAnalystPromptLoading tests that the Intent Analyst prompt loads correctly
func TestIntentAnalystPromptLoading(t *testing.T) {
	// Skip if no API key available
	if os.Getenv("GOOGLE_API_KEY") == "" {
		t.Skip("Skipping Intent Analyst prompt loading test - GOOGLE_API_KEY not set")
	}

	t.Logf("=== TESTING INTENT ANALYST PROMPT LOADING ===")

	// Initialize Genkit service
	genkitService, err := initializeGenkitServiceForTest(t)
	if err != nil {
		t.Fatalf("Failed to initialize Genkit service: %v", err)
	}

	// Test prompt loading directly
	prompt, err := genkitService.loadPrompt("intent_analyst")
	if err != nil {
		t.Fatalf("Failed to load intent_analyst prompt: %v", err)
	}

	if prompt == nil {
		t.Fatalf("Loaded prompt is nil")
	}

	t.Logf("✅ Intent Analyst prompt loaded successfully")
	t.Logf("Prompt type: %T", prompt)
}

// validateIntentAnalystResponse validates the structure of Intent Analyst response
func validateIntentAnalystResponse(t *testing.T, output map[string]interface{}) {
	requiredFields := []string{
		"is_automation_request",
		"required_services", 
		"can_fulfill",
		"missing_info",
		"next_action",
	}

	for _, field := range requiredFields {
		if _, exists := output[field]; !exists {
			t.Errorf("Missing required field in Intent Analyst response: %s", field)
		}
	}

	// Validate field types
	if _, ok := output["is_automation_request"].(bool); !ok {
		t.Errorf("is_automation_request should be boolean")
	}

	if _, ok := output["can_fulfill"].(bool); !ok {
		t.Errorf("can_fulfill should be boolean")
	}

	if _, ok := output["next_action"].(string); !ok {
		t.Errorf("next_action should be string")
	}

	// Validate next_action enum values
	nextAction, _ := output["next_action"].(string)
	validActions := []string{"generate_workflow", "need_clarification", "unsupported_request"}
	if !containsStringInSlice(nextAction, validActions) {
		t.Errorf("next_action '%s' is not valid. Must be one of: %v", nextAction, validActions)
	}
}

// containsAllServices checks if all expected services are present
func containsAllServices(actual, expected []string) bool {
	for _, expectedService := range expected {
		if !containsStringInSlice(expectedService, actual) {
			return false
		}
	}
	return true
}

// containsStringInSlice checks if a slice contains a string
func containsStringInSlice(item string, slice []string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// generateIntentAnalystTestFiles creates test files for debugging
func generateIntentAnalystTestFiles(t *testing.T, testName string, input map[string]interface{}, output map[string]interface{}) error {
	// Create test directory
	testDir := filepath.Join("test_output", "intent_analyst", fmt.Sprintf("%s_%d", testName, time.Now().Unix()))
	if err := os.MkdirAll(testDir, 0755); err != nil {
		return fmt.Errorf("failed to create test directory: %v", err)
	}

	// Save input
	inputJSON, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal input: %v", err)
	}
	inputFile := filepath.Join(testDir, "input.json")
	if err := os.WriteFile(inputFile, inputJSON, 0644); err != nil {
		return fmt.Errorf("failed to write input file: %v", err)
	}

	// Save output
	outputJSON, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal output: %v", err)
	}
	outputFile := filepath.Join(testDir, "output.json")
	if err := os.WriteFile(outputFile, outputJSON, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %v", err)
	}

	// Generate test report
	report := fmt.Sprintf(`# Intent Analyst Test Report: %s

**Generated:** %s

## Test Input
- **User Message:** %s
- **Available Services:** %v

## Test Output
- **Is Automation Request:** %v
- **Required Services:** %v
- **Can Fulfill:** %v
- **Missing Info:** %v
- **Next Action:** %v

## Files Generated
- ✅ input.json - Complete test input
- ✅ output.json - Complete LLM response
- ✅ test_report.md - This report

## Usage
These files can be used to:
1. Debug Intent Analyst issues
2. Validate LLM response format
3. Test prompt modifications
4. Analyze intent classification patterns

`, testName, time.Now().Format("2006-01-02 15:04:05"),
		getStringFromMap(input, "workflow_intent.UserMessage"),
		getServiceNames(input),
		output["is_automation_request"],
		output["required_services"],
		output["can_fulfill"],
		output["missing_info"],
		output["next_action"])

	reportFile := filepath.Join(testDir, "test_report.md")
	if err := os.WriteFile(reportFile, []byte(report), 0644); err != nil {
		return fmt.Errorf("failed to write test report: %v", err)
	}

	t.Logf("✅ Intent Analyst test files generated: %s", testDir)
	return nil
}

// getStringFromMap safely extracts a string from nested map
func getStringFromMap(data map[string]interface{}, path string) string {
	if workflowIntent, ok := data["workflow_intent"].(*types.WorkflowIntent); ok {
		return workflowIntent.UserMessage
	}
	return "unknown"
}

// getServiceNames extracts service names from service schemas
func getServiceNames(data map[string]interface{}) []string {
	var services []string
	if serviceSchemas, ok := data["service_schemas"].(map[string]types.ServiceSchema); ok {
		for name := range serviceSchemas {
			services = append(services, name)
		}
	}
	return services
}
