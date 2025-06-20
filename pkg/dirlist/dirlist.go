package dirlist

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// DirEntry represents a file or directory entry with metadata
type DirEntry struct {
	Name      string    // File/directory name
	Path      string    // Full path
	IsDir     bool      // Whether it's a directory
	Size      int64     // Size in bytes (0 for directories)
	ModTime   time.Time // Modification time
	Mode      fs.FileMode // File mode/permissions
	ItemCount int       // For directories: number of items (if known)
}

// ListOptions defines options for directory listing operations
type ListOptions struct {
	Pattern       string // Regex pattern for filtering
	FileType      string // "file", "dir", or extension (e.g., ".go")
	Recursive     bool   // Whether to list recursively
	MaxDepth      int    // Maximum recursion depth (0 = unlimited)
	MaxResults    int    // Maximum number of results
	IncludeHidden bool   // Whether to include hidden files
	IncludeMetadata bool // Whether to include detailed metadata
}

// ListingResult contains the results of a directory listing operation
type ListingResult struct {
	Entries      []DirEntry // List of directory entries
	TotalFiles   int        // Total number of files
	TotalDirs    int        // Total number of directories
	TotalSize    int64      // Total size in bytes
	Truncated    bool       // Whether results were truncated
	TotalEntries int        // Total entries before truncation
}

// DefaultListOptions returns ListOptions with sensible defaults
func DefaultListOptions() ListOptions {
	return ListOptions{
		Pattern:        "",
		FileType:       "",
		Recursive:      false,
		MaxDepth:       0,
		MaxResults:     100,
		IncludeHidden:  false,
		IncludeMetadata: true,
	}
}

// isHidden determines if a file or directory is hidden (starts with .)
func isHidden(name string) bool {
	return len(name) > 0 && name[0] == '.'
}

// ListDirectory lists the contents of a directory with the specified options
func ListDirectory(path string, options ListOptions) (ListingResult, error) {
	result := ListingResult{
		Entries: make([]DirEntry, 0),
	}

	// Validate path
	_, err := os.Stat(path)
	if err != nil {
		return result, err
	}

	// Compile regex pattern if provided
	var re *regexp.Regexp
	if options.Pattern != "" {
		re, err = regexp.Compile(options.Pattern)
		if err != nil {
			return result, err
		}
	}

	// Use a map to track unique entries for the result
	entriesMap := make(map[string]bool)
	
	// Process entries with Walk or simple ReadDir based on recursive option
	if options.Recursive {
		err = filepath.Walk(path, func(entryPath string, info fs.FileInfo, err error) error {
			// Skip if we've reached the maximum number of results
			if len(result.Entries) >= options.MaxResults {
				result.Truncated = true
				return filepath.SkipDir
			}
			
			// Skip the root path itself
			if entryPath == path {
				return nil
			}
			
			// Handle walk errors
			if err != nil {
				return nil
			}
			
			// Check depth
			if options.MaxDepth > 0 {
				// Calculate relative path to determine depth
				relPath, err := filepath.Rel(path, entryPath)
				if err != nil {
					return nil
				}
				
				// Count directory separators to determine depth
				depth := strings.Count(relPath, string(filepath.Separator)) + 1
				if depth > options.MaxDepth {
					if info.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
			}
			
			// Filter and process the entry
			entry, include := processEntry(entryPath, info, options, re)
			if include && !entriesMap[entryPath] {
				entriesMap[entryPath] = true
				result.Entries = append(result.Entries, entry)
				updateResultStats(&result, entry)
			}
			
			return nil
		})
	} else {
		// Non-recursive listing using ReadDir
		entries, err := os.ReadDir(path)
		if err != nil {
			return result, err
		}
		
		for _, dirEntry := range entries {
			// Skip if we've reached the maximum number of results
			if len(result.Entries) >= options.MaxResults {
				result.Truncated = true
				break
			}
			
			entryPath := filepath.Join(path, dirEntry.Name())
			info, err := dirEntry.Info()
			if err != nil {
				continue
			}
			
			entry, include := processEntry(entryPath, info, options, re)
			if include {
				entriesMap[entryPath] = true
				result.Entries = append(result.Entries, entry)
				updateResultStats(&result, entry)
			}
		}
	}
	
	// Count total entries (may be greater than what we collected due to MaxResults limit)
	if result.Truncated {
		// If truncated, we need to count all entries that would match
		totalEntries := 0
		if options.Recursive {
			filepath.Walk(path, func(entryPath string, info fs.FileInfo, err error) error {
				if err != nil || entryPath == path {
					return nil
				}
				
				// Apply same filters as in main processing
				if shouldIncludeEntry(entryPath, info, options, re) {
					totalEntries++
				}
				
				return nil
			})
		} else {
			entries, _ := os.ReadDir(path)
			for _, entry := range entries {
				info, err := entry.Info()
				if err == nil {
					entryPath := filepath.Join(path, entry.Name())
					if shouldIncludeEntry(entryPath, info, options, re) {
						totalEntries++
					}
				}
			}
		}
		result.TotalEntries = totalEntries
	} else {
		result.TotalEntries = len(result.Entries)
	}
	
	return result, err
}

// processEntry creates a DirEntry from fs.FileInfo and determines if it should be included based on filters
func processEntry(path string, info fs.FileInfo, options ListOptions, pattern *regexp.Regexp) (DirEntry, bool) {
	entry := DirEntry{
		Name:    info.Name(),
		Path:    path,
		IsDir:   info.IsDir(),
		Size:    0,
		ModTime: info.ModTime(),
		Mode:    info.Mode(),
	}
	
	// Check if entry should be included based on filters
	if !shouldIncludeEntry(path, info, options, pattern) {
		return entry, false
	}
	
	// Include size for files
	if !info.IsDir() {
		entry.Size = info.Size()
	} else if options.IncludeMetadata {
		// Get item count for directories if metadata is requested
		items, err := os.ReadDir(path)
		if err == nil {
			entry.ItemCount = len(items)
		}
	}
	
	return entry, true
}

// updateResultStats updates the statistics in the result based on a new entry
func updateResultStats(result *ListingResult, entry DirEntry) {
	if entry.IsDir {
		result.TotalDirs++
	} else {
		result.TotalFiles++
		result.TotalSize += entry.Size
	}
}


// shouldIncludeEntry determines if an entry should be included based on filters
func shouldIncludeEntry(path string, info fs.FileInfo, options ListOptions, pattern *regexp.Regexp) bool {
	// Skip hidden files if not included
	if !options.IncludeHidden && isHidden(info.Name()) {
		return false
	}
	
	// Apply pattern filter if specified
	if pattern != nil && !pattern.MatchString(info.Name()) {
		return false
	}
	
	// Apply file type filter if specified
	if options.FileType != "" {
		switch options.FileType {
		case "file":
			if info.IsDir() {
				return false
			}
		case "dir":
			if !info.IsDir() {
				return false
			}
		default:
			// Treat as file extension filter
			if !strings.HasSuffix(strings.ToLower(info.Name()), strings.ToLower(options.FileType)) {
				return false
			}
		}
	}
	
	return true
}