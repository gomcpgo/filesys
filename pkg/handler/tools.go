package handler

import (
	"context"
	"encoding/json"

	"github.com/gomcpgo/mcp/pkg/protocol"
)

// ListTools provides a list of all available tools in the filesystem handler
// This function is called by the MCP server to discover available tools
// Each tool has a name, description, and input schema defined in JSON format
// Added via insert_before_regex tool as a demo

func (h *FileSystemHandler) ListTools(ctx context.Context) (*protocol.ListToolsResponse, error) {
	tools := []protocol.Tool{
		{
			// Tool Definition
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
			// Tool Definition
			Name: "read_file",
			Description: "Read the contents of a file from the file system, with support for partial reading by line range. " +
				"For large files, you can specify start and end lines to read only a portion of the file. " +
				"Returns the exact file content as the primary response (preserving all formatting and whitespace). " +
				"For partial reads or truncated content, additional metadata is provided as a secondary response. " +
				"Small files are read efficiently in a single operation, while larger files use optimized line-by-line reading. " +
				"Only works within allowed directories.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "Path to the file to read"
					},
					"start_line": {
						"type": "integer",
						"description": "Line number to start reading from (1-indexed, optional). If not specified, starts from the first line.",
						"minimum": 1
					},
					"end_line": {
						"type": "integer",
						"description": "Line number to end reading at, inclusive (optional). If not specified, reads to the end of file.",
						"minimum": 1
					}
				},
				"required": ["path"]
			}`),
		},
		{
			// Tool Definition
			Name:        "read_multiple_files",
			Description: "Read the contents of multiple files simultaneously using optimized file reading. " +
				"Returns exact file content preserving all formatting and whitespace. " +
				"Automatically handles large files and provides truncation warnings when necessary. " +
				"Only works within allowed directories.",
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
			// Tool Definition
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
			// Tool Definition
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
			// Tool Definition
			Name:        "list_directory",
			Description: "Get a detailed listing of all files and directories in a specified path, with advanced filtering and recursion options.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "Path of the directory to list"
					},
					"pattern": {
						"type": "string",
						"description": "Regular expression pattern to filter entries by name (optional)"
					},
					"file_type": {
						"type": "string",
						"description": "Filter by file type: 'file' for regular files only, 'dir' for directories only, or file extension like '.txt' (optional)"
					},
					"recursive": {
						"type": "boolean",
						"description": "Whether to list contents recursively through subdirectories (default: false)",
						"default": false
					},
					"max_depth": {
						"type": "integer",
						"description": "Maximum recursion depth when recursive=true (0 for unlimited, default: 0)",
						"default": 0,
						"minimum": 0
					},
					"max_results": {
						"type": "integer",
						"description": "Maximum number of entries to return (default: 100)",
						"default": 100,
						"minimum": 1
					},
					"include_hidden": {
						"type": "boolean",
						"description": "Whether to include hidden files and directories (starting with '.') in the results (default: false)",
						"default": false
					},
					"include_metadata": {
						"type": "boolean",
						"description": "Whether to include detailed metadata for each entry (size, modification time, permissions) (default: true)",
						"default": true
					}
				},
				"required": ["path"]
			}`),
		},
		{
			// Tool Definition
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
			// Tool Definition
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
			// Tool Definition
			Name:        "list_allowed_directories",
			Description: "Returns the list of directories that this server is allowed to access.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {},
				"required": []
			}`),
		},
		{
			// Tool Definition
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
			// Tool Definition
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
			// Tool Definition
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
					},
					"dry_run": {
						"type": "boolean",
						"description": "Preview changes without applying them (default: false)",
						"default": false
					}
				},
				"required": ["path", "search", "replace"]
			}`),
		},
		{
			// Tool Definition
			Name:        "replace_in_file_regex",
			Description: "Replace content matching a regular expression pattern in a file. Supports capture groups in the replacement (use $1, $2, etc.).",
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
					"replace": {
						"type": "string",
						"description": "Content to replace matching patterns with. Can include capture group references ($1, $2, etc.)"
					},
					"occurrence": {
						"type": "integer",
						"description": "Which occurrence to replace (0 means all occurrences, 1+ for specific occurrence, default is all)",
						"minimum": 0
					},
					"case_sensitive": {
						"type": "boolean",
						"description": "Whether the search is case sensitive (default true)",
						"default": true
					},
					"multiline": {
						"type": "boolean",
						"description": "Enable multiline mode where . matches newlines and ^ $ match line boundaries (default: false)",
						"default": false
					},
					"dry_run": {
						"type": "boolean",
						"description": "Preview changes without applying them (default: false)",
						"default": false
					}
				},
				"required": ["path", "pattern", "replace"]
			}`),
		},
		{
			// Tool Definition
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
					},
					"autoIndent": {
						"type": "boolean",
						"description": "Automatically indent inserted content to match surrounding code (default: false)",
						"default": false
					},
					"multiline": {
						"type": "boolean",
						"description": "Enable multiline mode where . matches newlines and ^ $ match line boundaries (default: false)",
						"default": false
					},
					"dry_run": {
						"type": "boolean",
						"description": "Preview changes without applying them (default: false)",
						"default": false
					}
				},
				"required": ["path", "pattern", "content"]
			}`),
		},
		{
			// Tool Definition
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
					},
					"autoIndent": {
						"type": "boolean",
						"description": "Automatically indent inserted content to match surrounding code (default: false)",
						"default": false
					},
					"multiline": {
						"type": "boolean",
						"description": "Enable multiline mode where . matches newlines and ^ $ match line boundaries (default: false)",
						"default": false
					},
					"dry_run": {
						"type": "boolean",
						"description": "Preview changes without applying them (default: false)",
						"default": false
					}
				},
				"required": ["path", "pattern", "content"]
			}`),
		},
		{
			// Tool Definition
			Name:        "insert_after_line",
			Description: "Insert content after a specific line number in a file. More intuitive than regex for simple line-based insertions.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "Path to the file"
					},
					"line_number": {
						"type": "integer",
						"description": "Line number to insert after (1-indexed)",
						"minimum": 1
					},
					"content": {
						"type": "string",
						"description": "Content to insert"
					},
					"autoIndent": {
						"type": "boolean",
						"description": "Automatically indent inserted content to match the target line's indentation (default: false)",
						"default": false
					},
					"dry_run": {
						"type": "boolean",
						"description": "Preview changes without applying them (default: false)",
						"default": false
					}
				},
				"required": ["path", "line_number", "content"]
			}`),
		},
		{
			// Tool Definition
			Name:        "insert_before_line",
			Description: "Insert content before a specific line number in a file. More intuitive than regex for simple line-based insertions.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "Path to the file"
					},
					"line_number": {
						"type": "integer",
						"description": "Line number to insert before (1-indexed)",
						"minimum": 1
					},
					"content": {
						"type": "string",
						"description": "Content to insert"
					},
					"autoIndent": {
						"type": "boolean",
						"description": "Automatically indent inserted content to match the target line's indentation (default: false)",
						"default": false
					},
					"dry_run": {
						"type": "boolean",
						"description": "Preview changes without applying them (default: false)",
						"default": false
					}
				},
				"required": ["path", "line_number", "content"]
			}`),
		},
		{
			// Tool Definition
			Name:        "replace_in_files",
			Description: "Replace occurrences of a string across multiple files. Use search_in_files first to identify target files.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"paths": {
						"type": "array",
						"items": {
							"type": "string"
						},
						"description": "Array of file paths to modify"
					},
					"search": {
						"type": "string",
						"description": "String to search for"
					},
					"replace": {
						"type": "string",
						"description": "String to replace with"
					},
					"dry_run": {
						"type": "boolean",
						"description": "Preview changes without applying them (default: false)",
						"default": false
					}
				},
				"required": ["paths", "search", "replace"]
			}`),
		},
	}
	return &protocol.ListToolsResponse{Tools: tools}, nil
}
