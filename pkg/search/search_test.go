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
	
	// First test with lowercase pattern that should match
	options := SearchOptions{
		CaseSensitive: true,
	}
	
	result, err := Search(tmpDir, "file", options)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	
	if len(result.Matches) < 4 {
		t.Errorf("Case-sensitive search for 'file' found %d matches, expected at least 4", len(result.Matches))
		t.Logf("Matches found: %+v", result.Matches)
	}
	
	// Now test with uppercase pattern that should not match
	result, err = Search(tmpDir, "FILE", options)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	
	if len(result.Matches) > 0 {
		t.Errorf("Case-sensitive search for 'FILE' found %d matches, expected 0", len(result.Matches))
		t.Logf("Unexpected matches: %+v", result.Matches)
	}
}

// TestPathMatching tests searching in full paths instead of just filenames
func TestPathMatching(t *testing.T) {
	tmpDir := setupTestDir(t)
	defer cleanupTestDir(tmpDir)
	
	// Create a new directory with a specific name for testing
	testDirPath := filepath.Join(tmpDir, "special_test_directory")
	if err := os.MkdirAll(testDirPath, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	
	// Create a test file that doesn't have "special" in its name
	testFilePath := filepath.Join(testDirPath, "normal.txt")
	if err := os.WriteFile(testFilePath, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Test with filename-only matching (default)
	options := DefaultSearchOptions()
	result, err := Search(tmpDir, "special", options)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	
	t.Logf("Without path matching, search for 'special' found %d matches", len(result.Matches))
	for _, match := range result.Matches {
		t.Logf("Match: %s", match.Path)
	}
	
	// Should not find the normal.txt file when only matching filenames
	foundNormalTxt := false
	foundSpecialDir := false
	for _, match := range result.Matches {
		if strings.HasSuffix(match.Path, "normal.txt") {
			foundNormalTxt = true
		}
		if match.IsDir && strings.Contains(match.Path, "special_test_directory") {
			foundSpecialDir = true
		}
	}
	
	// The directory itself should match "special"
	if !foundSpecialDir {
		t.Errorf("Without path matching, did not find the 'special_test_directory' directory")
	}
	
	// normal.txt should not match "special" when only matching filenames
	if foundNormalTxt {
		t.Errorf("Without path matching, found normal.txt when searching for 'special'")
	}
	
	// Now test with path matching enabled
	options.MatchPath = true
	result, err = Search(tmpDir, "special", options)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	
	t.Logf("With path matching, search for 'special' found %d matches", len(result.Matches))
	for _, match := range result.Matches {
		t.Logf("Match: %s", match.Path)
	}
	
	// Should find normal.txt because it's in a path containing "special"
	foundNormalTxt = false
	for _, match := range result.Matches {
		if strings.HasSuffix(match.Path, "normal.txt") {
			foundNormalTxt = true
			break
		}
	}
	
	if !foundNormalTxt {
		t.Errorf("With path matching, did not find normal.txt in special_test_directory")
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
