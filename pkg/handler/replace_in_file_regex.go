package handler

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gomcpgo/filesys/pkg/search"
	"github.com/gomcpgo/mcp/pkg/protocol"
)

// formatRegexDryRunPreview formats the preview output for regex dry run mode
func formatRegexDryRunPreview(matches []search.RegexMatch, pattern string, wouldReplace int) string {
	var sb strings.Builder
	sb.WriteString("Preview (dry run - no changes applied):\n\n")

	for _, m := range matches {
		sb.WriteString(fmt.Sprintf("Line %d:\n", m.LineNum))
		sb.WriteString(fmt.Sprintf("- %s\n", m.OldLine))
		sb.WriteString(fmt.Sprintf("+ %s\n\n", m.NewLine))
	}

	sb.WriteString(fmt.Sprintf("%d occurrence(s) of pattern '%s' would be replaced.", wouldReplace, pattern))
	return sb.String()
}

// handleReplaceInFileRegex replaces content that matches a regex pattern in a file
func (h *FileSystemHandler) handleReplaceInFileRegex(args map[string]interface{}) (*protocol.CallToolResponse, error) {
	path, ok := args["path"].(string)
	if !ok {
		log.Printf("ERROR: replace_in_file_regex - invalid path type: %T", args["path"])
		return nil, fmt.Errorf("path must be a string")
	}

	pattern, ok := args["pattern"].(string)
	if !ok {
		log.Printf("ERROR: replace_in_file_regex - invalid pattern type: %T", args["pattern"])
		return nil, fmt.Errorf("regex pattern must be a string")
	}

	replaceString, ok := args["replace"].(string)
	if !ok {
		log.Printf("ERROR: replace_in_file_regex - invalid replace string type: %T", args["replace"])
		return nil, fmt.Errorf("replace string must be a string")
	}

	// Optional parameter for which occurrence to replace
	occurrence := 0 // 0 means replace all
	if occurrenceVal, ok := args["occurrence"].(float64); ok {
		occurrence = int(occurrenceVal)
	}

	// Optional parameter for case sensitivity
	caseSensitive := true // default is case sensitive
	if caseVal, ok := args["case_sensitive"].(bool); ok {
		caseSensitive = caseVal
	}

	// Optional parameter for dry run mode
	dryRun := false
	if dryRunVal, ok := args["dry_run"].(bool); ok {
		dryRun = dryRunVal
	}

	// Optional parameter for multiline mode
	multiline := false
	if multilineVal, ok := args["multiline"].(bool); ok {
		multiline = multilineVal
	}

	log.Printf("replace_in_file_regex - attempting to replace pattern '%s' with '%s' in %s (dry_run=%v, multiline=%v)",
		pattern, replaceString, path, dryRun, multiline)

	if !h.isPathAllowed(path) {
		log.Printf("ERROR: replace_in_file_regex - access denied to path: %s", path)
		return nil, NewAccessDeniedError(path)
	}

	var matches []search.RegexMatch
	var newContent string
	var replacementCount int
	var err error

	if multiline {
		// Multiline mode - use the multiline function
		newContent, replacementCount, err = search.ReplaceWithRegexMultiline(path, pattern, replaceString, occurrence, caseSensitive, true)
		if err != nil {
			log.Printf("ERROR: replace_in_file_regex - %v", err)
			return nil, err
		}

		if replacementCount == 0 {
			log.Printf("replace_in_file_regex - regex pattern '%s' not found in %s", pattern, path)
			return &protocol.CallToolResponse{
				Content: []protocol.ToolContent{
					{
						Type: "text",
						Text: fmt.Sprintf("Pattern '%s' not found in %s", pattern, path),
					},
				},
			}, nil
		}

		// Dry run mode for multiline - show simpler preview
		if dryRun {
			log.Printf("replace_in_file_regex - dry run: would replace %d occurrence(s) in %s (multiline)", replacementCount, path)
			return &protocol.CallToolResponse{
				Content: []protocol.ToolContent{
					{
						Type: "text",
						Text: fmt.Sprintf("Preview (dry run - no changes applied):\n\n%d occurrence(s) of pattern '%s' would be replaced.\n\nResulting content:\n%s",
							replacementCount, pattern, newContent),
					},
				},
			}, nil
		}
	} else {
		// Standard mode - use line-by-line matching for better preview
		matches, newContent, replacementCount, err = search.FindRegexMatches(path, pattern, replaceString, occurrence, caseSensitive)
		if err != nil {
			log.Printf("ERROR: replace_in_file_regex - %v", err)
			return nil, err
		}

		if replacementCount == 0 {
			log.Printf("replace_in_file_regex - regex pattern '%s' not found in %s", pattern, path)
			return &protocol.CallToolResponse{
				Content: []protocol.ToolContent{
					{
						Type: "text",
						Text: fmt.Sprintf("Pattern '%s' not found in %s", pattern, path),
					},
				},
			}, nil
		}

		// Dry run mode - return preview without modifying file
		if dryRun {
			log.Printf("replace_in_file_regex - dry run: would replace %d occurrence(s) in %s", replacementCount, path)
			return &protocol.CallToolResponse{
				Content: []protocol.ToolContent{
					{
						Type: "text",
						Text: formatRegexDryRunPreview(matches, pattern, replacementCount),
					},
				},
			}, nil
		}
	}

	// Write the new content back to the file
	err = os.WriteFile(path, []byte(newContent), 0644)
	if err != nil {
		log.Printf("ERROR: replace_in_file_regex - failed to write to %s: %v", path, err)
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Build response with line details
	var sb strings.Builder
	if occurrence == 0 {
		sb.WriteString(fmt.Sprintf("Successfully replaced %d occurrence(s) of pattern '%s' in %s:\n\n",
			replacementCount, pattern, path))
	} else {
		sb.WriteString(fmt.Sprintf("Successfully replaced occurrence %d of pattern '%s' in %s:\n\n",
			occurrence, pattern, path))
	}
	for _, m := range matches {
		sb.WriteString(fmt.Sprintf("  Line %d: %s\n", m.LineNum, m.NewLine))
	}

	log.Printf("replace_in_file_regex - successfully replaced %d occurrence(s) in %s", replacementCount, path)
	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: sb.String(),
			},
		},
	}, nil
}
