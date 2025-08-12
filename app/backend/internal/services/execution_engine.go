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
func (ee *ExecutionEngine) ValidateWorkflowServices(workflow *ParsedWorkflow) error {
	if workflow == nil {
		return fmt.Errorf("workflow is nil")
	}
	
	// Query live MCP service catalog for validation using centralized MCPService
	mcpServices, err := ee.mcpService.GetServiceCatalog()
	if err != nil {
		return fmt.Errorf("failed to query MCP service catalog for validation: %w", err)
	}
	
	// Use centralized parser to validate workflow services
	return ee.mcpParser.ValidateWorkflowServices(mcpServices, workflow)
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
func (ee *ExecutionEngine) PrepareExecution(cueworkflow string, userID string, user *types.User, intentAnalysis map[string]interface{}, oauthToken string) (*ExecutionPlan, error) {
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
	paramContext := ee.createParameterContext(intentAnalysis, user, oauthToken)

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
func (ee *ExecutionEngine) createParameterContext(intentAnalysis map[string]interface{}, user *types.User, oauthToken string) *ParameterContext {
	context := &ParameterContext{
		UserParameters:    make(map[string]interface{}),
		RuntimeParameters: make(map[string]interface{}),
		SystemParameters:  make(map[string]interface{}),
		StepOutputs:       make(map[string]interface{}),
	}

	// Extract user parameters from intent analysis
	if userParams, ok := intentAnalysis["user_parameters"].([]interface{}); ok {
		for _, param := range userParams {
			if paramMap, ok := param.(map[string]interface{}); ok {
				if name, exists := paramMap["name"].(string); exists {
					// Set default values or collect from user input
					context.UserParameters[name] = ee.resolveUserParameter(paramMap, user)
				}
			}
		}
	}

	// Set system parameters
	context.SystemParameters["current_date"] = time.Now().Format("2006-01-02")
	context.SystemParameters["current_datetime"] = time.Now().Format("2006-01-02T15:04:05")
	context.SystemParameters["user_email"] = user.Email
	context.SystemParameters["user_id"] = user.ID
	context.SystemParameters["oauth_token"] = oauthToken

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
		return fmt.Sprintf("${USER_INPUT:%s}", paramName)
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
	if strValue, ok := value.(string); ok {
		return ee.resolveStringParameter(strValue, context)
	}
	return value, nil
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

	// Handle user parameter references
	if strings.HasPrefix(value, "${USER_INPUT:") && strings.HasSuffix(value, "}") {
		paramName := strings.TrimSuffix(strings.TrimPrefix(value, "${USER_INPUT:"), "}")
		if userValue, exists := context.UserParameters[paramName]; exists {
			return userValue, nil
		}
		return value, fmt.Errorf("user parameter %s not provided", paramName)
	}

	// Handle system parameter references
	if strings.HasPrefix(value, "${SYSTEM:") && strings.HasSuffix(value, "}") {
		paramName := strings.TrimSuffix(strings.TrimPrefix(value, "${SYSTEM:"), "}")
		if systemValue, exists := context.SystemParameters[paramName]; exists {
			return systemValue, nil
		}
		return value, fmt.Errorf("system parameter %s not available", paramName)
	}

	return value, nil
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
	// Create CUE context
	ctx := cuecontext.New()
	
	// Parse the CUE content
	value := ctx.CompileString(cueContent)
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
		
		// Extract service
		if serviceValue := stepValue.LookupPath(cue.ParsePath("service")); serviceValue.Exists() {
			if service, err := serviceValue.String(); err != nil {
				return nil, fmt.Errorf("failed to extract service from step %d: %w", len(steps), err)
			} else {
				step.Service = service
			}
		}
		
		// Extract action
		if actionValue := stepValue.LookupPath(cue.ParsePath("action")); actionValue.Exists() {
			if action, err := actionValue.String(); err != nil {
				return nil, fmt.Errorf("failed to extract action from step %d: %w", len(steps), err)
			} else {
				step.Action = action
			}
		}
		
		// Extract inputs
		if inputsValue := stepValue.LookupPath(cue.ParsePath("inputs")); inputsValue.Exists() {
			inputsMap := make(map[string]interface{})
			inputsIter, _ := inputsValue.Fields()
			for inputsIter.Next() {
				key := inputsIter.Label()
				val := inputsIter.Value()
				
				// Convert CUE value to Go interface{}
				if goVal, err := ee.cueValueToInterface(val); err == nil {
					inputsMap[key] = goVal
				}
			}
			step.Inputs = inputsMap
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
	log.Printf("[ExecutionEngine] executeStep: Input parameters: %+v", step.Inputs)
	
	// Execute action via MCP service
	response, err := ee.mcpService.ExecuteAction(step.Service, step.Action, step.Inputs, oauthToken)
	if err != nil {
		log.Printf("[ExecutionEngine] executeStep: ERROR - MCP action execution failed for step %s: %v", step.ID, err)
		return fmt.Errorf("MCP action execution failed: %w", err)
	}
	
	log.Printf("[ExecutionEngine] executeStep: MCP service call successful for step %s", step.ID)
	log.Printf("[ExecutionEngine] executeStep: Response success: %t", response.Success)
	log.Printf("[ExecutionEngine] executeStep: Response data: %+v", response.Data)
	log.Printf("[ExecutionEngine] executeStep: Response error: %s", response.Error)
	
	// Update step outputs with MCP response data
	if response.Data != nil {
		log.Printf("[ExecutionEngine] executeStep: Updating step outputs with %d data fields", len(response.Data))
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
	} else {
		log.Printf("[ExecutionEngine] executeStep: No data returned from MCP service for step %s", step.ID)
	}
	
	log.Printf("[ExecutionEngine] executeStep: Step %s execution completed successfully", step.ID)
	return nil
}


