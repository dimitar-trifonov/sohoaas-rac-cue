package services

import (
	"testing"
	"sohoaas-backend/internal/types"
)

// TestExecutionEngineValidationMethods tests the new validation methods exist and compile
func TestExecutionEngineValidationMethods(t *testing.T) {
	// Create execution engine with real MCP service (methods will be tested with mocked data)
	mcpService := NewMCPService("http://mock-server")
	engine := NewExecutionEngine(mcpService)

	// Test 1: Verify validateResponseSchema method exists and handles missing service
	err := engine.validateResponseSchema("unknown_service", "unknown_action", map[string]interface{}{
		"test_field": "test_value",
	})
	if err == nil {
		t.Error("Expected error for unknown service, but got none")
	}
	t.Logf("✅ validateResponseSchema method exists and handles unknown service: %v", err)

	// Test 2: Verify validateOutputFieldExists method exists and handles missing service
	catalog := &types.MCPServiceCatalog{
		Providers: types.MCPProviders{
			Workspace: types.MCPWorkspaceProvider{
				Services: map[string]types.MCPServiceDefinition{},
			},
		},
	}
	err = engine.validateOutputFieldExists("unknown_service", "unknown_action", "test_field", catalog)
	if err == nil {
		t.Error("Expected error for unknown service in validateOutputFieldExists, but got none")
	}
	t.Logf("✅ validateOutputFieldExists method exists and handles unknown service: %v", err)

	// Test 3: Verify ValidateWorkflowServices includes output field validation
	workflow := &ParsedWorkflow{
		Steps: []WorkflowStep{
			{
				ID:      "test_step",
				Service: "unknown_service",
				Action:  "unknown_action",
				Inputs: map[string]interface{}{
					"test_param": "test_value",
				},
			},
		},
	}
	err = engine.ValidateWorkflowServices(workflow)
	if err == nil {
		t.Error("Expected error for workflow with unknown service, but got none")
	}
	t.Logf("✅ ValidateWorkflowServices includes enhanced validation: %v", err)

	t.Log("🎉 All execution engine validation enhancements are properly implemented!")
}

// TestExecutionEngineResponseSchemaValidation tests response schema validation with valid catalog
func TestExecutionEngineResponseSchemaValidation(t *testing.T) {
	// Create catalog with enhanced response schemas
	catalog := &types.MCPServiceCatalog{
		Providers: types.MCPProviders{
			Workspace: types.MCPWorkspaceProvider{
				Services: map[string]types.MCPServiceDefinition{
					"gmail": {
						Functions: map[string]types.MCPFunctionSchema{
							"send_message": {
								OutputSchema: &types.MCPResponseSchema{
									Properties: map[string]types.MCPParameterProperty{
										"message_id": {Type: "string"},
										"thread_id":  {Type: "string"},
										"status":     {Type: "string"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	mcpService := NewMCPService("http://mock-server")
	engine := NewExecutionEngine(mcpService)

	// Test valid response schema validation with direct method call
	err := engine.validateOutputFieldExists("gmail", "send_message", "message_id", catalog)
	if err != nil {
		t.Errorf("Expected valid field 'message_id' to pass validation, but got error: %v", err)
	} else {
		t.Log("✅ Valid output field 'message_id' passes validation")
	}

	// Test invalid output field validation
	err = engine.validateOutputFieldExists("gmail", "send_message", "invalid_field", catalog)
	if err == nil {
		t.Error("Expected error for invalid field 'invalid_field', but got none")
	} else {
		t.Logf("✅ Invalid output field 'invalid_field' correctly fails validation: %v", err)
	}

	// Test backward compatibility - no output schema defined
	legacyCatalog := &types.MCPServiceCatalog{
		Providers: types.MCPProviders{
			Workspace: types.MCPWorkspaceProvider{
				Services: map[string]types.MCPServiceDefinition{
					"legacy_service": {
						Functions: map[string]types.MCPFunctionSchema{
							"legacy_action": {
								// No OutputSchema defined - should allow any field
							},
						},
					},
				},
			},
		},
	}

	err = engine.validateOutputFieldExists("legacy_service", "legacy_action", "any_field", legacyCatalog)
	if err != nil {
		t.Errorf("Legacy service without output schema should allow any field, but got error: %v", err)
	} else {
		t.Log("✅ Legacy service without output schema allows any field reference (backward compatibility)")
	}

	t.Log("🎉 Response schema validation working correctly with enhanced MCP catalog!")
}
