package handler

import (
	"fmt"
	"log"
	"os"
	"strings"
	"github.com/gomcpgo/mcp/pkg/protocol"
)

// handleInsertBeforeString inserts content before a specific occurrence of a string
func (h *FileSystemHandler) handleInsertBeforeString(args map[string]interface{}) (*protocol.CallToolResponse, error) {
	path, ok := args["path"].(string)
	if !ok {
		log.Printf("ERROR: insert_before_string - invalid path type: %T", args["path"])
		return nil, fmt.Errorf("path must be a string")
	}
	
	searchString, ok := args["search"].(string)
	if !ok {
		log.Printf("ERROR: insert_before_string - invalid search string type: %T", args["search"])
		return nil, fmt.Errorf("search string must be a string")
	}
	
	contentToInsert, ok := args["content"].(string)
	if !ok {
		log.Printf("ERROR: insert_before_string - invalid content type: %T", args["content"])
		return nil, fmt.Errorf("content must be a string")
	}
	
	// Default to the first occurrence
	occurrence := 1
	if occurrenceVal, ok := args["occurrence"].(float64); ok {
		occurrence = int(occurrenceVal)
		if occurrence < 1 {
			log.Printf("ERROR: insert_before_string - invalid occurrence: %d", occurrence)
			return nil, fmt.Errorf("occurrence must be a positive integer")
		}
	}
	
	log.Printf("insert_before_string - attempting to insert before occurrence %d of '%s' in %s", 
		occurrence, searchString, path)
	
	if !h.isPathAllowed(path) {
		log.Printf("ERROR: insert_before_string - access denied to path: %s", path)
		return nil, fmt.Errorf("access to path is not allowed: %s", path)
	}
	
	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		log.Printf("ERROR: insert_before_string - failed to read file %s: %v", path, err)
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	fileContent := string(content)
	totalOccurrences := strings.Count(fileContent, searchString)
	
	if totalOccurrences == 0 {
		log.Printf("ERROR: insert_before_string - search string '%s' not found in %s", searchString, path)
		return nil, fmt.Errorf("string '%s' not found in %s", searchString, path)
	}
	
	if occurrence > totalOccurrences {
		log.Printf("ERROR: insert_before_string - specified occurrence %d exceeds total occurrences %d in %s", 
			occurrence, totalOccurrences, path)
		return nil, fmt.Errorf("specified occurrence %d exceeds total occurrences %d", 
			occurrence, totalOccurrences)
	}
	
	// Find the position before the specified occurrence
	currentOccurrence := 0
	lastIndex := 0
	insertPosition := -1
	
	for currentOccurrence < occurrence {
		index := strings.Index(fileContent[lastIndex:], searchString)
		if index == -1 {
			break
		}
		
		absoluteIndex := lastIndex + index
		currentOccurrence++
		
		if currentOccurrence == occurrence {
			insertPosition = absoluteIndex
			break
		}
		
		lastIndex = absoluteIndex + len(searchString)
	}
	
	if insertPosition == -1 {
		log.Printf("ERROR: insert_before_string - failed to find insertion position")
		return nil, fmt.Errorf("failed to find insertion position")
	}
	
	// Create new content with inserted text
	newContent := fileContent[:insertPosition] + contentToInsert + fileContent[insertPosition:]
	
	// Write back to file
	err = os.WriteFile(path, []byte(newContent), 0644)
	if err != nil {
		log.Printf("ERROR: insert_before_string - failed to write to %s: %v", path, err)
		return nil, fmt.Errorf("failed to write file: %w", err)
	}
	
	log.Printf("insert_before_string - successfully inserted content before occurrence %d of '%s' in %s", 
		occurrence, searchString, path)
	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully inserted content before occurrence %d of '%s' in %s", 
					occurrence, searchString, path),
			},
		},
	}, nil
}