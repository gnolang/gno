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
			expectError:    true,
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
			expectError:    true,
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
			expectError:    true,
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
			expectError:    true,
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
