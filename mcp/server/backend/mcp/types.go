package mcp

import (
	"encoding/json"
)

// MCP Protocol Types
// Based on Model Context Protocol specification

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// MCP Capability and Resource Types

// ServerCapabilities defines what the MCP server can do
type ServerCapabilities struct {
	Resources *ResourceCapability `json:"resources,omitempty"`
	Tools     *ToolCapability     `json:"tools,omitempty"`
	Prompts   *PromptCapability   `json:"prompts,omitempty"`
}

// ResourceCapability defines resource handling capabilities
type ResourceCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// ToolCapability defines tool handling capabilities
type ToolCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// PromptCapability defines prompt handling capabilities
type PromptCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// InitializeRequest represents the MCP initialize request
type InitializeRequest struct {
	ProtocolVersion string              `json:"protocolVersion"`
	Capabilities    ClientCapabilities  `json:"capabilities"`
	ClientInfo      ClientInfo          `json:"clientInfo"`
}

// ClientCapabilities defines what the client can handle
type ClientCapabilities struct {
	Roots    *RootCapability    `json:"roots,omitempty"`
	Sampling *SamplingCapability `json:"sampling,omitempty"`
}

// RootCapability defines root handling capabilities
type RootCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// SamplingCapability defines sampling capabilities
type SamplingCapability struct{}

// ClientInfo contains information about the MCP client
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializeResult represents the response to initialize
type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
}

// ServerInfo contains information about the MCP server
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Resource Types

// Resource represents an MCP resource
type Resource struct {
	URI         string      `json:"uri"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	MimeType    string      `json:"mimeType,omitempty"`
	Annotations interface{} `json:"annotations,omitempty"`
}

// ResourceContent represents the content of a resource
type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text,omitempty"`
	Blob     string `json:"blob,omitempty"`
}

// ListResourcesResult represents the result of listing resources
type ListResourcesResult struct {
	Resources []Resource `json:"resources"`
}

// ReadResourceRequest represents a request to read a resource
type ReadResourceRequest struct {
	URI string `json:"uri"`
}

// ReadResourceResult represents the result of reading a resource
type ReadResourceResult struct {
	Contents []ResourceContent `json:"contents"`
}

// Tool Types

// Tool represents an MCP tool
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	InputSchema interface{} `json:"inputSchema"`
}

// ListToolsResult represents the result of listing tools
type ListToolsResult struct {
	Tools []Tool `json:"tools"`
}

// CallToolRequest represents a request to call a tool
type CallToolRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// ToolResult represents the result of calling a tool
type ToolResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// ToolContent represents content returned by a tool
type ToolContent struct {
	Type     string      `json:"type"`
	Text     string      `json:"text,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	MimeType string      `json:"mimeType,omitempty"`
}

// Workspace-specific types for our implementation

// WorkspaceResource represents a Google Workspace resource
type WorkspaceResource struct {
	Resource
	ServiceType string                 `json:"serviceType"` // gmail, docs, drive, calendar
	Function    string                 `json:"function"`    // operation name
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// WorkspaceTool represents a Google Workspace operation as an MCP tool
type WorkspaceTool struct {
	Tool
	ServiceType    string   `json:"serviceType"`
	Function       string   `json:"function"`
	RequiredFields []string `json:"requiredFields"`
	ExamplePayload map[string]interface{} `json:"examplePayload"`
}
