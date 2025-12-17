package search

import (
	"fmt"
	"os"
	"regexp"
)

// InsertBeforeRegex inserts content before a specific occurrence of a regex pattern in a file.
// Parameters:
//   - path: Path to the file
//   - pattern: Regular expression pattern to search for
//   - content: Content to insert before the pattern match
//   - occurrence: Which occurrence to insert before (1-based indexing, 0 means all occurrences)
//   - autoIndent: If true, automatically indents inserted content to match surrounding code
//
// Returns:
//   - The new content with insertions
//   - Error if any
func InsertBeforeRegex(path, pattern, content string, occurrence int, autoIndent bool) (string, error) {
	// Read file content
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	fileContent := string(fileBytes)

	// Compile the regular expression
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("invalid regex pattern: %w", err)
	}

	// Find all matches
	indexes := re.FindAllStringIndex(fileContent, -1)
	if indexes == nil || len(indexes) == 0 {
		return "", fmt.Errorf("pattern '%s' not found in file", pattern)
	}

	totalMatches := len(indexes)

	// Check if requested occurrence exists
	if occurrence > totalMatches {
		return "", fmt.Errorf("specified occurrence %d exceeds total matches %d",
			occurrence, totalMatches)
	}

	// Build the new content with insertions
	var newContent string
	lastIndex := 0

	// Handle the "all occurrences" case or specific occurrence
	for i, idx := range indexes {
		matchStart := idx[0] // Start index of the match
		matchEnd := idx[1]   // End index of the match

		// If targeting specific occurrence and this isn't it, skip insertion
		if occurrence > 0 && i+1 != occurrence {
			continue
		}

		// Prepare the content to insert
		insertContent := content

		// If autoIndent is enabled, we need to insert on a new line with proper indentation
		if autoIndent {
			// Find the line containing the match to extract its indentation
			lineStart := matchStart
			for lineStart > 0 && fileContent[lineStart-1] != '\n' {
				lineStart--
			}

			// Extract the indentation of the current line
			lineIndent := ""
			for j := lineStart; j < len(fileContent) && (fileContent[j] == ' ' || fileContent[j] == '\t'); j++ {
				lineIndent += string(fileContent[j])
			}

			// For autoIndent, we insert content as new line(s) before the current line
			// First, add everything up to the line start
			if lastIndex <= lineStart {
				newContent += fileContent[lastIndex:lineStart]
			}

			// Apply indentation to each line of the content
			if lineIndent != "" {
				lines := splitLines(content)
				var indentedLines []string
				for _, line := range lines {
					if len(line) > 0 {
						indentedLines = append(indentedLines, lineIndent+line)
					} else {
						indentedLines = append(indentedLines, "")
					}
				}
				insertContent = joinLines(indentedLines)
				// Ensure trailing newline for proper line separation
				if len(insertContent) > 0 && insertContent[len(insertContent)-1] != '\n' {
					insertContent += "\n"
				}
			}

			// Add the indented content
			newContent += insertContent

			// Add the original line (from lineStart to matchEnd, which includes indentation and match)
			newContent += fileContent[lineStart:matchEnd]

			lastIndex = matchEnd
		} else {
			// Without autoIndent, just insert directly before the match
			newContent += fileContent[lastIndex:matchStart]
			newContent += insertContent
			newContent += fileContent[matchStart:matchEnd]
			lastIndex = matchEnd
		}

		// If targeting specific occurrence and we've handled it, we're done processing matches
		if occurrence > 0 && i+1 == occurrence {
			break
		}
	}

	// Add any remaining content after the last match or insertion
	if lastIndex < len(fileContent) {
		newContent += fileContent[lastIndex:]
	}

	return newContent, nil
}


// InsertBeforeRegexMultiline inserts content before a specific occurrence of a regex pattern in a file.
// When multiline is true:
//   - (?s) flag is added: dot (.) matches newlines
//   - (?m) flag is added: ^ and $ match line boundaries
//
// Parameters:
//   - path: Path to the file
//   - pattern: Regular expression pattern to search for
//   - content: Content to insert before the pattern match
//   - occurrence: Which occurrence to insert before (1-based indexing, 0 means all occurrences)
//   - autoIndent: If true, automatically indents inserted content to match surrounding code
//   - multiline: If true, enables multiline mode (dot matches newlines, ^ and $ match line boundaries)
//
// Returns:
//   - The new content with insertions
//   - Error if any
func InsertBeforeRegexMultiline(path, pattern, content string, occurrence int, autoIndent bool, multiline bool) (string, error) {
	// Read file content
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	fileContent := string(fileBytes)

	// Prepare pattern with multiline flags if enabled
	compiledPattern := pattern
	if multiline {
		// (?s) makes . match newlines, (?m) makes ^ and $ match line boundaries
		compiledPattern = "(?sm)" + pattern
	}

	// Compile the regular expression
	re, err := regexp.Compile(compiledPattern)
	if err != nil {
		return "", fmt.Errorf("invalid regex pattern: %w", err)
	}

	// Find all matches
	indexes := re.FindAllStringIndex(fileContent, -1)
	if indexes == nil || len(indexes) == 0 {
		return "", fmt.Errorf("pattern '%s' not found in file", pattern)
	}

	totalMatches := len(indexes)

	// Check if requested occurrence exists
	if occurrence > totalMatches {
		return "", fmt.Errorf("specified occurrence %d exceeds total matches %d",
			occurrence, totalMatches)
	}

	// Build the new content with insertions
	var newContent string
	lastIndex := 0

	// Handle the "all occurrences" case or specific occurrence
	for i, idx := range indexes {
		matchStart := idx[0] // Start index of the match
		matchEnd := idx[1]   // End index of the match

		// If targeting specific occurrence and this isn't it, skip insertion
		if occurrence > 0 && i+1 != occurrence {
			continue
		}

		// Prepare the content to insert
		insertContent := content

		// If autoIndent is enabled, we need to insert on a new line with proper indentation
		if autoIndent {
			// Find the line containing the match to extract its indentation
			lineStart := matchStart
			for lineStart > 0 && fileContent[lineStart-1] != '\n' {
				lineStart--
			}

			// Extract the indentation of the current line
			lineIndent := ""
			for j := lineStart; j < len(fileContent) && (fileContent[j] == ' ' || fileContent[j] == '\t'); j++ {
				lineIndent += string(fileContent[j])
			}

			// For autoIndent, we insert content as new line(s) before the current line
			// First, add everything up to the line start
			if lastIndex <= lineStart {
				newContent += fileContent[lastIndex:lineStart]
			}

			// Apply indentation to each line of the content
			if lineIndent != "" {
				lines := splitLines(content)
				var indentedLines []string
				for _, line := range lines {
					if len(line) > 0 {
						indentedLines = append(indentedLines, lineIndent+line)
					} else {
						indentedLines = append(indentedLines, "")
					}
				}
				insertContent = joinLines(indentedLines)
				// Ensure trailing newline for proper line separation
				if len(insertContent) > 0 && insertContent[len(insertContent)-1] != '\n' {
					insertContent += "\n"
				}
			}

			// Add the indented content
			newContent += insertContent

			// Add the original line (from lineStart to matchEnd, which includes indentation and match)
			newContent += fileContent[lineStart:matchEnd]

			lastIndex = matchEnd
		} else {
			// Without autoIndent, just insert directly before the match
			newContent += fileContent[lastIndex:matchStart]
			newContent += insertContent
			newContent += fileContent[matchStart:matchEnd]
			lastIndex = matchEnd
		}

		// If targeting specific occurrence and we've handled it, we're done processing matches
		if occurrence > 0 && i+1 == occurrence {
			break
		}
	}

	// Add any remaining content after the last match or insertion
	if lastIndex < len(fileContent) {
		newContent += fileContent[lastIndex:]
	}

	return newContent, nil
}
