package services

import (
	"testing"
)

func TestMCPCatalogParser_ValidateMCPFunctions(t *testing.T) {
	parser := NewMCPCatalogParser()

	// Mock MCP catalog with Gmail and Docs services (updated structure)
	mockCatalog := map[string]interface{}{
		"providers": map[string]interface{}{
			"workspace": map[string]interface{}{
				"gmail": map[string]interface{}{
					"display_name": "Gmail",
					"description":  "Gmail service for sending emails",
					"functions": map[string]interface{}{
						"send_message": map[string]interface{}{
							"name":         "send_message",
							"display_name": "Send Message",
							"description":  "Send email message",
							"required_fields": []interface{}{"to", "subject", "body"},
							"example_payload": map[string]interface{}{
								"to":      "user@example.com",
								"subject": "Test Subject",
								"body":    "Test Body",
							},
						},
					},
				},
				"docs": map[string]interface{}{
					"display_name": "Google Docs",
					"description":  "Google Docs service for document management",
					"functions": map[string]interface{}{
						"create_document": map[string]interface{}{
							"name":         "create_document",
							"display_name": "Create Document",
							"description":  "Create new document",
							"required_fields": []interface{}{"title"},
							"example_payload": map[string]interface{}{
								"title":   "New Document",
								"content": "Document content",
							},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name     string
		steps    []map[string]interface{}
		expected bool
		errorCount int
	}{
		{
			name: "Valid workflow with existing functions",
			steps: []map[string]interface{}{
				{
					"id":     "send_email",
					"action": "gmail.send_message",
					"parameters": map[string]interface{}{
						"to":      "${user.recipient_email}",
						"subject": "${user.email_subject}",
						"body":    "${user.message_body}",
					},
				},
				{
					"id":     "create_doc",
					"action": "docs.create_document",
					"parameters": map[string]interface{}{
						"title": "${user.document_title}",
					},
				},
			},
			expected: true,
			errorCount: 0,
		},
		{
			name: "Invalid workflow with non-existent function",
			steps: []map[string]interface{}{
				{
					"id":     "invalid_step",
					"action": "nonexistent.function",
					"parameters": map[string]interface{}{},
				},
			},
			expected: false,
			errorCount: 1,
		},
		{
			name: "Invalid workflow with wrong service",
			steps: []map[string]interface{}{
				{
					"id":     "wrong_service",
					"action": "unknown_service.some_function",
					"parameters": map[string]interface{}{},
				},
			},
			expected: false,
			errorCount: 1,
		},
		{
			name: "Invalid workflow with malformed action",
			steps: []map[string]interface{}{
				{
					"id":     "malformed",
					"action": "invalid_action_format",
					"parameters": map[string]interface{}{},
				},
			},
			expected: false,
			errorCount: 1,
		},
		{
			name: "Step missing action field",
			steps: []map[string]interface{}{
				{
					"id": "missing_action",
					"parameters": map[string]interface{}{},
				},
			},
			expected: false,
			errorCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to strongly-typed structures for testing
			catalog, err := parser.ParseMCPCatalog(mockCatalog)
			if err != nil {
				t.Fatalf("Failed to parse MCP catalog: %v", err)
			}
			typedSteps, err := parser.ParseWorkflowSteps(tt.steps)
			if err != nil {
				t.Fatalf("Failed to parse workflow steps: %v", err)
			}
			isValid, errors := parser.ValidateMCPFunctionsTyped(catalog, typedSteps)
			
			if isValid != tt.expected {
				t.Errorf("ValidateMCPFunctions() = %v, expected %v", isValid, tt.expected)
			}
			
			if len(errors) != tt.errorCount {
				t.Errorf("ValidateMCPFunctions() returned %d errors, expected %d. Errors: %v", len(errors), tt.errorCount, errors)
			}
		})
	}
}

func TestMCPCatalogParser_ValidateMCPParameters(t *testing.T) {
	parser := NewMCPCatalogParser()

	// Mock MCP catalog with parameter requirements (updated structure)
	mockCatalog := map[string]interface{}{
		"providers": map[string]interface{}{
			"workspace": map[string]interface{}{
				"gmail": map[string]interface{}{
					"display_name": "Gmail",
					"description":  "Gmail service for sending emails",
					"functions": map[string]interface{}{
						"send_message": map[string]interface{}{
							"name":         "send_message",
							"display_name": "Send Message",
							"description":  "Send email message",
							"required_fields": []interface{}{"to", "subject", "body"},
							"example_payload": map[string]interface{}{
								"to":      "user@example.com",
								"subject": "Test Subject",
								"body":    "Test Body",
							},
						},
					},
				},
				"docs": map[string]interface{}{
					"display_name": "Google Docs",
					"description":  "Google Docs service for document management",
					"functions": map[string]interface{}{
						"create_document": map[string]interface{}{
							"name":         "create_document",
							"display_name": "Create Document",
							"description":  "Create new document",
							"required_fields": []interface{}{"title"},
							"example_payload": map[string]interface{}{
								"title":   "New Document",
								"content": "Document content",
							},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name       string
		steps      []map[string]interface{}
		expected   bool
		errorCount int
	}{
		{
			name: "Valid parameters with all required fields",
			steps: []map[string]interface{}{
				{
					"id":     "send_email",
					"action": "gmail.send_message",
					"parameters": map[string]interface{}{
						"to":      "${user.recipient_email}",
						"subject": "${user.email_subject}",
						"body":    "${user.message_body}",
					},
				},
			},
			expected:   true,
			errorCount: 0,
		},
		{
			name: "Missing required parameters",
			steps: []map[string]interface{}{
				{
					"id":     "incomplete_email",
					"action": "gmail.send_message",
					"parameters": map[string]interface{}{
						"to": "${user.recipient_email}",
						// Missing required 'subject' and 'body' parameters
					},
				},
			},
			expected:   false,
			errorCount: 2, // Missing 'subject' and 'body'
		},
		{
			name: "Valid parameter references",
			steps: []map[string]interface{}{
				{
					"id":     "valid_refs",
					"action": "docs.create_document",
					"parameters": map[string]interface{}{
						"title":   "${user.document_title}",
						"content": "${steps.previous_step.outputs.generated_content}",
						"env_var": "${API_KEY}",
						"computed": "${computed.timestamp}",
					},
				},
			},
			expected:   true,
			errorCount: 0,
		},
		{
			name: "Invalid parameter references",
			steps: []map[string]interface{}{
				{
					"id":     "invalid_refs",
					"action": "docs.create_document",
					"parameters": map[string]interface{}{
						"title":        "${user.document_title}",
						"invalid_ref1": "${invalid_format}",
						"invalid_ref2": "${user.}",
						"invalid_ref3": "${}",
					},
				},
			},
			expected:   false,
			errorCount: 3, // Three invalid parameter references
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to strongly-typed structures for testing
			catalog, err := parser.ParseMCPCatalog(mockCatalog)
			if err != nil {
				t.Fatalf("Failed to parse MCP catalog: %v", err)
			}
			typedSteps, err := parser.ParseWorkflowSteps(tt.steps)
			if err != nil {
				t.Fatalf("Failed to parse workflow steps: %v", err)
			}
			isValid, errors := parser.ValidateMCPParametersTyped(catalog, typedSteps)
			
			if isValid != tt.expected {
				t.Errorf("ValidateMCPParameters() = %v, expected %v", isValid, tt.expected)
			}
			
			if len(errors) != tt.errorCount {
				t.Errorf("ValidateMCPParameters() returned %d errors, expected %d. Errors: %v", len(errors), tt.errorCount, errors)
			}
		})
	}
}

func TestMCPCatalogParser_ValidateServiceBindings(t *testing.T) {
	parser := NewMCPCatalogParser()

	// Mock MCP catalog (updated structure)
	mockCatalog := map[string]interface{}{
		"providers": map[string]interface{}{
			"workspace": map[string]interface{}{
				"gmail": map[string]interface{}{
					"display_name": "Gmail",
					"description":  "Gmail service for sending emails",
					"functions": map[string]interface{}{
						"send_message": map[string]interface{}{
							"name":         "send_message",
							"display_name": "Send Message",
							"description":  "Send email message",
							"required_fields": []interface{}{"to", "subject", "body"},
							"example_payload": map[string]interface{}{
								"to":      "user@example.com",
								"subject": "Test Subject",
								"body":    "Test Body",
							},
						},
					},
				},
				"docs": map[string]interface{}{
					"display_name": "Google Docs",
					"description":  "Google Docs service for document management",
					"functions": map[string]interface{}{
						"create_document": map[string]interface{}{
							"name":         "create_document",
							"display_name": "Create Document",
							"description":  "Create new document",
							"required_fields": []interface{}{"title"},
							"example_payload": map[string]interface{}{
								"title":   "New Document",
								"content": "Document content",
							},
						},
					},
				},
			},
		},
	}

	// Mock workflow steps requiring Gmail and Docs
	workflowSteps := []map[string]interface{}{
		{
			"id":     "send_email",
			"action": "gmail.send_message",
		},
		{
			"id":     "create_doc",
			"action": "docs.create_document",
		},
	}

	tests := []struct {
		name           string
		serviceBindings map[string]interface{}
		expected       bool
		errorCount     int
	}{
		{
			name: "Valid service bindings with OAuth",
			serviceBindings: map[string]interface{}{
				"gmail": map[string]interface{}{
					"auth_type": "oauth2",
					"oauth_config": map[string]interface{}{
						"scopes": []interface{}{
							"https://www.googleapis.com/auth/gmail.send",
							"https://www.googleapis.com/auth/gmail.readonly",
						},
					},
				},
				"docs": map[string]interface{}{
					"auth_type": "oauth2",
					"oauth_config": map[string]interface{}{
						"scopes": []interface{}{
							"https://www.googleapis.com/auth/documents",
						},
					},
				},
			},
			expected:   true,
			errorCount: 0,
		},
		{
			name: "Missing service binding",
			serviceBindings: map[string]interface{}{
				"gmail": map[string]interface{}{
					"auth_type": "oauth2",
					"oauth_config": map[string]interface{}{
						"scopes": []interface{}{
							"https://www.googleapis.com/auth/gmail.send",
							"https://www.googleapis.com/auth/gmail.readonly",
						},
					},
				},
				// Missing 'docs' service binding
			},
			expected:   false,
			errorCount: 1,
		},
		{
			name: "Missing OAuth scopes",
			serviceBindings: map[string]interface{}{
				"gmail": map[string]interface{}{
					"auth_type": "oauth2",
					"oauth_config": map[string]interface{}{
						"scopes": []interface{}{
							"https://www.googleapis.com/auth/gmail.readonly",
							// Missing gmail.send scope
						},
					},
				},
				"docs": map[string]interface{}{
					"auth_type": "oauth2",
					"oauth_config": map[string]interface{}{
						"scopes": []interface{}{
							"https://www.googleapis.com/auth/documents",
						},
					},
				},
			},
			expected:   false,
			errorCount: 1,
		},
		{
			name: "Missing OAuth configuration",
			serviceBindings: map[string]interface{}{
				"gmail": map[string]interface{}{
					"auth_type": "oauth2",
					// Missing oauth_config
				},
				"docs": map[string]interface{}{
					"auth_type": "oauth2",
					"oauth_config": map[string]interface{}{
						"scopes": []interface{}{
							"https://www.googleapis.com/auth/documents",
						},
					},
				},
			},
			expected:   false,
			errorCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to strongly-typed structures for testing
			catalog, err := parser.ParseMCPCatalog(mockCatalog)
			if err != nil {
				t.Fatalf("Failed to parse MCP catalog: %v", err)
			}
			typedSteps, err := parser.ParseWorkflowSteps(workflowSteps)
			if err != nil {
				t.Fatalf("Failed to parse workflow steps: %v", err)
			}
			typedBindings, err := parser.ParseServiceBindings(tt.serviceBindings)
			if err != nil {
				t.Fatalf("Failed to parse service bindings: %v", err)
			}
			isValid, errors := parser.ValidateServiceBindingsTyped(catalog, typedBindings, typedSteps)
			
			if isValid != tt.expected {
				t.Errorf("ValidateServiceBindings() = %v, expected %v", isValid, tt.expected)
			}
			
			if len(errors) != tt.errorCount {
				t.Errorf("ValidateServiceBindings() returned %d errors, expected %d. Errors: %v", len(errors), tt.errorCount, errors)
			}
		})
	}
}

func TestMCPCatalogParser_isValidParameterReference(t *testing.T) {
	parser := NewMCPCatalogParser()

	tests := []struct {
		name      string
		paramRef  string
		expected  bool
	}{
		{
			name:     "Valid user parameter reference",
			paramRef: "${user.recipient_email}",
			expected: true,
		},
		{
			name:     "Valid step output reference",
			paramRef: "${steps.create_doc.outputs.document_url}",
			expected: true,
		},
		{
			name:     "Valid computed reference",
			paramRef: "${computed.current_timestamp}",
			expected: true,
		},
		{
			name:     "Valid environment variable",
			paramRef: "${API_KEY}",
			expected: true,
		},
		{
			name:     "Valid environment variable with underscores",
			paramRef: "${GOOGLE_CLIENT_ID}",
			expected: true,
		},
		{
			name:     "Invalid reference - no braces",
			paramRef: "user.recipient_email",
			expected: false,
		},
		{
			name:     "Invalid reference - no dollar sign",
			paramRef: "{user.recipient_email}",
			expected: false,
		},
		{
			name:     "Invalid reference - empty",
			paramRef: "${}",
			expected: false,
		},
		{
			name:     "Invalid reference - malformed user param",
			paramRef: "${user.}",
			expected: false,
		},
		{
			name:     "Invalid reference - malformed step output",
			paramRef: "${steps.step_id.outputs.}",
			expected: false,
		},
		{
			name:     "Invalid reference - wrong format",
			paramRef: "${invalid_format}",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.isValidParameterReference(tt.paramRef)
			if result != tt.expected {
				t.Errorf("isValidParameterReference(%s) = %v, expected %v", tt.paramRef, result, tt.expected)
			}
		})
	}
}

func TestMCPCatalogParser_getRequiredScopes(t *testing.T) {
	parser := NewMCPCatalogParser()

	tests := []struct {
		name        string
		serviceName string
		expected    []string
	}{
		{
			name:        "Gmail service scopes",
			serviceName: "gmail",
			expected: []string{
				"https://www.googleapis.com/auth/gmail.send",
				"https://www.googleapis.com/auth/gmail.readonly",
			},
		},
		{
			name:        "Docs service scopes",
			serviceName: "docs",
			expected: []string{
				"https://www.googleapis.com/auth/documents",
			},
		},
		{
			name:        "Drive service scopes",
			serviceName: "drive",
			expected: []string{
				"https://www.googleapis.com/auth/drive.file",
			},
		},
		{
			name:        "Calendar service scopes",
			serviceName: "calendar",
			expected: []string{
				"https://www.googleapis.com/auth/calendar",
			},
		},
		{
			name:        "Unknown service",
			serviceName: "unknown",
			expected:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.getRequiredScopes(tt.serviceName)
			if len(result) != len(tt.expected) {
				t.Errorf("getRequiredScopes(%s) returned %d scopes, expected %d", tt.serviceName, len(result), len(tt.expected))
				return
			}
			
			for i, scope := range result {
				if scope != tt.expected[i] {
					t.Errorf("getRequiredScopes(%s)[%d] = %s, expected %s", tt.serviceName, i, scope, tt.expected[i])
				}
			}
		})
	}
}
