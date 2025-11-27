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

		// Step 1: Resolve symlinks in the allowed directory itself
		canonicalDir, err := filepath.EvalSymlinks(dir)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve symlinks for %q: %w", dir, err)
		}

		// Step 2: Convert to absolute path
		absDir, err := filepath.Abs(canonicalDir)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve absolute path for %q: %w", canonicalDir, err)
		}

		log.Printf("Canonical path: %q", absDir)

		// Step 3: Verify directory exists
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

// isPathAllowed checks if a path is within allowed directories.
//
// Security model:
// 1. Resolves ALL symbolic links to canonical paths using filepath.EvalSymlinks()
// 2. Converts to absolute path
// 3. Validates the canonical path is within allowed directories
// 4. Uses path separator in prefix matching to prevent "/allowed" matching "/allowed_attacker"
//
// For non-existent paths (write operations), validates the parent directory chain.
func (h *FileSystemHandler) isPathAllowed(path string) bool {
	// Step 1: Resolve symbolic links to get canonical path
	canonicalPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		// Check if path actually exists (could be broken symlink)
		_, statErr := os.Lstat(path)
		if statErr != nil && os.IsNotExist(statErr) {
			// Path truly doesn't exist - allow validation of parent for write operations
			return h.isPathAllowedNonExistent(path)
		}
		// SECURITY: Block broken symlinks and other resolution errors
		// If Lstat succeeded but EvalSymlinks failed, it's likely a broken symlink
		log.Printf("SECURITY: Access denied to %q - symlink resolution failed: %v", path, err)
		return false
	}

	// Step 2: Get absolute path (now that symlinks are resolved)
	absPath, err := filepath.Abs(canonicalPath)
	if err != nil {
		log.Printf("SECURITY: Access denied to %q - absolute path resolution failed: %v", canonicalPath, err)
		return false
	}

	log.Printf("Checking if path is allowed: %q (canonical: %q)", path, absPath)

	allowedDirs, err := getAllowedDirs()
	if err != nil {
		log.Printf("Error getting allowed directories: %v", err)
		return false
	}

	// Step 3: Validate against allowed directories with proper path separator handling
	for _, dir := range allowedDirs {
		// Ensure proper prefix matching: exact match OR prefix with separator
		if absPath == dir || strings.HasPrefix(absPath, dir+string(filepath.Separator)) {
			log.Printf("Path %q is allowed (matches %q)", absPath, dir)
			return true
		}
	}

	// SECURITY: Log blocked access attempts with details
	log.Printf("SECURITY: Access denied to %q (canonical: %q) - not in allowed directories %v", path, absPath, allowedDirs)
	return false
}

// isPathAllowedNonExistent handles validation for paths that don't exist yet.
// This is needed for write operations (write_file, create_directory, etc.)
func (h *FileSystemHandler) isPathAllowedNonExistent(path string) bool {
	// Get the parent directory and validate it exists and is allowed
	dir := filepath.Dir(path)

	// Keep going up until we find an existing directory
	for {
		canonicalDir, err := filepath.EvalSymlinks(dir)
		if err == nil {
			// Found existing parent - validate it
			absDir, err := filepath.Abs(canonicalDir)
			if err != nil {
				log.Printf("Error resolving absolute path for parent %q: %v", canonicalDir, err)
				return false
			}

			allowedDirs, err := getAllowedDirs()
			if err != nil {
				log.Printf("Error getting allowed directories: %v", err)
				return false
			}

			// Check if this parent is within allowed directories
			for _, allowedDir := range allowedDirs {
				if absDir == allowedDir || strings.HasPrefix(absDir, allowedDir+string(filepath.Separator)) {
					log.Printf("Non-existent path %q allowed (parent %q is within %q)", path, absDir, allowedDir)
					return true
				}
			}
			log.Printf("Non-existent path %q blocked (parent %q not in allowed directories)", path, absDir)
			return false
		}

		// If we've reached the root, stop
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			log.Printf("Non-existent path %q blocked (reached root without finding allowed parent)", path)
			return false
		}
		dir = parentDir
	}
}