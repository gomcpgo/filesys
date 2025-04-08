package search

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTestFileForBefore creates a temporary test file with provided content for testing
func setupTestFileForBefore(t *testing.T, name, content string) (string, func()) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "regex-before-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	// Create the test file
	testFile := filepath.Join(tempDir, name)
	err = os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create test file: %v", err)
	}
	// Return a cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}
	return testFile, cleanup
}

// TestInsertBeforeFirstFunction tests inserting content before the first function in a file
func TestInsertBeforeFirstFunction(t *testing.T) {
	content := `
func firstFunction() {
	// Code here
}
func secondFunction() {
	// More code
}`
	expected := `
// New content here
func firstFunction() {
	// Code here
}
func secondFunction() {
	// More code
}`
	testFile, cleanup := setupTestFileForBefore(t, "functions.go", content)
	defer cleanup()
	result, err := InsertBeforeRegex(testFile, `func firstFunction\(\)`, "// New content here\n", 1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}

// TestInsertBeforeAllFunctions tests inserting content before all functions in a file
func TestInsertBeforeAllFunctions(t *testing.T) {
	content := `
func firstFunction() {
	// Code here
}
func secondFunction() {
	// More code
}`
	expected := `
// Comment before function
func firstFunction() {
	// Code here
}
// Comment before function
func secondFunction() {
	// More code
}`
	testFile, cleanup := setupTestFileForBefore(t, "all_functions.go", content)
	defer cleanup()
	result, err := InsertBeforeRegex(testFile, `func \w+\(\)`, "// Comment before function\n", 0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}

// TestInsertBeforeSpecificLine tests inserting content before a specific line
func TestInsertBeforeSpecificLine(t *testing.T) {
	content := `line 1
line 2
line 3
line 4`
	expected := `line 1
INSERTED CONTENT
line 2
line 3
line 4`
	testFile, cleanup := setupTestFileForBefore(t, "lines.txt", content)
	defer cleanup()
	result, err := InsertBeforeRegex(testFile, `line 2`, "INSERTED CONTENT\n", 1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}

// TestInsertBeforeMultipleOccurrences tests inserting content before a specific occurrence
// when multiple occurrences of the pattern exist
func TestInsertBeforeMultipleOccurrences(t *testing.T) {
	content := `repeated text
some other content
repeated text
more content
repeated text`
	expected := `repeated text
some other content
INSERTED - repeated text
more content
repeated text`
	testFile, cleanup := setupTestFileForBefore(t, "repeated.txt", content)
	defer cleanup()
	result, err := InsertBeforeRegex(testFile, `repeated text`, "INSERTED - ", 2)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}

// TestPatternNotFoundBefore tests handling of a pattern that doesn't exist in the file
func TestPatternNotFoundBefore(t *testing.T) {
	content := `This is some content
that doesn't contain the pattern`
	testFile, cleanup := setupTestFileForBefore(t, "no_pattern.txt", content)
	defer cleanup()
	_, err := InsertBeforeRegex(testFile, `non-existent pattern`, "INSERTED CONTENT\n", 1)
	if err == nil {
		t.Error("Expected an error but got none")
	}
}

// TestInvalidRegexPatternBefore tests handling of an invalid regex pattern
func TestInvalidRegexPatternBefore(t *testing.T) {
	content := `This is some content
with valid text`
	testFile, cleanup := setupTestFileForBefore(t, "invalid_regex.txt", content)
	defer cleanup()
	_, err := InsertBeforeRegex(testFile, `[invalid regex`, "INSERTED CONTENT\n", 1)
	if err == nil {
		t.Error("Expected an error but got none")
	}
}

// TestOccurrenceOutOfRangeBefore tests handling when the specified occurrence exceeds available matches
func TestOccurrenceOutOfRangeBefore(t *testing.T) {
	content := `pattern here
some other text
pattern here`
	testFile, cleanup := setupTestFileForBefore(t, "occurrences.txt", content)
	defer cleanup()
	_, err := InsertBeforeRegex(testFile, `pattern here`, "INSERTED CONTENT\n", 3)
	if err == nil {
		t.Error("Expected an error but got none")
	}
}

// TestEmptyFileBefore tests handling of an empty file
func TestEmptyFileBefore(t *testing.T) {
	content := ``
	testFile, cleanup := setupTestFileForBefore(t, "empty.txt", content)
	defer cleanup()
	_, err := InsertBeforeRegex(testFile, `pattern`, "INSERTED CONTENT", 1)
	if err == nil {
		t.Error("Expected an error but got none")
	}
}
