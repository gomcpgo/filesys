package handler

import (
	"fmt"
	"log"
	"os"

	"github.com/gomcpgo/filesys/pkg/search"
	"github.com/gomcpgo/mcp/pkg/protocol"
)

// handleInsertAfterLine inserts content after a specific line number
func (h *FileSystemHandler) handleInsertAfterLine(args map[string]interface{}) (*protocol.CallToolResponse, error) {
	path, ok := args["path"].(string)
	if !ok {
		log.Printf("ERROR: insert_after_line - invalid path type: %T", args["path"])
		return nil, fmt.Errorf("path must be a string")
	}

	lineNumber, ok := args["line_number"].(float64)
	if !ok {
		log.Printf("ERROR: insert_after_line - invalid line_number type: %T", args["line_number"])
		return nil, fmt.Errorf("line_number must be an integer")
	}

	contentToInsert, ok := args["content"].(string)
	if !ok {
		log.Printf("ERROR: insert_after_line - invalid content type: %T", args["content"])
		return nil, fmt.Errorf("content must be a string")
	}

	// Check for autoIndent parameter (defaults to false)
	autoIndent := false
	if autoIndentVal, ok := args["autoIndent"].(bool); ok {
		autoIndent = autoIndentVal
	}

	// Check for dry_run parameter (defaults to false)
	dryRun := false
	if dryRunVal, ok := args["dry_run"].(bool); ok {
		dryRun = dryRunVal
	}

	log.Printf("insert_after_line - attempting to insert after line %d in %s (autoIndent: %v, dry_run: %v)",
		int(lineNumber), path, autoIndent, dryRun)

	if !h.isPathAllowed(path) {
		log.Printf("ERROR: insert_after_line - access denied to path: %s", path)
		return nil, NewAccessDeniedError(path)
	}

	// Use the search package to insert content after line
	newContent, err := search.InsertAfterLine(path, int(lineNumber), contentToInsert, autoIndent)
	if err != nil {
		log.Printf("ERROR: insert_after_line - %v", err)
		return nil, err
	}

	// Dry run mode - return preview without modifying file
	if dryRun {
		log.Printf("insert_after_line - dry run: would insert content after line %d in %s", int(lineNumber), path)
		return &protocol.CallToolResponse{
			Content: []protocol.ToolContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Preview (dry run - no changes applied):\n\nWould insert %d character(s) after line %d\n\nResulting content:\n%s",
						len(contentToInsert), int(lineNumber), newContent),
				},
			},
		}, nil
	}

	// Write the new content back to the file
	err = os.WriteFile(path, []byte(newContent), 0644)
	if err != nil {
		log.Printf("ERROR: insert_after_line - failed to write to %s: %v", path, err)
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	log.Printf("insert_after_line - successfully inserted content after line %d in %s", int(lineNumber), path)
	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: newContent,
			},
		},
	}, nil
}

// handleInsertBeforeLine inserts content before a specific line number
func (h *FileSystemHandler) handleInsertBeforeLine(args map[string]interface{}) (*protocol.CallToolResponse, error) {
	path, ok := args["path"].(string)
	if !ok {
		log.Printf("ERROR: insert_before_line - invalid path type: %T", args["path"])
		return nil, fmt.Errorf("path must be a string")
	}

	lineNumber, ok := args["line_number"].(float64)
	if !ok {
		log.Printf("ERROR: insert_before_line - invalid line_number type: %T", args["line_number"])
		return nil, fmt.Errorf("line_number must be an integer")
	}

	contentToInsert, ok := args["content"].(string)
	if !ok {
		log.Printf("ERROR: insert_before_line - invalid content type: %T", args["content"])
		return nil, fmt.Errorf("content must be a string")
	}

	// Check for autoIndent parameter (defaults to false)
	autoIndent := false
	if autoIndentVal, ok := args["autoIndent"].(bool); ok {
		autoIndent = autoIndentVal
	}

	// Check for dry_run parameter (defaults to false)
	dryRun := false
	if dryRunVal, ok := args["dry_run"].(bool); ok {
		dryRun = dryRunVal
	}

	log.Printf("insert_before_line - attempting to insert before line %d in %s (autoIndent: %v, dry_run: %v)",
		int(lineNumber), path, autoIndent, dryRun)

	if !h.isPathAllowed(path) {
		log.Printf("ERROR: insert_before_line - access denied to path: %s", path)
		return nil, NewAccessDeniedError(path)
	}

	// Use the search package to insert content before line
	newContent, err := search.InsertBeforeLine(path, int(lineNumber), contentToInsert, autoIndent)
	if err != nil {
		log.Printf("ERROR: insert_before_line - %v", err)
		return nil, err
	}

	// Dry run mode - return preview without modifying file
	if dryRun {
		log.Printf("insert_before_line - dry run: would insert content before line %d in %s", int(lineNumber), path)
		return &protocol.CallToolResponse{
			Content: []protocol.ToolContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Preview (dry run - no changes applied):\n\nWould insert %d character(s) before line %d\n\nResulting content:\n%s",
						len(contentToInsert), int(lineNumber), newContent),
				},
			},
		}, nil
	}

	// Write the new content back to the file
	err = os.WriteFile(path, []byte(newContent), 0644)
	if err != nil {
		log.Printf("ERROR: insert_before_line - failed to write to %s: %v", path, err)
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	log.Printf("insert_before_line - successfully inserted content before line %d in %s", int(lineNumber), path)
	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: newContent,
			},
		},
	}, nil
}
