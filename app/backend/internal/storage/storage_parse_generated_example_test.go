package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"sohoaas-backend/internal/types"
)

func TestParseGeneratedKickoffExampleCUE(t *testing.T) {
	// Resolve example path via RAC_CONTEXT_PATH
	cwd, err := os.Getwd()
	require.NoError(t, err)
	racRoot := os.Getenv("RAC_CONTEXT_PATH")
	if racRoot == "" {
		// Fallback: from app/backend/internal/storage to repo root is four levels up
		racRoot = filepath.Clean(filepath.Join(cwd, "../../../..", "rac"))
	}
	examplePath := filepath.Clean(filepath.Join(racRoot, "schemas", "examples", "generated_kickoff_example.cue"))
	content, err := os.ReadFile(examplePath)
	require.NoError(t, err, "failed to read generated example workflow cue at %s", examplePath)

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

	// User parameters must include keys from the real file
	up, ok := parsed.ParsedData["user_parameters"].(map[string]interface{})
	require.True(t, ok)
	_, hasTitle := up["document_title"]
	_, hasStart := up["event_start_time"]
	_, hasEnd := up["event_end_time"]
	_, hasFolder := up["folder_name"]
	require.True(t, hasTitle && hasStart && hasEnd && hasFolder)

	// Steps should be an array with at least 3 entries
	steps, ok := parsed.ParsedData["steps"].([]interface{})
	require.True(t, ok)
	require.GreaterOrEqual(t, len(steps), 3)

	// Dump ParsedData for inspection
	if b, err := json.MarshalIndent(parsed.ParsedData, "", "  "); err == nil {
		t.Logf("ParsedData (generated_kickoff_example):\n%s", string(b))
	} else {
		t.Logf("ParsedData dump failed: %v", err)
	}
}
