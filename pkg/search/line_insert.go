package search

import (
	"fmt"
	"os"
	"strings"
)

// InsertAfterLine inserts content after a specific line number in a file.
// Parameters:
//   - path: Path to the file
//   - lineNumber: Line number to insert after (1-based indexing)
//   - content: Content to insert
//   - autoIndent: If true, automatically indents inserted content to match the target line's indentation
//
// Returns:
//   - The new content with insertions
//   - Error if any
func InsertAfterLine(path string, lineNumber int, content string, autoIndent bool) (string, error) {
	// Read file content
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	fileContent := string(fileBytes)

	// Validate line number
	if lineNumber < 1 {
		return "", fmt.Errorf("line number must be at least 1, got %d", lineNumber)
	}

	// Split into lines
	lines := strings.Split(fileContent, "\n")
	totalLines := len(lines)

	// Handle empty file
	if totalLines == 0 || (totalLines == 1 && lines[0] == "") {
		return "", fmt.Errorf("cannot insert after line %d in empty file", lineNumber)
	}

	// Validate line number doesn't exceed total lines
	if lineNumber > totalLines {
		return "", fmt.Errorf("line number %d exceeds total lines %d", lineNumber, totalLines)
	}

	// Get the target line for indentation reference
	targetLine := lines[lineNumber-1]

	// Prepare content to insert
	insertContent := content
	if autoIndent {
		// Extract indentation from the target line
		indent := extractIndentation(targetLine)
		if indent != "" {
			insertContent = applyIndentation(content, indent)
		}
	}

	// Build new content
	var result []string
	for i, line := range lines {
		result = append(result, line)
		if i == lineNumber-1 {
			// Insert after this line
			result = append(result, insertContent)
		}
	}

	return strings.Join(result, "\n"), nil
}

// InsertBeforeLine inserts content before a specific line number in a file.
// Parameters:
//   - path: Path to the file
//   - lineNumber: Line number to insert before (1-based indexing)
//   - content: Content to insert
//   - autoIndent: If true, automatically indents inserted content to match the target line's indentation
//
// Returns:
//   - The new content with insertions
//   - Error if any
func InsertBeforeLine(path string, lineNumber int, content string, autoIndent bool) (string, error) {
	// Read file content
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	fileContent := string(fileBytes)

	// Validate line number
	if lineNumber < 1 {
		return "", fmt.Errorf("line number must be at least 1, got %d", lineNumber)
	}

	// Split into lines
	lines := strings.Split(fileContent, "\n")
	totalLines := len(lines)

	// Handle empty file
	if totalLines == 0 || (totalLines == 1 && lines[0] == "") {
		return "", fmt.Errorf("cannot insert before line %d in empty file", lineNumber)
	}

	// Validate line number doesn't exceed total lines
	if lineNumber > totalLines {
		return "", fmt.Errorf("line number %d exceeds total lines %d", lineNumber, totalLines)
	}

	// Get the target line for indentation reference
	targetLine := lines[lineNumber-1]

	// Prepare content to insert
	insertContent := content
	if autoIndent {
		// Extract indentation from the target line
		indent := extractIndentation(targetLine)
		if indent != "" {
			insertContent = applyIndentation(content, indent)
		}
	}

	// Build new content
	var result []string
	for i, line := range lines {
		if i == lineNumber-1 {
			// Insert before this line
			result = append(result, insertContent)
		}
		result = append(result, line)
	}

	return strings.Join(result, "\n"), nil
}

// extractIndentation returns the leading whitespace from a line
func extractIndentation(line string) string {
	var indent string
	for _, ch := range line {
		if ch == ' ' || ch == '\t' {
			indent += string(ch)
		} else {
			break
		}
	}
	return indent
}

// applyIndentation applies the given indentation to each line of content
func applyIndentation(content, indent string) string {
	lines := strings.Split(content, "\n")
	var indentedLines []string
	for _, line := range lines {
		if len(line) > 0 {
			indentedLines = append(indentedLines, indent+line)
		} else {
			indentedLines = append(indentedLines, "")
		}
	}
	return strings.Join(indentedLines, "\n")
}
