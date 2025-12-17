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
			return "", 0, fmt.Errorf("Pattern not found in file.\n"+
				"Pattern: %q\n"+
				"Note: For multiline patterns, the pattern must match content exactly including whitespace.\n"+
				"Consider using patterns that don't rely on \\n matching literal newlines.", pattern)
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
		return "", 0, fmt.Errorf("Pattern not found in file.\n"+
			"Pattern: %q\n"+
			"Note: For multiline patterns, the pattern must match content exactly including whitespace.\n"+
			"Consider using patterns that don't rely on \\n matching literal newlines.", pattern)
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
			return "", 0, fmt.Errorf("Pattern not found in content.\n"+
				"Pattern: %q\n"+
				"Note: For multiline patterns, the pattern must match content exactly including whitespace.\n"+
				"Consider using patterns that don't rely on \\n matching literal newlines.", pattern)
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
		return "", 0, fmt.Errorf("Pattern not found in content.\n"+
			"Pattern: %q\n"+
			"Note: For multiline patterns, the pattern must match content exactly including whitespace.\n"+
			"Consider using patterns that don't rely on \\n matching literal newlines.", pattern)
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


// RegexMatch holds information about a single regex match for preview
type RegexMatch struct {
	LineNum  int
	OldLine  string
	NewLine  string
	MatchStr string
}

// FindRegexMatches finds all matches of a regex pattern in a file and returns match details
// without modifying the file. Useful for dry-run/preview mode.
func FindRegexMatches(path, pattern, replacement string, occurrence int, caseSensitive bool) ([]RegexMatch, string, int, error) {
	// Read file content
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, "", 0, fmt.Errorf("failed to read file: %w", err)
	}
	fileContent := string(fileBytes)

	return FindRegexMatchesInString(fileContent, pattern, replacement, occurrence, caseSensitive)
}

// FindRegexMatchesInString finds all matches of a regex pattern in a string and returns match details
func FindRegexMatchesInString(content, pattern, replacement string, occurrence int, caseSensitive bool) ([]RegexMatch, string, int, error) {
	if len(content) == 0 {
		return nil, "", 0, nil
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
		return nil, "", 0, fmt.Errorf("invalid regex pattern: %w", err)
	}

	lines := splitLines(content)
	var matches []RegexMatch
	var newLines []string
	currentOccurrence := 0

	for lineNum, line := range lines {
		lineMatches := re.FindAllStringIndex(line, -1)
		if len(lineMatches) == 0 {
			newLines = append(newLines, line)
			continue
		}

		newLine := line
		matchedThisLine := false

		for _, match := range lineMatches {
			currentOccurrence++

			if occurrence == 0 || currentOccurrence == occurrence {
				matchStr := line[match[0]:match[1]]
				if !matchedThisLine {
					// Apply all replacements for this line
					if occurrence == 0 {
						newLine = re.ReplaceAllString(line, replacement)
					} else {
						// Replace only this specific match
						newLine = line[:match[0]] + re.ReplaceAllString(matchStr, replacement) + line[match[1]:]
					}
					matches = append(matches, RegexMatch{
						LineNum:  lineNum + 1,
						OldLine:  line,
						NewLine:  newLine,
						MatchStr: matchStr,
					})
					matchedThisLine = true
				}
			}
		}
		newLines = append(newLines, newLine)
	}

	newContent := joinLines(newLines)
	replacedCount := len(matches)
	if occurrence > 0 && len(matches) > 0 {
		replacedCount = 1
	}

	return matches, newContent, replacedCount, nil
}

// splitLines splits content into lines preserving line endings
func splitLines(content string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(content); i++ {
		if content[i] == '\n' {
			lines = append(lines, content[start:i])
			start = i + 1
		}
	}
	if start < len(content) {
		lines = append(lines, content[start:])
	}
	return lines
}

// joinLines joins lines with newline characters
func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	result := lines[0]
	for i := 1; i < len(lines); i++ {
		result += "\n" + lines[i]
	}
	return result
}
