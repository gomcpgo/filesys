package handler

import (
	"fmt"
	"os"
	"strings"

	"github.com/gomcpgo/mcp/pkg/protocol"
)

func (h *FileSystemHandler) handleCreateDirectory(args map[string]interface{}) (*protocol.CallToolResponse, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path must be a string")
	}

	if !h.isPathAllowed(path) {
		return nil, fmt.Errorf("access to path is not allowed: %s", path)
	}

	err := os.MkdirAll(path, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully created directory: %s", path),
			},
		},
	}, nil
}

func (h *FileSystemHandler) handleListDirectory(args map[string]interface{}) (*protocol.CallToolResponse, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path must be a string")
	}

	if !h.isPathAllowed(path) {
		return nil, fmt.Errorf("access to path is not allowed: %s", path)
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var listing []string
	for _, entry := range entries {
		prefix := "[FILE]"
		if entry.IsDir() {
			prefix = "[DIR ]"
		}
		listing = append(listing, fmt.Sprintf("%s %s", prefix, entry.Name()))
	}

	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: strings.Join(listing, "\n"),
			},
		},
	}, nil
}

func (h *FileSystemHandler) handleListAllowedDirectories() (*protocol.CallToolResponse, error) {
	dirs, err := getAllowedDirs()
	if err != nil {
		return nil, fmt.Errorf("failed to get allowed directories: %w", err)
	}

	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Allowed directories:\n%s", strings.Join(dirs, "\n")),
			},
		},
	}, nil
}