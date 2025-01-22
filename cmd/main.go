package main

import (
	"log"

	fshandler "github.com/gomcpgo/filesys/pkg/handler"
	"github.com/gomcpgo/mcp/pkg/handler"
	"github.com/gomcpgo/mcp/pkg/server"
)

func main() {
	// Create the filesystem handler
	fsHandler := fshandler.NewFileSystemHandler()

	// Create handler registry
	registry := handler.NewHandlerRegistry()
	registry.RegisterToolHandler(fsHandler)

	// Create and start server
	srv := server.New(server.Options{
		Name:     "filesystem-server",
		Version:  "1.0.0",
		Registry: registry,
	})

	log.Printf("Starting filesystem server")
	if err := srv.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}