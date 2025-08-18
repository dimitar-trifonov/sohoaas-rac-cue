package workspace

import (
	"context"
	"fmt"
	"time"

	"github.com/dimitar-trifonov/sohoaas/service-proxies/workflow"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// CalendarProxy implements WorkspaceProxy for Google Calendar service
type CalendarProxy struct {
	config *oauth2.Config
}

// NewCalendarProxy creates a new Calendar proxy instance
func NewCalendarProxy(config *oauth2.Config) *CalendarProxy {
	return &CalendarProxy{
		config: config,
	}
}

// Execute calls a Calendar function with the given payload
func (p *CalendarProxy) Execute(ctx context.Context, function string, token string, payload map[string]interface{}) (*workflow.ProxyResponse, error) {
	startTime := time.Now()

	// Debug logging
	fmt.Printf("[Calendar] Executing function: %s\n", function)
	fmt.Printf("[Calendar] Payload: %+v\n", payload)
	fmt.Printf("[Calendar] Token length: %d\n", len(token))

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

	// Initialize Calendar service
	oauthToken := &oauth2.Token{AccessToken: token}
	client := p.config.Client(ctx, oauthToken)
	service, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return &workflow.ProxyResponse{
			Success: false,
			Error: &workflow.ProxyError{
				Code:      string(ErrorCodeAuthenticationFailed),
				Message:   "Failed to initialize Calendar service",
				Details:   err.Error(),
				Retryable: true,
			},
		}, nil
	}

	// Execute the function
	var result map[string]interface{}
	var execErr error

	switch function {
	case CalendarFunctionCreateEvent:
		result, execErr = p.createEvent(ctx, service, payload)
	case CalendarFunctionGetEvent:
		result, execErr = p.getEvent(ctx, service, payload)
	case CalendarFunctionListEvents:
		result, execErr = p.listEvents(ctx, service, payload)
	case CalendarFunctionUpdateEvent:
		result, execErr = p.updateEvent(ctx, service, payload)
	case CalendarFunctionDeleteEvent:
		result, execErr = p.deleteEvent(ctx, service, payload)
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

// GetSupportedFunctions returns supported Calendar functions
func (p *CalendarProxy) GetSupportedFunctions() []string {
	return []string{
		CalendarFunctionCreateEvent,
		CalendarFunctionGetEvent,
		CalendarFunctionListEvents,
		CalendarFunctionUpdateEvent,
		CalendarFunctionDeleteEvent,
	}
}

// GetServiceType returns the service type
func (p *CalendarProxy) GetServiceType() string {
	return ServiceTypeCalendar
}

// GetServiceCapabilities returns the service capabilities
func (p *CalendarProxy) GetServiceCapabilities() map[string]interface{} {
	return map[string]interface{}{
		"service_type":        ServiceTypeCalendar,
		"supported_functions": p.GetSupportedFunctions(),
		"timezone_support":    true,
		"recurring_events":    true,
		"attendees":          true,
		"reminders":          true,
		"attachments":        true,
	}
}

// GetServiceMetadata returns metadata about the Calendar service and its functions
func (p *CalendarProxy) GetServiceMetadata() ServiceMetadata {
	return ServiceMetadata{
		ServiceType: ServiceTypeCalendar,
		DisplayName: "Google Calendar",
		Description: "Create, manage, and organize calendar events using Google Calendar API",
		Functions: map[string]FunctionMetadata{
			CalendarFunctionCreateEvent: {
				Name:        CalendarFunctionCreateEvent,
				DisplayName: "Create Event",
				Description: "Create a new calendar event",
				ExamplePayload: map[string]interface{}{
					"title":       "Meeting with client",
					"description": "Discuss project requirements",
					"startTime":   "2025-07-30T14:00:00Z",
					"endTime":     "2025-07-30T15:00:00Z",
					"attendees":   []string{"client@example.com"},
				},
				RequiredFields: []string{"title", "startTime", "endTime"},
				OutputSchema: &ResponseSchema{
					Type:        "object",
					Description: "Calendar event creation response",
					Properties: map[string]PropertySchema{
						"event_id": {
							Type:        "string",
							Description: "Google Calendar event ID",
						},
						"html_link": {
							Type:        "string",
							Description: "Event HTML link",
						},
						"title": {
							Type:        "string",
							Description: "Event title",
						},
						"description": {
							Type:        "string",
							Description: "Event description",
						},
						"start_time": {
							Type:        "string",
							Description: "Event start time",
						},
						"end_time": {
							Type:        "string",
							Description: "Event end time",
						},
						"status": {
							Type:        "string",
							Description: "Event status",
						},
						"created_at": {
							Type:        "string",
							Description: "ISO timestamp when created",
						},
						"updated_at": {
							Type:        "string",
							Description: "ISO timestamp when updated",
						},
					},
					Required: []string{"event_id", "title", "start_time", "end_time", "status"},
				},
				ErrorSchema: &ResponseSchema{
					Type:        "object",
					Description: "Calendar event creation error response",
					Properties: map[string]PropertySchema{
						"error_code": {
							Type:        "string",
							Description: "Error code",
						},
						"error_message": {
							Type:        "string",
							Description: "Error message",
						},
						"details": {
							Type:        "object",
							Description: "Additional error details",
						},
					},
					Required: []string{"error_code", "error_message"},
				},
			},
			CalendarFunctionGetEvent: {
				Name:        CalendarFunctionGetEvent,
				DisplayName: "Get Event",
				Description: "Retrieve details of a specific calendar event",
				ExamplePayload: map[string]interface{}{
					"event_id": "event123456",
				},
				RequiredFields: []string{"event_id"},
			},
			CalendarFunctionListEvents: {
				Name:        CalendarFunctionListEvents,
				DisplayName: "List Events",
				Description: "List calendar events within a time range",
				ExamplePayload: map[string]interface{}{
					"time_min":    "2025-07-30T00:00:00Z",
					"time_max":    "2025-07-31T00:00:00Z",
					"max_results": 10,
				},
				RequiredFields: []string{},
			},
			CalendarFunctionUpdateEvent: {
				Name:        CalendarFunctionUpdateEvent,
				DisplayName: "Update Event",
				Description: "Update an existing calendar event",
				ExamplePayload: map[string]interface{}{
					"event_id":    "event123456",
					"title":       "Updated meeting title",
					"description": "Updated description",
				},
				RequiredFields: []string{"event_id"},
			},
			CalendarFunctionDeleteEvent: {
				Name:        CalendarFunctionDeleteEvent,
				DisplayName: "Delete Event",
				Description: "Delete a calendar event",
				ExamplePayload: map[string]interface{}{
					"event_id": "event123456",
				},
				RequiredFields: []string{"event_id"},
			},
		},
	}
}

// GetFunctionMetadata returns metadata for a specific Calendar function
func (p *CalendarProxy) GetFunctionMetadata(function string) (FunctionMetadata, error) {
	metadata := p.GetServiceMetadata()
	if funcMetadata, exists := metadata.Functions[function]; exists {
		return funcMetadata, nil
	}
	return FunctionMetadata{}, fmt.Errorf("function %s not found", function)
}

// ValidateRequest validates a request (wrapper around ValidatePayload for interface compatibility)
func (p *CalendarProxy) ValidateRequest(function string, payload map[string]interface{}) error {
	return p.ValidatePayload(function, payload)
}

// ValidatePayload validates the payload for a given function
func (p *CalendarProxy) ValidatePayload(function string, payload map[string]interface{}) error {
	metadata, err := p.GetFunctionMetadata(function)
	if err != nil {
		return err
	}

	// Check required fields
	for _, field := range metadata.RequiredFields {
		if _, exists := payload[field]; !exists {
			return fmt.Errorf("required field '%s' is missing", field)
		}
	}

	return nil
}

// Private helper methods

func (p *CalendarProxy) isSupportedFunction(function string) bool {
	supportedFunctions := p.GetSupportedFunctions()
	for _, supportedFunc := range supportedFunctions {
		if supportedFunc == function {
			return true
		}
	}
	return false
}

func (p *CalendarProxy) createEvent(ctx context.Context, service *calendar.Service, payload map[string]interface{}) (map[string]interface{}, error) {
	title := payload["title"].(string)
	startTime := payload["startTime"].(string)
	endTime := payload["endTime"].(string)
	
	description := ""
	if desc, ok := payload["description"]; ok {
		description = desc.(string)
	}

	// Debug logging
	fmt.Printf("[Calendar] createEvent - Title: %s\n", title)
	fmt.Printf("[Calendar] createEvent - Start: %s, End: %s\n", startTime, endTime)
	fmt.Printf("[Calendar] createEvent - Making Calendar API call...\n")

	event := &calendar.Event{
		Summary:     title,
		Description: description,
		Start: &calendar.EventDateTime{
			DateTime: startTime,
		},
		End: &calendar.EventDateTime{
			DateTime: endTime,
		},
	}

	// Add attendees if provided
	if attendeesData, ok := payload["attendees"]; ok {
		if attendeesList, ok := attendeesData.([]interface{}); ok {
			attendees := make([]*calendar.EventAttendee, 0, len(attendeesList))
			for _, attendeeEmail := range attendeesList {
				if email, ok := attendeeEmail.(string); ok {
					attendees = append(attendees, &calendar.EventAttendee{
						Email: email,
					})
				}
			}
			event.Attendees = attendees
		}
	}

	createdEvent, err := service.Events.Insert("primary", event).Do()
	if err != nil {
		fmt.Printf("[Calendar] createEvent - Calendar API Error: %v\n", err)
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	fmt.Printf("[Calendar] createEvent - Success! Event created: %s\n", createdEvent.Id)

	return map[string]interface{}{
		"event_id":    createdEvent.Id,
		"html_link":   createdEvent.HtmlLink,
		"title":       createdEvent.Summary,
		"description": createdEvent.Description,
		"start_time":  createdEvent.Start.DateTime,
		"end_time":    createdEvent.End.DateTime,
		"status":      createdEvent.Status,
		"created_at":  createdEvent.Created,
		"updated_at":  createdEvent.Updated,
	}, nil
}

func (p *CalendarProxy) getEvent(ctx context.Context, service *calendar.Service, payload map[string]interface{}) (map[string]interface{}, error) {
	eventID := payload["event_id"].(string)

	// Debug logging
	fmt.Printf("[Calendar] getEvent - Event ID: %s\n", eventID)
	fmt.Printf("[Calendar] getEvent - Making Calendar API call...\n")

	event, err := service.Events.Get("primary", eventID).Do()
	if err != nil {
		fmt.Printf("[Calendar] getEvent - Calendar API Error: %v\n", err)
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	fmt.Printf("[Calendar] getEvent - Success! Event retrieved: %s\n", event.Id)

	return map[string]interface{}{
		"event_id":    event.Id,
		"html_link":   event.HtmlLink,
		"title":       event.Summary,
		"description": event.Description,
		"start_time":  event.Start.DateTime,
		"end_time":    event.End.DateTime,
		"status":      event.Status,
		"created_at":  event.Created,
		"updated_at":  event.Updated,
		"attendees":   event.Attendees,
	}, nil
}

func (p *CalendarProxy) listEvents(ctx context.Context, service *calendar.Service, payload map[string]interface{}) (map[string]interface{}, error) {
	// Optional parameters
	maxResults := int64(10) // default
	if mr, ok := payload["max_results"]; ok {
		if mrInt, ok := mr.(float64); ok {
			maxResults = int64(mrInt)
		}
	}

	listCall := service.Events.List("primary").MaxResults(maxResults).SingleEvents(true).OrderBy("startTime")

	if timeMin, ok := payload["time_min"]; ok {
		listCall = listCall.TimeMin(timeMin.(string))
	}

	if timeMax, ok := payload["time_max"]; ok {
		listCall = listCall.TimeMax(timeMax.(string))
	}

	events, err := listCall.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	eventList := make([]map[string]interface{}, 0, len(events.Items))
	for _, event := range events.Items {
		eventList = append(eventList, map[string]interface{}{
			"event_id":   event.Id,
			"title":      event.Summary,
			"start_time": event.Start.DateTime,
			"end_time":   event.End.DateTime,
			"status":     event.Status,
		})
	}

	return map[string]interface{}{
		"events":      eventList,
		"next_token":  events.NextPageToken,
		"total_count": len(eventList),
	}, nil
}

func (p *CalendarProxy) updateEvent(ctx context.Context, service *calendar.Service, payload map[string]interface{}) (map[string]interface{}, error) {
	eventID := payload["event_id"].(string)

	// Get existing event
	event, err := service.Events.Get("primary", eventID).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get event for update: %w", err)
	}

	// Update fields if provided
	if title, ok := payload["title"]; ok {
		event.Summary = title.(string)
	}
	if description, ok := payload["description"]; ok {
		event.Description = description.(string)
	}
	if startTime, ok := payload["startTime"]; ok {
		event.Start.DateTime = startTime.(string)
	}
	if endTime, ok := payload["endTime"]; ok {
		event.End.DateTime = endTime.(string)
	}

	updatedEvent, err := service.Events.Update("primary", eventID, event).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to update event: %w", err)
	}

	return map[string]interface{}{
		"event_id":    updatedEvent.Id,
		"title":       updatedEvent.Summary,
		"description": updatedEvent.Description,
		"start_time":  updatedEvent.Start.DateTime,
		"end_time":    updatedEvent.End.DateTime,
		"status":      updatedEvent.Status,
		"updated_at":  updatedEvent.Updated,
	}, nil
}

func (p *CalendarProxy) deleteEvent(ctx context.Context, service *calendar.Service, payload map[string]interface{}) (map[string]interface{}, error) {
	eventID := payload["event_id"].(string)

	// Debug logging
	fmt.Printf("[Calendar] deleteEvent - Event ID: %s\n", eventID)
	fmt.Printf("[Calendar] deleteEvent - Making Calendar API call...\n")

	err := service.Events.Delete("primary", eventID).Do()
	if err != nil {
		fmt.Printf("[Calendar] deleteEvent - Calendar API Error: %v\n", err)
		return nil, fmt.Errorf("failed to delete event: %w", err)
	}

	fmt.Printf("[Calendar] deleteEvent - Success! Event deleted: %s\n", eventID)

	return map[string]interface{}{
		"event_id": eventID,
		"status":   "deleted",
		"deleted_at": time.Now().Format(time.RFC3339),
	}, nil
}
