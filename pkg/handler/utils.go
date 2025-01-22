package handler

import (
	"path/filepath"
	"strings"
)

// isPathAllowed checks if a path is within allowed directories
func (h *FileSystemHandler) isPathAllowed(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	
	for _, dir := range h.allowedDirs {
		if strings.HasPrefix(absPath, dir) {
			return true
		}
	}
	return false
}