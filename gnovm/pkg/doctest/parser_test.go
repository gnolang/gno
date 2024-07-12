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
			wantOutput: "  Indented",
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
		name     string
		content  string
		expected string
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
			expected: "Print_main_Hello, World!",
		},
		{
			name: "Explicitly named code block",
			content: `
// @test: specified
package main

func main() {
	println("specified")
}`,
			expected: "specified",
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
			expected: "Calc_math_calculateArea_78.53981633974483",
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
			expected: "Test_testing_TestSquareRoot",
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
			expected: "Print_math_strings_main_3.141592653589793",
		},
		{
			name: "No function",
			content: `
package main

var x = 5
`,
			expected: "x",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateCodeBlockName(tt.content)
			if result != tt.expected {
				t.Errorf("generateCodeBlockName() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseExecutionOptions(t *testing.T) {
	tests := []struct {
		name      string
		language  string
		firstLine string
		want      ExecutionOption
	}{
		{
			name:      "No options",
			language:  "go",
			firstLine: "package main",
			want:      ExecutionOption{},
		},
		{
			name:      "Ignore option in language tag",
			language:  "go,ignore",
			firstLine: "package main",
			want:      ExecutionOption{Ignore: true},
		},
		{
			name:      "Should panic option in language tag",
			language:  "go,should_panic",
			firstLine: "package main",
			want:      ExecutionOption{ShouldPanic: ""},
		},
		{
			name:      "Should panic with message in comment",
			language:  "go,should_panic",
			firstLine: "// @should_panic=\"division by zero\"",
			want:      ExecutionOption{ShouldPanic: "division by zero"},
		},
		{
			name:      "Multiple options",
			language:  "go,ignore,should_panic",
			firstLine: "// @should_panic=\"runtime error\"",
			want:      ExecutionOption{Ignore: true, ShouldPanic: "runtime error"},
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

// ignore whitespace in the source code
func normalize(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, "\n", ""), "\r", ""), "\t", ""), " ", "")
}
