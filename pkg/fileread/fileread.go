package fileread

import (
	"bufio"
	"fmt"
	"io"
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
	IsPartial   bool   // Whether this was a partial file read (startLine > 1 or endLine specified)
}

// ReadFile is an optimized function to read file contents with optional line range control
// If startLine and endLine are both <= 0, it reads the entire file efficiently (if under maxSize)
// Otherwise, it performs line-by-line reading for the specified range
// Returns the file content exactly as-is in the original file
func ReadFile(path string, startLine int, endLine int, maxSize int) (FileReadResult, error) {
	result := FileReadResult{
		StartLine: startLine,
		EndLine:   endLine,
		IsPartial: startLine > 1 || endLine > 0,
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
	
	// Get file info
	fileInfo, err := os.Stat(path)
	if err != nil {
		return result, fmt.Errorf("failed to get file info: %w", err)
	}
	result.FileSize = fileInfo.Size()

	// OPTIMIZATION: If reading the entire file and it's smaller than maxSize, read it directly
	if (startLine <= 0 || startLine == 1) && endLine <= 0 && fileInfo.Size() <= int64(maxSize) {
		content, err := os.ReadFile(path)
		if err != nil {
			return result, fmt.Errorf("failed to read file: %w", err)
		}
		
		// Count lines for metadata
		lineCount := countLines(content)
		
		result.Content = string(content)
		result.ContentSize = len(content)
		result.StartLine = 1
		result.EndLine = lineCount
		result.TotalLines = lineCount
		result.ReadLines = lineCount
		result.IsPartial = false
		
		return result, nil
	}
	
	// For partial reads or large files, use line-by-line reading
	return readFileLineByLine(path, startLine, endLine, maxSize)
}

// Helper function to count lines in a byte array
func countLines(data []byte) int {
	lineCount := 0
	for _, b := range data {
		if b == '\n' {
			lineCount++
		}
	}
	// If file doesn't end with newline, count the last line
	if len(data) > 0 && data[len(data)-1] != '\n' {
		lineCount++
	}
	// Special case for empty file
	if len(data) == 0 {
		return 0
	}
	return lineCount
}

// readFileLineByLine performs line-by-line reading for partial file reads or large files
func readFileLineByLine(path string, startLine int, endLine int, maxSize int) (FileReadResult, error) {
	result := FileReadResult{
		IsPartial: startLine > 1 || endLine > 0,
	}

	// Open the file
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

	// Normalize parameters
	if startLine <= 0 {
		startLine = 1
	}
	result.StartLine = startLine

	// Create a scanner for line-by-line reading
	scanner := bufio.NewScanner(file)
	
	// For very large files or lines, increase buffer size
	const maxScanTokenSize = 1024 * 1024 // 1MB
	buf := make([]byte, maxScanTokenSize)
	scanner.Buffer(buf, maxScanTokenSize)

	// Skip lines before startLine
	lineCount := 0
	for lineCount < startLine-1 && scanner.Scan() {
		lineCount++
	}
	if err := scanner.Err(); err != nil {
		return result, fmt.Errorf("error while scanning file: %w", err)
	}

	// Read requested lines
	var contentBuilder strings.Builder
	readingLines := 0
	currentSize := 0

	for scanner.Scan() {
		lineCount++
		
		// Check if we've reached endLine
		if endLine > 0 && lineCount > endLine {
			break
		}

		line := scanner.Text()
		
		// Calculate size with newline character
		lineSize := len(line)
		if readingLines > 0 {
			lineSize++ // Account for newline
		}
		
		// Check if adding this line would exceed maxSize
		if currentSize+lineSize > maxSize {
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

	// Continue counting lines for metadata if we haven't reached EOF
	if !result.Truncated && (endLine <= 0 || lineCount <= endLine) {
		// Re-open the file to count total lines if we didn't read to EOF
		totalLines, err := countTotalLines(path)
		if err == nil {
			lineCount = totalLines
		}
	}

	// Fill in the result
	result.Content = contentBuilder.String()
	result.TotalLines = lineCount
	result.ReadLines = readingLines
	result.ContentSize = currentSize
	result.EndLine = result.StartLine + readingLines - 1

	return result, nil
}

// Helper function to count total lines in a file
func countTotalLines(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	count := 0
	buf := make([]byte, 32*1024)
	
	for {
		c, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return count, err
		}
		
		if c == 0 {
			break
		}
		
		for i := 0; i < c; i++ {
			if buf[i] == '\n' {
				count++
			}
		}
		
		if err == io.EOF {
			break
		}
	}
	
	// Check if file doesn't end with a newline
	if fi, err := file.Stat(); err == nil {
		if fi.Size() > 0 {
			// Check last byte
			if _, err := file.Seek(-1, io.SeekEnd); err == nil {
				lastByte := make([]byte, 1)
				if _, err := file.Read(lastByte); err == nil && lastByte[0] != '\n' {
					count++
				}
			}
		} else if fi.Size() == 0 {
			// Empty file has 0 lines
			return 0, nil
		}
	}
	
	return count, nil
}
