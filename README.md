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

The filesystem MCP server implements comprehensive security measures to prevent unauthorized file access:

### Path Validation
- **Symbolic link resolution**: All paths are resolved to their canonical form using `filepath.EvalSymlinks()` before validation
- **Canonical path checking**: Paths are validated against allowed directories only after resolving all symbolic links
- **Path traversal prevention**: Attempts to escape allowed directories using `../` or similar techniques are blocked
- **Prefix matching protection**: Directory prefixes are validated with path separators to prevent `/allowed` from matching `/allowed_attacker`

### Symbolic Link Handling
- **Legitimate symlinks**: Symbolic links within allowed directories are permitted (if their target is also within allowed directories)
- **Attack prevention**: Symbolic links pointing outside allowed directories are automatically blocked
- **Broken symlinks**: Broken symbolic links (pointing to non-existent targets) are rejected for security
- **Allowed directory symlinks**: Allowed directories themselves may be symbolic links (resolved during initialization)

### Write Operation Security
- **Non-existent paths**: When creating new files or directories, the parent directory chain is validated
- **Parent validation**: Only paths whose parent directories are within allowed areas can be created
- **Atomic validation**: Path resolution and validation occur atomically to prevent race conditions

### Attack Prevention
The server protects against:
- Symlink-based directory traversal attacks
- Path traversal with `../` sequences
- Broken symlink exploitation
- Prefix matching attacks (`/allowed` vs `/allowed_attacker`)
- Nested symlink chains escaping allowed directories

### Security Logging
- All blocked access attempts are logged with "SECURITY:" prefix
- Logs include both the requested path and its canonical resolution
- Helps detect and investigate potential attack attempts

### Best Practices
- Configure `MCP_ALLOWED_DIRS` with the minimum necessary directories
- Use absolute paths for allowed directories
- Monitor logs for "SECURITY:" messages indicating blocked access attempts
- Regularly review allowed directory configurations

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