package workspace

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dimitar-trifonov/sohoaas/service-proxies/workflow"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// DriveProxy implements WorkspaceProxy for Google Drive service
type DriveProxy struct {
	config *oauth2.Config
}

// NewDriveProxy creates a new Drive proxy instance
func NewDriveProxy(config *oauth2.Config) *DriveProxy {
	return &DriveProxy{
		config: config,
	}
}

// Execute calls a Drive function with the given payload
func (p *DriveProxy) Execute(ctx context.Context, function string, token string, payload map[string]interface{}) (*workflow.ProxyResponse, error) {
	startTime := time.Now()

	// Validate function
	if !p.isSupportedFunction(function) {
		return &workflow.ProxyResponse{
			Success: false,
			Error: &workflow.ProxyError{
				Code:      string(ErrorCodeInvalidFunction),
				Message:   fmt.Sprintf("Unsupported function: %s", function),
				Retryable: false,
			},
		}, nil
	}

	// Validate payload
	if err := p.ValidatePayload(function, payload); err != nil {
		return &workflow.ProxyResponse{
			Success: false,
			Error: &workflow.ProxyError{
				Code:      string(ErrorCodeInvalidPayload),
				Message:   err.Error(),
				Retryable: false,
			},
		}, nil
	}

	// Initialize Drive service
	oauthToken := &oauth2.Token{AccessToken: token}
	client := p.config.Client(ctx, oauthToken)
	service, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return &workflow.ProxyResponse{
			Success: false,
			Error: &workflow.ProxyError{
				Code:      string(ErrorCodeAuthenticationFailed),
				Message:   "Failed to initialize Drive service",
				Details:   err.Error(),
				Retryable: true,
			},
		}, nil
	}

	// Execute the function
	var result map[string]interface{}
	var execErr error

	switch function {
	case DriveFunctionCreateFolder:
		result, execErr = p.createFolder(ctx, service, payload)
	case DriveFunctionUploadFile:
		result, execErr = p.uploadFile(ctx, service, payload)
	case DriveFunctionGetFile:
		result, execErr = p.getFile(ctx, service, payload)
	case DriveFunctionListFiles:
		result, execErr = p.listFiles(ctx, service, payload)
	case DriveFunctionShareFile:
		result, execErr = p.shareFile(ctx, service, payload)
	case DriveFunctionMoveFile:
		result, execErr = p.moveFile(ctx, service, payload)
	default:
		execErr = fmt.Errorf("function not implemented: %s", function)
	}

	if execErr != nil {
		return &workflow.ProxyResponse{
			Success: false,
			Error: &workflow.ProxyError{
				Code:      string(ErrorCodeInternalError),
				Message:   "Function execution failed",
				Details:   execErr.Error(),
				Retryable: true,
			},
		}, nil
	}

	return &workflow.ProxyResponse{
		Success: true,
		Data:    result,
		Metadata: &workflow.ResponseMetadata{
			ExecutionTime: time.Since(startTime),
			Function:      function,
			Timestamp:     time.Now(),
		},
	}, nil
}

// GetSupportedFunctions returns supported Drive functions
func (p *DriveProxy) GetSupportedFunctions() []string {
	return []string{
		DriveFunctionCreateFolder,
		DriveFunctionUploadFile,
		DriveFunctionGetFile,
		DriveFunctionListFiles,
		DriveFunctionShareFile,
		DriveFunctionMoveFile,
	}
}

// GetServiceType returns the service type
func (p *DriveProxy) GetServiceType() string {
	return ServiceTypeDrive
}

// GetServiceCapabilities returns the service capabilities
func (p *DriveProxy) GetServiceCapabilities() map[string]interface{} {
	return map[string]interface{}{
		"service_type":        ServiceTypeDrive,
		"supported_functions": p.GetSupportedFunctions(),
		"max_file_size":       "5TB",
		"file_sharing":        true,
		"folder_management":   true,
		"version_control":     true,
		"search_capabilities": true,
	}
}

// GetServiceMetadata returns metadata about the Drive service and its functions
func (p *DriveProxy) GetServiceMetadata() ServiceMetadata {
	return ServiceMetadata{
		ServiceType: ServiceTypeDrive,
		DisplayName: "Google Drive",
		Description: "Store, share, and manage files using Google Drive API",
		Functions: map[string]FunctionMetadata{
			DriveFunctionUploadFile: {
				Name:        DriveFunctionUploadFile,
				DisplayName: "Upload File",
				Description: "Upload a file to Google Drive",
				ExamplePayload: map[string]interface{}{
					"name":      "test-file.txt",
					"content":   "VGVzdCBmaWxlIGNvbnRlbnQ=", // base64 encoded
					"parent_id": "root",
				},
				RequiredFields: []string{"name", "content"},
			},
			DriveFunctionCreateFolder: {
				Name:        DriveFunctionCreateFolder,
				DisplayName: "Create Folder",
				Description: "Create a new folder in Google Drive",
				ExamplePayload: map[string]interface{}{
					"name":      "New Folder",
					"parent_id": "root",
				},
				RequiredFields: []string{"name"},
			},
			DriveFunctionGetFile: {
				Name:        DriveFunctionGetFile,
				DisplayName: "Get File",
				Description: "Retrieve file information from Google Drive",
				ExamplePayload: map[string]interface{}{
					"file_id": "1234567890abcdef",
				},
				RequiredFields: []string{"file_id"},
			},
			DriveFunctionListFiles: {
				Name:        DriveFunctionListFiles,
				DisplayName: "List Files",
				Description: "List files and folders in Google Drive",
				ExamplePayload: map[string]interface{}{
					"folder_id": "root",
					"page_size": 10,
				},
				RequiredFields: []string{},
			},
			DriveFunctionShareFile: {
				Name:        DriveFunctionShareFile,
				DisplayName: "Share File",
				Description: "Share a file with another user",
				ExamplePayload: map[string]interface{}{
					"file_id": "1234567890abcdef",
					"email":   "user@example.com",
					"role":    "reader",
				},
				RequiredFields: []string{"file_id", "email", "role"},
			},
			DriveFunctionMoveFile: {
				Name:        DriveFunctionMoveFile,
				DisplayName: "Move File",
				Description: "Move a file to a different folder",
				ExamplePayload: map[string]interface{}{
					"file_id":       "1234567890abcdef",
					"new_parent_id": "0987654321fedcba",
				},
				RequiredFields: []string{"file_id", "new_parent_id"},
			},
		},
	}
}

// GetFunctionMetadata returns metadata for a specific Drive function
func (p *DriveProxy) GetFunctionMetadata(function string) (FunctionMetadata, error) {
	metadata := p.GetServiceMetadata()
	funcMeta, exists := metadata.Functions[function]
	if !exists {
		return FunctionMetadata{}, fmt.Errorf("function %s not supported by Drive service", function)
	}
	return funcMeta, nil
}

// ValidateRequest validates a request (wrapper around ValidatePayload for interface compatibility)
func (p *DriveProxy) ValidateRequest(function string, payload map[string]interface{}) error {
	return p.ValidatePayload(function, payload)
}

// ValidatePayload validates the payload for a given function
func (p *DriveProxy) ValidatePayload(function string, payload map[string]interface{}) error {
	switch function {
	case DriveFunctionCreateFolder:
		if _, ok := payload[PayloadFieldName]; !ok {
			return fmt.Errorf("missing required field: %s", PayloadFieldName)
		}
	case DriveFunctionUploadFile:
		if _, ok := payload[PayloadFieldName]; !ok {
			return fmt.Errorf("missing required field: %s", PayloadFieldName)
		}
		if _, ok := payload[PayloadFieldContent]; !ok {
			return fmt.Errorf("missing required field: %s", PayloadFieldContent)
		}
	case DriveFunctionGetFile:
		if _, ok := payload[PayloadFieldFileID]; !ok {
			return fmt.Errorf("missing required field: %s", PayloadFieldFileID)
		}
	case DriveFunctionListFiles:
		// Optional parameters, no validation needed
	case DriveFunctionShareFile:
		if _, ok := payload[PayloadFieldFileID]; !ok {
			return fmt.Errorf("missing required field: %s", PayloadFieldFileID)
		}
		if _, ok := payload[PayloadFieldEmail]; !ok {
			return fmt.Errorf("missing required field: %s", PayloadFieldEmail)
		}
		if _, ok := payload[PayloadFieldRole]; !ok {
			return fmt.Errorf("missing required field: %s", PayloadFieldRole)
		}
	case DriveFunctionMoveFile:
		if _, ok := payload[PayloadFieldFileID]; !ok {
			return fmt.Errorf("missing required field: %s", PayloadFieldFileID)
		}
		if _, ok := payload["new_parent_id"]; !ok {
			return fmt.Errorf("missing required field: new_parent_id")
		}
	}
	return nil
}

// Private helper methods

func (p *DriveProxy) isSupportedFunction(function string) bool {
	supportedFunctions := p.GetSupportedFunctions()
	for _, f := range supportedFunctions {
		if f == function {
			return true
		}
	}
	return false
}

func (p *DriveProxy) createFolder(ctx context.Context, service *drive.Service, payload map[string]interface{}) (map[string]interface{}, error) {
	name := payload[PayloadFieldName].(string)

	folder := &drive.File{
		Name:     name,
		MimeType: "application/vnd.google-apps.folder",
	}

	// Set parent folder if specified
	if parentID, ok := payload[PayloadFieldParentID]; ok && parentID != "" {
		folder.Parents = []string{parentID.(string)}
	}

	createdFolder, err := service.Files.Create(folder).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create folder: %w", err)
	}

	return map[string]interface{}{
		"folder_id":  createdFolder.Id,
		"name":       createdFolder.Name,
		"url":        fmt.Sprintf("https://drive.google.com/drive/folders/%s", createdFolder.Id),
		"mime_type":  createdFolder.MimeType,
		"created_at": createdFolder.CreatedTime,
		"status":     "created",
	}, nil
}

func (p *DriveProxy) uploadFile(ctx context.Context, service *drive.Service, payload map[string]interface{}) (map[string]interface{}, error) {
	name := payload[PayloadFieldName].(string)
	content := payload[PayloadFieldContent].(string)

	file := &drive.File{
		Name: name,
	}

	// Set parent folder if specified
	if parentID, ok := payload[PayloadFieldParentID]; ok && parentID != "" {
		file.Parents = []string{parentID.(string)}
	}

	// Set MIME type if specified
	if mimeType, ok := payload["mime_type"]; ok && mimeType != "" {
		file.MimeType = mimeType.(string)
	}

	// Convert content string to reader
	contentReader := strings.NewReader(content)

	uploadedFile, err := service.Files.Create(file).Media(contentReader).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return map[string]interface{}{
		"file_id":    uploadedFile.Id,
		"name":       uploadedFile.Name,
		"url":        fmt.Sprintf("https://drive.google.com/file/d/%s/view", uploadedFile.Id),
		"mime_type":  uploadedFile.MimeType,
		"size":       uploadedFile.Size,
		"created_at": uploadedFile.CreatedTime,
		"status":     "uploaded",
	}, nil
}

func (p *DriveProxy) getFile(ctx context.Context, service *drive.Service, payload map[string]interface{}) (map[string]interface{}, error) {
	fileID := payload[PayloadFieldFileID].(string)

	file, err := service.Files.Get(fileID).Fields("id,name,mimeType,size,createdTime,modifiedTime,webViewLink,parents").Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	return map[string]interface{}{
		"file_id":      file.Id,
		"name":         file.Name,
		"mime_type":    file.MimeType,
		"size":         file.Size,
		"url":          file.WebViewLink,
		"created_at":   file.CreatedTime,
		"modified_at":  file.ModifiedTime,
		"parents":      file.Parents,
		"status":       "retrieved",
		"retrieved_at": time.Now().Format(time.RFC3339),
	}, nil
}

func (p *DriveProxy) listFiles(ctx context.Context, service *drive.Service, payload map[string]interface{}) (map[string]interface{}, error) {
	// Build query
	query := ""
	if folderID, ok := payload[PayloadFieldFolderID]; ok && folderID != "" {
		query = fmt.Sprintf("'%s' in parents", folderID.(string))
	}

	// Set page size
	pageSize := int64(10) // default
	if ps, ok := payload["page_size"]; ok {
		if psFloat, ok := ps.(float64); ok {
			pageSize = int64(psFloat)
		}
	}

	listCall := service.Files.List().PageSize(pageSize).Fields("files(id,name,mimeType,size,createdTime,parents)")
	if query != "" {
		listCall = listCall.Q(query)
	}

	fileList, err := listCall.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	files := make([]map[string]interface{}, 0, len(fileList.Files))
	for _, file := range fileList.Files {
		files = append(files, map[string]interface{}{
			"file_id":    file.Id,
			"name":       file.Name,
			"mime_type":  file.MimeType,
			"size":       file.Size,
			"created_at": file.CreatedTime,
			"parents":    file.Parents,
		})
	}

	return map[string]interface{}{
		"files":           files,
		"total_files":     len(files),
		"next_page_token": fileList.NextPageToken,
		"query":           query,
		"listed_at":       time.Now().Format(time.RFC3339),
	}, nil
}

func (p *DriveProxy) shareFile(ctx context.Context, service *drive.Service, payload map[string]interface{}) (map[string]interface{}, error) {
	fileID := payload[PayloadFieldFileID].(string)
	email := payload[PayloadFieldEmail].(string)
	role := payload[PayloadFieldRole].(string)

	permission := &drive.Permission{
		EmailAddress: email,
		Role:         role,
		Type:         "user",
	}

	createdPermission, err := service.Permissions.Create(fileID, permission).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to share file: %w", err)
	}

	return map[string]interface{}{
		"file_id":       fileID,
		"permission_id": createdPermission.Id,
		"email":         email,
		"role":          role,
		"type":          createdPermission.Type,
		"status":        "shared",
		"shared_at":     time.Now().Format(time.RFC3339),
	}, nil
}

func (p *DriveProxy) moveFile(ctx context.Context, service *drive.Service, payload map[string]interface{}) (map[string]interface{}, error) {
	fileID := payload[PayloadFieldFileID].(string)
	newParentID := payload["new_parent_id"].(string)

	// Get current parents
	file, err := service.Files.Get(fileID).Fields("parents").Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get file parents: %w", err)
	}

	// Move file by removing old parents and adding new parent
	var previousParents []string
	if len(file.Parents) > 0 {
		previousParents = file.Parents
	}

	updateCall := service.Files.Update(fileID, nil).AddParents(newParentID)
	if len(previousParents) > 0 {
		updateCall = updateCall.RemoveParents(strings.Join(previousParents, ","))
	}

	updatedFile, err := updateCall.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to move file: %w", err)
	}

	return map[string]interface{}{
		"file_id":          fileID,
		"new_parent_id":    newParentID,
		"previous_parents": previousParents,
		"current_parents":  updatedFile.Parents,
		"status":           "moved",
		"moved_at":         time.Now().Format(time.RFC3339),
	}, nil
}
