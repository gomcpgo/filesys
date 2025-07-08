package handler

import (
	"context"
	"github.com/gomcpgo/mcp/pkg/protocol"
)

// Handler is the base interface that all MCP handlers must implement
type Handler interface {
	// Initialize handles server initialization
	Initialize(ctx context.Context, req *protocol.InitializeRequest) (*protocol.InitializeResponse, error)

	// HandleRequest processes incoming requests
	HandleRequest(ctx context.Context, method string, params []byte) (interface{}, error)
}

// ToolHandler handles tool-related operations
type ToolHandler interface {
	// ListTools returns available tools
	ListTools(ctx context.Context) (*protocol.ListToolsResponse, error)

	// CallTool executes a tool
	CallTool(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResponse, error)
}

// ResourceHandler handles resource-related operations
type ResourceHandler interface {
	// ListResources returns available resources
	ListResources(ctx context.Context) (*protocol.ListResourcesResponse, error)

	// ReadResource reads a specific resource
	ReadResource(ctx context.Context, req *protocol.ReadResourceRequest) (*protocol.ReadResourceResponse, error)
}

// PromptHandler handles prompt-related operations
type PromptHandler interface {
	// ListPrompts returns available prompts
	ListPrompts(ctx context.Context) (*protocol.ListPromptsResponse, error)

	// GetPrompt retrieves a specific prompt
	GetPrompt(ctx context.Context, req *protocol.GetPromptRequest) (*protocol.GetPromptResponse, error)
}

// HandlerRegistry maintains a collection of handlers for different capabilities
type HandlerRegistry struct {
	toolHandler     ToolHandler
	resourceHandler ResourceHandler
	promptHandler   PromptHandler
}

// NewHandlerRegistry creates a new handler registry
func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{}
}

// RegisterToolHandler registers a tool handler
func (r *HandlerRegistry) RegisterToolHandler(h ToolHandler) {
	r.toolHandler = h
}

// RegisterResourceHandler registers a resource handler
func (r *HandlerRegistry) RegisterResourceHandler(h ResourceHandler) {
	r.resourceHandler = h
}

// RegisterPromptHandler registers a prompt handler
func (r *HandlerRegistry) RegisterPromptHandler(h PromptHandler) {
	r.promptHandler = h
}

// GetToolHandler returns the registered tool handler
func (r *HandlerRegistry) GetToolHandler() ToolHandler {
	return r.toolHandler
}

// GetResourceHandler returns the registered resource handler
func (r *HandlerRegistry) GetResourceHandler() ResourceHandler {
	return r.resourceHandler
}

// GetPromptHandler returns the registered prompt handler
func (r *HandlerRegistry) GetPromptHandler() PromptHandler {
	return r.promptHandler
}

// HasToolHandler checks if a tool handler is registered
func (r *HandlerRegistry) HasToolHandler() bool {
	return r.toolHandler != nil
}

// HasResourceHandler checks if a resource handler is registered
func (r *HandlerRegistry) HasResourceHandler() bool {
	return r.resourceHandler != nil
}

// HasPromptHandler checks if a prompt handler is registered
func (r *HandlerRegistry) HasPromptHandler() bool {
	return r.promptHandler != nil
}