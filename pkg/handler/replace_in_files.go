package handler

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gomcpgo/mcp/pkg/protocol"
)

// fileReplaceResult holds the result of replacements in a single file
type fileReplaceResult struct {
	path         string
	matches      []replaceMatch
	replacements int
	err          error
}

// handleReplaceInFiles replaces occurrences of a string across multiple files
func (h *FileSystemHandler) handleReplaceInFiles(args map[string]interface{}) (*protocol.CallToolResponse, error) {
	// Extract paths array
	pathsArg, ok := args["paths"].([]interface{})
	if !ok {
		log.Printf("ERROR: replace_in_files - invalid paths type: %T", args["paths"])
		return nil, fmt.Errorf("paths must be an array of strings")
	}

	paths := make([]string, 0, len(pathsArg))
	for i, p := range pathsArg {
		path, ok := p.(string)
		if !ok {
			log.Printf("ERROR: replace_in_files - invalid path type at index %d: %T", i, p)
			return nil, fmt.Errorf("path at index %d must be a string", i)
		}
		paths = append(paths, path)
	}

	if len(paths) == 0 {
		return nil, fmt.Errorf("paths array cannot be empty")
	}

	searchString, ok := args["search"].(string)
	if !ok {
		log.Printf("ERROR: replace_in_files - invalid search string type: %T", args["search"])
		return nil, fmt.Errorf("search string must be a string")
	}

	replaceString, ok := args["replace"].(string)
	if !ok {
		log.Printf("ERROR: replace_in_files - invalid replace string type: %T", args["replace"])
		return nil, fmt.Errorf("replace string must be a string")
	}

	// Optional parameter for dry run mode
	dryRun := false
	if dryRunVal, ok := args["dry_run"].(bool); ok {
		dryRun = dryRunVal
	}

	log.Printf("replace_in_files - attempting to replace '%s' with '%s' in %d files (dry_run=%v)",
		searchString, replaceString, len(paths), dryRun)

	// Validate all paths are allowed first (fail fast)
	for _, path := range paths {
		if !h.isPathAllowed(path) {
			log.Printf("ERROR: replace_in_files - access denied to path: %s", path)
			return nil, NewAccessDeniedError(path)
		}
	}

	// Process each file
	results := make([]fileReplaceResult, 0, len(paths))
	totalReplacements := 0
	filesModified := 0

	for _, path := range paths {
		result := h.processFileReplacement(path, searchString, replaceString, dryRun)
		results = append(results, result)

		if result.err == nil && result.replacements > 0 {
			totalReplacements += result.replacements
			filesModified++
		}
	}

	// Format response
	responseText := formatBatchReplaceResponse(results, searchString, totalReplacements, filesModified, dryRun)

	log.Printf("replace_in_files - %s %d occurrence(s) across %d of %d files",
		map[bool]string{true: "would replace", false: "replaced"}[dryRun],
		totalReplacements, filesModified, len(paths))

	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: responseText,
			},
		},
	}, nil
}

// processFileReplacement handles replacement in a single file
func (h *FileSystemHandler) processFileReplacement(path, searchString, replaceString string, dryRun bool) fileReplaceResult {
	result := fileReplaceResult{path: path}

	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		result.err = fmt.Errorf("failed to read file: %w", err)
		return result
	}

	fileContent := string(content)

	// Find matches with line numbers (replace all occurrences)
	matches, newContent, replacedCount := findReplacementMatches(fileContent, searchString, replaceString, 0)
	result.matches = matches
	result.replacements = replacedCount

	if replacedCount == 0 {
		return result
	}

	// If not dry run, write the changes
	if !dryRun {
		err = os.WriteFile(path, []byte(newContent), 0644)
		if err != nil {
			result.err = fmt.Errorf("failed to write file: %w", err)
			return result
		}
	}

	return result
}

// formatBatchReplaceResponse formats the response for batch replacement
func formatBatchReplaceResponse(results []fileReplaceResult, searchString string, totalReplacements, filesModified int, dryRun bool) string {
	var sb strings.Builder

	if dryRun {
		sb.WriteString("Preview (dry run - no changes applied):\n\n")
	} else {
		sb.WriteString(fmt.Sprintf("Replaced in %d of %d files:\n\n", filesModified, len(results)))
	}

	for _, result := range results {
		sb.WriteString(fmt.Sprintf("%s:\n", result.path))

		if result.err != nil {
			sb.WriteString(fmt.Sprintf("  Error: %v\n\n", result.err))
			continue
		}

		if result.replacements == 0 {
			sb.WriteString("  No matches found\n\n")
			continue
		}

		for _, m := range result.matches {
			sb.WriteString(fmt.Sprintf("  Line %d: %s\n", m.lineNum, m.newLine))
		}
		sb.WriteString(fmt.Sprintf("  (%d replacement(s))\n\n", result.replacements))
	}

	if dryRun {
		sb.WriteString(fmt.Sprintf("Total: %d occurrence(s) of '%s' would be replaced across %d file(s)",
			totalReplacements, searchString, filesModified))
	} else {
		sb.WriteString(fmt.Sprintf("Total: %d replacement(s) across %d file(s)", totalReplacements, filesModified))
	}

	return sb.String()
}
