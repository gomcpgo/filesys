package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gomcpgo/mcp/pkg/handler"
	"github.com/gomcpgo/mcp/pkg/protocol"
	"github.com/gomcpgo/mcp/pkg/transport"
)

// Server represents an MCP server instance
type Server struct {
	options   Options
	registry  *handler.HandlerRegistry
	transport transport.Transport
}

// New creates a new MCP server instance with the provided options
func New(options Options) *Server {
	// Start with default options
	defaultOpts := DefaultOptions()

	// Override with provided options
	if options.Name != "" {
		defaultOpts.Name = options.Name
	}
	if options.Version != "" {
		defaultOpts.Version = options.Version
	}
	if options.Registry != nil {
		defaultOpts.Registry = options.Registry
	}
	if options.Transport != nil {
		defaultOpts.Transport = options.Transport
	}

	return &Server{
		options:   defaultOpts,
		registry:  defaultOpts.Registry,
		transport: defaultOpts.Transport,
	}
}

// Run starts the server and handles requests
func (s *Server) Run() error {
	ctx := context.Background()

	// Start transport
	if err := s.transport.Start(ctx); err != nil {
		return fmt.Errorf("failed to start transport: %w", err)
	}
	defer s.transport.Stop(ctx)

	// Process requests
	log.Printf("[SERVER] Starting request processing loop")
	requestCount := 0
	for {
		select {
		case err := <-s.transport.Errors():
			log.Printf("[SERVER] Transport error: %v", err)
			continue

		case req := <-s.transport.Receive():
			if req == nil {
				log.Printf("[SERVER] Received nil request after %d requests, shutting down", requestCount)
				return nil
			}
			requestCount++
			log.Printf("[SERVER] Received request #%d: method=%s", requestCount, req.Method)

			go s.handleRequest(ctx, req)
		}
	}
}

// handleRequest processes individual requests
func (s *Server) handleRequest(ctx context.Context, req *protocol.Request) {
	var result interface{}
	var err error

	log.Printf("[SERVER] Processing request: method=%s, id=%v", req.Method, req.ID)
	startTime := time.Now()
	defer func() {
		log.Printf("[SERVER] Request completed: method=%s, id=%v, duration=%v", req.Method, req.ID, time.Since(startTime))
	}()

	log.Printf("MCP server req received:\n%v\n", PrettyJSON(req))

	switch req.Method {
	case protocol.MethodInitialize:
		result, err = s.handleInitialize(ctx, req.Params)

	case protocol.MethodInitialized, protocol.NotificationInitialized:
		log.Printf("Server initialized successfully")
		// Initialization notification, no response needed
		return

	case protocol.MethodToolsList:
		if s.registry.HasToolHandler() {
			result, err = s.registry.GetToolHandler().ListTools(ctx)
		} else {
			// Return empty list instead of error
			result = &protocol.ListToolsResponse{
				Tools: []protocol.Tool{},
			}
		}

	case protocol.MethodToolsCall:
		if s.registry.HasToolHandler() {
			var toolReq protocol.CallToolRequest
			if err := json.Unmarshal(req.Params, &toolReq); err != nil {
				s.sendError(req.ID, protocol.InvalidParams, "Invalid tool parameters")
				return
			}
			result, err = s.registry.GetToolHandler().CallTool(ctx, &toolReq)
		} else {
			err = fmt.Errorf("tools not supported")
		}

	case protocol.MethodResourcesList:
		if s.registry.HasResourceHandler() {
			result, err = s.registry.GetResourceHandler().ListResources(ctx)
		} else {
			// Return empty list instead of error
			result = &protocol.ListResourcesResponse{
				Resources: []protocol.Resource{},
			}
		}

	case protocol.MethodResourcesRead:
		if s.registry.HasResourceHandler() {
			var resourceReq protocol.ReadResourceRequest
			if err := json.Unmarshal(req.Params, &resourceReq); err != nil {
				s.sendError(req.ID, protocol.InvalidParams, "Invalid resource parameters")
				return
			}
			result, err = s.registry.GetResourceHandler().ReadResource(ctx, &resourceReq)
		} else {
			err = fmt.Errorf("resources not supported")
		}

	case protocol.MethodPromptsList:
		if s.registry.HasPromptHandler() {
			result, err = s.registry.GetPromptHandler().ListPrompts(ctx)
		} else {
			// Return empty list instead of error
			result = &protocol.ListPromptsResponse{
				Prompts: []protocol.Prompt{},
			}
		}

	case protocol.MethodPromptsGet:
		if s.registry.HasPromptHandler() {
			var promptReq protocol.GetPromptRequest
			if err := json.Unmarshal(req.Params, &promptReq); err != nil {
				s.sendError(req.ID, protocol.InvalidParams, "Invalid prompt parameters")
				return
			}
			result, err = s.registry.GetPromptHandler().GetPrompt(ctx, &promptReq)
		} else {
			err = fmt.Errorf("prompts not supported")
		}

	default:
		err = fmt.Errorf("unknown method: %s", req.Method)
	}

	if err != nil {
		s.sendError(req.ID, protocol.InternalError, err.Error())
		return
	}

	s.sendResponse(req.ID, result)
}

// handleInitialize processes initialization requests
func (s *Server) handleInitialize(ctx context.Context, params json.RawMessage) (*protocol.InitializeResponse, error) {
	var initReq protocol.InitializeRequest
	if err := json.Unmarshal(params, &initReq); err != nil {
		return nil, fmt.Errorf("invalid initialization parameters: %w", err)
	}

	capabilities := protocol.Capabilities{}
	if s.registry.HasToolHandler() {
		capabilities.Tools = &protocol.ToolsInfo{}
	}
	if s.registry.HasResourceHandler() {
		capabilities.Resources = &protocol.ResourcesInfo{}
	}
	if s.registry.HasPromptHandler() {
		capabilities.Prompts = &protocol.PromptsInfo{}
	}

	return &protocol.InitializeResponse{
		ProtocolVersion: protocol.Version,
		ServerInfo: protocol.ServerInfo{
			Name:    s.options.Name,
			Version: s.options.Version,
		},
		Capabilities: capabilities,
	}, nil
}

// sendResponse sends a successful response
func (s *Server) sendResponse(id interface{}, result interface{}) {
	response := &protocol.Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	log.Printf("MCP server response:\n%v\n", PrettyJSON(response))
	if err := s.transport.Send(response); err != nil {
		log.Printf("Error sending response: %v", err)
	}
}

// sendError sends an error response
func (s *Server) sendError(id interface{}, code int, message string) {
	response := &protocol.Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &protocol.Error{
			Code:    code,
			Message: message,
		},
	}

	log.Printf("MCP server error response:\n%v\n", PrettyJSON(response))
	if err := s.transport.Send(response); err != nil {
		log.Printf("Error sending error response: %v", err)
	}
}
