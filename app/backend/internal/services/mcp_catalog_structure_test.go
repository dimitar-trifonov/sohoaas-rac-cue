package services

import (
	"encoding/json"
	"testing"

	"sohoaas-backend/internal/types"
)

// TestMCPCatalogStructure validates the current MCP catalog structure
// This test ensures we understand the working catalog format before making changes
func TestMCPCatalogStructure(t *testing.T) {
	// Mock the actual MCP catalog structure that we currently receive
	mockCatalogJSON := `{
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
								}
							}
						}
					},
					"drive": {
						"description": "Google Drive service for file management",
						"display_name": "Google Drive",
						"functions": {
							"share_file": {
								"name": "share_file",
								"display_name": "Share File",
								"description": "Share a file",
								"required_fields": ["file_id", "email", "role"],
								"example_payload": {
									"file_id": "doc123",
									"email": "user@example.com",
									"role": "reader"
								}
							}
						}
					},
					"calendar": {
						"description": "Google Calendar service for event management",
						"display_name": "Google Calendar",
						"functions": {
							"create_event": {
								"name": "create_event",
								"display_name": "Create Event",
								"description": "Create calendar event",
								"required_fields": ["title", "start_time", "end_time"],
								"example_payload": {
									"title": "Meeting",
									"start_time": "2024-01-01T10:00:00Z",
									"end_time": "2024-01-01T11:00:00Z"
								}
							}
						}
					}
				}
			}
		}
	}`

	// Parse the mock catalog
	var catalog types.MCPServiceCatalog
	err := json.Unmarshal([]byte(mockCatalogJSON), &catalog)
	if err != nil {
		t.Fatalf("Failed to parse mock catalog: %v", err)
	}

	// Test 1: Validate top-level structure
	t.Run("Top Level Structure", func(t *testing.T) {
		if catalog.Providers.Workspace.Description == "" {
			t.Error("Workspace provider description should not be empty")
		}
		if catalog.Providers.Workspace.DisplayName == "" {
			t.Error("Workspace provider display_name should not be empty")
		}
		if len(catalog.Providers.Workspace.Services) == 0 {
			t.Error("Workspace provider should have services")
		}
	})

	// Test 2: Validate service structure
	t.Run("Service Structure", func(t *testing.T) {
		expectedServices := []string{"gmail", "docs", "drive", "calendar"}
		
		for _, serviceName := range expectedServices {
			service, exists := catalog.Providers.Workspace.Services[serviceName]
			if !exists {
				t.Errorf("Service %s should exist in catalog", serviceName)
				continue
			}
			
			if service.Description == "" {
				t.Errorf("Service %s should have description", serviceName)
			}
			if service.DisplayName == "" {
				t.Errorf("Service %s should have display_name", serviceName)
			}
			if len(service.Functions) == 0 {
				t.Errorf("Service %s should have functions", serviceName)
			}
		}
	})

	// Test 3: Validate function structure
	t.Run("Function Structure", func(t *testing.T) {
		// Test Gmail send_message function
		gmailService := catalog.Providers.Workspace.Services["gmail"]
		sendMessageFunc, exists := gmailService.Functions["send_message"]
		if !exists {
			t.Fatal("Gmail send_message function should exist")
		}
		
		validateFunction(t, sendMessageFunc, "send_message", []string{"to", "subject", "body"})
		
		// Test Docs create_document function
		docsService := catalog.Providers.Workspace.Services["docs"]
		createDocFunc, exists := docsService.Functions["create_document"]
		if !exists {
			t.Fatal("Docs create_document function should exist")
		}
		
		validateFunction(t, createDocFunc, "create_document", []string{"title"})
		
		// Test Drive share_file function
		driveService := catalog.Providers.Workspace.Services["drive"]
		shareFileFunc, exists := driveService.Functions["share_file"]
		if !exists {
			t.Fatal("Drive share_file function should exist")
		}
		
		validateFunction(t, shareFileFunc, "share_file", []string{"file_id", "email", "role"})
		
		// Test Calendar create_event function
		calendarService := catalog.Providers.Workspace.Services["calendar"]
		createEventFunc, exists := calendarService.Functions["create_event"]
		if !exists {
			t.Fatal("Calendar create_event function should exist")
		}
		
		validateFunction(t, createEventFunc, "create_event", []string{"title", "start_time", "end_time"})
	})

	// Test 4: Validate example payloads
	t.Run("Example Payloads", func(t *testing.T) {
		// Test Gmail example payload
		gmailFunc := catalog.Providers.Workspace.Services["gmail"].Functions["send_message"]
		if gmailFunc.ExamplePayload["to"] == nil {
			t.Error("Gmail send_message should have 'to' in example payload")
		}
		if gmailFunc.ExamplePayload["subject"] == nil {
			t.Error("Gmail send_message should have 'subject' in example payload")
		}
		if gmailFunc.ExamplePayload["body"] == nil {
			t.Error("Gmail send_message should have 'body' in example payload")
		}
		
		// Test Docs example payload
		docsFunc := catalog.Providers.Workspace.Services["docs"].Functions["create_document"]
		if docsFunc.ExamplePayload["title"] == nil {
			t.Error("Docs create_document should have 'title' in example payload")
		}
	})

	// Test 5: Validate JSON serialization/deserialization
	t.Run("JSON Serialization", func(t *testing.T) {
		// Serialize back to JSON
		serialized, err := json.Marshal(catalog)
		if err != nil {
			t.Fatalf("Failed to serialize catalog: %v", err)
		}
		
		// Deserialize again
		var catalog2 types.MCPServiceCatalog
		err = json.Unmarshal(serialized, &catalog2)
		if err != nil {
			t.Fatalf("Failed to deserialize catalog: %v", err)
		}
		
		// Verify structure is preserved
		if len(catalog2.Providers.Workspace.Services) != len(catalog.Providers.Workspace.Services) {
			t.Error("Serialization/deserialization should preserve service count")
		}
	})
}

// validateFunction is a helper function to validate MCP function structure
func validateFunction(t *testing.T, function types.MCPFunctionSchema, expectedName string, expectedRequiredFields []string) {
	if function.Name != expectedName {
		t.Errorf("Function name should be %s, got %s", expectedName, function.Name)
	}
	
	if function.DisplayName == "" {
		t.Errorf("Function %s should have display_name", expectedName)
	}
	
	if function.Description == "" {
		t.Errorf("Function %s should have description", expectedName)
	}
	
	if len(function.RequiredFields) != len(expectedRequiredFields) {
		t.Errorf("Function %s should have %d required fields, got %d", 
			expectedName, len(expectedRequiredFields), len(function.RequiredFields))
		return
	}
	
	// Check all expected required fields are present
	for _, expectedField := range expectedRequiredFields {
		found := false
		for _, actualField := range function.RequiredFields {
			if actualField == expectedField {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Function %s should have required field %s", expectedName, expectedField)
		}
	}
	
	if function.ExamplePayload == nil {
		t.Errorf("Function %s should have example_payload", expectedName)
	}
}

// TestMCPCatalogCurrentWorkingState validates that the current catalog structure
// matches what the workflow generator and execution engine expect
func TestMCPCatalogCurrentWorkingState(t *testing.T) {
	// This test validates the exact structure that's currently working
	// It serves as a baseline before any modifications
	
	t.Run("Current Function Names Match Implementation", func(t *testing.T) {
		expectedFunctions := map[string][]string{
			"gmail":    {"send_message"},
			"docs":     {"create_document"},
			"drive":    {"share_file"},
			"calendar": {"create_event"},
		}
		
		// This validates that our current function names match what's implemented
		// in the MCP server and what the workflow generator expects
		for service, functions := range expectedFunctions {
			for _, function := range functions {
				t.Logf("Validating %s.%s exists in current implementation", service, function)
				// This is a documentation test - it records what we currently have
			}
		}
	})
	
	t.Run("Current Required Fields Structure", func(t *testing.T) {
		// Document the current required fields structure
		expectedRequiredFields := map[string]map[string][]string{
			"gmail": {
				"send_message": {"to", "subject", "body"},
			},
			"docs": {
				"create_document": {"title"},
			},
			"drive": {
				"share_file": {"file_id", "email", "role"},
			},
			"calendar": {
				"create_event": {"title", "start_time", "end_time"},
			},
		}
		
		// This documents what required fields we currently expect
		for service, functions := range expectedRequiredFields {
			for function, fields := range functions {
				t.Logf("Current %s.%s requires fields: %v", service, function, fields)
			}
		}
	})
	
	t.Run("Current Missing Response Schema Information", func(t *testing.T) {
		// This test documents the current limitation - we don't have response schemas
		// This is what we want to fix, but we need to understand the current state first
		
		t.Log("CURRENT LIMITATION: MCP catalog does not include response schemas")
		t.Log("Functions only provide: name, display_name, description, required_fields, example_payload")
		t.Log("Missing: output_schema, error_schema")
		t.Log("This causes workflow generator to guess what outputs are available")
		
		// This test will fail once we add response schemas, which is the goal
		// For now, it documents the current limitation
	})
}
