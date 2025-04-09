package dirlist

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// setupTestDirectory creates a test directory structure
func setupTestDirectory(t *testing.T) (string, func()) {
	t.Helper()
	
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "dirlist-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	
	// Create a cleanup function to remove the directory when done
	cleanup := func() {
		os.RemoveAll(tempDir)
	}
	
	// Create test directory structure
	createTestStructure(t, tempDir)
	
	return tempDir, cleanup
}

// createTestStructure creates a test directory structure with various files and subdirectories
func createTestStructure(t *testing.T, baseDir string) {
	t.Helper()
	
	// Define structure - make it deterministic for tests
	structure := map[string]string{
		"file1.txt":                   "This is file 1",
		"file2.go":                    "package main\n\nfunc main() {}\n",
		"file3.md":                    "# Markdown File",
		".hidden_file":                "Hidden file content",
		"subdir1/subfile1.txt":        "Subdir file content",
		"subdir1/subfile2.go":         "package sub\n",
		"subdir1/.hidden_subfile":     "Hidden subfile content",
		"subdir2/deep/nested/file.txt": "Deeply nested file",
		// No empty directories as they're handled inconsistently
	}
	
	// Create each file and necessary directories
	for path, content := range structure {
		fullPath := filepath.Join(baseDir, path)
		
		// Create parent directories
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err != nil {
			t.Fatalf("Failed to create directory %s: %v", filepath.Dir(fullPath), err)
		}
		
		// Write file content
		err = os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write file %s: %v", fullPath, err)
		}
	}
}

// TestBasicListing tests basic directory listing without options
func TestBasicListing(t *testing.T) {
	tempDir, cleanup := setupTestDirectory(t)
	defer cleanup()
	
	// Test basic listing with default options
	options := DefaultListOptions()
	result, err := ListDirectory(tempDir, options)
	
	if err != nil {
		t.Fatalf("ListDirectory failed: %v", err)
	}
	
	// Verify visible files and directories (excluding hidden)
	expectedFiles := 3 // file1.txt, file2.go, file3.md
	expectedDirs := 2  // subdir1, subdir2
	expectedEntries := expectedFiles + expectedDirs
	
	if len(result.Entries) != expectedEntries {
		t.Errorf("Expected %d entries, got %d", expectedEntries, len(result.Entries))
	}
	
	// Check totals
	if result.TotalFiles != expectedFiles {
		t.Errorf("Expected %d files, got %d", expectedFiles, result.TotalFiles)
	}
	
	if result.TotalDirs != expectedDirs {
		t.Errorf("Expected %d directories, got %d", expectedDirs, result.TotalDirs)
	}
	
	// Verify no truncation
	if result.Truncated {
		t.Error("Expected no truncation but got truncated results")
	}
}

// TestRecursiveListing tests recursive directory listing
func TestRecursiveListing(t *testing.T) {
	tempDir, cleanup := setupTestDirectory(t)
	defer cleanup()
	
	// Test recursive listing
	options := DefaultListOptions()
	options.Recursive = true
	result, err := ListDirectory(tempDir, options)
	
	if err != nil {
		t.Fatalf("ListDirectory failed: %v", err)
	}
	
	// Verify we have entries from subdirectories (don't check exact count as directory 
	// structure might vary slightly between OS/filesystem implementations)
	foundNestedFile := false
	foundSubdir1 := false
	
	for _, entry := range result.Entries {
		if strings.Contains(entry.Path, "subdir1/subfile1.txt") {
			foundNestedFile = true
		}
		if strings.HasSuffix(entry.Path, "subdir1") {
			foundSubdir1 = true
		}
	}
	
	if !foundNestedFile {
		t.Error("Expected to find nested file subdir1/subfile1.txt but didn't")
	}
	
	if !foundSubdir1 {
		t.Error("Expected to find subdir1 directory but didn't")
	}
	
	// Check that deeply nested file is included
	hasDeepNestedFile := false
	for _, entry := range result.Entries {
		if strings.Contains(entry.Path, "nested/file.txt") {
			hasDeepNestedFile = true
			break
		}
	}
	
	if !hasDeepNestedFile {
		t.Error("Expected to find deeply nested file but didn't")
	}
}

// TestDepthLimiting tests limiting the recursion depth
func TestDepthLimiting(t *testing.T) {
	tempDir, cleanup := setupTestDirectory(t)
	defer cleanup()
	
	// Test recursive listing with depth limit of 1
	options := DefaultListOptions()
	options.Recursive = true
	options.MaxDepth = 1
	result, err := ListDirectory(tempDir, options)
	
	if err != nil {
		t.Fatalf("ListDirectory failed: %v", err)
	}
	
	// Should only include files and dirs from first level of subdirectories
	for _, entry := range result.Entries {
		// Calculate path depth relative to tempDir
		relPath, err := filepath.Rel(tempDir, entry.Path)
		if err != nil {
			t.Fatalf("Failed to get relative path: %v", err)
		}
		
		depth := strings.Count(relPath, string(filepath.Separator)) + 1
		if depth > options.MaxDepth + 1 { // +1 because our entries start in tempDir
			t.Errorf("Entry exceeds max depth: %s (depth %d)", relPath, depth)
		}
	}
	
	// Should NOT include the deeply nested file
	for _, entry := range result.Entries {
		if strings.Contains(entry.Path, "deep/nested") {
			t.Error("Found deep/nested file despite depth limitation")
		}
	}
}

// TestPatternFiltering tests filtering by name pattern
func TestPatternFiltering(t *testing.T) {
	tempDir, cleanup := setupTestDirectory(t)
	defer cleanup()
	
	// Test with pattern for .go files
	options := DefaultListOptions()
	options.Pattern = `\.go$`
	result, err := ListDirectory(tempDir, options)
	
	if err != nil {
		t.Fatalf("ListDirectory failed: %v", err)
	}
	
	// Should only include .go files
	for _, entry := range result.Entries {
		if !entry.IsDir && !strings.HasSuffix(entry.Name, ".go") {
			t.Errorf("Expected only .go files but found: %s", entry.Name)
		}
	}
	
	// Verify file count
	expectedFiles := 1 // Only one .go file in the root directory
	if len(result.Entries) != expectedFiles {
		t.Errorf("Expected %d entries, got %d", expectedFiles, len(result.Entries))
	}
}

// TestHiddenFiles tests showing/hiding hidden files
func TestHiddenFiles(t *testing.T) {
	tempDir, cleanup := setupTestDirectory(t)
	defer cleanup()
	
	// Test with hidden files included
	options := DefaultListOptions()
	options.IncludeHidden = true
	result, err := ListDirectory(tempDir, options)
	
	if err != nil {
		t.Fatalf("ListDirectory failed: %v", err)
	}
	
	// Should include hidden files
	hasHiddenFile := false
	for _, entry := range result.Entries {
		if entry.Name == ".hidden_file" {
			hasHiddenFile = true
			break
		}
	}
	
	if !hasHiddenFile {
		t.Error("Expected to find hidden file but didn't")
	}
	
	// Test with hidden files excluded (default)
	options.IncludeHidden = false
	result, err = ListDirectory(tempDir, options)
	
	if err != nil {
		t.Fatalf("ListDirectory failed: %v", err)
	}
	
	// Should NOT include hidden files
	for _, entry := range result.Entries {
		if strings.HasPrefix(entry.Name, ".") {
			t.Errorf("Found hidden file despite IncludeHidden=false: %s", entry.Name)
		}
	}
}

// TestFileTypeFiltering tests filtering by file type
func TestFileTypeFiltering(t *testing.T) {
	tempDir, cleanup := setupTestDirectory(t)
	defer cleanup()
	
	// Test with file type = "file"
	options := DefaultListOptions()
	options.FileType = "file"
	result, err := ListDirectory(tempDir, options)
	
	if err != nil {
		t.Fatalf("ListDirectory failed: %v", err)
	}
	
	// Should only include files, not directories
	for _, entry := range result.Entries {
		if entry.IsDir {
			t.Errorf("Expected only files but found directory: %s", entry.Name)
		}
	}
	
	// Test with file type = "dir"
	options.FileType = "dir"
	result, err = ListDirectory(tempDir, options)
	
	if err != nil {
		t.Fatalf("ListDirectory failed: %v", err)
	}
	
	// Should only include directories, not files
	for _, entry := range result.Entries {
		if !entry.IsDir {
			t.Errorf("Expected only directories but found file: %s", entry.Name)
		}
	}
	
	// Test with file type = ".txt" (extension)
	options.FileType = ".txt"
	result, err = ListDirectory(tempDir, options)
	
	if err != nil {
		t.Fatalf("ListDirectory failed: %v", err)
	}
	
	// Should only include .txt files
	for _, entry := range result.Entries {
		if !entry.IsDir && !strings.HasSuffix(entry.Name, ".txt") {
			t.Errorf("Expected only .txt files but found: %s", entry.Name)
		}
	}
}

// TestMaxResults tests the maximum results limit
func TestMaxResults(t *testing.T) {
	tempDir, cleanup := setupTestDirectory(t)
	defer cleanup()
	
	// Test with very small max results
	options := DefaultListOptions()
	options.MaxResults = 2
	options.Recursive = true // Try to get more entries than the limit
	result, err := ListDirectory(tempDir, options)
	
	if err != nil {
		t.Fatalf("ListDirectory failed: %v", err)
	}
	
	// Should respect the limit
	if len(result.Entries) > options.MaxResults {
		t.Errorf("Expected maximum %d entries, got %d", options.MaxResults, len(result.Entries))
	}
	
	// Should indicate truncation
	if !result.Truncated {
		t.Error("Expected truncation flag to be true")
	}
	
	// Total entries should be greater than returned entries
	if result.TotalEntries <= len(result.Entries) {
		t.Errorf("Expected TotalEntries > %d, got %d", len(result.Entries), result.TotalEntries)
	}
}

// TestMetadataFields tests that the metadata fields are correctly populated
func TestMetadataFields(t *testing.T) {
	tempDir, cleanup := setupTestDirectory(t)
	defer cleanup()
	
	// Create a specific test file we'll check metadata for
	testFilePath := filepath.Join(tempDir, "metadata_test.txt")
	testContent := "Test content for metadata verification"
	err := os.WriteFile(testFilePath, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Get file info for comparison
	fileInfo, err := os.Stat(testFilePath)
	if err != nil {
		t.Fatalf("Failed to get file info: %v", err)
	}
	
	// Run directory listing
	options := DefaultListOptions()
	result, err := ListDirectory(tempDir, options)
	if err != nil {
		t.Fatalf("ListDirectory failed: %v", err)
	}
	
	// Find our test file in the results
	var testEntry *DirEntry
	for i, entry := range result.Entries {
		if entry.Path == testFilePath {
			testEntry = &result.Entries[i]
			break
		}
	}
	
	if testEntry == nil {
		t.Fatalf("Test file not found in results")
	}
	
	// Verify metadata fields
	if testEntry.Name != "metadata_test.txt" {
		t.Errorf("Expected name 'metadata_test.txt', got '%s'", testEntry.Name)
	}
	
	if testEntry.Size != int64(len(testContent)) {
		t.Errorf("Expected size %d, got %d", len(testContent), testEntry.Size)
	}
	
	if testEntry.IsDir {
		t.Error("Expected IsDir to be false")
	}
	
	// Time should be close to current time
	timeDiff := time.Since(testEntry.ModTime)
	if timeDiff > time.Minute {
		t.Errorf("File modification time is too old: %v", timeDiff)
	}
	
	if testEntry.Mode != fileInfo.Mode() {
		t.Errorf("Expected mode %v, got %v", fileInfo.Mode(), testEntry.Mode)
	}
}

// TestInvalidDirectory tests handling of an invalid directory
func TestInvalidDirectory(t *testing.T) {
	nonExistentDir := "/path/to/nonexistent/directory"
	
	options := DefaultListOptions()
	_, err := ListDirectory(nonExistentDir, options)
	
	if err == nil {
		t.Error("Expected error for non-existent directory, but got nil")
	}
}

// TestInvalidPattern tests handling of an invalid regex pattern
func TestInvalidPattern(t *testing.T) {
	tempDir, cleanup := setupTestDirectory(t)
	defer cleanup()
	
	options := DefaultListOptions()
	options.Pattern = "["  // Invalid regex pattern
	
	_, err := ListDirectory(tempDir, options)
	
	if err == nil {
		t.Error("Expected error for invalid regex pattern, but got nil")
	}
}

// TestCombinedFilters tests combining multiple filters
func TestCombinedFilters(t *testing.T) {
	tempDir, cleanup := setupTestDirectory(t)
	defer cleanup()
	
	// Test combination of filters: recursive, pattern, and file type
	options := DefaultListOptions()
	options.Recursive = true
	options.Pattern = `file\d`  // Files with "file" followed by a digit
	options.FileType = ".txt"
	options.IncludeHidden = false
	
	result, err := ListDirectory(tempDir, options)
	
	if err != nil {
		t.Fatalf("ListDirectory failed: %v", err)
	}
	
	// Check that we found file1.txt
	found := false
	for _, entry := range result.Entries {
		if entry.Name == "file1.txt" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Expected to find 'file1.txt' but it was not found")
	}
}

// TestDirectoryItemCount tests that directory item counts are correct
func TestDirectoryItemCount(t *testing.T) {
	tempDir, cleanup := setupTestDirectory(t)
	defer cleanup()
	
	// Create a subdirectory with a known number of items
	testDirPath := filepath.Join(tempDir, "count_test_dir")
	err := os.MkdirAll(testDirPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	
	// Create some files in this directory
	for i := 1; i <= 5; i++ {
		filePath := filepath.Join(testDirPath, fmt.Sprintf("item%d.txt", i))
		err := os.WriteFile(filePath, []byte("test"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}
	
	// Run directory listing
	options := DefaultListOptions()
	options.FileType = "dir"  // Only list directories
	result, err := ListDirectory(tempDir, options)
	if err != nil {
		t.Fatalf("ListDirectory failed: %v", err)
	}
	
	// Find our test directory in the results
	var testEntry *DirEntry
	for i, entry := range result.Entries {
		if entry.Path == testDirPath {
			testEntry = &result.Entries[i]
			break
		}
	}
	
	if testEntry == nil {
		t.Fatalf("Test directory not found in results")
	}
	
	// Verify item count
	if testEntry.ItemCount != 5 {
		t.Errorf("Expected item count 5, got %d", testEntry.ItemCount)
	}
}
