package search

import (
	"testing"
)

// TestDetectIndentation_SpaceIndented tests detecting space-based indentation
func TestDetectIndentation_SpaceIndented(t *testing.T) {
	content := `package main

import (
    "fmt"
    "os"
)`

	numSpaces, indentStr, err := DetectIndentation(content)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if numSpaces != 4 {
		t.Errorf("Expected 4 spaces, got %d", numSpaces)
	}

	if indentStr != "    " {
		t.Errorf("Expected 4 spaces as string, got %q", indentStr)
	}
}

// TestDetectIndentation_TabIndented tests detecting tab-based indentation
func TestDetectIndentation_TabIndented(t *testing.T) {
	content := `package main

import (
	"fmt"
	"os"
)`

	numSpaces, indentStr, err := DetectIndentation(content)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if numSpaces != 1 {
		t.Errorf("Expected 1 (tab), got %d", numSpaces)
	}

	if indentStr != "\t" {
		t.Errorf("Expected tab character, got %q", indentStr)
	}
}

// TestDetectIndentation_TwoSpaceIndentation tests detecting 2-space indentation
func TestDetectIndentation_TwoSpaceIndentation(t *testing.T) {
	content := `function test() {
  if (true) {
    console.log("hello");
  }
}`

	numSpaces, indentStr, err := DetectIndentation(content)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if numSpaces != 2 {
		t.Errorf("Expected 2 spaces, got %d", numSpaces)
	}

	if indentStr != "  " {
		t.Errorf("Expected 2 spaces, got %q", indentStr)
	}
}

// TestDetectIndentation_NoIndentation tests file with no indentation
func TestDetectIndentation_NoIndentation(t *testing.T) {
	content := `package main

func test() {
// No indent
}
`

	numSpaces, indentStr, err := DetectIndentation(content)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if numSpaces != 0 {
		t.Errorf("Expected 0, got %d", numSpaces)
	}

	if indentStr != "" {
		t.Errorf("Expected empty string, got %q", indentStr)
	}
}

// TestDetectIndentation_EmptyContent tests empty content
func TestDetectIndentation_EmptyContent(t *testing.T) {
	content := ""

	numSpaces, indentStr, err := DetectIndentation(content)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if numSpaces != 0 {
		t.Errorf("Expected 0, got %d", numSpaces)
	}

	if indentStr != "" {
		t.Errorf("Expected empty string, got %q", indentStr)
	}
}

// TestDetectIndentation_ContentWithoutLeadingWhitespace tests single-line content with no indent
func TestDetectIndentation_ContentWithoutLeadingWhitespace(t *testing.T) {
	content := "just some text"

	numSpaces, indentStr, err := DetectIndentation(content)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if numSpaces != 0 {
		t.Errorf("Expected 0, got %d", numSpaces)
	}

	if indentStr != "" {
		t.Errorf("Expected empty string, got %q", indentStr)
	}
}

// TestApplyIndentationToLines_SingleLine tests applying indent to single line
func TestApplyIndentationToLines_SingleLine(t *testing.T) {
	content := "hello world"
	indentStr := "    "

	result := ApplyIndentationToLines(content, indentStr)

	expected := "    hello world"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// TestApplyIndentationToLines_MultipleLines tests applying indent to multiple lines
func TestApplyIndentationToLines_MultipleLines(t *testing.T) {
	content := "line 1\nline 2\nline 3"
	indentStr := "  "

	result := ApplyIndentationToLines(content, indentStr)

	expected := "  line 1\n  line 2\n  line 3"
	if result != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, result)
	}
}

// TestApplyIndentationToLines_WithEmptyLines tests applying indent with empty lines preserved
func TestApplyIndentationToLines_WithEmptyLines(t *testing.T) {
	content := "line 1\n\nline 3"
	indentStr := "\t"

	result := ApplyIndentationToLines(content, indentStr)

	expected := "\tline 1\n\n\tline 3"
	if result != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, result)
	}
}

// TestApplyIndentationToLines_EmptyIndent tests applying empty indent (no-op)
func TestApplyIndentationToLines_EmptyIndent(t *testing.T) {
	content := "line 1\nline 2"
	indentStr := ""

	result := ApplyIndentationToLines(content, indentStr)

	if result != content {
		t.Errorf("Expected:\n%q\nGot:\n%q", content, result)
	}
}

// TestApplyIndentationToLines_EmptyContent tests applying indent to empty content
func TestApplyIndentationToLines_EmptyContent(t *testing.T) {
	content := ""
	indentStr := "    "

	result := ApplyIndentationToLines(content, indentStr)

	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}

// TestDetectIndentation_MixedIndentation tests detecting indentation with some mixed lines
func TestDetectIndentation_MixedIndentation(t *testing.T) {
	// File mostly uses 4 spaces but has some inconsistency
	content := `package main

import (
    "fmt"
  "os"
    "io"
)`

	numSpaces, indentStr, err := DetectIndentation(content)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should detect the most common indentation (4 spaces)
	if numSpaces != 4 {
		t.Errorf("Expected 4 spaces (most common), got %d", numSpaces)
	}

	if indentStr != "    " {
		t.Errorf("Expected 4 spaces, got %q", indentStr)
	}
}

// TestDetectIndentation_ComplexGoFile tests real-world Go file indentation detection
func TestDetectIndentation_ComplexGoFile(t *testing.T) {
	content := `package main

import "fmt"

func main() {
	if true {
		fmt.Println("hello")
	}
}

func helper() {
	return
}`

	numSpaces, indentStr, err := DetectIndentation(content)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if numSpaces != 1 {
		t.Errorf("Expected 1 (tab), got %d", numSpaces)
	}

	if indentStr != "\t" {
		t.Errorf("Expected tab, got %q", indentStr)
	}
}

// TestApplyIndentationToLines_CodeBlock tests applying indent to a code block
func TestApplyIndentationToLines_CodeBlock(t *testing.T) {
	content := `if condition {
    doSomething()
}`
	indentStr := "\t"

	result := ApplyIndentationToLines(content, indentStr)

	expected := "\tif condition {\n\t    doSomething()\n\t}"
	if result != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, result)
	}
}

// TestDetectIndentation_AllUnindentedContent tests content that's completely unindented
func TestDetectIndentation_AllUnindentedContent(t *testing.T) {
	content := `line 1
line 2
line 3
line 4`

	numSpaces, indentStr, err := DetectIndentation(content)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if numSpaces != 0 {
		t.Errorf("Expected 0, got %d", numSpaces)
	}

	if indentStr != "" {
		t.Errorf("Expected empty string, got %q", indentStr)
	}
}
