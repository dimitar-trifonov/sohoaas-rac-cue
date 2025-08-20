package storage

import (
	"sohoaas-backend/internal/types"
)

// WorkflowStorage defines the common interface for workflow storage backends
type WorkflowStorage interface {
	// Core workflow operations
	SaveWorkflow(userID string, workflowName string, cueContent string) (*types.WorkflowFile, error)
	GetWorkflow(userID string, workflowID string) (*types.WorkflowFile, error)
	ListUserWorkflows(userID string) ([]*types.WorkflowFile, error)
	
	// Artifact management
	SaveWorkflowArtifact(userID string, workflowID string, artifactType string, filename string, content string) error
	SavePrompt(userID string, workflowID string, promptName string, promptContent string) error
	SaveResponse(userID string, workflowID string, responseName string, responseContent string) error
	SaveExecutionLog(userID string, workflowID string, logContent string) error
	
	// Storage backend identification
	GetStorageType() string
	GetStorageInfo() map[string]interface{}
}

// StorageConfig holds configuration for different storage backends
type StorageConfig struct {
	Backend string `json:"backend"` // "local" or "gcs"
	
	// Local storage config
	LocalConfig LocalStorageConfig `json:"local,omitempty"`
	
	// GCS storage config  
	GCSConfig GCSStorageConfig `json:"gcs,omitempty"`
}

// LocalStorageConfig for filesystem-based storage
type LocalStorageConfig struct {
	WorkflowsDir string `json:"workflows_dir"`
}

// GCSStorageConfig for Google Cloud Storage
type GCSStorageConfig struct {
	BucketName           string `json:"bucket_name"`
	ServiceAccountKey    string `json:"service_account_key,omitempty"` // JSON content of service account key
	ProjectID            string `json:"project_id"`
	WorkflowsPrefix      string `json:"workflows_prefix"` // e.g., "workflows/" or "sohoaas/workflows/"
}

// StorageFactory creates storage backends based on configuration
type StorageFactory struct{}

// NewStorage creates a storage backend based on the provided configuration
func (f *StorageFactory) NewStorage(config StorageConfig) (WorkflowStorage, error) {
	switch config.Backend {
	case "local":
		return NewLocalStorage(config.LocalConfig)
	case "gcs":
		return NewGCSStorage(config.GCSConfig)
	default:
		// Default to local storage for backward compatibility
		return NewLocalStorage(LocalStorageConfig{
			WorkflowsDir: "./generated_workflows",
		})
	}
}
