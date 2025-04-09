package fileread

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// FileReadResult contains the result of a file read operation with line-based control
type FileReadResult struct {
	Content     string // The actual content read from the file
	TotalLines  int    // Total number of lines in the file (if known)
	ReadLines   int    // Number of lines actually read
	StartLine   int    // Effective start line of the read operation (1-indexed)
	EndLine     int    // Effective end line of the read operation (may be less than requested if truncated)
	Truncated   bool   // Whether the content was truncated due to size limits
	FileSize    int64  // Total size of the file in bytes
	ContentSize int    // Size of the returned content in bytes
}

// ReadFileLines reads a file with control over which lines to include and maximum size
// startLine and endLine are 1-indexed
// If startLine <= 0, reading starts from the first line
// If endLine <= 0, reading continues to the end of file (subject to size limits)
// If endLine < startLine, an error is returned
// maxSize controls the maximum size in bytes of the returned content
func ReadFileLines(path string, startLine int, endLine int, maxSize int) (FileReadResult, error) {
	result := FileReadResult{
		StartLine: startLine,
		EndLine:   endLine,
	}

	// Validate parameters
	if startLine < 0 {
		return result, fmt.Errorf("startLine cannot be negative")
	}
	if endLine < 0 {
		return result, fmt.Errorf("endLine cannot be negative")
	}
	if endLine > 0 && startLine > endLine {
		return result, fmt.Errorf("startLine (%d) cannot be greater than endLine (%d)", startLine, endLine)
	}
	if maxSize <= 0 {
		return result, fmt.Errorf("maxSize must be positive")
	}

	// Open the file and get info
	file, err := os.Open(path)
	if err != nil {
		return result, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file stats
	fileInfo, err := file.Stat()
	if err != nil {
		return result, fmt.Errorf("failed to get file stats: %w", err)
	}
	result.FileSize = fileInfo.Size()

	// Create a reader
	reader := bufio.NewReader(file)
	scanner := bufio.NewScanner(reader)

	// Normalize parameters:
	// If startLine is 0 or negative, start from line 1
	if startLine <= 0 {
		startLine = 1
	}
	result.StartLine = startLine

	var contentBuilder strings.Builder
	lineCount := 0
	currentSize := 0

	// Skip lines before startLine
	for lineCount < startLine-1 && scanner.Scan() {
		lineCount++
	}
	if err := scanner.Err(); err != nil {
		return result, fmt.Errorf("error while scanning file: %w", err)
	}

	// Read requested lines
	readingLines := 0
	for scanner.Scan() {
		lineCount++
		
		// Check if we've reached endLine
		if endLine > 0 && lineCount > endLine {
			break
		}

		line := scanner.Text()
		lineSize := len(line)
		
		// Check if adding this line would exceed maxSize
		if currentSize+lineSize+readingLines > maxSize {
			// We can't add this line without exceeding maxSize
			result.Truncated = true
			break
		}

		// Add a newline before each line (except the first)
		if readingLines > 0 {
			contentBuilder.WriteByte('\n')
		}
		
		contentBuilder.WriteString(line)
		currentSize += lineSize
		readingLines++
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return result, fmt.Errorf("error while scanning file: %w", err)
	}

	// Count total lines if we haven't reached the end of file yet
	if endLine > 0 && lineCount <= endLine && !result.Truncated {
		// Continue scanning to count total lines
		for scanner.Scan() {
			lineCount++
		}
		// Ignore any error since we're just counting
	}

	// Fill in the result
	result.Content = contentBuilder.String()
	result.TotalLines = lineCount
	result.ReadLines = readingLines
	result.ContentSize = currentSize
	result.EndLine = result.StartLine + readingLines - 1

	return result, nil
}
