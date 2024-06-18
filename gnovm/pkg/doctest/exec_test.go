package doctest

import (
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
		{
			name: "print multiple values",
			codeBlock: CodeBlock{
				Content: `
package main

func main() {
	count := 3
	for i := 0; i < count; i++ {
		println("Hello")
	}
}`,
				T:       "gno",
				Package: "main",
			},
			expected: "Hello\nHello\nHello\n",
		},
		{
			name: "import multiple go stdlib packages",
			codeBlock: CodeBlock{
				Content: `
package main

import (
	"math"
	"strings"
)

func main() {
	println(math.Pi)
	println(strings.ToUpper("Hello, World"))
}`,
				T:       "gno",
				Package: "main",
			},
			expected: "3.141592653589793\nHELLO, WORLD\n",
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
