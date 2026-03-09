package handler

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gomcpgo/mcp/pkg/protocol"
)

func (h *FileSystemHandler) handleWriteFile(args map[string]interface{}) (*protocol.CallToolResponse, error) {
	path, ok := args["path"].(string)
	if !ok {
		log.Printf("ERROR: write_file - invalid path type: %T", args["path"])
		return nil, fmt.Errorf("path must be a string")
	}
	content, ok := args["content"].(string)
	if !ok {
		log.Printf("ERROR: write_file - invalid content type: %T", args["content"])
		return nil, fmt.Errorf("content must be a string")
	}

	log.Printf("write_file - attempting to write %d bytes to: %s", len(content), path)
	if !h.isPathAllowed(path) {
		log.Printf("ERROR: write_file - access denied to path: %s", path)
		return nil, NewAccessDeniedError(path)
	}

	// Auto-create parent directories if they don't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("ERROR: write_file - failed to create parent directories for %s: %v", path, err)
		return nil, fmt.Errorf("failed to create parent directories: %w", err)
	}

	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		log.Printf("ERROR: write_file - failed to write to %s: %v", path, err)
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	bytesWritten := len(content)
	log.Printf("write_file - successfully wrote %d bytes to %s", bytesWritten, path)
	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully wrote %d bytes to: %s", bytesWritten, path),
			},
		},
	}, nil
}

func (h *FileSystemHandler) handleMoveFile(args map[string]interface{}) (*protocol.CallToolResponse, error) {
	source, ok := args["source"].(string)
	if !ok {
		log.Printf("ERROR: move_file - invalid source path type: %T", args["source"])
		return nil, fmt.Errorf("source must be a string")
	}
	destination, ok := args["destination"].(string)
	if !ok {
		log.Printf("ERROR: move_file - invalid destination path type: %T", args["destination"])
		return nil, fmt.Errorf("destination must be a string")
	}

	log.Printf("move_file - attempting to move %s to %s", source, destination)
	if !h.isPathAllowed(source) {
		log.Printf("ERROR: move_file - access denied to source path: %s", source)
		return nil, NewAccessDeniedError(source)
	}
	if !h.isPathAllowed(destination) {
		log.Printf("ERROR: move_file - access denied to destination path: %s", destination)
		return nil, NewAccessDeniedError(destination)
	}

	err := os.Rename(source, destination)
	if err != nil {
		log.Printf("ERROR: move_file - failed to move %s to %s: %v", source, destination, err)
		return nil, fmt.Errorf("failed to move file: %w", err)
	}

	log.Printf("move_file - successfully moved %s to %s", source, destination)
	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully moved %s to %s", source, destination),
			},
		},
	}, nil
}
