package storage

import (
	"fmt"
	"os"
)

// CreateStorageFromEnv creates a storage backend based on environment variables
func CreateStorageFromEnv() (WorkflowStorage, error) {
	factory := &StorageFactory{}
	
	// Determine storage backend from environment
	backend := os.Getenv("STORAGE_BACKEND")
	if backend == "" {
		backend = "local" // Default to local for backward compatibility
	}
	
	config := StorageConfig{
		Backend: backend,
	}
	
	switch backend {
	case "local":
		config.LocalConfig = LocalStorageConfig{
			WorkflowsDir: getEnvOrDefault("WORKFLOWS_DIR", "./generated_workflows"),
		}
	case "gcs":
		config.GCSConfig = GCSStorageConfig{
			BucketName:        getEnvOrDefault("GCS_BUCKET_NAME", "sohoaas-workflows"),
			ServiceAccountKey: os.Getenv("GCS_SERVICE_ACCOUNT_KEY"),
			ProjectID:         getEnvOrDefault("GCS_PROJECT_ID", os.Getenv("FIREBASE_PROJECT_ID")),
			WorkflowsPrefix:   getEnvOrDefault("GCS_WORKFLOWS_PREFIX", "workflows/"),
		}
	default:
		return nil, fmt.Errorf("unsupported storage backend: %s", backend)
	}
	
	return factory.NewStorage(config)
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
