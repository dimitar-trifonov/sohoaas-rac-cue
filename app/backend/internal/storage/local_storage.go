package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sohoaas-backend/internal/types"
)

// LocalStorage implements WorkflowStorage interface using local filesystem
type LocalStorage struct {
	workflowsDir string
}

// NewLocalStorage creates a new local storage backend
func NewLocalStorage(config LocalStorageConfig) (WorkflowStorage, error) {
	workflowsDir := config.WorkflowsDir
	if workflowsDir == "" {
		workflowsDir = "./generated_workflows"
	}

	// Create workflows directory if it doesn't exist
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create workflows directory: %v", err)
	}

	return &LocalStorage{
		workflowsDir: workflowsDir,
	}, nil
}

// SaveWorkflow saves a generated CUE workflow to local filesystem
func (ls *LocalStorage) SaveWorkflow(userID string, workflowName string, cueContent string) (*types.WorkflowFile, error) {
	timestamp := time.Now().Format("20060102_150405")
	workflowID := fmt.Sprintf("%s_%s", timestamp, strings.ReplaceAll(workflowName, " ", "_"))

	userDir := filepath.Join(ls.workflowsDir, userID, workflowID)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create user workflow directory: %v", err)
	}

	workflowPath := filepath.Join(userDir, "workflow.cue")
	if err := os.WriteFile(workflowPath, []byte(cueContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write workflow file: %v", err)
	}

	workflowFile := &types.WorkflowFile{
		ID:        fmt.Sprintf("%s_%s", userID, workflowID),
		Filename:  "workflow.cue",
		Path:      workflowPath,
		UserID:    userID,
		Name:      workflowName,
		Content:   cueContent,
		CreatedAt: time.Now(),
	}

	// Parse CUE content into structured data
	if parsed, err := parseCUEWorkflow(cueContent, workflowFile); err == nil {
		workflowFile = parsed
	}
	// Note: If parsing fails, we still return the file with raw content

	return workflowFile, nil
}

// GetWorkflow retrieves a specific workflow file from local filesystem
func (ls *LocalStorage) GetWorkflow(userID string, workflowID string) (*types.WorkflowFile, error) {
	// Extract the actual workflow directory name from the combined ID
	workflowDirName := strings.TrimPrefix(workflowID, userID+"_")
	workflowPath := filepath.Join(ls.workflowsDir, userID, workflowDirName, "workflow.cue")
	if _, err := os.Stat(workflowPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}

	info, err := os.Stat(workflowPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow info: %v", err)
	}

	// Read workflow content
	content, err := os.ReadFile(workflowPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow content: %v", err)
	}

	workflowFile := &types.WorkflowFile{
		ID:        fmt.Sprintf("%s_%s", userID, workflowID),
		Filename:  "workflow.cue",
		Path:      workflowPath,
		UserID:    userID,
		Content:   string(content),
		CreatedAt: info.ModTime(),
	}

	// Parse CUE content into structured data
	if parsed, err := parseCUEWorkflow(string(content), workflowFile); err == nil {
		workflowFile = parsed
	}
	// Note: If parsing fails, we still return the file with raw content

	return workflowFile, nil
}

// ListUserWorkflows lists all CUE workflow files for a user from local filesystem
func (ls *LocalStorage) ListUserWorkflows(userID string) ([]*types.WorkflowFile, error) {
	userDir := filepath.Join(ls.workflowsDir, userID)
	if _, err := os.Stat(userDir); os.IsNotExist(err) {
		return []*types.WorkflowFile{}, nil
	}

	entries, err := os.ReadDir(userDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read user directory: %v", err)
	}

	var workflows []*types.WorkflowFile
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		workflowPath := filepath.Join(userDir, entry.Name(), "workflow.cue")
		if _, err := os.Stat(workflowPath); os.IsNotExist(err) {
			continue
		}

		info, err := os.Stat(workflowPath)
		if err != nil {
			continue
		}

		// Read workflow content
		content, err := os.ReadFile(workflowPath)
		if err != nil {
			continue
		}

		workflow := &types.WorkflowFile{
			ID:        fmt.Sprintf("%s_%s", userID, entry.Name()),
			Filename:  "workflow.cue",
			Path:      workflowPath,
			UserID:    userID,
			Name:      entry.Name(),
			Content:   string(content),
			CreatedAt: info.ModTime(),
		}

		// Parse CUE content into structured data
		if parsed, err := parseCUEWorkflow(string(content), workflow); err == nil {
			workflow = parsed
		}
		// Note: If parsing fails, we still include the file with raw content

		workflows = append(workflows, workflow)
	}

	return workflows, nil
}

// SaveWorkflowArtifact saves an artifact to the workflow's artifact directory
func (ls *LocalStorage) SaveWorkflowArtifact(userID string, workflowID string, artifactType string, filename string, content string) error {
	// Handle root directory artifacts (artifactType = ".")
	var artifactPath string
	if artifactType == "." || artifactType == "" {
		workflowDir := filepath.Join(ls.workflowsDir, userID, workflowID)
		if err := os.MkdirAll(workflowDir, 0755); err != nil {
			return fmt.Errorf("failed to create workflow directory: %v", err)
		}
		artifactPath = filepath.Join(workflowDir, filename)
	} else {
		artifactDir := filepath.Join(ls.workflowsDir, userID, workflowID, artifactType)
		if err := os.MkdirAll(artifactDir, 0755); err != nil {
			return fmt.Errorf("failed to create artifact directory: %v", err)
		}
		artifactPath = filepath.Join(artifactDir, filename)
	}

	return os.WriteFile(artifactPath, []byte(content), 0644)
}

// SavePrompt saves a prompt used during workflow generation
func (ls *LocalStorage) SavePrompt(userID string, workflowID string, promptName string, promptContent string) error {
	return ls.SaveWorkflowArtifact(userID, workflowID, "prompts", fmt.Sprintf("%s.txt", promptName), promptContent)
}

// SaveResponse saves an LLM response during workflow generation
func (ls *LocalStorage) SaveResponse(userID string, workflowID string, responseName string, responseContent string) error {
	return ls.SaveWorkflowArtifact(userID, workflowID, "responses", fmt.Sprintf("%s.json", responseName), responseContent)
}

// SaveExecutionLog saves execution logs for the workflow
func (ls *LocalStorage) SaveExecutionLog(userID string, workflowID string, logContent string) error {
	timestamp := time.Now().Format("20060102_150405")
	return ls.SaveWorkflowArtifact(userID, workflowID, "logs", fmt.Sprintf("execution_%s.log", timestamp), logContent)
}

// GetStorageType returns the storage backend type
func (ls *LocalStorage) GetStorageType() string {
	return "local"
}

// GetStorageInfo returns information about the storage backend
func (ls *LocalStorage) GetStorageInfo() map[string]interface{} {
	return map[string]interface{}{
		"type":          "local",
		"workflows_dir": ls.workflowsDir,
		"created_at":    time.Now().Format(time.RFC3339),
	}
}
