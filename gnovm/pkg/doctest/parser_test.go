package doctest

import (
	"os"
	"strings"
	"testing"
)

func TestGetCodeBlocks(t *testing.T) {
	t.Parallel()
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
					Start:   6,
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
					Start:   29,
					End:     69,
					T:       "python",
					Index:   0,
				},
				{
					Content: "console.log(\"Hello, World!\");",
					Start:   103,
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
					Start:   4,
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
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			result := getCodeBlocks(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Failed %s: expected %d code blocks, got %d", tt.name, len(tt.expected), len(result))
			}

			for i, res := range result {
				if normalize(res.Content) != normalize(tt.expected[i].Content) {
					t.Errorf("Failed %s: expected content %s, got %s", tt.name, tt.expected[i].Content, res.Content)
				}

				if res.Start != tt.expected[i].Start {
					t.Errorf("Failed %s: expected start %d, got %d", tt.name, tt.expected[i].Start, res.Start)
				}

				if res.End != tt.expected[i].End {
					t.Errorf("Failed %s: expected end %d, got %d", tt.name, tt.expected[i].End, res.End)
				}

				if res.T != tt.expected[i].T {
					t.Errorf("Failed %s: expected type %s, got %s", tt.name, tt.expected[i].T, res.T)
				}

				if res.Index != tt.expected[i].Index {
					t.Errorf("Failed %s: expected index %d, got %d", tt.name, tt.expected[i].Index, res.Index)
				}
			}
		})
	}
}

func TestWriteCodeBlockToFile(t *testing.T) {
	t.Parallel()
	cb := CodeBlock{
		Content: "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}",
		T:       "go",
		Index:   1,
	}

	err := writeCodeBlockToFile(cb)
	if err != nil {
		t.Errorf("writeCodeBlockToFile failed: %v", err)
	}

	filename := "1.gno"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("file %s not created", filename)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Errorf("failed to read file %s: %v", filename, err)
	}

	expectedContent := cb.Content
	if string(content) != expectedContent {
		t.Errorf("file content mismatch\nexpected: %s\nactual: %s", expectedContent, string(content))
	}

	os.Remove(filename)
}

func normalize(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, "\n", ""), "\r", ""), "\t", ""), " ", "")
}
