package handler

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/gomcpgo/mcp/pkg/protocol"
)

// FileSystemHandler implements the MCP handler interfaces for filesystem operations
type FileSystemHandler struct {
	allowedDirs []string
}

// NewFileSystemHandler creates a new filesystem handler with allowed directories
func NewFileSystemHandler(allowedDirs []string) *FileSystemHandler {
	// Normalize paths
	normalized := make([]string, len(allowedDirs))
	for i, dir := range allowedDirs {
		abs, err := filepath.Abs(dir)
		if err == nil {
			normalized[i] = abs
		} else {
			normalized[i] = dir
		}
	}
	return &FileSystemHandler{
		allowedDirs: normalized,
	}
}

// CallTool handles execution of filesystem tools
func (h *FileSystemHandler) CallTool(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResponse, error) {
	switch req.Name {
	case "read_file":
		return h.handleReadFile(req.Arguments)
	case "read_multiple_files":
		return h.handleReadMultipleFiles(req.Arguments)
	case "write_file":
		return h.handleWriteFile(req.Arguments)
	case "create_directory":
		return h.handleCreateDirectory(req.Arguments)
	case "list_directory":
		return h.handleListDirectory(req.Arguments)
	case "move_file":
		return h.handleMoveFile(req.Arguments)
	case "search_files":
		return h.handleSearchFiles(req.Arguments)
	case "get_file_info":
		return h.handleGetFileInfo(req.Arguments)
	case "list_allowed_directories":
		return h.handleListAllowedDirectories()
	default:
		return nil, fmt.Errorf("unknown tool: %s", req.Name)
	}
}