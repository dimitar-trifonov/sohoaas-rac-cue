package storage

import (
	"encoding/json"
	"fmt"

	"cuelang.org/go/cue/cuecontext"
	"sohoaas-backend/internal/types"
)

// parseCUEWorkflow parses CUE content into structured data
func parseCUEWorkflow(cueContent string, workflow *types.WorkflowFile) (*types.WorkflowFile, error) {
	ctx := cuecontext.New()
	value := ctx.CompileString(cueContent)
	if value.Err() != nil {
		return workflow, fmt.Errorf("failed to compile CUE: %v", value.Err())
	}

	// Convert to JSON for structured data
	data, err := value.MarshalJSON()
	if err != nil {
		return workflow, fmt.Errorf("failed to marshal CUE to JSON: %v", err)
	}

	// Parse JSON into map
	var parsedData map[string]interface{}
	if err := json.Unmarshal(data, &parsedData); err != nil {
		return workflow, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	// Extract fields from nested workflow object
	if workflowData, ok := parsedData["workflow"].(map[string]interface{}); ok {
		workflow.ParsedData = workflowData
	} else {
		workflow.ParsedData = parsedData
	}

	return workflow, nil
}
