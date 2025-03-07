package handler

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gomcpgo/mcp/pkg/protocol"
)

func (h *FileSystemHandler) handleGetFileInfo(args map[string]interface{}) (*protocol.CallToolResponse, error) {
	path, ok := args["path"].(string)
	if !ok {
		log.Printf("ERROR: get_file_info - invalid path type: %T", args["path"])
		return nil, fmt.Errorf("path must be a string")
	}

	log.Printf("get_file_info - retrieving info for path: %s", path)
	if !h.isPathAllowed(path) {
		log.Printf("ERROR: get_file_info - access denied to path: %s", path)
		return nil, fmt.Errorf("access to path is not allowed: %s", path)
	}

	info, err := os.Stat(path)
	if err != nil {
		log.Printf("ERROR: get_file_info - failed to get file info for %s: %v", path, err)
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	mode := info.Mode()
	fileType := "file"
	if mode.IsDir() {
		fileType = "directory"
	}

	fileInfo := map[string]interface{}{
		"name":        info.Name(),
		"size":        info.Size(),
		"type":        fileType,
		"mode":        mode.String(),
		"permissions": fmt.Sprintf("%o", mode.Perm()),
		"modTime":     info.ModTime().Format(time.RFC3339),
	}

	var details []string
	details = append(details, fmt.Sprintf("Name: %s", fileInfo["name"]))
	details = append(details, fmt.Sprintf("Size: %d bytes", fileInfo["size"]))
	details = append(details, fmt.Sprintf("Type: %s", fileInfo["type"]))
	details = append(details, fmt.Sprintf("Mode: %s", fileInfo["mode"]))
	details = append(details, fmt.Sprintf("Permissions: %s", fileInfo["permissions"]))
	details = append(details, fmt.Sprintf("Last Modified: %s", fileInfo["modTime"]))

	log.Printf("get_file_info - successfully retrieved info for %s (%s, %d bytes)",
		path, fileType, info.Size())
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
		log.Printf("ERROR: search_files - invalid path type: %T", args["path"])
		return nil, fmt.Errorf("path must be a string")
	}
	pattern, ok := args["pattern"].(string)
	if !ok {
		log.Printf("ERROR: search_files - invalid pattern type: %T", args["pattern"])
		return nil, fmt.Errorf("pattern must be a string")
	}

	log.Printf("search_files - searching in %s for pattern: %s", path, pattern)
	if !h.isPathAllowed(path) {
		log.Printf("ERROR: search_files - access denied to path: %s", path)
		return nil, fmt.Errorf("access to path is not allowed: %s", path)
	}

	var matches []string
	err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			log.Printf("WARNING: search_files - error accessing path %s: %v", path, err)
			return err
		}
		if strings.Contains(strings.ToLower(info.Name()), strings.ToLower(pattern)) {
			matches = append(matches, path)
			log.Printf("search_files - found match: %s", path)
		}
		return nil
	})

	if err != nil {
		log.Printf("ERROR: search_files - failed to search files in %s: %v", path, err)
		return nil, fmt.Errorf("failed to search files: %w", err)
	}

	log.Printf("search_files - found %d matches for pattern %q in %s", len(matches), pattern, path)
	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: strings.Join(matches, "\n"),
			},
		},
	}, nil
}
