package fix

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStdsplit(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "simple std import removal",
			input: `package test
import "std"
func main() {}`,
			expected: `package test

func main() {}`,
		},
		{
			name: "std.Address rewrite",
			input: `package test
import "std"
func main() {
	addr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
}`,
			expected: `package test

import "chain"

func main() {
	addr := chain.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
}`,
		},
		{
			name: "std.AssertOriginCall rewrite",
			input: `package test
import "std"
func main() {
	std.AssertOriginCall()
}`,
			expected: `package test

import "chain/runtime"

func main() {
	runtime.AssertOriginCall()
}`,
		},
		{
			name: "multiple std functions",
			input: `package test
import "std"
func main() {
	addr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	std.AssertOriginCall()
}`,
			expected: `package test

import (
	"chain"
	"chain/runtime"
)

func main() {
	addr := chain.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	runtime.AssertOriginCall()
}`,
		},
		{
			name: "existing imports",
			input: `package test
import (
	"fmt"
	"std"
)
func main() {
	fmt.Println("hello")
	addr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
}`,
			expected: `package test

import (
	"chain"
	"fmt"
)

func main() {
	fmt.Println("hello")
	addr := chain.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
}`,
		},
		{
			name: "aliased import",
			input: `package test
import s "std"
func main() {
	addr := s.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
}`,
			expected: `package test

import "chain"

func main() {
	addr := chain.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "test.go", tc.input, parser.ParseComments)
			require.NoError(t, err)

			stdsplit(f)

			// Convert the AST back to source code for comparison
			output := astToString(t, fset, f)
			assert.Equal(t, tc.expected, output)
		})
	}
}

func TestCollisionHandling(t *testing.T) {
	// Test cases where there might be identifier collisions
	tests := []struct {
		name        string
		input       string
		shouldPanic bool
	}{
		{
			name: "collision with top-level identifier",
			input: `package test
import "std"
var chain int
func main() {
	addr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
}`,
			shouldPanic: true,
		},
		{
			name: "shadowing of runtime identifier",
			input: `package test
import "std"
func main() {
	runtime := 123
	std.AssertOriginCall()
}`,
			shouldPanic: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "test.go", tc.input, parser.ParseComments)
			require.NoError(t, err)

			if tc.shouldPanic {
				assert.Panics(t, func() {
					stdsplit(f)
				})
			} else {
				assert.NotPanics(t, func() {
					stdsplit(f)
				})
			}
		})
	}
}

func astToString(t *testing.T, fset *token.FileSet, f *ast.File) string {
	t.Helper()
	var buf bytes.Buffer
	err := format.Node(&buf, fset, f)
	require.NoError(t, err)
	return strings.TrimSuffix(buf.String(), "\n")
}
