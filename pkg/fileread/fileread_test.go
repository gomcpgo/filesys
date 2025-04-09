package fileread

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test utility function to create a temporary file with content
func createTempFile(t *testing.T, content string) (string, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "fileread-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	tempFile := filepath.Join(tempDir, "testfile.txt")
	err = os.WriteFile(tempFile, []byte(content), 0644)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Return file path and cleanup function
	return tempFile, func() {
		os.RemoveAll(tempDir)
	}
}

// Test reading an entire small file with optimized path
func TestOptimizedReadSmallFile(t *testing.T) {
	content := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"
	tempFile, cleanup := createTempFile(t, content)
	defer cleanup()

	// Read the entire file
	result, err := ReadFile(tempFile, 0, 0, 1024)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify the result
	if result.Content != content {
		t.Errorf("Expected content: %q, got: %q", content, result.Content)
	}
	if result.TotalLines != 5 {
		t.Errorf("Expected 5 total lines, got %d", result.TotalLines)
	}
	if result.ReadLines != 5 {
		t.Errorf("Expected 5 read lines, got %d", result.ReadLines)
	}
	if result.StartLine != 1 {
		t.Errorf("Expected start line 1, got %d", result.StartLine)
	}
	if result.EndLine != 5 {
		t.Errorf("Expected end line 5, got %d", result.EndLine)
	}
	if result.Truncated {
		t.Error("Expected not truncated, got truncated")
	}
	if result.IsPartial {
		t.Error("Expected IsPartial to be false for full file read")
	}
	fileInfo, _ := os.Stat(tempFile)
	if result.FileSize != fileInfo.Size() {
		t.Errorf("Expected file size %d, got %d", fileInfo.Size(), result.FileSize)
	}
}

// Test reading a specific line range
func TestReadLineRange(t *testing.T) {
	content := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"
	tempFile, cleanup := createTempFile(t, content)
	defer cleanup()

	// Read lines 2-4
	result, err := ReadFile(tempFile, 2, 4, 1024)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify the result
	expectedContent := "Line 2\nLine 3\nLine 4"
	if result.Content != expectedContent {
		t.Errorf("Expected content: %q, got: %q", expectedContent, result.Content)
	}
	if result.TotalLines != 5 {
		t.Errorf("Expected 5 total lines, got %d", result.TotalLines)
	}
	if result.ReadLines != 3 {
		t.Errorf("Expected 3 read lines, got %d", result.ReadLines)
	}
	if result.StartLine != 2 {
		t.Errorf("Expected start line 2, got %d", result.StartLine)
	}
	if result.EndLine != 4 {
		t.Errorf("Expected end line 4, got %d", result.EndLine)
	}
	if !result.IsPartial {
		t.Error("Expected IsPartial to be true for line range read")
	}
}

// Test size limit enforcement - entire file too large
func TestOptimizedSizeLimitEnforcement(t *testing.T) {
	// Create a file just over the size limit
	content := strings.Repeat("X", 150)
	tempFile, cleanup := createTempFile(t, content)
	defer cleanup()

	// Try to read with size limit lower than file size
	result, err := ReadFile(tempFile, 0, 0, 100)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify we got truncated content
	if !result.Truncated {
		t.Error("Expected truncation due to size limit, got no truncation")
	}
	if len(result.Content) > 100 {
		t.Errorf("Expected content size <= 100, got %d", len(result.Content))
	}
}

// Test size limit enforcement with line-by-line reading
func TestSizeLimitWithLineReading(t *testing.T) {
	// Create content with lines of different lengths
	var contentBuilder strings.Builder
	for i := 1; i <= 10; i++ {
		contentBuilder.WriteString(strings.Repeat("A", i*10))
		contentBuilder.WriteString("\n")
	}
	content := contentBuilder.String()
	
	tempFile, cleanup := createTempFile(t, content)
	defer cleanup()

	// Set a small max size that will truncate after a few lines
	result, err := ReadFile(tempFile, 1, 0, 50)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify the truncation
	if !result.Truncated {
		t.Error("Expected truncation due to size limit, got no truncation")
	}
	if result.ReadLines >= 10 {
		t.Errorf("Expected less than 10 lines read due to size limit, got %d", result.ReadLines)
	}
	if len(result.Content) > 50 {
		t.Errorf("Expected content size <= 50, got %d", len(result.Content))
	}
}

// Test invalid parameters
func TestInvalidParameters(t *testing.T) {
	content := "Line 1\nLine 2\nLine 3"
	tempFile, cleanup := createTempFile(t, content)
	defer cleanup()

	tests := []struct {
		name      string
		startLine int
		endLine   int
		maxSize   int
		expectErr bool
	}{
		{"Negative start line", -5, 10, 1024, true},
		{"Negative end line", 1, -10, 1024, true},
		{"Start > End", 5, 3, 1024, true},
		{"Zero maxSize", 1, 5, 0, true},
		{"Negative maxSize", 1, 5, -100, true},
		{"Valid parameters", 1, 3, 1024, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ReadFile(tempFile, tc.startLine, tc.endLine, tc.maxSize)
			if (err != nil) != tc.expectErr {
				t.Errorf("Expected error: %v, got error: %v", tc.expectErr, err)
			}
		})
	}
}

// Test reading from a non-existent file
func TestNonExistentFile(t *testing.T) {
	_, err := ReadFile("/path/to/nonexistent/file", 1, 10, 1024)
	if err == nil {
		t.Error("Expected error for non-existent file, got none")
	}
}

// Test reading an empty file
func TestEmptyFile(t *testing.T) {
	tempFile, cleanup := createTempFile(t, "")
	defer cleanup()

	result, err := ReadFile(tempFile, 1, 10, 1024)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Content != "" {
		t.Errorf("Expected empty content, got: %q", result.Content)
	}
	if result.TotalLines != 0 {
		t.Errorf("Expected 0 total lines for empty file, got %d", result.TotalLines)
	}
	if result.ReadLines != 0 {
		t.Errorf("Expected 0 read lines for empty file, got %d", result.ReadLines)
	}
}

// Test with startLine exceeding file length
func TestStartLineExceedsLength(t *testing.T) {
	content := "Line 1\nLine 2\nLine 3"
	tempFile, cleanup := createTempFile(t, content)
	defer cleanup()

	result, err := ReadFile(tempFile, 10, 20, 1024)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should return empty content but no error
	if result.Content != "" {
		t.Errorf("Expected empty content when startLine exceeds file length, got: %q", result.Content)
	}
	if result.ReadLines != 0 {
		t.Errorf("Expected 0 read lines when startLine exceeds file length, got %d", result.ReadLines)
	}
	if result.TotalLines != 3 {
		t.Errorf("Expected 3 total lines, got %d", result.TotalLines)
	}
}

// Test with a file containing long lines
func TestLongLines(t *testing.T) {
	// Create a file with a few very long lines
	var contentBuilder strings.Builder
	for i := 0; i < 5; i++ {
		contentBuilder.WriteString(strings.Repeat("X", 1000))
		contentBuilder.WriteString("\n")
	}
	content := contentBuilder.String()
	
	tempFile, cleanup := createTempFile(t, content)
	defer cleanup()

	// Set max size to just over 2 lines worth
	result, err := ReadFile(tempFile, 1, 0, 2100)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should read about 2 lines
	if result.ReadLines != 2 {
		t.Errorf("Expected about 2 lines to be read with 2100 byte limit, got %d", result.ReadLines)
	}
	if !result.Truncated {
		t.Error("Expected truncation due to size limit on long lines")
	}
}

// Test reading file with Windows-style line endings (CRLF)
func TestWindowsLineEndings(t *testing.T) {
	// Create content with CRLF line endings
	content := "Line 1\r\nLine 2\r\nLine 3\r\nLine 4\r\nLine 5"
	tempFile, cleanup := createTempFile(t, content)
	defer cleanup()

	// Read the file
	result, err := ReadFile(tempFile, 1, 0, 1024)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify the line count is correct
	if result.TotalLines != 5 {
		t.Errorf("Expected 5 total lines with CRLF endings, got %d", result.TotalLines)
	}
	if result.ReadLines != 5 {
		t.Errorf("Expected 5 read lines with CRLF endings, got %d", result.ReadLines)
	}
}

// Test code file with special formatting
func TestCodeFileFormatting(t *testing.T) {
	code := `package main

import (
	"fmt"
)

func main() {
	// This is a comment with special indentation
	fmt.Println("Hello, World!")
	
	if true {
		fmt.Println("Indented code")
	}
}`
	tempFile, cleanup := createTempFile(t, code)
	defer cleanup()

	// Read the entire file
	result, err := ReadFile(tempFile, 0, 0, 1024)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify exact content preservation
	if result.Content != code {
		t.Errorf("Expected exact code content to be preserved\nExpected:\n%s\n\nGot:\n%s", code, result.Content)
	}
}

// Test optimized path for small files
func TestOptimizationPathUsed(t *testing.T) {
	// Create a file with mixed line endings to test exact preservation
	content := "Line 1\nLine 2\r\nLine 3\nLine 4\r\n"
	tempFile, cleanup := createTempFile(t, content)
	defer cleanup()

	// Directly compare with file reading method
	fileBytes, _ := os.ReadFile(tempFile)
	directContent := string(fileBytes)

	// Now use our optimized function
	result, err := ReadFile(tempFile, 0, 0, 1024)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify exact byte-for-byte content preservation
	if result.Content != directContent {
		t.Errorf("Expected optimized path to preserve exact content\nExpected:\n%q\n\nGot:\n%q", 
			directContent, result.Content)
	}
}

// Test count lines function
func TestCountLines(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		expected int
	}{
		{"Empty file", []byte{}, 0},
		{"Single line no newline", []byte("single line"), 1},
		{"Single line with newline", []byte("single line\n"), 1},
		{"Multiple lines", []byte("line 1\nline 2\nline 3"), 3},
		{"Multiple lines with final newline", []byte("line 1\nline 2\nline 3\n"), 3},
		{"Just newlines", []byte("\n\n\n"), 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			count := countLines(tc.content)
			if count != tc.expected {
				t.Errorf("Expected %d lines, got %d", tc.expected, count)
			}
		})
	}
}
