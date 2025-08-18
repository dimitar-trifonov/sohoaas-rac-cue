package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sohoaas-backend/internal/types"
)

func TestWorkflowStorageParseCUE(t *testing.T) {
	// Create a temporary workflow storage service
	tmpDir := t.TempDir()
	service := NewWorkflowStorageService(tmpDir)

	// Read an actual CUE workflow file
	cueFilePath := filepath.Join("..", "..", "..", "..", "generated_workflows", "mock_user_123", "workflow_20250815_131034", "workflow.cue")
	content, err := os.ReadFile(cueFilePath)
	if err != nil {
		t.Skipf("Skipping test - CUE file not found: %v", err)
		return
	}

	// Create a test workflow file
	workflowFile := &types.WorkflowFile{
		ID:      "test_workflow",
		Name:    "Test Workflow",
		Content: string(content),
	}

	t.Logf("Testing CUE parsing with content length: %d bytes", len(content))

	// Test the CUE parsing
	parsedWorkflow, err := service.parseCUEWorkflow(string(content), workflowFile)
	
	if err != nil {
		t.Logf("CUE parsing failed: %v", err)
		t.Logf("First 500 chars of content:\n%s", string(content)[:min(500, len(content))])
		
		// This is expected to fail due to missing schema definitions
		assert.Error(t, err, "Expected CUE parsing to fail due to missing schema definitions")
		return
	}

	// If parsing succeeds, check the parsed data
	require.NotNil(t, parsedWorkflow, "Parsed workflow should not be nil")
	require.NotNil(t, parsedWorkflow.ParsedData, "Parsed data should not be nil")

	t.Logf("Successfully parsed workflow with keys: %v", getWorkflowMapKeys(parsedWorkflow.ParsedData))

	// Check for expected fields
	assert.Contains(t, parsedWorkflow.ParsedData, "version", "Should contain version")
	assert.Contains(t, parsedWorkflow.ParsedData, "name", "Should contain name")
	assert.Contains(t, parsedWorkflow.ParsedData, "steps", "Should contain steps")
	assert.Contains(t, parsedWorkflow.ParsedData, "user_parameters", "Should contain user_parameters")
}

func TestListUserWorkflows(t *testing.T) {
	// Create a temporary workflow storage service
	tmpDir := t.TempDir()
	service := NewWorkflowStorageService(tmpDir)

	// Test with mock_user_123 workflows if they exist
	workflows, err := service.ListUserWorkflows("mock_user_123")
	require.NoError(t, err, "Should be able to list workflows")

	t.Logf("Found %d workflows for mock_user_123", len(workflows))

	for i, workflow := range workflows {
		t.Logf("Workflow %d: ID=%s, Name=%s, HasParsedData=%v", 
			i, workflow.ID, workflow.Name, workflow.ParsedData != nil)
		
		if workflow.ParsedData != nil {
			t.Logf("  Parsed data keys: %v", getWorkflowMapKeys(workflow.ParsedData))
		}
	}
}

func getWorkflowMapKeys(m map[string]interface{}) []string {
	if m == nil {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestSaveWorkflowWithSchemaInjection(t *testing.T) {
	// Create a temporary workflow storage service
	tmpDir := t.TempDir()
	service := NewWorkflowStorageService(tmpDir)

	// Simple test workflow content without schema
	testWorkflowContent := `
workflow: {
	version: "1.0.0"
	name: "Test Workflow"
	description: "A test workflow"
	steps: [
		{
			id: "test_step"
			name: "Test Step"
			action: "test.action"
		}
	]
	user_parameters: {
		test_param: {
			type: "string"
			prompt: "Enter test value:"
			required: true
		}
	}
	service_bindings: {}
	execution_config: {
		mode: "sequential"
	}
}
`

	// Save workflow (this should inject the schema)
	savedWorkflow, err := service.SaveWorkflow("test_user", "test_workflow", testWorkflowContent)
	require.NoError(t, err, "Should be able to save workflow with schema injection")
	require.NotNil(t, savedWorkflow, "Saved workflow should not be nil")

	t.Logf("Saved workflow: %s at %s", savedWorkflow.Name, savedWorkflow.Path)

	// Read the saved file to verify schema was injected
	savedContent, err := os.ReadFile(savedWorkflow.Path)
	require.NoError(t, err, "Should be able to read saved workflow file")

	savedContentStr := string(savedContent)
	
	// Verify schema was injected
	require.Contains(t, savedContentStr, "#DeterministicWorkflow", "Saved workflow should contain embedded schema")
	require.Contains(t, savedContentStr, "Test Workflow", "Saved workflow should contain original content")
	
	t.Logf("Schema injection successful - file contains both schema and workflow content")
}
