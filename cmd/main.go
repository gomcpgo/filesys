package main

import (
	"log"
	"os"

	fshandler "github.com/gomcpgo/filesys/pkg/handler"
	"github.com/gomcpgo/mcp/pkg/handler"
	"github.com/gomcpgo/mcp/pkg/server"
)

func main() {
	// Get allowed directories from command-line arguments
	if len(os.Args) < 2 {
		log.Fatal("Usage: filesys <dir1> [dir2] [dir3] ...")
	}

	// Create the filesystem handler with allowed directories
	fsHandler := fshandler.NewFileSystemHandler(os.Args[1:])

	// Create handler registry
	registry := handler.NewHandlerRegistry()
	registry.RegisterToolHandler(fsHandler)

	// Create and start server
	srv := server.New(server.Options{
		Name:     "filesystem-server",
		Version:  "1.0.0",
		Registry: registry,
	})

	log.Printf("Starting filesystem server with allowed directories: %v", os.Args[1:])
	if err := srv.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}