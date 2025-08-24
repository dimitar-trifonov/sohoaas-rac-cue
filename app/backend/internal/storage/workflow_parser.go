package storage

import (
	"encoding/json"
	"fmt"
	"strings"

	"cuelang.org/go/cue/cuecontext"
	"sohoaas-backend/internal/types"
)

// parseCUEWorkflow parses CUE content into structured data
func parseCUEWorkflow(cueContent string, workflow *types.WorkflowFile) (*types.WorkflowFile, error) {
	// Sanitize: strip import declarations to avoid resolution errors during parsing-only use.
	sanitized := sanitizeCUEForParsing(cueContent)

	ctx := cuecontext.New()
	value := ctx.CompileString(sanitized)
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

// sanitizeCUEForParsing removes import declarations (single-line and block) so that
// CompileString doesn't require resolving external files when we only need a JSON view.
func sanitizeCUEForParsing(src string) string {
    lines := strings.Split(src, "\n")
    var out []string
    inImportBlock := false
    for _, ln := range lines {
        trimmed := strings.TrimSpace(ln)
        if inImportBlock {
            // End of import block
            if strings.HasSuffix(trimmed, ")") {
                inImportBlock = false
            }
            continue
        }
        // Start of multi-line import block: import (
        if trimmed == "import (" {
            inImportBlock = true
            continue
        }
        // Single-line import "path" or import alias "path"
        if strings.HasPrefix(trimmed, "import ") {
            continue
        }
        out = append(out, ln)
    }
    sanitized := strings.Join(out, "\n")
    // Also neutralize deterministic schema type conjunction to avoid unresolved refs:
    //   workflow: #DeterministicWorkflow & { ... }  ->  workflow: { ... }
    sanitized = strings.ReplaceAll(sanitized, "#DeterministicWorkflow &", "")
    return sanitized
}
