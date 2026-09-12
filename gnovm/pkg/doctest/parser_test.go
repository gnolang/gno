package doctest

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
				{content: "fmt.Println(\"Hello, World!\")", lang: "go", index: 0},
			},
		},
		{
			name:  "Single code block with additional backticks",
			input: "```go\nfmt.Println(\"Hello, World!\")\n``````",
			expected: []codeBlock{
				{content: "fmt.Println(\"Hello, World!\")", lang: "go", index: 0},
			},
		},
		{
			name:  "Single code block with tildes",
			input: "## Example\nprint hello world in go.\n~~~go\nfmt.Println(\"Hello, World!\")\n~~~",
			expected: []codeBlock{
				{content: "fmt.Println(\"Hello, World!\")", lang: "go", index: 0},
			},
		},
		{
			name:  "Multiple code blocks",
			input: "Here is some text.\n```python\ndef hello():\n    print(\"Hello, World!\")\n```\nSome more text.\n```javascript\nconsole.log(\"Hello, World!\");\n```",
			expected: []codeBlock{
				{content: "def hello():\n    print(\"Hello, World!\")", lang: "python", index: 0},
				{content: "console.log(\"Hello, World!\");", lang: "javascript", index: 1},
			},
		},
		{
			name:  "Code block with no language specifier",
			input: "```\nfmt.Println(\"Hello, World!\")\n```",
			expected: []codeBlock{
				{content: "fmt.Println(\"Hello, World!\")", lang: "plain", index: 0},
			},
		},
		{
			name:     "No code blocks",
			input:    "Just some text without any code blocks.",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := GetCodeBlocks(tt.input)
			if err != nil {
				t.Errorf("failed %s: unexpected error %v", tt.name, err)
			}
			if len(result) != len(tt.expected) {
				t.Errorf("failed %s: expected %d code blocks, got %d", tt.name, len(tt.expected), len(result))
			}

			for i, res := range result {
				assert.Equal(t, normalize(tt.expected[i].content), normalize(res.content))
				assert.Equal(t, tt.expected[i].lang, res.lang)
				assert.Equal(t, tt.expected[i].index, res.index)
			}
		})
	}
}

func TestParseExpectedResults(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		wantOutput string
		wantError  string
	}{
		{
			name: "Basic output",
			content: `
// Some code
fmt.Println("Hello, World!")
// Output:
// Hello, World!
`,
			wantOutput: "Hello, World!",
		},
		{
			name: "Basic error",
			content: `
// Some code that causes an error
panic("oops")
// Error:
// panic: oops
`,
			wantError: "panic: oops",
		},
		{
			name: "Output and error",
			content: `
// Some code with both output and error
fmt.Println("Start")
panic("oops")
// Output:
// Start
// Error:
// panic: oops
`,
			wantOutput: "Start",
			wantError:  "panic: oops",
		},
		{
			name: "Multiple lines in section",
			content: `
fmt.Println("Hello")
// Output:
// Hello
// World
`,
			wantOutput: "Hello\nWorld",
		},
		{
			name: "Preserve indentation",
			content: `
fmt.Println("  Indented")
// Output:
//   Indented
`,
			wantOutput: "  Indented",
		},
		{
			name: "Output with // in content",
			content: `
fmt.Println("// Comment")
// Output:
// // Comment
`,
			wantOutput: "// Comment",
		},
		{
			name: "No directive",
			content: `
// Just some comments
// No output or error
`,
		},
		{
			name: "Bare // ends section",
			content: `
// Output:
// Hello
//
// not part of output
`,
			wantOutput: "Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := parseBlockMetadata(tt.content)
			assert.Equal(t, tt.wantOutput, meta.output)
			assert.Equal(t, tt.wantError, meta.errOutput)
		})
	}
}

func TestParseBlockMetadataName(t *testing.T) {
	tests := []struct {
		name    string
		content string
		index   int
		want    string
	}{
		{
			name:    "explicit name via NAME directive",
			content: "// NAME: my_test\npackage main",
			want:    "my_test",
		},
		{
			name:    "NAME with surrounding whitespace",
			content: "package main\n    // NAME:    trimmed   \nfunc main() {}",
			want:    "trimmed",
		},
		{
			name:    "no directive falls back to block_<index>",
			content: "package main\nfunc main() {}",
			index:   2,
			want:    "block_2",
		},
		{
			name:    "empty NAME falls back",
			content: "// NAME:\npackage main",
			index:   1,
			want:    "block_1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := parseBlockMetadata(tt.content)
			got := meta.name
			if got == "" {
				got = fmt.Sprintf("block_%d", tt.index)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseExecutionOptions(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    ExecutionOptions
	}{
		{
			name:    "No options",
			content: "package main",
			want:    ExecutionOptions{},
		},
		{
			name:    "IGNORE directive",
			content: "// IGNORE:\npackage main",
			want:    ExecutionOptions{Ignore: true},
		},
		{
			name:    "SHOULD_PANIC without message",
			content: "// SHOULD_PANIC:\npackage main",
			want:    ExecutionOptions{ShouldPanic: true},
		},
		{
			name:    "SHOULD_PANIC with message",
			content: "// SHOULD_PANIC: division by zero\npackage main",
			want:    ExecutionOptions{ShouldPanic: true, PanicMessage: "division by zero"},
		},
		{
			name:    "Multiple directives",
			content: "// IGNORE:\n// SHOULD_PANIC: runtime error\npackage main",
			want:    ExecutionOptions{Ignore: true, ShouldPanic: true, PanicMessage: "runtime error"},
		},
		{
			name:    "Directive after NAME",
			content: "// NAME: my_test\n// SHOULD_PANIC: out of bounds\n",
			want:    ExecutionOptions{ShouldPanic: true, PanicMessage: "out of bounds"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := parseBlockMetadata(tt.content)
			assert.Equal(t, tt.want, meta.options)
		})
	}
}

func TestGetCodeBlocksWithOptions(t *testing.T) {
	input := `
` + "```gno" + `
// IGNORE:
package main

func main() {
    panic("This should not execute")
}
` + "```" + `

` + "```gno" + `
// SHOULD_PANIC: runtime error: index out of range
package main

func main() {
    arr := []int{1, 2, 3}
    fmt.Println(arr[5])
}
` + "```" + `

` + "```gno" + `
package main

func main() {
    fmt.Println("Hello, World!")
}
` + "```" + `
`

	blocks, err := GetCodeBlocks(input)
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, blocks, 3)
	assert.True(t, blocks[0].options.Ignore)
	assert.True(t, blocks[1].options.ShouldPanic)
	assert.Equal(t, "runtime error: index out of range", blocks[1].options.PanicMessage)
	assert.Equal(t, ExecutionOptions{}, blocks[2].options)
}

// ignore whitespace in the source code
func normalize(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, "\n", ""), "\r", ""), "\t", ""), " ", "")
}
