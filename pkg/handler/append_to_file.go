package handler

import (
	"fmt"
	"log"
	"os"
	"strings"
	"github.com/gomcpgo/mcp/pkg/protocol"
)

// handleAppendToFile adds content to the end of a file
func (h *FileSystemHandler) handleAppendToFile(args map[string]interface{}) (*protocol.CallToolResponse, error) {
	path, ok := args["path"].(string)
	if !ok {
		log.Printf("ERROR: append_to_file - invalid path type: %T", args["path"])
		return nil, fmt.Errorf("path must be a string")
	}
	
	content, ok := args["content"].(string)
	if !ok {
		log.Printf("ERROR: append_to_file - invalid content type: %T", args["content"])
		return nil, fmt.Errorf("content must be a string")
	}
	
	log.Printf("append_to_file - attempting to append %d bytes to: %s", len(content), path)
	
	if !h.isPathAllowed(path) {
		log.Printf("ERROR: append_to_file - access denied to path: %s", path)
		return nil, fmt.Errorf("access to path is not allowed: %s", path)
	}
	
	// Check if file exists
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// If file doesn't exist, create it
			err = os.WriteFile(path, []byte(content), 0644)
			if err != nil {
				log.Printf("ERROR: append_to_file - failed to create file %s: %v", path, err)
				return nil, fmt.Errorf("failed to create file: %w", err)
			}
			log.Printf("append_to_file - created new file %s with %d bytes", path, len(content))
			return &protocol.CallToolResponse{
				Content: []protocol.ToolContent{
					{
						Type: "text",
						Text: fmt.Sprintf("Created new file %s with provided content", path),
					},
				},
			}, nil
		}
		log.Printf("ERROR: append_to_file - failed to check file %s: %v", path, err)
		return nil, fmt.Errorf("failed to check file: %w", err)
	}
	
	// Make sure it's a regular file
	if !fileInfo.Mode().IsRegular() {
		log.Printf("ERROR: append_to_file - %s is not a regular file", path)
		return nil, fmt.Errorf("%s is not a regular file", path)
	}
	
	// Read existing content
	existingContent, err := os.ReadFile(path)
	if err != nil {
		log.Printf("ERROR: append_to_file - failed to read file %s: %v", path, err)
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	// Append new content
	var newContent string
	if len(existingContent) > 0 && !strings.HasSuffix(string(existingContent), "\n") {
		// Add a newline if the file doesn't end with one
		newContent = string(existingContent) + "\n" + content
	} else {
		newContent = string(existingContent) + content
	}
	
	// Write back to file
	err = os.WriteFile(path, []byte(newContent), 0644)
	if err != nil {
		log.Printf("ERROR: append_to_file - failed to write to %s: %v", path, err)
		return nil, fmt.Errorf("failed to write file: %w", err)
	}
	
	log.Printf("append_to_file - successfully appended %d bytes to %s", len(content), path)
	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully appended content to %s", path),
			},
		},
	}, nil
}