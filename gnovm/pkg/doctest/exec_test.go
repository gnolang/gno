package doctest

import (
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
				T: "go",
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
				T: "go",
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
				T: "go",
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
			res, err := executeCodeBlock(tc.codeBlock)
			if tc.isErr && err == nil {
				t.Errorf("%s did not return an error", tc.name)
			}

			if res != tc.expected {
				t.Errorf("%s = %v, want %v", tc.name, res, tc.expected)
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
			_, _ = executeCodeBlock(tc.codeBlock)
		})
	}
}
