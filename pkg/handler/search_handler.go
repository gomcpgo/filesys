package handler

import (
	"fmt"
	"log"

	"github.com/gomcpgo/filesys/pkg/search"
	"github.com/gomcpgo/mcp/pkg/protocol"
)

// handleSearchFiles implements the search_files tool using the search package
func (h *FileSystemHandler) handleSearchFiles(args map[string]interface{}) (*protocol.CallToolResponse, error) {
	// Extract path parameter
	path, ok := args["path"].(string)
	if !ok {
		log.Printf("ERROR: search_files - invalid path type: %T", args["path"])
		return nil, fmt.Errorf("path must be a string")
	}

	// Extract pattern parameter
	pattern, ok := args["pattern"].(string)
	if !ok {
		log.Printf("ERROR: search_files - invalid pattern type: %T", args["pattern"])
		return nil, fmt.Errorf("pattern must be a string")
	}

	log.Printf("search_files - searching in %s for pattern: %s", path, pattern)

	// Check if the path is allowed
	if !h.isPathAllowed(path) {
		log.Printf("ERROR: search_files - access denied to path: %s", path)
		return nil, fmt.Errorf("access to path is not allowed: %s", path)
	}

	// Use our new search package
	options := search.DefaultSearchOptions()

	// Extract optional parameters if provided
	if caseSensitive, ok := args["case_sensitive"].(bool); ok {
		options.CaseSensitive = caseSensitive
	}

	// MaxDepth option has been removed

	if matchPath, ok := args["match_path"].(bool); ok {
		options.MatchPath = matchPath
	}

	// Perform the search
	result, err := search.Search(path, pattern, options)
	if err != nil {
		log.Printf("ERROR: search_files - search failed: %v", err)
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Log warnings if any occurred during search
	for _, searchErr := range result.Errors {
		log.Printf("WARNING: search_files - %v", searchErr)
	}

	log.Printf("search_files - found %d matches for pattern %q in %s",
		len(result.Matches), pattern, path)

	// Format the results
	formattedMatches := result.FormatMatches()

	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: formattedMatches,
			},
		},
	}, nil
}
