#!/bin/bash


## export PERPLEXITY_API_KEY="your-api-key-here"

# Show usage if no command provided
function show_usage() {
    echo "Usage: ./run.sh [command]"
    echo "Commands:"
    echo "  build    Build the MCP server binary"
    echo "  run      Run the MCP server"
    echo "  test     Run all unit tests (no cache)"
    exit 1
}

# Handle different commands
case "$1" in
  build)
    echo "Building MCP server..."
    go build -o bin/filesystem-server ./cmd
    ;;
  run)
    echo "Running MCP server..."
    go run ./cmd/main.go
    ;;
  test)
    echo "Running all unit tests (no cache)..."
    go test -count=1 -v ./...
    ;;
  *)
    show_usage
    ;;
esac