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

	markdownContent := `# Go Code Examples

This document contains two simple examples written in Go.

## Example 1: Fibonacci Sequence

The first example prints the first 10 numbers of the Fibonacci sequence.

` + "```go" + `
package main

func main() {
    a, b := 0, 1
    for i := 0; i < 10; i++ {
        println(a)
        a, b = b, a+b
    }
}
` + "```" + `

## Example 2: String Reversal

The second example reverses a given string and prints it.

` + "```go" + `
package main

func main() {
    str := "Hello, Go!"
    runes := []rune(str)
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    println(string(runes))
}
` + "```" + `

These two examples demonstrate basic Go functionality without using concurrency, generics, or reflect.` + "## std Package" + `
` + "```go" + `
package main

import (
	"std"
)

func main() {
    addr := std.GetOrigCaller()
    println(addr)
}
`

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
			stdoutShouldContain: "0\n1\n1\n2\n3\n5\n8\n13\n21\n34\n\n",
		},
		{
			args:                []string{"doctest", "-path", mdFilePath, "-index", "1"},
			stdoutShouldContain: "!oG ,olleH\n",
		},
		{
			args:                []string{"doctest", "-path", mdFilePath, "-index", "2"},
			stdoutShouldContain: "g14ch5q26mhx3jk5cxl88t278nper264ces4m8nt\n",
		},
	}

	testMainCaseRun(t, tc)
}
