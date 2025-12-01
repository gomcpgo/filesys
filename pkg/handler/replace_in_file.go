package handler

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gomcpgo/mcp/pkg/protocol"
)

// replaceMatch holds information about a single replacement match
type replaceMatch struct {
	lineNum     int
	oldLine     string
	newLine     string
}

// findReplacementMatches finds all occurrences of searchString and returns match details with line numbers
func findReplacementMatches(content, searchString, replaceString string, occurrence int) ([]replaceMatch, string, int) {
	lines := strings.Split(content, "\n")
	var matches []replaceMatch
	var newLines []string
	currentOccurrence := 0

	for lineNum, line := range lines {
		if strings.Contains(line, searchString) {
			// Count occurrences in this line
			count := strings.Count(line, searchString)
			newLine := line

			if occurrence == 0 {
				// Replace all occurrences
				newLine = strings.ReplaceAll(line, searchString, replaceString)
				currentOccurrence += count
				matches = append(matches, replaceMatch{
					lineNum: lineNum + 1,
					oldLine: line,
					newLine: newLine,
				})
			} else {
				// Replace only the specified occurrence - need to track position
				searchPos := 0
				for i := 0; i < count; i++ {
					currentOccurrence++
					idx := strings.Index(line[searchPos:], searchString)
					if idx == -1 {
						break
					}
					actualIdx := searchPos + idx

					if currentOccurrence == occurrence {
						// Replace this specific occurrence
						newLine = line[:actualIdx] + replaceString + line[actualIdx+len(searchString):]
						matches = append(matches, replaceMatch{
							lineNum: lineNum + 1,
							oldLine: line,
							newLine: newLine,
						})
						break
					}
					searchPos = actualIdx + len(searchString)
				}
			}
			newLines = append(newLines, newLine)
		} else {
			newLines = append(newLines, line)
		}
	}

	// Deduplicate matches (multiple occurrences on same line should show once)
	seen := make(map[int]bool)
	var uniqueMatches []replaceMatch
	for _, m := range matches {
		if !seen[m.lineNum] {
			seen[m.lineNum] = true
			uniqueMatches = append(uniqueMatches, m)
		}
	}

	newContent := strings.Join(newLines, "\n")
	replacedCount := len(uniqueMatches)
	if occurrence > 0 && replacedCount > 0 {
		replacedCount = 1
	}

	return uniqueMatches, newContent, replacedCount
}

// formatDryRunPreview formats the preview output for dry run mode
func formatDryRunPreview(matches []replaceMatch, searchString string, wouldReplace int) string {
	var sb strings.Builder
	sb.WriteString("Preview (dry run - no changes applied):\n\n")

	for _, m := range matches {
		sb.WriteString(fmt.Sprintf("Line %d:\n", m.lineNum))
		sb.WriteString(fmt.Sprintf("- %s\n", m.oldLine))
		sb.WriteString(fmt.Sprintf("+ %s\n\n", m.newLine))
	}

	sb.WriteString(fmt.Sprintf("%d occurrence(s) of '%s' would be replaced.", wouldReplace, searchString))
	return sb.String()
}

// handleReplaceInFile replaces occurrences of a string in a file
func (h *FileSystemHandler) handleReplaceInFile(args map[string]interface{}) (*protocol.CallToolResponse, error) {
	path, ok := args["path"].(string)
	if !ok {
		log.Printf("ERROR: replace_in_file - invalid path type: %T", args["path"])
		return nil, fmt.Errorf("path must be a string")
	}

	searchString, ok := args["search"].(string)
	if !ok {
		log.Printf("ERROR: replace_in_file - invalid search string type: %T", args["search"])
		return nil, fmt.Errorf("search string must be a string")
	}

	replaceString, ok := args["replace"].(string)
	if !ok {
		log.Printf("ERROR: replace_in_file - invalid replace string type: %T", args["replace"])
		return nil, fmt.Errorf("replace string must be a string")
	}

	// Optional parameter for which occurrence to replace
	occurrence := 0 // 0 means replace all
	if occurrenceVal, ok := args["occurrence"].(float64); ok {
		occurrence = int(occurrenceVal)
	}

	// Optional parameter for dry run mode
	dryRun := false
	if dryRunVal, ok := args["dry_run"].(bool); ok {
		dryRun = dryRunVal
	}

	log.Printf("replace_in_file - attempting to replace '%s' with '%s' in %s (dry_run=%v)", searchString, replaceString, path, dryRun)

	if !h.isPathAllowed(path) {
		log.Printf("ERROR: replace_in_file - access denied to path: %s", path)
		return nil, NewAccessDeniedError(path)
	}

	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		log.Printf("ERROR: replace_in_file - failed to read file %s: %v", path, err)
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	fileContent := string(content)
	totalOccurrences := strings.Count(fileContent, searchString)

	if totalOccurrences == 0 {
		log.Printf("replace_in_file - search string '%s' not found in %s", searchString, path)
		return &protocol.CallToolResponse{
			Content: []protocol.ToolContent{
				{
					Type: "text",
					Text: fmt.Sprintf("String '%s' not found in %s", searchString, path),
				},
			},
		}, nil
	}

	if occurrence > totalOccurrences {
		log.Printf("ERROR: replace_in_file - specified occurrence %d exceeds total occurrences %d in %s",
			occurrence, totalOccurrences, path)
		return nil, fmt.Errorf("specified occurrence %d exceeds total occurrences %d",
			occurrence, totalOccurrences)
	}

	// Find matches with line numbers
	matches, newContent, replacedCount := findReplacementMatches(fileContent, searchString, replaceString, occurrence)

	// Dry run mode - return preview without modifying file
	if dryRun {
		log.Printf("replace_in_file - dry run: would replace %d occurrence(s) in %s", replacedCount, path)
		return &protocol.CallToolResponse{
			Content: []protocol.ToolContent{
				{
					Type: "text",
					Text: formatDryRunPreview(matches, searchString, replacedCount),
				},
			},
		}, nil
	}

	// Write back to file
	err = os.WriteFile(path, []byte(newContent), 0644)
	if err != nil {
		log.Printf("ERROR: replace_in_file - failed to write to %s: %v", path, err)
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Build response with line details
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Successfully replaced %d occurrence(s) of '%s' in %s:\n\n", replacedCount, searchString, path))
	for _, m := range matches {
		sb.WriteString(fmt.Sprintf("  Line %d: %s\n", m.lineNum, m.newLine))
	}

	log.Printf("replace_in_file - successfully replaced %d occurrence(s) in %s", replacedCount, path)
	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: sb.String(),
			},
		},
	}, nil
}
