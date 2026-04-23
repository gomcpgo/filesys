package main

import (
	_ "embed"
	"log"

	fshandler "github.com/gomcpgo/filesys/pkg/handler"
	"github.com/gomcpgo/mcp/pkg/handler"
	"github.com/gomcpgo/mcp/pkg/protocol"
	"github.com/gomcpgo/mcp/pkg/server"
)

//go:embed icon.svg
var iconSVG []byte

func main() {
	// Create the filesystem handler
	fsHandler := fshandler.NewFileSystemHandler()

	// Create handler registry
	registry := handler.NewHandlerRegistry()
	registry.RegisterToolHandler(fsHandler)

	// Create and start server
	srv := server.New(server.Options{
		Name:     "filesystem-server",
		Title:    "Filesystem",
		Version:  "1.0.0",
		Icons:    protocol.IconFromSVG(iconSVG),
		Registry: registry,
	})

	log.Printf("Starting filesystem server")
	if err := srv.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}