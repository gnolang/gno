package testutils

import (
	"os"
	"strings"
	"testing"
)

// NewTestCaseDir creates a new temporary directory for a test case.
// Returns the directory path and a cleanup function.
func NewTestCaseDir(t *testing.T) (string, func()) {
	t.Helper()

	// Replace any restricted character with safe ones (for nested tests)
	pattern := strings.ReplaceAll(t.Name()+"_", "/", "_")
	dir, err := os.MkdirTemp("", pattern)
	if err != nil {
		t.Fatalf("unable to generate temporary directory, %v", err)
	}

	return dir, func() { _ = os.RemoveAll(dir) }
}

// NewTestFile creates a new temporary file for a test case
func NewTestFile(t *testing.T) (*os.File, func()) {
	t.Helper()

	// Replace any restricted character with safe ones (for nested tests)
	pattern := strings.ReplaceAll(t.Name()+"-", "/", "_")
	file, err := os.CreateTemp("", pattern)
	if err != nil {
		t.Fatalf(
			"unable to create a temporary output file, %v",
			err,
		)
	}

	return file, func() { _ = os.RemoveAll(file.Name()) }
}
