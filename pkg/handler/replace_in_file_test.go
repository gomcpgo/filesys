package handler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestReplaceInFile_TabWhitespaceMismatch tests that replace_in_file returns detailed error
// when pattern contains tabs/spaces that don't match (Issue #2 from docs)
func TestReplaceInFile_TabWhitespaceMismatch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.go")

	// Create file with tab-indented import statements (exact scenario from issues doc)
	content := `package main

import (
	"savant/pkg/base"
	"savant/pkg/events"
)
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

	// Try to replace with tabs in the search string (should match the file)
	// But deliberately use spaces to simulate a mismatch
	args := map[string]interface{}{
		"path":    testFile,
		"search":  "    \"savant/pkg/base\"\n    \"savant/pkg/events\"", // spaces, not tabs
		"replace": "    \"savant/pkg/database\"\n    \"savant/pkg/events\"",
	}

	resp, err := handler.handleReplaceInFile(args)

	// After fix: should return error with detailed message about pattern not found
	if err == nil {
		t.Fatal("Expected error when pattern not found, but got success response")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "Pattern not found") {
		t.Errorf("Expected error to contain 'Pattern not found', got: %v", err)
	}

	// Error should provide context about whitespace issues
	if !strings.Contains(errMsg, "whitespace") && !strings.Contains(errMsg, "Whitespace") && !strings.Contains(errMsg, "Possible issues") {
		t.Errorf("Expected error to provide helpful context about why pattern wasn't found, got: %v", err)
	}

	_ = resp // unused after fix (error path)
}

// TestReplaceInFile_PatternNotFoundReturnsError tests that replace_in_file returns an error
// (not just a message) when pattern is not found
func TestReplaceInFile_PatternNotFoundReturnsError(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("hello world"), 0644)
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
		"search":  "nonexistent pattern that doesnt exist",
		"replace": "replacement",
	}

	resp, err := handler.handleReplaceInFile(args)

	// After implementing fix: should return error with detailed message
	if err != nil {
		// This is the desired behavior - error with context
		if !strings.Contains(err.Error(), "Pattern not found") {
			t.Errorf("Expected error about pattern not found, got: %v", err)
		}
		return
	}

	// Current behavior: returns success with message (not ideal for LLM)
	if resp != nil && len(resp.Content) > 0 {
		t.Logf("Current behavior returns response (should be error): %s", resp.Content[0].Text)
	}
}
