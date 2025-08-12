package services

import (
	"io/ioutil"
	"testing"
)

func TestParseCUEWorkflow(t *testing.T) {
	// Read the new test workflow file
	cueContent, err := ioutil.ReadFile("../../workflows/authenticated_user/workflow_weekly_reports_test.cue")
	if err != nil {
		t.Fatalf("Failed to read CUE file: %v", err)
	}

	t.Logf("=== CUE FILE CONTENT ===\n%s", string(cueContent))

	// Create execution engine for testing
	ee := NewExecutionEngine(nil) // nil MCP service for parsing test
	
	// Test the parsing
	workflow, err := ee.ParseCUEWorkflow(string(cueContent))
	if err != nil {
		t.Fatalf("Failed to parse CUE workflow: %v", err)
	}

	// Verify basic workflow structure
	if workflow.Name == "" {
		t.Error("Workflow name is empty")
	}
	if workflow.Description == "" {
		t.Error("Workflow description is empty")
	}
	if len(workflow.Steps) == 0 {
		t.Error("No workflow steps found")
	}

	t.Logf("=== PARSED WORKFLOW ===")
	t.Logf("Name: %s", workflow.Name)
	t.Logf("Description: %s", workflow.Description)
	t.Logf("Number of Steps: %d", len(workflow.Steps))

	// Check each step for required fields
	for i, step := range workflow.Steps {
		t.Logf("\nStep %d:", i)
		t.Logf("  ID: '%s'", step.ID)
		t.Logf("  Name: '%s'", step.Name)
		t.Logf("  Service: '%s'", step.Service)
		t.Logf("  Action: '%s'", step.Action)
		t.Logf("  Inputs: %+v", step.Inputs)
		t.Logf("  Outputs: %+v", step.Outputs)

		// Verify critical fields are not empty
		if step.ID == "" {
			t.Errorf("Step %d has empty ID", i)
		}
		if step.Service == "" {
			t.Errorf("Step %d has empty Service", i)
		}
		if step.Action == "" {
			t.Errorf("Step %d has empty Action", i)
		}
	}
}
