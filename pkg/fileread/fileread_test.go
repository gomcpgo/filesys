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

// Test reading an entire small file
func TestReadEntireSmallFile(t *testing.T) {
	content := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"
	tempFile, cleanup := createTempFile(t, content)
	defer cleanup()

	// Read the entire file
	result, err := ReadFileLines(tempFile, 0, 0, 1024)
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
	result, err := ReadFileLines(tempFile, 2, 4, 1024)
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
}

// Test size limit enforcement
func TestSizeLimitEnforcement(t *testing.T) {
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
	result, err := ReadFileLines(tempFile, 1, 0, 50)
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
			_, err := ReadFileLines(tempFile, tc.startLine, tc.endLine, tc.maxSize)
			if (err != nil) != tc.expectErr {
				t.Errorf("Expected error: %v, got error: %v", tc.expectErr, err)
			}
		})
	}
}

// Test reading from a non-existent file
func TestNonExistentFile(t *testing.T) {
	_, err := ReadFileLines("/path/to/nonexistent/file", 1, 10, 1024)
	if err == nil {
		t.Error("Expected error for non-existent file, got none")
	}
}

// Test reading an empty file
func TestEmptyFile(t *testing.T) {
	tempFile, cleanup := createTempFile(t, "")
	defer cleanup()

	result, err := ReadFileLines(tempFile, 1, 10, 1024)
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

	result, err := ReadFileLines(tempFile, 10, 20, 1024)
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
	result, err := ReadFileLines(tempFile, 1, 0, 2100)
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
	result, err := ReadFileLines(tempFile, 1, 0, 1024)
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
