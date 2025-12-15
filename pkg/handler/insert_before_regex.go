package handler

import (
	"fmt"
	"log"
	"os"
	"github.com/gomcpgo/filesys/pkg/search"
	"github.com/gomcpgo/mcp/pkg/protocol"
)

// handleInsertBeforeRegex inserts content before a specific occurrence of a regex pattern
func (h *FileSystemHandler) handleInsertBeforeRegex(args map[string]interface{}) (*protocol.CallToolResponse, error) {
	path, ok := args["path"].(string)
	if !ok {
		log.Printf("ERROR: insert_before_regex - invalid path type: %T", args["path"])
		return nil, fmt.Errorf("path must be a string")
	}
	
	pattern, ok := args["pattern"].(string)
	if !ok {
		log.Printf("ERROR: insert_before_regex - invalid pattern type: %T", args["pattern"])
		return nil, fmt.Errorf("pattern must be a string")
	}
	
	contentToInsert, ok := args["content"].(string)
	if !ok {
		log.Printf("ERROR: insert_before_regex - invalid content type: %T", args["content"])
		return nil, fmt.Errorf("content must be a string")
	}
	
	// Default to the first occurrence
	occurrence := 1
	if occurrenceVal, ok := args["occurrence"].(float64); ok {
		occurrence = int(occurrenceVal)
		if occurrence < 0 {
			log.Printf("ERROR: insert_before_regex - invalid occurrence: %d", occurrence)
			return nil, fmt.Errorf("occurrence must be a non-negative integer (0 for all occurrences, 1 or more for specific occurrence)")
		}
	}
	
	log.Printf("insert_before_regex - attempting to insert before occurrence %d of pattern '%s' in %s", 
		occurrence, pattern, path)
	
	if !h.isPathAllowed(path) {
		log.Printf("ERROR: insert_before_regex - access denied to path: %s", path)
		return nil, NewAccessDeniedError(path)
	}
	
	// Use the search package to insert content before regex pattern
	newContent, err := search.InsertBeforeRegex(path, pattern, contentToInsert, occurrence)
	if err != nil {
		log.Printf("ERROR: insert_before_regex - %v", err)
		return nil, err
	}
	
	// Write the new content back to the file
	err = os.WriteFile(path, []byte(newContent), 0644)
	if err != nil {
		log.Printf("ERROR: insert_before_regex - failed to write to %s: %v", path, err)
		return nil, fmt.Errorf("failed to write file: %w", err)
	}
	
	if occurrence == 0 {
		log.Printf("insert_before_regex - successfully inserted content before all occurrences of pattern '%s' in %s", 
			pattern, path)
		return &protocol.CallToolResponse{
			Content: []protocol.ToolContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Successfully inserted content before all occurrences of pattern '%s' in %s", 
						pattern, path),
				},
			},
		}, nil
	} else {
		log.Printf("insert_before_regex - successfully inserted content before occurrence %d of pattern '%s' in %s", 
			occurrence, pattern, path)
		return &protocol.CallToolResponse{
			Content: []protocol.ToolContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Successfully inserted content before occurrence %d of pattern '%s' in %s", 
						occurrence, pattern, path),
				},
			},
		}, nil
	}
}
