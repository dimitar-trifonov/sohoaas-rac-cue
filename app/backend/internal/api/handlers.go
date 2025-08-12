package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"sohoaas-backend/internal/manager"
	"sohoaas-backend/internal/services"
	"sohoaas-backend/internal/types"

	"github.com/gin-gonic/gin"
)

// getMapKeys returns the keys of a map[string]interface{} for logging
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Handler contains all the dependencies for API handlers
type Handler struct {
	agentManager    *manager.AgentManager
	mcpService      *services.MCPService
	workflowStorage *services.WorkflowStorageService
	executionEngine *services.ExecutionEngine
}

// NewHandler creates a new API handler instance
func NewHandler(agentManager *manager.AgentManager, mcpService *services.MCPService, workflowStorage *services.WorkflowStorageService, executionEngine *services.ExecutionEngine) *Handler {
	return &Handler{
		agentManager:    agentManager,
		mcpService:      mcpService,
		workflowStorage: workflowStorage,
		executionEngine: executionEngine,
	}
}

// HealthCheck returns the health status of the service
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "sohoaas-backend",
	})
}

// GetAgents returns all available agents
func (h *Handler) GetAgents(c *gin.Context) {
	agents := h.agentManager.GetAgents()
	
	c.JSON(http.StatusOK, gin.H{
		"agents": agents,
		"count":  len(agents),
	})
}

// GetPersonalCapabilities retrieves user's personal automation capabilities
func (h *Handler) GetPersonalCapabilities(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found in context",
		})
		return
	}
	
	userObj := user.(*types.User)
	
	response, err := h.agentManager.GetPersonalCapabilities(userObj.ID, userObj)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get personal capabilities",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"agent_response": response,
	})
}

// StartWorkflowDiscovery initiates a workflow discovery conversation
func (h *Handler) StartWorkflowDiscovery(c *gin.Context) {
	var request struct {
		Message string `json:"message" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}
	
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found in context",
		})
		return
	}
	
	userObj := user.(*types.User)
	
	// Create conversation history with the initial message
	conversationHistory := []types.ConversationMessage{
		{
			Role:      "user",
			Message:   request.Message,
			Timestamp: time.Now(),
		},
	}
	
	response, err := h.agentManager.ProcessUserMessage(userObj.ID, request.Message, conversationHistory, userObj)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process user message",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"agent_response": response,
		"conversation_id": "conv_" + userObj.ID + "_" + time.Now().Format("20060102150405"),
	})
}

// ContinueWorkflowDiscovery continues an existing workflow discovery conversation
func (h *Handler) ContinueWorkflowDiscovery(c *gin.Context) {
	var request struct {
		Message             string                        `json:"message" binding:"required"`
		ConversationHistory []types.ConversationMessage  `json:"conversation_history"`
		ConversationID      string                        `json:"conversation_id"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}
	
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found in context",
		})
		return
	}
	
	userObj := user.(*types.User)
	
	// Add the new message to conversation history
	request.ConversationHistory = append(request.ConversationHistory, types.ConversationMessage{
		Role:      "user",
		Message:   request.Message,
		Timestamp: time.Now(),
	})
	
	response, err := h.agentManager.ProcessUserMessage(userObj.ID, request.Message, request.ConversationHistory, userObj)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process user message",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"agent_response":    response,
		"conversation_id":   request.ConversationID,
	})
}

// AnalyzeIntent validates and analyzes a workflow intent
func (h *Handler) AnalyzeIntent(c *gin.Context) {
	var request struct {
		WorkflowIntent types.WorkflowIntent `json:"workflow_intent" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid workflow intent format",
		})
		return
	}
	
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found in context",
		})
		return
	}
	
	userObj := user.(*types.User)
	
	response, err := h.agentManager.AnalyzeIntent(userObj.ID, &request.WorkflowIntent, userObj)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to analyze intent",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"agent_response": response,
	})
}

// GenerateWorkflow generates a deterministic workflow from validated intent
func (h *Handler) GenerateWorkflow(c *gin.Context) {
	log.Printf("[API] === WORKFLOW GENERATION REQUEST RECEIVED ===")
	log.Printf("[API] Request Method: %s, URL: %s", c.Request.Method, c.Request.URL.String())
	log.Printf("[API] Request Headers: %+v", c.Request.Header)
	
	var request struct {
		ValidatedIntent map[string]interface{} `json:"validated_intent" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("[API] ERROR: Failed to bind JSON request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid validated intent format",
		})
		return
	}
	
	log.Printf("[API] Parsed validated intent: %+v", request.ValidatedIntent)
	
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found in context",
		})
		return
	}
	
	userObj := user.(*types.User)
	
	log.Printf("[API] Calling AgentManager.GenerateWorkflow for user %s", userObj.ID)
	response, err := h.agentManager.GenerateWorkflow(userObj.ID, request.ValidatedIntent, userObj)
	if err != nil {
		log.Printf("[API] ERROR: GenerateWorkflow failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate workflow",
		})
		return
	}
	
	log.Printf("[API] SUCCESS: Workflow generation completed")
	log.Printf("[API] Response AgentID: %s", response.AgentID)
	log.Printf("[API] Response Error: %s", response.Error)
	if response.Output != nil {
		log.Printf("[API] Response Output keys: %+v", getMapKeys(response.Output))
		if workflowFile, exists := response.Output["workflow_file"]; exists {
			log.Printf("[API] Workflow file saved: %+v", workflowFile)
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"agent_response": response,
	})
}

// ExecuteWorkflow executes a generated workflow
func (h *Handler) ExecuteWorkflow(c *gin.Context) {
	var request struct {
		WorkflowCUE    string                 `json:"workflow_cue" binding:"required"`
		UserParameters map[string]interface{} `json:"user_parameters"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid workflow execution request",
		})
		return
	}
	
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found in context",
		})
		return
	}
	
	token, exists := c.Get("token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Token not found in context",
		})
		return
	}
	
	userObj := user.(*types.User)
	tokenStr := token.(string)
	
	log.Printf("[API] === WORKFLOW EXECUTION STARTED ===")
	log.Printf("[API] User: %s", userObj.ID)
	log.Printf("[API] Workflow CUE length: %d characters", len(request.WorkflowCUE))
	log.Printf("[API] User parameters: %+v", request.UserParameters)
	
	// Create workflow execution
	execution := &types.WorkflowExecution{
		ID:          "exec_" + userObj.ID + "_" + time.Now().Format("20060102150405"),
		UserID:      userObj.ID,
		WorkflowCUE: request.WorkflowCUE,
		Status:      "pending",
		Steps:       []types.WorkflowStep{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	log.Printf("[API] Created execution plan: %s", execution.ID)
	
	// Prepare execution plan using the execution engine
	executionPlan, err := h.executionEngine.PrepareExecution(
		request.WorkflowCUE, 
		userObj.ID, 
		userObj, 
		request.UserParameters, 
		tokenStr,
	)
	if err != nil {
		log.Printf("[API] ERROR: Failed to prepare execution plan: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to prepare workflow execution",
			"details": err.Error(),
		})
		return
	}
	
	log.Printf("[API] Execution plan prepared successfully")
	log.Printf("[API] Workflow: %s (%s)", executionPlan.Name, executionPlan.Description)
	log.Printf("[API] Steps to execute: %d", len(executionPlan.ResolvedSteps))
	
	if len(executionPlan.ValidationErrors) > 0 {
		log.Printf("[API] WARNING: Validation errors found: %v", executionPlan.ValidationErrors)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Workflow validation failed",
			"validation_errors": executionPlan.ValidationErrors,
		})
		return
	}
	
	// Execute the workflow
	log.Printf("[API] Starting workflow execution...")
	execution.Status = "running"
	
	err = h.executionEngine.ExecuteWorkflow(executionPlan)
	if err != nil {
		log.Printf("[API] ERROR: Workflow execution failed: %v", err)
		execution.Status = "failed"
		c.JSON(http.StatusInternalServerError, gin.H{
			"execution_id": execution.ID,
			"status": "failed",
			"error": err.Error(),
			"execution_plan": executionPlan,
		})
		return
	}
	
	execution.Status = "completed"
	log.Printf("[API] === WORKFLOW EXECUTION COMPLETED SUCCESSFULLY ===")
	log.Printf("[API] Execution ID: %s", execution.ID)
	log.Printf("[API] Steps completed: %d", len(executionPlan.ResolvedSteps))
	
	c.JSON(http.StatusOK, gin.H{
		"execution_id": execution.ID,
		"status": "completed",
		"message": "Workflow executed successfully",
		"execution_plan": executionPlan,
		"steps_completed": len(executionPlan.ResolvedSteps),
	})
}

// GetUserServices retrieves user's connected MCP services
func (h *Handler) GetUserServices(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found in context",
		})
		return
	}
	
	token, exists := c.Get("token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Token not found in context",
		})
		return
	}
	
	userObj := user.(*types.User)
	tokenStr := token.(string)
	
	services, err := h.mcpService.GetUserServices(userObj.ID, tokenStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user services",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"services": services,
		"count":    len(services),
	})
}

// GetUserWorkflows retrieves user's saved CUE workflow files
func (h *Handler) GetUserWorkflows(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found in context",
		})
		return
	}

	userObj := user.(*types.User)

	workflows, err := h.workflowStorage.ListUserWorkflows(userObj.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user workflows",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":     len(workflows),
		"workflows": workflows,
	})
}

// GetWorkflow retrieves a specific workflow file by ID
func (h *Handler) GetWorkflow(c *gin.Context) {
	workflowID := c.Param("id")
	if workflowID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Workflow ID is required",
		})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found in context",
		})
		return
	}

	userObj := user.(*types.User)

	workflow, err := h.workflowStorage.GetWorkflow(userObj.ID, workflowID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Workflow not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"workflow": workflow,
	})
}

// TestCompleteWorkflowPipeline tests the complete end-to-end workflow pipeline
func (h *Handler) TestCompleteWorkflowPipeline(c *gin.Context) {
	start := time.Now()
	log.Printf("[API] Starting complete workflow pipeline test")
	
	// Get user from context (set by auth middleware)
	userInterface, exists := c.Get("user")
	if !exists {
		log.Printf("[API] ERROR: No user found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	
	user, ok := userInterface.(*types.User)
	if !ok {
		log.Printf("[API] ERROR: Invalid user type in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user data"})
		return
	}
	
	// Get OAuth token from context (set by auth middleware)
	token, exists := c.Get("token")
	if !exists {
		log.Printf("[API] ERROR: No OAuth token found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "OAuth token not found"})
		return
	}
	
	tokenStr := token.(string)
	
	// Create test workflow intent
	testIntent := &types.WorkflowIntent{
		WorkflowPattern: "Send weekly reports to my team every Friday",
		TriggerConditions: map[string]interface{}{
			"request_type": "automation",
			"frequency":    "weekly",
			"day":         "Friday",
		},
		ActionSequence: []types.WorkflowAction{
			{
				Service: "gmail",
				Action:  "send_email",
				Parameters: map[string]interface{}{
					"subject": "Weekly Report",
					"body":    "Weekly team report",
				},
			},
		},
	}
	
	log.Printf("[API] Testing pipeline for user %s with intent: %s", user.ID, testIntent.WorkflowPattern)
	
	// Phase 1: Test Intent Analysis
	log.Printf("[API] Phase 1: Testing Intent Analysis")
	intentResponse, err := h.agentManager.AnalyzeIntent(user.ID, testIntent, user)
	if err != nil {
		log.Printf("[API] ERROR: Intent analysis failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Intent analysis failed",
			"details": err.Error(),
			"phase": "intent_analysis",
		})
		return
	}
	
	if intentResponse.Error != "" {
		log.Printf("[API] WARNING: Intent analysis returned error: %s", intentResponse.Error)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Intent analysis error",
			"details": intentResponse.Error,
			"phase": "intent_analysis",
		})
		return
	}
	
	log.Printf("[API] Phase 1 SUCCESS: Intent analysis completed")
	
	// Phase 2: Test Workflow Generation
	log.Printf("[API] Phase 2: Testing Workflow Generation")
	workflowResponse, err := h.agentManager.GenerateWorkflow(user.ID, intentResponse.Output, user)
	if err != nil {
		log.Printf("[API] ERROR: Workflow generation failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Workflow generation failed",
			"details": err.Error(),
			"phase": "workflow_generation",
			"intent_analysis": intentResponse.Output,
		})
		return
	}
	
	if workflowResponse.Error != "" {
		log.Printf("[API] WARNING: Workflow generation returned error: %s", workflowResponse.Error)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Workflow generation error",
			"details": workflowResponse.Error,
			"phase": "workflow_generation",
			"intent_analysis": intentResponse.Output,
		})
		return
	}
	
	log.Printf("[API] Phase 2 SUCCESS: Workflow generation completed")
	
	// Phase 3: Test Execution Engine Preparation
	log.Printf("[API] Phase 3: Testing Execution Engine")
	var cueworkflow string
	if workflowCue, ok := workflowResponse.Output["workflow_cue"].(string); ok && workflowCue != "" {
		cueworkflow = workflowCue
	} else {
		log.Printf("[API] ERROR: No workflow_cue generated - workflow generation failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Workflow generation failed - no CUE content produced",
			"details": "The workflow generator must produce valid CUE content using live MCP catalog services",
		})
		return
	}
	
	executionPlan, err := h.executionEngine.PrepareExecution(cueworkflow, user.ID, user, intentResponse.Output, tokenStr)
	if err != nil {
		log.Printf("[API] ERROR: Execution preparation failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Execution preparation failed",
			"details": err.Error(),
			"phase": "execution_preparation",
			"intent_analysis": intentResponse.Output,
			"workflow_generation": workflowResponse.Output,
		})
		return
	}
	
	log.Printf("[API] Phase 3 SUCCESS: Execution preparation completed")
	
	// Complete pipeline test results
	duration := time.Since(start)
	log.Printf("[API] COMPLETE PIPELINE SUCCESS: All phases completed in %v", duration)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Complete workflow pipeline test successful",
		"duration_ms": duration.Milliseconds(),
		"phases": gin.H{
			"intent_analysis": gin.H{
				"status": "success",
				"agent_id": intentResponse.AgentID,
				"output": intentResponse.Output,
			},
			"workflow_generation": gin.H{
				"status": "success",
				"agent_id": workflowResponse.AgentID,
				"output": workflowResponse.Output,
			},
			"execution_preparation": gin.H{
				"status": "success",
				"workflow_id": executionPlan.WorkflowID,
				"steps_count": len(executionPlan.ResolvedSteps),
				"validation_errors": executionPlan.ValidationErrors,
			},
		},
		"user_id": user.ID,
		"test_intent": testIntent.WorkflowPattern,
	})
}

// ValidateServiceCatalog validates the service catalog integrity
func (h *Handler) ValidateServiceCatalog(c *gin.Context) {
	log.Printf("[API] Starting service catalog validation")
	
	serviceCatalog := h.agentManager.GetServiceCatalog()
	serviceSchemas := h.agentManager.GetServiceSchemas()
	
	// Validate catalog structure
	validationResults := gin.H{
		"catalog_valid": true,
		"services_count": len(serviceCatalog.Services),
		"services": gin.H{},
		"validation_errors": []string{},
	}
	
	var validationErrors []string
	
	// Test each service in the catalog
	for serviceName, serviceSchema := range serviceSchemas {
		serviceValidation := gin.H{
			"service_name": serviceName,
			"status": serviceSchema.Status,
			"actions_count": len(serviceSchema.Actions),
			"actions": gin.H{},
		}
		
		// Validate each action
		for actionName, actionSchema := range serviceSchema.Actions {
			actionValidation := gin.H{
				"action_name": actionName,
				"required_fields": len(actionSchema.RequiredFields),
				"optional_fields": len(actionSchema.OptionalFields),
				"description": actionSchema.Description,
			}
			
			// Validate field schemas
			if len(actionSchema.RequiredFields) == 0 {
				validationErrors = append(validationErrors, 
					fmt.Sprintf("Service %s action %s has no required fields", serviceName, actionName))
			}
			
			serviceValidation["actions"].(gin.H)[actionName] = actionValidation
		}
		
		// Validate service has at least one action
		if len(serviceSchema.Actions) == 0 {
			validationErrors = append(validationErrors, 
				fmt.Sprintf("Service %s has no actions defined", serviceName))
		}
		
		validationResults["services"].(gin.H)[serviceName] = serviceValidation
	}
	
	// Validate that MCP catalog contains at least one service (dynamic validation)
	mcpServices, err := h.mcpService.GetServiceCatalog()
	if err != nil {
		validationErrors = append(validationErrors, "Failed to query MCP service catalog")
	} else if len(mcpServices) == 0 {
		validationErrors = append(validationErrors, "No services available in MCP catalog")
	}
	
	validationResults["validation_errors"] = validationErrors
	validationResults["catalog_valid"] = len(validationErrors) == 0
	
	status := http.StatusOK
	if len(validationErrors) > 0 {
		status = http.StatusBadRequest
		log.Printf("[API] Service catalog validation failed with %d errors", len(validationErrors))
	} else {
		log.Printf("[API] Service catalog validation successful")
	}
	
	c.JSON(status, validationResults)
}
