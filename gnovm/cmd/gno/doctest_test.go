package main

import (
	"os"
	"testing"
)

func TestDoctest(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "doctest-test")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	markdownContent := "## Example\nprint hello world in gno.\n```go\npackage main\n\nfunc main() {\nprintln(\"Hello, World!\")\n}\n```"

	mdFile, err := os.CreateTemp(tempDir, "sample-*.md")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer mdFile.Close()

	_, err = mdFile.WriteString(markdownContent)
	if err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}

	mdFilePath := mdFile.Name()

	tc := []testMainCase{
		{
			args:        []string{"doctest -h"},
			errShouldBe: "flag: help requested",
		},
		{
			args:                []string{"doctest", "-path", mdFilePath, "-index", "0"},
			stdoutShouldContain: "Hello, World!\n",
		},
	}

	testMainCaseRun(t, tc)
}
