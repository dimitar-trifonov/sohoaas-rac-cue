package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockMCPServer provides a simple HTTP mock server for MCP integration testing
type MockMCPServer struct {
	server    *httptest.Server
	responses map[string]*ExecuteActionResponse
	catalog   map[string]interface{}
}

// NewMockMCPServer creates a new mock MCP server for testing
func NewMockMCPServer(t *testing.T) *MockMCPServer {
	mock := &MockMCPServer{
		responses: make(map[string]*ExecuteActionResponse),
		catalog:   buildMockServiceCatalog(),
	}
	
	// Create HTTP server with mock handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/api/services", mock.handleServiceCatalog)
	mux.HandleFunc("/api/v1/mcp/execute", mock.handleExecuteAction)
	
	mock.server = httptest.NewServer(mux)
	t.Logf("Mock MCP Server started at: %s", mock.server.URL)
	
	return mock
}

// Close shuts down the mock server
func (m *MockMCPServer) Close() {
	m.server.Close()
}

// URL returns the mock server URL
func (m *MockMCPServer) URL() string {
	return m.server.URL
}

// SetResponse configures a mock response for a specific service/action combination
func (m *MockMCPServer) SetResponse(service, action string, response *ExecuteActionResponse) {
	key := fmt.Sprintf("%s.%s", service, action)
	m.responses[key] = response
}

// SetDefaultGoogleWorkspaceResponses configures typical Google Workspace responses
func (m *MockMCPServer) SetDefaultGoogleWorkspaceResponses() {
	// Gmail responses
	m.SetResponse("gmail", "send_email", &ExecuteActionResponse{
		Success: true,
		Data: map[string]interface{}{
			"message_id": "mock_message_123",
			"status":     "sent",
		},
	})
	
	m.SetResponse("gmail", "list_messages", &ExecuteActionResponse{
		Success: true,
		Data: map[string]interface{}{
			"messages": []map[string]interface{}{
				{
					"id":      "msg_001",
					"subject": "Mock Email Subject",
					"from":    "test@example.com",
				},
			},
			"total_count": 1,
		},
	})
	
	// Google Docs responses
	m.SetResponse("docs", "create_document", &ExecuteActionResponse{
		Success: true,
		Data: map[string]interface{}{
			"document_id":  "mock_doc_456",
			"document_url": "https://docs.google.com/document/d/mock_doc_456/edit",
			"title":        "Mock Document",
		},
	})
	
	m.SetResponse("docs", "append_text", &ExecuteActionResponse{
		Success: true,
		Data: map[string]interface{}{
			"document_id": "mock_doc_456",
			"updated":     true,
		},
	})
	
	// Google Drive responses
	m.SetResponse("drive", "create_folder", &ExecuteActionResponse{
		Success: true,
		Data: map[string]interface{}{
			"folder_id":  "mock_folder_789",
			"folder_url": "https://drive.google.com/drive/folders/mock_folder_789",
			"name":       "Mock Folder",
		},
	})
	
	m.SetResponse("drive", "move_file", &ExecuteActionResponse{
		Success: true,
		Data: map[string]interface{}{
			"file_id":   "mock_doc_456",
			"folder_id": "mock_folder_789",
			"moved":     true,
		},
	})
	
	// Google Calendar responses
	m.SetResponse("calendar", "create_event", &ExecuteActionResponse{
		Success: true,
		Data: map[string]interface{}{
			"event_id":  "mock_event_101",
			"event_url": "https://calendar.google.com/calendar/event?eid=mock_event_101",
			"title":     "Mock Calendar Event",
		},
	})
}

// handleServiceCatalog handles GET /api/services requests
func (m *MockMCPServer) handleServiceCatalog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m.catalog)
}

// handleExecuteAction handles POST /api/v1/mcp/execute requests
func (m *MockMCPServer) handleExecuteAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Parse request
	var request ExecuteActionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// Validate OAuth token (basic check for PoC)
	if request.OAuthToken == "" || request.OAuthToken == "invalid_token" {
		response := &ExecuteActionResponse{
			Success: false,
			Error:   "Invalid or missing OAuth token",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}
	
	// Find mock response
	key := fmt.Sprintf("%s.%s", request.Service, request.Action)
	response, exists := m.responses[key]
	if !exists {
		response = &ExecuteActionResponse{
			Success: false,
			Error:   fmt.Sprintf("Mock response not configured for %s.%s", request.Service, request.Action),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		json.NewEncoder(w).Encode(response)
		return
	}
	
	// Return mock response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// buildMockServiceCatalog creates a mock service catalog matching real MCP structure
func buildMockServiceCatalog() map[string]interface{} {
	return map[string]interface{}{
		"providers": map[string]interface{}{
			"workspace": map[string]interface{}{
				"services": map[string]interface{}{
					"gmail": map[string]interface{}{
						"name":        "Gmail",
						"description": "Google Gmail service",
						"functions": map[string]interface{}{
							"send_email": map[string]interface{}{
								"description": "Send an email",
								"parameters":  map[string]interface{}{},
							},
							"list_messages": map[string]interface{}{
								"description": "List email messages",
								"parameters":  map[string]interface{}{},
							},
							"get_message": map[string]interface{}{
								"description": "Get a specific email message",
								"parameters":  map[string]interface{}{},
							},
							"search_messages": map[string]interface{}{
								"description": "Search email messages",
								"parameters":  map[string]interface{}{},
							},
						},
						"status": "connected",
					},
					"docs": map[string]interface{}{
						"name":        "Google Docs",
						"description": "Google Docs service",
						"functions": map[string]interface{}{
							"create_document": map[string]interface{}{
								"description": "Create a new document",
								"parameters":  map[string]interface{}{},
							},
							"get_document": map[string]interface{}{
								"description": "Get document content",
								"parameters":  map[string]interface{}{},
							},
							"append_text": map[string]interface{}{
								"description": "Append text to document",
								"parameters":  map[string]interface{}{},
							},
							"update_document": map[string]interface{}{
								"description": "Update document content",
								"parameters":  map[string]interface{}{},
							},
						},
						"status": "connected",
					},
					"drive": map[string]interface{}{
						"name":        "Google Drive",
						"description": "Google Drive service",
						"functions": map[string]interface{}{
							"create_folder": map[string]interface{}{
								"description": "Create a new folder",
								"parameters":  map[string]interface{}{},
							},
							"list_files": map[string]interface{}{
								"description": "List files in Drive",
								"parameters":  map[string]interface{}{},
							},
							"get_file": map[string]interface{}{
								"description": "Get file information",
								"parameters":  map[string]interface{}{},
							},
							"move_file": map[string]interface{}{
								"description": "Move file to folder",
								"parameters":  map[string]interface{}{},
							},
							"share_file": map[string]interface{}{
								"description": "Share file with users",
								"parameters":  map[string]interface{}{},
							},
						},
						"status": "connected",
					},
					"calendar": map[string]interface{}{
						"name":        "Google Calendar",
						"description": "Google Calendar service",
						"functions": map[string]interface{}{
							"create_event": map[string]interface{}{
								"description": "Create a calendar event",
								"parameters":  map[string]interface{}{},
							},
							"list_events": map[string]interface{}{
								"description": "List calendar events",
								"parameters":  map[string]interface{}{},
							},
							"get_event": map[string]interface{}{
								"description": "Get event details",
								"parameters":  map[string]interface{}{},
							},
							"update_event": map[string]interface{}{
								"description": "Update calendar event",
								"parameters":  map[string]interface{}{},
							},
							"delete_event": map[string]interface{}{
								"description": "Delete calendar event",
								"parameters":  map[string]interface{}{},
							},
						},
						"status": "connected",
					},
					"sheets": map[string]interface{}{
						"name":        "Google Sheets",
						"description": "Google Sheets service",
						"functions": map[string]interface{}{
							"create_spreadsheet": map[string]interface{}{
								"description": "Create a new spreadsheet",
								"parameters":  map[string]interface{}{},
							},
							"get_values": map[string]interface{}{
								"description": "Get cell values",
								"parameters":  map[string]interface{}{},
							},
							"update_values": map[string]interface{}{
								"description": "Update cell values",
								"parameters":  map[string]interface{}{},
							},
							"append_values": map[string]interface{}{
								"description": "Append values to sheet",
								"parameters":  map[string]interface{}{},
							},
						},
						"status": "connected",
					},
				},
			},
		},
	}
}
