package handler

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gomcpgo/mcp/pkg/protocol"
)

func (h *FileSystemHandler) handleCopyLines(args map[string]interface{}) (*protocol.CallToolResponse, error) {
	sourcePath, ok := args["source_path"].(string)
	if !ok {
		return nil, fmt.Errorf("source_path must be a string")
	}

	destPath, ok := args["destination_path"].(string)
	if !ok {
		return nil, fmt.Errorf("destination_path must be a string")
	}

	startLine := 1
	if v, ok := args["start_line"].(float64); ok {
		startLine = int(v)
	}

	endLine := 0 // 0 means EOF
	if v, ok := args["end_line"].(float64); ok {
		endLine = int(v)
	}

	appendMode := false
	if v, ok := args["append"].(bool); ok {
		appendMode = v
	}

	// Validate line range
	if startLine < 1 {
		startLine = 1
	}
	if endLine > 0 && startLine > endLine {
		return nil, fmt.Errorf("start_line (%d) cannot be greater than end_line (%d)", startLine, endLine)
	}

	log.Printf("copy_lines - %s lines %d-%d to %s (append=%v)", sourcePath, startLine, endLine, destPath, appendMode)

	// Validate both paths
	if !h.isPathAllowed(sourcePath) {
		log.Printf("ERROR: copy_lines - access denied to source: %s", sourcePath)
		return nil, NewAccessDeniedError(sourcePath)
	}
	if !h.isPathAllowed(destPath) {
		log.Printf("ERROR: copy_lines - access denied to destination: %s", destPath)
		return nil, NewAccessDeniedError(destPath)
	}

	// Open source file
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		log.Printf("ERROR: copy_lines - failed to open source: %v", err)
		return nil, fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// Count total lines while we copy (for metadata)
	scanner := bufio.NewScanner(sourceFile)
	const maxScanTokenSize = 1024 * 1024
	buf := make([]byte, maxScanTokenSize)
	scanner.Buffer(buf, maxScanTokenSize)

	// Ensure destination directory exists
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Open destination file
	var destFile *os.File
	if appendMode {
		destFile, err = os.OpenFile(destPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	} else {
		destFile, err = os.Create(destPath)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to open destination file: %w", err)
	}
	defer destFile.Close()

	writer := bufio.NewWriter(destFile)
	lineNum := 0
	copiedLines := 0
	bytesWritten := 0

	for scanner.Scan() {
		lineNum++

		// Past end_line, stop copying but keep counting for total
		if endLine > 0 && lineNum > endLine {
			break
		}

		// In range — write to destination
		if lineNum >= startLine {
			line := scanner.Text()
			n, err := writer.WriteString(line)
			if err != nil {
				return nil, fmt.Errorf("failed to write to destination: %w", err)
			}
			bytesWritten += n

			n, err = writer.WriteString("\n")
			if err != nil {
				return nil, fmt.Errorf("failed to write newline: %w", err)
			}
			bytesWritten += n

			copiedLines++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading source file: %w", err)
	}

	if err := writer.Flush(); err != nil {
		return nil, fmt.Errorf("failed to flush destination: %w", err)
	}

	// Count remaining lines for total if we stopped early
	totalLines := lineNum
	if endLine > 0 && lineNum == endLine+1 {
		// We broke out of the loop — need to count remaining lines
		for scanner.Scan() {
			totalLines++
		}
	}

	// Build metadata-only response
	effectiveEnd := startLine + copiedLines - 1
	if copiedLines == 0 {
		effectiveEnd = startLine
	}

	result := fmt.Sprintf("Copied %d lines (%d bytes) from %s to %s\nSource lines: %d-%d of %d",
		copiedLines, bytesWritten, sourcePath, destPath, startLine, effectiveEnd, totalLines)

	log.Printf("copy_lines - %s", result)

	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{Type: "text", Text: result},
		},
	}, nil
}
