package doctest

import (
	"reflect"
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

func TestParseExpectedResults(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		wantOutput     string
		wantError      string
		wantParseError bool
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
			wantError:  "",
		},
		{
			name: "Basic error",
			content: `
// Some code that causes an error
panic("oops")
// Error:
// panic: oops
`,
			wantOutput: "",
			wantError:  "panic: oops",
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
			name: "Multiple output sections",
			content: `
// First output
fmt.Println("Hello")
// Output:
// Hello
// World
`,
			wantOutput: "Hello\nWorld",
			wantError:  "",
		},
		{
			name: "Preserve indentation",
			content: `
// Indented output
fmt.Println("  Indented")
// Output:
//   Indented
`,
			wantOutput: "Indented",
			wantError:  "",
		},
		{
			name: "Output with // in content",
			content: `
// Output with //
fmt.Println("// Comment")
// Output:
// // Comment
`,
			wantOutput: "// Comment",
			wantError:  "",
		},
		{
			name: "Empty content",
			content: `
// Just some comments
// No output or error
`,
			wantOutput: "",
			wantError:  "",
		},
		{
			name: "simple code",
			content: `
package main

func main() {
	println("Actual output")
}
// Output:
// Actual output
`,
			wantOutput: "Actual output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOutput, gotError, err := parseExpectedResults(tt.content)
			if (err != nil) != tt.wantParseError {
				t.Errorf("parseExpectedResults() error = %v, wantParseError %v", err, tt.wantParseError)
				return
			}
			if gotOutput != tt.wantOutput {
				t.Errorf("parseExpectedResults() gotOutput = %v, want %v", gotOutput, tt.wantOutput)
			}
			if gotError != tt.wantError {
				t.Errorf("parseExpectedResults() gotError = %v, want %v", gotError, tt.wantError)
			}
		})
	}
}

func TestGenerateCodeBlockName(t *testing.T) {
	tests := []struct {
		name                 string
		content              string
		output               string
		expectedGenerateName string
	}{
		{
			name: "Simple print function",
			content: `
package main

func main() {
    println("Hello, World!")
}
// Output:
// Hello, World!
`,
			output:               "Hello, World!",
			expectedGenerateName: "Print_main_Hello, World!",
		},
		{
			name: "Explicitly named code block",
			content: `
// @test: specified
package main

func main() {
	println("specified")
}`,
			output:               "specified",
			expectedGenerateName: "specified",
		},
		{
			name: "Simple calculation",
			content: `
package main

import "math"

func calculateArea(radius float64) float64 {
    return math.Pi * radius * radius
}

func main() {
    println(calculateArea(5))
}
// Output:
// 78.53981633974483
`,
			output:               "78.53981633974483",
			expectedGenerateName: "Calc_calculateArea_78.53981633974483_math",
		},
		{
			name: "Test function",
			content: `
package main

import "testing"

func TestSquareRoot(t *testing.T) {
    got := math.Sqrt(4)
    if got != 2 {
        t.Errorf("Sqrt(4) = %f; want 2", got)
    }
}
`,
			expectedGenerateName: "Test_TestSquareRoot_testing",
		},
		{
			name: "Multiple imports",
			content: `
package main

import (
    "math"
    "strings"
)

func main() {
    println(math.Pi)
    println(strings.ToUpper("hello"))
}
// Output:
// 3.141592653589793
// HELLO
`,
			output:               "3.141592653589793\nHELLO",
			expectedGenerateName: "Print_main_3.141592653589793_math_strings",
		},
		{
			name: "No function",
			content: `
package main

var x = 5
`,
			expectedGenerateName: "x",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateCodeBlockName(tt.content, tt.output)
			if result != tt.expectedGenerateName {
				t.Errorf("generateCodeBlockName() = %v, want %v", result, tt.expectedGenerateName)
			}
		})
	}
}

func TestParseExecutionOptions(t *testing.T) {
	tests := []struct {
		name      string
		language  string
		firstLine string
		want      ExecutionOptions
	}{
		{
			name:      "No options",
			language:  "go",
			firstLine: "package main",
			want:      ExecutionOptions{},
		},
		{
			name:      "Ignore option in language tag",
			language:  "go,ignore",
			firstLine: "package main",
			want:      ExecutionOptions{Ignore: true},
		},
		{
			name:      "Should panic option in language tag",
			language:  "go,should_panic",
			firstLine: "package main",
			want:      ExecutionOptions{PanicMessage: ""},
		},
		{
			name:      "Should panic with message in comment",
			language:  "go,should_panic",
			firstLine: "// @should_panic=\"division by zero\"",
			want:      ExecutionOptions{PanicMessage: "division by zero"},
		},
		{
			name:      "Multiple options",
			language:  "go,ignore,should_panic",
			firstLine: "// @should_panic=\"runtime error\"",
			want:      ExecutionOptions{Ignore: true, PanicMessage: "runtime error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseExecutionOptions(tt.language, []byte(tt.firstLine))
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseExecutionOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCodeBlocksWithOptions(t *testing.T) {
	input := `
Some text here

` + "```go,ignore" + `
// This block should be ignored
func main() {
    panic("This should not execute")
}
` + "```" + `

Another paragraph

` + "```go,should_panic" + `
// @should_panic="runtime error: index out of range"
func main() {
    arr := []int{1, 2, 3}
    fmt.Println(arr[5])
}
` + "```" + `

` + "```go" + `
// Normal execution
func main() {
    fmt.Println("Hello, World!")
}
` + "```" + `
`

	blocks := GetCodeBlocks(input)

	if len(blocks) != 3 {
		t.Fatalf("Expected 3 code blocks, got %d", len(blocks))
	}

	// Check the first block (ignore)
	if !blocks[0].options.Ignore {
		t.Errorf("Expected first block to be ignored")
	}

	// Check the second block (should_panic)
	if blocks[1].options.PanicMessage != "runtime error: index out of range" {
		t.Errorf("Expected second block to have ShouldPanic option set to 'runtime error: index out of range', got '%s'", blocks[1].options.PanicMessage)
	}

	// Check the third block (normal execution)
	if blocks[2].options.Ignore || blocks[2].options.PanicMessage != "" {
		t.Errorf("Expected third block to have no special options")
	}
}

// ignore whitespace in the source code
func normalize(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, "\n", ""), "\r", ""), "\t", ""), " ", "")
}
