package search

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupLineInsertTestFile creates a temporary test file with provided content
func setupLineInsertTestFile(t *testing.T, name, content string) (string, func()) {
	tempDir, err := os.MkdirTemp("", "line-insert-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	testFile := filepath.Join(tempDir, name)
	err = os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create test file: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return testFile, cleanup
}

// TestInsertAfterLineBasic tests basic insertion after a specific line
func TestInsertAfterLineBasic(t *testing.T) {
	content := `line 1
line 2
line 3`

	testFile, cleanup := setupLineInsertTestFile(t, "after_line.txt", content)
	defer cleanup()

	result, err := InsertAfterLine(testFile, 2, "inserted line", false)
	if err != nil {
		t.Fatalf("InsertAfterLine failed: %v", err)
	}

	expected := `line 1
line 2
inserted line
line 3`

	if result != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}

// TestInsertAfterLineFirst tests insertion after the first line
func TestInsertAfterLineFirst(t *testing.T) {
	content := `first
second
third`

	testFile, cleanup := setupLineInsertTestFile(t, "after_first.txt", content)
	defer cleanup()

	result, err := InsertAfterLine(testFile, 1, "new", false)
	if err != nil {
		t.Fatalf("InsertAfterLine failed: %v", err)
	}

	expected := `first
new
second
third`

	if result != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}

// TestInsertAfterLineLast tests insertion after the last line
func TestInsertAfterLineLast(t *testing.T) {
	content := `first
second
third`

	testFile, cleanup := setupLineInsertTestFile(t, "after_last.txt", content)
	defer cleanup()

	result, err := InsertAfterLine(testFile, 3, "new", false)
	if err != nil {
		t.Fatalf("InsertAfterLine failed: %v", err)
	}

	expected := `first
second
third
new`

	if result != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}

// TestInsertAfterLineWithAutoIndent tests autoIndent preserves line indentation
func TestInsertAfterLineWithAutoIndent(t *testing.T) {
	content := `func example() {
	line1
	line2
}`

	testFile, cleanup := setupLineInsertTestFile(t, "after_indent.go", content)
	defer cleanup()

	// Insert after line 2 (which has tab indentation)
	result, err := InsertAfterLine(testFile, 2, "newLine", true)
	if err != nil {
		t.Fatalf("InsertAfterLine with autoIndent failed: %v", err)
	}

	// The inserted line should have the same indentation as line 2
	if !strings.Contains(result, "\tnewLine") {
		t.Errorf("Expected indented 'newLine', got:\n%s", result)
	}
}

// TestInsertAfterLineInvalidLineNumber tests error handling for invalid line numbers
func TestInsertAfterLineInvalidLineNumber(t *testing.T) {
	content := `line 1
line 2`

	testFile, cleanup := setupLineInsertTestFile(t, "invalid_line.txt", content)
	defer cleanup()

	// Line 0 is invalid
	_, err := InsertAfterLine(testFile, 0, "content", false)
	if err == nil {
		t.Errorf("Expected error for line 0, but got none")
	}

	// Line 5 exceeds total lines
	_, err = InsertAfterLine(testFile, 5, "content", false)
	if err == nil {
		t.Errorf("Expected error for line 5 (exceeds total), but got none")
	}
}

// TestInsertBeforeLineBasic tests basic insertion before a specific line
func TestInsertBeforeLineBasic(t *testing.T) {
	content := `line 1
line 2
line 3`

	testFile, cleanup := setupLineInsertTestFile(t, "before_line.txt", content)
	defer cleanup()

	result, err := InsertBeforeLine(testFile, 2, "inserted line", false)
	if err != nil {
		t.Fatalf("InsertBeforeLine failed: %v", err)
	}

	expected := `line 1
inserted line
line 2
line 3`

	if result != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}

// TestInsertBeforeLineFirst tests insertion before the first line
func TestInsertBeforeLineFirst(t *testing.T) {
	content := `first
second
third`

	testFile, cleanup := setupLineInsertTestFile(t, "before_first.txt", content)
	defer cleanup()

	result, err := InsertBeforeLine(testFile, 1, "new", false)
	if err != nil {
		t.Fatalf("InsertBeforeLine failed: %v", err)
	}

	expected := `new
first
second
third`

	if result != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}

// TestInsertBeforeLineLast tests insertion before the last line
func TestInsertBeforeLineLast(t *testing.T) {
	content := `first
second
third`

	testFile, cleanup := setupLineInsertTestFile(t, "before_last.txt", content)
	defer cleanup()

	result, err := InsertBeforeLine(testFile, 3, "new", false)
	if err != nil {
		t.Fatalf("InsertBeforeLine failed: %v", err)
	}

	expected := `first
second
new
third`

	if result != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}

// TestInsertBeforeLineWithAutoIndent tests autoIndent preserves line indentation
func TestInsertBeforeLineWithAutoIndent(t *testing.T) {
	content := `func example() {
	line1
	line2
}`

	testFile, cleanup := setupLineInsertTestFile(t, "before_indent.go", content)
	defer cleanup()

	// Insert before line 3 (line2 with tab indentation)
	result, err := InsertBeforeLine(testFile, 3, "newLine", true)
	if err != nil {
		t.Fatalf("InsertBeforeLine with autoIndent failed: %v", err)
	}

	// The inserted line should have the same indentation as line 3
	if !strings.Contains(result, "\tnewLine\n\tline2") {
		t.Errorf("Expected indented 'newLine' before 'line2', got:\n%s", result)
	}
}

// TestInsertBeforeLineInvalidLineNumber tests error handling for invalid line numbers
func TestInsertBeforeLineInvalidLineNumber(t *testing.T) {
	content := `line 1
line 2`

	testFile, cleanup := setupLineInsertTestFile(t, "invalid_before.txt", content)
	defer cleanup()

	// Line 0 is invalid
	_, err := InsertBeforeLine(testFile, 0, "content", false)
	if err == nil {
		t.Errorf("Expected error for line 0, but got none")
	}

	// Line 5 exceeds total lines
	_, err = InsertBeforeLine(testFile, 5, "content", false)
	if err == nil {
		t.Errorf("Expected error for line 5 (exceeds total), but got none")
	}
}

// TestInsertAfterLineMultilineContent tests inserting multiple lines of content
func TestInsertAfterLineMultilineContent(t *testing.T) {
	content := `line 1
line 2`

	testFile, cleanup := setupLineInsertTestFile(t, "multiline_content.txt", content)
	defer cleanup()

	multilineContent := `new line A
new line B
new line C`

	result, err := InsertAfterLine(testFile, 1, multilineContent, false)
	if err != nil {
		t.Fatalf("InsertAfterLine with multiline content failed: %v", err)
	}

	expected := `line 1
new line A
new line B
new line C
line 2`

	if result != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}

// TestInsertBeforeLineMultilineContent tests inserting multiple lines of content
func TestInsertBeforeLineMultilineContent(t *testing.T) {
	content := `line 1
line 2`

	testFile, cleanup := setupLineInsertTestFile(t, "before_multiline.txt", content)
	defer cleanup()

	multilineContent := `new line A
new line B`

	result, err := InsertBeforeLine(testFile, 2, multilineContent, false)
	if err != nil {
		t.Fatalf("InsertBeforeLine with multiline content failed: %v", err)
	}

	expected := `line 1
new line A
new line B
line 2`

	if result != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}

// TestInsertAfterLineEmptyFile tests error handling for empty files
func TestInsertAfterLineEmptyFile(t *testing.T) {
	content := ``

	testFile, cleanup := setupLineInsertTestFile(t, "empty.txt", content)
	defer cleanup()

	_, err := InsertAfterLine(testFile, 1, "content", false)
	if err == nil {
		t.Errorf("Expected error for empty file, but got none")
	}
}

// TestInsertAfterLineSingleLine tests inserting after a single-line file
func TestInsertAfterLineSingleLine(t *testing.T) {
	content := `only line`

	testFile, cleanup := setupLineInsertTestFile(t, "single.txt", content)
	defer cleanup()

	result, err := InsertAfterLine(testFile, 1, "new line", false)
	if err != nil {
		t.Fatalf("InsertAfterLine on single-line file failed: %v", err)
	}

	expected := `only line
new line`

	if result != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}
