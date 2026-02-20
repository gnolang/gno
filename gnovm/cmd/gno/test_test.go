package main

import (
	"path/filepath"
	"testing"
)

func TestTestApp(t *testing.T) {
	tc := []testMainCase{
		// Basic test command
		{
			args:                []string{"test"},
			stderrShouldContain: "[no test files]",
		},
		// Test valid package
		{
			args:                []string{"test", "./testdata/coverage"},
			stderrShouldContain: "ok",
		},
		// Test with verbose flag
		{
			args:                []string{"test", "-v", "./testdata/coverage"},
			stderrShouldContain: "=== RUN",
		},
		// Test with run filter
		{
			args:                []string{"test", "-run", "TestAdd", "./testdata/coverage"},
			stderrShouldContain: "ok",
		},
		// Test with non-existent package
		{
			args:             []string{"test", "non_existent_package"},
			errShouldContain: "no such file or directory",
		},
		// Test with print-runtime-metrics
		{
			args:                []string{"test", "-print-runtime-metrics", "./testdata/coverage"},
			stderrShouldContain: "runtime:",
		},
	}

	testMainCaseRun(t, tc)
}

func TestTestCoverageApp(t *testing.T) {
	tempDir := t.TempDir()
	coverProfile := filepath.Join(tempDir, "coverage.out")

	tc := []testMainCase{
		// Basic coverage test
		{
			args:                []string{"test", "-cover", "./testdata/coverage"},
			stderrShouldContain: "Coverage Report",
		},
		// Coverage with verbose
		{
			args:                []string{"test", "-cover", "-v", "./testdata/coverage"},
			stderrShouldContain: "Coverage Report",
		},
		// Coverage with profile output
		{
			args:                []string{"test", "-cover", "-coverprofile", coverProfile, "./testdata/coverage"},
			stderrShouldContain: "Coverage Report",
		},
		// Coverage with verbose and profile
		{
			args:                []string{"test", "-cover", "-v", "-coverprofile", coverProfile, "./testdata/coverage"},
			stderrShouldContain: "Coverage report written to:",
		},
		// Coverage with show flag
		{
			args:                []string{"test", "-cover", "-show", "main.gno", "./testdata/coverage"},
			stderrShouldContain: "Coverage Report",
		},
		// Coverage with show all files
		{
			args:                []string{"test", "-cover", "-show", "*.gno", "./testdata/coverage"},
			stderrShouldContain: "Coverage Report",
		},
	}

	testMainCaseRun(t, tc)
}
