package handler

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gomcpgo/mcp/pkg/protocol"
)

func (h *FileSystemHandler) handleCreateDirectory(args map[string]interface{}) (*protocol.CallToolResponse, error) {
	path, ok := args["path"].(string)
	if !ok {
		log.Printf("ERROR: create_directory - invalid path type: %T", args["path"])
		return nil, fmt.Errorf("path must be a string")
	}

	log.Printf("create_directory - attempting to create directory: %s", path)
	if !h.isPathAllowed(path) {
		log.Printf("ERROR: create_directory - access denied to path: %s", path)
		return nil, fmt.Errorf("access to path is not allowed: %s", path)
	}

	err := os.MkdirAll(path, 0755)
	if err != nil {
		log.Printf("ERROR: create_directory - failed to create directory %s: %v", path, err)
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	log.Printf("create_directory - successfully created directory: %s", path)
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
		log.Printf("ERROR: list_directory - invalid path type: %T", args["path"])
		return nil, fmt.Errorf("path must be a string")
	}

	log.Printf("list_directory - attempting to list directory: %s", path)
	if !h.isPathAllowed(path) {
		log.Printf("ERROR: list_directory - access denied to path: %s", path)
		return nil, fmt.Errorf("access to path is not allowed: %s", path)
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		log.Printf("ERROR: list_directory - failed to read directory %s: %v", path, err)
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

	log.Printf("list_directory - successfully listed %d entries in %s", len(entries), path)
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
	log.Printf("list_allowed_directories - retrieving allowed directories")
	dirs, err := getAllowedDirs()
	if err != nil {
		log.Printf("ERROR: list_allowed_directories - failed to get allowed directories: %v", err)
		return nil, fmt.Errorf("failed to get allowed directories: %w", err)
	}

	log.Printf("list_allowed_directories - found %d allowed directories", len(dirs))
	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Allowed directories:\n%s", strings.Join(dirs, "\n")),
			},
		},
	}, nil
}
