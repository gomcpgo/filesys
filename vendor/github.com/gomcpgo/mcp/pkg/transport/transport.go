package transport

import (
	"context"

	"github.com/gomcpgo/mcp/pkg/protocol"
)

// Transport defines the interface for MCP server transport
type Transport interface {
	// Start initializes and starts the transport
	Start(ctx context.Context) error

	// Stop gracefully shuts down the transport
	Stop(ctx context.Context) error

	// Send sends a response or notification
	Send(response *protocol.Response) error

	// Receive returns a channel that provides incoming requests
	Receive() <-chan *protocol.Request

	// Errors returns a channel that provides transport errors
	Errors() <-chan error
}

// Options holds configuration for transports
type Options struct {
	// Add common transport options here
}

// TransportType identifies different transport implementations
type TransportType string

const (
	TypeStdio TransportType = "stdio"
	TypeSSE   TransportType = "sse"
)
