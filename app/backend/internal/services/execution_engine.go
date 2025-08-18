package services

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"sohoaas-backend/internal/types"
)

// ExecutionEngine handles workflow execution with parameter replacement
type ExecutionEngine struct {
	mcpService     *MCPService
	mcpParser      *MCPCatalogParser
	serviceCatalog types.ServiceCatalog
}

// NewExecutionEngine creates a new execution engine
func NewExecutionEngine(mcpService *MCPService) *ExecutionEngine {
	return &ExecutionEngine{
		mcpService:     mcpService,
		mcpParser:      NewMCPCatalogParser(),
		serviceCatalog: types.ServiceCatalog{}, // Will be populated dynamically from MCP
	}
}

// ValidateWorkflowServices validates that all services in a workflow exist in the MCP service catalog
// and validates output field references against MCP response schemas
func (ee *ExecutionEngine) ValidateWorkflowServices(workflow *ParsedWorkflow) error {
	if workflow == nil {
		return fmt.Errorf("workflow is nil")
	}
	
	// Query live MCP service catalog for validation using centralized MCPService
	mcpServices, err := ee.mcpService.GetServiceCatalog()
	if err != nil {
		return fmt.Errorf("failed to query MCP service catalog for validation: %w", err)
	}
	
	// Validate workflow services against MCP catalog
	if err := ee.validateWorkflowServicesInternal(mcpServices, workflow); err != nil {
		return err
	}
	
	// Validate output field references against MCP response schemas
	return ee.validateOutputFieldReferences(mcpServices, workflow)
}

// validateWorkflowServicesInternal validates workflow services against MCP catalog
// Supports both legacy map[string]interface{} and new *types.MCPServiceCatalog
func (ee *ExecutionEngine) validateWorkflowServicesInternal(mcpCatalog interface{}, workflow *ParsedWorkflow) error {
	servicesData, err := ee.mcpParser.ParseServicesFromCatalog(mcpCatalog)
	if err != nil {
		return err
	}
	
	for i, step := range workflow.Steps {
		// Validate service exists in MCP catalog
		serviceData, exists := servicesData[step.Service]
		if !exists {
			return fmt.Errorf("unknown service '%s' in step %d (%s) - service not found in MCP catalog", step.Service, i, step.ID)
		}
		
		// Handle strongly-typed service definition
		if serviceDefinition, ok := serviceData.(types.MCPServiceDefinition); ok {
			// Check if action exists in the service's functions
			_, actionExists := serviceDefinition.Functions[step.Action]
			if !actionExists {
				return fmt.Errorf("unknown action '%s' for service '%s' in step %d (%s) - action not found in MCP catalog", step.Action, step.Service, i, step.ID)
			}
			continue
		}
		
		// Fallback: Handle legacy map format for backward compatibility
		serviceMap, ok := serviceData.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid service data for '%s' in MCP catalog", step.Service)
		}
		
		functions, ok := serviceMap["functions"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("no functions found for service '%s' in MCP catalog", step.Service)
		}
		
		_, actionExists := functions[step.Action]
		if !actionExists {
			return fmt.Errorf("unknown action '%s' for service '%s' in step %d (%s) - action not found in MCP catalog", step.Action, step.Service, i, step.ID)
		}
	}
	
	return nil
}

// validateOutputFieldReferences validates that workflow step output references exist in MCP response schemas
func (ee *ExecutionEngine) validateOutputFieldReferences(mcpCatalog *types.MCPServiceCatalog, workflow *ParsedWorkflow) error {
	stepOutputRegex := regexp.MustCompile(`\$\{steps\.([^.]+)\.outputs\.([^}]+)\}`)
	
	// Build map of step ID to service/action for lookup
	stepServiceMap := make(map[string]struct{service, action string})
	for _, step := range workflow.Steps {
		stepServiceMap[step.ID] = struct{service, action string}{step.Service, step.Action}
	}
	
	// Check each step's parameters for output field references
	for _, step := range workflow.Steps {
		// Check all parameter values recursively
		if err := ee.validateParameterOutputReferences(step.Inputs, stepOutputRegex, stepServiceMap, mcpCatalog); err != nil {
			return fmt.Errorf("invalid output reference in step %s: %w", step.ID, err)
		}
	}
	
	return nil
}

// validateParameterOutputReferences recursively validates output field references in parameters
func (ee *ExecutionEngine) validateParameterOutputReferences(params map[string]interface{}, stepOutputRegex *regexp.Regexp, stepServiceMap map[string]struct{service, action string}, mcpCatalog *types.MCPServiceCatalog) error {
	for paramName, paramValue := range params {
		switch v := paramValue.(type) {
		case string:
			// Check for step output references in string parameters
			matches := stepOutputRegex.FindAllStringSubmatch(v, -1)
			for _, match := range matches {
				stepID := match[1]
				outputField := match[2]
				
				// Get service and action for the referenced step
				stepInfo, exists := stepServiceMap[stepID]
				if !exists {
					return fmt.Errorf("parameter %s references unknown step: %s", paramName, stepID)
				}
				
				// Validate that the output field exists in the MCP function's output schema
				if err := ee.validateOutputFieldExists(stepInfo.service, stepInfo.action, outputField, mcpCatalog); err != nil {
					return fmt.Errorf("parameter %s references invalid output field %s.%s: %w", paramName, stepID, outputField, err)
				}
			}
		case map[string]interface{}:
			// Recursively validate nested objects
			if err := ee.validateParameterOutputReferences(v, stepOutputRegex, stepServiceMap, mcpCatalog); err != nil {
				return err
			}
		case []interface{}:
			// Validate array elements
			for i, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if err := ee.validateParameterOutputReferences(itemMap, stepOutputRegex, stepServiceMap, mcpCatalog); err != nil {
						return fmt.Errorf("array item %d: %w", i, err)
					}
				}
			}
		}
	}
	return nil
}

// validateOutputFieldExists checks if an output field exists in the MCP function's output schema
func (ee *ExecutionEngine) validateOutputFieldExists(service, action, outputField string, mcpCatalog *types.MCPServiceCatalog) error {
	// Check if service exists
	serviceDefinition, exists := mcpCatalog.Providers.Workspace.Services[service]
	if !exists {
		return fmt.Errorf("service '%s' not found in MCP catalog", service)
	}
	
	// Check if function exists
	functionSchema, exists := serviceDefinition.Functions[action]
	if !exists {
		return fmt.Errorf("function '%s' not found in service '%s'", action, service)
	}
	
	// If no output schema defined, allow any field reference (backward compatibility)
	if functionSchema.OutputSchema == nil || functionSchema.OutputSchema.Properties == nil {
		log.Printf("[ExecutionEngine] validateOutputFieldExists: No output schema defined for %s.%s, allowing field reference: %s", service, action, outputField)
		return nil
	}
	
	// Check if the output field exists in the schema
	if _, exists := functionSchema.OutputSchema.Properties[outputField]; !exists {
		availableFields := make([]string, 0, len(functionSchema.OutputSchema.Properties))
		for field := range functionSchema.OutputSchema.Properties {
			availableFields = append(availableFields, field)
		}
		return fmt.Errorf("output field '%s' not found in %s.%s schema. Available fields: %v", outputField, service, action, availableFields)
	}
	
	return nil
}

// ParameterContext holds all parameter values for workflow execution
type ParameterContext struct {
	UserParameters    map[string]interface{} `json:"user_parameters"`
	RuntimeParameters map[string]interface{} `json:"runtime_parameters"`
	SystemParameters  map[string]interface{} `json:"system_parameters"`
	StepOutputs       map[string]interface{} `json:"step_outputs"`
}

// ExecutionPlan represents a workflow ready for execution with resolved parameters
type ExecutionPlan struct {
	WorkflowID       string                 `json:"workflow_id"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	ResolvedSteps    []ResolvedStep         `json:"resolved_steps"`
	ParameterContext *ParameterContext      `json:"parameter_context"`
	ValidationErrors []string               `json:"validation_errors,omitempty"`
}

// ResolvedStep represents a workflow step with all parameters resolved
type ResolvedStep struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Service     string                 `json:"service"`
	Action      string                 `json:"action"`
	Inputs      map[string]interface{} `json:"inputs"`
	Outputs     map[string]interface{} `json:"outputs"`
	DependsOn   []string               `json:"depends_on,omitempty"`
	Status      string                 `json:"status"` // pending, running, completed, failed
}

// PrepareExecution analyzes a CUE workflow and creates an execution plan
func (ee *ExecutionEngine) PrepareExecution(cueworkflow string, userID string, user *types.User, intentAnalysis map[string]interface{}, oauthToken string, userTimezone string) (*ExecutionPlan, error) {
	// Parse the CUE workflow (simplified - would use actual CUE parser in production)
	workflow, err := ee.ParseCUEWorkflow(cueworkflow)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CUE workflow: %w", err)
	}

	// Validate all services in workflow against service catalog
	if err := ee.ValidateWorkflowServices(workflow); err != nil {
		return nil, fmt.Errorf("workflow service validation failed: %w", err)
	}

	// Create parameter context from intent analysis and user data
	paramContext := ee.createParameterContext(intentAnalysis, user, oauthToken, userTimezone)

	// Resolve all parameters in workflow steps
	resolvedSteps, validationErrors := ee.resolveWorkflowParameters(workflow.Steps, paramContext)

	executionPlan := &ExecutionPlan{
		WorkflowID:       fmt.Sprintf("%s_%d", userID, time.Now().Unix()),
		Name:             workflow.Name,
		Description:      workflow.Description,
		ResolvedSteps:    resolvedSteps,
		ParameterContext: paramContext,
		ValidationErrors: validationErrors,
	}

	return executionPlan, nil
}

// createParameterContext builds the parameter context from various sources
func (ee *ExecutionEngine) createParameterContext(intentAnalysis map[string]interface{}, user *types.User, oauthToken string, userTimezone string) *ParameterContext {
	context := &ParameterContext{
		UserParameters:    make(map[string]interface{}),
		RuntimeParameters: make(map[string]interface{}),
		SystemParameters:  make(map[string]interface{}),
		StepOutputs:       make(map[string]interface{}),
	}

	// Extract user parameters from intent analysis
	// Handle both array format (legacy) and map format (current standard)
	if userParams, ok := intentAnalysis["user_parameters"].([]interface{}); ok {
		// Legacy array format: [{"name": "param", "value": "val"}]
		for _, param := range userParams {
			if paramMap, ok := param.(map[string]interface{}); ok {
				if name, exists := paramMap["name"].(string); exists {
					context.UserParameters[name] = ee.resolveUserParameter(paramMap, user)
				}
			}
		}
	} else if userParamsMap, ok := intentAnalysis["user_parameters"].(map[string]interface{}); ok {
		// Current map format: {"param_name": "param_value"}
		for paramName, paramValue := range userParamsMap {
			context.UserParameters[paramName] = paramValue
		}
	} else {
		// Direct parameter map (when intentAnalysis IS the user parameters)
		for key, value := range intentAnalysis {
			if key != "user_parameters" {
				context.UserParameters[key] = value
			}
		}
	}

	// Set system parameters
	context.SystemParameters["current_date"] = time.Now().Format("2006-01-02")
	context.SystemParameters["current_datetime"] = time.Now().Format("2006-01-02T15:04:05")
	context.SystemParameters["user_email"] = user.Email
	context.SystemParameters["user_id"] = user.ID
	context.SystemParameters["oauth_token"] = oauthToken
	context.SystemParameters["user_timezone"] = userTimezone

	return context
}

// resolveUserParameter resolves a user parameter from various sources
func (ee *ExecutionEngine) resolveUserParameter(paramDef map[string]interface{}, user *types.User) interface{} {
	paramName := paramDef["name"].(string)
	
	// Check if we can auto-resolve from user profile
	switch paramName {
	case "user_email", "email":
		return user.Email
	case "user_id":
		return user.ID
	default:
		// Return default value if available, otherwise placeholder
		if defaultVal, exists := paramDef["default"]; exists {
			return defaultVal
		}
		return fmt.Sprintf("${user.%s}", paramName)
	}
}

// resolveWorkflowParameters resolves all parameters in workflow steps
func (ee *ExecutionEngine) resolveWorkflowParameters(steps []WorkflowStep, context *ParameterContext) ([]ResolvedStep, []string) {
	var resolvedSteps []ResolvedStep
	var validationErrors []string

	for _, step := range steps {
		resolvedStep := ResolvedStep{
			ID:        step.ID,
			Name:      step.Name,
			Service:   step.Service,
			Action:    step.Action,
			DependsOn: step.DependsOn,
			Status:    "pending",
			Inputs:    make(map[string]interface{}),
			Outputs:   make(map[string]interface{}),
		}

		// Resolve input parameters
		for key, value := range step.Inputs {
			resolvedValue, err := ee.resolveParameterValue(value, context)
			if err != nil {
				validationErrors = append(validationErrors, fmt.Sprintf("Step %s, input %s: %v", step.ID, key, err))
			}
			resolvedStep.Inputs[key] = resolvedValue
		}

		// Resolve output parameters
		for key, value := range step.Outputs {
			resolvedValue, err := ee.resolveParameterValue(value, context)
			if err != nil {
				validationErrors = append(validationErrors, fmt.Sprintf("Step %s, output %s: %v", step.ID, key, err))
			}
			resolvedStep.Outputs[key] = resolvedValue
		}

		resolvedSteps = append(resolvedSteps, resolvedStep)
	}

	return resolvedSteps, validationErrors
}

// resolveParameterValue resolves a single parameter value using various strategies
func (ee *ExecutionEngine) resolveParameterValue(value interface{}, context *ParameterContext) (interface{}, error) {
	switch v := value.(type) {
	case string:
		// Handle parameter references in strings
		resolved, err := ee.resolveStringParameter(v, context)
		if err != nil {
			return nil, err
		}
		
		// Apply timezone conversion to resolved string values
		if resolvedStr, ok := resolved.(string); ok && ee.isDateTimeValue(resolvedStr) {
			if userTimezone, exists := context.SystemParameters["user_timezone"]; exists {
				if timezone, ok := userTimezone.(string); ok && timezone != "" {
					// If datetime doesn't have timezone info, add it using user's timezone
					// Check for timezone offset indicators after the time part (not in date part)
					// Look for + or - after position 19 (after "2025-08-18T10:00:00") or Z at the end
					hasTimezoneOffset := false
					if len(resolvedStr) > 19 { // "2025-08-18T10:00:00" is 19 chars
						timezonePart := resolvedStr[19:] // Everything after the seconds
						hasTimezoneOffset = strings.Contains(timezonePart, "+") || strings.Contains(timezonePart, "-")
					}
					hasTimezone := hasTimezoneOffset || strings.HasSuffix(resolvedStr, "Z")
					
					if !hasTimezone {
						// Parse the datetime and add timezone
						if parsedTime, err := time.Parse("2006-01-02T15:04:05", resolvedStr); err == nil {
							// Load the user's timezone
							if loc, err := time.LoadLocation(timezone); err == nil {
								// Interpret the datetime as being in the user's timezone and format with offset
								localTime := time.Date(parsedTime.Year(), parsedTime.Month(), parsedTime.Day(),
									parsedTime.Hour(), parsedTime.Minute(), parsedTime.Second(), parsedTime.Nanosecond(), loc)
								return localTime.Format("2006-01-02T15:04:05-07:00"), nil
							}
						}
					}
				}
			}
		}
		return resolved, nil
	case map[string]interface{}:
		// Recursively resolve nested objects
		resolved := make(map[string]interface{})
		for key, val := range v {
			resolvedVal, err := ee.resolveParameterValue(val, context)
			if err != nil {
				return nil, err
			}
			resolved[key] = resolvedVal
		}
		return resolved, nil
	case []interface{}:
		// Recursively resolve arrays
		resolved := make([]interface{}, len(v))
		for i, val := range v {
			resolvedVal, err := ee.resolveParameterValue(val, context)
			if err != nil {
				return nil, err
			}
			resolved[i] = resolvedVal
		}
		return resolved, nil
	default:
		// Return primitive values as-is
		return value, nil
	}
}

// resolveStringParameter resolves string parameters with various expression types
func (ee *ExecutionEngine) resolveStringParameter(value string, context *ParameterContext) (interface{}, error) {
	// Handle runtime expressions: $(step.output.field)
	runtimeExpr := regexp.MustCompile(`\$\(([^.]+)\.outputs\.([^)]+)\)`)
	if matches := runtimeExpr.FindStringSubmatch(value); len(matches) == 3 {
		stepID := matches[1]
		outputField := matches[2]
		
		if stepOutputs, exists := context.StepOutputs[stepID]; exists {
			if outputMap, ok := stepOutputs.(map[string]interface{}); ok {
				if fieldValue, fieldExists := outputMap[outputField]; fieldExists {
					return fieldValue, nil
				}
			}
		}
		
		// Return placeholder for runtime resolution
		return fmt.Sprintf("${RUNTIME:%s.%s}", stepID, outputField), nil
	}

	// Handle system function calls: $(date '+%Y-%m-%d')
	systemFunc := regexp.MustCompile(`\$\(date\s+'([^']+)'\)`)
	if matches := systemFunc.FindStringSubmatch(value); len(matches) == 2 {
		format := matches[1]
		// Convert to Go time format (simplified)
		goFormat := strings.ReplaceAll(format, "%Y", "2006")
		goFormat = strings.ReplaceAll(goFormat, "%m", "01")
		goFormat = strings.ReplaceAll(goFormat, "%d", "02")
		goFormat = strings.ReplaceAll(goFormat, "%H", "15")
		goFormat = strings.ReplaceAll(goFormat, "%M", "04")
		goFormat = strings.ReplaceAll(goFormat, "%S", "05")
		
		return time.Now().Format(goFormat), nil
	}

	// Handle user parameter references (both pure and mixed with other text)
	userParamRegex := regexp.MustCompile(`\$\{user\.([^}]+)\}`)
	if userParamRegex.MatchString(value) {
		// Replace all user parameter references in the string
		result := userParamRegex.ReplaceAllStringFunc(value, func(match string) string {
			paramName := userParamRegex.FindStringSubmatch(match)[1]
			if userValue, exists := context.UserParameters[paramName]; exists {
				return fmt.Sprintf("%v", userValue)
			}
			// Return original match if parameter not found (will cause validation error later)
			return match
		})
		
		// Check if any parameters were not resolved (still contain ${user.})
		if strings.Contains(result, "${user.") {
			// Extract unresolved parameter names for error reporting
			unresolvedMatches := userParamRegex.FindAllStringSubmatch(value, -1)
			var missingParams []string
			for _, match := range unresolvedMatches {
				paramName := match[1]
				if _, exists := context.UserParameters[paramName]; !exists {
					missingParams = append(missingParams, paramName)
				}
			}
			if len(missingParams) > 0 {
				return value, fmt.Errorf("user parameter %s not provided", strings.Join(missingParams, ", "))
			}
		}
		
		return result, nil
	}


	// Handle step output references: ${steps.step_id.outputs.field}
	stepOutputRegex := regexp.MustCompile(`\$\{steps\.([^.]+)\.outputs\.([^}]+)\}`)
	if stepOutputRegex.MatchString(value) {
		result := stepOutputRegex.ReplaceAllStringFunc(value, func(match string) string {
			matches := stepOutputRegex.FindStringSubmatch(match)
			stepID := matches[1]
			outputField := matches[2]
			
			if stepOutputs, exists := context.StepOutputs[stepID]; exists {
				if outputMap, ok := stepOutputs.(map[string]interface{}); ok {
					if outputValue, exists := outputMap[outputField]; exists {
						return fmt.Sprintf("%v", outputValue)
					}
				}
			}
			return match // Keep original if not found during execution
		})
		
		// Only validate step output availability during actual execution, not pre-validation
		// During validation phase, step outputs won't exist yet - this is expected
		if strings.Contains(result, "${steps.") && len(context.StepOutputs) > 0 {
			// Only check for missing outputs if we're in execution phase (StepOutputs populated)
			unresolvedMatches := stepOutputRegex.FindAllStringSubmatch(value, -1)
			var missingRefs []string
			for _, match := range unresolvedMatches {
				stepID := match[1]
				outputField := match[2]
				if stepOutputs, exists := context.StepOutputs[stepID]; exists {
					if outputMap, ok := stepOutputs.(map[string]interface{}); ok {
						if _, exists := outputMap[outputField]; !exists {
							missingRefs = append(missingRefs, fmt.Sprintf("%s.%s", stepID, outputField))
						}
					}
				} else {
					missingRefs = append(missingRefs, fmt.Sprintf("%s.%s", stepID, outputField))
				}
			}
			if len(missingRefs) > 0 {
				return value, fmt.Errorf("step output reference %s not available", strings.Join(missingRefs, ", "))
			}
		}
		
		return result, nil
	}

	// Handle system parameter references: ${SYSTEM:param} format
	if strings.HasPrefix(value, "${SYSTEM:") && strings.HasSuffix(value, "}") {
		paramName := strings.TrimSuffix(strings.TrimPrefix(value, "${SYSTEM:"), "}")
		if systemValue, exists := context.SystemParameters[paramName]; exists {
			return systemValue, nil
		}
		return value, fmt.Errorf("system parameter %s not available", paramName)
	}

	// Handle standard system parameter references: ${param_name} format
	if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") && !strings.Contains(value, ".") {
		paramName := strings.TrimSuffix(strings.TrimPrefix(value, "${"), "}")
		if systemValue, exists := context.SystemParameters[paramName]; exists {
			return systemValue, nil
		}
		// Return as-is if not found (might be a literal string)
	}

	// Handle datetime values that need timezone information for API calls
	if ee.isDateTimeValue(value) {
		if userTimezone, exists := context.SystemParameters["user_timezone"]; exists {
			if timezone, ok := userTimezone.(string); ok && timezone != "" {
				// If datetime doesn't have timezone info, add it using user's timezone
				if !strings.Contains(value, "+") && !strings.Contains(value, "-") && !strings.HasSuffix(value, "Z") {
					// Parse the datetime and add timezone
					if parsedTime, err := time.Parse("2006-01-02T15:04:05", value); err == nil {
						// Load the user's timezone
						if loc, err := time.LoadLocation(timezone); err == nil {
							// Interpret the datetime as being in the user's timezone and format with offset
							localTime := time.Date(parsedTime.Year(), parsedTime.Month(), parsedTime.Day(),
								parsedTime.Hour(), parsedTime.Minute(), parsedTime.Second(), parsedTime.Nanosecond(), loc)
							return localTime.Format("2006-01-02T15:04:05-07:00"), nil
						}
					}
				}
			}
		}
	}



	// No parameter substitution needed, return as-is
	return value, nil
}

// isDateTimeValue checks if a string value looks like a datetime
func (ee *ExecutionEngine) isDateTimeValue(value string) bool {
	// Check if the value matches datetime patterns
	datetimePatterns := []string{
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05+07:00",
	}
	
	for _, pattern := range datetimePatterns {
		if _, err := time.Parse(pattern, value); err == nil {
			return true
		}
	}
	
	return false
}


// WorkflowStep represents a step in the workflow (simplified CUE parsing)
type WorkflowStep struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Service   string                 `json:"service"`
	Action    string                 `json:"action"`
	Inputs    map[string]interface{} `json:"inputs"`
	Outputs   map[string]interface{} `json:"outputs"`
	DependsOn []string               `json:"depends_on,omitempty"`
}

// ParsedWorkflow represents a parsed CUE workflow
type ParsedWorkflow struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Steps       []WorkflowStep `json:"steps"`
}

// ParseCUEWorkflow parses a CUE workflow string using the CUE library (public for testing)
func (ee *ExecutionEngine) ParseCUEWorkflow(cueContent string) (*ParsedWorkflow, error) {
	// Sanitize CUE content to remove illegal characters
	sanitizedContent := ee.sanitizeCUEContent(cueContent)
	
	// Create CUE context
	ctx := cuecontext.New()
	
	// Parse the CUE content (schema is already embedded in saved files)
	value := ctx.CompileString(sanitizedContent)
	if err := value.Err(); err != nil {
		return nil, fmt.Errorf("failed to compile CUE content: %w", err)
	}
	
	// Extract the workflow from the CUE value
	workflowValue := value.LookupPath(cue.ParsePath("workflow"))
	if !workflowValue.Exists() {
		return nil, fmt.Errorf("workflow field not found in CUE content")
	}
	
	// Parse workflow name
	nameValue := workflowValue.LookupPath(cue.ParsePath("name"))
	name, err := nameValue.String()
	if err != nil {
		return nil, fmt.Errorf("failed to extract workflow name: %w", err)
	}
	
	// Parse workflow description
	descValue := workflowValue.LookupPath(cue.ParsePath("description"))
	description, err := descValue.String()
	if err != nil {
		return nil, fmt.Errorf("failed to extract workflow description: %w", err)
	}
	
	// Parse workflow steps
	stepsValue := workflowValue.LookupPath(cue.ParsePath("steps"))
	if !stepsValue.Exists() {
		return nil, fmt.Errorf("steps field not found in workflow")
	}
	
	var steps []WorkflowStep
	stepsIter, err := stepsValue.List()
	if err != nil {
		return nil, fmt.Errorf("failed to iterate over steps: %w", err)
	}
	
	for stepsIter.Next() {
		stepValue := stepsIter.Value()
		
		// Parse step fields
		step := WorkflowStep{
			Inputs:  make(map[string]interface{}),
			Outputs: make(map[string]interface{}),
		}
		
		// Extract step ID
		if idValue := stepValue.LookupPath(cue.ParsePath("id")); idValue.Exists() {
			if id, err := idValue.String(); err != nil {
				return nil, fmt.Errorf("failed to extract id from step %d: %w", len(steps), err)
			} else {
				step.ID = id
			}
		}
		
		// Extract step name
		if nameValue := stepValue.LookupPath(cue.ParsePath("name")); nameValue.Exists() {
			if name, err := nameValue.String(); err != nil {
				return nil, fmt.Errorf("failed to extract name from step %d: %w", len(steps), err)
			} else {
				step.Name = name
			}
		}
		
		// Extract service field first (if exists)
		if serviceValue := stepValue.LookupPath(cue.ParsePath("service")); serviceValue.Exists() {
			if service, err := serviceValue.String(); err != nil {
				return nil, fmt.Errorf("failed to extract service from step %d: %w", len(steps), err)
			} else {
				step.Service = service
			}
		}
		
		// Extract action field
		if actionValue := stepValue.LookupPath(cue.ParsePath("action")); actionValue.Exists() {
			if action, err := actionValue.String(); err != nil {
				return nil, fmt.Errorf("failed to extract action from step %d: %w", len(steps), err)
			} else {
				// If action contains a dot and no service was set, split it
				if step.Service == "" && strings.Contains(action, ".") {
					parts := strings.SplitN(action, ".", 2)
					if len(parts) == 2 {
						step.Service = parts[0]  // e.g., "gmail"
						step.Action = parts[1]   // e.g., "get_messages"
					} else {
						step.Action = action
					}
				} else {
					step.Action = action
				}
			}
		}
		
		// Extract parameters/inputs (try both "parameters" and "inputs" fields)
		var inputsMap map[string]interface{}
		if parametersValue := stepValue.LookupPath(cue.ParsePath("parameters")); parametersValue.Exists() {
			inputsMap = make(map[string]interface{})
			parametersIter, _ := parametersValue.Fields()
			for parametersIter.Next() {
				key := parametersIter.Label()
				val := parametersIter.Value()
				
				// Convert CUE value to Go interface{}
				if goVal, err := ee.cueValueToInterface(val); err == nil {
					inputsMap[key] = goVal
				}
			}
		} else if inputsValue := stepValue.LookupPath(cue.ParsePath("inputs")); inputsValue.Exists() {
			inputsMap = make(map[string]interface{})
			inputsIter, _ := inputsValue.Fields()
			for inputsIter.Next() {
				key := inputsIter.Label()
				val := inputsIter.Value()
				
				// Convert CUE value to Go interface{}
				if goVal, err := ee.cueValueToInterface(val); err == nil {
					inputsMap[key] = goVal
				}
			}
		}
		if inputsMap != nil {
			step.Inputs = inputsMap // Store in Inputs field for execution engine compatibility
		}
		
		// Extract outputs (usually empty in workflow definition)
		if outputsValue := stepValue.LookupPath(cue.ParsePath("outputs")); outputsValue.Exists() {
			outputsMap := make(map[string]interface{})
			outputsIter, _ := outputsValue.Fields()
			for outputsIter.Next() {
				key := outputsIter.Label()
				val := outputsIter.Value()
				
				if goVal, err := ee.cueValueToInterface(val); err == nil {
					outputsMap[key] = goVal
				}
			}
			step.Outputs = outputsMap
		}
		
		// Extract dependencies
		if depsValue := stepValue.LookupPath(cue.ParsePath("depends_on")); depsValue.Exists() {
			var deps []string
			depsIter, _ := depsValue.List()
			for depsIter.Next() {
				if depStr, err := depsIter.Value().String(); err == nil {
					deps = append(deps, depStr)
				}
			}
			step.DependsOn = deps
		}
		
		steps = append(steps, step)
	}
	
	return &ParsedWorkflow{
		Name:        name,
		Description: description,
		Steps:       steps,
	}, nil
}

// cueValueToInterface converts a CUE value to a Go interface{}
func (ee *ExecutionEngine) cueValueToInterface(val cue.Value) (interface{}, error) {
	switch val.Kind() {
	case cue.StringKind:
		return val.String()
	case cue.IntKind:
		return val.Int64()
	case cue.FloatKind:
		return val.Float64()
	case cue.BoolKind:
		return val.Bool()
	case cue.ListKind:
		var list []interface{}
		iter, _ := val.List()
		for iter.Next() {
			if item, err := ee.cueValueToInterface(iter.Value()); err == nil {
				list = append(list, item)
			}
		}
		return list, nil
	case cue.StructKind:
		obj := make(map[string]interface{})
		iter, _ := val.Fields()
		for iter.Next() {
			key := iter.Label()
			if item, err := ee.cueValueToInterface(iter.Value()); err == nil {
				obj[key] = item
			}
		}
		return obj, nil
	default:
		// For other types, try to get as string
		if str, err := val.String(); err == nil {
			return str, nil
		}
		return nil, fmt.Errorf("unsupported CUE value kind: %v", val.Kind())
	}
}

// ExecuteWorkflow executes a prepared workflow plan
func (ee *ExecutionEngine) ExecuteWorkflow(plan *ExecutionPlan) error {
	log.Printf("[ExecutionEngine] === STARTING WORKFLOW EXECUTION ===")
	log.Printf("[ExecutionEngine] Workflow: %s (%s)", plan.Name, plan.Description)
	log.Printf("[ExecutionEngine] Total steps: %d", len(plan.ResolvedSteps))
	
	if len(plan.ValidationErrors) > 0 {
		log.Printf("[ExecutionEngine] ERROR: Workflow has validation errors: %v", plan.ValidationErrors)
		return fmt.Errorf("workflow has validation errors: %v", plan.ValidationErrors)
	}

	// Execute steps in dependency order
	for i := range plan.ResolvedSteps {
		step := &plan.ResolvedSteps[i]
		
		log.Printf("[ExecutionEngine] === EXECUTING STEP %d/%d ===", i+1, len(plan.ResolvedSteps))
		log.Printf("[ExecutionEngine] Step ID: %s", step.ID)
		log.Printf("[ExecutionEngine] Step Name: %s", step.Name)
		log.Printf("[ExecutionEngine] Service: %s", step.Service)
		log.Printf("[ExecutionEngine] Action: %s", step.Action)
		log.Printf("[ExecutionEngine] Dependencies: %v", step.DependsOn)
		log.Printf("[ExecutionEngine] Inputs: %+v", step.Inputs)
		
		// Check dependencies
		if !ee.areDependenciesMet(step.DependsOn, plan.ResolvedSteps) {
			log.Printf("[ExecutionEngine] ERROR: Dependencies not met for step %s", step.ID)
			step.Status = "failed"
			return fmt.Errorf("dependencies not met for step %s", step.ID)
		}
		
		log.Printf("[ExecutionEngine] Dependencies satisfied, executing step...")

		// Execute step via MCP service
		err := ee.executeStep(step, plan.ParameterContext)
		if err != nil {
			log.Printf("[ExecutionEngine] ERROR: Step %s failed: %v", step.ID, err)
			step.Status = "failed"
			return fmt.Errorf("step %s failed: %w", step.ID, err)
		}

		step.Status = "completed"
		log.Printf("[ExecutionEngine] SUCCESS: Step %s completed", step.ID)
		log.Printf("[ExecutionEngine] Step outputs: %+v", step.Outputs)
	}

	log.Printf("[ExecutionEngine] === WORKFLOW EXECUTION COMPLETED SUCCESSFULLY ===")
	log.Printf("[ExecutionEngine] All %d steps completed", len(plan.ResolvedSteps))
	return nil
}

// areDependenciesMet checks if all dependencies for a step are completed
func (ee *ExecutionEngine) areDependenciesMet(dependencies []string, steps []ResolvedStep) bool {
	for _, depID := range dependencies {
		found := false
		for _, step := range steps {
			if step.ID == depID && step.Status == "completed" {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// executeStep executes a single workflow step via MCP service
func (ee *ExecutionEngine) executeStep(step *ResolvedStep, context *ParameterContext) error {
	log.Printf("[ExecutionEngine] executeStep: Starting execution for step %s", step.ID)
	step.Status = "running"
	
	// Get OAuth token from context (should be passed from user authentication)
	oauthToken, ok := context.SystemParameters["oauth_token"].(string)
	if !ok || oauthToken == "" {
		log.Printf("[ExecutionEngine] executeStep: ERROR - Missing OAuth token for step %s", step.ID)
		return fmt.Errorf("missing OAuth token for MCP service execution")
	}
	
	log.Printf("[ExecutionEngine] executeStep: OAuth token found, calling MCP service...")
	log.Printf("[ExecutionEngine] executeStep: Service=%s, Action=%s", step.Service, step.Action)
	log.Printf("[ExecutionEngine] executeStep: Input parameters (before resolution): %+v", step.Inputs)
	
	// Resolve parameter references in step inputs at runtime
	resolvedInputs, err := ee.resolveStepInputs(step.Inputs, context)
	if err != nil {
		log.Printf("[ExecutionEngine] executeStep: ERROR - Parameter resolution failed for step %s: %v", step.ID, err)
		return fmt.Errorf("parameter resolution failed: %w", err)
	}
	log.Printf("[ExecutionEngine] executeStep: Input parameters (after resolution): %+v", resolvedInputs)
	
	// Log the resolved inputs being sent to MCP for debugging
	log.Printf("[ExecutionEngine] executeStep: Sending parameters to MCP service %s.%s:", step.Service, step.Action)
	for key, value := range resolvedInputs {
		log.Printf("[ExecutionEngine] executeStep:   %s: %v", key, value)
	}

	// Execute the MCP action
	response, err := ee.mcpService.ExecuteAction(step.Service, step.Action, resolvedInputs, oauthToken)
	if err != nil {
		log.Printf("[ExecutionEngine] executeStep: ERROR - MCP action execution failed for step %s: %v", step.ID, err)
		return fmt.Errorf("MCP action execution failed: %w", err)
	}
	
	log.Printf("[ExecutionEngine] executeStep: MCP service call successful for step %s", step.ID)
	log.Printf("[ExecutionEngine] executeStep: Response success: %t", response.Success)
	log.Printf("[ExecutionEngine] executeStep: Response data: %+v", response.Data)
	log.Printf("[ExecutionEngine] executeStep: Response error: %s", response.Error)
	
	// Validate and update step outputs with MCP response data
	if response.Data != nil {
		log.Printf("[ExecutionEngine] executeStep: Updating step outputs with %d data fields", len(response.Data))
		
		// Validate response against expected output schema
		if err := ee.validateResponseSchema(step.Service, step.Action, response.Data); err != nil {
			log.Printf("[ExecutionEngine] executeStep: WARNING - Response schema validation failed for step %s: %v", step.ID, err)
			// Continue execution but log validation warning for observability
		}
		
		for key, value := range response.Data {
			step.Outputs[key] = value
			log.Printf("[ExecutionEngine] executeStep: Set output %s = %v", key, value)
		}
		
		// Update context for next steps
		if context.StepOutputs[step.ID] == nil {
			context.StepOutputs[step.ID] = make(map[string]interface{})
		}
		stepOutputs := context.StepOutputs[step.ID].(map[string]interface{})
		for key, value := range response.Data {
			stepOutputs[key] = value
		}
		log.Printf("[ExecutionEngine] executeStep: Updated context with step outputs for %s", step.ID)
		log.Printf("[ExecutionEngine] executeStep: Available step outputs in context:")
		for stepID, outputs := range context.StepOutputs {
			if outputMap, ok := outputs.(map[string]interface{}); ok {
				for outputKey, outputValue := range outputMap {
					log.Printf("[ExecutionEngine] executeStep:   %s.%s = %v", stepID, outputKey, outputValue)
				}
			}
		}
	} else {
		log.Printf("[ExecutionEngine] executeStep: No data returned from MCP service for step %s", step.ID)
	}
	
	log.Printf("[ExecutionEngine] executeStep: Step %s execution completed successfully", step.ID)
	return nil
}

// resolveStepInputs resolves parameter references in step inputs at runtime
func (ee *ExecutionEngine) resolveStepInputs(inputs map[string]interface{}, context *ParameterContext) (map[string]interface{}, error) {
	resolved := make(map[string]interface{})
	
	for key, value := range inputs {
		resolvedValue, err := ee.resolveParameterValue(value, context)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve parameter %s: %w", key, err)
		}
		resolved[key] = resolvedValue
	}
	
	return resolved, nil
}

// validateResponseSchema validates MCP response data against expected output schema
func (ee *ExecutionEngine) validateResponseSchema(service, action string, responseData map[string]interface{}) error {
	// Get MCP service catalog for validation
	mcpCatalog, err := ee.mcpService.GetServiceCatalog()
	if err != nil {
		return fmt.Errorf("failed to get MCP catalog for response validation: %w", err)
	}
	
	// Check if service exists
	serviceDefinition, exists := mcpCatalog.Providers.Workspace.Services[service]
	if !exists {
		return fmt.Errorf("service '%s' not found in MCP catalog", service)
	}
	
	// Check if function exists
	functionSchema, exists := serviceDefinition.Functions[action]
	if !exists {
		return fmt.Errorf("function '%s' not found in service '%s'", action, service)
	}
	
	// If no output schema defined, skip validation (backward compatibility)
	if functionSchema.OutputSchema == nil || functionSchema.OutputSchema.Properties == nil {
		log.Printf("[ExecutionEngine] validateResponseSchema: No output schema defined for %s.%s, skipping validation", service, action)
		return nil
	}
	
	// Validate response fields against output schema
	var missingFields []string
	var unexpectedFields []string
	
	// Check for expected fields in response
	for expectedField := range functionSchema.OutputSchema.Properties {
		if _, exists := responseData[expectedField]; !exists {
			missingFields = append(missingFields, expectedField)
		}
	}
	
	// Check for unexpected fields in response (informational only)
	for responseField := range responseData {
		if _, exists := functionSchema.OutputSchema.Properties[responseField]; !exists {
			unexpectedFields = append(unexpectedFields, responseField)
		}
	}
	
	// Log validation results for observability
	if len(missingFields) > 0 {
		log.Printf("[ExecutionEngine] validateResponseSchema: Missing expected fields for %s.%s: %v", service, action, missingFields)
	}
	if len(unexpectedFields) > 0 {
		log.Printf("[ExecutionEngine] validateResponseSchema: Unexpected fields for %s.%s: %v", service, action, unexpectedFields)
	}
	
	// Return error only for missing critical fields (non-blocking for PoC)
	if len(missingFields) > 0 {
		return fmt.Errorf("response missing expected fields: %v", missingFields)
	}
	
	log.Printf("[ExecutionEngine] validateResponseSchema: Response schema validation passed for %s.%s", service, action)
	return nil
}



// sanitizeCUEContent removes illegal characters and formatting from CUE content
func (ee *ExecutionEngine) sanitizeCUEContent(cueContent string) string {
	// Remove backticks that cause CUE parsing errors
	sanitized := strings.ReplaceAll(cueContent, "`", "'")
	
	// Remove any markdown code block markers that might have been generated
	sanitized = strings.ReplaceAll(sanitized, "```cue", "")
	sanitized = strings.ReplaceAll(sanitized, "```", "")
	
	// Remove any other problematic characters that could cause CUE parsing issues
	sanitized = strings.ReplaceAll(sanitized, "\u0060", "'") // Unicode backtick
	
	return sanitized
}

// removePackageDeclaration removes the package declaration from CUE content to avoid conflicts
func (ee *ExecutionEngine) removePackageDeclaration(cueContent string) string {
	lines := strings.Split(cueContent, "\n")
	var filteredLines []string
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip package declarations
		if strings.HasPrefix(trimmed, "package ") {
			continue
		}
		filteredLines = append(filteredLines, line)
	}
	
	return strings.Join(filteredLines, "\n")
}


