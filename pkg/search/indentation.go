package search

import (
	"strings"
)

// DetectIndentation analyzes the content and detects the indentation pattern used.
// Returns:
//   - numSpaces: number of spaces (or 1 for tab)
//   - indentStr: the actual indentation string (\t or space sequence)
//   - error: if detection fails
func DetectIndentation(content string) (int, string, error) {
	if len(content) == 0 {
		return 0, "", nil
	}

	lines := strings.Split(content, "\n")
	var indentCounts map[int]int     // space count -> frequency
	var tabCount int                 // count of tab-indented lines
	var detectedIndentStr string     // the actual indent string

	indentCounts = make(map[int]int)

	// Analyze each line to find indentation patterns
	for _, line := range lines {
		if len(line) == 0 {
			continue // Skip empty lines
		}

		// Skip lines that don't start with whitespace
		if line[0] != ' ' && line[0] != '\t' {
			continue
		}

		// Check for tabs
		if line[0] == '\t' {
			tabCount++
			continue
		}

		// Count leading spaces
		spaceCount := 0
		for _, ch := range line {
			if ch == ' ' {
				spaceCount++
			} else {
				break
			}
		}

		if spaceCount > 0 {
			indentCounts[spaceCount]++
		}
	}

	// Determine if tabs or spaces are used
	if tabCount > 0 && len(indentCounts) == 0 {
		// File uses tabs exclusively
		return 1, "\t", nil
	}

	if len(indentCounts) == 0 {
		// No indentation found
		return 0, "", nil
	}

	// Find the most common space indentation
	maxCount := 0
	mostCommonIndent := 0
	for spaces, count := range indentCounts {
		if count > maxCount {
			maxCount = count
			mostCommonIndent = spaces
		}
	}

	detectedIndentStr = strings.Repeat(" ", mostCommonIndent)
	return mostCommonIndent, detectedIndentStr, nil
}

// ApplyIndentationToLines applies indentation to all lines in the content.
// Empty lines are preserved without indentation.
func ApplyIndentationToLines(content, indentStr string) string {
	if len(content) == 0 || len(indentStr) == 0 {
		return content
	}

	lines := strings.Split(content, "\n")
	var result []string

	for _, line := range lines {
		if len(line) == 0 {
			// Preserve empty lines without indentation
			result = append(result, "")
		} else {
			result = append(result, indentStr+line)
		}
	}

	return strings.Join(result, "\n")
}
