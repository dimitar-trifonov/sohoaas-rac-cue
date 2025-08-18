package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/dimitar-trifonov/sohoaas/service-proxies/mcp"
	"github.com/dimitar-trifonov/sohoaas/service-proxies/providers/workspace"
	"github.com/dimitar-trifonov/sohoaas/service-proxies/workflow"
)

func main() {
	fmt.Println("Service Proxies - Multi-Provider Workflow Engine")
	fmt.Println("================================================")

	// Create workflow engine
	engine := workflow.NewMultiProviderWorkflowEngine()

	// Load OAuth2 credentials from environment variables
	creds, err := loadGoogleCredentialsFromEnv()
	if err != nil {
		log.Fatalf("Failed to load Google credentials: %v", err)
	}

	// Initialize OAuth2 configuration with loaded credentials
	oauthConfig := &oauth2.Config{
		ClientID:     creds.Web.ClientID,
		ClientSecret: creds.Web.ClientSecret,
		RedirectURL:  getEnvOrDefault("OAUTH_REDIRECT_URL", "http://localhost:8080/api/v1/auth/google/callback"), // Match your Google Cloud Console config
		Scopes: []string{
			"https://www.googleapis.com/auth/gmail.modify",
			"https://www.googleapis.com/auth/documents",
			"https://www.googleapis.com/auth/drive",
			"https://www.googleapis.com/auth/calendar",
		},
		Endpoint: google.Endpoint,
	}

	// Initialize workspace proxies
	gmailProxy := workspace.NewGmailProxy(oauthConfig)
	docsProxy := workspace.NewDocsProxy(oauthConfig)
	driveProxy := workspace.NewDriveProxy(oauthConfig)
	calendarProxy := workspace.NewCalendarProxy(oauthConfig)

	// Register workspace services
	engine.RegisterServiceProxy("workspace", "gmail", gmailProxy)
	engine.RegisterServiceProxy("workspace", "docs", docsProxy)
	engine.RegisterServiceProxy("workspace", "drive", driveProxy)
	engine.RegisterServiceProxy("workspace", "calendar", calendarProxy)

	fmt.Printf("Registered providers: %v\n", engine.GetSupportedProviders())
	fmt.Printf("Workspace services: %v\n", engine.GetSupportedServices("workspace"))

	fmt.Println("\nService proxy backend initialized successfully!")
	fmt.Println("Ready to execute multi-provider workflows.")

	// Create workspace manager for MCP
	workspaceManager := workspace.NewProxyManager(&workspace.ProxyConfig{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("OAUTH_REDIRECT_URL"),
		GmailScopes:  []string{"https://www.googleapis.com/auth/gmail.send"},
		DocsScopes:   []string{"https://www.googleapis.com/auth/documents"},
		DriveScopes:  []string{"https://www.googleapis.com/auth/drive"},
		CalendarScopes: []string{"https://www.googleapis.com/auth/calendar"},
	})

	// Create MCP server
	mcpServer := mcp.NewMCPServer(workspaceManager, engine)

	// Start HTTP server for proxy API endpoints and MCP WebSocket
	startHTTPServer(engine, oauthConfig, gmailProxy, docsProxy, driveProxy, calendarProxy, mcpServer)
}

func startHTTPServer(engine *workflow.MultiProviderWorkflowEngine, oauthConfig *oauth2.Config, gmailProxy *workspace.GmailProxy, docsProxy *workspace.DocsProxy, driveProxy *workspace.DriveProxy, calendarProxy *workspace.CalendarProxy, mcpServer *mcp.MCPServer) {
	r := gin.Default()

	// Store OAuth2 state and token
	var currentToken *oauth2.Token
	oauthStates := make(map[string]bool)
	
	// Get frontend URL for OAuth2 redirects
	frontendURL := getEnvOrDefault("FRONTEND_URL", "http://localhost:3000")

	// OAuth2 authorization endpoint
	r.GET("/api/auth/login", func(c *gin.Context) {
		// Generate random state
		state := generateRandomState()
		oauthStates[state] = true

		// Generate authorization URL
		authURL := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)

		c.JSON(http.StatusOK, gin.H{
			"auth_url": authURL,
			"message":  "Visit this URL to authorize the application",
			"state":    state,
		})
	})

	// OAuth2 callback endpoint (matches Google Cloud Console config)
	r.GET("/api/v1/auth/google/callback", func(c *gin.Context) {
		code := c.Query("code")
		state := c.Query("state")
		error := c.Query("error")

		// Handle OAuth error (user denied access, etc.)
		if error != "" {
			// Redirect back to frontend with error
			c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/?auth_error="+error)
			return
		}

		// Verify state
		if !oauthStates[state] {
			// Redirect back to frontend with error
			c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/?auth_error=invalid_state")
			return
		}
		delete(oauthStates, state)

		// Exchange code for token
		token, err := oauthConfig.Exchange(context.Background(), code)
		if err != nil {
			// Redirect back to frontend with error
			c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/?auth_error=token_exchange_failed")
			return
		}

		// Store token (in production, associate with user)
		currentToken = token

		// Redirect back to frontend with success
		c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/oauth-success.html")
	})

	// OAuth2 callback endpoint for frontend route (alternative path)
	r.GET("/api/auth/callback", func(c *gin.Context) {
		code := c.Query("code")
		state := c.Query("state")
		error := c.Query("error")

		// Handle OAuth error (user denied access, etc.)
		if error != "" {
			// Redirect to OAuth error page for popup communication
			c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/oauth-error.html?error="+error)
			return
		}

		// Verify state
		if !oauthStates[state] {
			// Redirect to OAuth error page for popup communication
			c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/oauth-error.html?error=invalid_state")
			return
		}
		delete(oauthStates, state)

		// Exchange code for token
		token, err := oauthConfig.Exchange(context.Background(), code)
		if err != nil {
			// Redirect back to frontend with error
			c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/?auth_error=token_exchange_failed")
			return
		}

		// Store token (in production, associate with user)
		currentToken = token

		// Redirect back to frontend with success
		c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/oauth-success.html")
	})

	// Get current token endpoint
	r.GET("/api/auth/token", func(c *gin.Context) {
		if currentToken == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No token available. Please authorize first."})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"access_token": currentToken.AccessToken,
			"token_type":   currentToken.TokenType,
			"expires_in":   currentToken.Expiry.Unix(),
			"valid":        currentToken.Valid(),
		})
	})

	// Health endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":   "healthy",
			"providers": engine.GetSupportedProviders(),
		})
	})

	// Workflow execution endpoint
	r.POST("/api/workflow/execute", func(c *gin.Context) {
		var request struct {
			Steps []workflow.WorkflowStep `json:"steps"`
			Input map[string]interface{}  `json:"input"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Debug logging
		fmt.Printf("[DEBUG] Workflow execute request received:\n")
		fmt.Printf("[DEBUG] Steps count: %d\n", len(request.Steps))
		for i, step := range request.Steps {
			fmt.Printf("[DEBUG] Step %d: ID=%s, Provider=%s, Service=%s, Function=%s\n", i, step.ID, step.Provider, step.Service, step.Function)
		}
		fmt.Printf("[DEBUG] Input: %+v\n", request.Input)

		// Use current token if available
		if currentToken != nil {
			if request.Input == nil {
				request.Input = make(map[string]interface{})
			}
			request.Input["oauth_token"] = currentToken.AccessToken
			// Also set the token for the workspace provider
			engine.SetProviderToken("workspace", currentToken.AccessToken)
		}

		result, err := engine.ExecuteWorkflow(context.Background(), request.Steps, request.Input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	})

	// Provider info endpoints
	r.GET("/api/providers", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"providers": engine.GetSupportedProviders(),
		})
	})

	r.GET("/api/providers/:provider/services", func(c *gin.Context) {
		provider := c.Param("provider")
		services := engine.GetSupportedServices(provider)
		c.JSON(http.StatusOK, gin.H{
			"provider": provider,
			"services": services,
		})
	})

	// Service discovery endpoint with metadata
	r.GET("/api/services", func(c *gin.Context) {
		// Build service metadata for all providers
		providersMetadata := make(map[string]map[string]interface{})
		
		// For workspace provider, get metadata from all registered services
		workspaceServices := make(map[string]interface{})
		
		// Get Gmail service metadata
		gmailMetadata := gmailProxy.GetServiceMetadata()
		workspaceServices[gmailMetadata.ServiceType] = map[string]interface{}{
			"display_name": gmailMetadata.DisplayName,
			"description":  gmailMetadata.Description,
			"functions":    gmailMetadata.Functions,
		}
		
		// Get Docs service metadata
		docsMetadata := docsProxy.GetServiceMetadata()
		workspaceServices[docsMetadata.ServiceType] = map[string]interface{}{
			"display_name": docsMetadata.DisplayName,
			"description":  docsMetadata.Description,
			"functions":    docsMetadata.Functions,
		}
		
		// Get Drive service metadata
		driveMetadata := driveProxy.GetServiceMetadata()
		workspaceServices[driveMetadata.ServiceType] = map[string]interface{}{
			"display_name": driveMetadata.DisplayName,
			"description":  driveMetadata.Description,
			"functions":    driveMetadata.Functions,
		}
		
		// Get Calendar service metadata
		calendarMetadata := calendarProxy.GetServiceMetadata()
		workspaceServices[calendarMetadata.ServiceType] = map[string]interface{}{
			"display_name": calendarMetadata.DisplayName,
			"description":  calendarMetadata.Description,
			"functions":    calendarMetadata.Functions,
		}
		
		providersMetadata["workspace"] = map[string]interface{}{
			"display_name": "Google Workspace",
			"description":  "Google Workspace services including Gmail, Docs, Drive, and Calendar",
			"services":     workspaceServices,
		}

		c.JSON(http.StatusOK, gin.H{
			"providers": providersMetadata,
		})
	})

	// MCP WebSocket endpoint
	r.GET("/mcp", func(c *gin.Context) {
		// Check if this is a WebSocket upgrade request
		if c.Request.Header.Get("Upgrade") != "websocket" {
			c.JSON(400, gin.H{"error": "WebSocket upgrade required"})
			return
		}
		mcpServer.HandleWebSocket(c.Writer, c.Request)
	})

	// MCP REST API endpoints (for Genkit MCP plugin compatibility)
	// GET for listing operations (follows REST conventions)
	r.GET("/api/mcp/tools", func(c *gin.Context) {
		// Use the same service discovery as REST API
		var tools []map[string]interface{}
		
		// Get Gmail service metadata and convert to MCP tools
		gmailMetadata := gmailProxy.GetServiceMetadata()
		for functionName, functionInfo := range gmailMetadata.Functions {
			tools = append(tools, map[string]interface{}{
				"name": fmt.Sprintf("gmail.%s", functionName),
				"description": functionInfo.Description,
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"token": map[string]interface{}{
							"type": "string",
							"description": "OAuth2 access token",
						},
					},
					"required": append([]string{"token"}, functionInfo.RequiredFields...),
				},
			})
		}
		
		// Get Docs service metadata and convert to MCP tools
		docsMetadata := docsProxy.GetServiceMetadata()
		for functionName, functionInfo := range docsMetadata.Functions {
			tools = append(tools, map[string]interface{}{
				"name": fmt.Sprintf("docs.%s", functionName),
				"description": functionInfo.Description,
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"token": map[string]interface{}{
							"type": "string",
							"description": "OAuth2 access token",
						},
					},
					"required": append([]string{"token"}, functionInfo.RequiredFields...),
				},
			})
		}
		
		// Get Drive service metadata and convert to MCP tools
		driveMetadata := driveProxy.GetServiceMetadata()
		for functionName, functionInfo := range driveMetadata.Functions {
			tools = append(tools, map[string]interface{}{
				"name": fmt.Sprintf("drive.%s", functionName),
				"description": functionInfo.Description,
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"token": map[string]interface{}{
							"type": "string",
							"description": "OAuth2 access token",
						},
					},
					"required": append([]string{"token"}, functionInfo.RequiredFields...),
				},
			})
		}
		
		// Get Calendar service metadata and convert to MCP tools
		calendarMetadata := calendarProxy.GetServiceMetadata()
		for functionName, functionInfo := range calendarMetadata.Functions {
			tools = append(tools, map[string]interface{}{
				"name": fmt.Sprintf("calendar.%s", functionName),
				"description": functionInfo.Description,
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"token": map[string]interface{}{
							"type": "string",
							"description": "OAuth2 access token",
						},
					},
					"required": append([]string{"token"}, functionInfo.RequiredFields...),
				},
			})
		}
		
		c.JSON(http.StatusOK, gin.H{
			"tools": tools,
		})
	})

	// POST for tool execution (follows REST conventions)
	r.POST("/api/mcp/tools/call", func(c *gin.Context) {
		var request struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request format",
				"details": err.Error(),
			})
			return
		}

		result, err := mcpServer.ExecuteTool(request.Name, request.Arguments)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Tool execution failed",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"result": result,
		})
	})

	// GET for listing resources (follows REST conventions)
	r.GET("/api/mcp/resources", func(c *gin.Context) {
		resources := mcpServer.GetAvailableResources()
		c.JSON(http.StatusOK, gin.H{
			"resources": resources,
		})
	})

	// GET for reading specific resource with URI as query parameter
	r.GET("/api/mcp/resources/read", func(c *gin.Context) {
		uri := c.Query("uri")
		if uri == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "URI parameter is required",
			})
			return
		}

		content, err := mcpServer.ReadResource(uri)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Resource read failed",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"content": content,
		})
	})

	port := getEnvOrDefault("PORT", "8080")
	fmt.Printf("\nStarting HTTP server on :%s...\n", port)
	fmt.Println("Endpoints:")
	fmt.Println("  GET  /health")
	fmt.Println("  POST /workflow/execute")
	fmt.Println("  GET  /providers")
	fmt.Println("  GET  /providers/:provider/services")
	fmt.Println("  GET  /api/services")
	fmt.Println("  GET  /mcp (WebSocket - MCP Protocol)")
	fmt.Println("MCP REST API endpoints:")
	fmt.Println("  GET  /api/mcp/tools")
	fmt.Println("  POST /api/mcp/tools/call")
	fmt.Println("  GET  /api/mcp/resources")
	fmt.Println("  GET  /api/mcp/resources/read?uri=<resource_uri>")
	log.Printf("Server starting on :%s", port)
	log.Println("OAuth2 endpoints:")
	log.Println("  GET /auth/login   - Get authorization URL")
	log.Println("  GET /auth/callback - OAuth callback (automatic)")
	log.Println("  GET /auth/token   - Get current token")
	log.Fatal(r.Run(":" + port))
}

// GoogleCredentials represents the structure of Google OAuth2 credentials JSON file
type GoogleCredentials struct {
	Web struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	} `json:"web"`
}

// loadGoogleCredentialsFromEnv loads OAuth2 credentials from environment variables
func loadGoogleCredentialsFromEnv() (*GoogleCredentials, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

	if clientID == "" {
		return nil, fmt.Errorf("GOOGLE_CLIENT_ID environment variable is required")
	}
	if clientSecret == "" {
		return nil, fmt.Errorf("GOOGLE_CLIENT_SECRET environment variable is required")
	}

	creds := &GoogleCredentials{
		Web: struct {
			ClientID     string `json:"client_id"`
			ClientSecret string `json:"client_secret"`
		}{
			ClientID:     clientID,
			ClientSecret: clientSecret,
		},
	}

	return creds, nil
}

// generateRandomState generates a random state string for OAuth2
func generateRandomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
