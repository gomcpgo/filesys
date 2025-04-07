package handler

import (
	"fmt"
	"log"
	"strings"

	"github.com/gomcpgo/filesys/pkg/search"
	"github.com/gomcpgo/mcp/pkg/protocol"
)

// handleSearchInFiles searches for regex patterns within files in a directory
func (h *FileSystemHandler) handleSearchInFiles(args map[string]interface{}) (*protocol.CallToolResponse, error) {
	// Extract path parameter
	path, ok := args["path"].(string)
	if !ok {
		log.Printf("ERROR: search_in_files - invalid path type: %T", args["path"])
		return nil, fmt.Errorf("path must be a string")
	}

	// Extract pattern parameter
	pattern, ok := args["pattern"].(string)
	if !ok {
		log.Printf("ERROR: search_in_files - invalid pattern type: %T", args["pattern"])
		return nil, fmt.Errorf("pattern must be a string")
	}

	// Extract optional parameters
	maxResults := 100 // Default
	if maxResultsVal, ok := args["max_results"].(float64); ok {
		maxResults = int(maxResultsVal)
	}

	maxFileSearches := 100 // Default
	if maxFileSearchesVal, ok := args["max_file_searches"].(float64); ok {
		maxFileSearches = int(maxFileSearchesVal)
	}

	caseSensitive := true // Default
	if caseSensitiveVal, ok := args["case_sensitive"].(bool); ok {
		caseSensitive = caseSensitiveVal
	}

	fileExtensions := []string{} // Default: all extensions
	if extensionsInterface, ok := args["file_extensions"].([]interface{}); ok {
		for _, ext := range extensionsInterface {
			if extStr, ok := ext.(string); ok {
				fileExtensions = append(fileExtensions, extStr)
			}
		}
	}

	log.Printf("search_in_files - searching in %s for pattern: %s", path, pattern)

	// Check if path is allowed
	if !h.isPathAllowed(path) {
		log.Printf("ERROR: search_in_files - access denied to path: %s", path)
		return nil, fmt.Errorf("access to path is not allowed: %s", path)
	}

	// Configure search options
	options := search.SearchOptions{
		RootDir:         path,
		Pattern:         pattern,
		FileExtensions:  fileExtensions,
		MaxFileSearches: maxFileSearches,
		MaxResults:      maxResults,
		CaseSensitive:   caseSensitive,
	}

	// Perform the search
	result, err := search.Search(options)
	if err != nil {
		log.Printf("ERROR: search_in_files - search failed: %v", err)
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Format the results
	var lines []string
	lines = append(lines, fmt.Sprintf("Search results for pattern '%s'", pattern))
	lines = append(lines, fmt.Sprintf("Files searched: %d, Files matched: %d, Total matches: %d", 
		result.FilesSearched, result.FilesMatched, result.TotalMatches))
	lines = append(lines, "")

	// Group matches by file
	fileMatches := make(map[string][]search.SearchMatch)
	for _, match := range result.Matches {
		fileMatches[match.FilePath] = append(fileMatches[match.FilePath], match)
	}

	// Add formatted results
	for filePath, matches := range fileMatches {
		lines = append(lines, fmt.Sprintf("File: %s", filePath))
		for _, match := range matches {
			// Trim and clean line content for display
			lineContent := strings.TrimSpace(match.LineContent)
			if len(lineContent) > 100 {
				// Truncate long lines
				lineContent = lineContent[:97] + "..."
			}
			lines = append(lines, fmt.Sprintf("  Line %d: %s", match.LineNumber, lineContent))
		}
		lines = append(lines, "")
	}

	log.Printf("search_in_files - found %d matches in %d files", result.TotalMatches, result.FilesMatched)
	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: strings.Join(lines, "\n"),
			},
		},
	}, nil
}
