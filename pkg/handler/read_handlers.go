package handler

import (
	"fmt"
	"log"
	"strings"

	"github.com/gomcpgo/filesys/pkg/fileread"
	"github.com/gomcpgo/mcp/pkg/protocol"
)

const maxFileSize = 5 * 1024 * 1024 // 5MB in bytes

func (h *FileSystemHandler) handleReadFile(args map[string]interface{}) (*protocol.CallToolResponse, error) {
	path, ok := args["path"].(string)
	if !ok {
		log.Printf("ERROR: read_file - invalid path type: %T", args["path"])
		return nil, fmt.Errorf("path must be a string")
	}

	// Extract optional parameters
	startLine := 0
	endLine := 0
	
	if startLineVal, ok := args["start_line"].(float64); ok {
		startLine = int(startLineVal)
	}
	
	if endLineVal, ok := args["end_line"].(float64); ok {
		endLine = int(endLineVal)
	}

	log.Printf("read_file - attempting to read file: %s (lines %d to %d)", path, startLine, endLine)

	if !h.isPathAllowed(path) {
		log.Printf("ERROR: read_file - access denied to path: %s", path)
		return nil, NewAccessDeniedError(path)
	}

	// Use our smart file reading function
	result, err := fileread.ReadFile(path, startLine, endLine, maxFileSize)
	if err != nil {
		log.Printf("ERROR: read_file - failed to read file %s: %v", path, err)
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Create content array - first element is ALWAYS the exact file content
	contentArray := []protocol.ToolContent{
		{
			Type: "text",
			Text: result.Content,
		},
	}
	
	// Only add metadata if it's a partial read or truncated
	if result.IsPartial || result.Truncated {
		var metadataBuilder strings.Builder
		
		// If the file was truncated, add a warning message
		if result.Truncated {
			metadataBuilder.WriteString("⚠️ File content was truncated due to size limits.\n")
		}
		
		// Add information about file and line range
		metadataBuilder.WriteString(fmt.Sprintf("File: %s\n", path))
		metadataBuilder.WriteString(fmt.Sprintf("Total lines: %d\n", result.TotalLines))
		
		if result.StartLine > 1 || (result.EndLine > 0 && result.EndLine < result.TotalLines) {
			metadataBuilder.WriteString(fmt.Sprintf("Showing lines %d to %d\n", result.StartLine, result.EndLine))
		}
		
		metadataBuilder.WriteString(fmt.Sprintf("Content size: %d bytes\n", result.ContentSize))
		
		// Add metadata as second content element
		contentArray = append(contentArray, protocol.ToolContent{
			Type: "text",
			Text: metadataBuilder.String(),
		})
	}

	log.Printf("read_file - successfully read %d bytes (%d lines) from %s", 
		result.ContentSize, result.ReadLines, path)
		
	return &protocol.CallToolResponse{
		Content: contentArray,
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

		// Use our optimized file reading function 
		// For multiple files, we always read the entire file (no line ranges)
		result, err := fileread.ReadFile(path, 0, 0, maxFileSize)
		if err != nil {
			log.Printf("ERROR: read_multiple_files - failed to read file %s: %v", path, err)
			results = append(results, fmt.Sprintf("Error reading %s: %v", path, err))
			continue
		}

		// Add metadata if the file was truncated
		if result.Truncated {
			results = append(results, fmt.Sprintf("=== %s ===\n⚠️ File content was truncated (showing %d of %d bytes)\n%s", 
				path, result.ContentSize, result.FileSize, result.Content))
		} else {
			results = append(results, fmt.Sprintf("=== %s ===\n%s", path, result.Content))
		}
		
		log.Printf("read_multiple_files - successfully read %d bytes from %s", result.ContentSize, path)
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