package doctest

import (
	"testing"
)

func TestExecuteCodeBlock(t *testing.T) {
	codeBlock := CodeBlock{
		Content: "package main\n\nfunc main() { println(\"Hello, World!\") }",
		Start:   0,
		End:     50,
		T:       "go",
		Index:   0,
	}

	err := writeCodeBlockToFile(codeBlock)
	if err != nil {
		t.Errorf("Failed to write code block to file: %v", err)
	}

	res, err := executeCodeBlock(codeBlock)
	if err != nil {
		t.Errorf("Failed to execute code block: %v", err)
	}

	if res != "Hello, World!\n" {
		t.Errorf("Expected 'Hello, World!', got %s", res)
	}
}
