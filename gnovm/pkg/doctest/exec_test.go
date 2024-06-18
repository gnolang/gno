package doctest

import (
	"reflect"
	"testing"
)

func TestExecuteCodeBlock(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		codeBlock CodeBlock
		expected  string
	}{
		{
			name: "import go stdlib package",
			codeBlock: CodeBlock{
				Content: `
package main

func main() {
	println("Hello, World")
}`,
				T:       "gno",
				Package: "main",
			},
			expected: "Hello, World\n",
		},
		{
			name: "import go stdlib package",
			codeBlock: CodeBlock{
				Content: `
package main

import "std"

func main() {
	addr := std.GetOrigCaller()
	println(addr)
}`,
				T:       "gno",
				Package: "main",
			},
			expected: "g14ch5q26mhx3jk5cxl88t278nper264ces4m8nt\n",
		},
		{
			name: "import go stdlib package",
			codeBlock: CodeBlock{
				Content: `
package main

import "strings"

func main() {
	println(strings.ToUpper("Hello, World"))
}`,
				T:       "gno",
				Package: "main",
			},
			expected: "HELLO, WORLD\n",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			res, err := ExecuteCodeBlock(tt.codeBlock)
			if err != nil {
				t.Errorf("%s returned an error: %v", tt.name, err)
			}

			if res != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, res, tt.expected)
			}
		})
	}
}

func TestExtractOptions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "Ignore option",
			input: `
//gno: ignore
package main

func main() {
	println("This code should be ignored")
}
`,
			expected: []string{"ignore"},
		},
		{
			name: "No run option",
			input: `
//gno: no_run
package main

func main() {
	println("This code should not run")
}
`,
			expected: []string{"no_run"},
		},
		{
			name: "Should panic option",
			input: `
//gno: should_panic
package main

func main() {
	panic("Expected panic")
}
`,
			expected: []string{"should_panic"},
		},
		{
			name: "No options",
			input: `
package main

func main() {
	println("No options")
}
`,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codeBlock := CodeBlock{
				Content: tt.input,
				T:       "go",
			}
			options := extractOptions(codeBlock.Content)
			if !reflect.DeepEqual(options, tt.expected) {
				t.Errorf("got %v, want %v", options, tt.expected)
			}
		})
	}
}
