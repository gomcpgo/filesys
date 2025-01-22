package handler

import (
	"fmt"
	"os"
	"strings"

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

func (h *FileSystemHandler) handleUpdateFileSection(args map[string]interface{}) (*protocol.CallToolResponse, error) {
	// Extract and validate parameters
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path must be a string")
	}

	startLine, ok := args["startLine"].(float64) // JSON numbers come as float64
	if !ok {
		return nil, fmt.Errorf("startLine must be a number")
	}

	endLine, ok := args["endLine"].(float64)
	if !ok {
		return nil, fmt.Errorf("endLine must be a number")
	}

	newContent, ok := args["newContent"].(string)
	if !ok {
		return nil, fmt.Errorf("newContent must be a string")
	}

	// Convert to int and validate line numbers
	start := int(startLine)
	end := int(endLine)
	if start < 1 {
		return nil, fmt.Errorf("startLine must be >= 1")
	}
	if end < start {
		return nil, fmt.Errorf("endLine must be >= startLine")
	}

	// Check path permissions
	if !h.isPathAllowed(path) {
		return nil, fmt.Errorf("access to path is not allowed: %s", path)
	}

	// Read the entire file
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Split into lines
	lines := strings.Split(string(content), "\n")

	// Validate line numbers against file length
	if start > len(lines) {
		return nil, fmt.Errorf("startLine %d is beyond end of file (%d lines)", start, len(lines))
	}
	if end > len(lines) {
		return nil, fmt.Errorf("endLine %d is beyond end of file (%d lines)", end, len(lines))
	}

	// Create new content
	newLines := make([]string, 0, len(lines)-(end-start+1)+1)
	newLines = append(newLines, lines[:start-1]...)                 // Lines before the section
	newLines = append(newLines, strings.Split(newContent, "\n")...) // New content
	newLines = append(newLines, lines[end:]...)                     // Lines after the section

	// Write back to file
	err = os.WriteFile(path, []byte(strings.Join(newLines, "\n")), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully updated lines %d-%d in %s", start, end, path),
			},
		},
	}, nil
}
