package services

import (
	"testing"
	"sohoaas-backend/internal/types"
)

// TestExecuteWorkflowCore tests the main workflow orchestration method
func TestExecuteWorkflowCore(t *testing.T) {
	// Start mock MCP server
	mockServer := NewMockMCPServer(t)
	defer mockServer.Close()
	mockServer.SetDefaultGoogleWorkspaceResponses()

	// Create execution engine
	mcpService := NewMCPService(mockServer.URL())
	executionEngine := NewExecutionEngine(mcpService)

	// Create test user
	testUser := &types.User{
		ID:    "test_user_123",
		Email: "test@example.com",
	}

	// Simple test CUE workflow
	testWorkflowCUE := `
workflow: {
	name: "test_workflow"
	description: "Test workflow for core execution"
	steps: [
		{
			id: "step1"
			action: "gmail.send_message"
			parameters: {
				to: "${user.recipient_email}"
				subject: "Test Email"
				body: "This is a test email"
			}
			depends_on: []
		}
	]
	user_parameters: {
		recipient_email: {
			type: "string"
			prompt: "Enter recipient email:"
			required: true
		}
	}
}
`

	// Test parameters
	intentAnalysis := map[string]interface{}{
		"recipient_email": "recipient@example.com",
	}

	// Execute workflow preparation
	result, err := executionEngine.PrepareExecution(
		testWorkflowCUE,
		testUser.ID,
		testUser,
		intentAnalysis,
		"mock_oauth_token_valid",
		"America/New_York",
	)
	if err != nil {
		t.Fatalf("PrepareExecution failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected execution result, got nil")
	}

	if len(result.ResolvedSteps) != 1 {
		t.Errorf("Expected 1 step, got %d", len(result.ResolvedSteps))
	}

	// Verify step was resolved correctly
	step := result.ResolvedSteps[0]
	if step.ID != "step1" {
		t.Errorf("Expected step ID 'step1', got '%s'", step.ID)
	}

	if step.Service != "gmail" {
		t.Errorf("Expected service 'gmail', got '%s'", step.Service)
	}

	if step.Action != "send_message" {
		t.Errorf("Expected action 'send_message', got '%s'", step.Action)
	}

	// Test workflow execution
	err = executionEngine.ExecuteWorkflow(result)
	if err != nil {
		t.Fatalf("ExecuteWorkflow failed: %v", err)
	}

	// Verify step completed successfully (check the updated step from result)
	updatedStep := result.ResolvedSteps[0]
	if updatedStep.Status != "completed" {
		t.Errorf("Expected step status 'completed', got '%s'", updatedStep.Status)
	}

	t.Logf("✅ Core workflow execution test completed successfully")
}

// TestExecuteStepErrorHandling tests error scenarios during step execution
func TestExecuteStepErrorHandling(t *testing.T) {
	// Start mock MCP server
	mockServer := NewMockMCPServer(t)
	defer mockServer.Close()

	// Configure server to return error response
	errorResponse := &ExecuteActionResponse{
		Success: false,
		Error:   "QUOTA_EXCEEDED: Daily sending quota exceeded",
		Data:    nil,
	}
	mockServer.SetResponse("gmail", "send_message", errorResponse)

	// Create execution engine
	mcpService := NewMCPService(mockServer.URL())
	executionEngine := NewExecutionEngine(mcpService)

	// Create test step
	step := &ResolvedStep{
		ID:      "error_step",
		Service: "gmail",
		Action:  "send_message",
		Inputs: map[string]interface{}{
			"to":      "test@example.com",
			"subject": "Test",
			"body":    "Test body",
		},
		DependsOn: []string{},
	}

	// Create parameter context
	context := &ParameterContext{
		UserParameters: map[string]interface{}{},
		StepOutputs:    map[string]interface{}{},
		SystemParameters: map[string]interface{}{
			"user_id":       "test_user",
			"user_timezone": "America/New_York",
		},
	}

	// Execute step - should handle error gracefully
	err := executionEngine.executeStep(step, context)
	if err == nil {
		t.Fatal("Expected error from executeStep, got nil")
	}

	// Verify error message contains expected information
	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}

	t.Logf("Received expected error: %v", err)
}

// TestDependencyValidation tests step dependency validation
func TestDependencyValidation(t *testing.T) {
	// Create execution engine (no mock server needed for dependency validation)
	mcpService := NewMCPService("http://localhost:8080") // dummy URL
	executionEngine := NewExecutionEngine(mcpService)

	tests := []struct {
		name         string
		dependencies []string
		steps        []ResolvedStep
		expectReady  bool
	}{
		{
			name:         "No dependencies - ready",
			dependencies: []string{},
			steps:        []ResolvedStep{},
			expectReady:  true,
		},
		{
			name:         "Dependencies met - ready",
			dependencies: []string{"step1"},
			steps: []ResolvedStep{
				{ID: "step1", Status: "completed"},
			},
			expectReady: true,
		},
		{
			name:         "Dependencies not met - not ready",
			dependencies: []string{"step1"},
			steps: []ResolvedStep{
				{ID: "step1", Status: "pending"},
			},
			expectReady: false,
		},
		{
			name:         "Missing dependency step - not ready",
			dependencies: []string{"missing_step"},
			steps: []ResolvedStep{
				{ID: "step1", Status: "completed"},
			},
			expectReady: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use existing areDependenciesMet method
			ready := executionEngine.areDependenciesMet(tt.dependencies, tt.steps)

			if ready != tt.expectReady {
				t.Errorf("Expected ready=%t, got ready=%t", tt.expectReady, ready)
			}
		})
	}
}
