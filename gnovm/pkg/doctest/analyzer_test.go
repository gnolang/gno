package doctest

import (
	"testing"
)

func TestAnalyzeAndModifyCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "simple hello world without package and import",
			input: `
func main() {
	println("Hello, World")
}`,
			expected: `package main

func main() {
	println("Hello, World")
}
`,
		},
		{
			name: "main with address without package",
			input: `
import (
	"std"
)

func main() {
	addr := std.GetOrigCaller()
	println(addr)
}`,
			expected: `package main

import (
	"std"
)

func main() {
	addr := std.GetOrigCaller()
	println(addr)
}
`,
		},
		{
			name: "multiple imports without package and import statement",
			input: `
import (
	"math"
	"strings"
)

func main() {
	println(math.Pi)
	println(strings.ToUpper("Hello, World"))
}`,
			expected: `package main

import (
	"math"
	"strings"
)

func main() {
	println(math.Pi)
	println(strings.ToUpper("Hello, World"))
}
`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			modifiedCode, err := analyzeAndModifyCode(tt.input)
			if err != nil {
				t.Fatalf("AnalyzeAndModifyCode(%s) returned error: %v", tt.name, err)
			}
			if modifiedCode != tt.expected {
				t.Errorf("AnalyzeAndModifyCode(%s) = %v, want %v", tt.name, modifiedCode, tt.expected)
			}
		})
	}
}

func TestAnalyzeAndModifyCodeWithConflictingNames(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "local identifier with same name as stdlib package",
			input: `package main

func main() {
	math := 42
	println(math)
}`,
			expected: `package main

func main() {
	math := 42
	println(math)
}
`,
		},
		{
			name: "local function with same name as stdlib package",
			input: `
package main

func strings() string {
	return "local strings function"
}

func main() {
	println(strings())
}`,
			expected: `package main

func strings() string {
	return "local strings function"
}

func main() {
	println(strings())
}
`,
		},
		{
			name: "mixed use of local and stdlib identifiers",
			input: `package main

import (
	"fmt"
)

func strings() string {
    return "local strings function"
}

func main() {
	strings := strings()
	fmt.Println(strings)
	fmt.Println(strings.ToUpper("hello"))
}`,
			expected: `package main

import (
	"fmt"
	"strings"
)

func strings() string {
	return "local strings function"
}

func main() {
	strings := strings()
	fmt.Println(strings)
	fmt.Println(strings.ToUpper("hello"))
}
`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			modifiedCode, err := analyzeAndModifyCode(tt.input)
			if err != nil {
				t.Fatalf("AnalyzeAndModifyCode(%s) returned error: %v", tt.name, err)
			}
			if modifiedCode != tt.expected {
				t.Errorf("AnalyzeAndModifyCode(%s) = %v, want %v", tt.name, modifiedCode, tt.expected)
			}
		})
	}
}
