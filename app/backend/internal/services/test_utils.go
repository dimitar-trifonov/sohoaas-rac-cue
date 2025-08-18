package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// getTestOutputDir returns the test output directory using the same logic as WorkflowStorageService
func getTestOutputDir() string {
	// Use unified ARTIFACT_OUTPUT_DIR for all artifact storage
	if dir := os.Getenv("ARTIFACT_OUTPUT_DIR"); dir != "" {
		return dir
	}
	return "/tmp"
}

// saveTestArtifact saves test artifacts using WorkflowStorageService structure for consistency
func saveTestArtifact(testName, artifactType, filename, content string) error {
	// Skip artifact saving if not in test mode to prevent duplication with WorkflowStorageService
	if !strings.Contains(os.Args[0], "test") && !strings.HasSuffix(os.Args[0], ".test") {
		return nil // Skip saving in production to avoid duplicate directories
	}
	
	// Use consistent directory structure: {base}/test_artifacts/{testName}/{artifactType}/
	baseDir := getTestOutputDir()
	artifactDir := filepath.Join(baseDir, "test_artifacts", testName, artifactType)
	
	if err := os.MkdirAll(artifactDir, 0755); err != nil {
		return fmt.Errorf("failed to create artifact directory: %w", err)
	}
	
	artifactPath := filepath.Join(artifactDir, filename)
	if err := os.WriteFile(artifactPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write artifact file: %w", err)
	}
	
	return nil
}
