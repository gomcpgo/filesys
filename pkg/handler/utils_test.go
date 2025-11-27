package handler

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTestDirs creates a test directory structure with allowed and restricted areas
func setupTestDirs(t *testing.T) (allowed string, restricted string, cleanup func()) {
	tmpDir, err := os.MkdirTemp("", "filesys-security-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	allowedDir := filepath.Join(tmpDir, "allowed")
	restrictedDir := filepath.Join(tmpDir, "restricted")

	if err := os.MkdirAll(allowedDir, 0755); err != nil {
		t.Fatalf("Failed to create allowed dir: %v", err)
	}
	if err := os.MkdirAll(restrictedDir, 0755); err != nil {
		t.Fatalf("Failed to create restricted dir: %v", err)
	}

	// Create a sensitive file in restricted area
	if err := os.WriteFile(filepath.Join(restrictedDir, "secret.txt"), []byte("sensitive data"), 0644); err != nil {
		t.Fatalf("Failed to write secret file: %v", err)
	}

	cleanup = func() {
		_ = os.RemoveAll(tmpDir)
	}

	return allowedDir, restrictedDir, cleanup
}

// Test 1: Symlink Attack - Direct symlink to restricted directory
func TestSymlinkAttackToRestrictedDirectory(t *testing.T) {
	allowed, restricted, cleanup := setupTestDirs(t)
	defer cleanup()

	// Set environment
	os.Setenv("MCP_ALLOWED_DIRS", allowed)
	defer os.Unsetenv("MCP_ALLOWED_DIRS")

	// Clear cache to force reload
	allowedDirsMutex.Lock()
	allowedDirsCache = nil
	allowedDirsMutex.Unlock()

	// Create symlink in allowed directory pointing to restricted directory
	symlinkPath := filepath.Join(allowed, "escape")
	err := os.Symlink(restricted, symlinkPath)
	if err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	handler := NewFileSystemHandler()

	// Attempt to access restricted file through symlink
	attackPath := filepath.Join(symlinkPath, "secret.txt")

	if handler.isPathAllowed(attackPath) {
		t.Errorf("SECURITY VULNERABILITY: Symlink attack succeeded! Path %q should be blocked", attackPath)
	}
}

// Test 2: Symlink Attack - Nested symlinks
func TestSymlinkAttackNested(t *testing.T) {
	allowed, restricted, cleanup := setupTestDirs(t)
	defer cleanup()

	os.Setenv("MCP_ALLOWED_DIRS", allowed)
	defer os.Unsetenv("MCP_ALLOWED_DIRS")

	allowedDirsMutex.Lock()
	allowedDirsCache = nil
	allowedDirsMutex.Unlock()

	// Create chain: allowed/link1 -> allowed/link2 -> restricted/
	link1 := filepath.Join(allowed, "link1")
	link2 := filepath.Join(allowed, "link2")

	err := os.Symlink(link2, link1)
	if err != nil {
		t.Fatalf("Failed to create first symlink: %v", err)
	}
	err = os.Symlink(restricted, link2)
	if err != nil {
		t.Fatalf("Failed to create second symlink: %v", err)
	}

	handler := NewFileSystemHandler()
	attackPath := filepath.Join(link1, "secret.txt")

	if handler.isPathAllowed(attackPath) {
		t.Errorf("SECURITY VULNERABILITY: Nested symlink attack succeeded!")
	}
}

// Test 3: Legitimate symlink within allowed directory
func TestLegitimateSymlinkWithinAllowed(t *testing.T) {
	allowed, _, cleanup := setupTestDirs(t)
	defer cleanup()

	os.Setenv("MCP_ALLOWED_DIRS", allowed)
	defer os.Unsetenv("MCP_ALLOWED_DIRS")

	allowedDirsMutex.Lock()
	allowedDirsCache = nil
	allowedDirsMutex.Unlock()

	// Create legitimate file and symlink within allowed directory
	realFile := filepath.Join(allowed, "real.txt")
	symlinkFile := filepath.Join(allowed, "link.txt")

	err := os.WriteFile(realFile, []byte("legitimate data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create real file: %v", err)
	}
	err = os.Symlink(realFile, symlinkFile)
	if err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	handler := NewFileSystemHandler()

	// Both should be allowed
	if !handler.isPathAllowed(realFile) {
		t.Errorf("Real file should be allowed: %q", realFile)
	}

	if !handler.isPathAllowed(symlinkFile) {
		t.Errorf("Legitimate symlink within allowed directory should be allowed: %q", symlinkFile)
	}
}

// Test 4: Path traversal with ..
func TestPathTraversalAttack(t *testing.T) {
	allowed, _, cleanup := setupTestDirs(t)
	defer cleanup()

	os.Setenv("MCP_ALLOWED_DIRS", allowed)
	defer os.Unsetenv("MCP_ALLOWED_DIRS")

	allowedDirsMutex.Lock()
	allowedDirsCache = nil
	allowedDirsMutex.Unlock()

	handler := NewFileSystemHandler()

	// Try to escape using ../
	attackPath := filepath.Join(allowed, "..", "restricted", "secret.txt")

	if handler.isPathAllowed(attackPath) {
		t.Errorf("SECURITY VULNERABILITY: Path traversal attack succeeded! Path %q should be blocked", attackPath)
	}
}

// Test 5: Non-existent file in allowed directory (write operation)
func TestNonExistentFileInAllowedDirectory(t *testing.T) {
	allowed, _, cleanup := setupTestDirs(t)
	defer cleanup()

	os.Setenv("MCP_ALLOWED_DIRS", allowed)
	defer os.Unsetenv("MCP_ALLOWED_DIRS")

	allowedDirsMutex.Lock()
	allowedDirsCache = nil
	allowedDirsMutex.Unlock()

	handler := NewFileSystemHandler()

	// Non-existent file in allowed directory should be allowed (for write operations)
	newFile := filepath.Join(allowed, "newfile.txt")

	if !handler.isPathAllowed(newFile) {
		t.Errorf("Non-existent file in allowed directory should be allowed: %q", newFile)
	}
}

// Test 6: Non-existent file in restricted area
func TestNonExistentFileInRestrictedArea(t *testing.T) {
	allowed, restricted, cleanup := setupTestDirs(t)
	defer cleanup()

	os.Setenv("MCP_ALLOWED_DIRS", allowed)
	defer os.Unsetenv("MCP_ALLOWED_DIRS")

	allowedDirsMutex.Lock()
	allowedDirsCache = nil
	allowedDirsMutex.Unlock()

	handler := NewFileSystemHandler()

	// Non-existent file in restricted area should be blocked
	newFile := filepath.Join(restricted, "newfile.txt")

	if handler.isPathAllowed(newFile) {
		t.Errorf("Non-existent file in restricted directory should be blocked: %q", newFile)
	}
}

// Test 7: Broken symlink
func TestBrokenSymlink(t *testing.T) {
	allowed, _, cleanup := setupTestDirs(t)
	defer cleanup()

	os.Setenv("MCP_ALLOWED_DIRS", allowed)
	defer os.Unsetenv("MCP_ALLOWED_DIRS")

	allowedDirsMutex.Lock()
	allowedDirsCache = nil
	allowedDirsMutex.Unlock()

	// Create broken symlink
	brokenLink := filepath.Join(allowed, "broken")
	err := os.Symlink("/nonexistent/path", brokenLink)
	if err != nil {
		t.Fatalf("Failed to create broken symlink: %v", err)
	}

	handler := NewFileSystemHandler()

	// Broken symlink should be blocked (can't verify target)
	if handler.isPathAllowed(brokenLink) {
		t.Errorf("Broken symlink should be blocked: %q", brokenLink)
	}
}

// Test 8: Allowed directory is itself a symlink
func TestAllowedDirectoryIsSymlink(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-security-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create real directory and symlink to it
	realDir := filepath.Join(tmpDir, "real")
	symlinkDir := filepath.Join(tmpDir, "symlink")

	err = os.MkdirAll(realDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create real directory: %v", err)
	}
	err = os.Symlink(realDir, symlinkDir)
	if err != nil {
		t.Fatalf("Failed to create symlink directory: %v", err)
	}

	// Create test file
	testFile := filepath.Join(realDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Use symlink as allowed directory
	os.Setenv("MCP_ALLOWED_DIRS", symlinkDir)
	defer os.Unsetenv("MCP_ALLOWED_DIRS")

	allowedDirsMutex.Lock()
	allowedDirsCache = nil
	allowedDirsMutex.Unlock()

	handler := NewFileSystemHandler()

	// Should allow access to file in real directory
	if !handler.isPathAllowed(testFile) {
		t.Errorf("File in real directory should be allowed when allowed dir is symlink: %q", testFile)
	}
}

// Test 9: Prefix matching edge case
func TestPrefixMatchingEdgeCase(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-security-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create directories with similar prefixes
	allowed := filepath.Join(tmpDir, "allowed")
	attacker := filepath.Join(tmpDir, "allowed_attacker")

	err = os.MkdirAll(allowed, 0755)
	if err != nil {
		t.Fatalf("Failed to create allowed directory: %v", err)
	}
	err = os.MkdirAll(attacker, 0755)
	if err != nil {
		t.Fatalf("Failed to create attacker directory: %v", err)
	}

	err = os.WriteFile(filepath.Join(attacker, "evil.txt"), []byte("evil"), 0644)
	if err != nil {
		t.Fatalf("Failed to create evil file: %v", err)
	}

	os.Setenv("MCP_ALLOWED_DIRS", allowed)
	defer os.Unsetenv("MCP_ALLOWED_DIRS")

	allowedDirsMutex.Lock()
	allowedDirsCache = nil
	allowedDirsMutex.Unlock()

	handler := NewFileSystemHandler()

	// Should NOT allow access to attacker directory
	attackPath := filepath.Join(attacker, "evil.txt")
	if handler.isPathAllowed(attackPath) {
		t.Errorf("SECURITY VULNERABILITY: Prefix matching allowed similar directory name: %q", attackPath)
	}
}

// Test 10: Multiple allowed directories
func TestMultipleAllowedDirectories(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filesys-security-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dir1 := filepath.Join(tmpDir, "dir1")
	dir2 := filepath.Join(tmpDir, "dir2")
	dir3 := filepath.Join(tmpDir, "dir3")

	err = os.MkdirAll(dir1, 0755)
	if err != nil {
		t.Fatalf("Failed to create dir1: %v", err)
	}
	err = os.MkdirAll(dir2, 0755)
	if err != nil {
		t.Fatalf("Failed to create dir2: %v", err)
	}
	err = os.MkdirAll(dir3, 0755)
	if err != nil {
		t.Fatalf("Failed to create dir3: %v", err)
	}

	file1 := filepath.Join(dir1, "test1.txt")
	file2 := filepath.Join(dir2, "test2.txt")
	file3 := filepath.Join(dir3, "test3.txt")

	err = os.WriteFile(file1, []byte("test1"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	err = os.WriteFile(file2, []byte("test2"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}
	err = os.WriteFile(file3, []byte("test3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file3: %v", err)
	}

	// Allow dir1 and dir2, but not dir3
	os.Setenv("MCP_ALLOWED_DIRS", dir1+","+dir2)
	defer os.Unsetenv("MCP_ALLOWED_DIRS")

	allowedDirsMutex.Lock()
	allowedDirsCache = nil
	allowedDirsMutex.Unlock()

	handler := NewFileSystemHandler()

	if !handler.isPathAllowed(file1) {
		t.Errorf("File in first allowed directory should be allowed: %q", file1)
	}

	if !handler.isPathAllowed(file2) {
		t.Errorf("File in second allowed directory should be allowed: %q", file2)
	}

	if handler.isPathAllowed(file3) {
		t.Errorf("File in non-allowed directory should be blocked: %q", file3)
	}
}
