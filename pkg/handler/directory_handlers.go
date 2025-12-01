package handler

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gomcpgo/filesys/pkg/dirlist"
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
		return nil, NewAccessDeniedError(path)
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
		return nil, NewAccessDeniedError(path)
	}
	
	// Extract optional parameters
	options := dirlist.DefaultListOptions()
	
	if pattern, ok := args["pattern"].(string); ok {
		options.Pattern = pattern
	}
	
	if fileType, ok := args["file_type"].(string); ok {
		options.FileType = fileType
	}
	
	if recursive, ok := args["recursive"].(bool); ok {
		options.Recursive = recursive
	}
	
	if maxDepth, ok := args["max_depth"].(float64); ok {
		options.MaxDepth = int(maxDepth)
	}
	
	if maxResults, ok := args["max_results"].(float64); ok {
		options.MaxResults = int(maxResults)
	}
	
	if includeHidden, ok := args["include_hidden"].(bool); ok {
		options.IncludeHidden = includeHidden
	}
	
	if includeMetadata, ok := args["include_metadata"].(bool); ok {
		options.IncludeMetadata = includeMetadata
	}
	
	// Get directory listing using the dirlist package
	result, err := dirlist.ListDirectory(path, options)
	if err != nil {
		log.Printf("ERROR: list_directory - failed to list directory %s: %v", path, err)
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}
	
	// Format the results
	var lines []string
	
	// Add header and summary
	lines = append(lines, fmt.Sprintf("DIRECTORY LISTING: %s", path))
	lines = append(lines, fmt.Sprintf("Total entries: %d (%d files, %d directories)", 
		result.TotalFiles+result.TotalDirs, result.TotalFiles, result.TotalDirs))
	
	if result.TotalFiles > 0 {
		lines = append(lines, fmt.Sprintf("Total size: %d bytes", result.TotalSize))
	}
	
	if result.Truncated {
		lines = append(lines, fmt.Sprintf("Note: Results truncated (showing %d of %d entries)", 
			len(result.Entries), result.TotalEntries))
	}
	
	lines = append(lines, "")
	
	// Add entries
	for _, entry := range result.Entries {
		var entryInfo string
		
		if entry.IsDir {
			// Format directory entry
			if options.IncludeMetadata {
				entryInfo = fmt.Sprintf("[DIR ] %s | %d items | %s | %s", 
					entry.Name, entry.ItemCount, entry.ModTime.Format(time.RFC3339), entry.Mode)
			} else {
				entryInfo = fmt.Sprintf("[DIR ] %s", entry.Name)
			}
		} else {
			// Format file entry
			if options.IncludeMetadata {
				entryInfo = fmt.Sprintf("[FILE] %s | %d bytes | %s | %s", 
					entry.Name, entry.Size, entry.ModTime.Format(time.RFC3339), entry.Mode)
			} else {
				entryInfo = fmt.Sprintf("[FILE] %s", entry.Name)
			}
		}
		
		lines = append(lines, entryInfo)
	}
	
	log.Printf("list_directory - successfully listed %d entries in %s", len(result.Entries), path)
	
	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: strings.Join(lines, "\n"),
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
