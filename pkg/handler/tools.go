package handler

import (
	"context"
	"encoding/json"

	"github.com/gomcpgo/mcp/pkg/protocol"
)

func (h *FileSystemHandler) ListTools(ctx context.Context) (*protocol.ListToolsResponse, error) {
	tools := []protocol.Tool{
		{
			Name:        "search_in_files",
			Description: "Search for text content inside files using regular expressions. This tool searches through file contents recursively in a directory and returns matches with file paths, line numbers, and the matched text lines. Only searches text files and skips binary files automatically. Use this for finding code, text strings, or patterns across multiple files.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "Directory path to search in. Will search recursively through all subdirectories."
					},
					"pattern": {
						"type": "string",
						"description": "Regular expression pattern to search for within file contents. Supports full regex syntax including anchors (^$), character classes ([a-z]), quantifiers (*, +, ?), and more."
					},
					"file_extensions": {
						"type": "array",
						"items": {
							"type": "string"
						},
						"description": "File extensions to include (e.g., [\".txt\", \".go\"]). Each extension should include the dot. If empty, searches common text file types including .txt, .go, .js, .html, .css, .md, .json, .yml, .yaml, .xml."
					},
					"max_results": {
						"type": "integer",
						"description": "Maximum number of matches to return (default 100)",
						"default": 100
					},
					"max_file_searches": {
						"type": "integer",
						"description": "Maximum number of files to examine (default 100)",
						"default": 100
					},
					"case_sensitive": {
						"type": "boolean",
						"description": "Whether the search is case sensitive. If false, case will be ignored when matching (default true)",
						"default": true
					}
				},
				"required": ["path", "pattern"]
			}`),
		},
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
		{
			Name:        "insert_after_regex",
			Description: "Insert content after a specific occurrence of a regular expression pattern in a file.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "Path to the file"
					},
					"pattern": {
						"type": "string",
						"description": "Regular expression pattern to search for. Supports full regex syntax including anchors (^$), character classes ([a-z]), quantifiers (*, +, ?), groups, and more."
					},
					"content": {
						"type": "string",
						"description": "Content to insert"
					},
					"occurrence": {
						"type": "integer",
						"description": "Which occurrence to insert after (0 means all occurrences, 1+ for specific occurrence, default is 1)",
						"minimum": 0
					}
				},
				"required": ["path", "pattern", "content"]
			}`),
		},
		{
			Name:        "insert_before_regex",
			Description: "Insert content before a specific occurrence of a regular expression pattern in a file.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "Path to the file"
					},
					"pattern": {
						"type": "string",
						"description": "Regular expression pattern to search for. Supports full regex syntax including anchors (^$), character classes ([a-z]), quantifiers (*, +, ?), groups, and more."
					},
					"content": {
						"type": "string",
						"description": "Content to insert"
					},
					"occurrence": {
						"type": "integer",
						"description": "Which occurrence to insert before (0 means all occurrences, 1+ for specific occurrence, default is 1)",
						"minimum": 0
					}
				},
				"required": ["path", "pattern", "content"]
			}`),
		},
	}
	return &protocol.ListToolsResponse{Tools: tools}, nil
}
