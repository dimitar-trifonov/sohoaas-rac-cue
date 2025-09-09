package services

import "sohoaas-backend/internal/types"

// Simplified Intent Analyst Input/Output Types for Genkit compatibility
type IntentAnalystInput struct {
	UserMessage       string   `json:"user_message"`
	AvailableServices []string `json:"available_services"`
}

type IntentAnalystOutput struct {
	IsAutomationRequest bool     `json:"is_automation_request"`
	RequiredServices    []string `json:"required_services"`
	CanFulfill          bool     `json:"can_fulfill"`
	MissingInfo         []string `json:"missing_info"`
	NextAction          string   `json:"next_action"`
	Explanation         string   `json:"explanation"`
}

type ValidatedIntent struct {
	IsAutomationRequest bool     `json:"is_automation_request"`
	RequiredServices    []string `json:"required_services"`
	CanFulfill          bool     `json:"can_fulfill"`
	MissingInfo         []string `json:"missing_info"`
	NextAction          string   `json:"next_action"`
	Explanation         string   `json:"explanation,omitempty"`
	Confidence          float64  `json:"confidence,omitempty"`
	WorkflowPattern     string   `json:"workflow_pattern,omitempty"`
}

type WorkflowGeneratorInput struct {
	UserIntent        string          `json:"user_intent"`
	ValidatedIntent   ValidatedIntent `json:"validated_intent"`
	AvailableServices string          `json:"available_services"`
	RacContext        string          `json:"rac_context"`
}

type WorkflowGeneratorOutput struct {
	Version        string                         `json:"version"`
	Name           string                         `json:"name"`
	Description    string                         `json:"description"`
	OriginalIntent string                         `json:"original_intent"`
	Steps          []types.WorkflowStep           `json:"steps"`
	UserParameters map[string]types.UserParameter `json:"user_parameters"`
	Services       map[string]interface{}         `json:"services"`
	DecisionLog    *DecisionLog                   `json:"decision_log,omitempty"`
}

// DecisionLog captures a concise, non-authoritative trace emitted by the LLM
// for debugging. It should contain short reason codes and references to RaC
// patterns, not chain-of-thought.
type DecisionLog struct {
	Summary string             `json:"summary"`
	Events  []DecisionLogEvent `json:"events"`
}

type DecisionLogEvent struct {
	EventID         string                 `json:"event_id"`
	StepID          string                 `json:"step_id"`
	Action          string                 `json:"action"`
	ReasonCode      string                 `json:"reason_code"`
	RacRefs         []string               `json:"rac_refs,omitempty"`
	InputsMapped    map[string]interface{} `json:"inputs_mapped,omitempty"`
	ValidationFlags []string               `json:"validation_flags,omitempty"`
}
