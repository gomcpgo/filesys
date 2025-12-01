package handler

import (
	"os"
	"path/filepath"
	"testing"
)

// TestReadFileWithValidPath tests reading a file in allowed directory
func TestReadFileWithValidPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-integration-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "Hello, World!"
	err = os.WriteFile(testFile, []byte(testContent), 0644)
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
		"path": testFile,
	}

	resp, err := handler.handleReadFile(args)
	if err != nil {
		t.Fatalf("Read file failed: %v", err)
	}

	if len(resp.Content) == 0 {
		t.Fatal("No content returned")
	}

	if resp.Content[0].Text != testContent {
		t.Errorf("Expected %q, got %q", testContent, resp.Content[0].Text)
	}
}

// TestReadFileWithInvalidPath tests reading a file outside allowed directory
func TestReadFileWithInvalidPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-integration-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	allowedDir := filepath.Join(tmpDir, "allowed")
	restrictedDir := filepath.Join(tmpDir, "restricted")

	err = os.MkdirAll(allowedDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create allowed dir: %v", err)
	}
	err = os.MkdirAll(restrictedDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create restricted dir: %v", err)
	}

	testFile := filepath.Join(restrictedDir, "secret.txt")
	err = os.WriteFile(testFile, []byte("secret"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	os.Setenv("MCP_ALLOWED_DIRS", allowedDir)
	defer os.Unsetenv("MCP_ALLOWED_DIRS")

	allowedDirsMutex.Lock()
	allowedDirsCache = nil
	allowedDirsMutex.Unlock()

	handler := NewFileSystemHandler()

	args := map[string]interface{}{
		"path": testFile,
	}

	_, err = handler.handleReadFile(args)
	if err == nil {
		t.Fatal("Expected error when reading file outside allowed directory")
	}
}

// TestWriteFileInAllowedDirectory tests writing to allowed directory
func TestWriteFileInAllowedDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-integration-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	os.Setenv("MCP_ALLOWED_DIRS", tmpDir)
	defer os.Unsetenv("MCP_ALLOWED_DIRS")

	allowedDirsMutex.Lock()
	allowedDirsCache = nil
	allowedDirsMutex.Unlock()

	handler := NewFileSystemHandler()

	testFile := filepath.Join(tmpDir, "newfile.txt")
	testContent := "New content"

	args := map[string]interface{}{
		"path":    testFile,
		"content": testContent,
	}

	_, err = handler.handleWriteFile(args)
	if err != nil {
		t.Fatalf("Write file failed: %v", err)
	}

	// Verify file was created
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("Expected %q, got %q", testContent, string(content))
	}
}

// TestWriteFileInRestrictedDirectory tests writing to restricted directory
func TestWriteFileInRestrictedDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-integration-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	allowedDir := filepath.Join(tmpDir, "allowed")
	restrictedDir := filepath.Join(tmpDir, "restricted")

	err = os.MkdirAll(allowedDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create allowed dir: %v", err)
	}
	err = os.MkdirAll(restrictedDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create restricted dir: %v", err)
	}

	os.Setenv("MCP_ALLOWED_DIRS", allowedDir)
	defer os.Unsetenv("MCP_ALLOWED_DIRS")

	allowedDirsMutex.Lock()
	allowedDirsCache = nil
	allowedDirsMutex.Unlock()

	handler := NewFileSystemHandler()

	testFile := filepath.Join(restrictedDir, "newfile.txt")
	testContent := "Should not be created"

	args := map[string]interface{}{
		"path":    testFile,
		"content": testContent,
	}

	_, err = handler.handleWriteFile(args)
	if err == nil {
		t.Fatal("Expected error when writing file to restricted directory")
	}

	// Verify file was NOT created
	if _, err := os.Stat(testFile); err == nil {
		t.Error("File should not have been created in restricted directory")
	}
}

// TestAccessDeniedErrorFormat tests that access denied errors include allowed directories
func TestAccessDeniedErrorFormat(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	allowedDir := filepath.Join(tmpDir, "allowed")
	err = os.MkdirAll(allowedDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create allowed dir: %v", err)
	}

	os.Setenv("MCP_ALLOWED_DIRS", allowedDir)
	defer os.Unsetenv("MCP_ALLOWED_DIRS")

	allowedDirsMutex.Lock()
	allowedDirsCache = nil
	allowedDirsMutex.Unlock()

	// Create the error
	testPath := "/some/restricted/path"
	err = NewAccessDeniedError(testPath)

	// Check error message contains expected parts
	errMsg := err.Error()
	if !contains(errMsg, testPath) {
		t.Errorf("Error message should contain requested path %q, got: %s", testPath, errMsg)
	}
	if !contains(errMsg, allowedDir) {
		t.Errorf("Error message should contain allowed directory %q, got: %s", allowedDir, errMsg)
	}
	if !contains(errMsg, "list_allowed_directories") {
		t.Errorf("Error message should contain hint about list_allowed_directories, got: %s", errMsg)
	}
}

// TestReplaceInFileDryRun tests dry run mode for replace_in_file
func TestReplaceInFileDryRun(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	originalContent := "Hello World\nHello Again\nGoodbye"
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
		"search":  "Hello",
		"replace": "Hi",
		"dry_run": true,
	}

	resp, err := handler.handleReplaceInFile(args)
	if err != nil {
		t.Fatalf("Dry run failed: %v", err)
	}

	t.Logf("Response from handleReplaceInFile %v ", resp)

	// Check response indicates dry run
	if !contains(resp.Content[0].Text, "dry run") {
		t.Errorf("Response should indicate dry run mode, got: %s", resp.Content[0].Text)
	}
	if !contains(resp.Content[0].Text, "would be replaced") {
		t.Errorf("Response should say 'would be replaced', got: %s", resp.Content[0].Text)
	}

	// Verify file was NOT modified
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(content) != originalContent {
		t.Errorf("File should not be modified in dry run. Expected %q, got %q", originalContent, string(content))
	}
}

// TestReplaceInFileDryRunMultiline tests dry run with multiple lines and verifies preview format
func TestReplaceInFileDryRunMultiline(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "code.go")
	originalContent := `package main

func oldFunction() {
	fmt.Println("oldFunction called")
}

func anotherFunc() {
	oldFunction()
}

func oldFunction2() {
	// This also has oldFunction in comment
}`
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
		"search":  "oldFunction",
		"replace": "newFunction",
		"dry_run": true,
	}

	resp, err := handler.handleReplaceInFile(args)
	if err != nil {
		t.Fatalf("Dry run failed: %v", err)
	}

	t.Logf("Multiple lines replacement test response: %v", resp)

	responseText := resp.Content[0].Text

	// Verify dry run indicator
	if !contains(responseText, "dry run") {
		t.Errorf("Response should indicate dry run mode, got: %s", responseText)
	}

	// Verify line numbers are shown
	if !contains(responseText, "Line 3") {
		t.Errorf("Response should show Line 3 (func oldFunction), got: %s", responseText)
	}
	if !contains(responseText, "Line 4") {
		t.Errorf("Response should show Line 4 (fmt.Println), got: %s", responseText)
	}

	// Verify old and new lines are shown (diff format)
	if !contains(responseText, "- ") && !contains(responseText, "+ ") {
		t.Errorf("Response should show diff format with - and +, got: %s", responseText)
	}

	// Verify replacement text appears in preview
	if !contains(responseText, "newFunction") {
		t.Errorf("Response should show replacement text 'newFunction', got: %s", responseText)
	}

	// Verify count message
	if !contains(responseText, "would be replaced") {
		t.Errorf("Response should say 'would be replaced', got: %s", responseText)
	}

	// CRITICAL: Verify file was NOT modified
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(content) != originalContent {
		t.Errorf("File should NOT be modified in dry run.\nExpected:\n%s\n\nGot:\n%s", originalContent, string(content))
	}

	// Verify original content still has oldFunction
	if !contains(string(content), "oldFunction") {
		t.Error("Original file should still contain 'oldFunction'")
	}
}

// TestReplaceInFileActual tests actual replacement (not dry run)
func TestReplaceInFileActual(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	originalContent := "Hello World\nHello Again"
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
		"search":  "Hello",
		"replace": "Hi",
	}

	resp, err := handler.handleReplaceInFile(args)
	if err != nil {
		t.Fatalf("Replace failed: %v", err)
	}

	// Check response indicates success
	if !contains(resp.Content[0].Text, "Successfully replaced") {
		t.Errorf("Response should indicate success, got: %s", resp.Content[0].Text)
	}

	// Verify file WAS modified
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	expectedContent := "Hi World\nHi Again"
	if string(content) != expectedContent {
		t.Errorf("File should be modified. Expected %q, got %q", expectedContent, string(content))
	}
}

// TestReplaceInFilesDryRun tests dry run mode for batch replace
func TestReplaceInFilesDryRun(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create two test files
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	content1 := "Hello from file1"
	content2 := "Hello from file2"

	err = os.WriteFile(file1, []byte(content1), 0644)
	if err != nil {
		t.Fatalf("Failed to write file1: %v", err)
	}
	err = os.WriteFile(file2, []byte(content2), 0644)
	if err != nil {
		t.Fatalf("Failed to write file2: %v", err)
	}

	os.Setenv("MCP_ALLOWED_DIRS", tmpDir)
	defer os.Unsetenv("MCP_ALLOWED_DIRS")

	allowedDirsMutex.Lock()
	allowedDirsCache = nil
	allowedDirsMutex.Unlock()

	handler := NewFileSystemHandler()

	args := map[string]interface{}{
		"paths":   []interface{}{file1, file2},
		"search":  "Hello",
		"replace": "Hi",
		"dry_run": true,
	}

	resp, err := handler.handleReplaceInFiles(args)
	if err != nil {
		t.Fatalf("Batch dry run failed: %v", err)
	}

	// Check response indicates dry run
	if !contains(resp.Content[0].Text, "dry run") {
		t.Errorf("Response should indicate dry run mode, got: %s", resp.Content[0].Text)
	}

	// Verify files were NOT modified
	actual1, _ := os.ReadFile(file1)
	actual2, _ := os.ReadFile(file2)
	if string(actual1) != content1 {
		t.Errorf("File1 should not be modified. Expected %q, got %q", content1, string(actual1))
	}
	if string(actual2) != content2 {
		t.Errorf("File2 should not be modified. Expected %q, got %q", content2, string(actual2))
	}
}

// TestReplaceInFilesActual tests actual batch replacement
func TestReplaceInFilesActual(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create two test files
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	err = os.WriteFile(file1, []byte("Hello from file1"), 0644)
	if err != nil {
		t.Fatalf("Failed to write file1: %v", err)
	}
	err = os.WriteFile(file2, []byte("Hello from file2"), 0644)
	if err != nil {
		t.Fatalf("Failed to write file2: %v", err)
	}

	os.Setenv("MCP_ALLOWED_DIRS", tmpDir)
	defer os.Unsetenv("MCP_ALLOWED_DIRS")

	allowedDirsMutex.Lock()
	allowedDirsCache = nil
	allowedDirsMutex.Unlock()

	handler := NewFileSystemHandler()

	args := map[string]interface{}{
		"paths":   []interface{}{file1, file2},
		"search":  "Hello",
		"replace": "Hi",
	}

	resp, err := handler.handleReplaceInFiles(args)
	if err != nil {
		t.Fatalf("Batch replace failed: %v", err)
	}

	// Check response indicates success
	if !contains(resp.Content[0].Text, "Replaced in 2 of 2 files") {
		t.Errorf("Response should indicate 2 files modified, got: %s", resp.Content[0].Text)
	}

	// Verify files WERE modified
	actual1, _ := os.ReadFile(file1)
	actual2, _ := os.ReadFile(file2)
	if string(actual1) != "Hi from file1" {
		t.Errorf("File1 should be modified. Expected %q, got %q", "Hi from file1", string(actual1))
	}
	if string(actual2) != "Hi from file2" {
		t.Errorf("File2 should be modified. Expected %q, got %q", "Hi from file2", string(actual2))
	}
}

// TestReplaceInFilesRestrictedPath tests that batch replace fails if any path is restricted
func TestReplaceInFilesRestrictedPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	allowedDir := filepath.Join(tmpDir, "allowed")
	restrictedDir := filepath.Join(tmpDir, "restricted")
	os.MkdirAll(allowedDir, 0755)
	os.MkdirAll(restrictedDir, 0755)

	allowedFile := filepath.Join(allowedDir, "file1.txt")
	restrictedFile := filepath.Join(restrictedDir, "file2.txt")

	os.WriteFile(allowedFile, []byte("Hello"), 0644)
	os.WriteFile(restrictedFile, []byte("Hello"), 0644)

	os.Setenv("MCP_ALLOWED_DIRS", allowedDir)
	defer os.Unsetenv("MCP_ALLOWED_DIRS")

	allowedDirsMutex.Lock()
	allowedDirsCache = nil
	allowedDirsMutex.Unlock()

	handler := NewFileSystemHandler()

	args := map[string]interface{}{
		"paths":   []interface{}{allowedFile, restrictedFile},
		"search":  "Hello",
		"replace": "Hi",
	}

	_, err = handler.handleReplaceInFiles(args)
	if err == nil {
		t.Fatal("Expected error when one path is restricted")
	}

	// Verify the allowed file was NOT modified (fail fast behavior)
	actual, _ := os.ReadFile(allowedFile)
	if string(actual) != "Hello" {
		t.Errorf("Allowed file should not be modified when batch fails. Expected %q, got %q", "Hello", string(actual))
	}
}

// helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
