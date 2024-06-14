package doctest

import (
	"reflect"
	"testing"
)

func TestGetCodeBlocks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []CodeBlock
	}{
		{
			name:  "Single code block",
			input: "```go\nfmt.Println(\"Hello, World!\")\n```",
			expected: []CodeBlock{
				{
					Content: "fmt.Println(\"Hello, World!\")",
					Start:   0,
					End:     38,
					T:       "go",
					Index:   0,
				},
			},
		},
		{
			name:  "Multiple code blocks",
			input: "Here is some text.\n```python\ndef hello():\n    print(\"Hello, World!\")\n```\nSome more text.\n```javascript\nconsole.log(\"Hello, World!\");\n```",
			expected: []CodeBlock{
				{
					Content: "def hello():\n    print(\"Hello, World!\")",
					Start:   19,
					End:     72,
					T:       "python",
					Index:   0,
				},
				{
					Content: "console.log(\"Hello, World!\");",
					Start:   89,
					End:     136,
					T:       "javascript",
					Index:   1,
				},
			},
		},
		{
			name:  "Code block with no language specifier",
			input: "```\nfmt.Println(\"Hello, World!\")\n```",
			expected: []CodeBlock{
				{
					Content: "fmt.Println(\"Hello, World!\")",
					Start:   0,
					End:     36,
					T:       "plain",
					Index:   0,
				},
			},
		},
		{
			name:     "No code blocks",
			input:    "Just some text without any code blocks.",
			expected: nil,
		},
		{
			name:     "malformed code block",
			input:    "```go\nfmt.Println(\"Hello, World!\")",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCodeBlocks(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Failed %s: expected %d code blocks, got %d", tt.name, len(tt.expected), len(result))
			}
			for i, res := range result {
				if !reflect.DeepEqual(res, tt.expected[i]) {
					t.Errorf("Failed %s: expected %v, got %v", tt.name, tt.expected[i], res)
				}
			}
		})
	}
}
