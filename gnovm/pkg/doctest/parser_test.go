package doctest

import (
	"strings"
	"testing"
)

func TestGetCodeBlocks(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected []codeBlock
	}{
		{
			name:  "Single code block with backticks",
			input: "```go\nfmt.Println(\"Hello, World!\")\n```",
			expected: []codeBlock{
				{
					content: "fmt.Println(\"Hello, World!\")",
					start:   6,
					end:     35,
					lang:    "go",
					index:   0,
				},
			},
		},
		{
			name:  "Single code block with additional backticks",
			input: "```go\nfmt.Println(\"Hello, World!\")\n``````",
			expected: []codeBlock{
				{
					content: "fmt.Println(\"Hello, World!\")",
					start:   6,
					end:     35,
					lang:    "go",
					index:   0,
				},
			},
		},
		{
			name:  "Single code block with tildes",
			input: "## Example\nprint hello world in go.\n~~~go\nfmt.Println(\"Hello, World!\")\n~~~",
			expected: []codeBlock{
				{
					content: "fmt.Println(\"Hello, World!\")",
					start:   42,
					end:     71,
					lang:    "go",
					index:   0,
				},
			},
		},
		{
			name:  "Multiple code blocks",
			input: "Here is some text.\n```python\ndef hello():\n    print(\"Hello, World!\")\n```\nSome more text.\n```javascript\nconsole.log(\"Hello, World!\");\n```",
			expected: []codeBlock{
				{
					content: "def hello():\n    print(\"Hello, World!\")",
					start:   29,
					end:     69,
					lang:    "python",
					index:   0,
				},
				{
					content: "console.log(\"Hello, World!\");",
					start:   103,
					end:     133,
					lang:    "javascript",
					index:   1,
				},
			},
		},
		{
			name:  "Code block with no language specifier",
			input: "```\nfmt.Println(\"Hello, World!\")\n```",
			expected: []codeBlock{
				{
					content: "fmt.Println(\"Hello, World!\")",
					start:   4,
					end:     33,
					lang:    "plain",
					index:   0,
				},
			},
		},
		{
			name:     "No code blocks",
			input:    "Just some text without any code blocks.",
			expected: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := GetCodeBlocks(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Failed %s: expected %d code blocks, got %d", tt.name, len(tt.expected), len(result))
			}

			for i, res := range result {
				if normalize(res.content) != normalize(tt.expected[i].content) {
					t.Errorf("Failed %s: expected content %s, got %s", tt.name, tt.expected[i].content, res.content)
				}

				if res.start != tt.expected[i].start {
					t.Errorf("Failed %s: expected start %d, got %d", tt.name, tt.expected[i].start, res.start)
				}

				if res.end != tt.expected[i].end {
					t.Errorf("Failed %s: expected end %d, got %d", tt.name, tt.expected[i].end, res.end)
				}

				if res.lang != tt.expected[i].lang {
					t.Errorf("Failed %s: expected type %s, got %s", tt.name, tt.expected[i].lang, res.lang)
				}

				if res.index != tt.expected[i].index {
					t.Errorf("Failed %s: expected index %d, got %d", tt.name, tt.expected[i].index, res.index)
				}
			}
		})
	}
}

// ignore whitespace in the source code
func normalize(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, "\n", ""), "\r", ""), "\t", ""), " ", "")
}
