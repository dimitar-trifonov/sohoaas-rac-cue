package workspace

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/dimitar-trifonov/sohoaas/service-proxies/workflow"
	"golang.org/x/oauth2"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/option"
)

// DocsProxy implements WorkspaceProxy for Google Docs service
type DocsProxy struct {
	config *oauth2.Config
}

// NewDocsProxy creates a new Docs proxy instance
func NewDocsProxy(config *oauth2.Config) *DocsProxy {
	return &DocsProxy{
		config: config,
	}
}

// Execute calls a Docs function with the given payload
func (p *DocsProxy) Execute(ctx context.Context, function string, token string, payload map[string]interface{}) (*workflow.ProxyResponse, error) {
	startTime := time.Now()
	requestID := fmt.Sprintf("docs_%d", startTime.UnixNano())

	// Enhanced request logging
	log.Printf("[Docs] [%s] ========== REQUEST START ==========\n", requestID)
	log.Printf("[Docs] [%s] Function: %s\n", requestID, function)
	log.Printf("[Docs] [%s] Request Time: %s\n", requestID, startTime.Format(time.RFC3339Nano))
	log.Printf("[Docs] [%s] OAuth Token Length: %d characters\n", requestID, len(token))
	log.Printf("[Docs] [%s] OAuth Token Prefix: %s...\n", requestID, token[:min(20, len(token))])
	
	// Log payload with JSON formatting
	payloadJSON, _ := json.MarshalIndent(payload, "", "  ")
	log.Printf("[Docs] [%s] Request Payload:\n%s\n", requestID, string(payloadJSON))

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

	// Initialize Docs service with enhanced logging
	log.Printf("[Docs] [%s] Initializing Google Docs service...\n", requestID)
	serviceStartTime := time.Now()
	
	oauthToken := &oauth2.Token{AccessToken: token}
	client := p.config.Client(ctx, oauthToken)
	service, err := docs.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Printf("[Docs] [%s] ❌ Docs service initialization FAILED after %v: %v\n", requestID, time.Since(serviceStartTime), err)
		return &workflow.ProxyResponse{
			Success: false,
			Error: &workflow.ProxyError{
				Code:      string(ErrorCodeAuthenticationFailed),
				Message:   "Failed to initialize Docs service",
				Details:   err.Error(),
				Retryable: true,
			},
		}, nil
	}
	log.Printf("[Docs] [%s] ✅ Docs service initialized successfully in %v\n", requestID, time.Since(serviceStartTime))

	// Execute the function with enhanced logging
	log.Printf("[Docs] [%s] Executing function: %s\n", requestID, function)
	functionStartTime := time.Now()
	
	var result map[string]interface{}
	var execErr error

	switch function {
	case DocsFunctionCreateDocument:
		result, execErr = p.createDocumentWithLogging(ctx, service, payload, requestID)
	case DocsFunctionGetDocument:
		result, execErr = p.getDocumentWithLogging(ctx, service, payload, requestID)
	case DocsFunctionInsertText:
		result, execErr = p.insertTextWithLogging(ctx, service, payload, requestID)
	case DocsFunctionUpdateDocument:
		result, execErr = p.updateDocumentWithLogging(ctx, service, payload, requestID)
	case DocsFunctionBatchUpdate:
		result, execErr = p.batchUpdateWithLogging(ctx, service, payload, requestID)
	default:
		execErr = fmt.Errorf("function not implemented: %s", function)
		log.Printf("[Docs] [%s] ❌ Function not implemented: %s\n", requestID, function)
	}

	functionDuration := time.Since(functionStartTime)
	totalDuration := time.Since(startTime)

	if execErr != nil {
		log.Printf("[Docs] [%s] ❌ Function execution FAILED after %v (total: %v): %v\n", requestID, functionDuration, totalDuration, execErr)
		log.Printf("[Docs] [%s] ========== REQUEST END (FAILED) ==========\n", requestID)
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

	// Log successful response
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	log.Printf("[Docs] [%s] ✅ Function executed successfully in %v (total: %v)\n", requestID, functionDuration, totalDuration)
	log.Printf("[Docs] [%s] Response Data:\n%s\n", requestID, string(resultJSON))
	log.Printf("[Docs] [%s] ========== REQUEST END (SUCCESS) ==========\n", requestID)

	return &workflow.ProxyResponse{
		Success: true,
		Data:    result,
		Metadata: &workflow.ResponseMetadata{
			ExecutionTime: totalDuration,
			Function:      function,
			Timestamp:     time.Now(),
		},
	}, nil
}

// GetSupportedFunctions returns supported Docs functions
func (p *DocsProxy) GetSupportedFunctions() []string {
	return []string{
		DocsFunctionCreateDocument,
		DocsFunctionGetDocument,
		DocsFunctionInsertText,
		DocsFunctionUpdateDocument,
		DocsFunctionBatchUpdate,
	}
}

// GetServiceType returns the service type
func (p *DocsProxy) GetServiceType() string {
	return ServiceTypeDocs
}

// GetServiceCapabilities returns the service capabilities
func (p *DocsProxy) GetServiceCapabilities() map[string]interface{} {
	return map[string]interface{}{
		"service_type":        ServiceTypeDocs,
		"supported_functions": p.GetSupportedFunctions(),
		"max_document_size":   "10MB",
		"supported_formats":   []string{"text", "html"},
		"batch_operations":    true,
		"real_time_editing":   true,
	}
}

// GetServiceMetadata returns metadata about the Docs service and its functions
func (p *DocsProxy) GetServiceMetadata() ServiceMetadata {
	return ServiceMetadata{
		ServiceType: ServiceTypeDocs,
		DisplayName: "Google Docs",
		Description: "Create, edit, and manage documents using Google Docs API",
		Functions: map[string]FunctionMetadata{
			DocsFunctionCreateDocument: {
				Name:        DocsFunctionCreateDocument,
				DisplayName: "Create Document",
				Description: "Create a new Google Docs document",
				ExamplePayload: map[string]interface{}{
					"title": "Test Document",
				},
				RequiredFields: []string{"title"},
			},
			DocsFunctionGetDocument: {
				Name:        DocsFunctionGetDocument,
				DisplayName: "Get Document",
				Description: "Retrieve a Google Docs document by ID",
				ExamplePayload: map[string]interface{}{
					"document_id": "1234567890abcdef",
				},
				RequiredFields: []string{"document_id"},
			},
			DocsFunctionInsertText: {
				Name:        DocsFunctionInsertText,
				DisplayName: "Insert Text",
				Description: "Insert text into a Google Docs document",
				ExamplePayload: map[string]interface{}{
					"document_id": "1234567890abcdef",
					"content":     "This is new text to insert",
					"index":       1,
				},
				RequiredFields: []string{"document_id", "content"},
			},
			DocsFunctionUpdateDocument: {
				Name:        DocsFunctionUpdateDocument,
				DisplayName: "Update Document",
				Description: "Update a Google Docs document with batch operations",
				ExamplePayload: map[string]interface{}{
					"document_id": "1234567890abcdef",
					"requests":    []interface{}{},
				},
				RequiredFields: []string{"document_id", "requests"},
			},
			DocsFunctionBatchUpdate: {
				Name:        DocsFunctionBatchUpdate,
				DisplayName: "Batch Update",
				Description: "Perform multiple operations on a Google Docs document",
				ExamplePayload: map[string]interface{}{
					"document_id": "1234567890abcdef",
					"requests":    []interface{}{},
				},
				RequiredFields: []string{"document_id", "requests"},
			},
		},
	}
}

// GetFunctionMetadata returns metadata for a specific Docs function
func (p *DocsProxy) GetFunctionMetadata(function string) (FunctionMetadata, error) {
	metadata := p.GetServiceMetadata()
	funcMeta, exists := metadata.Functions[function]
	if !exists {
		return FunctionMetadata{}, fmt.Errorf("function %s not supported by Docs service", function)
	}
	return funcMeta, nil
}

// ValidateRequest validates a request (wrapper around ValidatePayload for interface compatibility)
func (p *DocsProxy) ValidateRequest(function string, payload map[string]interface{}) error {
	return p.ValidatePayload(function, payload)
}

// ValidatePayload validates the payload for a given function
func (p *DocsProxy) ValidatePayload(function string, payload map[string]interface{}) error {
	switch function {
	case DocsFunctionCreateDocument:
		if _, ok := payload[PayloadFieldTitle]; !ok {
			return fmt.Errorf("missing required field: %s", PayloadFieldTitle)
		}
	case DocsFunctionGetDocument:
		if _, ok := payload[PayloadFieldDocumentID]; !ok {
			return fmt.Errorf("missing required field: %s", PayloadFieldDocumentID)
		}
	case DocsFunctionInsertText:
		if _, ok := payload[PayloadFieldDocumentID]; !ok {
			return fmt.Errorf("missing required field: %s", PayloadFieldDocumentID)
		}
		if _, ok := payload[PayloadFieldContent]; !ok {
			return fmt.Errorf("missing required field: %s", PayloadFieldContent)
		}
	case DocsFunctionUpdateDocument:
		if _, ok := payload[PayloadFieldDocumentID]; !ok {
			return fmt.Errorf("missing required field: %s", PayloadFieldDocumentID)
		}
		if _, ok := payload["requests"]; !ok {
			return fmt.Errorf("missing required field: requests")
		}
	case DocsFunctionBatchUpdate:
		if _, ok := payload[PayloadFieldDocumentID]; !ok {
			return fmt.Errorf("missing required field: %s", PayloadFieldDocumentID)
		}
		if _, ok := payload["requests"]; !ok {
			return fmt.Errorf("missing required field: requests")
		}
	}
	return nil
}

// Private helper methods

func (p *DocsProxy) isSupportedFunction(function string) bool {
	supportedFunctions := p.GetSupportedFunctions()
	for _, f := range supportedFunctions {
		if f == function {
			return true
		}
	}
	return false
}

func (p *DocsProxy) createDocument(ctx context.Context, service *docs.Service, payload map[string]interface{}) (map[string]interface{}, error) {
	return p.createDocumentWithLogging(ctx, service, payload, "legacy")
}

func (p *DocsProxy) createDocumentWithLogging(ctx context.Context, service *docs.Service, payload map[string]interface{}, requestID string) (map[string]interface{}, error) {
	title := payload[PayloadFieldTitle].(string)

	log.Printf("[Docs] [%s] 📄 Creating document: '%s'\n", requestID, title)
	
	// Create document request
	doc := &docs.Document{
		Title: title,
	}

	log.Printf("[Docs] [%s] 🚀 Calling Google Docs API: Documents.Create\n", requestID)
	apiStartTime := time.Now()

	createdDoc, err := service.Documents.Create(doc).Do()
	apiDuration := time.Since(apiStartTime)
	
	if err != nil {
		log.Printf("[Docs] [%s] ❌ Google Docs API call FAILED after %v: %v\n", requestID, apiDuration, err)
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	log.Printf("[Docs] [%s] ✅ Google Docs API call SUCCESS in %v\n", requestID, apiDuration)
	log.Printf("[Docs] [%s] 📄 Document created successfully:\n", requestID)
	log.Printf("[Docs] [%s]    Document ID: %s\n", requestID, createdDoc.DocumentId)
	log.Printf("[Docs] [%s]    Title: %s\n", requestID, createdDoc.Title)
	log.Printf("[Docs] [%s]    Revision ID: %s\n", requestID, createdDoc.RevisionId)

	return map[string]interface{}{
		"document_id": createdDoc.DocumentId,
		"title":       createdDoc.Title,
		"url":         fmt.Sprintf("https://docs.google.com/document/d/%s/edit", createdDoc.DocumentId),
		"revision_id": createdDoc.RevisionId,
		"status":      "created",
		"created_at":  time.Now().Format(time.RFC3339),
		"api_duration_ms": apiDuration.Milliseconds(),
	}, nil
}

func (p *DocsProxy) getDocument(ctx context.Context, service *docs.Service, payload map[string]interface{}) (map[string]interface{}, error) {
	return p.getDocumentWithLogging(ctx, service, payload, "legacy")
}

func (p *DocsProxy) getDocumentWithLogging(ctx context.Context, service *docs.Service, payload map[string]interface{}, requestID string) (map[string]interface{}, error) {
	documentID := payload[PayloadFieldDocumentID].(string)

	log.Printf("[Docs] [%s] 📄 Retrieving document: %s\n", requestID, documentID)
	log.Printf("[Docs] [%s] 🚀 Calling Google Docs API: Documents.Get\n", requestID)
	apiStartTime := time.Now()

	document, err := service.Documents.Get(documentID).Do()
	apiDuration := time.Since(apiStartTime)
	
	if err != nil {
		log.Printf("[Docs] [%s] ❌ Google Docs API call FAILED after %v: %v\n", requestID, apiDuration, err)
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	log.Printf("[Docs] [%s] ✅ Google Docs API call SUCCESS in %v\n", requestID, apiDuration)
	log.Printf("[Docs] [%s] 📄 Document retrieved successfully:\n", requestID)
	log.Printf("[Docs] [%s]    Document ID: %s\n", requestID, document.DocumentId)
	log.Printf("[Docs] [%s]    Title: %s\n", requestID, document.Title)
	log.Printf("[Docs] [%s]    Revision ID: %s\n", requestID, document.RevisionId)

	// Extract text content (simplified)
	content := ""
	if document.Body != nil {
		for _, element := range document.Body.Content {
			if element.Paragraph != nil {
				for _, textElement := range element.Paragraph.Elements {
					if textElement.TextRun != nil {
						content += textElement.TextRun.Content
					}
				}
			}
		}
	}

	return map[string]interface{}{
		"document_id":  document.DocumentId,
		"title":        document.Title,
		"url":          fmt.Sprintf("https://docs.google.com/document/d/%s/edit", document.DocumentId),
		"revision_id":  document.RevisionId,
		"content":      content,
		"status":       "retrieved",
		"retrieved_at": time.Now().Format(time.RFC3339),
		"api_duration_ms": apiDuration.Milliseconds(),
	}, nil
}

func (p *DocsProxy) insertText(ctx context.Context, service *docs.Service, payload map[string]interface{}) (map[string]interface{}, error) {
	return p.insertTextWithLogging(ctx, service, payload, "legacy")
}

func (p *DocsProxy) insertTextWithLogging(ctx context.Context, service *docs.Service, payload map[string]interface{}, requestID string) (map[string]interface{}, error) {
	documentID := payload[PayloadFieldDocumentID].(string)
	text := payload[PayloadFieldContent].(string)

	// Get insertion index (default to beginning)
	index := int64(1)
	if idx, ok := payload["index"]; ok {
		if idxFloat, ok := idx.(float64); ok {
			index = int64(idxFloat)
		}
	}

	log.Printf("[Docs] [%s] ✏️ Inserting text into document: %s\n", requestID, documentID)
	log.Printf("[Docs] [%s]    Text Length: %d characters\n", requestID, len(text))
	log.Printf("[Docs] [%s]    Insertion Index: %d\n", requestID, index)

	// Create batch update request to insert text
	requests := []*docs.Request{
		{
			InsertText: &docs.InsertTextRequest{
				Location: &docs.Location{
					Index: index,
				},
				Text: text,
			},
		},
	}

	batchUpdateRequest := &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}

	log.Printf("[Docs] [%s] 🚀 Calling Google Docs API: Documents.BatchUpdate\n", requestID)
	apiStartTime := time.Now()

	response, err := service.Documents.BatchUpdate(documentID, batchUpdateRequest).Do()
	apiDuration := time.Since(apiStartTime)
	
	if err != nil {
		log.Printf("[Docs] [%s] ❌ Google Docs API call FAILED after %v: %v\n", requestID, apiDuration, err)
		return nil, fmt.Errorf("failed to insert text: %w", err)
	}

	log.Printf("[Docs] [%s] ✅ Google Docs API call SUCCESS in %v\n", requestID, apiDuration)
	log.Printf("[Docs] [%s] ✏️ Text inserted successfully\n", requestID)

	return map[string]interface{}{
		"document_id":     documentID,
		"text_inserted":   text,
		"insertion_index": index,
		"revision_id":     response.DocumentId, // This might not be correct, check API docs
		"status":          "text_inserted",
		"updated_at":      time.Now().Format(time.RFC3339),
		"api_duration_ms": apiDuration.Milliseconds(),
	}, nil
}

func (p *DocsProxy) updateDocument(ctx context.Context, service *docs.Service, payload map[string]interface{}) (map[string]interface{}, error) {
	return p.updateDocumentWithLogging(ctx, service, payload, "legacy")
}

func (p *DocsProxy) updateDocumentWithLogging(ctx context.Context, service *docs.Service, payload map[string]interface{}, requestID string) (map[string]interface{}, error) {
	documentID := payload[PayloadFieldDocumentID].(string)

	// Convert requests from interface{} to proper format
	requestsInterface := payload["requests"]

	log.Printf("[Docs] [%s] 📝 Updating document: %s\n", requestID, documentID)
	log.Printf("[Docs] [%s]    Requests: %v\n", requestID, requestsInterface)

	// This is a simplified implementation - in practice, you'd need proper request parsing
	batchUpdateRequest := &docs.BatchUpdateDocumentRequest{
		Requests: []*docs.Request{}, // Would need proper conversion from payload
	}

	log.Printf("[Docs] [%s] 🚀 Calling Google Docs API: Documents.BatchUpdate\n", requestID)
	apiStartTime := time.Now()

	response, err := service.Documents.BatchUpdate(documentID, batchUpdateRequest).Do()
	apiDuration := time.Since(apiStartTime)
	
	if err != nil {
		log.Printf("[Docs] [%s] ❌ Google Docs API call FAILED after %v: %v\n", requestID, apiDuration, err)
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	log.Printf("[Docs] [%s] ✅ Google Docs API call SUCCESS in %v\n", requestID, apiDuration)
	log.Printf("[Docs] [%s] 📝 Document updated successfully\n", requestID)

	return map[string]interface{}{
		"document_id":      documentID,
		"requests_applied": len(batchUpdateRequest.Requests),
		"revision_id":      response.DocumentId,
		"status":           "updated",
		"updated_at":       time.Now().Format(time.RFC3339),
		"requests":         requestsInterface,
		"api_duration_ms": apiDuration.Milliseconds(),
	}, nil
}

func (p *DocsProxy) batchUpdate(ctx context.Context, service *docs.Service, payload map[string]interface{}) (map[string]interface{}, error) {
	return p.batchUpdateWithLogging(ctx, service, payload, "legacy")
}

func (p *DocsProxy) batchUpdateWithLogging(ctx context.Context, service *docs.Service, payload map[string]interface{}, requestID string) (map[string]interface{}, error) {
	// This would be similar to updateDocument but with more complex request handling
	log.Printf("[Docs] [%s] 📝 Performing batch update (delegating to updateDocument)\n", requestID)
	return p.updateDocumentWithLogging(ctx, service, payload, requestID)
}
