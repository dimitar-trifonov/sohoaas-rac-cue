package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testWorkflowCUE = `package workflow

workflow: {
	version: "1.0"
	name: "Test Workflow"
	description: "Integration test workflow"
	user_parameters: {
		test_param: {
			type: "string"
			required: true
		}
	}
	steps: [{
		id: "test_step"
		name: "Test Step"
		service: "gmail"
		action: "send_message"
		parameters: {
			to: "${user.test_param}"
			subject: "Test Email"
			body: "This is a test"
		}
	}]
}`

func TestStorageImplementations(t *testing.T) {
	// Setup test storages
	localConfig := LocalStorageConfig{
		WorkflowsDir: t.TempDir(),
	}
	localStorage, err := NewLocalStorage(localConfig)
	require.NoError(t, err)

	mockStorage := NewMockStorage()

	// Test all storage implementations
	storages := []struct {
		name    string
		storage WorkflowStorage
	}{
		{"LocalStorage", localStorage},
		{"MockStorage", mockStorage},
	}

	for _, s := range storages {
		t.Run(s.name, func(t *testing.T) {
			// Save workflow
			workflow, err := s.storage.SaveWorkflow("test_user", "test_workflow", testWorkflowCUE)
			require.NoError(t, err)
			require.NotNil(t, workflow)

			// Verify ParsedData is populated
			assert.NotNil(t, workflow.ParsedData, "ParsedData should be populated")
			assert.Contains(t, workflow.ParsedData, "version")
			assert.Contains(t, workflow.ParsedData, "name")
			assert.Contains(t, workflow.ParsedData, "user_parameters")
			assert.Contains(t, workflow.ParsedData, "steps")

			// Verify steps structure
			steps, ok := workflow.ParsedData["steps"].([]interface{})
			require.True(t, ok, "steps should be an array")
			require.Len(t, steps, 1)

			step := steps[0].(map[string]interface{})
			assert.Equal(t, "test_step", step["id"])
			assert.Equal(t, "gmail", step["service"])
			assert.Equal(t, "send_message", step["action"])

			// Verify user parameters
			params, ok := workflow.ParsedData["user_parameters"].(map[string]interface{})
			require.True(t, ok, "user_parameters should be a map")
			testParam, ok := params["test_param"].(map[string]interface{})
			require.True(t, ok, "test_param should be a map")
			assert.Equal(t, "string", testParam["type"])
			assert.Equal(t, true, testParam["required"])

			// Get workflow and verify ParsedData is still present
			retrieved, err := s.storage.GetWorkflow("test_user", workflow.ID)
			require.NoError(t, err)
			require.NotNil(t, retrieved)
			assert.NotNil(t, retrieved.ParsedData, "ParsedData should be present in retrieved workflow")
			assert.Equal(t, workflow.ParsedData, retrieved.ParsedData)

			// List workflows and verify ParsedData is present
			workflows, err := s.storage.ListUserWorkflows("test_user")
			require.NoError(t, err)
			require.Len(t, workflows, 1)
			assert.NotNil(t, workflows[0].ParsedData, "ParsedData should be present in listed workflow")
			assert.Equal(t, workflow.ParsedData, workflows[0].ParsedData)
		})
	}
}
