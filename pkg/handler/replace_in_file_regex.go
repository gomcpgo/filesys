package handler

import (
	"fmt"
	"log"
	"os"
	"github.com/gomcpgo/filesys/pkg/search"
	"github.com/gomcpgo/mcp/pkg/protocol"
)

// handleReplaceInFileRegex replaces content that matches a regex pattern in a file
func (h *FileSystemHandler) handleReplaceInFileRegex(args map[string]interface{}) (*protocol.CallToolResponse, error) {
	path, ok := args["path"].(string)
	if !ok {
		log.Printf("ERROR: replace_in_file_regex - invalid path type: %T", args["path"])
		return nil, fmt.Errorf("path must be a string")
	}
	
	pattern, ok := args["pattern"].(string)
	if !ok {
		log.Printf("ERROR: replace_in_file_regex - invalid pattern type: %T", args["pattern"])
		return nil, fmt.Errorf("regex pattern must be a string")
	}
	
	replaceString, ok := args["replace"].(string)
	if !ok {
		log.Printf("ERROR: replace_in_file_regex - invalid replace string type: %T", args["replace"])
		return nil, fmt.Errorf("replace string must be a string")
	}
	
	// Optional parameter for which occurrence to replace
	occurrence := 0 // 0 means replace all
	if occurrenceVal, ok := args["occurrence"].(float64); ok {
		occurrence = int(occurrenceVal)
	}
	
	// Optional parameter for case sensitivity
	caseSensitive := true // default is case sensitive
	if caseVal, ok := args["case_sensitive"].(bool); ok {
		caseSensitive = caseVal
	}
	
	log.Printf("replace_in_file_regex - attempting to replace pattern '%s' with '%s' in %s", 
		pattern, replaceString, path)
	
	if !h.isPathAllowed(path) {
		log.Printf("ERROR: replace_in_file_regex - access denied to path: %s", path)
		return nil, fmt.Errorf("access to path is not allowed: %s", path)
	}
	
	// Use our regex replacement function
	newContent, replacementCount, err := search.ReplaceWithRegex(path, pattern, replaceString, occurrence, caseSensitive)
	if err != nil {
		log.Printf("ERROR: replace_in_file_regex - %v", err)
		return nil, err
	}
	
	if replacementCount == 0 {
		log.Printf("replace_in_file_regex - regex pattern '%s' not found in %s", pattern, path)
		return &protocol.CallToolResponse{
			Content: []protocol.ToolContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Pattern '%s' not found in %s", pattern, path),
				},
			},
		}, nil
	}
	
	// Write the new content back to the file
	err = os.WriteFile(path, []byte(newContent), 0644)
	if err != nil {
		log.Printf("ERROR: replace_in_file_regex - failed to write to %s: %v", path, err)
		return nil, fmt.Errorf("failed to write file: %w", err)
	}
	
	if occurrence == 0 {
		log.Printf("replace_in_file_regex - successfully replaced %d occurrence(s) of pattern '%s' in %s", 
			replacementCount, pattern, path)
		return &protocol.CallToolResponse{
			Content: []protocol.ToolContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Successfully replaced %d occurrence(s) of pattern '%s' in %s", 
						replacementCount, pattern, path),
				},
			},
		}, nil
	} else {
		log.Printf("replace_in_file_regex - successfully replaced occurrence %d of pattern '%s' in %s", 
			occurrence, pattern, path)
		return &protocol.CallToolResponse{
			Content: []protocol.ToolContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Successfully replaced occurrence %d of pattern '%s' in %s", 
						occurrence, pattern, path),
				},
			},
		}, nil
	}
}
