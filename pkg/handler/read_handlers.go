package handler

import (
	"fmt"
	"os"
	"strings"

	"github.com/gomcpgo/mcp/pkg/protocol"
)

func (h *FileSystemHandler) handleReadFile(args map[string]interface{}) (*protocol.CallToolResponse, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path must be a string")
	}

	if !h.isPathAllowed(path) {
		return nil, fmt.Errorf("access to path is not allowed: %s", path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: string(content),
			},
		},
	}, nil
}

func (h *FileSystemHandler) handleReadMultipleFiles(args map[string]interface{}) (*protocol.CallToolResponse, error) {
	pathsInterface, ok := args["paths"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("paths must be an array")
	}

	var results []string
	for _, pathInterface := range pathsInterface {
		path, ok := pathInterface.(string)
		if !ok {
			results = append(results, fmt.Sprintf("Invalid path type: %v", pathInterface))
			continue
		}

		if !h.isPathAllowed(path) {
			results = append(results, fmt.Sprintf("Access denied: %s", path))
			continue
		}

		content, err := os.ReadFile(path)
		if err != nil {
			results = append(results, fmt.Sprintf("Error reading %s: %v", path, err))
			continue
		}

		results = append(results, fmt.Sprintf("=== %s ===\n%s", path, string(content)))
	}

	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: strings.Join(results, "\n\n"),
			},
		},
	}, nil
}