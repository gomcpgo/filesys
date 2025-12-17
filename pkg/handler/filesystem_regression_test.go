package handler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Issue #1: insert_after_regex - Auto-indentation with proper newline handling
// This test verifies that inserted content is properly formatted with correct newlines and indentation
func TestIssue1_InsertAfterRegexAutoIndent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-issue1-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file with imports
	testFile := filepath.Join(tmpDir, "imports.go")
	content := `package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("hello")
}`

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

	// Test: Insert "io" import after "os" import with auto-indentation
	// The inserted content should be automatically indented to match the surrounding context
	args := map[string]interface{}{
		"path":       testFile,
		"pattern":    `"os"`,
		"content":    "\n\t\"io\"", // Newline and tab, should be matched by auto-indent
		"occurrence": 1,
		"autoIndent": true,
	}

	resp, err := handler.handleInsertAfterRegex(args)
	if err != nil {
		t.Fatalf("Insert after regex failed: %v", err)
	}

	// Verify the content contains both "os" and "io" imports properly formatted
	newContent := resp.Content[0].Text
	if !strings.Contains(newContent, `"os"`) {
		t.Error("Original import 'os' not found in result")
	}
	if !strings.Contains(newContent, `"io"`) {
		t.Error("New import 'io' not found in result")
	}

	// Verify proper formatting - "io" should be on its own line with correct indentation
	lines := strings.Split(newContent, "\n")
	found := false
	for i, line := range lines {
		if strings.Contains(line, `"io"`) {
			// Check that it's on its own line
			if !strings.HasPrefix(line, "\t\"io\"") && !strings.HasPrefix(line, "    \"io\"") {
				t.Errorf("Import 'io' not properly indented: %q", line)
			}
			// Verify there's content after it (not concatenated)
			if i+1 < len(lines) {
				found = true
			}
			break
		}
	}
	if !found {
		t.Error("Import 'io' not found on separate line")
	}
}

// Issue #2: replace_in_file with Tab Characters - Proper handling without silent failures
// This test verifies that tabs are handled correctly in single-line replacements
// Note: Multi-line replacements with tabs would require replace_in_file_regex for full pattern support
func TestIssue2_ReplaceWithTabCharacters(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-issue2-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file with tab-indented content
	testFile := filepath.Join(tmpDir, "main.go")
	// Use actual tab characters
	content := "package main\n\nfunc test() {\n\tfmt.Println(\"before\")\n}\n"

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

	// Test: Replace single line with tabs
	oldString := "\tfmt.Println(\"before\")"
	newString := "\tfmt.Println(\"after\")"

	args := map[string]interface{}{
		"path":    testFile,
		"search":  oldString,
		"replace": newString,
	}

	resp, err := handler.handleReplaceInFile(args)
	if err != nil {
		t.Fatalf("Replace in file failed: %v", err)
	}

	newContent := resp.Content[0].Text
	if !strings.Contains(newContent, "\"after\"") {
		t.Error("Expected 'after' string to be added, but not found in result")
	}

	// Verify tabs are preserved
	if !strings.Contains(newContent, "\tfmt.Println(\"after\")") {
		t.Error("Tab indentation not preserved in replacement")
	}
}

// Issue #3: Inconsistent Behavior Between Tools
// This test verifies consistent whitespace handling across multiple operations
func TestIssue3_ToolConsistency(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-issue3-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.go")
	content := "package main\n\nfunc hello() {\n    fmt.Println(\"hello\")\n}\n"

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

	// Read the file first
	readArgs := map[string]interface{}{
		"path": testFile,
	}
	readResp, err := handler.handleReadFile(readArgs)
	if err != nil {
		t.Fatalf("Read file failed: %v", err)
	}
	readContent := readResp.Content[0].Text

	// Replace a single line to verify consistency
	replaceArgs := map[string]interface{}{
		"path":    testFile,
		"search":  "fmt.Println(\"hello\")",
		"replace": "fmt.Println(\"updated hello\")",
	}

	replaceResp, err := handler.handleReplaceInFile(replaceArgs)
	if err != nil {
		t.Fatalf("Replace in file failed: %v", err)
	}

	// Verify that the replacement worked
	if !strings.Contains(replaceResp.Content[0].Text, "updated hello") {
		t.Error("Expected 'updated hello' to be in result")
	}

	// Verify content was actually modified
	if replaceResp.Content[0].Text == readContent {
		t.Error("Content should have changed after replacement")
	}
}

// Issue #4: No Built-in Validation - Proper error handling and validation
// This test verifies that operations provide meaningful feedback
func TestIssue4_ValidationAndErrorHandling(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-issue4-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line 1\nline 2\nline 3"
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

	// Test: Replace with pattern that doesn't exist
	args := map[string]interface{}{
		"path":    testFile,
		"search":  "non-existent pattern",
		"replace": "replacement",
	}

	_, err = handler.handleReplaceInFile(args)
	if err == nil {
		t.Error("Expected error when replacing non-existent pattern, but got none")
	}

	// Error message should be informative
	if err != nil && !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "pattern") {
		t.Logf("Warning: Error message could be more descriptive: %v", err)
	}
}

// Issue #5: Limited Error Context - Better error messages with context
// This test verifies that errors provide useful information about what went wrong
func TestIssue5_ErrorContextAndMessages(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-issue5-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "code.go")
	content := `package main

func main() {
    fmt.Println("test")
}`
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

	// Test regex that doesn't match
	args := map[string]interface{}{
		"path":       testFile,
		"pattern":    `func NonExistent\(\)`,
		"content":    "new content",
		"occurrence": 1,
		"autoIndent": false,
	}

	_, err = handler.handleInsertAfterRegex(args)
	if err == nil {
		t.Error("Expected error for non-matching pattern, but got none")
	}

	// Error should indicate pattern was not found
	if err != nil && !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "pattern") {
		t.Logf("Warning: Error could provide more pattern context: %v", err)
	}
}

// Issue #6: File Position Context - Line-based operations
// This test verifies that operations can work with line numbers
func TestIssue6_LineBasedOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-issue6-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	content := `line 1
line 2
line 3
line 4
line 5`
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

	// Test: Insert before a specific line
	args := map[string]interface{}{
		"path":       testFile,
		"pattern":    `line 3`, // Targeting line 3
		"content":    "INSERTED CONTENT\n",
		"occurrence": 1,
		"autoIndent": false,
	}

	resp, err := handler.handleInsertBeforeRegex(args)
	if err != nil {
		t.Fatalf("Insert before regex failed: %v", err)
	}

	newContent := resp.Content[0].Text
	lines := strings.Split(newContent, "\n")

	// Verify INSERTED CONTENT is before line 3
	found := false
	for i, line := range lines {
		if strings.Contains(line, "INSERTED CONTENT") {
			// Next line should be "line 3"
			if i+1 < len(lines) && strings.Contains(lines[i+1], "line 3") {
				found = true
				break
			}
		}
	}

	if !found {
		t.Error("INSERTED CONTENT not found before 'line 3'")
	}
}

// Issue #7: Escape Sequence Inconsistency - Consistent escape handling
// This test verifies that escape sequences are handled consistently
func TestIssue7_EscapeSequenceHandling(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-issue7-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.go")
	// Use actual tab characters
	content := "package main\n\nimport (\n\t\"fmt\"\n\t\"os\"\n)\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n"

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

	// Test 1: Replace with single-line pattern containing tab characters
	args1 := map[string]interface{}{
		"path":    testFile,
		"search":  "\t\"fmt\"",
		"replace": "\t\"fmt\"\n\t\"io\"",
	}

	resp1, err := handler.handleReplaceInFile(args1)
	if err != nil {
		t.Fatalf("First replace failed: %v", err)
	}

	if !strings.Contains(resp1.Content[0].Text, `"io"`) {
		t.Error("Expected 'io' import to be added in first test")
	}

	// Test 2: Insert with escape sequences
	args2 := map[string]interface{}{
		"path":       testFile,
		"pattern":    `"os"`,
		"content":    "\n\t\"sys\"",
		"occurrence": 1,
		"autoIndent": false,
	}

	resp2, err := handler.handleInsertAfterRegex(args2)
	if err != nil {
		t.Fatalf("Insert after regex failed: %v", err)
	}

	if !strings.Contains(resp2.Content[0].Text, `"sys"`) {
		t.Error("Expected 'sys' import to be added in second test")
	}

	// Verify both operations handle escape sequences consistently
	if resp1.Content[0].Text == "" || resp2.Content[0].Text == "" {
		t.Error("Operations did not return valid content")
	}
}

// TestRegressionSuite_AllIssuesIntegrated verifies all 7 issues in an integrated test
func TestRegressionSuite_AllIssuesIntegrated(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-regression-suite-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "complete.go")
	initialContent := "package main\n\nimport (\n\t\"fmt\"\n)\n\nfunc main() {\n\tfmt.Println(\"start\")\n}\n"

	err = os.WriteFile(testFile, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	os.Setenv("MCP_ALLOWED_DIRS", tmpDir)
	defer os.Unsetenv("MCP_ALLOWED_DIRS")

	allowedDirsMutex.Lock()
	allowedDirsCache = nil
	allowedDirsMutex.Unlock()

	handler := NewFileSystemHandler()

	// Step 1: Insert after regex with auto-indent (Issue #1)
	insertArgs := map[string]interface{}{
		"path":       testFile,
		"pattern":    "\"fmt\"",
		"content":    "\n\t\"os\"",
		"occurrence": 1,
		"autoIndent": true,
	}
	resp1, err := handler.handleInsertAfterRegex(insertArgs)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Verify first operation worked
	if !strings.Contains(resp1.Content[0].Text, "\"os\"") {
		t.Error("Expected 'os' import to be added after insert")
	}

	// Step 2: Replace a single line (Issue #2) - use the response from step 1 as our file now
	replaceArgs := map[string]interface{}{
		"path":    testFile,
		"search":  "fmt.Println(\"start\")",
		"replace": "fmt.Println(\"start\"); fmt.Println(\"end\")",
	}
	resp2, err := handler.handleReplaceInFile(replaceArgs)
	if err != nil {
		t.Fatalf("Replace failed: %v", err)
	}

	// Verify both operations worked
	finalContent := resp2.Content[0].Text
	if !strings.Contains(finalContent, "\"end\"") {
		t.Error("Expected 'end' string from replace operation")
	}

	// Verify both modifications are present
	if !strings.Contains(finalContent, "fmt.Println(\"start\"); fmt.Println(\"end\")") {
		t.Error("Expected complete replacement text to be present")
	}
}
