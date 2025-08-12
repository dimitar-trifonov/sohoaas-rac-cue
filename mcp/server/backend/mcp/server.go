package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/dimitar-trifonov/sohoaas/service-proxies/providers/workspace"
	"github.com/dimitar-trifonov/sohoaas/service-proxies/workflow"
)

// MCPServer represents the MCP server implementation
type MCPServer struct {
	workspaceManager *workspace.ProxyManager
	workflowEngine   *workflow.MultiProviderWorkflowEngine
	upgrader         websocket.Upgrader
	connections      map[string]*websocket.Conn
	connMutex        sync.RWMutex
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer(workspaceManager *workspace.ProxyManager, workflowEngine *workflow.MultiProviderWorkflowEngine) *MCPServer {
	return &MCPServer{
		workspaceManager: workspaceManager,
		workflowEngine:   workflowEngine,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow connections from any origin for development
				// In production, implement proper origin checking
				return true
			},
		},
		connections: make(map[string]*websocket.Conn),
	}
}

// HandleWebSocket handles WebSocket connections for MCP
func (s *MCPServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Printf("MCP WebSocket connection attempt from %s", r.RemoteAddr)
	log.Printf("Headers: Upgrade=%s, Connection=%s", r.Header.Get("Upgrade"), r.Header.Get("Connection"))
	
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Generate connection ID
	connID := fmt.Sprintf("conn_%d", len(s.connections))
	s.connMutex.Lock()
	s.connections[connID] = conn
	s.connMutex.Unlock()

	defer func() {
		s.connMutex.Lock()
		delete(s.connections, connID)
		s.connMutex.Unlock()
	}()

	log.Printf("MCP client connected: %s", connID)

	// Handle messages
	for {
		var request JSONRPCRequest
		err := conn.ReadJSON(&request)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		response := s.handleRequest(request)
		
		err = conn.WriteJSON(response)
		if err != nil {
			log.Printf("Error writing response: %v", err)
			break
		}
	}
}

// handleRequest processes incoming JSON-RPC requests
func (s *MCPServer) handleRequest(request JSONRPCRequest) JSONRPCResponse {
	switch request.Method {
	case "initialize":
		return s.handleInitialize(request)
	case "resources/list":
		return s.handleListResources(request)
	case "resources/read":
		return s.handleReadResource(request)
	case "tools/list":
		return s.handleListTools(request)
	case "tools/call":
		return s.handleCallTool(request)
	default:
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &RPCError{
				Code:    -32601,
				Message: "Method not found",
				Data:    fmt.Sprintf("Unknown method: %s", request.Method),
			},
		}
	}
}

// handleInitialize handles the MCP initialize request
func (s *MCPServer) handleInitialize(request JSONRPCRequest) JSONRPCResponse {
	var initReq InitializeRequest
	if err := json.Unmarshal(request.Params, &initReq); err != nil {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &RPCError{
				Code:    -32602,
				Message: "Invalid params",
				Data:    err.Error(),
			},
		}
	}

	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: ServerCapabilities{
			Resources: &ResourceCapability{
				Subscribe:   false,
				ListChanged: false,
			},
			Tools: &ToolCapability{
				ListChanged: false,
			},
		},
		ServerInfo: ServerInfo{
			Name:    "Workspace MCP Server",
			Version: "1.0.0",
		},
	}

	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}
}

// handleListResources handles the resources/list request
func (s *MCPServer) handleListResources(request JSONRPCRequest) JSONRPCResponse {
	resources := []Resource{
		{
			URI:         "workspace://gmail/functions",
			Name:        "Gmail Functions",
			Description: "Available Gmail operations",
			MimeType:    "application/json",
		},
		{
			URI:         "workspace://docs/functions",
			Name:        "Google Docs Functions",
			Description: "Available Google Docs operations",
			MimeType:    "application/json",
		},
		{
			URI:         "workspace://drive/functions",
			Name:        "Google Drive Functions",
			Description: "Available Google Drive operations",
			MimeType:    "application/json",
		},
		{
			URI:         "workspace://calendar/functions",
			Name:        "Google Calendar Functions",
			Description: "Available Google Calendar operations",
			MimeType:    "application/json",
		},
		{
			URI:         "workspace://workflow/status",
			Name:        "Workflow Status",
			Description: "Current workflow execution status",
			MimeType:    "application/json",
		},
	}

	result := ListResourcesResult{
		Resources: resources,
	}

	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}
}

// handleReadResource handles the resources/read request
func (s *MCPServer) handleReadResource(request JSONRPCRequest) JSONRPCResponse {
	var readReq ReadResourceRequest
	if err := json.Unmarshal(request.Params, &readReq); err != nil {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &RPCError{
				Code:    -32602,
				Message: "Invalid params",
				Data:    err.Error(),
			},
		}
	}

	content, err := s.getResourceContent(readReq.URI)
	if err != nil {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &RPCError{
				Code:    -32603,
				Message: "Internal error",
				Data:    err.Error(),
			},
		}
	}

	result := ReadResourceResult{
		Contents: []ResourceContent{content},
	}

	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}
}

// handleListTools handles the tools/list request
func (s *MCPServer) handleListTools(request JSONRPCRequest) JSONRPCResponse {
	tools := s.getAvailableTools()

	result := ListToolsResult{
		Tools: tools,
	}

	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}
}

// handleCallTool handles the tools/call request
func (s *MCPServer) handleCallTool(request JSONRPCRequest) JSONRPCResponse {
	var callReq CallToolRequest
	if err := json.Unmarshal(request.Params, &callReq); err != nil {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &RPCError{
				Code:    -32602,
				Message: "Invalid params",
				Data:    err.Error(),
			},
		}
	}

	result, err := s.executeTool(callReq.Name, callReq.Arguments)
	if err != nil {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &RPCError{
				Code:    -32603,
				Message: "Tool execution failed",
				Data:    err.Error(),
			},
		}
	}

	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}
}

// getResourceContent retrieves content for a specific resource URI
func (s *MCPServer) getResourceContent(uri string) (ResourceContent, error) {
	switch uri {
	case "workspace://gmail/functions":
		functions := s.getServiceFunctions("gmail")
		data, _ := json.Marshal(functions)
		return ResourceContent{
			URI:      uri,
			MimeType: "application/json",
			Text:     string(data),
		}, nil
	case "workspace://docs/functions":
		functions := s.getServiceFunctions("docs")
		data, _ := json.Marshal(functions)
		return ResourceContent{
			URI:      uri,
			MimeType: "application/json",
			Text:     string(data),
		}, nil
	case "workspace://drive/functions":
		functions := s.getServiceFunctions("drive")
		data, _ := json.Marshal(functions)
		return ResourceContent{
			URI:      uri,
			MimeType: "application/json",
			Text:     string(data),
		}, nil
	case "workspace://calendar/functions":
		functions := s.getServiceFunctions("calendar")
		data, _ := json.Marshal(functions)
		return ResourceContent{
			URI:      uri,
			MimeType: "application/json",
			Text:     string(data),
		}, nil
	case "workspace://workflow/status":
		status := map[string]interface{}{
			"engine_status": "ready",
			"active_workflows": 0,
			"supported_providers": []string{"google_workspace"},
		}
		data, _ := json.Marshal(status)
		return ResourceContent{
			URI:      uri,
			MimeType: "application/json",
			Text:     string(data),
		}, nil
	default:
		return ResourceContent{}, fmt.Errorf("resource not found: %s", uri)
	}
}

// getServiceFunctions returns available functions for a service
func (s *MCPServer) getServiceFunctions(serviceType string) []map[string]interface{} {
	// Return predefined function metadata for each service
	switch serviceType {
	case "gmail":
		return []map[string]interface{}{
			{
				"name": "send_email",
				"description": "Send an email via Gmail",
				"required_fields": []string{"to", "subject", "body"},
			},
			{
				"name": "send_followup",
				"description": "Send a follow-up email",
				"required_fields": []string{"to", "subject", "body"},
			},
			{
				"name": "track_response",
				"description": "Track email responses",
				"required_fields": []string{"message_id"},
			},
		}
	case "docs":
		return []map[string]interface{}{
			{
				"name": "create_from_template",
				"description": "Create document from template",
				"required_fields": []string{"template_id", "title"},
			},
			{
				"name": "create_contract",
				"description": "Create a contract document",
				"required_fields": []string{"template_id", "title"},
			},
			{
				"name": "populate_data",
				"description": "Populate document with data",
				"required_fields": []string{"document_id", "data"},
			},
		}
	case "drive":
		return []map[string]interface{}{
			{
				"name": "share_document",
				"description": "Share a document",
				"required_fields": []string{"file_id", "email", "role"},
			},
			{
				"name": "organize_files",
				"description": "Organize files in folders",
				"required_fields": []string{"file_id", "folder_id"},
			},
			{
				"name": "track_document_status",
				"description": "Track document status",
				"required_fields": []string{"file_id"},
			},
		}
	case "calendar":
		return []map[string]interface{}{
			{
				"name": "create_reminder",
				"description": "Create a calendar reminder",
				"required_fields": []string{"title", "start_time", "end_time"},
			},
			{
				"name": "create_event",
				"description": "Create a calendar event",
				"required_fields": []string{"title", "start_time", "end_time"},
			},
		}
	default:
		return []map[string]interface{}{}
	}
}

// getAvailableTools returns all available MCP tools
func (s *MCPServer) getAvailableTools() []Tool {
	tools := []Tool{
		{
			Name:        "gmail.send_email",
			Description: "Send an email via Gmail",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"token": map[string]interface{}{
						"type":        "string",
						"description": "OAuth2 access token",
					},
					"to": map[string]interface{}{
						"type":        "string",
						"description": "Recipient email address",
					},
					"subject": map[string]interface{}{
						"type":        "string",
						"description": "Email subject",
					},
					"body": map[string]interface{}{
						"type":        "string",
						"description": "Email body content",
					},
				},
				"required": []string{"token", "to", "subject", "body"},
			},
		},
		{
			Name:        "docs.create_from_template",
			Description: "Create a Google Doc from a template",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"token": map[string]interface{}{
						"type":        "string",
						"description": "OAuth2 access token",
					},
					"template_id": map[string]interface{}{
						"type":        "string",
						"description": "Template document ID",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "New document title",
					},
					"replacements": map[string]interface{}{
						"type":        "object",
						"description": "Key-value pairs for template replacement",
					},
				},
				"required": []string{"token", "template_id", "title"},
			},
		},
		{
			Name:        "drive.share_document",
			Description: "Share a Google Drive document",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"token": map[string]interface{}{
						"type":        "string",
						"description": "OAuth2 access token",
					},
					"file_id": map[string]interface{}{
						"type":        "string",
						"description": "File ID to share",
					},
					"email": map[string]interface{}{
						"type":        "string",
						"description": "Email address to share with",
					},
					"role": map[string]interface{}{
						"type":        "string",
						"description": "Permission role (reader, writer, commenter)",
						"enum":        []string{"reader", "writer", "commenter"},
					},
				},
				"required": []string{"token", "file_id", "email", "role"},
			},
		},
		{
			Name:        "calendar.create_reminder",
			Description: "Create a calendar reminder",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"token": map[string]interface{}{
						"type":        "string",
						"description": "OAuth2 access token",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "Event title",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Event description",
					},
					"start_time": map[string]interface{}{
						"type":        "string",
						"description": "Start time in RFC3339 format",
					},
					"end_time": map[string]interface{}{
						"type":        "string",
						"description": "End time in RFC3339 format",
					},
				},
				"required": []string{"token", "title", "start_time", "end_time"},
			},
		},
	}

	return tools
}

// executeTool executes a specific MCP tool
func (s *MCPServer) executeTool(toolName string, arguments map[string]interface{}) (ToolResult, error) {
	ctx := context.Background()

	// Extract common parameters
	token, ok := arguments["token"].(string)
	if !ok {
		return ToolResult{}, fmt.Errorf("token is required")
	}

	// Parse tool name: service.function
	parts := strings.Split(toolName, ".")
	if len(parts) != 2 {
		return ToolResult{}, fmt.Errorf("invalid tool name format: %s (expected service.function)", toolName)
	}
	
	service := parts[0]
	function := parts[1]
	
	// Use unified workflow execution for all services
	return s.executeToolViaWorkflow(ctx, service, function, token, arguments)
}

// executeToolViaWorkflow executes any tool using the unified workflow engine
func (s *MCPServer) executeToolViaWorkflow(ctx context.Context, service, function, token string, arguments map[string]interface{}) (ToolResult, error) {
	// Create workflow step for the tool execution
	steps := []workflow.WorkflowStep{
		{
			ID:       fmt.Sprintf("%s_%s", service, function),
			Provider: "workspace",
			Service:  service,
			Function: function,
			Payload:  arguments, // Use arguments directly as payload
		},
	}

	input := map[string]interface{}{
		"oauth_token": token,
	}

	// Debug logging
	log.Printf("[MCP] Executing %s.%s via workflow engine", service, function)
	log.Printf("[MCP] Token length: %d", len(token))
	log.Printf("[MCP] Arguments: %+v", arguments)

	// Set provider token (same as REST API)
	s.workflowEngine.SetProviderToken("workspace", token)

	// Execute using the same workflow engine as REST API
	result, err := s.workflowEngine.ExecuteWorkflow(ctx, steps, input)
	if err != nil {
		log.Printf("[MCP] %s.%s workflow ERROR: %v", service, function, err)
		return ToolResult{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Error executing %s.%s: %v", service, function, err),
			}},
			IsError: true,
		}, nil
	}

	// Debug logging for result
	log.Printf("[MCP] %s.%s workflow result: %+v", service, function, result)

	// Extract result from the step
	stepID := fmt.Sprintf("%s_%s", service, function)
	stepResult, exists := result.StepResults[stepID]
	if !exists {
		log.Printf("[MCP] %s.%s step result not found", service, function)
		return ToolResult{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("%s.%s execution failed: step result not found", service, function),
			}},
			IsError: true,
		}, nil
	}

	if !stepResult.Success {
		errorMsg := "Unknown error"
		if stepResult.Error != nil {
			errorMsg = stepResult.Error.Message
			if stepResult.Error.Details != "" {
				errorMsg += ": " + stepResult.Error.Details
			}
		}
		log.Printf("[MCP] %s.%s step FAILED: %s", service, function, errorMsg)
		return ToolResult{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Failed to execute %s.%s: %s", service, function, errorMsg),
			}},
			IsError: true,
		}, nil
	}

	// Format success response based on service type
	var responseText string
	switch service {
	case "gmail":
		if msgID, exists := stepResult.Data["message_id"]; exists {
			responseText = fmt.Sprintf("Email sent successfully. Message ID: %v", msgID)
		} else {
			responseText = "Email operation completed successfully"
		}
	case "docs":
		if docID, exists := stepResult.Data["document_id"]; exists {
			responseText = fmt.Sprintf("Document operation completed successfully. Document ID: %v", docID)
		} else {
			responseText = "Document operation completed successfully"
		}
	case "drive":
		if fileID, exists := stepResult.Data["file_id"]; exists {
			responseText = fmt.Sprintf("Drive operation completed successfully. File ID: %v", fileID)
		} else {
			responseText = "Drive operation completed successfully"
		}
	case "calendar":
		if eventID, exists := stepResult.Data["event_id"]; exists {
			responseText = fmt.Sprintf("Calendar operation completed successfully. Event ID: %v", eventID)
		} else {
			responseText = "Calendar operation completed successfully"
		}
	default:
		responseText = fmt.Sprintf("%s.%s completed successfully", service, function)
	}

	log.Printf("[MCP] %s.%s workflow SUCCESS", service, function)
	return ToolResult{
		Content: []ToolContent{{
			Type: "text",
			Text: responseText,
		}},
		IsError: false,
	}, nil
}

// Tool execution methods

func (s *MCPServer) executeGmailSendEmail(ctx context.Context, token string, args map[string]interface{}) (ToolResult, error) {
	// Use the same workflow execution path as REST API
	steps := []workflow.WorkflowStep{
		{
			ID:       "gmail_send_email",
			Provider: "workspace",
			Service:  "gmail",
			Function: "send_message",
			Payload: map[string]interface{}{
				"to":      args["to"],
				"subject": args["subject"],
				"body":    args["body"],
			},
		},
	}

	input := map[string]interface{}{
		"oauth_token": token,
	}

	// Debug logging
	log.Printf("[MCP] Gmail send_email using workflow execution")
	log.Printf("[MCP] Token length: %d", len(token))

	// Set provider token (same as REST API)
	s.workflowEngine.SetProviderToken("workspace", token)

	// Execute using the same workflow engine as REST API
	result, err := s.workflowEngine.ExecuteWorkflow(ctx, steps, input)
	if err != nil {
		log.Printf("[MCP] Gmail send_email workflow ERROR: %v", err)
		return ToolResult{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Error sending email: %v", err),
			}},
			IsError: true,
		}, nil
	}

	// Debug logging for result
	log.Printf("[MCP] Gmail send_email workflow result: %+v", result)

	// Extract result from the step
	stepResult, exists := result.StepResults["gmail_send_email"]
	if !exists {
		log.Printf("[MCP] Gmail send_email step result not found")
		return ToolResult{
			Content: []ToolContent{{
				Type: "text",
				Text: "Email sending failed: step result not found",
			}},
			IsError: true,
		}, nil
	}

	if !stepResult.Success {
		errorMsg := "Unknown error"
		if stepResult.Error != nil {
			errorMsg = stepResult.Error.Message
			if stepResult.Error.Details != "" {
				errorMsg += ": " + stepResult.Error.Details
			}
		}
		log.Printf("[MCP] Gmail send_email step FAILED: %s", errorMsg)
		return ToolResult{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Failed to send email: %s", errorMsg),
			}},
			IsError: true,
		}, nil
	}

	// Extract message ID from step result data
	messageID := "unknown"
	if stepResult.Data != nil {
		if msgID, exists := stepResult.Data["message_id"]; exists {
			messageID = fmt.Sprintf("%v", msgID)
		}
	}

	log.Printf("[MCP] Gmail send_email workflow SUCCESS: Message ID = %s", messageID)
	return ToolResult{
		Content: []ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("Email sent successfully. Message ID: %s", messageID),
		}},
		IsError: false,
	}, nil
}

func (s *MCPServer) executeDocsCreateFromTemplate(ctx context.Context, token string, args map[string]interface{}) (ToolResult, error) {
	request := &workspace.ProxyRequest{
		Function:    "create_from_template",
		Token:       token,
		ServiceType: "docs",
		Payload: map[string]interface{}{
			"template_id":   args["template_id"],
			"title":         args["title"],
			"replacements":  args["replacements"],
		},
	}

	response, err := s.workspaceManager.Execute(ctx, request)
	if err != nil {
		return ToolResult{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Error creating document: %v", err),
			}},
			IsError: true,
		}, nil
	}

	return ToolResult{
		Content: []ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("Document created successfully. Document ID: %v", response.Data),
		}},
		IsError: false,
	}, nil
}

func (s *MCPServer) executeDriveShareDocument(ctx context.Context, token string, args map[string]interface{}) (ToolResult, error) {
	request := &workspace.ProxyRequest{
		Function:    "share_file",
		Token:       token,
		ServiceType: "drive",
		Payload: map[string]interface{}{
			"file_id": args["file_id"],
			"email":   args["email"],
			"role":    args["role"],
		},
	}

	response, err := s.workspaceManager.Execute(ctx, request)
	if err != nil {
		return ToolResult{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Error sharing document: %v", err),
			}},
			IsError: true,
		}, nil
	}

	return ToolResult{
		Content: []ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("Document shared successfully: %v", response.Data),
		}},
		IsError: false,
	}, nil
}

func (s *MCPServer) executeCalendarCreateReminder(ctx context.Context, token string, args map[string]interface{}) (ToolResult, error) {
	request := &workspace.ProxyRequest{
		Function:    "create_event",
		Token:       token,
		ServiceType: "calendar",
		Payload: map[string]interface{}{
			"title":       args["title"],
			"description": args["description"],
			"start_time":  args["start_time"],
			"end_time":    args["end_time"],
		},
	}

	response, err := s.workspaceManager.Execute(ctx, request)
	if err != nil {
		return ToolResult{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Error creating calendar event: %v", err),
			}},
			IsError: true,
		}, nil
	}

	return ToolResult{
		Content: []ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("Calendar event created successfully: %v", response.Data),
		}},
		IsError: false,
	}, nil
}

// Public REST API methods for external access

// GetAvailableTools returns all available MCP tools (public method)
func (s *MCPServer) GetAvailableTools() []Tool {
	return s.getAvailableTools()
}

// ExecuteTool executes a specific MCP tool (public method)
func (s *MCPServer) ExecuteTool(toolName string, arguments map[string]interface{}) (ToolResult, error) {
	return s.executeTool(toolName, arguments)
}

// GetAvailableResources returns all available MCP resources (public method)
func (s *MCPServer) GetAvailableResources() []Resource {
	return []Resource{
		{
			URI:         "workspace://gmail/functions",
			Name:        "Gmail Functions",
			Description: "Available Gmail operations",
			MimeType:    "application/json",
		},
		{
			URI:         "workspace://docs/functions",
			Name:        "Docs Functions",
			Description: "Available Google Docs operations",
			MimeType:    "application/json",
		},
		{
			URI:         "workspace://drive/functions",
			Name:        "Drive Functions",
			Description: "Available Google Drive operations",
			MimeType:    "application/json",
		},
		{
			URI:         "workspace://calendar/functions",
			Name:        "Calendar Functions",
			Description: "Available Google Calendar operations",
			MimeType:    "application/json",
		},
	}
}

// ReadResource reads a specific resource by URI (public method)
func (s *MCPServer) ReadResource(uri string) (ResourceContent, error) {
	return s.getResourceContent(uri)
}
