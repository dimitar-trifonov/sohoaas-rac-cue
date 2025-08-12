package workspace

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ProxyHandler handles HTTP requests for the proxy system
type ProxyHandler struct {
	proxyManager *ProxyManager
}

// NewProxyHandler creates a new proxy handler
func NewProxyHandler(proxyManager *ProxyManager) *ProxyHandler {
	return &ProxyHandler{
		proxyManager: proxyManager,
	}
}

// SetupRoutes sets up HTTP routes for the proxy system
func (h *ProxyHandler) SetupRoutes(router *gin.Engine) {
	proxyGroup := router.Group("/api/proxy")
	{
		// Core proxy endpoints
		proxyGroup.POST("/execute", h.handleExecute)
		proxyGroup.POST("/batch", h.handleBatchExecute)

		// Information endpoints
		proxyGroup.GET("/services", h.handleGetServices)
		proxyGroup.GET("/services/:service/functions", h.handleGetFunctions)
		proxyGroup.GET("/capabilities", h.handleGetCapabilities)

		// Validation endpoints
		proxyGroup.POST("/validate", h.handleValidateRequest)
	}
}

// handleExecute handles single proxy execution requests
func (h *ProxyHandler) handleExecute(c *gin.Context) {
	var request ProxyRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": map[string]interface{}{
				"code":    ErrorCodeInvalidPayload,
				"message": "Invalid request format",
				"details": err.Error(),
			},
		})
		return
	}

	// Validate request
	if err := h.proxyManager.ValidateRequest(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": map[string]interface{}{
				"code":    ErrorCodeInvalidPayload,
				"message": "Request validation failed",
				"details": err.Error(),
			},
		})
		return
	}

	// Execute request
	response, err := h.proxyManager.Execute(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": map[string]interface{}{
				"code":    ErrorCodeInternalError,
				"message": "Execution failed",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// handleBatchExecute handles batch proxy execution requests
func (h *ProxyHandler) handleBatchExecute(c *gin.Context) {
	var requests []*ProxyRequest
	if err := c.ShouldBindJSON(&requests); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": map[string]interface{}{
				"code":    ErrorCodeInvalidPayload,
				"message": "Invalid batch request format",
				"details": err.Error(),
			},
		})
		return
	}

	// Validate all requests
	for i, request := range requests {
		if err := h.proxyManager.ValidateRequest(request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": map[string]interface{}{
					"code":          ErrorCodeInvalidPayload,
					"message":       "Request validation failed",
					"details":       err.Error(),
					"request_index": i,
				},
			})
			return
		}
	}

	// Execute batch
	responses, err := h.proxyManager.BatchExecute(c.Request.Context(), requests)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": map[string]interface{}{
				"code":    ErrorCodeInternalError,
				"message": "Batch execution failed",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"responses": responses,
		"total":     len(responses),
	})
}



// handleGetServices returns list of supported services
func (h *ProxyHandler) handleGetServices(c *gin.Context) {
	services := h.proxyManager.GetSupportedServices()
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"services": services,
		"total":    len(services),
	})
}

// handleGetFunctions returns supported functions for a service
func (h *ProxyHandler) handleGetFunctions(c *gin.Context) {
	serviceType := c.Param("service")

	functions, err := h.proxyManager.GetSupportedFunctions(serviceType)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error": map[string]interface{}{
				"code":    ErrorCodeNotFound,
				"message": "Service not found",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"service_type": serviceType,
		"functions":    functions,
		"total":        len(functions),
	})
}

// handleGetCapabilities returns capabilities for all services
func (h *ProxyHandler) handleGetCapabilities(c *gin.Context) {
	capabilities := h.proxyManager.GetServiceCapabilities()
	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"capabilities": capabilities,
	})
}

// handleValidateRequest validates a proxy request without executing it
func (h *ProxyHandler) handleValidateRequest(c *gin.Context) {
	var request ProxyRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"valid":   false,
			"error": map[string]interface{}{
				"code":    ErrorCodeInvalidPayload,
				"message": "Invalid request format",
				"details": err.Error(),
			},
		})
		return
	}

	// Validate request
	if err := h.proxyManager.ValidateRequest(&request); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"valid":   false,
			"error": map[string]interface{}{
				"code":    ErrorCodeInvalidPayload,
				"message": "Request validation failed",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"valid":   true,
		"message": "Request is valid",
	})
}
