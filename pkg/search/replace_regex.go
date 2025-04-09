package search

import (
	"fmt"
	"os"
	"regexp"
)

// ReplaceWithRegex replaces content that matches a regex pattern in a file.
// Parameters:
//   - path: Path to the file
//   - pattern: Regular expression pattern to search for
//   - replacement: Content to replace the pattern matches with (can use capture groups with $1, $2, etc.)
//   - occurrence: Which occurrence to replace (0 for all occurrences, 1+ for specific occurrence)
//   - caseSensitive: Whether the search is case sensitive
//
// Returns:
//   - The new content with replacements
//   - Number of replacements made
//   - Error if any
func ReplaceWithRegex(path, pattern, replacement string, occurrence int, caseSensitive bool) (string, int, error) {
	// Read file content
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read file: %w", err)
	}
	fileContent := string(fileBytes)

	// Special case: empty file should return empty content with 0 replacements
	if len(fileContent) == 0 {
		return "", 0, nil
	}

	// Compile the regular expression
	var re *regexp.Regexp
	if caseSensitive {
		re, err = regexp.Compile(pattern)
	} else {
		re, err = regexp.Compile("(?i)" + pattern)
	}
	if err != nil {
		return "", 0, fmt.Errorf("invalid regex pattern: %w", err)
	}

	// If we're replacing all occurrences, we can use the regexp package's built-in function
	if occurrence == 0 {
		// FindAllStringSubmatchIndex gives us access to capture groups
		allMatches := re.FindAllStringSubmatchIndex(fileContent, -1)
		if len(allMatches) == 0 {
			return fileContent, 0, nil
		}

		newContent := ""
		lastIndex := 0
		replacementCount := 0

		for _, match := range allMatches {
			// Add the text before this match
			newContent += fileContent[lastIndex:match[0]]

			// Apply the replacement with capture groups
			matchText := fileContent[match[0]:match[1]]
			replacedText := re.ReplaceAllString(matchText, replacement)
			newContent += replacedText

			lastIndex = match[1]
			replacementCount++
		}

		// Add any remaining content
		if lastIndex < len(fileContent) {
			newContent += fileContent[lastIndex:]
		}

		return newContent, replacementCount, nil
	}

	// For specific occurrences, we need more control
	allMatches := re.FindAllStringSubmatchIndex(fileContent, -1)
	if len(allMatches) == 0 {
		return fileContent, 0, nil
	}

	if occurrence > len(allMatches) {
		return "", 0, fmt.Errorf("specified occurrence %d exceeds total matches %d",
			occurrence, len(allMatches))
	}

	// Get the specific match we want to replace (1-indexed for occurrences)
	matchIndex := occurrence - 1
	match := allMatches[matchIndex]

	// Build the new content with the replacement
	newContent := fileContent[:match[0]]
	matchText := fileContent[match[0]:match[1]]
	replacedText := re.ReplaceAllString(matchText, replacement)
	newContent += replacedText
	newContent += fileContent[match[1]:]

	return newContent, 1, nil
}

// ReplaceWithRegexInString replaces content that matches a regex pattern in a string.
// Similar to ReplaceWithRegex but operates on a string directly instead of a file.
// This is useful for testing or when working with in-memory content.
func ReplaceWithRegexInString(content, pattern, replacement string, occurrence int, caseSensitive bool) (string, int, error) {
	// Special case: empty content should return empty string with 0 replacements
	if len(content) == 0 {
		return "", 0, nil
	}

	// Compile the regular expression
	var re *regexp.Regexp
	var err error
	if caseSensitive {
		re, err = regexp.Compile(pattern)
	} else {
		re, err = regexp.Compile("(?i)" + pattern)
	}
	if err != nil {
		return "", 0, fmt.Errorf("invalid regex pattern: %w", err)
	}

	// If we're replacing all occurrences, we can use the regexp package's built-in function
	if occurrence == 0 {
		// FindAllStringSubmatchIndex gives us access to capture groups
		allMatches := re.FindAllStringSubmatchIndex(content, -1)
		if len(allMatches) == 0 {
			return content, 0, nil
		}

		newContent := ""
		lastIndex := 0
		replacementCount := 0

		for _, match := range allMatches {
			// Add the text before this match
			newContent += content[lastIndex:match[0]]

			// Apply the replacement with capture groups
			matchText := content[match[0]:match[1]]
			replacedText := re.ReplaceAllString(matchText, replacement)
			newContent += replacedText

			lastIndex = match[1]
			replacementCount++
		}

		// Add any remaining content
		if lastIndex < len(content) {
			newContent += content[lastIndex:]
		}

		return newContent, replacementCount, nil
	}

	// For specific occurrences, we need more control
	allMatches := re.FindAllStringSubmatchIndex(content, -1)
	if len(allMatches) == 0 {
		return content, 0, nil
	}

	if occurrence > len(allMatches) {
		return "", 0, fmt.Errorf("specified occurrence %d exceeds total matches %d",
			occurrence, len(allMatches))
	}

	// Get the specific match we want to replace (1-indexed for occurrences)
	matchIndex := occurrence - 1
	match := allMatches[matchIndex]

	// Build the new content with the replacement
	newContent := content[:match[0]]
	matchText := content[match[0]:match[1]]
	replacedText := re.ReplaceAllString(matchText, replacement)
	newContent += replacedText
	newContent += content[match[1]:]

	return newContent, 1, nil
}
