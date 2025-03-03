# Filesystem MCP Server

A secure Model Context Protocol (MCP) server that provides filesystem operations with controlled access to specified directories.

## Features

- Directory access controlled via environment variables
- File operations within allowed directories only
- Thread-safe caching of allowed directories
- Proper handling of paths with spaces

## Installation

```bash
go get github.com/gomcpgo/filesys
```

## Configuration

Set allowed directories using the environment variable:

```bash
export MCP_ALLOWED_DIRS="/path1,/path2,/path with spaces/dir3"
```

## Tools

### File Reading
- `read_file`: Read single file contents
- `read_multiple_files`: Read multiple files simultaneously

### File Writing
- `write_file`: Create or overwrite files

### Directory Operations
- `create_directory`: Create new directories
- `list_directory`: List directory contents
- `list_allowed_directories`: Show accessible directories

### File Management
- `move_file`: Move or rename files and directories
- `get_file_info`: Get file metadata
- `search_files`: Search files recursively with pattern matching

## Usage with Claude Desktop

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "/path/to/filesys",
      "env": {
        "MCP_ALLOWED_DIRS": "/path1,/path2,/path with spaces/dir3"
      }
    }
  }
}
```

## Tool Examples

### Reading a File
```javascript
{
    "name": "read_file",
    "arguments": {
        "path": "/allowed/path/file.txt"
    }
}
```

### Listing Directory
```javascript
{
    "name": "list_directory",
    "arguments": {
        "path": "/allowed/path"
    }
}
```

## Security

- All operations restricted to allowed directories
- Path traversal prevention
- Permission validation before operations
- Proper error handling and logging

## Building

```bash
go build -o bin/filesys cmd/main.go
```

## License

MIT License

## Contributing

Pull requests welcome. Please ensure:
- Tests pass
- New features include documentation
- Code follows project style