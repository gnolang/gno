package doctest

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"
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

	if hashKey1 == hashKey2 {
		t.Errorf("hash key for code block 1 and 2 are the same: %v", hashKey1)
	}
	if hashKey2 == hashKey3 {
		t.Errorf("hash key for code block 2 and 3 are the same: %v", hashKey2)
	}
	if hashKey1 == hashKey3 {
		t.Errorf("hash key for code block 1 and 3 are the same: %v", hashKey1)
	}
}

func TestExecuteCodeBlock(t *testing.T) {
	tests := []struct {
		name           string
		codeBlock      codeBlock
		expectedResult string
		expectError    bool
	}{
		{
			name: "Simple print without expected output",
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
			name: "Print with expected output",
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
			name: "Print with incorrect expected output",
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
			name: "Code with expected error",
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
			name: "Code with unexpected error",
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
			name: "Unsupported language",
			codeBlock: codeBlock{
				content: `print("Hello")`,
				lang:    "python",
			},
			expectedResult: "SKIPPED (Unsupported language: python)",
		},
		{
			name: "Ignored code block",
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
			name: "Should panic code block",
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
			name: "Should panic but doesn't",
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
			name: "Should panic with specific message",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExecuteCodeBlock(tt.codeBlock, GetStdlibsDir())

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if strings.TrimSpace(result) != strings.TrimSpace(tt.expectedResult) {
				t.Errorf("Expected result %q, but got %q", tt.expectedResult, result)
			}
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
			name:           "Exact match",
			actual:         "Hello, World!",
			expectedOutput: "Hello, World!",
		},
		{
			name:           "Mismatch",
			actual:         "Hello, World!",
			expectedOutput: "Hello, Gno!",
			wantErr:        true,
		},
		{
			name:           "Regex match",
			actual:         "Hello, World!",
			expectedOutput: "regex:Hello, \\w+!",
		},
		{
			name:           "Numbers Regex match",
			actual:         "1234567890",
			expectedOutput: "regex:\\d+",
		},
		{
			name:           "Complex Regex match (e-mail format)",
			actual:         "foobar12456@somemail.com",
			expectedOutput: "regex:[a-zA-Z0-9]+@[a-zA-Z0-9]+\\.[a-zA-Z0-9]+",
		},
		{
			name:          "Error match",
			actual:        "Error: division by zero",
			expectedError: "Error: division by zero",
		},
		{
			name:          "Error mismatch",
			actual:        "Error: division by zero",
			expectedError: "Error: null pointer",
			wantErr:       true,
		},
		{
			name:          "Error regex match",
			actual:        "Error: division by zero",
			expectedError: "regex:Error: .+",
		},
		{
			name:           "Empty expected",
			actual:         "Hello, World!",
			expectedOutput: "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := compareResults(tt.actual, tt.expectedOutput, tt.expectedError)
			if (err != nil) != tt.wantErr {
				t.Errorf("compareResults() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr {
				if tt.expectedOutput != "" && !strings.Contains(err.Error(), tt.expectedOutput) {
					t.Errorf("compareResults() error = %v, should contain %v", err, tt.expectedOutput)
				}
				if tt.expectedError != "" && !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("compareResults() error = %v, should contain %v", err, tt.expectedError)
				}
			}
		})
	}
}

func TestExecuteMatchingCodeBlock(t *testing.T) {
	testCases := []struct {
		name           string
		content        string
		pattern        string
		expectedResult []string
		expectError    bool
	}{
		{
			name: "Single matching block",
			content: `
Some text here
` + "```go" + `
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
			name: "Multiple matching blocks",
			content: `
` + "```go" + `
// @test: test1
package main

func main() {
    println("First")
}
` + "```" + `
` + "```go" + `
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
			name: "No matching blocks",
			content: `
` + "```go" + `
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
			name: "Error in code block",
			content: `
` + "```go" + `
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
` + "```go" + `
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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			results, err := ExecuteMatchingCodeBlock(ctx, tc.content, tc.pattern)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if len(results) == 0 && len(tc.expectedResult) == 0 {
					// do nothing
				} else if !reflect.DeepEqual(results, tc.expectedResult) {
					t.Errorf("Expected results %v, but got %v", tc.expectedResult, results)
				}
			}

			for _, expected := range tc.expectedResult {
				found := false
				for _, result := range results {
					if strings.Contains(result, strings.TrimSpace(expected)) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected result not found: %s", expected)
				}
			}
		})
	}
}
