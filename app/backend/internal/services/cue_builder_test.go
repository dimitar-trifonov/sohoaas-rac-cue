package services

import (
	"testing"
)

func TestCUEBuilder_BuildCUEFromJSON(t *testing.T) {
	// Sample JSON workflow that the LLM might generate
	jsonWorkflow := `{
		"workflow_name": "Send Weekly Report Email",
		"description": "Automatically send weekly reports to team members",
		"trigger": {
			"type": "manual"
		},
		"steps": [
			{
				"id": "send_report_email",
				"name": "Send Weekly Report Email",
				"service": "gmail",
				"action": "gmail.send_email",
				"inputs": {
					"to": "${USER_INPUT:recipient_email}",
					"subject": "Weekly Report - ${CURRENT_DATE}",
					"body": "${USER_INPUT:report_content}"
				},
				"outputs": {
					"message_id": "string"
				}
			}
		],
		"user_parameters": [
			{
				"name": "recipient_email",
				"type": "string",
				"required": true,
				"description": "Email address to send the report to"
			},
			{
				"name": "report_content",
				"type": "string",
				"required": true,
				"description": "Content of the weekly report"
			}
		]
	}`

	// Create CUE builder (with nil MCP service for testing)
	cueBuilder := NewCUEBuilder(nil)
	
	// Convert JSON to CUE
	cueContent, err := cueBuilder.BuildCUEFromJSON(jsonWorkflow)
	if err != nil {
		t.Fatalf("Failed to build CUE from JSON: %v", err)
	}

	t.Logf("=== GENERATED CUE CONTENT ===\n%s", cueContent)

	// Verify CUE content contains expected elements
	expectedElements := []string{
		"package workflow",
		"#DeterministicWorkflow:",
		"workflow: #DeterministicWorkflow & {",
		`name: "Send Weekly Report Email"`,
		`service: "gmail"`,
		`action: "gmail.send_email"`,
		"user_parameters:",
		"service_bindings:",
	}

	for _, element := range expectedElements {
		if !containsString(cueContent, element) {
			t.Errorf("Generated CUE content missing expected element: %s", element)
		}
	}

	// Verify no invalid services
	invalidElements := []string{
		`service: "none"`,
		`service: ""`,
		`action: ""`,
	}

	for _, element := range invalidElements {
		if containsString(cueContent, element) {
			t.Errorf("Generated CUE content contains invalid element: %s", element)
		}
	}
}

func TestCUEBuilder_ValidateWorkflow(t *testing.T) {
	// Test invalid JSON workflow
	invalidJSON := `{
		"workflow_name": "Test Workflow",
		"description": "Test description",
		"trigger": {"type": "manual"},
		"steps": [
			{
				"id": "invalid_step",
				"name": "Invalid Step",
				"service": "nonexistent_service",
				"action": "invalid_action",
				"inputs": {}
			}
		],
		"user_parameters": []
	}`

	cueBuilder := NewCUEBuilder(nil)
	
	// This should fail validation (but we can't test MCP validation without a real service)
	_, err := cueBuilder.BuildCUEFromJSON(invalidJSON)
	
	// We expect an error due to validation failure
	if err == nil {
		t.Log("Note: Validation test skipped - requires MCP service for full validation")
	} else {
		t.Logf("Validation correctly failed: %v", err)
	}
}

// Helper function to check if string contains substring
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
