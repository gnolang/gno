package doctest

import (
	"reflect"
	"testing"
)

func TestExecuteCodeBlock(t *testing.T) {
	tests := []struct {
		name      string
		codeBlock CodeBlock
		expected  string
		isErr     bool
	}{
		{
			name: "Hello, World!",
			codeBlock: CodeBlock{
				Content: `
package main

func main() {
	println("Hello, World!")
}`,
				T:       "go",
				Package: "main",
			},
			expected: "Hello, World!\n",
		},
		{
			name: "Multiple prints",
			codeBlock: CodeBlock{
				Content: `
package main

func main() {
	println("Hello");
	println("World")
}`,
				T:       "go",
				Package: "main",
			},
			expected: "Hello\nWorld\n",
		},
		{
			name: "Print variables",
			codeBlock: CodeBlock{
				Content: `
package main

func main() {
	a := 10
	b := 20
	println(a + b)
}`,
				T:       "go",
				Package: "main",
			},
			expected: "30\n",
		},
		{
			name: "unsupported language",
			codeBlock: CodeBlock{
				Content: `
data Tree a = Empty | Node a (Tree a) (Tree a)
    deriving (Eq, Show)

data Direction = LH | RH
    deriving (Eq, Show)

splay :: (Ord a) => a -> Tree a -> Tree a
splay a t = rebuild $ path a t [(undefined,t)]`,
				T: "haskell",
			},
			isErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// res, err := executeCodeBlock(tc.codeBlock)
			res, err := ExecuteCodeBlock(tc.codeBlock)
			if tc.isErr && err == nil {
				t.Errorf("%s did not return an error", tc.name)
			}

			if res != tc.expected {
				t.Errorf("%s = %v, want %v", tc.name, res, tc.expected)
			}

			if tc.codeBlock.T == "go" {
				if tc.codeBlock.Package != "" {
					if tc.codeBlock.Package != "main" {
						t.Errorf("%s = %v, want %v", tc.name, tc.codeBlock.Package, "main")
					}
				}
			}
		})
	}
}

func TestExecuteCodeBlock_ImportPackage(t *testing.T) {
	t.Skip("skipping test for now")
	tests := []struct {
		name      string
		codeBlock CodeBlock
		expected  string
	}{
		{
			name: "import go stdlib package",
			codeBlock: CodeBlock{
				Content: `package main

import (
	"strings"
)

func main() {
	println(strings.Join([]string{"Hello", "World"}, ", "))
}`,
				T:       "go",
				Package: "main",
			},
			expected: "Hello, World\n",
		},
		{
			name: "import realm",
			codeBlock: CodeBlock{
				Content: `package main

import (
	"gno.land/p/demo/ufmt"
)

func main() {
	ufmt.Println("Hello, World!")
}`,
				T:       "go",
				Package: "main",
			},
			expected: "Hello, World!\n",
		},
	}

	for _, tt := range tests {
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

func TestExecuteCodeBlock_ShouldPanic(t *testing.T) {
	tests := []struct {
		name      string
		codeBlock CodeBlock
	}{
		{
			name: "syntax error",
			codeBlock: CodeBlock{
				Content: "package main\n\nfunc main() { println(\"Hello, World!\")",
				T:       "go",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("%s did not panic", tc.name)
				}
			}()
			_, _ = ExecuteCodeBlock(tc.codeBlock)
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
