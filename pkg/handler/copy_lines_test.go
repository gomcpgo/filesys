package handler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCopyLinesEntireFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-copylines-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sourceContent := "line1\nline2\nline3\nline4\nline5\n"
	sourceFile := filepath.Join(tmpDir, "source.txt")
	destFile := filepath.Join(tmpDir, "dest.txt")
	os.WriteFile(sourceFile, []byte(sourceContent), 0644)

	os.Setenv("MCP_ALLOWED_DIRS", tmpDir)
	defer os.Unsetenv("MCP_ALLOWED_DIRS")
	allowedDirsMutex.Lock()
	allowedDirsCache = nil
	allowedDirsMutex.Unlock()

	handler := NewFileSystemHandler()
	args := map[string]interface{}{
		"source_path":      sourceFile,
		"destination_path": destFile,
	}

	resp, err := handler.handleCopyLines(args)
	if err != nil {
		t.Fatalf("copy_lines failed: %v", err)
	}

	// Verify destination has the content
	actual, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("Failed to read dest file: %v", err)
	}
	if string(actual) != sourceContent {
		t.Errorf("Expected %q, got %q", sourceContent, string(actual))
	}

	// Verify response is metadata only, not content
	if contains(resp.Content[0].Text, "line1") {
		t.Error("Response should contain metadata only, not file content")
	}
	if !contains(resp.Content[0].Text, "Copied 5 lines") {
		t.Errorf("Response should report line count, got: %s", resp.Content[0].Text)
	}
}

func TestCopyLinesRange(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-copylines-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sourceContent := "line1\nline2\nline3\nline4\nline5\n"
	sourceFile := filepath.Join(tmpDir, "source.txt")
	destFile := filepath.Join(tmpDir, "dest.txt")
	os.WriteFile(sourceFile, []byte(sourceContent), 0644)

	os.Setenv("MCP_ALLOWED_DIRS", tmpDir)
	defer os.Unsetenv("MCP_ALLOWED_DIRS")
	allowedDirsMutex.Lock()
	allowedDirsCache = nil
	allowedDirsMutex.Unlock()

	handler := NewFileSystemHandler()
	args := map[string]interface{}{
		"source_path":      sourceFile,
		"destination_path": destFile,
		"start_line":       float64(2),
		"end_line":         float64(4),
	}

	_, err = handler.handleCopyLines(args)
	if err != nil {
		t.Fatalf("copy_lines range failed: %v", err)
	}

	actual, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("Failed to read dest file: %v", err)
	}

	expected := "line2\nline3\nline4\n"
	if string(actual) != expected {
		t.Errorf("Expected %q, got %q", expected, string(actual))
	}
}

func TestCopyLinesAppendMode(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-copylines-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sourceContent := "line1\nline2\nline3\n"
	sourceFile := filepath.Join(tmpDir, "source.txt")
	destFile := filepath.Join(tmpDir, "dest.txt")
	os.WriteFile(sourceFile, []byte(sourceContent), 0644)
	os.WriteFile(destFile, []byte("existing\n"), 0644)

	os.Setenv("MCP_ALLOWED_DIRS", tmpDir)
	defer os.Unsetenv("MCP_ALLOWED_DIRS")
	allowedDirsMutex.Lock()
	allowedDirsCache = nil
	allowedDirsMutex.Unlock()

	handler := NewFileSystemHandler()
	args := map[string]interface{}{
		"source_path":      sourceFile,
		"destination_path": destFile,
		"append":           true,
	}

	_, err = handler.handleCopyLines(args)
	if err != nil {
		t.Fatalf("copy_lines append failed: %v", err)
	}

	actual, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("Failed to read dest file: %v", err)
	}

	expected := "existing\nline1\nline2\nline3\n"
	if string(actual) != expected {
		t.Errorf("Expected %q, got %q", expected, string(actual))
	}
}

func TestCopyLinesAccessDeniedSource(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-copylines-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	allowedDir := filepath.Join(tmpDir, "allowed")
	restrictedDir := filepath.Join(tmpDir, "restricted")
	os.MkdirAll(allowedDir, 0755)
	os.MkdirAll(restrictedDir, 0755)

	sourceFile := filepath.Join(restrictedDir, "source.txt")
	destFile := filepath.Join(allowedDir, "dest.txt")
	os.WriteFile(sourceFile, []byte("secret"), 0644)

	os.Setenv("MCP_ALLOWED_DIRS", allowedDir)
	defer os.Unsetenv("MCP_ALLOWED_DIRS")
	allowedDirsMutex.Lock()
	allowedDirsCache = nil
	allowedDirsMutex.Unlock()

	handler := NewFileSystemHandler()
	args := map[string]interface{}{
		"source_path":      sourceFile,
		"destination_path": destFile,
	}

	_, err = handler.handleCopyLines(args)
	if err == nil {
		t.Fatal("Expected error for access denied source")
	}
}

func TestCopyLinesAccessDeniedDest(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-copylines-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	allowedDir := filepath.Join(tmpDir, "allowed")
	restrictedDir := filepath.Join(tmpDir, "restricted")
	os.MkdirAll(allowedDir, 0755)
	os.MkdirAll(restrictedDir, 0755)

	sourceFile := filepath.Join(allowedDir, "source.txt")
	destFile := filepath.Join(restrictedDir, "dest.txt")
	os.WriteFile(sourceFile, []byte("content"), 0644)

	os.Setenv("MCP_ALLOWED_DIRS", allowedDir)
	defer os.Unsetenv("MCP_ALLOWED_DIRS")
	allowedDirsMutex.Lock()
	allowedDirsCache = nil
	allowedDirsMutex.Unlock()

	handler := NewFileSystemHandler()
	args := map[string]interface{}{
		"source_path":      sourceFile,
		"destination_path": destFile,
	}

	_, err = handler.handleCopyLines(args)
	if err == nil {
		t.Fatal("Expected error for access denied destination")
	}
}

func TestCopyLinesSourceNotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-copylines-test-")
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
	args := map[string]interface{}{
		"source_path":      filepath.Join(tmpDir, "nonexistent.txt"),
		"destination_path": filepath.Join(tmpDir, "dest.txt"),
	}

	_, err = handler.handleCopyLines(args)
	if err == nil {
		t.Fatal("Expected error for nonexistent source")
	}
}

func TestCopyLinesStartGreaterThanEnd(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-copylines-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	sourceFile := filepath.Join(tmpDir, "source.txt")
	os.WriteFile(sourceFile, []byte("line1\nline2\n"), 0644)

	os.Setenv("MCP_ALLOWED_DIRS", tmpDir)
	defer os.Unsetenv("MCP_ALLOWED_DIRS")
	allowedDirsMutex.Lock()
	allowedDirsCache = nil
	allowedDirsMutex.Unlock()

	handler := NewFileSystemHandler()
	args := map[string]interface{}{
		"source_path":      sourceFile,
		"destination_path": filepath.Join(tmpDir, "dest.txt"),
		"start_line":       float64(5),
		"end_line":         float64(2),
	}

	_, err = handler.handleCopyLines(args)
	if err == nil {
		t.Fatal("Expected error for start_line > end_line")
	}
}

func TestReadFileUnboundedCap(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-readcap-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file larger than 40KB
	var sb strings.Builder
	lineContent := strings.Repeat("x", 100) // 100 chars per line
	for i := 0; i < 600; i++ {              // 600 lines × ~101 bytes = ~60KB
		sb.WriteString(lineContent)
		sb.WriteByte('\n')
	}
	largeFile := filepath.Join(tmpDir, "large.txt")
	os.WriteFile(largeFile, []byte(sb.String()), 0644)

	os.Setenv("MCP_ALLOWED_DIRS", tmpDir)
	defer os.Unsetenv("MCP_ALLOWED_DIRS")
	allowedDirsMutex.Lock()
	allowedDirsCache = nil
	allowedDirsMutex.Unlock()

	handler := NewFileSystemHandler()
	args := map[string]interface{}{
		"path": largeFile,
	}

	resp, err := handler.handleReadFile(args)
	if err != nil {
		t.Fatalf("read_file failed: %v", err)
	}

	// Should have metadata element indicating truncation
	if len(resp.Content) < 2 {
		t.Fatal("Expected metadata element for capped read")
	}

	metadata := resp.Content[1].Text
	if !contains(metadata, "start_line") {
		t.Errorf("Metadata should suggest using start_line/end_line, got: %s", metadata)
	}

	// Content should be less than the full file
	if len(resp.Content[0].Text) >= sb.Len() {
		t.Error("Content should be capped, not the full file")
	}
}

func TestReadFileSmallFileNoCap(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-readcap-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	smallContent := "line1\nline2\nline3\n"
	smallFile := filepath.Join(tmpDir, "small.txt")
	os.WriteFile(smallFile, []byte(smallContent), 0644)

	os.Setenv("MCP_ALLOWED_DIRS", tmpDir)
	defer os.Unsetenv("MCP_ALLOWED_DIRS")
	allowedDirsMutex.Lock()
	allowedDirsCache = nil
	allowedDirsMutex.Unlock()

	handler := NewFileSystemHandler()
	args := map[string]interface{}{
		"path": smallFile,
	}

	resp, err := handler.handleReadFile(args)
	if err != nil {
		t.Fatalf("read_file failed: %v", err)
	}

	// Small file should be returned in full with no metadata
	if resp.Content[0].Text != smallContent {
		t.Errorf("Expected full content %q, got %q", smallContent, resp.Content[0].Text)
	}
	if len(resp.Content) > 1 {
		t.Error("Small file should not have metadata element")
	}
}

func TestReadFileWithRangeAllowed(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-readcap-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file where 200 lines is under 100KB but full file is large
	var sb strings.Builder
	lineContent := strings.Repeat("x", 100)
	for i := 0; i < 600; i++ {
		sb.WriteString(lineContent)
		sb.WriteByte('\n')
	}
	largeFile := filepath.Join(tmpDir, "large.txt")
	os.WriteFile(largeFile, []byte(sb.String()), 0644)

	os.Setenv("MCP_ALLOWED_DIRS", tmpDir)
	defer os.Unsetenv("MCP_ALLOWED_DIRS")
	allowedDirsMutex.Lock()
	allowedDirsCache = nil
	allowedDirsMutex.Unlock()

	handler := NewFileSystemHandler()
	// Request 200 lines explicitly — should work under the 100KB ranged cap
	args := map[string]interface{}{
		"path":       largeFile,
		"start_line": float64(1),
		"end_line":   float64(200),
	}

	resp, err := handler.handleReadFile(args)
	if err != nil {
		t.Fatalf("read_file with range failed: %v", err)
	}

	// 200 lines × 101 bytes = ~20KB, well under 100KB ranged cap
	lines := strings.Split(strings.TrimRight(resp.Content[0].Text, "\n"), "\n")
	if len(lines) != 200 {
		t.Errorf("Expected 200 lines, got %d", len(lines))
	}
}

func TestReadMultipleFilesCap(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-readcap-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a large file (>40KB)
	var sb strings.Builder
	lineContent := strings.Repeat("x", 100)
	for i := 0; i < 600; i++ {
		sb.WriteString(lineContent)
		sb.WriteByte('\n')
	}
	largeFile := filepath.Join(tmpDir, "large.txt")
	os.WriteFile(largeFile, []byte(sb.String()), 0644)

	os.Setenv("MCP_ALLOWED_DIRS", tmpDir)
	defer os.Unsetenv("MCP_ALLOWED_DIRS")
	allowedDirsMutex.Lock()
	allowedDirsCache = nil
	allowedDirsMutex.Unlock()

	handler := NewFileSystemHandler()
	args := map[string]interface{}{
		"paths": []interface{}{largeFile},
	}

	resp, err := handler.handleReadMultipleFiles(args)
	if err != nil {
		t.Fatalf("read_multiple_files failed: %v", err)
	}

	// Should contain truncation info and guidance
	if !contains(resp.Content[0].Text, "start_line") {
		t.Errorf("Should suggest using start_line/end_line, got: %s", resp.Content[0].Text)
	}
}
