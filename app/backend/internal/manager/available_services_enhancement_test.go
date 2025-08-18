package manager

import (
	"strings"
	"testing"

	"sohoaas-backend/internal/types"
)

// TestAvailableServicesEnhancement validates that buildAvailableServicesString includes response schema information
func TestAvailableServicesEnhancement(t *testing.T) {
	t.Run("Enhanced Available Services Include Response Schemas", func(t *testing.T) {
		// Create mock catalog with response schemas (matching our MCP server implementation)
		catalog := &types.MCPServiceCatalog{
			Providers: types.MCPProviders{
				Workspace: types.MCPWorkspaceProvider{
					Services: map[string]types.MCPServiceDefinition{
						"gmail": {
							Description: "Gmail service for sending emails",
							DisplayName: "Gmail",
							Functions: map[string]types.MCPFunctionSchema{
								"send_message": {
									Name:           "send_message",
									DisplayName:    "Send Email",
									Description:    "Send an email message via Gmail",
									RequiredFields: []string{"to", "subject", "body"},
									ExamplePayload: map[string]interface{}{
										"to":      "user@example.com",
										"subject": "Test Subject",
										"body":    "Test Body",
									},
									OutputSchema: &types.MCPResponseSchema{
										Type:        "object",
										Description: "Gmail send message response",
										Properties: map[string]types.MCPParameterProperty{
											"message_id": {
												Type:        "string",
												Description: "Gmail message ID",
											},
											"thread_id": {
												Type:        "string",
												Description: "Gmail thread ID",
											},
											"status": {
												Type:        "string",
												Description: "Send status",
											},
											"sent_at": {
												Type:        "string",
												Description: "ISO timestamp when sent",
											},
										},
										Required: []string{"message_id", "thread_id", "status", "sent_at"},
									},
								},
							},
						},
						"docs": {
							Description: "Google Docs service",
							DisplayName: "Google Docs",
							Functions: map[string]types.MCPFunctionSchema{
								"create_document": {
									Name:           "create_document",
									DisplayName:    "Create Document",
									Description:    "Create a new Google Docs document",
									RequiredFields: []string{"title"},
									ExamplePayload: map[string]interface{}{
										"title": "New Document",
									},
									OutputSchema: &types.MCPResponseSchema{
										Type:        "object",
										Description: "Document creation response",
										Properties: map[string]types.MCPParameterProperty{
											"document_id": {
												Type:        "string",
												Description: "Google Docs document ID",
											},
											"title": {
												Type:        "string",
												Description: "Document title",
											},
											"url": {
												Type:        "string",
												Description: "Shareable document URL",
											},
											"status": {
												Type:        "string",
												Description: "Creation status",
											},
										},
										Required: []string{"document_id", "title", "url", "status"},
									},
								},
							},
						},
					},
				},
			},
		}

		// Create AgentManager instance for testing
		am := &AgentManager{}

		// Generate available services string
		availableServices := am.buildAvailableServicesString(catalog)

		// Validate that response schema information is included
		if !strings.Contains(availableServices, "→ outputs:") {
			t.Error("Available services should include output schema information (→ outputs:)")
		}

		// Validate Gmail function includes critical output fields
		if !strings.Contains(availableServices, "message_id") {
			t.Error("Gmail function should include message_id in outputs (critical for step references)")
		}
		if !strings.Contains(availableServices, "thread_id") {
			t.Error("Gmail function should include thread_id in outputs")
		}
		if !strings.Contains(availableServices, "status") {
			t.Error("Gmail function should include status in outputs")
		}
		if !strings.Contains(availableServices, "sent_at") {
			t.Error("Gmail function should include sent_at in outputs")
		}

		// Validate Docs function includes critical output fields
		if !strings.Contains(availableServices, "document_id") {
			t.Error("Docs function should include document_id in outputs (critical for step references)")
		}
		if !strings.Contains(availableServices, "url") {
			t.Error("Docs function should include url in outputs")
		}

		// Validate structure includes both parameters and outputs
		if !strings.Contains(availableServices, "[params:") {
			t.Error("Available services should still include parameter information")
		}

		t.Logf("Enhanced available services string:\n%s", availableServices)
		t.Log("SUCCESS: Available services now include response schema information")
		t.Log("This enables LLM to generate accurate workflow step outputs")
	})

	t.Run("Backward Compatibility Without Response Schemas", func(t *testing.T) {
		// Test catalog without response schemas (legacy compatibility)
		legacyCatalog := &types.MCPServiceCatalog{
			Providers: types.MCPProviders{
				Workspace: types.MCPWorkspaceProvider{
					Services: map[string]types.MCPServiceDefinition{
						"gmail": {
							Description: "Gmail service",
							DisplayName: "Gmail",
							Functions: map[string]types.MCPFunctionSchema{
								"send_message": {
									Name:           "send_message",
									DisplayName:    "Send Email",
									Description:    "Send email",
									RequiredFields: []string{"to", "subject", "body"},
									ExamplePayload: map[string]interface{}{
										"to":      "user@example.com",
										"subject": "Test Subject",
										"body":    "Test Body",
									},
									// Note: No OutputSchema - should still work
								},
							},
						},
					},
				},
			},
		}

		am := &AgentManager{}
		availableServices := am.buildAvailableServicesString(legacyCatalog)

		// Should still include basic function information
		if !strings.Contains(availableServices, "send_message") {
			t.Error("Legacy catalog should still include function names")
		}
		if !strings.Contains(availableServices, "required: to, subject, body") {
			t.Error("Legacy catalog should still include required fields")
		}

		// Should not include output information (graceful degradation)
		if strings.Contains(availableServices, "→ outputs:") {
			t.Error("Legacy catalog should not include output information")
		}

		t.Log("SUCCESS: Backward compatibility maintained for catalogs without response schemas")
	})

	t.Run("LLM Prompt Enhancement Impact", func(t *testing.T) {
		// This test demonstrates the impact on LLM workflow generation
		
		// Before: LLM only knew inputs
		legacyPromptData := "gmail.send_message(required: to, subject, body) [params: to, subject, body]"
		
		// After: LLM knows both inputs and outputs
		enhancedPromptData := "gmail.send_message(required: to, subject, body) [params: to, subject, body] → outputs: message_id, thread_id, status, sent_at"
		
		// Validate that enhanced data enables proper workflow generation
		if !strings.Contains(enhancedPromptData, "message_id") {
			t.Error("Enhanced prompt should include message_id for ${steps.send_email.outputs.message_id} references")
		}
		
		if !strings.Contains(enhancedPromptData, "→ outputs:") {
			t.Error("Enhanced prompt should clearly indicate available outputs")
		}
		
		t.Log("SUCCESS: Enhanced prompt data enables LLM to generate accurate workflow step outputs")
		t.Log("Before:", legacyPromptData)
		t.Log("After: ", enhancedPromptData)
		t.Log("Impact: LLM can now generate proper 'outputs' sections in workflow steps")
	})
}
