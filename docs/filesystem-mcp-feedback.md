# Filesystem MCP Server - Usage Feedback & Suggestions

Feedback from extensive usage during email MCP server development session.

## What Worked Well

| Tool | Strength |
|------|----------|
| `read_multiple_files` | Efficient batch reading - one call instead of multiple |
| `list_directory` | Rich metadata (sizes, dates), recursive option, filtering |
| `replace_in_file` | Simple, safe exact string matching |

## Issues Encountered

### 1. Path Access Restriction - Poor Error Message

**Problem:** When accessing `/Users/prasanth/Library/Application Support/Savant/...`:
```
MCP error -32603: access to path is not allowed
```

**Issues:**
- No indication of what paths ARE allowed
- Had to fall back to built-in `Read` tool which had broader access
- Confusing that different tools have different access

**Suggestion:** Include allowed paths in error message:
```
access to path is not allowed. Allowed directories: [/Users/prasanth/MyProjects/...]
```

### 2. No Proactive Path Guidance

`list_allowed_directories` exists but isn't obvious to call. Consider mentioning it in error messages.

## Feature Suggestions

### 1. Regex Support for Replace

Current `replace_in_file` only does exact matching. Add regex option:

```json
{
  "path": "/path/to/file.go",
  "pattern": "func (old\\w+)\\(",
  "replace": "func new$1(",
  "regex": true
}
```

### 2. Line-Based Editing

Replace specific line ranges:

```json
{
  "path": "/path/to/file.go",
  "start_line": 50,
  "end_line": 60,
  "content": "new content for these lines"
}
```

### 3. Batch Replace Across Files

Refactor same pattern in multiple files:

```json
{
  "paths": ["/path/a.go", "/path/b.go"],
  "search": "oldFunction",
  "replace": "newFunction"
}
```

### 4. Dry Run / Preview Mode

Show what would change before applying:

```json
{
  "path": "/path/to/file.go",
  "search": "oldName",
  "replace": "newName",
  "dry_run": true
}
```

**Response:**
```
Preview (not applied):

Line 45:
- func oldName(ctx context.Context) error {
+ func newName(ctx context.Context) error {

2 occurrences would be replaced.
```

**Simpler alternative:** Return affected lines in success response:
```
Replaced 2 occurrences:
  Line 45: func newName(ctx context.Context) error {
  Line 123: // newName handles the request
```

### 5. Read File with Pattern Filter

Read only lines matching a pattern (useful for large files):

```json
{
  "path": "/path/to/large-file.go",
  "pattern": "func.*Handler",
  "context_lines": 3
}
```

## Tool Preference Summary

| Scenario | Preferred Tool | Reason |
|----------|---------------|--------|
| Read multiple related files | `mcp__filesystem__read_multiple_files` | Single call efficiency |
| Explore project structure | `mcp__filesystem__list_directory` | Rich metadata + recursion |
| Simple string replacement | `mcp__filesystem__replace_in_file` | Safe, straightforward |
| Access restricted paths | Built-in `Read` | Broader access |
| Code search | Built-in `Grep` | Optimized ripgrep |
| Precise line edits | Built-in `Edit` | Line number support |
