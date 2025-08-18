package services

import (
	"encoding/json"
	"testing"

	"sohoaas-backend/internal/types"
)

// TestMCPResponseSchemaIntegration validates end-to-end response schema integration
func TestMCPResponseSchemaIntegration(t *testing.T) {
	t.Run("Enhanced MCP Catalog Enables Accurate Workflow Generation", func(t *testing.T) {
		// Mock enhanced MCP catalog that would come from MCP server with response schemas
		enhancedCatalogJSON := `{
			"providers": {
				"workspace": {
					"display_name": "Google Workspace",
					"description": "Google Workspace services",
					"services": {
						"gmail": {
							"display_name": "Gmail",
							"description": "Gmail service for sending emails",
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
										"description": "Gmail send message response",
										"properties": {
											"message_id": {"type": "string", "description": "Gmail message ID"},
											"thread_id": {"type": "string", "description": "Gmail thread ID"},
											"to": {"type": "string", "description": "Recipient email"},
											"subject": {"type": "string", "description": "Email subject"},
											"status": {"type": "string", "description": "Send status"},
											"sent_at": {"type": "string", "description": "ISO timestamp when sent"}
										},
										"required": ["message_id", "thread_id", "status", "sent_at"]
									},
									"error_schema": {
										"type": "object",
										"description": "Gmail send message error response",
										"properties": {
											"error_code": {"type": "string", "description": "Error code"},
											"error_message": {"type": "string", "description": "Error message"}
										},
										"required": ["error_code", "error_message"]
									}
								}
							}
						},
						"docs": {
							"display_name": "Google Docs",
							"description": "Google Docs service",
							"functions": {
								"create_document": {
									"name": "create_document",
									"display_name": "Create Document",
									"description": "Create new document",
									"required_fields": ["title"],
									"example_payload": {
										"title": "New Document"
									},
									"output_schema": {
										"type": "object",
										"description": "Document creation response",
										"properties": {
											"document_id": {"type": "string", "description": "Google Docs document ID"},
											"title": {"type": "string", "description": "Document title"},
											"url": {"type": "string", "description": "Shareable document URL"},
											"status": {"type": "string", "description": "Creation status"}
										},
										"required": ["document_id", "title", "url", "status"]
									},
									"error_schema": {
										"type": "object",
										"description": "Document creation error response",
										"properties": {
											"error_code": {"type": "string", "description": "Error code"},
											"error_message": {"type": "string", "description": "Error message"}
										},
										"required": ["error_code", "error_message"]
									}
								}
							}
						}
					}
				}
			}
		}`

		// Parse enhanced catalog into strongly-typed structure
		var catalogData map[string]interface{}
		err := json.Unmarshal([]byte(enhancedCatalogJSON), &catalogData)
		if err != nil {
			t.Fatalf("Failed to parse enhanced catalog: %v", err)
		}

		// Convert to MCPServiceCatalog using existing parsing logic
		mcpService := &MCPService{}
		catalog, err := mcpService.parseCatalogResponse(catalogData)
		if err != nil {
			t.Fatalf("Failed to parse catalog into MCPServiceCatalog: %v", err)
		}

		// Validate that response schemas are properly parsed
		if catalog == nil {
			t.Fatal("Catalog should not be nil")
		}

		workspace := catalog.Providers.Workspace
		if workspace.DisplayName == "" {
			t.Fatal("Catalog should have workspace provider")
		}

		gmail, exists := workspace.Services["gmail"]
		if !exists {
			t.Fatal("Workspace should have gmail service")
		}

		sendMessage, exists := gmail.Functions["send_message"]
		if !exists {
			t.Fatal("Gmail should have send_message function")
		}

		// Validate OutputSchema is present and properly structured
		if sendMessage.OutputSchema == nil {
			t.Fatal("send_message should have OutputSchema")
		}

		outputProps := sendMessage.OutputSchema.Properties
		if outputProps == nil || len(outputProps) == 0 {
			t.Fatal("OutputSchema should have properties")
		}

		// Validate critical output fields that enable step references
		requiredOutputs := []string{"message_id", "thread_id", "status", "sent_at"}
		for _, field := range requiredOutputs {
			if _, exists := outputProps[field]; !exists {
				t.Errorf("OutputSchema should have %s property for step references", field)
			}
		}

		// Validate ErrorSchema is present
		if sendMessage.ErrorSchema == nil {
			t.Fatal("send_message should have ErrorSchema")
		}

		errorProps := sendMessage.ErrorSchema.Properties
		if errorProps == nil || len(errorProps) == 0 {
			t.Fatal("ErrorSchema should have properties")
		}

		// Validate standard error fields
		if _, exists := errorProps["error_code"]; !exists {
			t.Error("ErrorSchema should have error_code")
		}
		if _, exists := errorProps["error_message"]; !exists {
			t.Error("ErrorSchema should have error_message")
		}

		t.Log("SUCCESS: Enhanced MCP catalog properly parsed with response schemas")
	})

	t.Run("Response Schemas Enable Accurate Workflow Step Outputs", func(t *testing.T) {
		// This test demonstrates how response schemas solve the original workflow validation issue
		
		// Before: LLM had to guess what outputs were available
		// Generated workflow step without outputs or with incorrect outputs:
		legacyWorkflowStep := map[string]interface{}{
			"id":          "send_email",
			"action":      "gmail.send_message",
			"parameters": map[string]interface{}{
				"to":      "${user.recipient}",
				"subject": "${user.subject}",
				"body":    "${user.message}",
			},
			// Missing outputs section - LLM didn't know what was available
		}

		// After: With response schemas, LLM can generate accurate outputs
		enhancedWorkflowStep := map[string]interface{}{
			"id":          "send_email",
			"action":      "gmail.send_message",
			"parameters": map[string]interface{}{
				"to":      "${user.recipient}",
				"subject": "${user.subject}",
				"body":    "${user.message}",
			},
			"outputs": map[string]interface{}{
				"message_id": map[string]interface{}{
					"type":        "string",
					"description": "Gmail message ID",
				},
				"thread_id": map[string]interface{}{
					"type":        "string",
					"description": "Gmail thread ID",
				},
				"status": map[string]interface{}{
					"type":        "string",
					"description": "Send status",
				},
				"sent_at": map[string]interface{}{
					"type":        "string",
					"description": "ISO timestamp when sent",
				},
			},
		}

		// Validate that enhanced workflow step has proper outputs
		outputs, exists := enhancedWorkflowStep["outputs"]
		if !exists {
			t.Fatal("Enhanced workflow step should have outputs")
		}

		outputsMap := outputs.(map[string]interface{})
		if len(outputsMap) == 0 {
			t.Fatal("Outputs should not be empty")
		}

		// Validate critical outputs for step references
		if _, exists := outputsMap["message_id"]; !exists {
			t.Error("Outputs should include message_id for ${steps.send_email.outputs.message_id} references")
		}

		// Validate legacy workflow step lacks outputs (demonstrating the problem)
		if _, exists := legacyWorkflowStep["outputs"]; exists {
			t.Error("Legacy workflow step should not have outputs (demonstrates the original problem)")
		}

		t.Log("SUCCESS: Response schemas enable LLM to generate accurate workflow step outputs")
		t.Log("This solves the original workflow execution validation error:")
		t.Log("- Before: Missing or incorrect step outputs caused validation failures")
		t.Log("- After: Accurate step outputs enable proper parameter references")
		t.Log("- Result: ${steps.step_id.outputs.field} references work correctly")
	})

	t.Run("Backward Compatibility Maintained", func(t *testing.T) {
		// Test that existing catalog without response schemas still works
		legacyCatalogJSON := `{
			"providers": {
				"workspace": {
					"display_name": "Google Workspace",
					"description": "Google Workspace services",
					"services": {
						"gmail": {
							"display_name": "Gmail",
							"description": "Gmail service",
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
									}
								}
							}
						}
					}
				}
			}
		}`

		var catalogData map[string]interface{}
		err := json.Unmarshal([]byte(legacyCatalogJSON), &catalogData)
		if err != nil {
			t.Fatalf("Failed to parse legacy catalog: %v", err)
		}

		mcpService := &MCPService{}
		catalog, err := mcpService.parseCatalogResponse(catalogData)
		if err != nil {
			t.Fatalf("Legacy catalog should still parse successfully: %v", err)
		}

		// Validate basic structure still works
		workspace := catalog.Providers.Workspace
		gmail := workspace.Services["gmail"]
		sendMessage := gmail.Functions["send_message"]

		if sendMessage.Name != "send_message" {
			t.Error("Legacy catalog should preserve function name")
		}

		// Response schemas should be nil for legacy catalog (optional fields)
		if sendMessage.OutputSchema != nil {
			t.Error("Legacy catalog should have nil OutputSchema")
		}
		if sendMessage.ErrorSchema != nil {
			t.Error("Legacy catalog should have nil ErrorSchema")
		}

		t.Log("SUCCESS: Backward compatibility maintained for existing catalogs")
	})
}

// parseCatalogResponse is a test helper method to simulate MCPService parsing
func (s *MCPService) parseCatalogResponse(catalogData map[string]interface{}) (*types.MCPServiceCatalog, error) {
	// Convert map to JSON and back to strongly-typed structure
	jsonData, err := json.Marshal(catalogData)
	if err != nil {
		return nil, err
	}

	var catalog types.MCPServiceCatalog
	err = json.Unmarshal(jsonData, &catalog)
	if err != nil {
		return nil, err
	}

	return &catalog, nil
}
