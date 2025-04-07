package handler

import (
	"context"
	"encoding/json"

	"github.com/gomcpgo/mcp/pkg/protocol"
)

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
			Description: "Recursively search for files and directories matching a pattern. " +
				"The search looks for the pattern anywhere in the file or directory names.",
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
					},
					"case_sensitive": {
						"type": "boolean",
						"description": "Whether to perform a case-sensitive search (default: false)"
					},
					"max_depth": {
						"type": "integer",
						"description": "Maximum depth for recursive search (-1 = unlimited, 0 = current directory only, default: -1)"
					},
					"match_path": {
						"type": "boolean",
						"description": "Whether to search in full paths instead of just filenames (default: false)"
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
		{
			Name:        "append_to_file",
			Description: "Add content to the end of a file. If the file doesn't exist, it will be created.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "Path to the file"
					},
					"content": {
						"type": "string",
						"description": "Content to append to the file"
					}
				},
				"required": ["path", "content"]
			}`),
		},
		{
			Name:        "prepend_to_file",
			Description: "Add content to the beginning of a file. If the file doesn't exist, it will be created.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "Path to the file"
					},
					"content": {
						"type": "string",
						"description": "Content to prepend to the file"
					}
				},
				"required": ["path", "content"]
			}`),
		},
		{
			Name:        "replace_in_file",
			Description: "Replace occurrences of a string in a file with new content.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "Path to the file"
					},
					"search": {
						"type": "string",
						"description": "String to search for"
					},
					"replace": {
						"type": "string",
						"description": "String to replace with"
					},
					"occurrence": {
						"type": "integer",
						"description": "Which occurrence to replace (0 means all, default is all)",
						"minimum": 0
					}
				},
				"required": ["path", "search", "replace"]
			}`),
		},
		{
			Name:        "insert_after_string",
			Description: "Insert content after a specific occurrence of a string in a file.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "Path to the file"
					},
					"search": {
						"type": "string",
						"description": "String to search for"
					},
					"content": {
						"type": "string",
						"description": "Content to insert"
					},
					"occurrence": {
						"type": "integer",
						"description": "Which occurrence to insert after (default is 1, the first occurrence)",
						"minimum": 1
					}
				},
				"required": ["path", "search", "content"]
			}`),
		},
		{
			Name:        "insert_before_string",
			Description: "Insert content before a specific occurrence of a string in a file.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "Path to the file"
					},
					"search": {
						"type": "string",
						"description": "String to search for"
					},
					"content": {
						"type": "string",
						"description": "Content to insert"
					},
					"occurrence": {
						"type": "integer",
						"description": "Which occurrence to insert before (default is 1, the first occurrence)",
						"minimum": 1
					}
				},
				"required": ["path", "search", "content"]
			}`),
		},
	}
	return &protocol.ListToolsResponse{Tools: tools}, nil
}
