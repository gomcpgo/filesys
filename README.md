# Filesystem MCP Server

A secure, single-binary MCP server for filesystem operations. No runtime dependencies — download, configure allowed directories, and connect to any MCP client.

Tested extensively with Claude Desktop and Claude Code across real-world coding workflows including file editing, codebase search, refactoring, and multi-file batch operations.

## Why this server?

- **Single binary** — no Node.js, Python, or other runtime needed. Download and run
- **Tested with real AI workflows** — battle-tested with Claude Desktop and Claude Code for day-to-day coding tasks
- **18 tools** — goes beyond basic read/write with regex search, pattern-based replacement, auto-indented code insertion, and batch operations
- **Dry-run preview** — preview changes before applying them for replacement and insertion tools
- **Secure by default** — sandboxed to configured directories with symlink attack prevention and path traversal protection
- **Detailed error messages** — when access is denied, errors explain why and suggest fixes

## Installation

### Download a release binary

Download the latest binary for your platform from the [Releases](../../releases) page:

| Platform | Binary |
|----------|--------|
| macOS (Apple Silicon) | `filesystem-mcp-darwin-arm64` |
| macOS (Intel) | `filesystem-mcp-darwin-amd64` |
| Linux (x86_64) | `filesystem-mcp-linux-amd64` |
| Linux (ARM64) | `filesystem-mcp-linux-arm64` |
| Windows | `filesystem-mcp-windows-amd64.exe` |

Make it executable (macOS/Linux):

```bash
chmod +x filesystem-mcp-darwin-arm64
```

### Build from source

```bash
go build -ldflags="-s -w" -o bin/filesystem-mcp ./cmd
```

## Configuration

Set allowed directories using the environment variable:

```bash
export MCP_ALLOWED_DIRS="/path1,/path2,/path with spaces/dir3"
```

## Tools

### Reading

- **`read_file`** — Read a single file, with optional `start_line`/`end_line` for partial reads
- **`read_multiple_files`** — Read multiple files simultaneously in one call
- **`search_in_files`** — Recursive regex search across files. Returns file paths, line numbers, and matched text. Skips binary files automatically. Params: `path`, `pattern`, `file_extensions`, `max_results`, `case_sensitive`

### Writing

- **`write_file`** — Create or overwrite a file. Auto-creates parent directories
- **`append_to_file`** — Add content to end of file. Creates file if it doesn't exist
- **`prepend_to_file`** — Add content to beginning of file. Creates file if it doesn't exist

### Text Replacement

All replacement tools support `dry_run` to preview changes without applying them.

- **`replace_in_file`** — Replace exact string occurrences in a file. Params: `path`, `search`, `replace`, `occurrence` (0=all), `dry_run`
- **`replace_in_file_regex`** — Replace regex pattern matches with capture group support (`$1`, `$2`). Params: `path`, `pattern`, `replace`, `occurrence`, `case_sensitive`, `dry_run`
- **`replace_in_files`** — Batch replace a string across multiple files. Validates all paths before applying. Params: `paths`, `search`, `replace`, `dry_run`

### Regex-Based Insertion

All insertion tools support `dry_run` and `autoIndent` (match surrounding indentation).

- **`insert_after_regex`** — Insert content after a regex pattern match. Params: `path`, `pattern`, `content`, `occurrence` (0=all, default 1), `autoIndent`, `dry_run`
- **`insert_before_regex`** — Insert content before a regex pattern match. Same params as above

### Line Copying

- **`copy_lines`** — Copy a line range from source to destination file directly on disk (no context overhead). Params: `source_path`, `destination_path`, `start_line`, `end_line`, `append`

### Directory Operations

- **`list_directory`** — List directory contents with filtering by pattern, file type, recursion depth, hidden files, and metadata. Params: `path`, `pattern`, `file_type`, `recursive`, `max_depth`, `max_results`, `include_hidden`, `include_metadata`
- **`create_directory`** — Create directory and parents (idempotent)
- **`list_allowed_directories`** — Show accessible directories

### File Management

- **`move_file`** — Move or rename files and directories
- **`get_file_info`** — Get file metadata (size, permissions, modification time)

## Usage with Claude Desktop

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "/path/to/filesystem-mcp-darwin-arm64",
      "env": {
        "MCP_ALLOWED_DIRS": "/path1,/path2,/path with spaces/dir3"
      }
    }
  }
}
```

## Security

The server implements defense-in-depth security to prevent unauthorized file access.

### Path Validation
- **Symbolic link resolution**: All paths resolved to canonical form via `filepath.EvalSymlinks()` before validation
- **Path traversal prevention**: Blocks `../` escape attempts
- **Prefix matching protection**: Validates with path separators to prevent `/allowed` matching `/allowed_attacker`

### Symbolic Link Handling
- Symlinks within allowed directories are permitted if their target is also within allowed directories
- Symlinks pointing outside allowed directories are blocked
- Broken symlinks are rejected
- Allowed directories themselves may be symbolic links (resolved during initialization)

### Write Operation Security
- Parent directory chain is validated for new file creation
- Path resolution and validation occur atomically

### Security Logging
- All blocked access attempts are logged with `SECURITY:` prefix
- Logs include both the requested path and its canonical resolution

### Best Practices
- Configure `MCP_ALLOWED_DIRS` with the minimum necessary directories
- Use absolute paths for allowed directories
- Monitor logs for `SECURITY:` messages

## License

MIT License
