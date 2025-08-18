package services

import (
	"strings"
	"testing"
)

func TestResolveStepInputs(t *testing.T) {
	// Create execution engine instance
	executionEngine := &ExecutionEngine{}

	tests := []struct {
		name           string
		inputs         map[string]interface{}
		context        *ParameterContext
		expectedOutput map[string]interface{}
		expectError    bool
		errorContains  string
	}{
		{
			name: "Resolve user parameter references",
			inputs: map[string]interface{}{
				"title":       "${user.document_title}",
				"folder_name": "${user.folder_name}",
				"static_val":  "unchanged",
			},
			context: &ParameterContext{
				UserParameters: map[string]interface{}{
					"document_title": "My Important Document",
					"folder_name":    "Work Documents",
				},
				StepOutputs:       make(map[string]interface{}),
				SystemParameters:  make(map[string]interface{}),
				RuntimeParameters: make(map[string]interface{}),
			},
			expectedOutput: map[string]interface{}{
				"title":       "My Important Document",
				"folder_name": "Work Documents",
				"static_val":  "unchanged",
			},
			expectError: false,
		},
		{
			name: "Resolve step output references",
			inputs: map[string]interface{}{
				"file_id":        "${steps.create_document.outputs.document_id}",
				"parent_folder":  "${steps.create_folder.outputs.folder_id}",
				"mixed_content":  "Document ${steps.create_document.outputs.document_id} in folder",
			},
			context: &ParameterContext{
				UserParameters: make(map[string]interface{}),
				StepOutputs: map[string]interface{}{
					"create_document": map[string]interface{}{
						"document_id": "1BxY8Z9AbCdEfGhIjKlMnOpQrStUvWxYz",
						"document_url": "https://docs.google.com/document/d/1BxY8Z9AbCdEfGhIjKlMnOpQrStUvWxYz/edit",
					},
					"create_folder": map[string]interface{}{
						"folder_id": "1FoLdEr9AbCdEfGhIjKlMnOpQrStUvWxYz",
						"folder_url": "https://drive.google.com/drive/folders/1FoLdEr9AbCdEfGhIjKlMnOpQrStUvWxYz",
					},
				},
				SystemParameters:  make(map[string]interface{}),
				RuntimeParameters: make(map[string]interface{}),
			},
			expectedOutput: map[string]interface{}{
				"file_id":        "1BxY8Z9AbCdEfGhIjKlMnOpQrStUvWxYz",
				"parent_folder":  "1FoLdEr9AbCdEfGhIjKlMnOpQrStUvWxYz",
				"mixed_content":  "Document 1BxY8Z9AbCdEfGhIjKlMnOpQrStUvWxYz in folder",
			},
			expectError: false,
		},
		{
			name: "Resolve mixed parameter types",
			inputs: map[string]interface{}{
				"user_email":     "${user.email}",
				"document_id":    "${steps.create_doc.outputs.id}",
				"system_token":   "${oauth_token}",
				"literal_value":  "no parameters here",
			},
			context: &ParameterContext{
				UserParameters: map[string]interface{}{
					"email": "user@example.com",
				},
				StepOutputs: map[string]interface{}{
					"create_doc": map[string]interface{}{
						"id": "doc123456789",
					},
				},
				SystemParameters: map[string]interface{}{
					"oauth_token": "ya29.token_value_here",
				},
				RuntimeParameters: make(map[string]interface{}),
			},
			expectedOutput: map[string]interface{}{
				"user_email":     "user@example.com",
				"document_id":    "doc123456789",
				"system_token":   "ya29.token_value_here",
				"literal_value":  "no parameters here",
			},
			expectError: false,
		},
		{
			name: "Handle nested objects and arrays",
			inputs: map[string]interface{}{
				"config": map[string]interface{}{
					"title":     "${user.title}",
					"folder_id": "${steps.folder.outputs.id}",
					"settings": map[string]interface{}{
						"owner": "${user.email}",
					},
				},
				"recipients": []interface{}{
					"${user.email}",
					"admin@company.com",
					"${user.manager_email}",
				},
			},
			context: &ParameterContext{
				UserParameters: map[string]interface{}{
					"title":         "Project Report",
					"email":         "john@company.com",
					"manager_email": "manager@company.com",
				},
				StepOutputs: map[string]interface{}{
					"folder": map[string]interface{}{
						"id": "folder_abc123",
					},
				},
				SystemParameters:  make(map[string]interface{}),
				RuntimeParameters: make(map[string]interface{}),
			},
			expectedOutput: map[string]interface{}{
				"config": map[string]interface{}{
					"title":     "Project Report",
					"folder_id": "folder_abc123",
					"settings": map[string]interface{}{
						"owner": "john@company.com",
					},
				},
				"recipients": []interface{}{
					"john@company.com",
					"admin@company.com",
					"manager@company.com",
				},
			},
			expectError: false,
		},
		{
			name: "Error on missing user parameter",
			inputs: map[string]interface{}{
				"title": "${user.missing_param}",
			},
			context: &ParameterContext{
				UserParameters:    map[string]interface{}{},
				StepOutputs:       make(map[string]interface{}),
				SystemParameters:  make(map[string]interface{}),
				RuntimeParameters: make(map[string]interface{}),
			},
			expectedOutput: nil,
			expectError:    true,
			errorContains:  "user parameter missing_param not provided",
		},
		{
			name: "Step output reference during validation phase (empty StepOutputs)",
			inputs: map[string]interface{}{
				"file_id": "${steps.missing_step.outputs.id}",
			},
			context: &ParameterContext{
				UserParameters:    make(map[string]interface{}),
				StepOutputs:       make(map[string]interface{}), // Empty - validation phase
				SystemParameters:  make(map[string]interface{}),
				RuntimeParameters: make(map[string]interface{}),
			},
			expectedOutput: map[string]interface{}{
				"file_id": "${steps.missing_step.outputs.id}", // Unresolved during validation
			},
			expectError: false, // No error during validation phase
		},
		{
			name: "Error on missing step output during execution phase",
			inputs: map[string]interface{}{
				"file_id": "${steps.missing_step.outputs.id}",
			},
			context: &ParameterContext{
				UserParameters: make(map[string]interface{}),
				StepOutputs: map[string]interface{}{
					"existing_step": map[string]interface{}{
						"some_output": "value",
					},
					// missing_step not present - this should error during execution
				},
				SystemParameters:  make(map[string]interface{}),
				RuntimeParameters: make(map[string]interface{}),
			},
			expectedOutput: nil,
			expectError:    true,
			errorContains:  "step output reference missing_step.id not available",
		},
		{
			name: "Handle primitive types unchanged",
			inputs: map[string]interface{}{
				"number":  42,
				"boolean": true,
				"float":   3.14,
				"null":    nil,
			},
			context: &ParameterContext{
				UserParameters:    make(map[string]interface{}),
				StepOutputs:       make(map[string]interface{}),
				SystemParameters:  make(map[string]interface{}),
				RuntimeParameters: make(map[string]interface{}),
			},
			expectedOutput: map[string]interface{}{
				"number":  42,
				"boolean": true,
				"float":   3.14,
				"null":    nil,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executionEngine.resolveStepInputs(tt.inputs, tt.context)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !deepEqual(result, tt.expectedOutput) {
				t.Errorf("Expected output:\n%+v\nGot:\n%+v", tt.expectedOutput, result)
			}
		})
	}
}

func TestResolveParameterValue(t *testing.T) {
	executionEngine := &ExecutionEngine{}

	tests := []struct {
		name           string
		value          interface{}
		context        *ParameterContext
		expectedOutput interface{}
		expectError    bool
	}{
		{
			name:  "String with user parameter",
			value: "${user.name}",
			context: &ParameterContext{
				UserParameters: map[string]interface{}{
					"name": "John Doe",
				},
				StepOutputs:       make(map[string]interface{}),
				SystemParameters:  make(map[string]interface{}),
				RuntimeParameters: make(map[string]interface{}),
			},
			expectedOutput: "John Doe",
			expectError:    false,
		},
		{
			name:  "String with step output",
			value: "${steps.create_doc.outputs.id}",
			context: &ParameterContext{
				UserParameters: make(map[string]interface{}),
				StepOutputs: map[string]interface{}{
					"create_doc": map[string]interface{}{
						"id": "doc_12345",
					},
				},
				SystemParameters:  make(map[string]interface{}),
				RuntimeParameters: make(map[string]interface{}),
			},
			expectedOutput: "doc_12345",
			expectError:    false,
		},
		{
			name:  "Literal string unchanged",
			value: "literal string",
			context: &ParameterContext{
				UserParameters:    make(map[string]interface{}),
				StepOutputs:       make(map[string]interface{}),
				SystemParameters:  make(map[string]interface{}),
				RuntimeParameters: make(map[string]interface{}),
			},
			expectedOutput: "literal string",
			expectError:    false,
		},
		{
			name:  "Number unchanged",
			value: 42,
			context: &ParameterContext{
				UserParameters:    make(map[string]interface{}),
				StepOutputs:       make(map[string]interface{}),
				SystemParameters:  make(map[string]interface{}),
				RuntimeParameters: make(map[string]interface{}),
			},
			expectedOutput: 42,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executionEngine.resolveParameterValue(tt.value, tt.context)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !deepEqual(result, tt.expectedOutput) {
				t.Errorf("Expected: %+v, Got: %+v", tt.expectedOutput, result)
			}
		})
	}
}

func TestRuntimeParameterResolutionIntegration(t *testing.T) {
	// Test the complete flow: step execution with parameter resolution
	executionEngine := &ExecutionEngine{}

	// Simulate a context with previous step outputs
	context := &ParameterContext{
		UserParameters: map[string]interface{}{
			"document_title": "Weekly Report",
			"folder_name":    "Reports",
		},
		StepOutputs: map[string]interface{}{
			"create_document": map[string]interface{}{
				"document_id":  "1BxY8Z9AbCdEfGhIjKlMnOpQrStUvWxYz",
				"document_url": "https://docs.google.com/document/d/1BxY8Z9AbCdEfGhIjKlMnOpQrStUvWxYz/edit",
			},
			"create_folder": map[string]interface{}{
				"folder_id":  "1FoLdEr9AbCdEfGhIjKlMnOpQrStUvWxYz",
				"folder_url": "https://drive.google.com/drive/folders/1FoLdEr9AbCdEfGhIjKlMnOpQrStUvWxYz",
			},
		},
		SystemParameters: map[string]interface{}{
			"oauth_token":   "ya29.token_value",
			"user_email":    "user@example.com",
			"user_timezone": "America/New_York",
		},
		RuntimeParameters: make(map[string]interface{}),
	}

	// Test step inputs that need resolution (similar to the failing step from logs)
	stepInputs := map[string]interface{}{
		"file_id":       "${steps.create_document.outputs.document_id}",
		"new_parent_id": "${steps.create_folder.outputs.folder_id}",
		"title":         "${user.document_title} - Final Version",
	}

	// Resolve the inputs
	resolvedInputs, err := executionEngine.resolveStepInputs(stepInputs, context)
	if err != nil {
		t.Fatalf("Failed to resolve step inputs: %v", err)
	}

	// Verify the resolution worked correctly
	expectedResolved := map[string]interface{}{
		"file_id":       "1BxY8Z9AbCdEfGhIjKlMnOpQrStUvWxYz",
		"new_parent_id": "1FoLdEr9AbCdEfGhIjKlMnOpQrStUvWxYz",
		"title":         "Weekly Report - Final Version",
	}

	if !deepEqual(resolvedInputs, expectedResolved) {
		t.Errorf("Parameter resolution failed.\nExpected: %+v\nGot: %+v", expectedResolved, resolvedInputs)
	}

	t.Logf("✅ Parameter resolution successful:")
	t.Logf("   file_id: %s", resolvedInputs["file_id"])
	t.Logf("   new_parent_id: %s", resolvedInputs["new_parent_id"])
	t.Logf("   title: %s", resolvedInputs["title"])
}

func TestTimezoneHandling(t *testing.T) {
	executionEngine := &ExecutionEngine{}

	tests := []struct {
		name           string
		inputs         map[string]interface{}
		userTimezone   string
		expectedOutput map[string]interface{}
		expectError    bool
	}{
		{
			name: "Datetime parameters get timezone added",
			inputs: map[string]interface{}{
				"startTime": "2025-08-18T10:00:00",
				"endTime":   "2025-08-18T11:00:00",
				"title":     "Meeting",
			},
			userTimezone: "Europe/Sofia",
			expectedOutput: map[string]interface{}{
				"startTime": "2025-08-18T10:00:00+03:00",
				"endTime":   "2025-08-18T11:00:00+03:00",
				"title":     "Meeting",
			},
			expectError: false,
		},
		{
			name: "Datetime with timezone unchanged",
			inputs: map[string]interface{}{
				"startTime": "2025-08-18T10:00:00+03:00",
			},
			userTimezone: "Europe/Sofia",
			expectedOutput: map[string]interface{}{
				"startTime": "2025-08-18T10:00:00+03:00",
			},
			expectError: false,
		},
		{
			name: "UTC datetime unchanged",
			inputs: map[string]interface{}{
				"startTime": "2025-08-18T10:00:00Z",
			},
			userTimezone: "Europe/Sofia",
			expectedOutput: map[string]interface{}{
				"startTime": "2025-08-18T10:00:00Z",
			},
			expectError: false,
		},
		{
			name: "America/New_York timezone",
			inputs: map[string]interface{}{
				"startTime": "2025-08-18T10:00:00",
			},
			userTimezone: "America/New_York",
			expectedOutput: map[string]interface{}{
				"startTime": "2025-08-18T10:00:00-04:00",
			},
			expectError: false,
		},
		{
			name: "Pacific timezone (larger negative offset)",
			inputs: map[string]interface{}{
				"startTime": "2025-08-18T10:00:00",
			},
			userTimezone: "America/Los_Angeles",
			expectedOutput: map[string]interface{}{
				"startTime": "2025-08-18T10:00:00-07:00",
			},
			expectError: false,
		},
		{
			name: "Datetime with negative timezone unchanged",
			inputs: map[string]interface{}{
				"startTime": "2025-08-18T10:00:00-05:00",
			},
			userTimezone: "Europe/Sofia",
			expectedOutput: map[string]interface{}{
				"startTime": "2025-08-18T10:00:00-05:00",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := &ParameterContext{
				UserParameters:    make(map[string]interface{}),
				StepOutputs:       make(map[string]interface{}),
				RuntimeParameters: make(map[string]interface{}),
				SystemParameters: map[string]interface{}{
					"user_timezone": tt.userTimezone,
				},
			}

			result, err := executionEngine.resolveStepInputs(tt.inputs, context)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !deepEqual(result, tt.expectedOutput) {
				t.Errorf("Expected: %+v, Got: %+v", tt.expectedOutput, result)
			}
		})
	}
}

func TestIsDateTimeValue(t *testing.T) {
	executionEngine := &ExecutionEngine{}

	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{
			name:     "ISO datetime without timezone",
			value:    "2025-08-18T10:00:00",
			expected: true,
		},
		{
			name:     "ISO datetime with timezone offset",
			value:    "2025-08-18T10:00:00+03:00",
			expected: true,
		},
		{
			name:     "ISO datetime with negative timezone offset",
			value:    "2025-08-18T10:00:00-05:00",
			expected: true,
		},
		{
			name:     "ISO datetime UTC",
			value:    "2025-08-18T10:00:00Z",
			expected: true,
		},
		{
			name:     "Regular string",
			value:    "not a datetime",
			expected: false,
		},
		{
			name:     "Date only",
			value:    "2025-08-18",
			expected: false,
		},
		{
			name:     "Time only",
			value:    "10:00:00",
			expected: false,
		},
		{
			name:     "Invalid datetime format",
			value:    "2025-08-18 10:00:00",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executionEngine.isDateTimeValue(tt.value)
			if result != tt.expected {
				t.Errorf("Expected: %v, Got: %v for value: %s", tt.expected, result, tt.value)
			}
		})
	}
}

// Helper functions for testing

func deepEqual(a, b interface{}) bool {
	// Simple deep equality check for test purposes
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	switch va := a.(type) {
	case map[string]interface{}:
		vb, ok := b.(map[string]interface{})
		if !ok || len(va) != len(vb) {
			return false
		}
		for k, v := range va {
			if !deepEqual(v, vb[k]) {
				return false
			}
		}
		return true
	case []interface{}:
		vb, ok := b.([]interface{})
		if !ok || len(va) != len(vb) {
			return false
		}
		for i, v := range va {
			if !deepEqual(v, vb[i]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}
