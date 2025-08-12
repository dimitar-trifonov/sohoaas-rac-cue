package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"sohoaas-backend/internal/services"
	"sohoaas-backend/internal/types"
)

// AuthMiddleware validates OAuth2 tokens with MCP service
func AuthMiddleware(mcpService *services.MCPService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Check if it's a Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		// Extract the token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token is required",
			})
			c.Abort()
			return
		}

		// TODO: Validate token with MCP service (to be implemented)
		// For PoC: Create a mock user - services will be populated dynamically from MCP catalog
		user := &types.User{
			ID:    "mock_user_123",
			Email: "user@example.com",
			Name:  "Mock User",
			OAuthTokens: map[string]interface{}{
				"google": map[string]interface{}{
					"access_token": token,
				},
			},
			ConnectedServices: []string{}, // Will be populated dynamically from MCP catalog
		}

		// Store user and token in context for use by handlers
		c.Set("user", user)
		c.Set("token", token)
		c.Next()
	}
}

// CORS middleware for cross-origin requests
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
