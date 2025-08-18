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
	mux.HandleFunc("/api/mcp/tools/call", mock.handleExecuteAction)
	
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
	
	m.SetResponse("gmail", "send_message", &ExecuteActionResponse{
		Success: true,
		Data: map[string]interface{}{
			"message_id": "mock_message_456",
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
	
	// Parse MCP tools/call request format
	var toolsRequest struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.NewDecoder(r.Body).Decode(&toolsRequest); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Extract OAuth token from arguments
	oauthToken, _ := toolsRequest.Arguments["token"].(string)

	// Validate OAuth token (basic check for PoC)
	if oauthToken == "" || oauthToken == "invalid_token" {
		toolsResponse := map[string]interface{}{
			"result": map[string]interface{}{
				"isError": true,
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": "Invalid or missing OAuth token",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(toolsResponse)
		return
	}

	// Find mock response using tool name
	key := toolsRequest.Name
	response, exists := m.responses[key]
	if !exists {
		// Return MCP tools/call error format for unknown actions
		toolsResponse := map[string]interface{}{
			"result": map[string]interface{}{
				"isError": true,
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": fmt.Sprintf("Mock response not configured for %s", key),
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		json.NewEncoder(w).Encode(toolsResponse)
		return
	}
	
	// Convert ExecuteActionResponse to MCP tools/call format with service-specific responses
	var responseText string
	switch key {
	case "gmail.list_messages":
		responseText = `{"messages": [{"id": "msg1", "subject": "Test Email", "from": "test@example.com"}]}`
	case "docs.create_document":
		responseText = `{"document_id": "doc_123", "document_url": "https://docs.google.com/document/d/doc_123"}`
	case "gmail.send_message", "gmail.send_email":
		responseText = `{"message_id": "msg_456", "status": "sent"}`
	case "calendar.create_event":
		responseText = `{"event_id": "evt_789", "event_url": "https://calendar.google.com/event/evt_789"}`
	default:
		responseText = fmt.Sprintf(`{"result": "Mock response for %s"}`, key)
	}

	toolsResponse := map[string]interface{}{
		"result": map[string]interface{}{
			"isError": !response.Success,
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": responseText,
				},
			},
		},
	}
	
	if !response.Success {
		toolsResponse["result"].(map[string]interface{})["content"] = []map[string]interface{}{
			{
				"type": "text",
				"text": response.Error,
			},
		}
	}
	
	// Return the MCP tools/call response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(toolsResponse)
}

// buildMockServiceCatalog creates a minimal mock service catalog for testing
func buildMockServiceCatalog() map[string]interface{} {
	return map[string]interface{}{
		"providers": map[string]interface{}{
			"workspace": map[string]interface{}{
				"description": "Google Workspace services",
				"display_name": "Google Workspace",
				"services": map[string]interface{}{
					"gmail": map[string]interface{}{
						"display_name": "Gmail",
						"description":  "Send, receive, and manage emails using Gmail API",
						"functions": map[string]interface{}{
							"send_message": map[string]interface{}{
								"name":         "send_message",
								"display_name": "Send Email",
								"description":  "Send an email message via Gmail",
								"required_fields": []interface{}{"to", "subject", "body"},
								"example_payload": map[string]interface{}{
									"to":      "recipient@example.com",
									"subject": "Test Email",
									"body":    "This is a test email from SOHOaaS",
								},
							},
							"send_email": map[string]interface{}{
								"name":         "send_email",
								"display_name": "Send Email (Legacy)",
								"description":  "Send an email message via Gmail (legacy action name)",
								"required_fields": []interface{}{"to", "subject", "body"},
								"example_payload": map[string]interface{}{
									"to":      "recipient@example.com",
									"subject": "Test Email",
									"body":    "This is a test email from SOHOaaS",
								},
							},
							"list_messages": map[string]interface{}{
								"name":         "list_messages",
								"display_name": "List Messages",
								"description":  "List Gmail messages with optional query",
								"required_fields": []interface{}{},
								"example_payload": map[string]interface{}{
									"max_results": 10,
									"query": "is:unread",
								},
							},
						},
					},
					"docs": map[string]interface{}{
						"display_name": "Google Docs",
						"description":  "Create and manage Google Docs documents",
						"functions": map[string]interface{}{
							"create_document": map[string]interface{}{
								"name":         "create_document",
								"display_name": "Create Document",
								"description":  "Create a new Google Docs document",
								"required_fields": []interface{}{"title"},
								"example_payload": map[string]interface{}{
									"title": "My Document",
								},
							},
						},
					},
					"calendar": map[string]interface{}{
						"display_name": "Google Calendar",
						"description":  "Create and manage calendar events",
						"functions": map[string]interface{}{
							"create_event": map[string]interface{}{
								"name":         "create_event",
								"display_name": "Create Event",
								"description":  "Create a new calendar event",
								"required_fields": []interface{}{"title", "start_time"},
								"example_payload": map[string]interface{}{
									"title": "Meeting",
									"start_time": "2024-01-01T10:00:00Z",
								},
							},
						},
					},
				},
			},
		},
	}
}
