package handler

import (
	"fmt"
	"log"
	"os"
	"strings"
	"github.com/gomcpgo/mcp/pkg/protocol"
)

// handleReplaceInFile replaces occurrences of a string in a file
func (h *FileSystemHandler) handleReplaceInFile(args map[string]interface{}) (*protocol.CallToolResponse, error) {
	path, ok := args["path"].(string)
	if !ok {
		log.Printf("ERROR: replace_in_file - invalid path type: %T", args["path"])
		return nil, fmt.Errorf("path must be a string")
	}
	
	searchString, ok := args["search"].(string)
	if !ok {
		log.Printf("ERROR: replace_in_file - invalid search string type: %T", args["search"])
		return nil, fmt.Errorf("search string must be a string")
	}
	
	replaceString, ok := args["replace"].(string)
	if !ok {
		log.Printf("ERROR: replace_in_file - invalid replace string type: %T", args["replace"])
		return nil, fmt.Errorf("replace string must be a string")
	}
	
	// Optional parameter for which occurrence to replace
	occurrence := 0 // 0 means replace all
	if occurrenceVal, ok := args["occurrence"].(float64); ok {
		occurrence = int(occurrenceVal)
	}
	
	log.Printf("replace_in_file - attempting to replace '%s' with '%s' in %s", searchString, replaceString, path)
	
	if !h.isPathAllowed(path) {
		log.Printf("ERROR: replace_in_file - access denied to path: %s", path)
		return nil, fmt.Errorf("access to path is not allowed: %s", path)
	}
	
	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		log.Printf("ERROR: replace_in_file - failed to read file %s: %v", path, err)
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	fileContent := string(content)
	totalOccurrences := strings.Count(fileContent, searchString)
	
	if totalOccurrences == 0 {
		log.Printf("replace_in_file - search string '%s' not found in %s", searchString, path)
		return &protocol.CallToolResponse{
			Content: []protocol.ToolContent{
				{
					Type: "text",
					Text: fmt.Sprintf("String '%s' not found in %s", searchString, path),
				},
			},
		}, nil
	}
	
	var newContent string
	var replacedCount int
	
	if occurrence == 0 {
		// Replace all occurrences
		newContent = strings.ReplaceAll(fileContent, searchString, replaceString)
		replacedCount = totalOccurrences
	} else {
		if occurrence > totalOccurrences {
			log.Printf("ERROR: replace_in_file - specified occurrence %d exceeds total occurrences %d in %s", 
				occurrence, totalOccurrences, path)
			return nil, fmt.Errorf("specified occurrence %d exceeds total occurrences %d", 
				occurrence, totalOccurrences)
		}
		
		// Replace nth occurrence only
		currentOccurrence := 0
		lastIndex := 0
		newContent = ""
		
		for {
			index := strings.Index(fileContent[lastIndex:], searchString)
			if index == -1 {
				// Add the remaining content
				newContent += fileContent[lastIndex:]
				break
			}
			
			absoluteIndex := lastIndex + index
			currentOccurrence++
			
			// Add content up to this occurrence
			newContent += fileContent[lastIndex:absoluteIndex]
			
			// For the specified occurrence, add the replacement; for others, keep the original
			if currentOccurrence == occurrence {
				newContent += replaceString
				replacedCount = 1
			} else {
				newContent += searchString
			}
			
			lastIndex = absoluteIndex + len(searchString)
			
			// If we've processed all content, break
			if lastIndex >= len(fileContent) {
				break
			}
		}
	}
	
	// Write back to file
	err = os.WriteFile(path, []byte(newContent), 0644)
	if err != nil {
		log.Printf("ERROR: replace_in_file - failed to write to %s: %v", path, err)
		return nil, fmt.Errorf("failed to write file: %w", err)
	}
	
	log.Printf("replace_in_file - successfully replaced %d occurrence(s) in %s", replacedCount, path)
	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully replaced %d occurrence(s) of '%s' in %s", replacedCount, searchString, path),
			},
		},
	}, nil
}