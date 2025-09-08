package storage

import (
	"log"
	"sohoaas-backend/internal/types"
)

// ParsingStorage decorates a WorkflowStorage to ensure ParsedData is populated
// uniformly for save/get/list operations, regardless of the underlying backend.
type parsingStorage struct {
	inner WorkflowStorage
}

// NewParsingStorage wraps a WorkflowStorage with parsing behavior.
func NewParsingStorage(inner WorkflowStorage) WorkflowStorage {
	return &parsingStorage{inner: inner}
}

// SaveWorkflow delegates to inner then parses the result's content.
func (ps *parsingStorage) SaveWorkflow(userID string, workflowName string, cueContent string) (*types.WorkflowFile, error) {
	wf, err := ps.inner.SaveWorkflow(userID, workflowName, cueContent)
	if err != nil {
		return nil, err
	}
	if wf != nil {
		if parsed, perr := parseCUEWorkflow(wf.Content, wf); perr == nil {
			wf = parsed
		} else {
			log.Printf("[ParsingStorage] SaveWorkflow: parse error for workflow %s: %v", wf.ID, perr)
		}
	}
	return wf, nil
}

// GetWorkflow delegates to inner then parses the result's content.
func (ps *parsingStorage) GetWorkflow(userID string, workflowID string) (*types.WorkflowFile, error) {
	wf, err := ps.inner.GetWorkflow(userID, workflowID)
	if err != nil {
		return nil, err
	}
	if wf != nil {
		if parsed, perr := parseCUEWorkflow(wf.Content, wf); perr == nil {
			wf = parsed
		} else {
			log.Printf("[ParsingStorage] GetWorkflow: parse error for workflow %s: %v", wf.ID, perr)
		}
	}
	return wf, nil
}

// ListUserWorkflows delegates to inner then parses each workflow's content.
func (ps *parsingStorage) ListUserWorkflows(userID string) ([]*types.WorkflowFile, error) {
	list, err := ps.inner.ListUserWorkflows(userID)
	if err != nil {
		return nil, err
	}
	for i, wf := range list {
		if wf == nil {
			continue
		}
		if parsed, perr := parseCUEWorkflow(wf.Content, wf); perr == nil {
			list[i] = parsed
		} else {
			log.Printf("[ParsingStorage] ListUserWorkflows: parse error for workflow %s: %v", wf.ID, perr)
		}
	}
	return list, nil
}

// Artifact management passthrough
func (ps *parsingStorage) SaveWorkflowArtifact(userID string, workflowID string, artifactType string, filename string, content string) error {
	return ps.inner.SaveWorkflowArtifact(userID, workflowID, artifactType, filename, content)
}

func (ps *parsingStorage) SavePrompt(userID string, workflowID string, promptName string, promptContent string) error {
	return ps.inner.SavePrompt(userID, workflowID, promptName, promptContent)
}

func (ps *parsingStorage) SaveResponse(userID string, workflowID string, responseName string, responseContent string) error {
	return ps.inner.SaveResponse(userID, workflowID, responseName, responseContent)
}

func (ps *parsingStorage) SaveExecutionLog(userID string, workflowID string, logContent string) error {
	return ps.inner.SaveExecutionLog(userID, workflowID, logContent)
}

// Storage backend identification passthrough
func (ps *parsingStorage) GetStorageType() string {
	return ps.inner.GetStorageType()
}

func (ps *parsingStorage) GetStorageInfo() map[string]interface{} {
	return ps.inner.GetStorageInfo()
}

// DeleteWorkflow passthrough to inner storage
func (ps *parsingStorage) DeleteWorkflow(userID string, workflowID string) error {
	return ps.inner.DeleteWorkflow(userID, workflowID)
}
