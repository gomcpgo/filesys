package search

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInsertAfterRegex(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "regex-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up after tests

	// Test cases
	tests := []struct {
		name           string
		fileContent    string
		pattern        string
		insertContent  string
		occurrence     int
		expectedResult string
		expectError    bool
	}{
		{
			name: "Insert after first function",
			fileContent: `
func firstFunction() {
	// Code here
}

func secondFunction() {
	// More code
}`,
			pattern:        `func \w+\(\) {[\s\S]*?}`,
			insertContent:  "\n\n// New content here",
			occurrence:     1,
			expectedResult: `
func firstFunction() {
	// Code here
}

// New content here

func secondFunction() {
	// More code
}`,
			expectError: false,
		},
		{
			name: "Insert after all functions",
			fileContent: `
func firstFunction() {
	// Code here
}

func secondFunction() {
	// More code
}`,
			pattern:        `func \w+\(\) {[\s\S]*?}`,
			insertContent:  "\n// Comment after function",
			occurrence:     0, // 0 means all occurrences
			expectedResult: `
func firstFunction() {
	// Code here
}
// Comment after function

func secondFunction() {
	// More code
}
// Comment after function`,
			expectError: false,
		},
		{
			name: "Insert after specific line",
			fileContent: `line 1
line 2
line 3
line 4`,
			pattern:        `line 2`,
			insertContent:  "\nINSERTED CONTENT",
			occurrence:     1,
			expectedResult: `line 1
line 2
INSERTED CONTENT
line 3
line 4`,
			expectError: false,
		},
		{
			name: "Multiple occurrences of pattern",
			fileContent: `repeated text
some other content
repeated text
more content
repeated text`,
			pattern:        `repeated text`,
			insertContent:  " - INSERTED",
			occurrence:     2,
			expectedResult: `repeated text
some other content
repeated text - INSERTED
more content
repeated text`,
			expectError: false,
		},
		{
			name: "Pattern not found",
			fileContent: `This is some content
that doesn't contain the pattern`,
			pattern:        `non-existent pattern`,
			insertContent:  "\nINSERTED CONTENT",
			occurrence:     1,
			expectedResult: ``,
			expectError:    true,
		},
		{
			name: "Invalid regex pattern",
			fileContent: `This is some content
with valid text`,
			pattern:        `[invalid regex`,
			insertContent:  "\nINSERTED CONTENT",
			occurrence:     1,
			expectedResult: ``,
			expectError:    true,
		},
		{
			name: "Occurrence out of range",
			fileContent: `pattern here
some other text
pattern here`,
			pattern:        `pattern here`,
			insertContent:  "\nINSERTED CONTENT",
			occurrence:     3, // There are only 2 occurrences
			expectedResult: ``,
			expectError:    true,
		},
		{
			name: "Empty file",
			fileContent: ``,
			pattern:        `pattern`,
			insertContent:  "INSERTED CONTENT",
			occurrence:     1,
			expectedResult: ``,
			expectError:    true,
		},
		{
			name: "Insert at beginning with look-behind",
			fileContent: `package main

import "fmt"

func main() {
    fmt.Println("Hello")
}`,
			pattern:       `^package .*`,
			insertContent: "\n// Project: Example",
			occurrence:    1,
			expectedResult: `package main
// Project: Example

import "fmt"

func main() {
    fmt.Println("Hello")
}`,
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test file
			testFile := filepath.Join(tempDir, "test-"+strings.ReplaceAll(tc.name, " ", "-")+".txt")
			err := os.WriteFile(testFile, []byte(tc.fileContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Call the function
			result, err := InsertAfterRegex(testFile, tc.pattern, tc.insertContent, tc.occurrence)

			// Check errors
			if tc.expectError && err == nil {
				t.Errorf("Expected an error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Did not expect an error but got: %v", err)
			}

			// If no error is expected, validate the result
			if !tc.expectError {
				if result != tc.expectedResult {
					t.Errorf("Expected result:\n%s\n\nGot:\n%s", tc.expectedResult, result)
				}
			}
		})
	}
}

// TestInsertAfterRegexEdgeCases tests edge cases and more complex scenarios
func TestInsertAfterRegexEdgeCases(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "regex-edge-cases")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up after tests

	// Test cases for edge cases
	tests := []struct {
		name           string
		fileContent    string
		pattern        string
		insertContent  string
		occurrence     int
		expectedResult string
		expectError    bool
	}{
		{
			name: "Inserting empty content",
			fileContent: `Function one
Function two
Function three`,
			pattern:        `Function two`,
			insertContent:  "",
			occurrence:     1,
			expectedResult: `Function one
Function two
Function three`,
			expectError: false,
		},
		{
			name: "Complex regex with capture groups",
			fileContent: `<div class="container">
  <h1>Title</h1>
  <p>Paragraph 1</p>
  <p>Paragraph 2</p>
</div>`,
			pattern:        `<p>(.*?)</p>`,
			insertContent:  "\n  <!-- Comment -->",
			occurrence:     1,
			expectedResult: `<div class="container">
  <h1>Title</h1>
  <p>Paragraph 1</p>
  <!-- Comment -->
  <p>Paragraph 2</p>
</div>`,
			expectError: false,
		},
		{
			name: "Unicode characters",
			fileContent: `function greet() {
  console.log("你好，世界");
}

function farewell() {
  console.log("再见");
}`,
			pattern:        `console\.log\("你好，世界"\);`,
			insertContent:  `\n  console.log("Processing...");`,
			occurrence:     1,
			expectedResult: `function greet() {
  console.log("你好，世界");
  console.log("Processing...");
}

function farewell() {
  console.log("再见");
}`,
			expectError: false,
		},
		{
			name: "Insert at last occurrence",
			fileContent: `Example 1
Example 2
Example 3`,
			pattern:        `Example \d`,
			insertContent:  " - Last",
			occurrence:     3,
			expectedResult: `Example 1
Example 2
Example 3 - Last`,
			expectError: false,
		},
		{
			name: "Overlapping patterns",
			fileContent: `aaaa`,
			pattern:        `aa`,
			insertContent:  "X",
			occurrence:     0, // All occurrences
			expectedResult: `aaXaaX`, // Should match at positions 0 and 2
			expectError:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test file
			testFile := filepath.Join(tempDir, "test-"+strings.ReplaceAll(tc.name, " ", "-")+".txt")
			err := os.WriteFile(testFile, []byte(tc.fileContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Call the function
			result, err := InsertAfterRegex(testFile, tc.pattern, tc.insertContent, tc.occurrence)

			// Check errors
			if tc.expectError && err == nil {
				t.Errorf("Expected an error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Did not expect an error but got: %v", err)
			}

			// If no error is expected, validate the result
			if !tc.expectError {
				if result != tc.expectedResult {
					t.Errorf("Expected result:\n%s\n\nGot:\n%s", tc.expectedResult, result)
				}
			}
		})
	}
}
