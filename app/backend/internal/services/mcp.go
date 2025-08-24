package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"sohoaas-backend/internal/types"
)

// MCPService handles communication with the MCP service
type MCPService struct {
	baseURL string
	client  *http.Client
}

// NewMCPService creates a new MCP service instance
func NewMCPService(baseURL string) *MCPService {
	return &MCPService{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetUserServices retrieves all available services for a user (PoC: all services available)
func (m *MCPService) GetUserServices(userID, token string) ([]types.MCPService, error) {
	log.Printf("[MCPService] Getting user services for user: %s", userID)
	// For PoC: Use the working /api/services endpoint and return all services
	// Since all services are available for all users
	catalog, err := m.GetServiceCatalog()
	if err != nil {
		log.Printf("[MCPService] ERROR: Failed to get service catalog for user %s: %v", userID, err)
		return nil, fmt.Errorf("failed to get service catalog: %w", err)
	}
	
	// Convert to MCPService slice - all services available for user
	var userServices []types.MCPService
	for serviceName, serviceDefinition := range catalog.Providers.Workspace.Services {
		// Convert functions from catalog to MCPFunction slice
		var functions []types.MCPFunction
		for functionName, functionSchema := range serviceDefinition.Functions {
			functions = append(functions, types.MCPFunction{
				Name:        functionName,
				Description: functionSchema.Description,
				Parameters:  functionSchema.ExamplePayload, // Use example payload as parameters
				Required:    functionSchema.RequiredFields,
			})
		}
		
		userServices = append(userServices, types.MCPService{
			Service:   serviceName,
			Functions: functions,
			Status:    "connected", // PoC: assume all services are connected
			Metadata: map[string]interface{}{
				"enabled": true,
			},
		})
	}
	
	return userServices, nil
}

// GetServiceCatalog retrieves the service catalog from MCP service
func (m *MCPService) GetServiceCatalog() (*types.MCPServiceCatalog, error) {
	url := m.baseURL + "/api/services"
	log.Printf("[MCPService] === CALLING MCP SERVICE CATALOG ===")
	log.Printf("[MCPService] MCP URL: %s", url)
	
	resp, err := m.client.Get(url)
	if err != nil {
		log.Printf("[MCPService] ERROR: Failed to call MCP service: %v", err)
		return nil, fmt.Errorf("failed to query MCP service catalog: %w", err)
	}
	defer resp.Body.Close()
	
	log.Printf("[MCPService] MCP Response Status: %d", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		log.Printf("[MCPService] ERROR: MCP service returned non-200 status: %d", resp.StatusCode)
		return nil, fmt.Errorf("MCP service catalog returned status %d", resp.StatusCode)
	}
	
	var catalog types.MCPServiceCatalog
	if err := json.NewDecoder(resp.Body).Decode(&catalog); err != nil {
		log.Printf("[MCPService] ERROR: Failed to decode MCP response: %v", err)
		return nil, fmt.Errorf("failed to decode MCP service catalog: %w", err)
	}
	
	log.Printf("[MCPService] SUCCESS: Retrieved MCP catalog with %d services", len(catalog.Providers.Workspace.Services))
	return &catalog, nil
}

// ExecuteActionRequest represents a request to execute an MCP action
type ExecuteActionRequest struct {
	Service    string                 `json:"service"`
	Action     string                 `json:"action"`
	Parameters map[string]interface{} `json:"parameters"`
	OAuthToken string                 `json:"oauth_token"`
}

// ExecuteActionResponse represents the response from MCP action execution
type ExecuteActionResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

// ExecuteAction executes an action via the MCP service
func (m *MCPService) ExecuteAction(service, action string, parameters map[string]interface{}, oauthToken string) (*ExecuteActionResponse, error) {
	url := m.baseURL + "/api/mcp/tools/call"
	
	// Convert to MCP tools/call expected format
	toolName := fmt.Sprintf("%s.%s", service, action)
	arguments := make(map[string]interface{})
	
	// Add OAuth token to arguments
	arguments["token"] = oauthToken
	
	// Add all parameters to arguments
	for key, value := range parameters {
		arguments[key] = value
	}
	
	request := struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}{
		Name:      toolName,
		Arguments: arguments,
	}
	
	log.Printf("[MCPService] === EXECUTING MCP ACTION ===")
	log.Printf("[MCPService] Service: %s, Action: %s", service, action)
	log.Printf("[MCPService] URL: %s", url)
	log.Printf("[MCPService] Parameters: %+v", parameters)
	log.Printf("[MCPService] OAuth token length: %d characters", len(oauthToken))
	
	// Marshal request to JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		log.Printf("[MCPService] ERROR: Failed to marshal request: %v", err)
		return nil, fmt.Errorf("failed to marshal MCP execute request: %w", err)
	}
	
	log.Printf("[MCPService] Request body length: %d bytes", len(requestBody))
	// Redact token value in logged JSON
	redactedArgs := make(map[string]interface{}, len(arguments))
	for k, v := range arguments {
		if k == "token" {
			redactedArgs[k] = "[REDACTED]"
		} else {
			redactedArgs[k] = v
		}
	}
	redactedReq := struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}{
		Name:      toolName,
		Arguments: redactedArgs,
	}
	if redactedJSON, err := json.Marshal(redactedReq); err == nil {
		log.Printf("[MCPService] Request JSON (redacted): %s", string(redactedJSON))
	} else {
		log.Printf("[MCPService] Request JSON (redacted) marshal error: %v", err)
	}
	
	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewReader(requestBody))
	if err != nil {
		log.Printf("[MCPService] ERROR: Failed to create HTTP request: %v", err)
		return nil, fmt.Errorf("failed to create MCP execute request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	log.Printf("[MCPService] Sending HTTP POST request to MCP server...")
	
	// Execute request
	resp, err := m.client.Do(req)
	if err != nil {
		log.Printf("[MCPService] ERROR: Failed to execute MCP action: %v", err)
		return nil, fmt.Errorf("failed to execute MCP action: %w", err)
	}
	defer resp.Body.Close()
	
	log.Printf("[MCPService] MCP Execute Response Status: %d", resp.StatusCode)
	log.Printf("[MCPService] Response headers: %+v", resp.Header)
	
	// Read response body first for logging
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[MCPService] ERROR: Failed to read response body: %v", err)
		return nil, fmt.Errorf("failed to read MCP execute response: %w", err)
	}
	
	log.Printf("[MCPService] Response body length: %d bytes", len(responseBody))
	// Truncate very long responses in logs to keep them readable
	const maxLogBody = 2000
	if len(responseBody) > maxLogBody {
		log.Printf("[MCPService] Response body (truncated): %s... [truncated %d bytes]", string(responseBody[:maxLogBody]), len(responseBody)-maxLogBody)
	} else {
		log.Printf("[MCPService] Response body: %s", string(responseBody))
	}
	
	// Parse response from /api/mcp/tools/call
	var toolResponse struct {
		Result struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
			IsError bool `json:"isError"`
		} `json:"result"`
	}
	
	if err := json.Unmarshal(responseBody, &toolResponse); err != nil {
		log.Printf("[MCPService] ERROR: Failed to decode MCP tools/call response: %v", err)
		log.Printf("[MCPService] Raw response: %s", string(responseBody))
		return nil, fmt.Errorf("failed to decode MCP tools/call response: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		log.Printf("[MCPService] ERROR: MCP tools/call failed with status %d", resp.StatusCode)
		return nil, fmt.Errorf("MCP tools/call failed with status %d", resp.StatusCode)
	}
	
	// Convert tools/call response to ExecuteActionResponse
	executeResponse := &ExecuteActionResponse{
		Success: !toolResponse.Result.IsError,
		Data:    make(map[string]interface{}),
		Error:   "",
	}
	
	// Extract content from response
	if len(toolResponse.Result.Content) > 0 {
		// If there's an error, extract error message
		if toolResponse.Result.IsError {
			executeResponse.Error = toolResponse.Result.Content[0].Text
		} else {
			// For successful responses, try to parse the result as JSON data
			resultText := toolResponse.Result.Content[0].Text
			log.Printf("[MCPService] Parsing result text: %s", resultText)
			var resultData map[string]interface{}
			if err := json.Unmarshal([]byte(resultText), &resultData); err == nil {
				executeResponse.Data = resultData
				log.Printf("[MCPService] Successfully parsed JSON data: %+v", resultData)
			} else {
				log.Printf("[MCPService] Failed to parse as JSON, storing as plain text: %v", err)
				// If not JSON, store as plain text
				executeResponse.Data = map[string]interface{}{
					"result": resultText,
				}
			}
		}
	}
	
	if !executeResponse.Success {
		log.Printf("[MCPService] ERROR: MCP tool execution failed: %s", executeResponse.Error)
		return executeResponse, fmt.Errorf("MCP tool execution failed: %s", executeResponse.Error)
	}
	
	log.Printf("[MCPService] SUCCESS: MCP tool executed successfully")
	return executeResponse, nil
}
