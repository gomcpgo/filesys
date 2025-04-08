package search

import (
	"fmt"
	"os"
	"regexp"
)

// InsertAfterRegex inserts content after a specific occurrence of a regex pattern in a file.
// Parameters:
//   - path: Path to the file
//   - pattern: Regular expression pattern to search for
//   - content: Content to insert after the pattern match
//   - occurrence: Which occurrence to insert after (1-based indexing, 0 means all occurrences)
//
// Returns:
//   - The new content with insertions
//   - Error if any
func InsertAfterRegex(path, pattern, content string, occurrence int) (string, error) {
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
		matchEnd := idx[1] // End index of the match

		// If targeting specific occurrence and this isn't it, skip insertion
		if occurrence > 0 && i+1 != occurrence {
			continue
		}

		// Add content up to and including the matched pattern
		newContent += fileContent[lastIndex:matchEnd]

		// Add the content to insert - handle literal characters in the content string
		// This fixes issues with escape sequences like \n being treated literally
		newContent += content

		lastIndex = matchEnd

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
