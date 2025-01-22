package handler

import (
	"context"
	"encoding/json"

	"github.com/gomcpgo/mcp/pkg/protocol"
)

// ListTools returns the available filesystem tools
func (h *FileSystemHandler) ListTools(ctx context.Context) (*protocol.ListToolsResponse, error) {
	tools := []protocol.Tool{
		{
			Name: "read_file",
			Description: "Read the complete contents of a file from the file system. " +
				"Handles various text encodings and provides detailed error messages " +
				"if the file cannot be read. Use this tool when you need to examine " +
				"the contents of a single file. Only works within allowed directories.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "Path to the file to read"
					}
				},
				"required": ["path"]
			}`),
		},
		{
			Name: "update_file_section",
			Description: "Update a specific section of a file by replacing content between given line numbers. " +
				"Ideal for modifying specific functions or blocks of code while preserving the rest of the file unchanged. " +
				"Use this instead of complete file rewrites when only a small section needs to be modified.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "Path to the file to update"
					},
					"startLine": {
						"type": "integer",
						"description": "Starting line number (1-based)",
						"minimum": 1
					},
					"endLine": {
						"type": "integer",
						"description": "Ending line number (1-based, inclusive)",
						"minimum": 1
					},
					"newContent": {
						"type": "string",
						"description": "New content to insert between start and end lines"
					}
				},
				"required": ["path", "startLine", "endLine", "newContent"]
			}`),
		},
		{
			Name:        "read_multiple_files",
			Description: "Read the contents of multiple files simultaneously.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"paths": {
						"type": "array",
						"items": {
							"type": "string"
						},
						"description": "Array of file paths to read"
					}
				},
				"required": ["paths"]
			}`),
		},
		{
			Name:        "write_file",
			Description: "Create a new file or completely overwrite an existing file with new content.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "Path where the file should be written"
					},
					"content": {
						"type": "string",
						"description": "Content to write to the file"
					}
				},
				"required": ["path", "content"]
			}`),
		},
		{
			Name:        "create_directory",
			Description: "Create a new directory or ensure a directory exists.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "Path of the directory to create"
					}
				},
				"required": ["path"]
			}`),
		},
		{
			Name:        "list_directory",
			Description: "Get a detailed listing of all files and directories in a specified path.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "Path of the directory to list"
					}
				},
				"required": ["path"]
			}`),
		},
		{
			Name:        "move_file",
			Description: "Move or rename files and directories.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"source": {
						"type": "string",
						"description": "Source path"
					},
					"destination": {
						"type": "string",
						"description": "Destination path"
					}
				},
				"required": ["source", "destination"]
			}`),
		},
		{
			Name:        "search_files",
			Description: "Recursively search for files and directories matching a pattern.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "Path to search in"
					},
					"pattern": {
						"type": "string",
						"description": "Search pattern to match"
					}
				},
				"required": ["path", "pattern"]
			}`),
		},
		{
			Name:        "get_file_info",
			Description: "Retrieve detailed metadata about a file or directory.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "Path to get information about"
					}
				},
				"required": ["path"]
			}`),
		},
		{
			Name:        "list_allowed_directories",
			Description: "Returns the list of directories that this server is allowed to access.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {},
				"required": []
			}`),
		},
	}

	return &protocol.ListToolsResponse{Tools: tools}, nil
}
