package search

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTestFile creates a temporary test file with provided content for testing
func setupTestFile(t *testing.T, name, content string) (string, func()) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "regex-test")
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

// TestInsertAfterFirstFunction tests inserting content after the first function in a file
func TestInsertAfterFirstFunction(t *testing.T) {
	content := `
func firstFunction() {
	// Code here
}

func secondFunction() {
	// More code
}`

	expected := `
func firstFunction() {
	// Code here
}

// New content here

func secondFunction() {
	// More code
}`

	testFile, cleanup := setupTestFile(t, "functions.go", content)
	defer cleanup()

	result, err := InsertAfterRegex(testFile, `func \w+\(\) {[\s\S]*?}`, "\n\n// New content here", 1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}

// TestInsertAfterAllFunctions tests inserting content after all functions in a file
func TestInsertAfterAllFunctions(t *testing.T) {
	content := `
func firstFunction() {
	// Code here
}

func secondFunction() {
	// More code
}`

	expected := `
func firstFunction() {
	// Code here
}
// Comment after function

func secondFunction() {
	// More code
}
// Comment after function`

	testFile, cleanup := setupTestFile(t, "all_functions.go", content)
	defer cleanup()

	result, err := InsertAfterRegex(testFile, `func \w+\(\) {[\s\S]*?}`, "\n// Comment after function", 0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}

// TestInsertAfterSpecificLine tests inserting content after a specific line
func TestInsertAfterSpecificLine(t *testing.T) {
	content := `line 1
line 2
line 3
line 4`

	expected := `line 1
line 2
INSERTED CONTENT
line 3
line 4`

	testFile, cleanup := setupTestFile(t, "lines.txt", content)
	defer cleanup()

	result, err := InsertAfterRegex(testFile, `line 2`, "\nINSERTED CONTENT", 1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}

// TestInsertAfterMultipleOccurrences tests inserting content after a specific occurrence
// when multiple occurrences of the pattern exist
func TestInsertAfterMultipleOccurrences(t *testing.T) {
	content := `repeated text
some other content
repeated text
more content
repeated text`

	expected := `repeated text
some other content
repeated text - INSERTED
more content
repeated text`

	testFile, cleanup := setupTestFile(t, "repeated.txt", content)
	defer cleanup()

	result, err := InsertAfterRegex(testFile, `repeated text`, " - INSERTED", 2)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}

// TestPatternNotFound tests handling of a pattern that doesn't exist in the file
func TestPatternNotFound(t *testing.T) {
	content := `This is some content
that doesn't contain the pattern`

	testFile, cleanup := setupTestFile(t, "no_pattern.txt", content)
	defer cleanup()

	_, err := InsertAfterRegex(testFile, `non-existent pattern`, "\nINSERTED CONTENT", 1)
	if err == nil {
		t.Error("Expected an error but got none")
	}
}

// TestInvalidRegexPattern tests handling of an invalid regex pattern
func TestInvalidRegexPattern2(t *testing.T) {
	content := `This is some content
with valid text`

	testFile, cleanup := setupTestFile(t, "invalid_regex.txt", content)
	defer cleanup()

	_, err := InsertAfterRegex(testFile, `[invalid regex`, "\nINSERTED CONTENT", 1)
	if err == nil {
		t.Error("Expected an error but got none")
	}
}

// TestOccurrenceOutOfRange tests handling when the specified occurrence exceeds available matches
func TestOccurrenceOutOfRange(t *testing.T) {
	content := `pattern here
some other text
pattern here`

	testFile, cleanup := setupTestFile(t, "occurrences.txt", content)
	defer cleanup()

	_, err := InsertAfterRegex(testFile, `pattern here`, "\nINSERTED CONTENT", 3)
	if err == nil {
		t.Error("Expected an error but got none")
	}
}

// TestEmptyFile tests handling of an empty file
func TestEmptyFile(t *testing.T) {
	content := ``

	testFile, cleanup := setupTestFile(t, "empty.txt", content)
	defer cleanup()

	_, err := InsertAfterRegex(testFile, `pattern`, "INSERTED CONTENT", 1)
	if err == nil {
		t.Error("Expected an error but got none")
	}
}

// TestInsertWithLookbehind tests inserting at the beginning with look-behind pattern
func TestInsertWithLookbehind(t *testing.T) {
	content := `package main

import "fmt"

func main() {
    fmt.Println("Hello")
}`

	expected := `package main
// Project: Example

import "fmt"

func main() {
    fmt.Println("Hello")
}`

	testFile, cleanup := setupTestFile(t, "package.go", content)
	defer cleanup()

	result, err := InsertAfterRegex(testFile, `^package .*`, "\n// Project: Example", 1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}

// TestInsertingEmptyContent tests inserting empty content after a match
func TestInsertingEmptyContent(t *testing.T) {
	content := `Function one
Function two
Function three`

	expected := `Function one
Function two
Function three`

	testFile, cleanup := setupTestFile(t, "empty_insert.txt", content)
	defer cleanup()

	result, err := InsertAfterRegex(testFile, `Function two`, "", 1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}

// TestComplexRegexWithCaptureGroups tests using complex regex with capture groups
func TestComplexRegexWithCaptureGroups(t *testing.T) {
	content := `<div class="container">
  <h1>Title</h1>
  <p>Paragraph 1</p>
  <p>Paragraph 2</p>
</div>`

	expected := `<div class="container">
  <h1>Title</h1>
  <p>Paragraph 1</p>
  <!-- Comment -->
  <p>Paragraph 2</p>
</div>`

	testFile, cleanup := setupTestFile(t, "complex.html", content)
	defer cleanup()

	result, err := InsertAfterRegex(testFile, `<p>(.*?)</p>`, "\n  <!-- Comment -->", 1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}

// TestUnicodeCharacters tests handling unicode characters in the file and regex
func TestUnicodeCharacters(t *testing.T) {
	content := `function greet() {
  console.log("你好，世界");
}

function farewell() {
  console.log("再见");
}`

	expected := `function greet() {
  console.log("你好，世界");
  console.log("Processing...");
}

function farewell() {
  console.log("再见");
}`

	testFile, cleanup := setupTestFile(t, "unicode.js", content)
	defer cleanup()

	// Note: We're now using an actual newline character followed by spaces
	// instead of the escape sequence \n in the insert content
	result, err := InsertAfterRegex(testFile, `console\.log\("你好，世界"\);`, "\n  console.log(\"Processing...\");", 1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}

// TestInsertAtLastOccurrence tests inserting content after the last occurrence
func TestInsertAtLastOccurrence(t *testing.T) {
	content := `Example 1
Example 2
Example 3`

	expected := `Example 1
Example 2
Example 3 - Last`

	testFile, cleanup := setupTestFile(t, "last_occurrence.txt", content)
	defer cleanup()

	result, err := InsertAfterRegex(testFile, `Example \d`, " - Last", 3)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}

// TestOverlappingPatterns tests handling overlapping patterns
func TestOverlappingPatterns(t *testing.T) {
	content := `aaaa`

	expected := `aaXaaX`

	testFile, cleanup := setupTestFile(t, "overlapping.txt", content)
	defer cleanup()

	result, err := InsertAfterRegex(testFile, `aa`, "X", 0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}
