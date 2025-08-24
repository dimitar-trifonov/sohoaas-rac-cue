package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"sohoaas-backend/internal/types"
)

func TestParseExampleWorkflowCUE(t *testing.T) {
	// Locate example CUE file via RAC_CONTEXT_PATH if set
	cwd, err := os.Getwd()
	require.NoError(t, err)
	racRoot := os.Getenv("RAC_CONTEXT_PATH")
	if racRoot == "" {
		// Fallback: from app/backend/internal/storage to repo root is four levels up
		racRoot = filepath.Clean(filepath.Join(cwd, "../../../..", "rac"))
	}
	examplePath := filepath.Clean(filepath.Join(racRoot, "schemas", "examples", "example_workflow.cue"))
	content, err := os.ReadFile(examplePath)
	require.NoError(t, err, "failed to read example workflow cue at %s", examplePath)

	wf := &types.WorkflowFile{}
	parsed, err := parseCUEWorkflow(string(content), wf)
	require.NoError(t, err)
	require.NotNil(t, parsed)

	// Basic assertions on ParsedData
	require.NotNil(t, parsed.ParsedData)
	if name, ok := parsed.ParsedData["name"].(string); ok {
		require.NotEmpty(t, name)
	} else {
		t.Fatalf("missing name in ParsedData")
	}

	// User parameters should be present and include expected keys
	up, ok := parsed.ParsedData["user_parameters"].(map[string]interface{})
	require.True(t, ok)
	_, hasDoc := up["document_title"]
	_, hasStart := up["event_start_time"]
	_, hasEnd := up["event_end_time"]
	require.True(t, hasDoc && hasStart && hasEnd)

	// Steps should be an array with at least 2 entries
	steps, ok := parsed.ParsedData["steps"].([]interface{})
	require.True(t, ok)
	require.GreaterOrEqual(t, len(steps), 2)
}
