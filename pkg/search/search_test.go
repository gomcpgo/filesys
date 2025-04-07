package search

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestDir creates a temporary directory with test files and returns its path
func setupTestDir(t *testing.T) string {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "search-test")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	
	// Create test directory structure
	files := []string{
		"file1.txt",
		"another_file.md",
		"test_file.go",
		".hidden_file.txt",
		"subdir/nested_file.txt",
		"subdir/another_test.go",
		"subdir/.hidden_nested.md",
		"subdir/deeper/deep_test_file.json",
	}
	
	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		dir := filepath.Dir(path)
		
		// Create parent directories if needed
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		
		// Create empty file
		if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", path, err)
		}
	}
	
	return tmpDir
}

// cleanupTestDir removes the temporary test directory
func cleanupTestDir(tmpDir string) {
	os.RemoveAll(tmpDir)
}

// TestBasicSearch tests basic search functionality with default options
func TestBasicSearch(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer cleanupTestDir(tmpDir)
	
	result, err := Search(tmpDir, "file", DefaultSearchOptions())
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	
	// We expect multiple matches including file1.txt, test_file.go, nested_file.txt, etc.
	expectedMinCount := 4
	if len(result.Matches) < expectedMinCount {
		t.Errorf("Search() found %d matches, expected at least %d", len(result.Matches), expectedMinCount)
		t.Logf("Actual matches: %+v", result.Matches)
	}
	
	// Verify that each match actually contains the pattern
	for _, match := range result.Matches {
		if !strings.Contains(strings.ToLower(match.Name), "file") {
			t.Errorf("Match %s does not contain pattern 'file' (case-insensitive)", match.Name)
		}
	}
}

// TestCaseSensitiveSearch tests search with case sensitivity enabled
func TestCaseSensitiveSearch(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer cleanupTestDir(tmpDir)
	
	options := SearchOptions{
		CaseSensitive: true,
		MaxDepth:      -1,
	}
	
	result, err := Search(tmpDir, "File", options)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	
	// None of our test files have "File" with a capital 'F'
	if len(result.Matches) != 0 {
		t.Errorf("Case-sensitive search found %d matches, expected 0", len(result.Matches))
		t.Logf("Unexpected matches: %+v", result.Matches)
	}
	
	// Now try with lowercase "file" which should match
	result, err = Search(tmpDir, "file", options)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	
	if len(result.Matches) < 4 {
		t.Errorf("Case-sensitive search for 'file' found %d matches, expected at least 4", len(result.Matches))
	}
}

// TestDepthLimitedSearch tests search with a depth limit
func TestDepthLimitedSearch(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer cleanupTestDir(tmpDir)
	
	options := SearchOptions{
		CaseSensitive: false,
		MaxDepth:      0, // Only search in the root directory
	}
	
	// First test with depth 0 (only the current directory)
	result, err := Search(tmpDir, "file", options)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	
	// Only files in the root directory should be found
	for _, match := range result.Matches {
		relPath, err := filepath.Rel(tmpDir, match.Path)
		if err != nil {
			t.Errorf("Failed to get relative path: %v", err)
			continue
		}
		
		if strings.Contains(relPath, string(filepath.Separator)) {
			t.Errorf("Depth-0 search returned file from subdirectory: %s", relPath)
		}
	}
	
	// Now test with depth 1 to look for a file we know exists in the first subdirectory level
	options.MaxDepth = 1
	result, err = Search(tmpDir, "nested", options) // Use "nested" to match nested_file.txt
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	
	// Debug output
	t.Logf("Depth-1 search for 'nested' found %d matches:", len(result.Matches))
	for _, match := range result.Matches {
		t.Logf("Match: %s", match.Path)
	}
	
	// At least one file should be found in the first subdirectory level
	foundInSubdir := false
	foundTooDeep := false
	
	for _, match := range result.Matches {
		relPath, err := filepath.Rel(tmpDir, match.Path)
		if err != nil {
			t.Errorf("Failed to get relative path: %v", err)
			continue
		}
		
		parts := strings.Split(relPath, string(filepath.Separator))
		
		// A file is in a first-level subdirectory if it has exactly one separator
		if len(parts) == 2 { 
			foundInSubdir = true
			t.Logf("Found file in first-level subdirectory: %s", relPath)
		}
		
		// It should not find files in deeper subdirectories
		if len(parts) > 2 {
			foundTooDeep = true
			t.Logf("Found file too deep: %s", relPath)
		}
	}
	
	if !foundInSubdir {
		t.Errorf("Depth-1 search did not find files in first-level subdirectories")
	}
	
	if foundTooDeep {
		t.Errorf("Depth-1 search found files in deeper subdirectories than expected")
	}
	
	// Now let's test with unlimited depth
	options.MaxDepth = -1
	result, err = Search(tmpDir, "deep", options) // Should find deep_test_file.json
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	
	foundDeepFile := false
	for _, match := range result.Matches {
		relPath, err := filepath.Rel(tmpDir, match.Path)
		if err != nil {
			t.Errorf("Failed to get relative path: %v", err)
			continue
		}
		
		if strings.Contains(relPath, "deeper") {
			foundDeepFile = true
			break
		}
	}
	
	if !foundDeepFile {
		t.Errorf("Unlimited depth search did not find files in deep subdirectories")
	}
}

// TestPathMatching tests searching in full paths instead of just filenames
func TestPathMatching(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer cleanupTestDir(tmpDir)
	
	// First test without path matching (default)
	options := DefaultSearchOptions()
	result, err := Search(tmpDir, "deeper", options)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	
	// Should only find files with "deeper" in their names
	deeperInName := 0
	for _, match := range result.Matches {
		if strings.Contains(strings.ToLower(match.Name), "deeper") {
			deeperInName++
		}
	}
	
	// There are no files with "deeper" in the name, only a directory
	if deeperInName > 0 {
		t.Errorf("Without path matching, found %d files with 'deeper' in name; expected 0", deeperInName)
	}
	
	// Now test with path matching enabled
	options.MatchPath = true
	result, err = Search(tmpDir, "deeper", options)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	
	// Should find files in the "deeper" directory
	if len(result.Matches) == 0 {
		t.Errorf("With path matching, found 0 matches for 'deeper', expected at least 1")
	}
	
	foundDeepFile := false
	for _, match := range result.Matches {
		relPath, err := filepath.Rel(tmpDir, match.Path)
		if err != nil {
			t.Errorf("Failed to get relative path: %v", err)
			continue
		}
		
		if strings.Contains(relPath, "deeper") && strings.Contains(relPath, "deep_test_file.json") {
			foundDeepFile = true
			break
		}
	}
	
	if !foundDeepFile {
		t.Errorf("With path matching, did not find deep_test_file.json in deeper directory")
	}
}

// TestNonExistentDirectory tests search with a non-existent directory
func TestNonExistentDirectory(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer cleanupTestDir(tmpDir)
	
	nonExistentDir := filepath.Join(tmpDir, "non-existent-dir")
	_, err := Search(nonExistentDir, "test", DefaultSearchOptions())
	
	if err == nil {
		t.Errorf("Expected error when searching non-existent directory, got nil")
	}
}

// TestFileAsBasePath tests search with a file as the base path
func TestFileAsBasePath(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer cleanupTestDir(tmpDir)
	
	filePath := filepath.Join(tmpDir, "file1.txt")
	_, err := Search(filePath, "test", DefaultSearchOptions())
	
	if err == nil {
		t.Errorf("Expected error when using file as base path, got nil")
	}
}

// TestFormatMatches tests the string formatting of matches
func TestFormatMatches(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer cleanupTestDir(tmpDir)
	
	// Perform a search
	result, err := Search(tmpDir, "file", DefaultSearchOptions())
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	
	// Format the matches
	formatted := result.FormatMatches()
	
	// Check that the formatted string contains all matches
	for _, match := range result.Matches {
		if !strings.Contains(formatted, match.Path) {
			t.Errorf("FormatMatches() output does not contain %s", match.Path)
		}
	}
	
	// Test empty results
	emptyResult := &SearchResult{
		Pattern:  "nonexistent",
		BasePath: tmpDir,
		Matches:  []FileMatch{},
	}
	
	emptyFormatted := emptyResult.FormatMatches()
	if emptyFormatted != "No matches found." {
		t.Errorf("Expected 'No matches found.' for empty results, got: %s", emptyFormatted)
	}
}
