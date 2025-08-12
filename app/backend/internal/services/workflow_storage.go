package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sohoaas-backend/internal/types"
)

// WorkflowStorageService handles storing generated CUE workflows to disk
type WorkflowStorageService struct {
	workflowsDir string
}

// NewWorkflowStorageService creates a new workflow storage service
func NewWorkflowStorageService(workflowsDir string) *WorkflowStorageService {
	// Create workflows directory if it doesn't exist
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create workflows directory: %v", err))
	}

	return &WorkflowStorageService{
		workflowsDir: workflowsDir,
	}
}

// SaveWorkflow saves a generated CUE workflow to disk with organized artifacts
func (w *WorkflowStorageService) SaveWorkflow(userID string, workflowName string, cueContent string) (*types.WorkflowFile, error) {
	// Create user-specific directory
	userDir := filepath.Join(w.workflowsDir, userID)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create user directory: %w", err)
	}

	// Generate unique workflow ID and timestamp
	timestamp := time.Now().Format("20060102_150405")
	workflowID := fmt.Sprintf("%s_%s", workflowName, timestamp)
	
	// Create dedicated workflow folder
	workflowDir := filepath.Join(userDir, workflowID)
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create workflow directory: %w", err)
	}

	// Create artifact subdirectories
	artifactDirs := []string{"prompts", "responses", "metadata", "logs"}
	for _, dir := range artifactDirs {
		if err := os.MkdirAll(filepath.Join(workflowDir, dir), 0755); err != nil {
			return nil, fmt.Errorf("failed to create artifact directory %s: %w", dir, err)
		}
	}

	// Save main CUE workflow file
	cueFilename := fmt.Sprintf("%s.cue", workflowName)
	cueFilepath := filepath.Join(workflowDir, cueFilename)
	if err := os.WriteFile(cueFilepath, []byte(cueContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write workflow file: %w", err)
	}

	// Create workflow metadata file
	metadata := fmt.Sprintf(`{
	"workflow_id": "%s",
	"name": "%s",
	"created_at": "%s",
	"user_id": "%s",
	"status": "draft",
	"artifacts": {
		"cue_file": "%s",
		"prompts_dir": "prompts/",
		"responses_dir": "responses/",
		"metadata_dir": "metadata/",
		"logs_dir": "logs/"
	}
}`, workflowID, workflowName, time.Now().Format(time.RFC3339), userID, cueFilename)

	metadataPath := filepath.Join(workflowDir, "metadata", "workflow.json")
	if err := os.WriteFile(metadataPath, []byte(metadata), 0644); err != nil {
		return nil, fmt.Errorf("failed to write metadata file: %w", err)
	}

	return &types.WorkflowFile{
		ID:          fmt.Sprintf("%s_%s", userID, timestamp),
		Name:        workflowName,
		Description: fmt.Sprintf("Generated workflow: %s", workflowName),
		Status:      "draft", // New workflows start as draft
		Filename:    cueFilename,
		Path:        cueFilepath,
		UserID:      userID,
		Content:     cueContent,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

// SaveWorkflowArtifact saves an artifact to the workflow's artifact directory
func (w *WorkflowStorageService) SaveWorkflowArtifact(userID string, workflowID string, artifactType string, filename string, content string) error {
	// Find workflow directory
	userDir := filepath.Join(w.workflowsDir, userID)
	workflowDir := filepath.Join(userDir, workflowID)
	
	// Check if workflow directory exists
	if _, err := os.Stat(workflowDir); os.IsNotExist(err) {
		return fmt.Errorf("workflow directory not found: %s", workflowID)
	}
	
	// Create artifact file path
	artifactDir := filepath.Join(workflowDir, artifactType)
	if err := os.MkdirAll(artifactDir, 0755); err != nil {
		return fmt.Errorf("failed to create artifact directory %s: %w", artifactType, err)
	}
	
	artifactPath := filepath.Join(artifactDir, filename)
	if err := os.WriteFile(artifactPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write artifact file: %w", err)
	}
	
	return nil
}

// SavePrompt saves a prompt used during workflow generation
func (w *WorkflowStorageService) SavePrompt(userID string, workflowID string, promptName string, promptContent string) error {
	filename := fmt.Sprintf("%s_%s.txt", promptName, time.Now().Format("150405"))
	return w.SaveWorkflowArtifact(userID, workflowID, "prompts", filename, promptContent)
}

// SaveResponse saves an LLM response during workflow generation
func (w *WorkflowStorageService) SaveResponse(userID string, workflowID string, responseName string, responseContent string) error {
	filename := fmt.Sprintf("%s_%s.json", responseName, time.Now().Format("150405"))
	return w.SaveWorkflowArtifact(userID, workflowID, "responses", filename, responseContent)
}

// SaveExecutionLog saves execution logs for the workflow
func (w *WorkflowStorageService) SaveExecutionLog(userID string, workflowID string, logContent string) error {
	filename := fmt.Sprintf("execution_%s.log", time.Now().Format("20060102_150405"))
	return w.SaveWorkflowArtifact(userID, workflowID, "logs", filename, logContent)
}

// ListUserWorkflows lists all CUE workflow files for a user
func (w *WorkflowStorageService) ListUserWorkflows(userID string) ([]*types.WorkflowFile, error) {
	userDir := filepath.Join(w.workflowsDir, userID)
	
	// Check if user directory exists
	if _, err := os.Stat(userDir); os.IsNotExist(err) {
		return []*types.WorkflowFile{}, nil
	}

	entries, err := os.ReadDir(userDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read user workflows directory: %w", err)
	}

	var workflows []*types.WorkflowFile
	for _, entry := range entries {
		// Look for workflow directories (new structure) or legacy .cue files
		if entry.IsDir() {
			// New structure: workflow folder with artifacts
			workflowDir := filepath.Join(userDir, entry.Name())
			
			// Look for .cue file in the workflow directory
			workflowFiles, err := os.ReadDir(workflowDir)
			if err != nil {
				continue // Skip directories we can't read
			}
			
			var cueFile string
			for _, file := range workflowFiles {
				if !file.IsDir() && filepath.Ext(file.Name()) == ".cue" {
					cueFile = file.Name()
					break
				}
			}
			
			if cueFile == "" {
				continue // No CUE file found in this directory
			}
			
			cueFilePath := filepath.Join(workflowDir, cueFile)
			content, err := os.ReadFile(cueFilePath)
			if err != nil {
				continue // Skip files we can't read
			}

			info, err := entry.Info()
			if err != nil {
				continue
			}

			// Extract workflow name from directory name (remove timestamp)
			workflowName := entry.Name()
			if idx := strings.LastIndex(workflowName, "_"); idx > 0 {
				workflowName = workflowName[:idx]
			}
			
			workflows = append(workflows, &types.WorkflowFile{
				ID:          fmt.Sprintf("%s_%s", userID, entry.Name()),
				Name:        workflowName,
				Description: fmt.Sprintf("Generated workflow: %s", workflowName),
				Status:      "draft", // Default status for existing workflows
				Filename:    cueFile,
				Path:        cueFilePath,
				UserID:      userID,
				Content:     string(content),
				CreatedAt:   info.ModTime(),
				UpdatedAt:   info.ModTime(),
			})
		} else if filepath.Ext(entry.Name()) == ".cue" {
			// Legacy structure: direct .cue files in user directory
			filePath := filepath.Join(userDir, entry.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				continue // Skip files we can't read
			}

			info, err := entry.Info()
			if err != nil {
				continue
			}

			// Extract workflow name from filename (remove timestamp and extension)
			workflowName := entry.Name()
			if idx := strings.LastIndex(workflowName, "_"); idx > 0 {
				workflowName = workflowName[:idx]
			}
			workflowName = strings.TrimSuffix(workflowName, ".cue")
			
			workflows = append(workflows, &types.WorkflowFile{
				ID:          fmt.Sprintf("%s_%s", userID, entry.Name()),
				Name:        workflowName,
				Description: fmt.Sprintf("Generated workflow: %s", workflowName),
				Status:      "draft", // Default status for existing workflows
				Filename:    entry.Name(),
				Path:        filePath,
				UserID:      userID,
				Content:     string(content),
				CreatedAt:   info.ModTime(),
				UpdatedAt:   info.ModTime(),
			})
		}
	}

	return workflows, nil
}

// GetWorkflow retrieves a specific workflow file
func (w *WorkflowStorageService) GetWorkflow(userID string, workflowID string) (*types.WorkflowFile, error) {
	workflows, err := w.ListUserWorkflows(userID)
	if err != nil {
		return nil, err
	}

	for _, workflow := range workflows {
		if workflow.ID == workflowID {
			return workflow, nil
		}
	}

	return nil, fmt.Errorf("workflow not found: %s", workflowID)
}
