package search

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// SearchOptions defines parameters for the search operation
type SearchOptions struct {
	RootDir         string   // Directory to search in
	Pattern         string   // Regex pattern to search for
	FileExtensions  []string // File extensions to include (e.g., ".go", ".txt")
	MaxFileSearches int      // Maximum number of files to search (default 100)
	MaxResults      int      // Maximum number of results to return (default 100)
	CaseSensitive   bool     // Whether search is case sensitive (default true)
}

// SearchMatch represents a single match in a file
type SearchMatch struct {
	FilePath    string // Full path of the file
	LineNumber  int    // Line number of the match
	LineContent string // The content of the line with the match
}

// SearchResult contains all matches from the search operation
type SearchResult struct {
	Matches       []SearchMatch // Array of matches
	FilesSearched int           // Number of files searched
	FilesMatched  int           // Number of files with matches
	TotalMatches  int           // Total number of matches found
}

// DefaultSearchOptions returns SearchOptions with sensible defaults
func DefaultSearchOptions() SearchOptions {
	return SearchOptions{
		MaxFileSearches: 100,
		MaxResults:      100,
		CaseSensitive:   true,
		FileExtensions:  []string{".txt", ".go", ".js", ".html", ".css", ".md", ".json", ".yml", ".yaml", ".xml"},
	}
}

// Search performs a regex search in files within the root directory
func Search(opts SearchOptions) (SearchResult, error) {
	// Apply defaults for zero values
	if opts.MaxFileSearches <= 0 {
		opts.MaxFileSearches = 100
	}
	if opts.MaxResults <= 0 {
		opts.MaxResults = 100
	}

	// Check if root directory exists
	_, err := os.Stat(opts.RootDir)
	if err != nil {
		return SearchResult{}, fmt.Errorf("root directory error: %w", err)
	}

	// Compile regex pattern
	var regexPattern *regexp.Regexp
	if opts.CaseSensitive {
		regexPattern, err = regexp.Compile(opts.Pattern)
	} else {
		regexPattern, err = regexp.Compile("(?i)" + opts.Pattern)
	}
	if err != nil {
		return SearchResult{}, fmt.Errorf("invalid regex pattern: %w", err)
	}

	// Initialize result
	result := SearchResult{
		Matches: make([]SearchMatch, 0),
	}

	// Create a map of allowed extensions for faster lookup
	allowedExts := make(map[string]bool)
	for _, ext := range opts.FileExtensions {
		allowedExts[strings.ToLower(ext)] = true
	}

	// Walk through directory
	err = filepath.Walk(opts.RootDir, func(path string, info os.FileInfo, err error) error {
		// Skip if there was an error accessing this path
		if err != nil {
			return nil
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Stop if we've reached the maximum number of files to search
		if result.FilesSearched >= opts.MaxFileSearches {
			return filepath.SkipDir
		}

		// Check file extension if extensions were specified
		if len(allowedExts) > 0 {
			ext := strings.ToLower(filepath.Ext(path))
			if !allowedExts[ext] {
				return nil
			}
		}

		// Skip if the file is too large (greater than 10MB) or likely binary
		if !isTextFile(path, info) {
			return nil
		}

		// Increment files searched counter
		result.FilesSearched++

		// Search the file
		fileMatches, err := searchFile(path, regexPattern)
		if err != nil {
			// Just skip files that can't be read
			return nil
		}

		// Add matches to result, up to the maximum
		if len(fileMatches) > 0 {
			result.FilesMatched++
			for _, match := range fileMatches {
				if result.TotalMatches >= opts.MaxResults {
					break
				}
				result.Matches = append(result.Matches, match)
				result.TotalMatches++
			}
		}

		// Stop if we've reached the maximum number of results
		if result.TotalMatches >= opts.MaxResults {
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return result, fmt.Errorf("error walking directory: %w", err)
	}

	return result, nil
}

// searchFile searches a single file for the regex pattern
func searchFile(filePath string, pattern *regexp.Regexp) ([]SearchMatch, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Scan the file line by line
	scanner := bufio.NewScanner(file)
	lineNum := 0
	var matches []SearchMatch

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Check if the line matches the pattern
		if pattern.MatchString(line) {
			matches = append(matches, SearchMatch{
				FilePath:    filePath,
				LineNumber:  lineNum,
				LineContent: line,
			})
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return matches, nil
}

// isTextFile checks if a file is likely to be a text file
// It checks both the size and content
func isTextFile(filePath string, info os.FileInfo) bool {
	// Skip large files (greater than 10MB)
	if info.Size() > 10*1024*1024 {
		return false
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	// Check the first 512 bytes for null bytes
	// Files with null bytes are likely binary
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false
	}

	// Check for null bytes which usually indicate a binary file
	for i := 0; i < n; i++ {
		if buffer[i] == 0 {
			return false
		}
	}

	return true
}
