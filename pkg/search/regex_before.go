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
			if i+1 < occurrence {
				continue
			}
			// For occurrences after the one we're targeting,
			// just add the content up to this point and continue
			newContent += fileContent[lastIndex:matchEnd]
			lastIndex = matchEnd
			continue
		}

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

		// Add content up to the line start (not including the line's indentation)
		newContent += fileContent[lastIndex:lineStart]

		// Prepare the content to insert
		insertContent := content

		// If autoIndent is enabled, apply indentation from the current line
		if autoIndent {
			// Check if the first non-empty line of content already starts with whitespace
			// Only auto-indent if it doesn't
			shouldAutoIndent := false
			if lineIndent != "" {
				lines := splitLines(content)
				for _, line := range lines {
					if len(line) > 0 {
						// Only auto-indent if the line doesn't start with whitespace
						if line[0] != ' ' && line[0] != '\t' {
							shouldAutoIndent = true
						}
						break
					}
				}
			}

			// Apply indentation if needed
			if shouldAutoIndent {
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
				// Preserve trailing newline if it was in the original content
				if len(content) > 0 && content[len(content)-1] == '\n' {
					insertContent += "\n"
				}
			}
		}

		// Insert the indented content before the match
		newContent += insertContent

		// Add the original line's indentation (if it had any) plus the match
		if lineIndent != "" {
			newContent += lineIndent
		}
		newContent += fileContent[matchStart:matchEnd]

		// Update lastIndex to point to the end of this match
		lastIndex = matchEnd
	}

	// Add any remaining content after the last processed match
	if lastIndex < len(fileContent) {
		newContent += fileContent[lastIndex:]
	}

	return newContent, nil
}
