package services

import (
	"log"
	"sohoaas-backend/internal/types"
)

// ExecutePersonalCapabilitiesAgent executes the Personal Capabilities Agent
func (g *GenkitService) ExecutePersonalCapabilitiesAgent(input map[string]interface{}) (*types.AgentResponse, error) {
	// Execute the pre-defined flow (uses inline prompts for now)
	result, err := g.personalCapabilitiesFlow.Run(g.ctx, input)
	if err != nil {
		return &types.AgentResponse{
			AgentID: "personal_capabilities",
			Error:   err.Error(),
		}, nil
	}

	return &types.AgentResponse{
		AgentID: "personal_capabilities",
		Output:  result,
	}, nil
}

// ExecuteIntentGathererAgent executes the Intent Gatherer Agent
func (g *GenkitService) ExecuteIntentGathererAgent(input map[string]interface{}) (*types.AgentResponse, error) {
	// Execute the pre-defined flow (uses inline prompts for now)
	result, err := g.intentGathererFlow.Run(g.ctx, input)
	if err != nil {
		return &types.AgentResponse{
			AgentID: "intent_gatherer",
			Error:   err.Error(),
		}, nil
	}

	return &types.AgentResponse{
		AgentID: "intent_gatherer",
		Output:  result,
	}, nil
}

// ExecuteIntentAnalystAgent executes the Intent Analyst Agent
func (g *GenkitService) ExecuteIntentAnalystAgent(input map[string]interface{}) (*types.AgentResponse, error) {
	// Convert map[string]interface{} input to simplified IntentAnalystInput
	typedInput := IntentAnalystInput{}

	// Extract user message from workflow_intent (proper Go type handling)
	if workflowIntent, ok := input["workflow_intent"].(*types.WorkflowIntent); ok {
		typedInput.UserMessage = workflowIntent.UserMessage
	}

	// Extract available services from service_schemas (proper Go type handling)
	availableServices := []string{}
	if serviceSchemas, ok := input["service_schemas"].(map[string]types.ServiceSchema); ok {
		for serviceName := range serviceSchemas {
			availableServices = append(availableServices, serviceName)
		}
	}
	typedInput.AvailableServices = availableServices

	log.Printf("[DEBUG] ExecuteIntentAnalystAgent: Simplified input: %+v", typedInput)

	// Execute the pre-defined flow with typed input
	result, err := g.intentAnalystFlow.Run(g.ctx, typedInput)
	if err != nil {
		return &types.AgentResponse{
			AgentID: "intent_analyst",
			Error:   err.Error(),
		}, nil
	}

	// Convert typed output back to map[string]interface{} for compatibility
	outputMap := map[string]interface{}{
		"is_automation_request": result.IsAutomationRequest,
		"required_services":     result.RequiredServices,
		"can_fulfill":           result.CanFulfill,
		"missing_info":          result.MissingInfo,
		"next_action":           result.NextAction,
	}

	return &types.AgentResponse{
		AgentID: "intent_analyst",
		Output:  outputMap,
	}, nil
}
