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
//
// Returns:
//   - The new content with insertions
//   - Error if any
func InsertBeforeRegex(path, pattern, content string, occurrence int) (string, error) {
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

		// If targeting specific occurrence and this isn't it, skip insertion
		if occurrence > 0 && i+1 != occurrence {
			if i+1 < occurrence {
				continue
			}
			// For occurrences after the one we're targeting,
			// just add the content up to this point and continue
			newContent += fileContent[lastIndex:idx[1]]
			lastIndex = idx[1]
			continue
		}

		// Add content up to the matched pattern
		newContent += fileContent[lastIndex:matchStart]

		// Insert the new content before the match
		newContent += content

		// Add the matched pattern itself
		newContent += fileContent[matchStart:idx[1]]

		// Update lastIndex to point to the end of this match
		lastIndex = idx[1]
	}

	// Add any remaining content after the last processed match
	if lastIndex < len(fileContent) {
		newContent += fileContent[lastIndex:]
	}

	return newContent, nil
}
