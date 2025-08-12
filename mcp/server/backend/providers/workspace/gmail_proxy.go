package workspace

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/dimitar-trifonov/sohoaas/service-proxies/workflow"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// GmailProxy implements WorkspaceProxy for Gmail service
type GmailProxy struct {
	config *oauth2.Config
}

// NewGmailProxy creates a new Gmail proxy instance
func NewGmailProxy(config *oauth2.Config) *GmailProxy {
	return &GmailProxy{
		config: config,
	}
}

// Execute calls a Gmail function with the given payload
func (p *GmailProxy) Execute(ctx context.Context, function string, token string, payload map[string]interface{}) (*workflow.ProxyResponse, error) {
	startTime := time.Now()
	requestID := fmt.Sprintf("gmail_%d", startTime.UnixNano())

	// Enhanced request logging
	log.Printf("[Gmail] [%s] ========== REQUEST START ==========\n", requestID)
	log.Printf("[Gmail] [%s] Function: %s\n", requestID, function)
	log.Printf("[Gmail] [%s] Request Time: %s\n", requestID, startTime.Format(time.RFC3339Nano))
	log.Printf("[Gmail] [%s] OAuth Token Length: %d characters\n", requestID, len(token))
	log.Printf("[Gmail] [%s] OAuth Token Prefix: %s...\n", requestID, token[:min(20, len(token))])
	
	// Log payload with JSON formatting
	payloadJSON, _ := json.MarshalIndent(payload, "", "  ")
	log.Printf("[Gmail] [%s] Request Payload:\n%s\n", requestID, string(payloadJSON))

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

	// Initialize Gmail service with enhanced logging
	log.Printf("[Gmail] [%s] Initializing Gmail service...\n", requestID)
	serviceStartTime := time.Now()
	
	oauthToken := &oauth2.Token{AccessToken: token}
	client := p.config.Client(ctx, oauthToken)
	service, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Printf("[Gmail] [%s] ❌ Gmail service initialization FAILED after %v: %v\n", requestID, time.Since(serviceStartTime), err)
		return &workflow.ProxyResponse{
			Success: false,
			Error: &workflow.ProxyError{
				Code:      string(ErrorCodeAuthenticationFailed),
				Message:   "Failed to initialize Gmail service",
				Details:   err.Error(),
				Retryable: true,
			},
		}, nil
	}
	log.Printf("[Gmail] [%s] ✅ Gmail service initialized successfully in %v\n", requestID, time.Since(serviceStartTime))

	// Execute the function with enhanced logging
	log.Printf("[Gmail] [%s] Executing function: %s\n", requestID, function)
	functionStartTime := time.Now()
	
	var result map[string]interface{}
	var execErr error

	switch function {
	case GmailFunctionSendMessage:
		result, execErr = p.sendMessageWithLogging(ctx, service, payload, requestID)
	case GmailFunctionGetMessage:
		result, execErr = p.getMessageWithLogging(ctx, service, payload, requestID)
	case GmailFunctionListMessages:
		result, execErr = p.listMessagesWithLogging(ctx, service, payload, requestID)
	case GmailFunctionSearchMessages:
		result, execErr = p.searchMessagesWithLogging(ctx, service, payload, requestID)
	default:
		execErr = fmt.Errorf("function not implemented: %s", function)
		log.Printf("[Gmail] [%s] ❌ Function not implemented: %s\n", requestID, function)
	}

	functionDuration := time.Since(functionStartTime)
	totalDuration := time.Since(startTime)

	if execErr != nil {
		log.Printf("[Gmail] [%s] ❌ Function execution FAILED after %v (total: %v): %v\n", requestID, functionDuration, totalDuration, execErr)
		log.Printf("[Gmail] [%s] ========== REQUEST END (FAILED) ==========\n", requestID)
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
	log.Printf("[Gmail] [%s] ✅ Function executed successfully in %v (total: %v)\n", requestID, functionDuration, totalDuration)
	log.Printf("[Gmail] [%s] Response Data:\n%s\n", requestID, string(resultJSON))
	log.Printf("[Gmail] [%s] ========== REQUEST END (SUCCESS) ==========\n", requestID)

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

// GetSupportedFunctions returns supported Gmail functions
func (p *GmailProxy) GetSupportedFunctions() []string {
	return []string{
		GmailFunctionSendMessage,
		GmailFunctionGetMessage,
		GmailFunctionListMessages,
		GmailFunctionSearchMessages,
	}
}

// GetServiceType returns the service type
func (p *GmailProxy) GetServiceType() string {
	return ServiceTypeGmail
}

// GetServiceCapabilities returns the service capabilities
func (p *GmailProxy) GetServiceCapabilities() map[string]interface{} {
	return map[string]interface{}{
		"service_type":        ServiceTypeGmail,
		"supported_functions": p.GetSupportedFunctions(),
		"max_message_size":    "25MB",
		"attachments":         true,
		"threading":           true,
		"search_capabilities": true,
		"labels":              true,
	}
}

// GetServiceMetadata returns metadata about the Gmail service and its functions
func (p *GmailProxy) GetServiceMetadata() ServiceMetadata {
	return ServiceMetadata{
		ServiceType: ServiceTypeGmail,
		DisplayName: "Gmail",
		Description: "Send, receive, and manage emails using Gmail API",
		Functions: map[string]FunctionMetadata{
			GmailFunctionSendMessage: {
				Name:        GmailFunctionSendMessage,
				DisplayName: "Send Email",
				Description: "Send an email message via Gmail",
				ExamplePayload: map[string]interface{}{
					"to":      "recipient@example.com",
					"subject": "Test Email",
					"body":    "This is a test email from SOHOaaS",
				},
				RequiredFields: []string{"to", "subject", "body"},
			},
			GmailFunctionSearchMessages: {
				Name:        GmailFunctionSearchMessages,
				DisplayName: "Search Emails",
				Description: "Search for emails in Gmail using query syntax",
				ExamplePayload: map[string]interface{}{
					"query":       "from:noreply@example.com",
					"max_results": 10,
				},
				RequiredFields: []string{"query"},
			},
			GmailFunctionGetMessage: {
				Name:        GmailFunctionGetMessage,
				DisplayName: "Get Email",
				Description: "Retrieve a specific email message by ID",
				ExamplePayload: map[string]interface{}{
					"message_id": "1234567890abcdef",
				},
				RequiredFields: []string{"message_id"},
			},
			GmailFunctionListMessages: {
				Name:        GmailFunctionListMessages,
				DisplayName: "List Emails",
				Description: "List email messages with optional filtering",
				ExamplePayload: map[string]interface{}{
					"max_results": 10,
					"query":       "is:unread",
				},
				RequiredFields: []string{},
			},
		},
	}
}

// GetFunctionMetadata returns metadata for a specific Gmail function
func (p *GmailProxy) GetFunctionMetadata(function string) (FunctionMetadata, error) {
	metadata := p.GetServiceMetadata()
	funcMeta, exists := metadata.Functions[function]
	if !exists {
		return FunctionMetadata{}, fmt.Errorf("function %s not supported by Gmail service", function)
	}
	return funcMeta, nil
}

// ValidateRequest validates a request (wrapper around ValidatePayload for interface compatibility)
func (p *GmailProxy) ValidateRequest(function string, payload map[string]interface{}) error {
	return p.ValidatePayload(function, payload)
}

// ValidatePayload validates the payload for a given function
func (p *GmailProxy) ValidatePayload(function string, payload map[string]interface{}) error {
	switch function {
	case GmailFunctionSendMessage:
		if _, ok := payload[PayloadFieldTo]; !ok {
			return fmt.Errorf("missing required field: %s", PayloadFieldTo)
		}
		if _, ok := payload[PayloadFieldSubject]; !ok {
			return fmt.Errorf("missing required field: %s", PayloadFieldSubject)
		}
		if _, ok := payload[PayloadFieldBody]; !ok {
			return fmt.Errorf("missing required field: %s", PayloadFieldBody)
		}
	case GmailFunctionGetMessage:
		if _, ok := payload["message_id"]; !ok {
			return fmt.Errorf("missing required field: message_id")
		}
	case GmailFunctionListMessages:
		// Optional parameters, no validation needed
	case GmailFunctionSearchMessages:
		if _, ok := payload["query"]; !ok {
			return fmt.Errorf("missing required field: query")
		}
	}
	return nil
}

// Private helper methods

func (p *GmailProxy) isSupportedFunction(function string) bool {
	supportedFunctions := p.GetSupportedFunctions()
	for _, f := range supportedFunctions {
		if f == function {
			return true
		}
	}
	return false
}

func (p *GmailProxy) sendMessage(ctx context.Context, service *gmail.Service, payload map[string]interface{}) (map[string]interface{}, error) {
	return p.sendMessageWithLogging(ctx, service, payload, "legacy")
}

func (p *GmailProxy) sendMessageWithLogging(ctx context.Context, service *gmail.Service, payload map[string]interface{}, requestID string) (map[string]interface{}, error) {
	to := payload[PayloadFieldTo].(string)
	subject := payload[PayloadFieldSubject].(string)
	body := payload[PayloadFieldBody].(string)

	log.Printf("[Gmail] [%s] 📧 Preparing to send email\n", requestID)
	log.Printf("[Gmail] [%s]    To: %s\n", requestID, to)
	log.Printf("[Gmail] [%s]    Subject: %s\n", requestID, subject)
	log.Printf("[Gmail] [%s]    Body Length: %d characters\n", requestID, len(body))

	// Create email message
	rawMessage := p.createRawMessage(to, subject, body)
	log.Printf("[Gmail] [%s] 📝 Raw message created (length: %d)\n", requestID, len(rawMessage))
	
	message := &gmail.Message{
		Raw: rawMessage,
	}

	// Send message with timing
	log.Printf("[Gmail] [%s] 🚀 Calling Gmail API: Users.Messages.Send\n", requestID)
	apiStartTime := time.Now()
	
	sentMessage, err := service.Users.Messages.Send("me", message).Do()
	apiDuration := time.Since(apiStartTime)
	
	if err != nil {
		log.Printf("[Gmail] [%s] ❌ Gmail API call FAILED after %v: %v\n", requestID, apiDuration, err)
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	log.Printf("[Gmail] [%s] ✅ Gmail API call SUCCESS in %v\n", requestID, apiDuration)
	log.Printf("[Gmail] [%s] 📨 Message sent successfully:\n", requestID)
	log.Printf("[Gmail] [%s]    Message ID: %s\n", requestID, sentMessage.Id)
	log.Printf("[Gmail] [%s]    Thread ID: %s\n", requestID, sentMessage.ThreadId)
	log.Printf("[Gmail] [%s]    Label IDs: %v\n", requestID, sentMessage.LabelIds)
	log.Printf("[Gmail] [%s]    Snippet: %s\n", requestID, sentMessage.Snippet)

	return map[string]interface{}{
		"message_id": sentMessage.Id,
		"thread_id":  sentMessage.ThreadId,
		"label_ids":  sentMessage.LabelIds,
		"snippet":    sentMessage.Snippet,
		"to":         to,
		"subject":    subject,
		"status":     "sent",
		"sent_at":    time.Now().Format(time.RFC3339),
		"api_duration_ms": apiDuration.Milliseconds(),
	}, nil
}

func (p *GmailProxy) getMessage(ctx context.Context, service *gmail.Service, payload map[string]interface{}) (map[string]interface{}, error) {
	return p.getMessageWithLogging(ctx, service, payload, "legacy")
}

func (p *GmailProxy) getMessageWithLogging(ctx context.Context, service *gmail.Service, payload map[string]interface{}, requestID string) (map[string]interface{}, error) {
	messageID := payload["message_id"].(string)

	log.Printf("[Gmail] [%s] 📬 Retrieving message: %s\n", requestID, messageID)
	log.Printf("[Gmail] [%s] 🚀 Calling Gmail API: Users.Messages.Get\n", requestID)
	apiStartTime := time.Now()

	message, err := service.Users.Messages.Get("me", messageID).Do()
	apiDuration := time.Since(apiStartTime)
	
	if err != nil {
		log.Printf("[Gmail] [%s] ❌ Gmail API call FAILED after %v: %v\n", requestID, apiDuration, err)
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	log.Printf("[Gmail] [%s] ✅ Gmail API call SUCCESS in %v\n", requestID, apiDuration)
	log.Printf("[Gmail] [%s] 📨 Message retrieved successfully:\n", requestID)
	log.Printf("[Gmail] [%s]    Message ID: %s\n", requestID, message.Id)
	log.Printf("[Gmail] [%s]    Thread ID: %s\n", requestID, message.ThreadId)
	log.Printf("[Gmail] [%s]    Size Estimate: %d bytes\n", requestID, message.SizeEstimate)
	log.Printf("[Gmail] [%s]    Label IDs: %v\n", requestID, message.LabelIds)

	// Extract headers
	headers := make(map[string]string)
	for _, header := range message.Payload.Headers {
		headers[header.Name] = header.Value
	}

	return map[string]interface{}{
		"message_id":    message.Id,
		"thread_id":     message.ThreadId,
		"label_ids":     message.LabelIds,
		"snippet":       message.Snippet,
		"history_id":    message.HistoryId,
		"internal_date": message.InternalDate,
		"size_estimate": message.SizeEstimate,
		"headers":       headers,
		"subject":       headers["Subject"],
		"from":          headers["From"],
		"to":            headers["To"],
		"date":          headers["Date"],
	}, nil
}

func (p *GmailProxy) listMessages(ctx context.Context, service *gmail.Service, payload map[string]interface{}) (map[string]interface{}, error) {
	return p.listMessagesWithLogging(ctx, service, payload, "legacy")
}

func (p *GmailProxy) listMessagesWithLogging(ctx context.Context, service *gmail.Service, payload map[string]interface{}, requestID string) (map[string]interface{}, error) {
	// Optional parameters
	maxResults := int64(10) // default
	if mr, ok := payload["max_results"]; ok {
		if mrInt, ok := mr.(float64); ok {
			maxResults = int64(mrInt)
		}
	}

	query := ""
	if q, ok := payload["query"]; ok {
		query = q.(string)
	}

	log.Printf("[Gmail] [%s] 📋 Listing messages (maxResults: %d, query: '%s')\n", requestID, maxResults, query)
	log.Printf("[Gmail] [%s] 🚀 Calling Gmail API: Users.Messages.List\n", requestID)
	apiStartTime := time.Now()

	listCall := service.Users.Messages.List("me").MaxResults(maxResults)
	if query != "" {
		listCall = listCall.Q(query)
	}

	messageList, err := listCall.Do()
	apiDuration := time.Since(apiStartTime)
	
	if err != nil {
		log.Printf("[Gmail] [%s] ❌ Gmail API call FAILED after %v: %v\n", requestID, apiDuration, err)
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	log.Printf("[Gmail] [%s] ✅ Gmail API call SUCCESS in %v\n", requestID, apiDuration)
	log.Printf("[Gmail] [%s] 📋 Messages listed successfully: %d messages found\n", requestID, len(messageList.Messages))

	messages := make([]map[string]interface{}, 0, len(messageList.Messages))
	for _, msg := range messageList.Messages {
		messages = append(messages, map[string]interface{}{
			"message_id": msg.Id,
			"thread_id":  msg.ThreadId,
		})
	}

	return map[string]interface{}{
		"messages":             messages,
		"next_page_token":      messageList.NextPageToken,
		"result_size_estimate": messageList.ResultSizeEstimate,
		"total_messages":       len(messages),
		"api_duration_ms":      apiDuration.Milliseconds(),
	}, nil
}

func (p *GmailProxy) searchMessages(ctx context.Context, service *gmail.Service, payload map[string]interface{}) (map[string]interface{}, error) {
	return p.searchMessagesWithLogging(ctx, service, payload, "legacy")
}

func (p *GmailProxy) searchMessagesWithLogging(ctx context.Context, service *gmail.Service, payload map[string]interface{}, requestID string) (map[string]interface{}, error) {
	query := payload["query"].(string)

	maxResults := int64(10) // default
	if mr, ok := payload["max_results"]; ok {
		if mrInt, ok := mr.(float64); ok {
			maxResults = int64(mrInt)
		}
	}

	log.Printf("[Gmail] [%s] 🔍 Searching messages (query: '%s', maxResults: %d)\n", requestID, query, maxResults)
	log.Printf("[Gmail] [%s] 🚀 Calling Gmail API: Users.Messages.List (with query)\n", requestID)
	apiStartTime := time.Now()

	messageList, err := service.Users.Messages.List("me").Q(query).MaxResults(maxResults).Do()
	apiDuration := time.Since(apiStartTime)
	
	if err != nil {
		log.Printf("[Gmail] [%s] ❌ Gmail API call FAILED after %v: %v\n", requestID, apiDuration, err)
		return nil, fmt.Errorf("failed to search messages: %w", err)
	}

	log.Printf("[Gmail] [%s] ✅ Gmail API call SUCCESS in %v\n", requestID, apiDuration)
	log.Printf("[Gmail] [%s] 🔍 Search completed: %d matches found\n", requestID, len(messageList.Messages))

	messages := make([]map[string]interface{}, 0, len(messageList.Messages))
	for _, msg := range messageList.Messages {
		messages = append(messages, map[string]interface{}{
			"message_id": msg.Id,
			"thread_id":  msg.ThreadId,
		})
	}

	return map[string]interface{}{
		"query":                query,
		"messages":             messages,
		"next_page_token":      messageList.NextPageToken,
		"result_size_estimate": messageList.ResultSizeEstimate,
		"total_matches":        len(messages),
		"api_duration_ms":      apiDuration.Milliseconds(),
	}, nil
}

func (p *GmailProxy) createRawMessage(to, subject, body string) string {
	// Create RFC 2822 compliant email message with proper headers
	message := fmt.Sprintf(
		"From: me\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+
		"MIME-Version: 1.0\r\n"+
		"\r\n"+
		"%s",
		to, subject, body)

	// Base64 encode the message (Gmail requires base64url encoding)
	return base64.URLEncoding.EncodeToString([]byte(message))
}
