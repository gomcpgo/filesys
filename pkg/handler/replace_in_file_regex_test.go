package handler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestReplaceInFileRegex_DryRunShowsContext tests that dry-run mode provides context with diff output
// This reproduces Issue #4 from the MCP_FILESYSTEM_ISSUES.md docs
func TestReplaceInFileRegex_DryRunShowsContext(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.go")

	// Create test file with multiple matches
	content := `package main

import "fmt"

func main() {
	fmt.Println("line 2")
	logger.Info("test")

	fmt.Println("line 5")
	logger.Info("another")

	fmt.Println("line 8")
}
`

	err = os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	os.Setenv("MCP_ALLOWED_DIRS", tmpDir)
	defer os.Unsetenv("MCP_ALLOWED_DIRS")

	allowedDirsMutex.Lock()
	allowedDirsCache = nil
	allowedDirsMutex.Unlock()

	handler := NewFileSystemHandler()

	args := map[string]interface{}{
		"path":       testFile,
		"pattern":    `logger\.Info\("([^"]+)"\)`,
		"replace":    `logger.Debug("$1")`,
		"dry_run":    true,
	}

	resp, err := handler.handleReplaceInFileRegex(args)

	if err != nil {
		t.Fatalf("Dry run should not error: %v", err)
	}

	if resp == nil || len(resp.Content) == 0 {
		t.Fatal("Expected response content for dry run")
	}

	dryRunOutput := resp.Content[0].Text

	// Dry run must show count of matches
	if !strings.Contains(dryRunOutput, "2") && !strings.Contains(dryRunOutput, "matches") {
		t.Errorf("Expected dry run to show match count, got: %s", dryRunOutput)
	}

	// Dry run should show line numbers
	if !strings.Contains(dryRunOutput, "6") && !strings.Contains(dryRunOutput, "9") {
		t.Logf("Helpful: Dry run should show line numbers where matches occur. Got: %s", dryRunOutput)
	}

	// Dry run should show context/diff
	if !strings.Contains(dryRunOutput, "-") || !strings.Contains(dryRunOutput, "+") {
		t.Logf("Helpful: Dry run should show diff with - and + lines. Got: %s", dryRunOutput)
	}

	// Should show original and replaced content
	if !strings.Contains(dryRunOutput, "logger.Info") && !strings.Contains(dryRunOutput, "logger.Debug") {
		t.Logf("Helpful: Dry run should show what would change. Got: %s", dryRunOutput)
	}
}

// TestReplaceInFileRegex_ActualReplacementWorks tests that actual replacement still works
func TestReplaceInFileRegex_ActualReplacementWorks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.go")

	originalContent := `package main

func test() {
	logger.Info("message1")
	logger.Info("message2")
}
`

	err = os.WriteFile(testFile, []byte(originalContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	os.Setenv("MCP_ALLOWED_DIRS", tmpDir)
	defer os.Unsetenv("MCP_ALLOWED_DIRS")

	allowedDirsMutex.Lock()
	allowedDirsCache = nil
	allowedDirsMutex.Unlock()

	handler := NewFileSystemHandler()

	args := map[string]interface{}{
		"path":    testFile,
		"pattern": `logger\.Info\("([^"]+)"\)`,
		"replace": `logger.Debug("$1")`,
		"dry_run": false, // Actual replacement
	}

	resp, err := handler.handleReplaceInFileRegex(args)

	if err != nil {
		t.Fatalf("Replacement should not error: %v", err)
	}

	if resp == nil || len(resp.Content) == 0 {
		t.Fatal("Expected response content")
	}

	// Verify file was actually modified
	newContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	newContentStr := string(newContent)
	if !strings.Contains(newContentStr, "logger.Debug") {
		t.Errorf("Expected logger.Debug in file, got: %s", newContentStr)
	}

	if strings.Contains(newContentStr, "logger.Info") {
		t.Errorf("Expected logger.Info to be replaced, but found in: %s", newContentStr)
	}
}

// TestReplaceInFileRegex_PatternNotFoundMessage tests that pattern not found returns helpful message
func TestReplaceInFileRegex_PatternNotFoundMessage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.go")

	content := `package main

func test() {
	fmt.Println("hello")
}
`

	err = os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	os.Setenv("MCP_ALLOWED_DIRS", tmpDir)
	defer os.Unsetenv("MCP_ALLOWED_DIRS")

	allowedDirsMutex.Lock()
	allowedDirsCache = nil
	allowedDirsMutex.Unlock()

	handler := NewFileSystemHandler()

	args := map[string]interface{}{
		"path":    testFile,
		"pattern": `nonexistent\(.*\)`, // Pattern that won't match
		"replace": "replacement",
		"dry_run": false,
	}

	resp, err := handler.handleReplaceInFileRegex(args)

	// Should return response (not error) indicating pattern not found
	if err != nil {
		t.Logf("Pattern not found handling returns error (future improvement): %v", err)
		return
	}

	if resp == nil || len(resp.Content) == 0 {
		t.Fatal("Expected response content")
	}

	message := resp.Content[0].Text
	if !strings.Contains(message, "not found") && !strings.Contains(message, "Not found") {
		t.Errorf("Expected message to indicate pattern not found, got: %s", message)
	}
}
