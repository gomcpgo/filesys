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
