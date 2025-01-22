package handler

import (
	"fmt"
	"os"

	"github.com/gomcpgo/mcp/pkg/protocol"
)

func (h *FileSystemHandler) handleWriteFile(args map[string]interface{}) (*protocol.CallToolResponse, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path must be a string")
	}

	content, ok := args["content"].(string)
	if !ok {
		return nil, fmt.Errorf("content must be a string")
	}

	if !h.isPathAllowed(path) {
		return nil, fmt.Errorf("access to path is not allowed: %s", path)
	}

	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully wrote file: %s", path),
			},
		},
	}, nil
}

func (h *FileSystemHandler) handleMoveFile(args map[string]interface{}) (*protocol.CallToolResponse, error) {
	source, ok := args["source"].(string)
	if !ok {
		return nil, fmt.Errorf("source must be a string")
	}

	destination, ok := args["destination"].(string)
	if !ok {
		return nil, fmt.Errorf("destination must be a string")
	}

	if !h.isPathAllowed(source) || !h.isPathAllowed(destination) {
		return nil, fmt.Errorf("access to path is not allowed")
	}

	err := os.Rename(source, destination)
	if err != nil {
		return nil, fmt.Errorf("failed to move file: %w", err)
	}

	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully moved %s to %s", source, destination),
			},
		},
	}, nil
}