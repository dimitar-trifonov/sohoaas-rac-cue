package workflow

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ServiceProxy represents a generic interface for any service provider proxy
type ServiceProxy interface {
	Execute(ctx context.Context, function string, token string, payload map[string]interface{}) (*ProxyResponse, error)
	GetSupportedFunctions() []string
	GetServiceCapabilities() map[string]interface{}
	ValidateRequest(function string, payload map[string]interface{}) error
}

// ProxyResponse represents a unified response from any service proxy
type ProxyResponse struct {
	Success  bool                   `json:"success"`
	Data     map[string]interface{} `json:"data,omitempty"`
	Error    *ProxyError            `json:"error,omitempty"`
	Metadata *ResponseMetadata      `json:"metadata,omitempty"`
}

type ProxyError struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Details     string `json:"details,omitempty"`
	ServiceType string `json:"service_type"`
	Retryable   bool   `json:"retryable"`
}

type ResponseMetadata struct {
	ExecutionTime time.Duration `json:"execution_time"`
	ServiceType   string        `json:"service_type"`
	Function      string        `json:"function"`
	Timestamp     time.Time     `json:"timestamp"`
}

// RetryPolicy defines how to handle step failures
type RetryPolicy struct {
	MaxRetries    int           `json:"max_retries"`
	RetryDelay    time.Duration `json:"retry_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
}

// WorkflowStep represents a step in a multi-provider workflow
type WorkflowStep struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Provider       string                 `json:"provider"`   // workspace, office365, etc.
	Service        string                 `json:"service"`    // gmail, docs, drive, calendar, outlook, teams, etc.
	Function       string                 `json:"function"`   // Function name to call
	Payload        map[string]interface{} `json:"payload"`    // Function parameters
	DependsOn      []string               `json:"depends_on"` // Step IDs this step depends on
	RetryPolicy    *RetryPolicy           `json:"retry_policy,omitempty"`
	TimeoutSeconds int                    `json:"timeout_seconds,omitempty"`
}

// WorkflowExecution represents the execution state of a workflow
type WorkflowExecution struct {
	ID           string                    `json:"id"`
	Steps        []WorkflowStep            `json:"steps"`
	StepResults  map[string]*ProxyResponse `json:"step_results"`
	Input        map[string]interface{}    `json:"input"`
	Status       string                    `json:"status"`
	StartTime    time.Time                 `json:"start_time"`
	EndTime      *time.Time                `json:"end_time,omitempty"`
	ErrorMessage string                    `json:"error_message,omitempty"`
}

// MultiProviderWorkflowEngine orchestrates workflows across multiple service providers
type MultiProviderWorkflowEngine struct {
	serviceProxies map[string]ServiceProxy // provider_service -> proxy (e.g., "workspace_gmail", "office365_outlook")
	tokens         map[string]string       // provider -> oauth_token (e.g., "workspace" -> token, "office365" -> token)
}

// NewMultiProviderWorkflowEngine creates a new provider-agnostic workflow engine
func NewMultiProviderWorkflowEngine() *MultiProviderWorkflowEngine {
	return &MultiProviderWorkflowEngine{
		serviceProxies: make(map[string]ServiceProxy),
		tokens:         make(map[string]string),
	}
}

// RegisterServiceProxy registers a service proxy for a specific provider and service
func (e *MultiProviderWorkflowEngine) RegisterServiceProxy(provider, service string, proxy ServiceProxy) {
	key := fmt.Sprintf("%s_%s", provider, service)
	e.serviceProxies[key] = proxy
}

// SetProviderToken sets the OAuth token for a specific provider
func (e *MultiProviderWorkflowEngine) SetProviderToken(provider string, token string) {
	e.tokens[provider] = token
}

// ExecuteWorkflow executes a complete workflow using the multi-provider proxy architecture
func (e *MultiProviderWorkflowEngine) ExecuteWorkflow(ctx context.Context, steps []WorkflowStep, input map[string]interface{}) (*WorkflowExecution, error) {
	execution := &WorkflowExecution{
		ID:          fmt.Sprintf("workflow_%d", time.Now().Unix()),
		Steps:       steps,
		StepResults: make(map[string]*ProxyResponse),
		Input:       input,
		Status:      "running",
		StartTime:   time.Now(),
	}

	// Execute steps in dependency order
	for _, step := range steps {
		// Check if dependencies are satisfied
		if !e.areDependenciesSatisfied(step, execution) {
			execution.Status = "failed"
			execution.ErrorMessage = fmt.Sprintf("Dependencies not satisfied for step %s", step.ID)
			endTime := time.Now()
			execution.EndTime = &endTime
			return execution, fmt.Errorf("dependencies not satisfied for step %s", step.ID)
		}

		// Resolve payload with data from previous steps
		resolvedPayload := e.resolvePayload(step.Payload, execution)

		// Execute the step using the appropriate service proxy
		response, err := e.executeStep(ctx, step, resolvedPayload)
		if err != nil {
			execution.Status = "failed"
			execution.ErrorMessage = fmt.Sprintf("Step %s failed: %v", step.ID, err)
			endTime := time.Now()
			execution.EndTime = &endTime
			return execution, err
		}

		// Store the result
		execution.StepResults[step.ID] = response
	}

	execution.Status = "completed"
	endTime := time.Now()
	execution.EndTime = &endTime
	return execution, nil
}

// executeStep executes a single workflow step using the appropriate service proxy
func (e *MultiProviderWorkflowEngine) executeStep(ctx context.Context, step WorkflowStep, payload map[string]interface{}) (*ProxyResponse, error) {
	// Get the service proxy key
	proxyKey := fmt.Sprintf("%s_%s", step.Provider, step.Service)

	// Find the appropriate service proxy
	proxy, exists := e.serviceProxies[proxyKey]
	if !exists {
		return nil, fmt.Errorf("no proxy found for %s", proxyKey)
	}

	// Get the provider token
	token, exists := e.tokens[step.Provider]
	if !exists {
		return nil, fmt.Errorf("no token found for provider %s", step.Provider)
	}

	// Execute the step with retry logic if configured
	if step.RetryPolicy != nil {
		return e.executeWithRetry(ctx, proxy, step, token, payload)
	}

	// Execute without retry
	return proxy.Execute(ctx, step.Function, token, payload)
}

// executeWithRetry executes a step with retry logic
func (e *MultiProviderWorkflowEngine) executeWithRetry(ctx context.Context, proxy ServiceProxy, step WorkflowStep, token string, payload map[string]interface{}) (*ProxyResponse, error) {
	var lastErr error
	delay := step.RetryPolicy.RetryDelay

	for attempt := 0; attempt <= step.RetryPolicy.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
			// Increase delay for next attempt
			delay = time.Duration(float64(delay) * step.RetryPolicy.BackoffFactor)
		}

		response, err := proxy.Execute(ctx, step.Function, token, payload)
		if err == nil {
			return response, nil
		}

		lastErr = err

		// Check if the error is retryable
		if response != nil && response.Error != nil && !response.Error.Retryable {
			break
		}
	}

	return nil, fmt.Errorf("step failed after %d retries: %v", step.RetryPolicy.MaxRetries, lastErr)
}

// areDependenciesSatisfied checks if all dependencies for a step are satisfied
func (e *MultiProviderWorkflowEngine) areDependenciesSatisfied(step WorkflowStep, execution *WorkflowExecution) bool {
	for _, depID := range step.DependsOn {
		if _, exists := execution.StepResults[depID]; !exists {
			return false
		}
		// Check if the dependency step was successful
		if result := execution.StepResults[depID]; result != nil && !result.Success {
			return false
		}
	}
	return true
}

// resolvePayload resolves payload references to data from previous steps
func (e *MultiProviderWorkflowEngine) resolvePayload(payload map[string]interface{}, execution *WorkflowExecution) map[string]interface{} {
	resolved := make(map[string]interface{})

	for key, value := range payload {
		resolved[key] = e.resolveValue(value, execution)
	}

	return resolved
}

// resolveValue resolves a single value, handling references to previous step results
func (e *MultiProviderWorkflowEngine) resolveValue(value interface{}, execution *WorkflowExecution) interface{} {
	switch v := value.(type) {
	case string:
		// Handle step result references like "${step_id.field_name}"
		if strings.HasPrefix(v, "${") && strings.HasSuffix(v, "}") {
			ref := v[2 : len(v)-1] // Remove ${ and }
			parts := strings.Split(ref, ".")
			if len(parts) >= 2 {
				stepID := parts[0]
				fieldPath := strings.Join(parts[1:], ".")

				if result, exists := execution.StepResults[stepID]; exists && result.Success {
					if resolvedValue := e.getNestedValue(result.Data, fieldPath); resolvedValue != nil {
						return resolvedValue
					}
					// If field not found, return original template for debugging
					return fmt.Sprintf("[UNRESOLVED: %s]", v)
				}
			}
		}
		return v
	case map[string]interface{}:
		resolved := make(map[string]interface{})
		for k, val := range v {
			resolved[k] = e.resolveValue(val, execution)
		}
		return resolved
	case []interface{}:
		resolved := make([]interface{}, len(v))
		for i, val := range v {
			resolved[i] = e.resolveValue(val, execution)
		}
		return resolved
	default:
		return v
	}
}

// getNestedValue retrieves a nested value from a map using dot notation
func (e *MultiProviderWorkflowEngine) getNestedValue(data map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		if i == len(parts)-1 {
			return current[part]
		}

		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return nil
		}
	}

	return nil
}

// GetSupportedProviders returns a list of all registered providers
func (e *MultiProviderWorkflowEngine) GetSupportedProviders() []string {
	providers := make(map[string]bool)
	for key := range e.serviceProxies {
		parts := strings.Split(key, "_")
		if len(parts) > 0 {
			providers[parts[0]] = true
		}
	}

	result := make([]string, 0, len(providers))
	for provider := range providers {
		result = append(result, provider)
	}
	return result
}

// GetSupportedServices returns a list of all supported services for a provider
func (e *MultiProviderWorkflowEngine) GetSupportedServices(provider string) []string {
	services := make(map[string]bool)
	prefix := provider + "_"

	for key := range e.serviceProxies {
		if strings.HasPrefix(key, prefix) {
			service := strings.TrimPrefix(key, prefix)
			services[service] = true
		}
	}

	result := make([]string, 0, len(services))
	for service := range services {
		result = append(result, service)
	}
	return result
}
