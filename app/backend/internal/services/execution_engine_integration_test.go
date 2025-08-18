package services

import (
	"testing"

	"sohoaas-backend/internal/types"
)

// TestExecutionEngineIntegrationWithMockMCP tests the complete execution pipeline with HTTP Mock MCP server
func TestExecutionEngineIntegrationWithMockMCP(t *testing.T) {
	t.Logf("=== EXECUTION ENGINE INTEGRATION TEST WITH MOCK MCP ===")
	
	// Start mock MCP server
	mockServer := NewMockMCPServer(t)
	defer mockServer.Close()
	
	// Configure default Google Workspace responses
	mockServer.SetDefaultGoogleWorkspaceResponses()
	
	// Create execution engine with mock MCP service
	mcpService := NewMCPService(mockServer.URL())
	executionEngine := NewExecutionEngine(mcpService)
	
	// Create test user
	testUser := &types.User{
		ID:    "test_user_123",
		Email: "test@example.com",
		Name:  "Test User",
	}
	
	// Create test workflow CUE content
	testWorkflowCUE := `
workflow: {
	name: "Daily Email Report"
	description: "Send daily email report with Gmail messages"
	
	steps: [
		{
			id: "fetch_emails"
			name: "Fetch recent emails"
			service: "gmail"
			action: "list_messages"
			inputs: {
				max_results: 5
				query: "is:unread"
			}
			outputs: {
				messages: "RUNTIME"
			}
		},
		{
			id: "create_report_doc"
			name: "Create report document"
			service: "docs"
			action: "create_document"
			inputs: {
				title: "Daily Email Report - ${system.current_date}"
			}
			outputs: {
				document_id: "RUNTIME"
				document_url: "RUNTIME"
			}
		},
		{
			id: "send_summary"
			name: "Send email summary"
			service: "gmail"
			action: "send_message"
			inputs: {
				to: "${user.email}"
				subject: "Daily Email Report"
				body: "Report created: ${steps.create_report_doc.outputs.document_url}"
			}
			outputs: {
				message_id: "RUNTIME"
			}
			depends_on: ["fetch_emails", "create_report_doc"]
		}
	]
	
	user_parameters: [
		{
			name: "report_frequency"
			type: "string"
			required: true
			default: "daily"
			prompt: "How often should the report be generated?"
		}
	]
	
	service_bindings: {
		gmail: {
			type: "mcp_service"
			provider: "workspace"
			auth: {
				type: "oauth2"
				scopes: ["https://www.googleapis.com/auth/gmail.readonly", "https://www.googleapis.com/auth/gmail.send"]
			}
		}
		docs: {
			type: "mcp_service"
			provider: "workspace"
			auth: {
				type: "oauth2"
				scopes: ["https://www.googleapis.com/auth/documents"]
			}
		}
	}
	
	execution_config: {
		mode: "sequential"
		timeout: "5m"
		environment: "test"
	}
}`
	
	// Create intent analysis (simulating Intent Analyst output)
	intentAnalysis := map[string]interface{}{
		"is_automation_request": true,
		"required_services":     []string{"gmail", "docs"},
		"can_fulfill":          true,
		"missing_info":         []string{},
		"next_action":          "generate_workflow",
		"user_parameters": map[string]interface{}{
			"report_frequency": "daily",
			"email": testUser.Email,
		},
	}
	
	// Test OAuth token (mock)
	testOAuthToken := "mock_oauth_token_valid"
	
	t.Logf("Phase 1: Testing Execution Preparation")
	
	// Test execution preparation
	executionPlan, err := executionEngine.PrepareExecution(
		testWorkflowCUE,
		testUser.ID,
		testUser,
		intentAnalysis,
		testOAuthToken,
		"America/New_York",
	)
	
	if err != nil {
		t.Fatalf("Execution preparation failed: %v", err)
	}
	
	// Validate execution plan
	if executionPlan == nil {
		t.Fatal("Execution plan is nil")
	}
	
	if len(executionPlan.ValidationErrors) > 0 {
		t.Fatalf("Execution plan has validation errors: %v", executionPlan.ValidationErrors)
	}
	
	if len(executionPlan.ResolvedSteps) != 3 {
		t.Fatalf("Expected 3 resolved steps, got %d", len(executionPlan.ResolvedSteps))
	}
	
	t.Logf("✅ Execution preparation successful: %d steps resolved", len(executionPlan.ResolvedSteps))
	
	// Validate parameter resolution
	if executionPlan.ParameterContext.SystemParameters["oauth_token"] != testOAuthToken {
		t.Error("OAuth token not properly set in system parameters")
	}
	
	if executionPlan.ParameterContext.SystemParameters["user_email"] != testUser.Email {
		t.Error("User email not properly set in system parameters")
	}
	
	t.Logf("Phase 2: Testing Workflow Execution")
	
	// Test workflow execution
	err = executionEngine.ExecuteWorkflow(executionPlan)
	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}
	
	t.Logf("✅ Workflow execution successful")
	
	// Validate step execution results
	for i, step := range executionPlan.ResolvedSteps {
		if step.Status != "completed" {
			t.Errorf("Step %d (%s) not completed, status: %s", i, step.ID, step.Status)
		}
		
		// Validate step outputs were populated
		if len(step.Outputs) == 0 {
			t.Errorf("Step %d (%s) has no outputs", i, step.ID)
		}
		
		t.Logf("Step %s: %s ✅", step.ID, step.Status)
	}
	
	// Validate step output propagation
	createDocStep := executionPlan.ResolvedSteps[1] // create_report_doc
	sendEmailStep := executionPlan.ResolvedSteps[2] // send_summary
	
	if createDocStep.Outputs["document_url"] == nil {
		t.Error("Document URL not set in create_report_doc step outputs")
	}
	
	if sendEmailStep.Outputs["message_id"] == nil {
		t.Error("Message ID not set in send_summary step outputs")
	}
	
	t.Logf("Phase 3: Testing Error Scenarios")
	
	// Test invalid OAuth token
	invalidTokenPlan, err := executionEngine.PrepareExecution(
		testWorkflowCUE,
		testUser.ID,
		testUser,
		intentAnalysis,
		"invalid_token",
		"America/New_York",
	)
	
	if err != nil {
		t.Fatalf("Preparation with invalid token failed: %v", err)
	}
	
	// This should fail during execution
	err = executionEngine.ExecuteWorkflow(invalidTokenPlan)
	if err == nil {
		t.Error("Expected execution to fail with invalid OAuth token, but it succeeded")
	} else {
		t.Logf("✅ Invalid OAuth token properly rejected: %v", err)
	}
	
	t.Logf("=== INTEGRATION TEST COMPLETED SUCCESSFULLY ===")
}

// TestExecutionEngineServiceValidation tests service validation against mock MCP catalog
func TestExecutionEngineServiceValidation(t *testing.T) {
	t.Logf("=== TESTING SERVICE VALIDATION WITH MOCK MCP ===")
	
	// Start mock MCP server
	mockServer := NewMockMCPServer(t)
	defer mockServer.Close()
	
	// Create execution engine
	mcpService := NewMCPService(mockServer.URL())
	executionEngine := NewExecutionEngine(mcpService)
	
	// Test valid workflow
	validWorkflow := &ParsedWorkflow{
		Name:        "Valid Workflow",
		Description: "Uses valid services",
		Steps: []WorkflowStep{
			{
				ID:      "step1",
				Service: "gmail",
				Action:  "send_message",
			},
			{
				ID:      "step2",
				Service: "docs",
				Action:  "create_document",
			},
		},
	}
	
	err := executionEngine.ValidateWorkflowServices(validWorkflow)
	if err != nil {
		t.Errorf("Valid workflow validation failed: %v", err)
	} else {
		t.Logf("✅ Valid workflow passed validation")
	}
	
	// Test invalid workflow
	invalidWorkflow := &ParsedWorkflow{
		Name:        "Invalid Workflow",
		Description: "Uses invalid services",
		Steps: []WorkflowStep{
			{
				ID:      "step1",
				Service: "nonexistent_service",
				Action:  "some_action",
			},
		},
	}
	
	err = executionEngine.ValidateWorkflowServices(invalidWorkflow)
	if err == nil {
		t.Error("Expected invalid workflow to fail validation, but it passed")
	} else {
		t.Logf("✅ Invalid workflow properly rejected: %v", err)
	}
	
	t.Logf("=== SERVICE VALIDATION TEST COMPLETED ===")
}

// TestMockMCPServerResponses tests the mock server responses directly
func TestMockMCPServerResponses(t *testing.T) {
	t.Logf("=== TESTING MOCK MCP SERVER RESPONSES ===")
	
	// Start mock server
	mockServer := NewMockMCPServer(t)
	defer mockServer.Close()
	
	mockServer.SetDefaultGoogleWorkspaceResponses()
	
	// Create MCP service client
	mcpService := NewMCPService(mockServer.URL())
	
	// Test service catalog
	catalog, err := mcpService.GetServiceCatalog()
	if err != nil {
		t.Fatalf("Failed to get service catalog: %v", err)
	}
	
	if catalog == nil {
		t.Fatal("Service catalog is nil")
	}
	
	t.Logf("✅ Service catalog retrieved successfully")
	
	// Test Gmail action
	response, err := mcpService.ExecuteAction("gmail", "send_message", map[string]interface{}{
		"to":      "test@example.com",
		"subject": "Test Email",
		"body":    "Test message",
	}, "mock_oauth_token_valid")
	
	if err != nil {
		t.Fatalf("Gmail action failed: %v", err)
	}
	
	if !response.Success {
		t.Fatalf("Gmail action returned success=false: %s", response.Error)
	}
	
	if response.Data["message_id"] == nil {
		t.Error("Gmail response missing message_id")
	}
	
	t.Logf("✅ Gmail send_message action successful: %v", response.Data)
	
	// Test invalid OAuth token
	_, err = mcpService.ExecuteAction("gmail", "send_message", map[string]interface{}{
		"to": "test@example.com",
	}, "invalid_token")
	
	if err == nil {
		t.Error("Expected invalid token to fail, but it succeeded")
	} else {
		t.Logf("✅ Invalid OAuth token properly rejected")
	}
	
	t.Logf("=== MOCK SERVER RESPONSE TEST COMPLETED ===")
}
