package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"sohoaas-backend/internal/api"
	"sohoaas-backend/internal/config"
	"sohoaas-backend/internal/manager"
	"sohoaas-backend/internal/middleware"
	"sohoaas-backend/internal/services"
	"sohoaas-backend/internal/storage"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize configuration
	cfg := config.New()

	// Initialize workflow storage service with pluggable backend first
	workflowStorage, err := storage.CreateStorageFromEnv()
	if err != nil {
		log.Fatalf("Failed to initialize workflow storage: %v", err)
	}
	log.Printf("Initialized workflow storage: %s", workflowStorage.GetStorageType())

	// Initialize services
	mcpService := services.NewMCPService(cfg.MCP.BaseURL)
	genkitService := services.NewGenkitService(cfg.OpenAI.APIKey, mcpService, workflowStorage)

	// Initialize Firebase Authentication using environment variables
	firebaseAuth, err := services.NewFirebaseAuthService()
	if err != nil {
		log.Fatalf("Failed to initialize Firebase Auth: %v", err)
	}

	// Initialize Agent Manager with all agents
	agentManager := manager.NewAgentManager(genkitService, mcpService)

	// Initialize Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	// Initialize execution engine
	executionEngine := services.NewExecutionEngine(mcpService)

	// Initialize token manager
	tokenManager := services.NewTokenManager()
	tokenManager.StartCleanupRoutine()

	// Initialize API handler
	apiHandler := api.NewHandler(agentManager, mcpService, workflowStorage, executionEngine, tokenManager)
	api.SetupRoutes(router, apiHandler, middleware.FirebaseAuthMiddleware(firebaseAuth))

	// Start server
	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting SOHOAAS backend server on port %s", port)
	log.Printf("Environment: %s", cfg.Environment)
	
	// Log all available endpoints
	logEndpoints(port)
	
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// logEndpoints prints all available API endpoints to the console
func logEndpoints(port string) {
	log.Println("Endpoints:")
	log.Println("  GET  /health")
	log.Println("")
	log.Println("API v1 endpoints:")
	log.Println("Public endpoints:")
	log.Println("  GET  /api/v1/health")
	log.Println("")
	log.Println("Protected endpoints (require authentication):")
	log.Println("Agent management:")
	log.Println("  GET  /api/v1/agents")
	log.Println("")
	log.Println("Personal capabilities:")
	log.Println("  GET  /api/v1/capabilities")
	log.Println("")
	log.Println("Workflow discovery:")
	log.Println("  POST /api/v1/workflow/discover")
	log.Println("  POST /api/v1/workflow/continue")
	log.Println("")
	log.Println("Intent analysis:")
	log.Println("  POST /api/v1/intent/analyze")
	log.Println("")
	log.Println("Workflow generation:")
	log.Println("  POST /api/v1/workflow/generate")
	log.Println("")
	log.Println("Workflow execution:")
	log.Println("  POST /api/v1/workflow/execute")
	log.Println("")
	log.Println("User services:")
	log.Println("  GET  /api/v1/services")
	log.Println("")
	log.Println("Workflow management:")
	log.Println("  GET  /api/v1/workflows")
	log.Println("  GET  /api/v1/workflows/:id")
	log.Println("")
	log.Println("Testing and validation:")
	log.Println("  POST /api/v1/test/pipeline")
	log.Println("  GET  /api/v1/validate/catalog")
	log.Println("")
	log.Printf("Server running at: http://localhost:%s", port)
}
