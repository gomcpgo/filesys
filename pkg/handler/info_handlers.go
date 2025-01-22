package handler

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gomcpgo/mcp/pkg/protocol"
)

func (h *FileSystemHandler) handleGetFileInfo(args map[string]interface{}) (*protocol.CallToolResponse, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path must be a string")
	}

	if !h.isPathAllowed(path) {
		return nil, fmt.Errorf("access to path is not allowed: %s", path)
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	mode := info.Mode()
	fileType := "file"
	if mode.IsDir() {
		fileType = "directory"
	}

	fileInfo := map[string]interface{}{
		"name":         info.Name(),
		"size":        info.Size(),
		"type":         fileType,
		"mode":         mode.String(),
		"permissions":  fmt.Sprintf("%o", mode.Perm()),
		"modTime":      info.ModTime().Format(time.RFC3339),
	}

	// Convert to formatted string
	var details []string
	details = append(details, fmt.Sprintf("Name: %s", fileInfo["name"]))
	details = append(details, fmt.Sprintf("Size: %d bytes", fileInfo["size"]))
	details = append(details, fmt.Sprintf("Type: %s", fileInfo["type"]))
	details = append(details, fmt.Sprintf("Mode: %s", fileInfo["mode"]))
	details = append(details, fmt.Sprintf("Permissions: %s", fileInfo["permissions"]))
	details = append(details, fmt.Sprintf("Last Modified: %s", fileInfo["modTime"]))

	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: strings.Join(details, "\n"),
			},
		},
	}, nil
}

func (h *FileSystemHandler) handleSearchFiles(args map[string]interface{}) (*protocol.CallToolResponse, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path must be a string")
	}

	pattern, ok := args["pattern"].(string)
	if !ok {
		return nil, fmt.Errorf("pattern must be a string")
	}

	if !h.isPathAllowed(path) {
		return nil, fmt.Errorf("access to path is not allowed: %s", path)
	}

	var matches []string
	err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.Contains(strings.ToLower(info.Name()), strings.ToLower(pattern)) {
			matches = append(matches, path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to search files: %w", err)
	}

	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: strings.Join(matches, "\n"),
			},
		},
	}, nil
}