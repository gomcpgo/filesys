package handler

import (
	"fmt"
	"log"
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

	log.Printf("Loading allowed directories from: %q", dirsStr)

	// Split by comma but preserve spaces in paths
	dirs := strings.Split(dirsStr, ",")
	cleanDirs := make([]string, 0, len(dirs))
	
	for _, dir := range dirs {
		// Only trim spaces around the entire path, not within it
		dir = strings.TrimSpace(dir)
		if dir == "" {
			continue
		}

		log.Printf("Processing directory: %q", dir)
		
		// Convert to absolute path (handles spaces correctly)
		absDir, err := filepath.Abs(dir)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve absolute path for %q: %w", dir, err)
		}
		
		log.Printf("Absolute path: %q", absDir)
		
		// Verify directory exists
		info, err := os.Stat(absDir)
		if err != nil {
			return nil, fmt.Errorf("directory %q does not exist: %w", absDir, err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("%q is not a directory", absDir)
		}
		
		cleanDirs = append(cleanDirs, absDir)
		log.Printf("Added allowed directory: %q", absDir)
	}

	if len(cleanDirs) == 0 {
		return nil, fmt.Errorf("no valid directories found in %s=%q", AllowedDirsEnvVar, dirsStr)
	}

	log.Printf("Final allowed directories: %q", cleanDirs)
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
		log.Printf("Error resolving absolute path for %q: %v", path, err)
		return false
	}

	log.Printf("Checking if path is allowed: %q", absPath)

	allowedDirs, err := getAllowedDirs()
	if err != nil {
		log.Printf("Error getting allowed directories: %v", err)
		return false
	}
	
	for _, dir := range allowedDirs {
		if strings.HasPrefix(absPath, dir) {
			log.Printf("Path %q is allowed (matches prefix %q)", absPath, dir)
			return true
		}
	}

	log.Printf("Path %q is not allowed in any directory %q", absPath, allowedDirs)
	return false
}