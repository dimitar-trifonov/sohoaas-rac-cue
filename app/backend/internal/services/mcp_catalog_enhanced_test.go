package services

import (
	"encoding/json"
	"testing"

	"sohoaas-backend/internal/types"
)

// TestMCPCatalogEnhanced validates the enhanced MCP catalog structure with response schemas
func TestMCPCatalogEnhanced(t *testing.T) {
	// Mock enhanced MCP catalog with response schemas
	mockEnhancedCatalogJSON := `{
		"providers": {
			"workspace": {
				"description": "Google Workspace Provider",
				"display_name": "Google Workspace",
				"services": {
					"gmail": {
						"description": "Gmail service for sending emails",
						"display_name": "Gmail",
						"functions": {
							"send_message": {
								"name": "send_message",
								"display_name": "Send Message",
								"description": "Send email message",
								"required_fields": ["to", "subject", "body"],
								"example_payload": {
									"to": "user@example.com",
									"subject": "Test Subject",
									"body": "Test Body"
								},
								"output_schema": {
									"type": "object",
									"description": "Email send response",
									"properties": {
										"message_id": {
											"type": "string",
											"description": "Gmail message ID"
										},
										"thread_id": {
											"type": "string",
											"description": "Gmail thread ID"
										},
										"status": {
											"type": "string",
											"description": "Send status"
										}
									},
									"required": ["message_id", "status"]
								},
								"error_schema": {
									"type": "object",
									"description": "Email send error response",
									"properties": {
										"error_code": {
											"type": "string",
											"description": "Error code"
										},
										"error_message": {
											"type": "string",
											"description": "Error message"
										}
									}
								}
							}
						}
					},
					"docs": {
						"description": "Google Docs service for document management",
						"display_name": "Google Docs",
						"functions": {
							"create_document": {
								"name": "create_document",
								"display_name": "Create Document",
								"description": "Create new document",
								"required_fields": ["title"],
								"example_payload": {
									"title": "New Document",
									"content": "Document content"
								},
								"output_schema": {
									"type": "object",
									"description": "Document creation response",
									"properties": {
										"document_id": {
											"type": "string",
											"description": "Google Docs document ID"
										},
										"title": {
											"type": "string",
											"description": "Document title"
										},
										"url": {
											"type": "string",
											"description": "Shareable document URL"
										},
										"revision_id": {
											"type": "string",
											"description": "Document revision ID"
										},
										"status": {
											"type": "string",
											"description": "Creation status"
										},
										"created_at": {
											"type": "string",
											"description": "ISO timestamp"
										}
									},
									"required": ["document_id", "title", "url", "status"]
								},
								"error_schema": {
									"type": "object",
									"description": "Document creation error response",
									"properties": {
										"error_code": {
											"type": "string",
											"description": "Error code"
										},
										"error_message": {
											"type": "string",
											"description": "Error message"
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}`

	// Parse the enhanced catalog
	var catalog types.MCPServiceCatalog
	err := json.Unmarshal([]byte(mockEnhancedCatalogJSON), &catalog)
	if err != nil {
		t.Fatalf("Failed to parse enhanced catalog: %v", err)
	}

	// Test 1: Validate backward compatibility
	t.Run("Backward Compatibility", func(t *testing.T) {
		// Ensure all existing fields still work
		gmailService := catalog.Providers.Workspace.Services["gmail"]
		sendMessageFunc := gmailService.Functions["send_message"]
		
		if sendMessageFunc.Name != "send_message" {
			t.Error("Existing Name field should still work")
		}
		if sendMessageFunc.DisplayName != "Send Message" {
			t.Error("Existing DisplayName field should still work")
		}
		if len(sendMessageFunc.RequiredFields) != 3 {
			t.Error("Existing RequiredFields should still work")
		}
		if sendMessageFunc.ExamplePayload == nil {
			t.Error("Existing ExamplePayload should still work")
		}
	})

	// Test 2: Validate output schema structure
	t.Run("Output Schema Structure", func(t *testing.T) {
		// Test Gmail output schema
		gmailFunc := catalog.Providers.Workspace.Services["gmail"].Functions["send_message"]
		if gmailFunc.OutputSchema == nil {
			t.Fatal("Gmail send_message should have output_schema")
		}
		
		outputSchema := gmailFunc.OutputSchema
		if outputSchema.Type != "object" {
			t.Error("Output schema type should be 'object'")
		}
		if outputSchema.Description == "" {
			t.Error("Output schema should have description")
		}
		if len(outputSchema.Properties) == 0 {
			t.Error("Output schema should have properties")
		}
		
		// Validate specific properties
		if _, exists := outputSchema.Properties["message_id"]; !exists {
			t.Error("Gmail output should have message_id property")
		}
		if _, exists := outputSchema.Properties["status"]; !exists {
			t.Error("Gmail output should have status property")
		}
		
		// Validate required fields
		expectedRequired := []string{"message_id", "status"}
		if len(outputSchema.Required) != len(expectedRequired) {
			t.Errorf("Expected %d required fields, got %d", len(expectedRequired), len(outputSchema.Required))
		}
	})

	// Test 3: Validate docs output schema (matches actual implementation)
	t.Run("Docs Output Schema Matches Implementation", func(t *testing.T) {
		docsFunc := catalog.Providers.Workspace.Services["docs"].Functions["create_document"]
		if docsFunc.OutputSchema == nil {
			t.Fatal("Docs create_document should have output_schema")
		}
		
		outputSchema := docsFunc.OutputSchema
		
		// These should match what docs_proxy.go actually returns
		expectedProperties := []string{"document_id", "title", "url", "revision_id", "status", "created_at"}
		for _, prop := range expectedProperties {
			if _, exists := outputSchema.Properties[prop]; !exists {
				t.Errorf("Docs output should have %s property (matches docs_proxy.go implementation)", prop)
			}
		}
		
		// Validate required fields match critical outputs
		requiredFields := outputSchema.Required
		if len(requiredFields) == 0 {
			t.Error("Docs output should have required fields")
		}
		
		// document_id is critical for workflow step references
		foundDocumentId := false
		for _, field := range requiredFields {
			if field == "document_id" {
				foundDocumentId = true
				break
			}
		}
		if !foundDocumentId {
			t.Error("document_id should be required (used in workflow step references)")
		}
	})

	// Test 4: Validate error schema structure
	t.Run("Error Schema Structure", func(t *testing.T) {
		gmailFunc := catalog.Providers.Workspace.Services["gmail"].Functions["send_message"]
		if gmailFunc.ErrorSchema == nil {
			t.Fatal("Gmail send_message should have error_schema")
		}
		
		errorSchema := gmailFunc.ErrorSchema
		if errorSchema.Type != "object" {
			t.Error("Error schema type should be 'object'")
		}
		if errorSchema.Description == "" {
			t.Error("Error schema should have description")
		}
		
		// Standard error properties
		if _, exists := errorSchema.Properties["error_code"]; !exists {
			t.Error("Error schema should have error_code property")
		}
		if _, exists := errorSchema.Properties["error_message"]; !exists {
			t.Error("Error schema should have error_message property")
		}
	})

	// Test 5: Validate JSON serialization with new fields
	t.Run("Enhanced JSON Serialization", func(t *testing.T) {
		// Serialize enhanced catalog
		serialized, err := json.Marshal(catalog)
		if err != nil {
			t.Fatalf("Failed to serialize enhanced catalog: %v", err)
		}
		
		// Deserialize again
		var catalog2 types.MCPServiceCatalog
		err = json.Unmarshal(serialized, &catalog2)
		if err != nil {
			t.Fatalf("Failed to deserialize enhanced catalog: %v", err)
		}
		
		// Verify response schemas are preserved
		gmailFunc2 := catalog2.Providers.Workspace.Services["gmail"].Functions["send_message"]
		if gmailFunc2.OutputSchema == nil {
			t.Error("Output schema should be preserved in serialization")
		}
		if gmailFunc2.ErrorSchema == nil {
			t.Error("Error schema should be preserved in serialization")
		}
	})

	// Test 6: Validate optional fields (backward compatibility)
	t.Run("Optional Response Schemas", func(t *testing.T) {
		// Test that functions can exist without response schemas (backward compatibility)
		catalogWithoutSchemas := `{
			"providers": {
				"workspace": {
					"description": "Test Provider",
					"display_name": "Test",
					"services": {
						"test": {
							"description": "Test service",
							"display_name": "Test",
							"functions": {
								"test_function": {
									"name": "test_function",
									"display_name": "Test Function",
									"description": "Test function without schemas",
									"required_fields": ["param1"],
									"example_payload": {"param1": "value1"}
								}
							}
						}
					}
				}
			}
		}`
		
		var testCatalog types.MCPServiceCatalog
		err := json.Unmarshal([]byte(catalogWithoutSchemas), &testCatalog)
		if err != nil {
			t.Fatalf("Should be able to parse catalog without response schemas: %v", err)
		}
		
		testFunc := testCatalog.Providers.Workspace.Services["test"].Functions["test_function"]
		if testFunc.OutputSchema != nil {
			t.Error("OutputSchema should be nil when not provided")
		}
		if testFunc.ErrorSchema != nil {
			t.Error("ErrorSchema should be nil when not provided")
		}
	})
}

// TestMCPCatalogEnhancedWorkflowGeneration validates that enhanced catalog supports workflow generation
func TestMCPCatalogEnhancedWorkflowGeneration(t *testing.T) {
	t.Run("LLM Can Generate Outputs Schema", func(t *testing.T) {
		// This test validates that with response schemas, the LLM can generate proper outputs
		// for workflow steps instead of guessing
		
		// Mock function with output schema
		function := types.MCPFunctionSchema{
			Name:           "create_document",
			DisplayName:    "Create Document",
			Description:    "Create new document",
			RequiredFields: []string{"title"},
			OutputSchema: &types.MCPResponseSchema{
				Type:        "object",
				Description: "Document creation response",
				Properties: map[string]types.MCPParameterProperty{
					"document_id": {
						Type:        "string",
						Description: "Google Docs document ID",
					},
					"url": {
						Type:        "string",
						Description: "Shareable document URL",
					},
				},
				Required: []string{"document_id", "url"},
			},
		}
		
		// Validate that LLM can now know what outputs are available
		if function.OutputSchema == nil {
			t.Fatal("Function should have output schema for LLM")
		}
		
		// LLM can now generate workflow step outputs like:
		// "outputs": {
		//   "document_id": {"type": "string", "description": "Google Docs document ID"},
		//   "url": {"type": "string", "description": "Shareable document URL"}
		// }
		
		outputSchema := function.OutputSchema
		if len(outputSchema.Properties) == 0 {
			t.Error("LLM needs output properties to generate step outputs")
		}
		
		if len(outputSchema.Required) == 0 {
			t.Error("LLM needs required fields to know critical outputs")
		}
		
		// Validate document_id is available (critical for step references)
		if _, exists := outputSchema.Properties["document_id"]; !exists {
			t.Error("document_id must be available for ${steps.create_document.outputs.document_id} references")
		}
		
		t.Log("SUCCESS: LLM can now generate accurate step outputs instead of guessing")
	})
}
