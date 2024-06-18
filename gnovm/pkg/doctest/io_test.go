package doctest

import (
	"os"
	"testing"
)

func TestReadMarkdownFile(t *testing.T) {
	t.Parallel()

	tmpFile, err := os.CreateTemp("", "*.md")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	expectedContent := "# Test Markdown\nThis is a test."
	if _, err := tmpFile.WriteString(expectedContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Test: Read the content of the temporary markdown file
	content, err := ReadMarkdownFile(tmpFile.Name())
	if err != nil {
		t.Errorf("ReadMarkdownFile returned an error: %v", err)
	}
	if content != expectedContent {
		t.Errorf("ReadMarkdownFile content mismatch. Got %v, want %v", content, expectedContent)
	}

	// Test: Attempt to read a non-existent file
	_, err = ReadMarkdownFile("non_existent_file.md")
	if err == nil {
		t.Error("ReadMarkdownFile did not return an error for a non-existent file")
	}
}
