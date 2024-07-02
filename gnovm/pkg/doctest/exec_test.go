package doctest

import (
	"testing"
)

func TestExecuteCodeBlock(t *testing.T) {
	clearCache()
	tests := []struct {
		name      string
		codeBlock codeBlock
		expected  string
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
			expected: "Hello, World\n",
		},
		{
			name: "import go stdlib package",
			codeBlock: codeBlock{
				content: `
package main

import "std"

func main() {
	addr := std.GetOrigCaller()
	println(addr)
}`,
				lang: "gno",
			},
			expected: "g14ch5q26mhx3jk5cxl88t278nper264ces4m8nt\n",
		},
		{
			name: "import go stdlib package",
			codeBlock: codeBlock{
				content: `
package main

import "strings"

func main() {
	println(strings.ToUpper("Hello, World"))
}`,
				lang: "gno",
			},
			expected: "HELLO, WORLD\n",
		},
		{
			name: "print multiple values",
			codeBlock: codeBlock{
				content: `
package main

func main() {
	count := 3
	for i := 0; i < count; i++ {
		println("Hello")
	}
}`,
				lang: "gno",
			},
			expected: "Hello\nHello\nHello\n",
		},
		{
			name: "import subpackage without package declaration",
			codeBlock: codeBlock{
				content: `
func main() {
	println(math.Pi)
	println(strings.ToUpper("Hello, World"))
}`,
				lang: "gno",
			},
			expected: "3.141592653589793\nHELLO, WORLD\n",
		},
		{
			name: "test",
			codeBlock: codeBlock{
				content: "package main\n\nfunc main() {\nprintln(\"Hello, World!\")\n}",
				lang:    "gno",
			},
			expected: "Hello, World!\n",
		},
		{
			name: "missing package declaration",
			codeBlock: codeBlock{
				content: "func main() {\nprintln(\"Hello, World!\")\n}",
				lang:    "gno",
			},
			expected: "Hello, World!\n",
		},
		{
			name: "missing package and import declaration",
			codeBlock: codeBlock{
				content: `
func main() {
	s := strings.ToUpper("Hello, World")
	println(s)
}`,
				lang: "gno",
			},
			expected: "HELLO, WORLD\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := ExecuteCodeBlock(tt.codeBlock, STDLIBS_DIR)
			if err != nil {
				t.Errorf("%s returned an error: %v", tt.name, err)
			}

			if res != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, res, tt.expected)
			}
		})
	}
}

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
			_, err := ExecuteCodeBlock(tt.codeBlock, STDLIBS_DIR)
			if err != nil {
				t.Errorf("%s returned an error: %v", tt.name, err)
			}

			cachedRes, err := ExecuteCodeBlock(tt.codeBlock, STDLIBS_DIR)
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
