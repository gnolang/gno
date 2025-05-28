package doctest

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHashCodeBlock(t *testing.T) {
	t.Parallel()
	codeBlock1 := codeBlock{
		content: `
package main

func main() {
	println("Hello, World")
}`,
		lang: "gno",
	}
	codeBlock2 := codeBlock{
		content: `
package main

func main() {
	println("Hello, World!")
}`,
		lang: "gno",
	}
	codeBlock3 := codeBlock{
		content: `
package main

func main() {
    println("Hello, World!")
}`,
		lang: "gno",
	}

	hashKey1 := hashCodeBlock(codeBlock1)
	hashKey2 := hashCodeBlock(codeBlock2)
	hashKey3 := hashCodeBlock(codeBlock3)

	assert.NotEqual(t, hashKey1, hashKey2)
	assert.NotEqual(t, hashKey2, hashKey3)
	assert.NotEqual(t, hashKey1, hashKey3)
}

func TestExecuteCodeBlock(t *testing.T) {
	tests := []struct {
		name           string
		codeBlock      codeBlock
		expectedResult string
		expectError    bool
	}{
		{
			name: "simple print without expected output",
			codeBlock: codeBlock{
				content: `
package main

func main() {
	println("Hello, World!")
}`,
				lang: "gno",
			},
			expectedResult: "Hello, World!\n",
		},
		{
			name: "print with expected output",
			codeBlock: codeBlock{
				content: `
package main

func main() {
	println("Hello, Gno!")
}
// Output:
// Hello, Gno!`,
				lang:           "gno",
				expectedOutput: "Hello, Gno!",
			},
			expectedResult: "Hello, Gno!\n",
		},
		{
			name: "print with incorrect expected output",
			codeBlock: codeBlock{
				content: `
package main

func main() {
	println("Hello, Gno!")
}
// Output:
// Hello, World!`,
				lang:           "gno",
				expectedOutput: "Hello, World!",
			},
			expectError: true,
		},
		{
			name: "code with expected error",
			codeBlock: codeBlock{
				content: `
package main

func main() {
	panic("oops")
}
// Error:
// panic: oops`,
				lang:          "gno",
				expectedError: "panic: oops",
			},
			expectError: true,
		},
		{
			name: "code with unexpected error",
			codeBlock: codeBlock{
				content: `
package main

func main() {
	panic("unexpected error")
}`,
				lang: "gno",
			},
			expectError: true,
		},
		{
			name: "Multiple print statements",
			codeBlock: codeBlock{
				content: `
package main

func main() {
	println("Line 1")
	println("Line 2")
}
// Output:
// Line 1
// Line 2`,
				lang:           "gno",
				expectedOutput: "Line 1\nLine 2",
			},
			expectedResult: "Line 1\nLine 2\n",
		},
		{
			name: "ignored code block",
			codeBlock: codeBlock{
				content: `
package main

func main() {
	panic("This should not execute")
}`,
				lang: "gno",
				options: ExecutionOptions{
					Ignore: true,
				},
			},
			expectedResult: "IGNORED",
		},
		{
			name: "should panic code block",
			codeBlock: codeBlock{
				content: `
package main

func main() {
	panic("Expected panic")
}`,
				lang: "gno",
				options: ExecutionOptions{
					PanicMessage: "Expected panic",
				},
			},
			expectedResult: "panicked as expected: Expected panic",
		},
		{
			name: "should panic but doesn't",
			codeBlock: codeBlock{
				content: `
package main

func main() {
	println("No panic")
}`,
				lang: "gno",
				options: ExecutionOptions{
					PanicMessage: "Expected panic",
				},
			},
			expectError: true,
		},
		{
			name: "should panic with specific message",
			codeBlock: codeBlock{
				content: `
package main

func main() {
	panic("Specific error message")
}`,
				lang: "gno",
				options: ExecutionOptions{
					PanicMessage: "Specific error message",
				},
			},
			expectedResult: "panicked as expected: Specific error message",
		},
		{
			name: "unsupported language 1",
			codeBlock: codeBlock{
				content: `
package main

func main() {
	println("Hello, World!")
}`,
				lang: "go",
			},
			expectedResult: "SKIPPED (Unsupported language: go)",
		},
		{
			name: "unsupported language 2",
			codeBlock: codeBlock{
				content: `print("Hello")`,
				lang:    "python",
			},
			expectedResult: "SKIPPED (Unsupported language: python)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExecuteCodeBlock(tt.codeBlock, GetStdlibsDir())

			if tt.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			assert.Equal(t, strings.TrimSpace(result), strings.TrimSpace(tt.expectedResult))
		})
	}
}

func TestCompareResults(t *testing.T) {
	tests := []struct {
		name           string
		actual         string
		expectedOutput string
		expectedError  string
		wantErr        bool
	}{
		{
			name:           "exact match",
			actual:         "Hello, World!",
			expectedOutput: "Hello, World!",
		},
		{
			name:           "mismatch",
			actual:         "Hello, World!",
			expectedOutput: "Hello, Gno!",
			wantErr:        true,
		},
		{
			name:           "regex match",
			actual:         "Hello, World!",
			expectedOutput: "regex:Hello, \\w+!",
		},
		{
			name:           "numbers regex match",
			actual:         "1234567890",
			expectedOutput: "regex:\\d+",
		},
		{
			name:           "complex regex match (e-mail format)",
			actual:         "foobar12456@somemail.com",
			expectedOutput: "regex:[a-zA-Z0-9]+@[a-zA-Z0-9]+\\.[a-zA-Z0-9]+",
		},
		{
			name:          "error match",
			actual:        "Error: division by zero",
			expectedError: "Error: division by zero",
		},
		{
			name:          "error mismatch",
			actual:        "Error: division by zero",
			expectedError: "Error: null pointer",
			wantErr:       true,
		},
		{
			name:          "error regex match",
			actual:        "Error: division by zero",
			expectedError: "regex:Error: .+",
		},
		{
			name:           "empty expected",
			actual:         "Hello, World!",
			expectedOutput: "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := compareResults(tt.actual, tt.expectedOutput, tt.expectedError)
			assert.Equal(t, tt.wantErr, err != nil)
			if err != nil && tt.wantErr {
				if tt.expectedOutput != "" {
					assert.Contains(t, err.Error(), tt.expectedOutput)
				}
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			}
		})
	}
}

func TestExecuteMatchingCodeBlock(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		pattern        string
		expectedResult []string
		expectError    bool
	}{
		{
			name: "single matching block",
			content: `
Some text here
` + "```gno" + `
// @test: test1
package main

func main() {
    println("Hello, World!")
}
` + "```" + `
More text
`,
			pattern:        "test1",
			expectedResult: []string{"\n=== test1 ===\n\nHello, World!\n\n"},
			expectError:    false,
		},
		{
			name: "multiple matching blocks",
			content: `
` + "```gno" + `
// @test: test1
package main

func main() {
    println("First")
}
` + "```" + `
` + "```gno" + `
// @test: test2
package main

func main() {
    println("Second")
}
` + "```" + `
`,
			pattern:        "test*",
			expectedResult: []string{"\n=== test1 ===\n\nFirst\n\n", "\n=== test2 ===\n\nSecond\n\n"},
			expectError:    false,
		},
		{
			name: "no matching blocks",
			content: `
` + "```gno" + `
// @test: test1
func main() {
    println("Hello")
}
` + "```" + `
`,
			pattern:        "nonexistent",
			expectedResult: []string{},
			expectError:    false,
		},
		{
			name: "error in code block",
			content: `
` + "```gno" + `
// @test: error_test
package main

func main() {
    panic("This should cause an error")
}
` + "```" + `
`,
			pattern:     "error_test",
			expectError: true,
		},
		{
			name: "expected output is nothing but actual output is something",
			content: `
` + "```gno" + `
// @test: foo
package main

func main() {
	println("This is an unexpected output")
}

// Output:
` + "```" + `
`,
			pattern:        "foo",
			expectedResult: []string{"\n=== foo ===\n\nThis is an unexpected output\n\n"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			results, err := ExecuteMatchingCodeBlock(ctx, tt.content, tt.pattern)

			if tt.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expectedResult, results)
			}

			for _, expected := range tt.expectedResult {
				found := false
				for _, result := range results {
					if strings.Contains(result, strings.TrimSpace(expected)) {
						found = true
						break
					}
				}
				assert.True(t, found)
			}
		})
	}
}

func TestShowingPropoerType(t *testing.T) {
	src := `
package main

type ints []int

func main() {
	a := ints{1, 2, 3}
	println(a)
}
`

	expected := "(slice[(1 int),(2 int),(3 int)] main.ints)\n"

	codeBlock := codeBlock{
		content: src,
		lang:    "gno",
	}

	result, err := ExecuteCodeBlock(codeBlock, GetStdlibsDir())
	assert.Nil(t, err)
	assert.Equal(t, expected, result)
}
