package doctest

import (
	"testing"
)

func TestAnalyzeAndModifyCode(t *testing.T) {
	t.Skip()
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
		t.Run(tt.name, func(t *testing.T) {
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
