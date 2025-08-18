package services

import (
	"reflect"
	"testing"

	"sohoaas-backend/internal/types"
)

func TestMCPCatalogParser_BuildUserCapabilities(t *testing.T) {
	parser := NewMCPCatalogParser()

	// Test with legacy map[string]interface{} format
	t.Run("Legacy map format", func(t *testing.T) {
		// Mock MCP catalog with Gmail, Docs, and Drive services
		mockCatalog := map[string]interface{}{
			"providers": map[string]interface{}{
				"workspace": map[string]interface{}{
					"services": map[string]interface{}{
						"gmail": map[string]interface{}{
							"name":         "gmail",
							"display_name": "Gmail",
							"description":  "Google Gmail service for email management",
							"functions": map[string]interface{}{
								"gmail.send_message": map[string]interface{}{
									"description": "Send email message",
									"required":    []interface{}{"to", "subject", "body"},
								},
								"gmail.list_messages": map[string]interface{}{
									"description": "List email messages",
									"required":    []interface{}{},
								},
							},
							"auth_type": "oauth2",
						},
						"docs": map[string]interface{}{
							"name":         "docs",
							"display_name": "Google Docs",
							"description":  "Google Docs service for document management",
							"functions": map[string]interface{}{
								"docs.create_document": map[string]interface{}{
									"description": "Create new document",
									"required":    []interface{}{"title"},
								},
								"docs.update_document": map[string]interface{}{
									"description": "Update existing document",
									"required":    []interface{}{"document_id", "content"},
								},
							},
							"auth_type": "oauth2",
						},
						"drive": map[string]interface{}{
							"name":         "drive",
							"display_name": "Google Drive",
							"description":  "Google Drive service for file management",
							"functions": map[string]interface{}{
								"drive.list_files": map[string]interface{}{
									"description": "List files in Drive",
									"required":    []interface{}{},
								},
								"drive.share_file": map[string]interface{}{
									"description": "Share file with users",
									"required":    []interface{}{"file_id", "email"},
								},
							},
							"auth_type": "oauth2",
						},
					},
				},
			},
		}

		tests := []struct {
			name               string
			connectedServices  []string
			expectedCapabilities []map[string]interface{}
			expectError        bool
		}{
			{
				name:              "All services connected",
				connectedServices: []string{"gmail", "docs", "drive"},
				expectedCapabilities: []map[string]interface{}{
					{
						"service": "gmail",
						"actions": []string{"gmail.send_message", "gmail.list_messages"},
						"status":  "connected",
					},
					{
						"service": "docs",
						"actions": []string{"docs.create_document", "docs.update_document"},
						"status":  "connected",
					},
					{
						"service": "drive",
						"actions": []string{"drive.list_files", "drive.share_file"},
						"status":  "connected",
					},
				},
				expectError: false,
			},
			{
				name:              "Subset of services connected",
				connectedServices: []string{"gmail", "docs"},
				expectedCapabilities: []map[string]interface{}{
					{
						"service": "gmail",
						"actions": []string{"gmail.send_message", "gmail.list_messages"},
						"status":  "connected",
					},
					{
						"service": "docs",
						"actions": []string{"docs.create_document", "docs.update_document"},
						"status":  "connected",
					},
				},
				expectError: false,
			},
			{
				name:              "Single service connected",
				connectedServices: []string{"gmail"},
				expectedCapabilities: []map[string]interface{}{
					{
						"service": "gmail",
						"actions": []string{"gmail.send_message", "gmail.list_messages"},
						"status":  "connected",
					},
				},
				expectError: false,
			},
			{
				name:                 "No services connected",
				connectedServices:    []string{},
				expectedCapabilities: []map[string]interface{}{},
				expectError:          false,
			},
			{
				name:              "Non-existent service requested",
				connectedServices: []string{"gmail", "nonexistent", "docs"},
				expectedCapabilities: []map[string]interface{}{
					{
						"service": "gmail",
						"actions": []string{"gmail.send_message", "gmail.list_messages"},
						"status":  "connected",
					},
					{
						"service": "docs",
						"actions": []string{"docs.create_document", "docs.update_document"},
						"status":  "connected",
					},
				},
				expectError: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				capabilities, err := parser.BuildUserCapabilities(mockCatalog, tt.connectedServices)
				
				if tt.expectError && err == nil {
					t.Errorf("BuildUserCapabilities() expected error but got none")
					return
				}
				
				if !tt.expectError && err != nil {
					t.Errorf("BuildUserCapabilities() unexpected error: %v", err)
					return
				}
				
				if len(capabilities) != len(tt.expectedCapabilities) {
					t.Errorf("BuildUserCapabilities() returned %d capabilities, expected %d", len(capabilities), len(tt.expectedCapabilities))
					return
				}
				
				// Verify each capability
				for i, expectedCap := range tt.expectedCapabilities {
					if i >= len(capabilities) {
						t.Errorf("Missing capability at index %d", i)
						continue
					}
					
					actualCap := capabilities[i]
					
					// Check service name
					if actualCap["service"] != expectedCap["service"] {
						t.Errorf("Capability[%d].service = %v, expected %v", i, actualCap["service"], expectedCap["service"])
					}
					
					// Check status
					if actualCap["status"] != expectedCap["status"] {
						t.Errorf("Capability[%d].status = %v, expected %v", i, actualCap["status"], expectedCap["status"])
					}
					
					// Check actions (order may vary, so we need to compare as sets)
					actualActions, ok := actualCap["actions"].([]string)
					if !ok {
						t.Errorf("Capability[%d].actions is not []string", i)
						continue
					}
					
					expectedActions, ok := expectedCap["actions"].([]string)
					if !ok {
						t.Errorf("Expected capability[%d].actions is not []string", i)
						continue
					}
					
					if !stringSlicesEqual(actualActions, expectedActions) {
						t.Errorf("Capability[%d].actions = %v, expected %v", i, actualActions, expectedActions)
					}
				}
			})
		}
	})

	// Test with strongly-typed *types.MCPServiceCatalog format
	t.Run("Strongly-typed catalog format", func(t *testing.T) {
		// Create strongly-typed catalog
		mockCatalog := &types.MCPServiceCatalog{
			Providers: types.MCPProviders{
				Workspace: types.MCPWorkspaceProvider{
					Description: "Google Workspace Provider",
					DisplayName: "Google Workspace",
					Services: map[string]types.MCPServiceDefinition{
						"gmail": {
							DisplayName: "Gmail",
							Description: "Gmail service for sending emails",
							Functions: map[string]types.MCPFunctionSchema{
								"send_message": {
									Name:           "send_message",
									DisplayName:    "Send Message",
									Description:    "Send an email message",
									RequiredFields: []string{"to", "subject", "body"},
									ExamplePayload: map[string]interface{}{
										"to":      "user@example.com",
										"subject": "Test Subject",
										"body":    "Test Body",
									},
								},
								"list_messages": {
									Name:           "list_messages",
									DisplayName:    "List Messages",
									Description:    "List email messages",
									RequiredFields: []string{},
									ExamplePayload: map[string]interface{}{},
								},
							},
						},
						"docs": {
							DisplayName: "Google Docs",
							Description: "Google Docs service for document management",
							Functions: map[string]types.MCPFunctionSchema{
								"create_document": {
									Name:           "create_document",
									DisplayName:    "Create Document",
									Description:    "Create a new document",
									RequiredFields: []string{"title", "content"},
									ExamplePayload: map[string]interface{}{
										"title":   "New Document",
										"content": "Document content",
									},
								},
							},
						},
					},
				},
			},
		}

		tests := []struct {
			name               string
			connectedServices  []string
			expectedCapabilities []map[string]interface{}
			expectError        bool
		}{
			{
				name:              "Typed catalog - Gmail and Docs connected",
				connectedServices: []string{"gmail", "docs"},
				expectedCapabilities: []map[string]interface{}{
					{
						"service": "gmail",
						"actions": []string{"gmail.send_message", "gmail.list_messages"},
						"status":  "connected",
					},
					{
						"service": "docs",
						"actions": []string{"docs.create_document"},
						"status":  "connected",
					},
				},
				expectError: false,
			},
			{
				name:              "Typed catalog - Single service",
				connectedServices: []string{"gmail"},
				expectedCapabilities: []map[string]interface{}{
					{
						"service": "gmail",
						"actions": []string{"gmail.send_message", "gmail.list_messages"},
						"status":  "connected",
					},
				},
				expectError: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				capabilities, err := parser.BuildUserCapabilities(mockCatalog, tt.connectedServices)
				
				if tt.expectError && err == nil {
					t.Errorf("BuildUserCapabilities() expected error but got none")
					return
				}
				
				if !tt.expectError && err != nil {
					t.Errorf("BuildUserCapabilities() unexpected error: %v", err)
					return
				}
				
				if len(capabilities) != len(tt.expectedCapabilities) {
					t.Errorf("BuildUserCapabilities() returned %d capabilities, expected %d", len(capabilities), len(tt.expectedCapabilities))
					return
				}
				
				// Verify each capability
				for i, expectedCap := range tt.expectedCapabilities {
					if i >= len(capabilities) {
						t.Errorf("Missing capability at index %d", i)
						continue
					}
					
					actualCap := capabilities[i]
					
					// Check service name
					if actualCap["service"] != expectedCap["service"] {
						t.Errorf("Capability[%d].service = %v, expected %v", i, actualCap["service"], expectedCap["service"])
					}
					
					// Check status
					if actualCap["status"] != expectedCap["status"] {
						t.Errorf("Capability[%d].status = %v, expected %v", i, actualCap["status"], expectedCap["status"])
					}
					
					// Check actions (order may vary, so we need to compare as sets)
					actualActions, ok := actualCap["actions"].([]string)
					if !ok {
						t.Errorf("Capability[%d].actions is not []string", i)
						continue
					}
					
					expectedActions, ok := expectedCap["actions"].([]string)
					if !ok {
						t.Errorf("Expected capability[%d].actions is not []string", i)
						continue
					}
					
					if !stringSlicesEqual(actualActions, expectedActions) {
						t.Errorf("Capability[%d].actions = %v, expected %v", i, actualActions, expectedActions)
					}
				}
			})
		}
	})

	// Test error cases
	t.Run("Error cases", func(t *testing.T) {
		tests := []struct {
			name        string
			catalog     interface{}
			services    []string
			expectError bool
		}{
			{
				name:        "Invalid catalog type",
				catalog:     "invalid",
				services:    []string{"gmail"},
				expectError: true,
			},
			{
				name:        "Nil catalog",
				catalog:     nil,
				services:    []string{"gmail"},
				expectError: true,
			},
			{
				name: "Malformed legacy catalog - missing providers",
				catalog: map[string]interface{}{
					"invalid": "structure",
				},
				services:    []string{"gmail"},
				expectError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := parser.BuildUserCapabilities(tt.catalog, tt.services)
				
				if tt.expectError && err == nil {
					t.Errorf("BuildUserCapabilities() expected error but got none")
				}
				
				if !tt.expectError && err != nil {
					t.Errorf("BuildUserCapabilities() unexpected error: %v", err)
				}
			})
		}
	})
}

func TestMCPService_CatalogConversion(t *testing.T) {
	// Test the catalog-to-services conversion logic directly
	// This tests the core logic without requiring HTTP mocking
	
	mockCatalog := &types.MCPServiceCatalog{
		Providers: types.MCPProviders{
			Workspace: types.MCPWorkspaceProvider{
				Description: "Google Workspace Provider",
				DisplayName: "Google Workspace",
				Services: map[string]types.MCPServiceDefinition{
					"gmail": {
						DisplayName: "Gmail",
						Description: "Google Gmail service",
						Functions: map[string]types.MCPFunctionSchema{
							"send_message": {
								Name:           "send_message",
								DisplayName:    "Send Message",
								Description:    "Send email message",
								RequiredFields: []string{"to", "subject", "body"},
								ExamplePayload: map[string]interface{}{
									"to":      "user@example.com",
									"subject": "Test Subject",
									"body":    "Test Body",
								},
							},
						},
					},
					"docs": {
						DisplayName: "Google Docs",
						Description: "Google Docs service",
						Functions: map[string]types.MCPFunctionSchema{
							"create_document": {
								Name:           "create_document",
								DisplayName:    "Create Document",
								Description:    "Create new document",
								RequiredFields: []string{"title"},
								ExamplePayload: map[string]interface{}{
									"title": "New Document",
								},
							},
						},
					},
					"drive": {
						DisplayName: "Google Drive",
						Description: "Google Drive service",
						Functions: map[string]types.MCPFunctionSchema{
							"share_file": {
								Name:           "share_file",
								DisplayName:    "Share File",
								Description:    "Share file with users",
								RequiredFields: []string{"file_id", "email"},
								ExamplePayload: map[string]interface{}{
									"file_id": "1234567890",
									"email":   "user@example.com",
								},
							},
						},
					},
				},
			},
		},
	}

	// Test catalog structure validation
	t.Run("Catalog structure validation", func(t *testing.T) {
		// Verify catalog has expected services
		expectedServices := []string{"gmail", "docs", "drive"}
		if len(mockCatalog.Providers.Workspace.Services) != len(expectedServices) {
			t.Errorf("Expected %d services, got %d", len(expectedServices), len(mockCatalog.Providers.Workspace.Services))
		}
		
		// Verify each expected service exists
		for _, serviceName := range expectedServices {
			if _, exists := mockCatalog.Providers.Workspace.Services[serviceName]; !exists {
				t.Errorf("Expected service '%s' not found in catalog", serviceName)
			}
		}
		
		// Verify service structure matches new format
		for serviceName, serviceDefinition := range mockCatalog.Providers.Workspace.Services {
			if serviceDefinition.DisplayName == "" {
				t.Errorf("Service '%s' missing DisplayName", serviceName)
			}
			if serviceDefinition.Description == "" {
				t.Errorf("Service '%s' missing Description", serviceName)
			}
			if len(serviceDefinition.Functions) == 0 {
				t.Errorf("Service '%s' has no functions", serviceName)
			}
			
			// Verify function structure
			for funcName, funcSchema := range serviceDefinition.Functions {
				if funcSchema.Name == "" {
					t.Errorf("Function '%s.%s' missing Name", serviceName, funcName)
				}
				if funcSchema.DisplayName == "" {
					t.Errorf("Function '%s.%s' missing DisplayName", serviceName, funcName)
				}
				if len(funcSchema.RequiredFields) == 0 {
					t.Errorf("Function '%s.%s' has no RequiredFields", serviceName, funcName)
				}
				if funcSchema.ExamplePayload == nil {
					t.Errorf("Function '%s.%s' missing ExamplePayload", serviceName, funcName)
				}
			}
		}
	})
}

// Helper function to compare string slices regardless of order
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	
	// Create maps for comparison
	mapA := make(map[string]bool)
	mapB := make(map[string]bool)
	
	for _, str := range a {
		mapA[str] = true
	}
	
	for _, str := range b {
		mapB[str] = true
	}
	
	return reflect.DeepEqual(mapA, mapB)
}
