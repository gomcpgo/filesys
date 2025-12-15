# Filesystem MCP Server Improvements - Implementation Plan

## Scope

Three improvements based on user feedback:
1. Better error messages for path access restrictions
2. Batch replace across multiple files (new tool)
3. Dry run/preview mode for replace operations

## Implementation Details

### 1. Better Error Messages for Path Access

**Current behavior:** Error returns `"access to path is not allowed"`

**New behavior:** Error includes allowed directories:
```
access to path '/Users/prasanth/Library/...' is not allowed.
Allowed directories: /Users/prasanth/MyProjects, /Users/prasanth/Documents
Hint: Use list_allowed_directories tool to see all accessible paths.
```

**Files to modify:**
- `pkg/handler/utils.go` - Update `isPathAllowed()` and `isPathAllowedNonExistent()` to return more informative errors

**Implementation:**
1. Create helper function `formatAccessDeniedError(requestedPath string) error` that:
   - Gets allowed directories from cache/env
   - Formats them into a readable list
   - Includes the requested path in the error
   - Adds hint about `list_allowed_directories` tool

2. Update error returns in path validation functions to use this helper

---

### 2. Batch Replace Across Files (New Tool)

**New tool:** `replace_in_files` (plural)

**Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| paths | string[] | Yes | List of file paths to modify |
| search | string | Yes | String to search for |
| replace | string | Yes | Replacement string |
| dry_run | bool | No | Preview changes without applying (default: false) |

**Response format:**
```
Replaced in 3 of 4 files:

/path/to/file1.go:
  Line 45: func newName(ctx context.Context) error {
  Line 123: // newName handles the request
  (2 replacements)

/path/to/file2.go:
  Line 12: import "newName"
  (1 replacement)

/path/to/file3.go:
  Line 89: newName := value
  (1 replacement)

/path/to/file4.go:
  No matches found

Total: 4 replacements across 3 files
```

**Design Decision:** Uses explicit file paths only (no glob patterns) for safety. Users should use `search_in_files` first to identify target files.

**Files to create/modify:**
- `pkg/handler/replace_in_files.go` - New handler
- `pkg/handler/tools.go` - Add tool definition
- `pkg/handler/filesystem.go` - Add case to CallTool switch

**Implementation:**
1. Create `handleReplaceInFiles()` function that:
   - Validates all paths are allowed first (fail fast)
   - For each file:
     - Read content
     - Find all occurrences with line numbers
     - If not dry_run, write modified content
     - Build result summary
   - Return aggregated results

---

### 3. Dry Run/Preview Mode for Replace Operations

**Add `dry_run` parameter to existing tools:**
- `replace_in_file`
- `replace_in_file_regex`

**Behavior when `dry_run: true`:**
- Find all matches with line numbers and context
- Do NOT write changes to file
- Return preview showing what would change

**Response format:**
```
Preview (dry run - no changes applied):

Line 45:
- func oldName(ctx context.Context) error {
+ func newName(ctx context.Context) error {

Line 123:
- // oldName handles the request
+ // newName handles the request

2 occurrences would be replaced.
```

**Files to modify:**
- `pkg/handler/replace_in_file.go` - Add dry_run logic
- `pkg/handler/replace_in_file_regex.go` - Add dry_run logic
- `pkg/handler/tools.go` - Update tool schemas with dry_run parameter

**Implementation:**
1. Update tool JSON schemas to include `dry_run` boolean parameter
2. Modify handlers to:
   - Extract `dry_run` from params (default false)
   - Track matches with line numbers during search
   - If dry_run, skip file write and format preview response
   - Otherwise, proceed with existing replacement logic

---

## Implementation Order

1. **Better error messages** (simplest, self-contained change)
2. **Dry run for existing tools** (modifies existing code, needed for #3)
3. **Batch replace tool** (new tool, builds on patterns from #2)

## Critical Files Summary

| File | Changes |
|------|---------|
| `pkg/handler/utils.go` | Add `formatAccessDeniedError()`, update error returns |
| `pkg/handler/tools.go` | Add `dry_run` params to schemas, add new tool definition |
| `pkg/handler/filesystem.go` | Add case for `replace_in_files` |
| `pkg/handler/replace_in_file.go` | Add dry_run support, line tracking |
| `pkg/handler/replace_in_file_regex.go` | Add dry_run support, line tracking |
| `pkg/handler/replace_in_files.go` | New file for batch replacement |

## Testing

After each change, run:
```bash
cd mcp_servers/filesys && go test ./...
```

Add tests for:
- Error message includes allowed paths
- Dry run returns preview without modifying file
- Batch replace handles partial failures gracefully
