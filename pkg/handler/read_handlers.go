package handler

import (
	"fmt"
	"log"
	"os"
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
		return nil, fmt.Errorf("access to path is not allowed: %s", path)
	}

	// Use our smart file reading function
	result, err := fileread.ReadFileLines(path, startLine, endLine, maxFileSize)
	if err != nil {
		log.Printf("ERROR: read_file - failed to read file %s: %v", path, err)
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Prepare a message with metadata about the read operation
	var metadataBuilder strings.Builder
	
	// If the file was truncated, add a warning message
	if result.Truncated {
		metadataBuilder.WriteString("⚠️ File content was truncated due to size limits.\n")
	}
	
	// Add information about line range
	metadataBuilder.WriteString(fmt.Sprintf("File: %s\n", path))
	metadataBuilder.WriteString(fmt.Sprintf("Total lines: %d\n", result.TotalLines))
	
	if result.StartLine > 1 || result.EndLine < result.TotalLines {
		metadataBuilder.WriteString(fmt.Sprintf("Showing lines %d to %d\n", result.StartLine, result.EndLine))
	}
	
	metadataBuilder.WriteString(fmt.Sprintf("Content size: %d bytes\n", result.ContentSize))
	
	// If there's metadata and content, add a separator
	var responseText string
	if metadataBuilder.Len() > 0 && len(result.Content) > 0 {
		responseText = metadataBuilder.String() + "\n---\n\n" + result.Content
	} else if metadataBuilder.Len() > 0 {
		responseText = metadataBuilder.String()
	} else {
		responseText = result.Content
	}

	log.Printf("read_file - successfully read %d bytes (%d lines) from %s", 
		result.ContentSize, result.ReadLines, path)
		
	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: responseText,
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

		// Get file info to check size before reading
		fileInfo, err := os.Stat(path)
		if err != nil {
			log.Printf("ERROR: read_multiple_files - failed to get file info %s: %v", path, err)
			results = append(results, fmt.Sprintf("Error getting file info for %s: %v", path, err))
			continue
		}

		// Check file size
		if fileInfo.Size() > maxFileSize {
			log.Printf("ERROR: read_multiple_files - file size %d bytes exceeds maximum allowed size of %d bytes: %s", fileInfo.Size(), maxFileSize, path)
			results = append(results, fmt.Sprintf("File size %d bytes exceeds maximum allowed size of %d bytes: %s", fileInfo.Size(), maxFileSize, path))
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