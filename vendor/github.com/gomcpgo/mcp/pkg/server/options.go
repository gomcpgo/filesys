package server

import (
	"github.com/gomcpgo/mcp/pkg/handler"
	"github.com/gomcpgo/mcp/pkg/transport"
)

// Options configures the MCP server
type Options struct {
	Name      string
	Version   string
	Registry  *handler.HandlerRegistry
	Transport transport.Transport
}

// Option is a function that can be used to configure the server
type Option func(*Options)

// WithName sets the server name
func WithName(name string) Option {
	return func(o *Options) {
		o.Name = name
	}
}

// WithVersion sets the server version
func WithVersion(version string) Option {
	return func(o *Options) {
		o.Version = version
	}
}

// WithRegistry sets the handler registry
func WithRegistry(registry *handler.HandlerRegistry) Option {
	return func(o *Options) {
		o.Registry = registry
	}
}

// WithTransport sets the transport
func WithTransport(transport transport.Transport) Option {
	return func(o *Options) {
		o.Transport = transport
	}
}

// DefaultOptions returns the default server options
func DefaultOptions() Options {
	return Options{
		Name:      "mcp-server",
		Version:   "1.0.0",
		Transport: transport.NewStdioTransport(),
		Registry:  handler.NewHandlerRegistry(),
	}
}
