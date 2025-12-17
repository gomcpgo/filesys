package search

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupMultilineTestFile creates a temporary test file with provided content
func setupMultilineTestFile(t *testing.T, name, content string) (string, func()) {
	tempDir, err := os.MkdirTemp("", "multiline-test")
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

// TestInsertAfterRegexMultilineDotMatchesNewline tests that with multiline=true,
// the dot (.) matches newlines, allowing patterns to span multiple lines
func TestInsertAfterRegexMultilineDotMatchesNewline(t *testing.T) {
	content := `func example() {
	line1
	line2
}

func other() {}`

	// Pattern that spans multiple lines - dot should match newlines with multiline=true
	// Use non-greedy .+? to match only the first closing brace
	pattern := `func example\(\) \{.+?\}`

	testFile, cleanup := setupMultilineTestFile(t, "multiline_dot.go", content)
	defer cleanup()

	// Without multiline, this pattern would fail to match (dot doesn't match newlines)
	// With multiline=true, it should match the entire function block
	result, err := InsertAfterRegexMultiline(testFile, pattern, "\n// inserted after", 1, false, true)
	if err != nil {
		t.Fatalf("InsertAfterRegexMultiline failed: %v", err)
	}

	if !strings.Contains(result, "// inserted after") {
		t.Errorf("Expected inserted content, got:\n%s", result)
	}

	// Verify the insertion is after the closing brace of example()
	expectedPart := `}
// inserted after

func other()`
	if !strings.Contains(result, expectedPart) {
		t.Errorf("Expected content after function block, got:\n%s", result)
	}
}

// TestInsertAfterRegexMultilineCaretDollarMatchLineEnds tests that with multiline=true,
// ^ and $ match at line boundaries, not just string boundaries
func TestInsertAfterRegexMultilineCaretDollarMatchLineEnds(t *testing.T) {
	content := `first line
second line
third line`

	// Pattern using ^ to match line start
	pattern := `^second`

	testFile, cleanup := setupMultilineTestFile(t, "multiline_caret.txt", content)
	defer cleanup()

	// With multiline=true, ^ should match the start of "second line"
	result, err := InsertAfterRegexMultiline(testFile, pattern, " MODIFIED", 1, false, true)
	if err != nil {
		t.Fatalf("InsertAfterRegexMultiline failed: %v", err)
	}

	if !strings.Contains(result, "second MODIFIED") {
		t.Errorf("Expected 'second MODIFIED', got:\n%s", result)
	}
}

// TestInsertBeforeRegexMultilineDotMatchesNewline tests InsertBeforeRegex with multiline
func TestInsertBeforeRegexMultilineDotMatchesNewline(t *testing.T) {
	content := `package main

func main() {
	// main code
}

func helper() {
	// helper code
}`

	// Pattern to match entire helper function using multiline dot
	pattern := `func helper\(\) \{.+\}`

	testFile, cleanup := setupMultilineTestFile(t, "before_multiline.go", content)
	defer cleanup()

	result, err := InsertBeforeRegexMultiline(testFile, pattern, "// before helper\n", 1, false, true)
	if err != nil {
		t.Fatalf("InsertBeforeRegexMultiline failed: %v", err)
	}

	if !strings.Contains(result, "// before helper\nfunc helper()") {
		t.Errorf("Expected content before helper function, got:\n%s", result)
	}
}

// TestReplaceWithRegexMultilineDotMatchesNewline tests ReplaceWithRegex with multiline
func TestReplaceWithRegexMultilineDotMatchesNewline(t *testing.T) {
	content := `func old() {
	oldCode1
	oldCode2
}

func keep() {}`

	// Pattern to match entire old function
	pattern := `func old\(\) \{.+?\}`

	replacement := `func new() {
	newCode
}`

	testFile, cleanup := setupMultilineTestFile(t, "replace_multiline.go", content)
	defer cleanup()

	result, count, err := ReplaceWithRegexMultiline(testFile, pattern, replacement, 0, true, true)
	if err != nil {
		t.Fatalf("ReplaceWithRegexMultiline failed: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 replacement, got %d", count)
	}

	if !strings.Contains(result, "func new()") {
		t.Errorf("Expected replaced function, got:\n%s", result)
	}

	if strings.Contains(result, "func old()") {
		t.Errorf("Old function should have been replaced, got:\n%s", result)
	}
}

// TestReplaceWithRegexMultilineCaretDollar tests ^ and $ with multiline mode
func TestReplaceWithRegexMultilineCaretDollar(t *testing.T) {
	content := `line one
line two
line three`

	// Replace all lines starting with "line "
	pattern := `^line `
	replacement := `item `

	testFile, cleanup := setupMultilineTestFile(t, "caret_dollar.txt", content)
	defer cleanup()

	result, count, err := ReplaceWithRegexMultiline(testFile, pattern, replacement, 0, true, true)
	if err != nil {
		t.Fatalf("ReplaceWithRegexMultiline failed: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected 3 replacements, got %d", count)
	}

	expected := `item one
item two
item three`

	if result != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}

// TestMultilineFalseDoesNotMatchAcrossLines verifies default behavior (multiline=false)
func TestMultilineFalseDoesNotMatchAcrossLines(t *testing.T) {
	content := `func example() {
	line1
}`

	// This pattern should NOT match when multiline=false because dot doesn't match newlines
	pattern := `func example\(\) \{.+\}`

	testFile, cleanup := setupMultilineTestFile(t, "no_multiline.go", content)
	defer cleanup()

	_, err := InsertAfterRegexMultiline(testFile, pattern, "// should not insert", 1, false, false)

	// Should fail because pattern doesn't match (dot doesn't cross lines)
	if err == nil {
		t.Errorf("Expected error when multiline=false and pattern spans lines, but got none")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

// TestMultilineWithAutoIndent tests that autoIndent works correctly with multiline patterns
func TestMultilineWithAutoIndent(t *testing.T) {
	content := `package main

	func indented() {
		code
	}

	func another() {}`

	pattern := `func indented\(\) \{.+?\}`

	testFile, cleanup := setupMultilineTestFile(t, "multiline_indent.go", content)
	defer cleanup()

	result, err := InsertAfterRegexMultiline(testFile, pattern, "\n// comment", 1, true, true)
	if err != nil {
		t.Fatalf("InsertAfterRegexMultiline with autoIndent failed: %v", err)
	}

	// Should have inserted with proper indentation
	if !strings.Contains(result, "// comment") {
		t.Errorf("Expected inserted comment, got:\n%s", result)
	}
}
