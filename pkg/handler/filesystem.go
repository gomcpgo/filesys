package handler

import (
	"context"
	"fmt"

	"github.com/gomcpgo/mcp/pkg/protocol"
)

// FileSystemHandler implements the MCP handler interfaces for filesystem operations
type FileSystemHandler struct{}

// NewFileSystemHandler creates a new filesystem handler
func NewFileSystemHandler() *FileSystemHandler {
	return &FileSystemHandler{}
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
	// Removed old search_files handler in favor of search_in_files
	case "get_file_info":
		return h.handleGetFileInfo(req.Arguments)
	case "list_allowed_directories":
		return h.handleListAllowedDirectories()
	// New tools
	case "append_to_file":
		return h.handleAppendToFile(req.Arguments)
	case "prepend_to_file":
		return h.handlePrependToFile(req.Arguments)
	case "replace_in_file":
		return h.handleReplaceInFile(req.Arguments)
	case "insert_after_string":
		return h.handleInsertAfterString(req.Arguments)
	case "insert_before_string":
		return h.handleInsertBeforeString(req.Arguments)
	case "search_in_files":
		return h.handleSearchInFiles(req.Arguments)
	default:
		return nil, fmt.Errorf("unknown tool: %s", req.Name)
	}
}
