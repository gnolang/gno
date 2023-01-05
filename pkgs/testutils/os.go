package testutils

import (
	"os"
	"testing"
)

// NewTestCaseDir creates a new temporary directory for a test case.
// Returns the directory path and a cleanup function.
func NewTestCaseDir(t *testing.T) (string, func()) {
	t.Helper()

	dir, err := os.MkdirTemp("", t.Name()+"_")
	if err != nil {
		t.Fatalf("unable to generate temporary directory, %v", err)
	}

	return dir, func() { _ = os.RemoveAll(dir) }
}

// NewTestFile creates a new temporary file for a test case
func NewTestFile(t *testing.T) (*os.File, func()) {
	t.Helper()

	file, err := os.CreateTemp("", t.Name()+"-")
	if err != nil {
		t.Fatalf(
			"unable to create a temporary output file, %v",
			err,
		)
	}

	return file, func() { _ = os.RemoveAll(file.Name()) }
}
