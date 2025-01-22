package handler

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gomcpgo/mcp/pkg/protocol"
)

func (h *FileSystemHandler) handleReadFile(args map[string]interface{}) (*protocol.CallToolResponse, error) {
	path, ok := args["path"].(string)
	if !ok {
		log.Printf("ERROR: read_file - invalid path type: %T", args["path"])
		return nil, fmt.Errorf("path must be a string")
	}
	log.Printf("read_file - attempting to read file: %s", path)

	if !h.isPathAllowed(path) {
		log.Printf("ERROR: read_file - access denied to path: %s", path)
		return nil, fmt.Errorf("access to path is not allowed: %s", path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		log.Printf("ERROR: read_file - failed to read file %s: %v", path, err)
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	log.Printf("read_file - successfully read %d bytes from %s", len(content), path)
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
		log.Printf("ERROR: read_multiple_files - invalid paths type: %T", args["paths"])
		return nil, fmt.Errorf("paths must be an array")
	}

	log.Printf("read_multiple_files - attempting to read %d files", len(pathsInterface))
	var results []string

	for i, pathInterface := range pathsInterface {
		path, ok := pathInterface.(string)
		if !ok {
			log.Printf("ERROR: read_multiple_files - invalid path type at index %d: %T", i, pathInterface)
			results = append(results, fmt.Sprintf("Invalid path type: %v", pathInterface))
			continue
		}

		log.Printf("read_multiple_files - processing file %d/%d: %s", i+1, len(pathsInterface), path)

		if !h.isPathAllowed(path) {
			log.Printf("ERROR: read_multiple_files - access denied to path: %s", path)
			results = append(results, fmt.Sprintf("Access denied: %s", path))
			continue
		}

		content, err := os.ReadFile(path)
		if err != nil {
			log.Printf("ERROR: read_multiple_files - failed to read file %s: %v", path, err)
			results = append(results, fmt.Sprintf("Error reading %s: %v", path, err))
			continue
		}

		log.Printf("read_multiple_files - successfully read %d bytes from %s", len(content), path)
		results = append(results, fmt.Sprintf("=== %s ===\n%s", path, string(content)))
	}

	log.Printf("read_multiple_files - completed reading files: %d successful results", len(results))
	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: strings.Join(results, "\n\n"),
			},
		},
	}, nil
}
