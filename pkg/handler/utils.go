package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	// Environment variable for allowed directories
	AllowedDirsEnvVar = "MCP_ALLOWED_DIRS"
)

var (
	// Cache for allowed directories
	allowedDirsCache []string
	allowedDirsMutex sync.RWMutex
)

// loadAllowedDirectories loads and validates allowed directories from environment variable
func loadAllowedDirectories() ([]string, error) {
	dirsStr := os.Getenv(AllowedDirsEnvVar)
	if dirsStr == "" {
		return nil, fmt.Errorf("environment variable %s not set", AllowedDirsEnvVar)
	}

	// Split by comma and trim spaces
	dirs := strings.Split(dirsStr, ",")
	cleanDirs := make([]string, 0, len(dirs))
	
	for _, dir := range dirs {
		dir = strings.TrimSpace(dir)
		if dir != "" {
			// Convert to absolute path
			absDir, err := filepath.Abs(dir)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve absolute path for %s: %w", dir, err)
			}
			
			// Verify directory exists
			info, err := os.Stat(absDir)
			if err != nil {
				return nil, fmt.Errorf("directory %s does not exist: %w", absDir, err)
			}
			if !info.IsDir() {
				return nil, fmt.Errorf("%s is not a directory", absDir)
			}
			
			cleanDirs = append(cleanDirs, absDir)
		}
	}

	if len(cleanDirs) == 0 {
		return nil, fmt.Errorf("no valid directories found in %s", AllowedDirsEnvVar)
	}

	return cleanDirs, nil
}

// getAllowedDirs gets allowed directories with caching
func getAllowedDirs() ([]string, error) {
	// Try to get from cache first
	allowedDirsMutex.RLock()
	if dirs := allowedDirsCache; dirs != nil {
		allowedDirsMutex.RUnlock()
		return dirs, nil
	}
	allowedDirsMutex.RUnlock()

	// Cache miss - load directories
	allowedDirsMutex.Lock()
	defer allowedDirsMutex.Unlock()

	// Double check after acquiring write lock
	if dirs := allowedDirsCache; dirs != nil {
		return dirs, nil
	}

	dirs, err := loadAllowedDirectories()
	if err != nil {
		return nil, err
	}

	allowedDirsCache = dirs
	return dirs, nil
}

// isPathAllowed checks if a path is within allowed directories
func (h *FileSystemHandler) isPathAllowed(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	allowedDirs, err := getAllowedDirs()
	if err != nil {
		return false
	}
	
	for _, dir := range allowedDirs {
		if strings.HasPrefix(absPath, dir) {
			return true
		}
	}
	return false
}