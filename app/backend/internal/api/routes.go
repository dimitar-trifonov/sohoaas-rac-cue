package api

import (
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes for the SOHOAAS backend
func SetupRoutes(router *gin.Engine, handler *Handler, authMiddleware gin.HandlerFunc) {
	// Health check endpoint (no auth required)
	router.GET("/health", handler.HealthCheck)
	
	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public routes (no auth required)
		public := v1.Group("/")
		{
			public.GET("/health", handler.HealthCheck)
		}
		
		// Protected routes (auth required)
		protected := v1.Group("/")
		protected.Use(authMiddleware)
		{
			// Token management endpoints
			protected.POST("/auth/store-google-token", handler.StoreGoogleToken)
			protected.GET("/auth/token-info", handler.GetTokenInfo)
			
			// Agent management
			protected.GET("/agents", handler.GetAgents)
			
			// Personal capabilities
			protected.GET("/capabilities", handler.GetPersonalCapabilities)
			
			// Workflow discovery
			protected.POST("/workflow/discover", handler.StartWorkflowDiscovery)
			protected.POST("/workflow/continue", handler.ContinueWorkflowDiscovery)
			
			// Intent analysis
			protected.POST("/intent/analyze", handler.AnalyzeIntent)
			
			// Workflow generation
			protected.POST("/workflow/generate", handler.GenerateWorkflow)
			
			// Workflow execution
			protected.POST("/workflow/execute", handler.ExecuteWorkflow)
			
			// Workflow management
			protected.GET("/workflows", handler.GetUserWorkflows)
			protected.GET("/workflows/:id", handler.GetWorkflow)
			
			// User services
			protected.GET("/services", handler.GetUserServices)
			
			// Testing and validation
			protected.POST("/test/pipeline", handler.TestCompleteWorkflowPipeline)
			protected.GET("/validate/catalog", handler.ValidateServiceCatalog)
		}
	}
}
