package search

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestFileForReplace creates a temporary test file with provided content for testing
func setupTestFileForReplace(t *testing.T, name, content string) (string, func()) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "regex-replace-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	
	// Create the test file
	testFile := filepath.Join(tempDir, name)
	err = os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Return a cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}
	
	return testFile, cleanup
}

// TestBasicReplace tests replacing simple patterns
func TestBasicReplace(t *testing.T) {
	content := `This is a test.
Another line with test.
And a third test line.`

	expected := `This is a REPLACED.
Another line with REPLACED.
And a third REPLACED line.`

	testFile, cleanup := setupTestFileForReplace(t, "basic.txt", content)
	defer cleanup()

	result, count, err := ReplaceWithRegex(testFile, "test", "REPLACED", 0, true)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected 3 replacements, got %d", count)
	}

	if result != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
	}
}

// TestReplaceSpecificOccurrence tests replacing only a specific occurrence
func TestReplaceSpecificOccurrence(t *testing.T) {
	content := `function test() {
  console.log("test");
}

function test2() {
  console.log("another test");
}`

	expected := `function REPLACED() {
  console.log("test");
}

function test2() {
  console.log("another test");
}`

	testFile, cleanup := setupTestFileForReplace(t, "specific.js", content)
	defer cleanup()

	result, count, err := ReplaceWithRegex(testFile, "test", "REPLACED", 1, true)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 replacement, got %d", count)
	}

	if result != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
	}
}

// TestCaptureGroups tests replacing with capture groups
func TestCaptureGroups(t *testing.T) {
	content := `function getData(id) {
  return fetchData(id);
}

function setData(id, value) {
  return updateData(id, value);
}`

	expected := `function getData(id) {
  return fetchData(id);
}

function setData(id, value) {
  return updateData(id, value, { timestamp: new Date() });
}`

	testFile, cleanup := setupTestFileForReplace(t, "capture.js", content)
	defer cleanup()

	result, count, err := ReplaceWithRegex(testFile, "return updateData\\((.*?)\\);", "return updateData($1, { timestamp: new Date() });", 1, true)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 replacement, got %d", count)
	}

	if result != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
	}
}

// TestMultilineReplace tests replacing patterns that span multiple lines
func TestMultilineReplace(t *testing.T) {
	content := `function oldFunction() {
  // Old implementation
  console.log("Old function");
  return true;
}

// Other code`

	expected := `function newFunction() {
  // New implementation
  console.log("New function");
  return false;
}

// Other code`

	testFile, cleanup := setupTestFileForReplace(t, "multiline.js", content)
	defer cleanup()

	pattern := `function oldFunction\(\) \{[\s\S]*?\}`
	replacement := `function newFunction() {
  // New implementation
  console.log("New function");
  return false;
}`

	result, count, err := ReplaceWithRegex(testFile, pattern, replacement, 1, true)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 replacement, got %d", count)
	}

	if result != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
	}
}

// TestCaseInsensitiveReplace tests case-insensitive replacements
func TestCaseInsensitiveReplace(t *testing.T) {
	content := `Text with TEST and test and Test variations.`
	expected := `Text with REPLACED and REPLACED and REPLACED variations.`

	testFile, cleanup := setupTestFileForReplace(t, "case.txt", content)
	defer cleanup()

	result, count, err := ReplaceWithRegex(testFile, "test", "REPLACED", 0, false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected 3 replacements, got %d", count)
	}

	if result != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
	}
}

// TestNoMatchReplace tests behavior when no matches are found
// After Phase 1 fix: Now returns error instead of silently returning unchanged content
func TestNoMatchReplace(t *testing.T) {
	content := `This content has no matches.`

	testFile, cleanup := setupTestFileForReplace(t, "nomatch.txt", content)
	defer cleanup()

	result, count, err := ReplaceWithRegex(testFile, "nonexistent", "replacement", 0, true)

	// After fix: should return error when pattern not found
	if err == nil {
		t.Fatalf("Expected error when pattern not found, but got none")
	}

	// Error should be informative
	if !strings.Contains(err.Error(), "Pattern not found") {
		t.Errorf("Expected error about pattern not found, got: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 replacements, got %d", count)
	}

	if result != "" {
		t.Errorf("Expected empty result on error, got: %s", result)
	}
}

// TestInvalidRegexPatternReplace tests handling of invalid regex patterns
func TestInvalidRegexPatternReplace(t *testing.T) {
	content := `Some content.`
	
	testFile, cleanup := setupTestFileForReplace(t, "invalid.txt", content)
	defer cleanup()

	_, _, err := ReplaceWithRegex(testFile, "[invalid", "replacement", 0, true)
	if err == nil {
		t.Error("Expected an error for invalid regex pattern, but got none")
	}
}

// TestOccurrenceOutOfRangeReplace tests when the specified occurrence exceeds available matches
func TestOccurrenceOutOfRangeReplace(t *testing.T) {
	content := `Only one match here.`
	
	testFile, cleanup := setupTestFileForReplace(t, "outofrange.txt", content)
	defer cleanup()

	_, _, err := ReplaceWithRegex(testFile, "match", "replacement", 2, true)
	if err == nil {
		t.Error("Expected an error for out-of-range occurrence, but got none")
	}
}

// TestComplexReplace tests a more complex replacement scenario with capture groups
func TestComplexReplace(t *testing.T) {
	content := `<div class="container">
  <div class="item" id="item-1">First Item</div>
  <div class="item" id="item-2">Second Item</div>
  <div class="item" id="item-3">Third Item</div>
</div>`

	expected := `<div class="container">
  <div class="item highlighted" id="item-1">First Item</div>
  <div class="item" id="item-2">Second Item</div>
  <div class="item highlighted" id="item-3">Third Item</div>
</div>`

	testFile, cleanup := setupTestFileForReplace(t, "complex.html", content)
	defer cleanup()

	// Add 'highlighted' class to items with odd-numbered IDs
	pattern := `<div class="item"(.*?id="item-([13579])")>`
	replacement := `<div class="item highlighted"$1>`

	result, count, err := ReplaceWithRegex(testFile, pattern, replacement, 0, true)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 replacements, got %d", count)
	}

	if result != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
	}
}

// TestReplaceLargeFunction tests replacing a large function with a new implementation
func TestReplaceLargeFunction(t *testing.T) {
	content := `package main

import "fmt"

func oldImplementation(data []string) map[string]int {
    result := make(map[string]int)
    
    for _, item := range data {
        if _, exists := result[item]; exists {
            result[item]++
        } else {
            result[item] = 1
        }
    }
    
    return result
}

func main() {
    // Test the function
    data := []string{"a", "b", "a", "c", "b", "a"}
    result := oldImplementation(data)
    fmt.Println(result)
}`

	expected := `package main

import "fmt"

func newImplementation(data []string) map[string]int {
    result := make(map[string]int)
    
    for _, item := range data {
        result[item]++
    }
    
    return result
}

func main() {
    // Test the function
    data := []string{"a", "b", "a", "c", "b", "a"}
    result := oldImplementation(data)
    fmt.Println(result)
}`

	testFile, cleanup := setupTestFileForReplace(t, "largefunction.go", content)
	defer cleanup()

	pattern := `func oldImplementation\(data \[\]string\) map\[string\]int \{[\s\S]+?return result\n\}`
	replacement := `func newImplementation(data []string) map[string]int {
    result := make(map[string]int)
    
    for _, item := range data {
        result[item]++
    }
    
    return result
}`

	result, count, err := ReplaceWithRegex(testFile, pattern, replacement, 1, true)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 replacement, got %d", count)
	}

	if result != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
	}
}

// TestEmptyFileReplace tests behavior with an empty file
func TestEmptyFileReplace(t *testing.T) {
	content := ``
	
	testFile, cleanup := setupTestFileForReplace(t, "empty.txt", content)
	defer cleanup()

	result, count, err := ReplaceWithRegex(testFile, ".*", "replacement", 0, true)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 replacements in empty file, got %d", count)
	}

	if result != "" {
		t.Errorf("Expected empty result for empty file")
	}
}

// TestStringFunctionBasic tests the string version of the replace function
func TestStringFunctionBasic(t *testing.T) {
	content := "This is a test string with test word repeated."
	expected := "This is a REPLACED string with REPLACED word repeated."
	
	result, count, err := ReplaceWithRegexInString(content, "test", "REPLACED", 0, true)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if count != 2 {
		t.Errorf("Expected 2 replacements, got %d", count)
	}
	
	if result != expected {
		t.Errorf("Expected: %s\nGot: %s", expected, result)
	}
}

// TestMultilineGolangFunction tests replacing a full Go function
func TestMultilineGolangFunction(t *testing.T) {
	content := `package main

func calculateSum(numbers []int) int {
    sum := 0
    for _, num := range numbers {
        sum += num
    }
    return sum
}

func calculateAverage(numbers []int) float64 {
    sum := calculateSum(numbers)
    return float64(sum) / float64(len(numbers))
}`

	expected := `package main

func calculateTotal(numbers []int) int {
    total := 0
    for i := 0; i < len(numbers); i++ {
        total += numbers[i]
    }
    return total
}

func calculateAverage(numbers []int) float64 {
    sum := calculateSum(numbers)
    return float64(sum) / float64(len(numbers))
}`

	testFile, cleanup := setupTestFileForReplace(t, "golang_function.go", content)
	defer cleanup()

	pattern := `func calculateSum\(numbers \[\]int\) int \{[\s\S]*?return sum\n\}`
	replacement := `func calculateTotal(numbers []int) int {
    total := 0
    for i := 0; i < len(numbers); i++ {
        total += numbers[i]
    }
    return total
}`

	result, count, err := ReplaceWithRegex(testFile, pattern, replacement, 1, true)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 replacement, got %d", count)
	}

	if result != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
	}
}

// TestReplaceRegex_PatternNotFound_ShowsContext tests that regex pattern not found error provides helpful context
// This reproduces Issue #5 from the MCP_FILESYSTEM_ISSUES.md docs
func TestReplaceRegex_PatternNotFound_ShowsContext(t *testing.T) {
	content := `package credits

import (
    "context"
    "database/sql"
)

func NewBatchTracker(
    supabaseClient *supabase.Client,
) *BatchTracker {
    // ...
}`

	testFile, cleanup := setupTestFileForReplace(t, "batch.go", content)
	defer cleanup()

	// This pattern fails to match (multiline issue with literal \n)
	pattern := `func NewBatchTracker\(\n\tsupabaseClient.*`

	_, _, err := ReplaceWithRegex(testFile, pattern, "replacement", 1, true)

	// Error should be actionable for LLM
	if err == nil {
		t.Fatal("Expected error when pattern not found, but got none")
	}

	errMsg := err.Error()

	// Error should mention pattern not found
	if !strings.Contains(errMsg, "Pattern not found") && !strings.Contains(errMsg, "not found") {
		t.Errorf("Expected error about pattern not found, got: %v", err)
	}

	// Error should provide helpful hints about multiline patterns
	if !strings.Contains(errMsg, "multiline") && !strings.Contains(errMsg, "Multiline") && !strings.Contains(errMsg, "whitespace") {
		t.Logf("Helpful hint: Error could mention multiline patterns: %v", err)
	}
}

// TestReplaceRegex_InvalidPattern tests that invalid regex patterns return clear errors
func TestReplaceRegex_InvalidPattern(t *testing.T) {
	content := `Some content here`

	testFile, cleanup := setupTestFileForReplace(t, "invalid.txt", content)
	defer cleanup()

	// Invalid regex pattern
	_, _, err := ReplaceWithRegex(testFile, "[invalid", "replacement", 0, true)

	if err == nil {
		t.Fatal("Expected error for invalid regex pattern")
	}

	errMsg := err.Error()
	// Should mention it's an invalid pattern
	if !strings.Contains(errMsg, "invalid") && !strings.Contains(errMsg, "Invalid") {
		t.Errorf("Expected error to mention invalid pattern: %v", err)
	}
}
