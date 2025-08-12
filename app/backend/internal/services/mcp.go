package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
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
	mcpServices, err := m.GetServiceCatalog()
	if err != nil {
		log.Printf("[MCPService] ERROR: Failed to get service catalog for user %s: %v", userID, err)
		return nil, fmt.Errorf("failed to get service catalog: %w", err)
	}
	
	// Parse the MCP catalog structure: providers → workspace → services
	providersWrapper, ok := mcpServices["providers"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid MCP catalog structure: missing providers")
	}
	
	workspaceWrapper, ok := providersWrapper["workspace"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid MCP catalog structure: missing workspace")
	}
	
	servicesMap, ok := workspaceWrapper["services"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid MCP catalog structure: missing services")
	}
	
	// Convert to MCPService slice - all services available for user
	var userServices []types.MCPService
	for serviceName := range servicesMap {
		userServices = append(userServices, types.MCPService{
			Service:   serviceName,
			Functions: []types.MCPFunction{}, // PoC: empty functions for now
			Status:    "connected",           // PoC: assume all services are connected
			Metadata: map[string]interface{}{
				"enabled": true,
			},
		})
	}
	
	return userServices, nil
}

// GetServiceCatalog retrieves the service catalog from MCP service
func (m *MCPService) GetServiceCatalog() (map[string]interface{}, error) {
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
	
	var mcpServices map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&mcpServices); err != nil {
		log.Printf("[MCPService] ERROR: Failed to decode MCP response: %v", err)
		return nil, fmt.Errorf("failed to decode MCP service catalog: %w", err)
	}
	
	log.Printf("[MCPService] SUCCESS: Retrieved MCP catalog with %d top-level keys", len(mcpServices))
	return mcpServices, nil
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
	url := m.baseURL + "/api/v1/mcp/execute"
	
	request := ExecuteActionRequest{
		Service:    service,
		Action:     action,
		Parameters: parameters,
		OAuthToken: oauthToken,
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
	log.Printf("[MCPService] Request JSON: %s", string(requestBody))
	
	// Create HTTP request
	req, err := http.NewRequest("POST", url, strings.NewReader(string(requestBody)))
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
	log.Printf("[MCPService] Response body: %s", string(responseBody))
	
	// Parse response
	var executeResponse ExecuteActionResponse
	if err := json.Unmarshal(responseBody, &executeResponse); err != nil {
		log.Printf("[MCPService] ERROR: Failed to decode MCP execute response: %v", err)
		log.Printf("[MCPService] Raw response: %s", string(responseBody))
		return nil, fmt.Errorf("failed to decode MCP execute response: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		log.Printf("[MCPService] ERROR: MCP execute failed with status %d: %s", resp.StatusCode, executeResponse.Error)
		return &executeResponse, fmt.Errorf("MCP execute failed with status %d: %s", resp.StatusCode, executeResponse.Error)
	}
	
	if !executeResponse.Success {
		log.Printf("[MCPService] ERROR: MCP execute returned success=false: %s", executeResponse.Error)
		return &executeResponse, fmt.Errorf("MCP execute failed: %s", executeResponse.Error)
	}
	
	log.Printf("[MCPService] SUCCESS: MCP action executed successfully")
	return &executeResponse, nil
}
