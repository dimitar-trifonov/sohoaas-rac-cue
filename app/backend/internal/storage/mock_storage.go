package storage

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"sohoaas-backend/internal/types"
)

// MockStorage implements WorkflowStorage interface for testing
type MockStorage struct {
	workflows map[string]*types.WorkflowFile // key: userID_workflowID
	artifacts map[string]string              // key: userID_workflowID_type_filename
	mu        sync.RWMutex
}

// NewMockStorage creates a new mock storage backend
func NewMockStorage() *MockStorage {
	return &MockStorage{
		workflows: make(map[string]*types.WorkflowFile),
		artifacts: make(map[string]string),
	}
}

// SaveWorkflow saves a workflow to mock storage
func (m *MockStorage) SaveWorkflow(userID string, workflowName string, cueContent string) (*types.WorkflowFile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	timestamp := time.Now().Format("20060102_150405")
	workflowID := fmt.Sprintf("%s_%s", timestamp, workflowName)
	id := fmt.Sprintf("%s_%s", userID, workflowID)

	workflowFile := &types.WorkflowFile{
		ID:        id,
		Filename:  "workflow.cue",
		Path:      fmt.Sprintf("/mock/%s/workflow.cue", id),
		UserID:    userID,
		Name:      workflowName,
		Content:   cueContent,
		CreatedAt: time.Now(),
	}

	// Parse CUE content into structured data
	if parsed, err := parseCUEWorkflow(cueContent, workflowFile); err == nil {
		workflowFile = parsed
	}

	m.workflows[id] = workflowFile
	return workflowFile, nil
}

// GetWorkflow retrieves a workflow from mock storage
func (m *MockStorage) GetWorkflow(userID string, workflowID string) (*types.WorkflowFile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	workflow, exists := m.workflows[workflowID]
	if !exists {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}
	return workflow, nil
}

// ListUserWorkflows lists workflows for a user from mock storage
func (m *MockStorage) ListUserWorkflows(userID string) ([]*types.WorkflowFile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var workflows []*types.WorkflowFile
	for _, workflow := range m.workflows {
		if workflow.UserID == userID {
			workflows = append(workflows, workflow)
		}
	}
	return workflows, nil
}

// SaveWorkflowArtifact saves an artifact to mock storage
func (m *MockStorage) SaveWorkflowArtifact(userID string, workflowID string, artifactType string, filename string, content string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s_%s_%s_%s", userID, workflowID, artifactType, filename)
	m.artifacts[key] = content
	return nil
}

// SavePrompt saves a prompt to mock storage
func (m *MockStorage) SavePrompt(userID string, workflowID string, promptName string, promptContent string) error {
	return m.SaveWorkflowArtifact(userID, workflowID, "prompts", fmt.Sprintf("%s.txt", promptName), promptContent)
}

// SaveResponse saves a response to mock storage
func (m *MockStorage) SaveResponse(userID string, workflowID string, responseName string, responseContent string) error {
	return m.SaveWorkflowArtifact(userID, workflowID, "responses", fmt.Sprintf("%s.json", responseName), responseContent)
}

// SaveExecutionLog saves execution logs to mock storage
func (m *MockStorage) SaveExecutionLog(userID string, workflowID string, logContent string) error {
	timestamp := time.Now().Format("20060102_150405")
	return m.SaveWorkflowArtifact(userID, workflowID, "logs", fmt.Sprintf("execution_%s.log", timestamp), logContent)
}

// GetStorageType returns the storage type
func (m *MockStorage) GetStorageType() string {
	return "mock"
}

// GetStorageInfo returns storage info
func (m *MockStorage) GetStorageInfo() map[string]interface{} {
	return map[string]interface{}{
		"type":       "mock",
		"created_at": time.Now().Format(time.RFC3339),
	}
}

// DeleteWorkflow removes a workflow from mock storage
func (m *MockStorage) DeleteWorkflow(userID string, workflowID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.workflows[workflowID]; !ok {
		return fmt.Errorf("workflow not found: %s", workflowID)
	}
	delete(m.workflows, workflowID)
	// Optionally clean artifacts for this workflow
	for key := range m.artifacts {
		if len(key) >= len(userID)+1+len(workflowID) && key[:len(userID)] == userID {
			// key format: userID_workflowID_type_filename (best-effort cleanup)
			// We simply check contains workflowID to avoid over-complication
			if strings.Contains(key, workflowID) {
				delete(m.artifacts, key)
			}
		}
	}
	return nil
}
