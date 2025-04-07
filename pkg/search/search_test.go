package search

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestFiles creates temporary test files for searching
func setupTestFiles(t *testing.T) (string, func()) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "search-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create test files - keep content simple and predictable
	testFiles := map[string]string{
		"file1.txt":         "Line one of file1\nThis contains apple\nAnother line here\n",
		"file2.txt":         "First line of file2\nWith an apple and orange\nLast line of file2\n",
		"file3.go":          "package main\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}\n",
		"binary.bin":        string([]byte{0, 1, 2, 3, 4, 5}),
		"large.txt":         "This is a test line\n",
		"subdir/nested.txt": "Nested file content\nWith apple inside\nEnd of nested file\n",
	}

	// Create subdirectory
	err = os.Mkdir(filepath.Join(tempDir, "subdir"), 0755)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Write the test files
	for filename, content := range testFiles {
		fullPath := filepath.Join(tempDir, filename)

		// Create parent directory if needed
		if dir := filepath.Dir(fullPath); dir != tempDir {
			if err := os.MkdirAll(dir, 0755); err != nil {
				os.RemoveAll(tempDir)
				t.Fatalf("Failed to create directory %s: %v", dir, err)
			}
		}

		err = os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			os.RemoveAll(tempDir)
			t.Fatalf("Failed to write test file %s: %v", filename, err)
		}

		// Make the "large.txt" file actually large by appending to it
		if filename == "large.txt" {
			f, err := os.OpenFile(fullPath, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				os.RemoveAll(tempDir)
				t.Fatalf("Failed to open large file for appending: %v", err)
			}

			// Add 1000 lines to make it larger
			for i := 0; i < 1000; i++ {
				_, err = f.WriteString("This is a filler line to make the file larger\n")
				if err != nil {
					f.Close()
					os.RemoveAll(tempDir)
					t.Fatalf("Failed to append to large file: %v", err)
				}
			}

			f.Close()
		}
	}

	// Return a cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

// TestBasicSearch tests searching for a simple string in files
func TestBasicSearch(t *testing.T) {
	tempDir, cleanup := setupTestFiles(t)
	defer cleanup()

	options := SearchOptions{
		RootDir:         tempDir,
		Pattern:         "apple",
		FileExtensions:  []string{".txt", ".go"},
		MaxFileSearches: 100,
		MaxResults:      100,
		CaseSensitive:   true,
	}

	result, err := Search(options)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedCount := 3
	if len(result.Matches) != expectedCount {
		t.Errorf("Expected %d matches, got %d", expectedCount, len(result.Matches))
	}

	// Check for expected files
	expectedFiles := map[string]bool{
		filepath.Join(tempDir, "file1.txt"):         true,
		filepath.Join(tempDir, "file2.txt"):         true,
		filepath.Join(tempDir, "subdir/nested.txt"): true,
	}

	for _, match := range result.Matches {
		if !expectedFiles[match.FilePath] {
			t.Errorf("Unexpected file in results: %s", match.FilePath)
		}
	}
}

// TestCaseInsensitiveSearch tests case-insensitive searching
func TestCaseInsensitiveSearch(t *testing.T) {
	tempDir, cleanup := setupTestFiles(t)
	defer cleanup()

	options := SearchOptions{
		RootDir:         tempDir,
		Pattern:         "APPLE",
		FileExtensions:  []string{".txt"},
		MaxFileSearches: 100,
		MaxResults:      100,
		CaseSensitive:   false,
	}

	result, err := Search(options)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedCount := 3
	if len(result.Matches) != expectedCount {
		t.Errorf("Expected %d matches, got %d", expectedCount, len(result.Matches))
	}
}

// TestFileExtensionFiltering tests filtering by file extension
func TestFileExtensionFiltering(t *testing.T) {
	tempDir, cleanup := setupTestFiles(t)
	defer cleanup()

	options := SearchOptions{
		RootDir:         tempDir,
		Pattern:         "line",
		FileExtensions:  []string{".go"}, // Should only match in Go files
		MaxFileSearches: 100,
		MaxResults:      100,
		CaseSensitive:   true,
	}

	result, err := Search(options)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedCount := 0 // "line" is in txt files but not in our .go file
	if len(result.Matches) != expectedCount {
		t.Errorf("Expected %d matches, got %d", expectedCount, len(result.Matches))
	}
}

// TestInvalidDirectory tests searching in a non-existent directory
func TestInvalidDirectory(t *testing.T) {
	tempDir, cleanup := setupTestFiles(t)
	defer cleanup()

	options := SearchOptions{
		RootDir:         filepath.Join(tempDir, "nonexistent"),
		Pattern:         "apple",
		FileExtensions:  []string{".txt"},
		MaxFileSearches: 100,
		MaxResults:      100,
		CaseSensitive:   true,
	}

	_, err := Search(options)
	if err == nil {
		t.Error("Expected error but got nil")
	}
}

// TestInvalidRegexPattern tests searching with an invalid regex pattern
func TestInvalidRegexPattern(t *testing.T) {
	tempDir, cleanup := setupTestFiles(t)
	defer cleanup()

	options := SearchOptions{
		RootDir:         tempDir,
		Pattern:         "a[", // Invalid regex
		FileExtensions:  []string{".txt"},
		MaxFileSearches: 100,
		MaxResults:      100,
		CaseSensitive:   true,
	}

	_, err := Search(options)
	if err == nil {
		t.Error("Expected error but got nil")
	}
}

// TestMaxResultsLimit tests limiting the number of results returned
func TestMaxResultsLimit(t *testing.T) {
	tempDir, cleanup := setupTestFiles(t)
	defer cleanup()

	options := SearchOptions{
		RootDir:         tempDir,
		Pattern:         "line|file|with", // Should match multiple lines
		FileExtensions:  []string{".txt"},
		MaxFileSearches: 100,
		MaxResults:      2, // Only get 2 results
		CaseSensitive:   true,
	}

	result, err := Search(options)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedCount := 2 // Should be limited to 2
	if len(result.Matches) != expectedCount {
		t.Errorf("Expected %d matches, got %d", expectedCount, len(result.Matches))
	}
}

// TestComplexRegex tests searching with a complex regex pattern
func TestComplexRegex(t *testing.T) {
	tempDir, cleanup := setupTestFiles(t)
	defer cleanup()

	// Create a subdirectory for this specific test
	complexDir := filepath.Join(tempDir, "complex_test_dir")
	if err := os.MkdirAll(complexDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create a specific test file with exactly the pattern we want to match
	complexTestFile := filepath.Join(complexDir, "complex_test.txt")
	complexContent := "This is line 1\nNot a match\nThis is line 3\nAlso not matching\n"
	if err := os.WriteFile(complexTestFile, []byte(complexContent), 0644); err != nil {
		t.Fatalf("Failed to create complex test file: %v", err)
	}

	options := SearchOptions{
		RootDir:         complexDir, // Search only in this subdirectory
		Pattern:         "^This.*line",
		FileExtensions:  []string{".txt"},
		MaxFileSearches: 100,
		MaxResults:      100,
		CaseSensitive:   true,
	}

	result, err := Search(options)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// We should find exactly 2 matches for the pattern "^This.*line"
	expectedCount := 2 // "This is line 1" and "This is line 3"
	if len(result.Matches) != expectedCount {
		t.Errorf("Expected %d matches, got %d", expectedCount, len(result.Matches))
		for i, match := range result.Matches {
			t.Logf("Match %d: %s:%d - %s", i, match.FilePath, match.LineNumber, match.LineContent)
		}
	}

	// Verify that we got exactly the right lines
	matchLines := map[string]bool{}
	for _, match := range result.Matches {
		matchLines[strings.TrimSpace(match.LineContent)] = true
	}

	if !matchLines["This is line 1"] {
		t.Error("Expected to find 'This is line 1' in matches")
	}
	if !matchLines["This is line 3"] {
		t.Error("Expected to find 'This is line 3' in matches")
	}
}

// TestSearchWithDefaults tests using default search options
func TestSearchWithDefaults(t *testing.T) {
	tempDir, cleanup := setupTestFiles(t)
	defer cleanup()

	// Test with default options
	opts := DefaultSearchOptions()
	opts.RootDir = tempDir
	opts.Pattern = "apple"

	result, err := Search(opts)
	if err != nil {
		t.Fatalf("Unexpected error with default options: %v", err)
	}

	expectedCount := 3
	if len(result.Matches) != expectedCount {
		t.Errorf("Expected %d matches with default options, got %d", expectedCount, len(result.Matches))
	}
}

// TestIsTextFile tests the text file detection logic
func TestIsTextFile(t *testing.T) {
	tempDir, cleanup := setupTestFiles(t)
	defer cleanup()

	tests := []struct {
		name     string
		filePath string
		expected bool
	}{
		{
			name:     "Text file",
			filePath: filepath.Join(tempDir, "file1.txt"),
			expected: true,
		},
		{
			name:     "Binary file with null bytes",
			filePath: filepath.Join(tempDir, "binary.bin"),
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			info, err := os.Stat(tc.filePath)
			if err != nil {
				t.Fatalf("Failed to stat file: %v", err)
			}

			result := isTextFile(tc.filePath, info)
			if result != tc.expected {
				t.Errorf("Expected isTextFile to return %v for %s, got %v",
					tc.expected, tc.filePath, result)
			}
		})
	}
}
