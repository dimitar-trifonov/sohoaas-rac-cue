package workspace

import (
	"context"
	"fmt"
	"sync"

	"github.com/dimitar-trifonov/sohoaas/service-proxies/workflow"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// ProxyManager manages all workspace service proxies
type ProxyManager struct {
	proxies map[string]WorkspaceProxy
	configs map[string]*oauth2.Config
	mutex   sync.RWMutex
}

// ProxyConfig holds configuration for all services
type ProxyConfig struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURL  string   `json:"redirect_url"`
	GmailScopes  []string `json:"gmail_scopes"`
	DocsScopes   []string `json:"docs_scopes"`
	DriveScopes  []string `json:"drive_scopes"`
	CalendarScopes []string `json:"calendar_scopes"`
}

// NewProxyManager creates a new proxy manager
func NewProxyManager(config *ProxyConfig) *ProxyManager {
	manager := &ProxyManager{
		proxies: make(map[string]WorkspaceProxy),
		configs: make(map[string]*oauth2.Config),
	}

	// Initialize OAuth configs for each service
	manager.configs[ServiceTypeGmail] = &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       config.GmailScopes,
		Endpoint:     google.Endpoint,
	}

	manager.configs[ServiceTypeDocs] = &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       config.DocsScopes,
		Endpoint:     google.Endpoint,
	}

	manager.configs[ServiceTypeDrive] = &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       config.DriveScopes,
		Endpoint:     google.Endpoint,
	}

	manager.configs[ServiceTypeCalendar] = &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       config.CalendarScopes,
		Endpoint:     google.Endpoint,
	}

	// Initialize proxy services
	manager.proxies[ServiceTypeGmail] = NewGmailProxy(manager.configs[ServiceTypeGmail])
	manager.proxies[ServiceTypeDocs] = NewDocsProxy(manager.configs[ServiceTypeDocs])
	manager.proxies[ServiceTypeDrive] = NewDriveProxy(manager.configs[ServiceTypeDrive])
	manager.proxies[ServiceTypeCalendar] = NewCalendarProxy(manager.configs[ServiceTypeCalendar])

	return manager
}

// Execute executes a function on the specified service
func (m *ProxyManager) Execute(ctx context.Context, request *ProxyRequest) (*workflow.ProxyResponse, error) {
	m.mutex.RLock()
	proxy, exists := m.proxies[request.ServiceType]
	m.mutex.RUnlock()

	if !exists {
		return &workflow.ProxyResponse{
			Success: false,
			Error: &workflow.ProxyError{
				Code:      string(ErrorCodeServiceUnavailable),
				Message:   fmt.Sprintf("Service not available: %s", request.ServiceType),
				Retryable: false,
			},
		}, nil
	}

	response, err := proxy.Execute(ctx, request.Function, request.Token, request.Payload)
	if err != nil {
		return &workflow.ProxyResponse{
			Success: false,
			Error: &workflow.ProxyError{
				Code:      string(ErrorCodeInternalError),
				Message:   "Proxy execution failed",
				Details:   err.Error(),
				Retryable: true,
			},
		}, nil
	}

	// Note: workflow.ProxyResponse doesn't have RequestID field, so we return response as-is
	return response, nil
}

// ExecuteWorkflowStep executes a workflow step using the proxy system
func (m *ProxyManager) ExecuteWorkflowStep(ctx context.Context, serviceType, function, token string, payload map[string]interface{}) (*workflow.ProxyResponse, error) {
	request := &ProxyRequest{
		Function:    function,
		Token:       token,
		Payload:     payload,
		ServiceType: serviceType,
		RequestID:   fmt.Sprintf("%s_%s_%d", serviceType, function, ctx.Value("timestamp")),
	}

	return m.Execute(ctx, request)
}

// GetSupportedServices returns a list of supported service types
func (m *ProxyManager) GetSupportedServices() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	services := make([]string, 0, len(m.proxies))
	for serviceType := range m.proxies {
		services = append(services, serviceType)
	}
	return services
}

// GetSupportedFunctions returns supported functions for a service
func (m *ProxyManager) GetSupportedFunctions(serviceType string) ([]string, error) {
	m.mutex.RLock()
	proxy, exists := m.proxies[serviceType]
	m.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("service not found: %s", serviceType)
	}

	return proxy.GetSupportedFunctions(), nil
}

// ValidateRequest validates a proxy request
func (m *ProxyManager) ValidateRequest(request *ProxyRequest) error {
	if request.ServiceType == "" {
		return fmt.Errorf("service_type is required")
	}

	if request.Function == "" {
		return fmt.Errorf("function is required")
	}

	if request.Token == "" {
		return fmt.Errorf("token is required")
	}

	if request.Payload == nil {
		return fmt.Errorf("payload is required")
	}

	m.mutex.RLock()
	proxy, exists := m.proxies[request.ServiceType]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("unsupported service type: %s", request.ServiceType)
	}

	// Validate function and payload
	supportedFunctions := proxy.GetSupportedFunctions()
	functionSupported := false
	for _, f := range supportedFunctions {
		if f == request.Function {
			functionSupported = true
			break
		}
	}

	if !functionSupported {
		return fmt.Errorf("unsupported function %s for service %s", request.Function, request.ServiceType)
	}

	return proxy.ValidatePayload(request.Function, request.Payload)
}

// GetServiceCapabilities returns capabilities for all services
func (m *ProxyManager) GetServiceCapabilities() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	capabilities := make(map[string]interface{})
	for serviceType, proxy := range m.proxies {
		capabilities[serviceType] = map[string]interface{}{
			"service_type":         proxy.GetServiceType(),
			"supported_functions":  proxy.GetSupportedFunctions(),
			"total_functions":      len(proxy.GetSupportedFunctions()),
		}
	}

	return capabilities
}

// BatchExecute executes multiple requests in parallel
func (m *ProxyManager) BatchExecute(ctx context.Context, requests []*ProxyRequest) ([]*workflow.ProxyResponse, error) {
	if len(requests) == 0 {
		return []*workflow.ProxyResponse{}, nil
	}

	responses := make([]*workflow.ProxyResponse, len(requests))
	var wg sync.WaitGroup
	var mutex sync.Mutex

	for i, request := range requests {
		wg.Add(1)
		go func(index int, req *ProxyRequest) {
			defer wg.Done()
			
			response, err := m.Execute(ctx, req)
			if err != nil {
				response = &workflow.ProxyResponse{
					Success: false,
					Error: &workflow.ProxyError{
						Code:      string(ErrorCodeInternalError),
						Message:   "Batch execution failed",
						Details:   err.Error(),
						Retryable: true,
					},
				}
			}

			mutex.Lock()
			responses[index] = response
			mutex.Unlock()
		}(i, request)
	}

	wg.Wait()
	return responses, nil
}

// GetOAuthConfig returns the OAuth config for a service
func (m *ProxyManager) GetOAuthConfig(serviceType string) (*oauth2.Config, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	config, exists := m.configs[serviceType]
	if !exists {
		return nil, fmt.Errorf("OAuth config not found for service: %s", serviceType)
	}

	return config, nil
}

// RegisterProxy allows registering custom proxies (for extensibility)
func (m *ProxyManager) RegisterProxy(serviceType string, proxy WorkspaceProxy, config *oauth2.Config) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.proxies[serviceType] = proxy
	m.configs[serviceType] = config
}
