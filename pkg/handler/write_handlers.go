package handler

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gomcpgo/mcp/pkg/protocol"
)

func (h *FileSystemHandler) handleWriteFile(args map[string]interface{}) (*protocol.CallToolResponse, error) {
	path, ok := args["path"].(string)
	if !ok {
		log.Printf("ERROR: write_file - invalid path type: %T", args["path"])
		return nil, fmt.Errorf("path must be a string")
	}
	content, ok := args["content"].(string)
	if !ok {
		log.Printf("ERROR: write_file - invalid content type: %T", args["content"])
		return nil, fmt.Errorf("content must be a string")
	}

	log.Printf("write_file - attempting to write %d bytes to: %s", len(content), path)
	if !h.isPathAllowed(path) {
		log.Printf("ERROR: write_file - access denied to path: %s", path)
		return nil, fmt.Errorf("access to path is not allowed: %s", path)
	}

	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		log.Printf("ERROR: write_file - failed to write to %s: %v", path, err)
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	log.Printf("write_file - successfully wrote %d bytes to %s", len(content), path)
	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully wrote file: %s", path),
			},
		},
	}, nil
}

func (h *FileSystemHandler) handleMoveFile(args map[string]interface{}) (*protocol.CallToolResponse, error) {
	source, ok := args["source"].(string)
	if !ok {
		log.Printf("ERROR: move_file - invalid source path type: %T", args["source"])
		return nil, fmt.Errorf("source must be a string")
	}
	destination, ok := args["destination"].(string)
	if !ok {
		log.Printf("ERROR: move_file - invalid destination path type: %T", args["destination"])
		return nil, fmt.Errorf("destination must be a string")
	}

	log.Printf("move_file - attempting to move %s to %s", source, destination)
	if !h.isPathAllowed(source) || !h.isPathAllowed(destination) {
		log.Printf("ERROR: move_file - access denied to path(s): source=%s, destination=%s", source, destination)
		return nil, fmt.Errorf("access to path is not allowed")
	}

	err := os.Rename(source, destination)
	if err != nil {
		log.Printf("ERROR: move_file - failed to move %s to %s: %v", source, destination, err)
		return nil, fmt.Errorf("failed to move file: %w", err)
	}

	log.Printf("move_file - successfully moved %s to %s", source, destination)
	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully moved %s to %s", source, destination),
			},
		},
	}, nil
}

func (h *FileSystemHandler) handleUpdateFileSection(args map[string]interface{}) (*protocol.CallToolResponse, error) {
	path, ok := args["path"].(string)
	if !ok {
		log.Printf("ERROR: update_file_section - invalid path type: %T", args["path"])
		return nil, fmt.Errorf("path must be a string")
	}
	startLine, ok := args["startLine"].(float64)
	if !ok {
		log.Printf("ERROR: update_file_section - invalid startLine type: %T", args["startLine"])
		return nil, fmt.Errorf("startLine must be a number")
	}
	endLine, ok := args["endLine"].(float64)
	if !ok {
		log.Printf("ERROR: update_file_section - invalid endLine type: %T", args["endLine"])
		return nil, fmt.Errorf("endLine must be a number")
	}
	newContent, ok := args["newContent"].(string)
	if !ok {
		log.Printf("ERROR: update_file_section - invalid newContent type: %T", args["newContent"])
		return nil, fmt.Errorf("newContent must be a string")
	}

	start := int(startLine)
	end := int(endLine)
	if start < 1 {
		log.Printf("ERROR: update_file_section - invalid startLine: %d (must be >= 1)", start)
		return nil, fmt.Errorf("startLine must be >= 1")
	}
	if end < start {
		log.Printf("ERROR: update_file_section - invalid endLine: %d (must be >= startLine %d)", end, start)
		return nil, fmt.Errorf("endLine must be >= startLine")
	}

	log.Printf("update_file_section - attempting to update lines %d-%d in %s", start, end, path)
	if !h.isPathAllowed(path) {
		log.Printf("ERROR: update_file_section - access denied to path: %s", path)
		return nil, fmt.Errorf("access to path is not allowed: %s", path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		log.Printf("ERROR: update_file_section - failed to read file %s: %v", path, err)
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	if start > len(lines) {
		log.Printf("ERROR: update_file_section - startLine %d is beyond end of file (%d lines)", start, len(lines))
		return nil, fmt.Errorf("startLine %d is beyond end of file (%d lines)", start, len(lines))
	}
	if end > len(lines) {
		log.Printf("ERROR: update_file_section - endLine %d is beyond end of file (%d lines)", end, len(lines))
		return nil, fmt.Errorf("endLine %d is beyond end of file (%d lines)", end, len(lines))
	}

	newLines := make([]string, 0, len(lines)-(end-start+1)+1)
	newLines = append(newLines, lines[:start-1]...)
	newLines = append(newLines, strings.Split(newContent, "\n")...)
	newLines = append(newLines, lines[end:]...)

	log.Printf("update_file_section - writing updated content to %s", path)
	err = os.WriteFile(path, []byte(strings.Join(newLines, "\n")), 0644)
	if err != nil {
		log.Printf("ERROR: update_file_section - failed to write updated content to %s: %v", path, err)
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	log.Printf("update_file_section - successfully updated lines %d-%d in %s", start, end, path)
	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully updated lines %d-%d in %s", start, end, path),
			},
		},
	}, nil
}
