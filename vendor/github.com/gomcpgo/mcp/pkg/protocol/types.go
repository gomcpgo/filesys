package protocol

import "encoding/json"

// JSON-RPC 2.0 message types
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// MCP Protocol types
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Capabilities struct {
	Tools     *ToolsInfo     `json:"tools,omitempty"`
	Resources *ResourcesInfo `json:"resources,omitempty"`
	Prompts   *PromptsInfo   `json:"prompts,omitempty"`
}

type ToolsInfo struct {
	// Tool-specific capabilities
}

type ResourcesInfo struct {
	// Resource-specific capabilities
}

type PromptsInfo struct {
	// Prompt-specific capabilities
}

// Initialize types
type InitializeRequest struct {
	// Initialization parameters (can be extended)
}

type InitializeResponse struct {
	ProtocolVersion string       `json:"protocolVersion"`
	ServerInfo      ServerInfo   `json:"serverInfo"`
	Capabilities    Capabilities `json:"capabilities"`
}

// Tool types
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

type ListToolsResponse struct {
	Tools []Tool `json:"tools"`
}

type CallToolRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type CallToolResponse struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type ToolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Resource types
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

type ListResourcesResponse struct {
	Resources []Resource `json:"resources"`
}

type ReadResourceRequest struct {
	URI string `json:"uri"`
}

type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
	Blob     string `json:"blob,omitempty"`
}

type ReadResourceResponse struct {
	Contents []ResourceContent `json:"contents"`
}

// Prompt types
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

type Prompt struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
}

type ListPromptsResponse struct {
	Prompts []Prompt `json:"prompts"`
}

type GetPromptRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type GetPromptResponse struct {
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string         `json:"role"`
	Content MessageContent `json:"content"`
}

type MessageContent struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	Resource *Resource `json:"resource,omitempty"`
}

// Constants
const (
	Version                 = "2024-11-05"
	MethodInitialize        = "initialize"
	NotificationInitialized = "notifications/initialized"
	MethodInitialized       = "initialized"
	MethodToolsList         = "tools/list"
	MethodToolsCall         = "tools/call"
	MethodResourcesList     = "resources/list"
	MethodResourcesRead     = "resources/read"
	MethodPromptsList       = "prompts/list"
	MethodPromptsGet        = "prompts/get"
)

// Error codes as per JSON-RPC 2.0
const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)
