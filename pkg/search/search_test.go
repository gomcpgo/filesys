package search

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTestFiles creates temporary test files for searching
func setupTestFiles(t *testing.T) (string, func()) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "search-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create test files
	testFiles := map[string]string{
		"file1.txt": "This is line 1\nThis contains apple\nThis is line 3\n",
		"file2.txt": "Another file\nWith an apple and orange\nAnd more content\n",
		"file3.go":  "package main\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}\n",
		"binary.bin": string([]byte{0, 1, 2, 3, 4, 5}),
		"large.txt": "This is a test line\n",
		"subdir/nested.txt": "Nested file\nWith apple content\n",
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

func TestSearch(t *testing.T) {
	// Setup test files
	tempDir, cleanup := setupTestFiles(t)
	defer cleanup()

	tests := []struct {
		name           string
		options        SearchOptions
		expectedCount  int
		expectedError  bool
		checkFilePaths bool
	}{
		{
			name: "Basic search for 'apple'",
			options: SearchOptions{
				RootDir:         tempDir,
				Pattern:         "apple",
				FileExtensions:  []string{".txt", ".go"},
				MaxFileSearches: 100,
				MaxResults:      100,
				CaseSensitive:   true,
			},
			expectedCount:  3,
			expectedError:  false,
			checkFilePaths: true,
		},
		{
			name: "Case insensitive search",
			options: SearchOptions{
				RootDir:         tempDir,
				Pattern:         "APPLE",
				FileExtensions:  []string{".txt"},
				CaseSensitive:   false,
			},
			expectedCount: 3,
			expectedError: false,
		},
		{
			name: "Limit file extensions",
			options: SearchOptions{
				RootDir:         tempDir,
				Pattern:         "line",
				FileExtensions:  []string{".go"}, // Should only match in Go files
				CaseSensitive:   true,
			},
			expectedCount: 0, // "line" is in txt files but not in our .go file
			expectedError: false,
		},
		{
			name: "Invalid directory",
			options: SearchOptions{
				RootDir:         filepath.Join(tempDir, "nonexistent"),
				Pattern:         "apple",
				CaseSensitive:   true,
			},
			expectedCount: 0,
			expectedError: true,
		},
		{
			name: "Invalid regex pattern",
			options: SearchOptions{
				RootDir:         tempDir,
				Pattern:         "a[", // Invalid regex
				CaseSensitive:   true,
			},
			expectedCount: 0,
			expectedError: true,
		},
		{
			name: "Max results limit",
			options: SearchOptions{
				RootDir:         tempDir,
				Pattern:         "line|file|with", // Should match multiple lines
				FileExtensions:  []string{".txt"},
				MaxResults:      2, // Only get 2 results
				CaseSensitive:   true,
			},
			expectedCount: 2, // Should be limited to 2
			expectedError: false,
		},
		{
			name: "Complex regex",
			options: SearchOptions{
				RootDir:         tempDir,
				Pattern:         "^This.*line",
				FileExtensions:  []string{".txt"},
				CaseSensitive:   true,
			},
			expectedCount: 2, // "This is line 1" and "This is line 3"
			expectedError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Run the search
			result, err := Search(tc.options)

			// Check error
			if tc.expectedError && err == nil {
				t.Errorf("Expected error but got nil")
			}
			if !tc.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// If expecting an error, don't check the results
			if tc.expectedError {
				return
			}

			// Check number of matches
			if len(result.Matches) != tc.expectedCount {
				t.Errorf("Expected %d matches, got %d", tc.expectedCount, len(result.Matches))
			}

			// Check for expected files in the first test case
			if tc.checkFilePaths && len(result.Matches) > 0 {
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
		})
	}
}

func TestSearchWithDefaults(t *testing.T) {
	// Setup test files
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

	if len(result.Matches) != 3 {
		t.Errorf("Expected 3 matches with default options, got %d", len(result.Matches))
	}
}

func TestIsTextFile(t *testing.T) {
	// Setup test files
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
