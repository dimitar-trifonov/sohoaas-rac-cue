package services

import (
	"context"
	"os"
	"testing"
)

// TestPromptSchemaValidation tests prompt schema loading directly
func TestPromptSchemaValidation(t *testing.T) {
	// Skip if no API key available
	if os.Getenv("GOOGLE_API_KEY") == "" {
		t.Skip("Skipping prompt schema test - GOOGLE_API_KEY not set")
	}

	t.Logf("=== TESTING PROMPT SCHEMA VALIDATION ===")

	// Initialize Genkit service
	genkitService, err := initializeGenkitServiceForTest(t)
	if err != nil {
		t.Fatalf("Failed to initialize Genkit service: %v", err)
	}

	// Test both prompts
	prompts := []string{"intent_analyst", "workflow_generator"}
	
	for _, promptName := range prompts {
		t.Run(promptName, func(t *testing.T) {
			t.Logf("Testing prompt: %s", promptName)
			
			// Load prompt
			prompt, err := genkitService.loadPrompt(promptName)
			if err != nil {
				t.Fatalf("Failed to load %s prompt: %v", promptName, err)
			}
			
			if prompt == nil {
				t.Fatalf("Loaded prompt is nil for %s", promptName)
			}
			
			t.Logf("✅ %s prompt loaded successfully", promptName)
			t.Logf("Prompt type: %T", prompt)
			
			// Test type assertion for execution capability
			if executablePrompt, ok := prompt.(interface{ Execute(context.Context, ...interface{}) (interface{}, error) }); ok {
				t.Logf("✅ %s prompt supports execution", promptName)
				_ = executablePrompt // Use the variable to avoid unused warning
			} else {
				t.Errorf("❌ %s prompt does not support execution", promptName)
				t.Logf("Available methods on prompt type %T:", prompt)
				// Try to get more information about the prompt type
				if stringer, ok := prompt.(interface{ String() string }); ok {
					t.Logf("Prompt string representation: %s", stringer.String())
				}
			}
		})
	}
}

// TestIntentAnalystPromptContent tests the content and structure of the intent analyst prompt
func TestIntentAnalystPromptContent(t *testing.T) {
	t.Logf("=== TESTING INTENT ANALYST PROMPT CONTENT ===")
	
	// Read the prompt file directly
	promptPath := "prompts/intent_analyst.prompt"
	content, err := os.ReadFile(promptPath)
	if err != nil {
		t.Fatalf("Failed to read prompt file: %v", err)
	}
	
	contentStr := string(content)
	t.Logf("Intent Analyst prompt content length: %d characters", len(contentStr))
	
	// Check for required sections
	requiredSections := []string{
		"---",           // YAML front matter start
		"model:",        // Model specification
		"input:",        // Input schema
		"output:",       // Output schema
		"schema:",       // Schema definitions
		"type: object",  // Object type definitions
		"properties:",   // Property definitions
	}
	
	for _, section := range requiredSections {
		if !containsText(contentStr, section) {
			t.Errorf("Missing required section in intent_analyst.prompt: %s", section)
		} else {
			t.Logf("✅ Found required section: %s", section)
		}
	}
	
	// Check for potential issues
	if containsText(contentStr, "type: array") && !containsText(contentStr, "items:") {
		t.Errorf("❌ Found array type without items definition")
	} else {
		t.Logf("✅ Array types have proper items definitions")
	}
}

// containsText checks if a string contains a substring
func containsText(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
