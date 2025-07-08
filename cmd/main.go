package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	fshandler "github.com/gomcpgo/filesys/pkg/handler"
	"github.com/gomcpgo/mcp/pkg/handler"
	"github.com/gomcpgo/mcp/pkg/server"
)

func main() {
	// Set up logging with timestamps
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	
	log.Printf("[FILESYS] Starting filesystem server - PID: %d, PPID: %d", os.Getpid(), os.Getppid())
	
	// Set up signal handling to log when we receive signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGPIPE, syscall.SIGHUP)
	go func() {
		for sig := range sigChan {
			log.Printf("[FILESYS] Received signal: %v", sig)
		}
	}()
	
	// Monitor parent process
	go func() {
		ppid := os.Getppid()
		for {
			time.Sleep(5 * time.Second)
			currentPpid := os.Getppid()
			if currentPpid != ppid {
				log.Printf("[FILESYS] Parent process changed from %d to %d", ppid, currentPpid)
				ppid = currentPpid
			}
			// Check if stdin is still valid
			if _, err := os.Stdin.Stat(); err != nil {
				log.Printf("[FILESYS] stdin error: %v", err)
			}
		}
	}()
	
	// Create the filesystem handler
	fsHandler := fshandler.NewFileSystemHandler()
	log.Printf("[FILESYS] Created filesystem handler")

	// Create handler registry
	registry := handler.NewHandlerRegistry()
	registry.RegisterToolHandler(fsHandler)
	log.Printf("[FILESYS] Registered tool handler")

	// Create and start server
	srv := server.New(server.Options{
		Name:     "filesystem-server",
		Version:  "1.0.0",
		Registry: registry,
	})
	log.Printf("[FILESYS] Created MCP server instance")

	// Add a defer to log when main exits
	defer func() {
		log.Printf("[FILESYS] Main function exiting")
		if r := recover(); r != nil {
			log.Printf("[FILESYS] Panic in main: %v", r)
		}
	}()
	
	log.Printf("[FILESYS] Starting server.Run() at %s", time.Now().Format(time.RFC3339))
	if err := srv.Run(); err != nil {
		log.Printf("[FILESYS] Server.Run() returned error: %v", err)
		log.Fatalf("[FILESYS] Server error: %v", err)
	}
	log.Printf("[FILESYS] Server.Run() completed normally - this means stdin was closed or EOF received")
}