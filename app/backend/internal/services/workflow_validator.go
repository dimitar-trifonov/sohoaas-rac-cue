package services

import (
	"fmt"
	"strings"

	"sohoaas-backend/internal/types"
)

// RaC-compliant validation result structures
type ValidationResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors,omitempty"`
}

type WorkflowValidationState struct {
	ParameterValidation  ValidationResult `json:"parameter_validation"`
	ServiceValidation    ValidationResult `json:"service_validation"`
	DependencyValidation ValidationResult `json:"dependency_validation"`
	OAuthValidation      ValidationResult `json:"oauth_validation"`
	ExecutionReady       bool             `json:"execution_ready"`
	Status               string           `json:"status"`
}

// WorkflowValidator provides comprehensive workflow validation capabilities
// This implements the validation actions referenced in rac/agents/workflow_validator.cue
type WorkflowValidator struct {
	mcpParser *MCPCatalogParser
}

// NewWorkflowValidator creates a new workflow validator
func NewWorkflowValidator() *WorkflowValidator {
	return &WorkflowValidator{
		mcpParser: NewMCPCatalogParser(),
	}
}


// ValidateParameterReferencesTyped validates parameter references using strongly-typed structures
func (wv *WorkflowValidator) ValidateParameterReferencesTyped(steps []types.WorkflowStepValidation, userParameters map[string]interface{}) (bool, []string) {
	var errors []string
	
	for _, step := range steps {
		stepErrors := wv.validateParameterReferencesInMap(step.Parameters, step.ID, userParameters, nil)
		errors = append(errors, stepErrors...)
	}
	
	return len(errors) == 0, errors
}

// validateParameterReferencesInMap recursively validates parameter references in a map
func (wv *WorkflowValidator) validateParameterReferencesInMap(data map[string]interface{}, stepID string, userParameters map[string]interface{}, allSteps []map[string]interface{}) []string {
	var errors []string
	
	for _, value := range data {
		switch v := value.(type) {
		case string:
			if strings.Contains(v, "${") {
				paramErrors := wv.validateParameterReference(v, stepID, userParameters, allSteps)
				errors = append(errors, paramErrors...)
			}
		case map[string]interface{}:
			nestedErrors := wv.validateParameterReferencesInMap(v, stepID, userParameters, allSteps)
			errors = append(errors, nestedErrors...)
		case []interface{}:
			for _, item := range v {
				if itemStr, ok := item.(string); ok && strings.Contains(itemStr, "${") {
					paramErrors := wv.validateParameterReference(itemStr, stepID, userParameters, allSteps)
					errors = append(errors, paramErrors...)
				} else if itemMap, ok := item.(map[string]interface{}); ok {
					nestedErrors := wv.validateParameterReferencesInMap(itemMap, stepID, userParameters, allSteps)
					errors = append(errors, nestedErrors...)
				}
			}
		}
	}
	
	return errors
}

// validateParameterReference validates a single parameter reference
func (wv *WorkflowValidator) validateParameterReference(paramRef, stepID string, userParameters map[string]interface{}, allSteps []map[string]interface{}) []string {
	var errors []string
	
	// Parse the parameter reference using existing parser
	parsed := wv.mcpParser.ParseParameterReference(paramRef)
	if !parsed.IsValid {
		errors = append(errors, fmt.Sprintf("step %s: invalid parameter reference format '%s'", stepID, paramRef))
		return errors
	}
	
	// Validate based on parameter type
	switch parsed.Type {
	case types.ParamRefUser:
		// Validate user parameter exists
		if len(parsed.Path) < 2 {
			errors = append(errors, fmt.Sprintf("step %s: invalid user parameter reference '%s'", stepID, paramRef))
			return errors
		}
		paramName := parsed.Path[1]
		if _, exists := userParameters[paramName]; !exists {
			errors = append(errors, fmt.Sprintf("step %s: user parameter '%s' not found in workflow parameters", stepID, paramName))
		}
		
	case types.ParamRefStep:
		// Validate step output reference: ${steps.step_id.outputs.field}
		if len(parsed.Path) < 4 {
			errors = append(errors, fmt.Sprintf("step %s: invalid step output reference '%s' - expected format ${steps.step_id.outputs.field}", stepID, paramRef))
			return errors
		}
		
		referencedStepID := parsed.Path[1]
		
		// Check if referenced step exists
		stepExists := false
		if allSteps != nil {
			for _, step := range allSteps {
				if stepIDStr, ok := step["id"].(string); ok && stepIDStr == referencedStepID {
					stepExists = true
					break
				}
			}
		}
		
		if !stepExists && allSteps != nil {
			errors = append(errors, fmt.Sprintf("step %s: referenced step '%s' not found in workflow", stepID, referencedStepID))
		}
		
		// Validate that we're not referencing ourselves (would cause circular dependency)
		if referencedStepID == stepID {
			errors = append(errors, fmt.Sprintf("step %s: cannot reference own outputs - circular dependency detected", stepID))
		}
		
	case types.ParamRefEnv:
		// Environment variables are assumed to be available at runtime
		
	case types.ParamRefComputed:
		// Computed values are validated at execution time
		
	default:
		errors = append(errors, fmt.Sprintf("step %s: unsupported parameter reference type in '%s'", stepID, paramRef))
	}
	
	return errors
}


// ValidateStepDependenciesTyped validates step dependencies using strongly-typed structures
func (wv *WorkflowValidator) ValidateStepDependenciesTyped(steps []types.WorkflowStepValidation) (bool, []string) {
	var errors []string
	
	// Build dependency graph
	stepMap := make(map[string]types.WorkflowStepValidation)
	dependencies := make(map[string][]string)
	
	for _, step := range steps {
		if step.ID == "" {
			errors = append(errors, "step missing required 'id' field")
			continue
		}
		
		stepMap[step.ID] = step
		
		// Extract explicit dependencies
		for _, dep := range step.DependsOn {
			dependencies[step.ID] = append(dependencies[step.ID], dep)
		}
		
		// Extract implicit dependencies from parameter references
		implicitDeps := wv.extractImplicitDependencies(step.Parameters)
		for _, dep := range implicitDeps {
			wv.addUniqueDependency(dependencies, step.ID, dep)
		}
	}
	
	// Validate that all referenced steps exist
	for stepID, deps := range dependencies {
		for _, depID := range deps {
			if _, exists := stepMap[depID]; !exists {
				errors = append(errors, fmt.Sprintf("step %s: dependency '%s' not found in workflow", stepID, depID))
			}
		}
	}
	
	// Detect circular dependencies using DFS
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)
	
	for stepID := range stepMap {
		if !visited[stepID] {
			if wv.hasCyclicDependency(stepID, dependencies, visited, recursionStack) {
				errors = append(errors, fmt.Sprintf("circular dependency detected involving step '%s'", stepID))
			}
		}
	}
	
	return len(errors) == 0, errors
}

// extractImplicitDependencies extracts step dependencies from parameter references
func (wv *WorkflowValidator) extractImplicitDependencies(data map[string]interface{}) []string {
	var dependencies []string
	
	for _, value := range data {
		switch v := value.(type) {
		case string:
			stepRefs := wv.extractStepReferences(v)
			dependencies = append(dependencies, stepRefs...)
		case map[string]interface{}:
			nestedDeps := wv.extractImplicitDependencies(v)
			dependencies = append(dependencies, nestedDeps...)
		case []interface{}:
			for _, item := range v {
				if itemStr, ok := item.(string); ok {
					stepRefs := wv.extractStepReferences(itemStr)
					dependencies = append(dependencies, stepRefs...)
				} else if itemMap, ok := item.(map[string]interface{}); ok {
					nestedDeps := wv.extractImplicitDependencies(itemMap)
					dependencies = append(dependencies, nestedDeps...)
				}
			}
		}
	}
	
	return dependencies
}

// extractStepReferences extracts step references from a parameter value
func (wv *WorkflowValidator) extractStepReferences(value string) []string {
	var refs []string
	if strings.Contains(value, "${steps.") {
		parsed := wv.mcpParser.ParseParameterReference(value)
		if parsed.Type == types.ParamRefStep && len(parsed.Path) >= 2 {
			refs = append(refs, parsed.Path[1])
		}
	}
	return refs
}

// addUniqueDependency adds a dependency if it doesn't already exist
func (wv *WorkflowValidator) addUniqueDependency(dependencies map[string][]string, stepID, dep string) {
	for _, existing := range dependencies[stepID] {
		if existing == dep {
			return // Already exists
		}
	}
	dependencies[stepID] = append(dependencies[stepID], dep)
}

// hasCyclicDependency detects circular dependencies using DFS
func (wv *WorkflowValidator) hasCyclicDependency(stepID string, dependencies map[string][]string, visited, recursionStack map[string]bool) bool {
	visited[stepID] = true
	recursionStack[stepID] = true
	
	for _, depID := range dependencies[stepID] {
		if !visited[depID] {
			if wv.hasCyclicDependency(depID, dependencies, visited, recursionStack) {
				return true
			}
		} else if recursionStack[depID] {
			return true
		}
	}
	
	recursionStack[stepID] = false
	return false
}

// ComputeExecutionOrder computes a valid execution order using topological sort
// This supports the sequential execution approach decided for the PoC
func (wv *WorkflowValidator) ComputeExecutionOrder(steps []map[string]interface{}) ([]string, error) {
	// Build dependency graph
	dependencies := make(map[string][]string)
	stepMap := make(map[string]bool)
	
	for _, step := range steps {
		stepID, ok := step["id"].(string)
		if !ok {
			return nil, fmt.Errorf("step missing required 'id' field")
		}
		
		stepMap[stepID] = true
		
		// Extract explicit dependencies
		if dependsOn, ok := step["depends_on"].([]interface{}); ok {
			for _, dep := range dependsOn {
				if depStr, ok := dep.(string); ok {
					dependencies[stepID] = append(dependencies[stepID], depStr)
				}
			}
		}
		
		// Extract implicit dependencies from parameter references
		if inputs, ok := step["inputs"].(map[string]interface{}); ok {
			implicitDeps := wv.extractImplicitDependencies(inputs)
			for _, dep := range implicitDeps {
				wv.addUniqueDependency(dependencies, stepID, dep)
			}
		}
	}
	
	// Compute execution order using topological sort
	inDegree := make(map[string]int)
	
	// Initialize all steps
	for stepID := range stepMap {
		inDegree[stepID] = 0
	}
	
	// Count dependencies - increment in-degree for steps that depend on others
	for stepID, deps := range dependencies {
		for _, depID := range deps {
			if stepMap[depID] { // Only count if dependency exists
				inDegree[stepID]++
			}
		}
	}
	
	// Find steps with no dependencies
	var queue []string
	for stepID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, stepID)
		}
	}
	
	var result []string
	
	// Process steps in topological order
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)
		
		// Reduce in-degree for steps that depend on the current step
		for stepID, deps := range dependencies {
			for _, depID := range deps {
				if depID == current && stepMap[stepID] {
					inDegree[stepID]--
					if inDegree[stepID] == 0 {
						queue = append(queue, stepID)
					}
				}
			}
		}
	}
	
	// Check if all steps were processed (no circular dependencies)
	if len(result) != len(stepMap) {
		return nil, fmt.Errorf("circular dependencies detected - cannot determine execution order")
	}
	
	return result, nil
}

// =============================================
// RaC-COMPLIANT VALIDATION METHODS
// =============================================
// These methods implement the exact actions specified in rac/agents/workflow_validator.cue

// CheckUserParameters implements validator.check_user_parameters from RaC specification
func (wv *WorkflowValidator) CheckUserParameters(steps []map[string]interface{}, userParameters map[string]interface{}) ValidationResult {
	// Convert to strongly-typed structures for validation
	typedSteps, err := wv.mcpParser.ParseWorkflowSteps(steps)
	if err != nil {
		return ValidationResult{
			Valid:  false,
			Errors: []string{fmt.Sprintf("Failed to parse workflow steps: %v", err)},
		}
	}

	// Use strongly-typed parameter reference validation
	isValid, errors := wv.ValidateParameterReferencesTyped(typedSteps, userParameters)
	return ValidationResult{
		Valid:  isValid,
		Errors: errors,
	}
}

// CheckServiceAvailability implements validator.check_service_availability from RaC specification
func (wv *WorkflowValidator) CheckServiceAvailability(mcpCatalog map[string]interface{}, steps []map[string]interface{}) ValidationResult {
	// Convert to strongly-typed structures for validation
	catalog, err := wv.mcpParser.ParseMCPCatalog(mcpCatalog)
	if err != nil {
		return ValidationResult{
			Valid:  false,
			Errors: []string{fmt.Sprintf("Failed to parse MCP catalog: %v", err)},
		}
	}

	typedSteps, err := wv.mcpParser.ParseWorkflowSteps(steps)
	if err != nil {
		return ValidationResult{
			Valid:  false,
			Errors: []string{fmt.Sprintf("Failed to parse workflow steps: %v", err)},
		}
	}

	// Use strongly-typed MCP validation for service availability
	isValid, errors := wv.mcpParser.ValidateMCPFunctionsTyped(catalog, typedSteps)
	return ValidationResult{
		Valid:  isValid,
		Errors: errors,
	}
}

// CheckStepDependencies implements validator.check_step_dependencies from RaC specification
func (wv *WorkflowValidator) CheckStepDependencies(steps []map[string]interface{}) ValidationResult {
	// Convert to strongly-typed structures for validation
	typedSteps, err := wv.mcpParser.ParseWorkflowSteps(steps)
	if err != nil {
		return ValidationResult{
			Valid:  false,
			Errors: []string{fmt.Sprintf("Failed to parse workflow steps: %v", err)},
		}
	}

	// Use strongly-typed step dependency validation
	isValid, errors := wv.ValidateStepDependenciesTyped(typedSteps)
	return ValidationResult{
		Valid:  isValid,
		Errors: errors,
	}
}

// CheckOAuthPermissions implements validator.check_oauth_permissions from RaC specification
func (wv *WorkflowValidator) CheckOAuthPermissions(mcpCatalog map[string]interface{}, serviceBindings map[string]interface{}, steps []map[string]interface{}) ValidationResult {
	// Convert to strongly-typed structures for validation
	catalog, err := wv.mcpParser.ParseMCPCatalog(mcpCatalog)
	if err != nil {
		return ValidationResult{
			Valid:  false,
			Errors: []string{fmt.Sprintf("Failed to parse MCP catalog: %v", err)},
		}
	}

	typedSteps, err := wv.mcpParser.ParseWorkflowSteps(steps)
	if err != nil {
		return ValidationResult{
			Valid:  false,
			Errors: []string{fmt.Sprintf("Failed to parse workflow steps: %v", err)},
		}
	}

	typedBindings, err := wv.mcpParser.ParseServiceBindings(serviceBindings)
	if err != nil {
		return ValidationResult{
			Valid:  false,
			Errors: []string{fmt.Sprintf("Failed to parse service bindings: %v", err)},
		}
	}

	// Use strongly-typed service binding validation for OAuth permissions
	isValid, errors := wv.mcpParser.ValidateServiceBindingsTyped(catalog, typedBindings, typedSteps)
	return ValidationResult{
		Valid:  isValid,
		Errors: errors,
	}
}

// ValidateWorkflow implements the comprehensive validation workflow from RaC specification
// This is the main entry point that matches the RaC agent workflow_validation state
func (wv *WorkflowValidator) ValidateWorkflow(mcpCatalog map[string]interface{}, steps []map[string]interface{}, userParameters map[string]interface{}, serviceBindings map[string]interface{}) WorkflowValidationState {
	// Execute all validation steps as specified in RaC
	paramValidation := wv.CheckUserParameters(steps, userParameters)
	serviceValidation := wv.CheckServiceAvailability(mcpCatalog, steps)
	dependencyValidation := wv.CheckStepDependencies(steps)
	oauthValidation := wv.CheckOAuthPermissions(mcpCatalog, serviceBindings, steps)
	
	// Determine overall execution readiness
	executionReady := paramValidation.Valid && serviceValidation.Valid && dependencyValidation.Valid && oauthValidation.Valid
	
	// Set status based on validation results
	status := "ready"
	if !executionReady {
		status = "blocked"
	}
	
	return WorkflowValidationState{
		ParameterValidation:  paramValidation,
		ServiceValidation:    serviceValidation,
		DependencyValidation: dependencyValidation,
		OAuthValidation:      oauthValidation,
		ExecutionReady:       executionReady,
		Status:               status,
	}
}
