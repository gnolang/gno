package doctest

import (
	"strings"
	"testing"
)

func clearCache() {
	cache.Lock()
	cache.m = make(map[string]string)
	cache.Unlock()
}

func TestExecuteCodeBlockWithCache(t *testing.T) {
	t.Parallel()
	clearCache()

	tests := []struct {
		name      string
		codeBlock codeBlock
		expect    string
	}{
		{
			name: "import go stdlib package",
			codeBlock: codeBlock{
				content: `
package main

func main() {
	println("Hello, World")
}`,
				lang: "gno",
			},
			expect: "Hello, World\n (cached)\n",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			stdlibDir := GetStdlibsDir()
			_, err := ExecuteCodeBlock(tt.codeBlock, stdlibDir)
			if err != nil {
				t.Errorf("%s returned an error: %v", tt.name, err)
			}

			cachedRes, err := ExecuteCodeBlock(tt.codeBlock, stdlibDir)
			if err != nil {
				t.Errorf("%s returned an error: %v", tt.name, err)
			}
			if cachedRes == tt.expect {
				t.Errorf("%s = %v, want %v", tt.name, cachedRes, tt.expect)
			}
		})
	}

	clearCache()
}

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
			expectError: true,
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
				options: ExecutionOption{
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
				options: ExecutionOption{
					ShouldPanic: "Expected panic",
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
				options: ExecutionOption{
					ShouldPanic: "Expected panic",
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
				options: ExecutionOption{
					ShouldPanic: "Specific error message",
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
			expectedError:  "",
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
