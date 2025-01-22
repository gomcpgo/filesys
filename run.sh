#!/bin/bash


## export PERPLEXITY_API_KEY="your-api-key-here"

# Show usage if no command provided
function show_usage() {
    echo "Usage: ./run.sh [command]"
    echo "Commands:"
    echo "  build    Build the  MCP server binary"
    echo "  run      Run the  MCP server"
    exit 1
}

# Handle different commands
case "$1" in
  build)
    echo "Building  MCP server..."
    go build -o bin/filesystem-server ./cmd
    ;;
  run)
    echo "Running  MCP server..."
    go run ./cmd/main.go
    ;;
  *)
    show_usage
    ;;
esac