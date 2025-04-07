package search

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SearchOptions contains configuration options for the file search
type SearchOptions struct {
	// CaseSensitive determines if pattern matching should be case-sensitive
	CaseSensitive bool
	// MaxDepth limits the recursion depth (0 = current directory only, -1 = unlimited)
	MaxDepth int
	// MatchPath determines if the pattern should be matched against the full path
	// instead of just the filename
	MatchPath bool
}

// DefaultSearchOptions provides sensible default search options
func DefaultSearchOptions() SearchOptions {
	return SearchOptions{
		CaseSensitive: false,
		MaxDepth:      -1, // Unlimited depth
		MatchPath:     false, // Only match filename by default
	}
}

// FileMatch represents a matched file or directory
type FileMatch struct {
	// Path is the full path to the matched file or directory
	Path string
	// Name is the base name of the matched file or directory
	Name string
	// IsDir indicates if this is a directory
	IsDir bool
	// Size is the file size in bytes (0 for directories)
	Size int64
	// ModTime is the last modification time
	ModTime time.Time
}

// SearchResult contains the results of a file search operation
type SearchResult struct {
	// Pattern is the search pattern that was used
	Pattern string
	// BasePath is the directory where the search was performed
	BasePath string
	// Matches contains all the matching files and directories
	Matches []FileMatch
	// Errors contains any errors encountered during the search
	Errors []error
}

// FormatMatches returns a string representation of all matches
func (sr *SearchResult) FormatMatches() string {
	if len(sr.Matches) == 0 {
		return "No matches found."
	}
	
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Found %d matches for pattern '%s' in %s:\n\n", 
		len(sr.Matches), sr.Pattern, sr.BasePath))
	
	for _, match := range sr.Matches {
		// Add the file type indicator
		fileType := "F"
		if match.IsDir {
			fileType = "D"
		}
		
		// Add the size for files
		size := ""
		if !match.IsDir {
			size = fmt.Sprintf(" (%d bytes)", match.Size)
		}
		
		builder.WriteString(fmt.Sprintf("[%s] %s%s\n", fileType, match.Path, size))
	}
	
	if len(sr.Errors) > 0 {
		builder.WriteString("\nWarnings encountered during search:\n")
		for _, err := range sr.Errors {
			builder.WriteString(fmt.Sprintf("- %v\n", err))
		}
	}
	
	return builder.String()
}

// Search searches for files and directories whose names contain the specified pattern
func Search(basePath string, pattern string, options SearchOptions) (*SearchResult, error) {
	// Ensure the base path exists and is a directory
	info, err := os.Stat(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to access base path: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("base path is not a directory: %s", basePath)
	}
	
	result := &SearchResult{
		Pattern:  pattern,
		BasePath: basePath,
		Matches:  []FileMatch{},
		Errors:   []error{},
	}
	
	// If we want case-insensitive search, convert pattern to lowercase
	searchPattern := pattern
	if !options.CaseSensitive {
		searchPattern = strings.ToLower(pattern)
	}
	
	// Walk the directory tree
	err = filepath.Walk(basePath, func(path string, info fs.FileInfo, err error) error {
		// Handle errors during traversal
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("error accessing %s: %w", path, err))
			return nil // Continue despite errors
		}
		
		// Check depth limit
		if options.MaxDepth >= 0 {
			relPath, err := filepath.Rel(basePath, path)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("error determining relative path for %s: %w", path, err))
				return nil
			}
			
			// Skip if beyond max depth
			depth := 0
			if relPath != "." {
				depth = len(strings.Split(relPath, string(filepath.Separator)))
			}
			
			if depth > options.MaxDepth {
				if info.IsDir() {
					return filepath.SkipDir // Skip directories beyond max depth
				}
				return nil
			}
		}
		
		// Determine what to match against (name or full path)
		var stringToMatch string
		if options.MatchPath {
			// Use the path relative to base path to avoid matching on the base path itself
			relPath, err := filepath.Rel(basePath, path)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("error determining relative path for %s: %w", path, err))
				relPath = path // Fall back to full path if relative path can't be determined
			}
			stringToMatch = relPath
		} else {
			stringToMatch = info.Name()
		}
		
		// Check if it matches the pattern
		var matches bool
		if options.CaseSensitive {
			matches = strings.Contains(stringToMatch, searchPattern)
		} else {
			matches = strings.Contains(strings.ToLower(stringToMatch), searchPattern)
		}
		
		if matches {
			match := FileMatch{
				Path:    path,
				Name:    info.Name(),
				IsDir:   info.IsDir(),
				Size:    info.Size(),
				ModTime: info.ModTime(),
			}
			result.Matches = append(result.Matches, match)
		}
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("error walking directory tree: %w", err)
	}
	
	return result, nil
}


