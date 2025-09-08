package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"sohoaas-backend/internal/types"
)

// GCSStorage implements WorkflowStorage interface using Google Cloud Storage
type GCSStorage struct {
	client          *storage.Client
	bucketName      string
	workflowsPrefix string
	ctx             context.Context
}

// NewGCSStorage creates a new GCS storage backend
func NewGCSStorage(config GCSStorageConfig) (WorkflowStorage, error) {
	ctx := context.Background()
	
	var client *storage.Client
	var err error
	
	// Initialize GCS client with service account JSON content
	if config.ServiceAccountKey != "" {
		client, err = storage.NewClient(ctx, option.WithCredentialsJSON([]byte(config.ServiceAccountKey)))
	} else {
		// Use default credentials (for GCP environments)
		client, err = storage.NewClient(ctx)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %v", err)
	}
	
	workflowsPrefix := config.WorkflowsPrefix
	if workflowsPrefix == "" {
		workflowsPrefix = "workflows/"
	}
	if !strings.HasSuffix(workflowsPrefix, "/") {
		workflowsPrefix += "/"
	}
	
	return &GCSStorage{
		client:          client,
		bucketName:      config.BucketName,
		workflowsPrefix: workflowsPrefix,
		ctx:             ctx,
	}, nil
}

// SaveWorkflow saves a generated CUE workflow to GCS
func (gcs *GCSStorage) SaveWorkflow(userID string, workflowName string, cueContent string) (*types.WorkflowFile, error) {
	// Generate workflow ID with timestamp
	timestamp := time.Now().Format("20060102_150405")
	workflowID := fmt.Sprintf("%s_%s", timestamp, strings.ReplaceAll(workflowName, " ", "_"))
	
	// Create workflow file path: workflows/{userID}/{workflowID}/workflow.cue
	objectPath := fmt.Sprintf("%s%s/%s/workflow.cue", gcs.workflowsPrefix, userID, workflowID)
	
	// Upload workflow content to GCS
	obj := gcs.client.Bucket(gcs.bucketName).Object(objectPath)
	writer := obj.NewWriter(gcs.ctx)
	writer.ContentType = "text/plain"
	
	if _, err := writer.Write([]byte(cueContent)); err != nil {
		writer.Close()
		return nil, fmt.Errorf("failed to write workflow to GCS: %v", err)
	}
	
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close GCS writer: %v", err)
	}
	
	// Create workflow file metadata
	workflowFile := &types.WorkflowFile{
		ID:       fmt.Sprintf("%s_%s", userID, workflowID),
		Filename: "workflow.cue",
		Path:     objectPath,
		Content:  cueContent,
		UserID:   userID,
		Name:     workflowName, // Add Name field to match local storage interface
		CreatedAt: time.Now(),
	}
	
	return workflowFile, nil
}

// GetWorkflow retrieves a specific workflow file from GCS
func (gcs *GCSStorage) GetWorkflow(userID string, workflowID string) (*types.WorkflowFile, error) {
	// Remove userID prefix if present
	cleanWorkflowID := strings.TrimPrefix(workflowID, userID+"_")
	
	// Create workflow file path: workflows/{userID}/{workflowID}/workflow.cue
	objectPath := fmt.Sprintf("%s%s/%s/workflow.cue", gcs.workflowsPrefix, userID, cleanWorkflowID)
	
	// Read workflow content from GCS
	obj := gcs.client.Bucket(gcs.bucketName).Object(objectPath)
	reader, err := obj.NewReader(gcs.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow from GCS: %v", err)
	}
	defer reader.Close()
	
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow content: %v", err)
	}
	
	// Get object attributes for metadata
	attrs, err := obj.Attrs(gcs.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get object attributes: %v", err)
	}
	
	workflowFile := &types.WorkflowFile{
		ID:       fmt.Sprintf("%s_%s", userID, cleanWorkflowID),
		Filename: "workflow.cue",
		Path:     objectPath,
		Content:  string(content),
		UserID:   userID,
		Name:     cleanWorkflowID, // Add Name field to match local storage interface
		CreatedAt: attrs.Created,
	}
	
	return workflowFile, nil
}

// ListUserWorkflows lists all CUE workflow files for a user from GCS
func (gcs *GCSStorage) ListUserWorkflows(userID string) ([]*types.WorkflowFile, error) {
	prefix := fmt.Sprintf("%s%s/", gcs.workflowsPrefix, userID)
	
	var workflows []*types.WorkflowFile
	
	// List objects with the user's workflow prefix
	it := gcs.client.Bucket(gcs.bucketName).Objects(gcs.ctx, &storage.Query{
		Prefix: prefix,
	})
	
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			// No more items - this is normal, not an error
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list workflows: %v", err)
		}
		
		// Only include workflow.cue files
		if !strings.HasSuffix(attrs.Name, "/workflow.cue") {
			continue
		}
		
		// Extract workflow ID from path
		relativePath := strings.TrimPrefix(attrs.Name, prefix)
		workflowID := strings.TrimSuffix(relativePath, "/workflow.cue")
		
		// Read the actual workflow content
		obj := gcs.client.Bucket(gcs.bucketName).Object(attrs.Name)
		reader, err := obj.NewReader(gcs.ctx)
		if err != nil {
			// Log error but continue with other workflows
			log.Printf("Failed to read workflow content for %s: %v", attrs.Name, err)
			continue
		}
		
		content, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			log.Printf("Failed to read workflow content for %s: %v", attrs.Name, err)
			continue
		}
		
		workflowFile := &types.WorkflowFile{
			ID:       fmt.Sprintf("%s_%s", userID, workflowID),
			Filename: "workflow.cue",
			Path:     attrs.Name,
			Content:  string(content),
			UserID:   userID,
			Name:     workflowID, // Add Name field to match local storage interface
			CreatedAt: attrs.Created,
		}
		
		workflows = append(workflows, workflowFile)
	}
	
	return workflows, nil
}

// SaveWorkflowArtifact saves an artifact to the workflow's artifact directory in GCS
func (gcs *GCSStorage) SaveWorkflowArtifact(userID string, workflowID string, artifactType string, filename string, content string) error {
	cleanWorkflowID := strings.TrimPrefix(workflowID, userID+"_")
	
	// Handle root directory artifacts (artifactType = ".")
	var objectPath string
	if artifactType == "." || artifactType == "" {
		objectPath = fmt.Sprintf("%s%s/%s/%s", gcs.workflowsPrefix, userID, cleanWorkflowID, filename)
	} else {
		objectPath = fmt.Sprintf("%s%s/%s/%s/%s", gcs.workflowsPrefix, userID, cleanWorkflowID, artifactType, filename)
	}
	
	obj := gcs.client.Bucket(gcs.bucketName).Object(objectPath)
	writer := obj.NewWriter(gcs.ctx)
	writer.ContentType = "text/plain"
	
	if _, err := writer.Write([]byte(content)); err != nil {
		writer.Close()
		return fmt.Errorf("failed to write artifact to GCS: %v", err)
	}
	
	return writer.Close()
}

// SavePrompt saves a prompt used during workflow generation
func (gcs *GCSStorage) SavePrompt(userID string, workflowID string, promptName string, promptContent string) error {
	filename := fmt.Sprintf("%s_%s.txt", promptName, time.Now().Format("150405"))
	return gcs.SaveWorkflowArtifact(userID, workflowID, "prompts", filename, promptContent)
}

// SaveResponse saves an LLM response during workflow generation
func (gcs *GCSStorage) SaveResponse(userID string, workflowID string, responseName string, responseContent string) error {
	filename := fmt.Sprintf("%s_%s.json", responseName, time.Now().Format("150405"))
	return gcs.SaveWorkflowArtifact(userID, workflowID, "responses", filename, responseContent)
}

// SaveExecutionLog saves execution logs for the workflow
func (gcs *GCSStorage) SaveExecutionLog(userID string, workflowID string, logContent string) error {
	filename := fmt.Sprintf("execution_%s.log", time.Now().Format("20060102_150405"))
	return gcs.SaveWorkflowArtifact(userID, workflowID, "logs", filename, logContent)
}

// GetStorageType returns the storage backend type
func (gcs *GCSStorage) GetStorageType() string {
	return "gcs"
}

// GetStorageInfo returns information about the storage backend
func (gcs *GCSStorage) GetStorageInfo() map[string]interface{} {
	return map[string]interface{}{
		"type":             "gcs",
		"bucket_name":      gcs.bucketName,
		"workflows_prefix": gcs.workflowsPrefix,
		"created_at":       time.Now().Format(time.RFC3339),
	}
}

// DeleteWorkflow deletes all objects under the workflow prefix for the given user and workflow ID
func (gcs *GCSStorage) DeleteWorkflow(userID string, workflowID string) error {
	cleanWorkflowID := strings.TrimPrefix(workflowID, userID+"_")
	prefix := fmt.Sprintf("%s%s/%s/", gcs.workflowsPrefix, userID, cleanWorkflowID)

	it := gcs.client.Bucket(gcs.bucketName).Objects(gcs.ctx, &storage.Query{
		Prefix: prefix,
	})

	// Track if we found anything to delete
	found := false
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to iterate objects for deletion: %v", err)
		}
		found = true
		obj := gcs.client.Bucket(gcs.bucketName).Object(attrs.Name)
		if err := obj.Delete(gcs.ctx); err != nil {
			return fmt.Errorf("failed to delete object %s: %v", attrs.Name, err)
		}
	}

	if !found {
		return fmt.Errorf("workflow not found: %s", workflowID)
	}
	return nil
}
