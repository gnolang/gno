package coverage

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImportHandling(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, result string)
	}{
		{
			name: "gno_file_with_bare_import",
			input: `package avl

import

//----------------------------------------
// Node

// Node represents a node in an AVL tree.
type Node struct {
	key    string
	value  any
	height int8
}`,
			validate: func(t *testing.T, result string) {
				// The result should have a valid import block
				assert.Contains(t, result, "import", "Should contain import keyword")
				assert.Contains(t, result, "testing", "Should contain testing import")

				// Check that the import is properly formatted
				lines := strings.Split(result, "\n")
				importFound := false
				for i, line := range lines {
					if strings.TrimSpace(line) == "import" {
						// Next line should be the import spec or opening paren
						if i+1 < len(lines) {
							nextLine := strings.TrimSpace(lines[i+1])
							assert.True(t,
								nextLine == `"testing"` ||
									nextLine == "(" ||
									strings.HasPrefix(nextLine, `"`),
								"Import should be followed by valid syntax, got: %s", nextLine)
						}
						importFound = true
						break
					}
				}
				assert.True(t, importFound, "Import statement should be found")
			},
		},
		{
			name: "gno_file_with_gno_imports",
			input: `package int256

import (
	"errors"

	"gno.land/p/demo/uint256"
)

type Int struct {
	value uint256.Uint
}

func New() *Int {
	return &Int{}
}`,
			validate: func(t *testing.T, result string) {
				// Should preserve gno.land imports
				assert.Contains(t, result, "gno.land/p/demo/uint256", "Should preserve gno.land imports")
				assert.Contains(t, result, "testing", "Should add testing import")

				// Verify the AST is valid
				fset := token.NewFileSet()
				_, err := parser.ParseFile(fset, "test.go", result, parser.ParseComments)
				assert.NoError(t, err, "Result should be valid Go/Gno code")
			},
		},
		{
			name: "file_with_comment_between_package_and_import",
			input: `package demo

// This is a demo package
// with multiple comment lines

import "fmt"

func Demo() {
	fmt.Println("demo")
}`,
			validate: func(t *testing.T, result string) {
				// Should preserve comments
				assert.Contains(t, result, "This is a demo package", "Should preserve comments")
				assert.Contains(t, result, "testing", "Should add testing import")

				// Check that import structure is valid
				fset := token.NewFileSet()
				f, err := parser.ParseFile(fset, "test.go", result, parser.ParseComments)
				require.NoError(t, err)

				// Should have imports
				assert.NotEmpty(t, f.Imports, "Should have imports")

				// Check that testing is imported
				testingImported := false
				for _, imp := range f.Imports {
					if imp.Path.Value == `"testing"` {
						testingImported = true
						break
					}
				}
				assert.True(t, testingImported, "Testing should be imported")
			},
		},
		{
			name: "single_line_import_conversion",
			input: `package test

import "gno.land/p/demo/avl"

func UseAVL() {
	_ = avl.NewTree()
}`,
			validate: func(t *testing.T, result string) {
				// Should convert single import to multi-line when adding testing
				assert.Contains(t, result, "testing", "Should add testing import")

				// Parse and verify
				fset := token.NewFileSet()
				f, err := parser.ParseFile(fset, "test.go", result, parser.ParseComments)
				require.NoError(t, err)

				assert.Len(t, f.Imports, 2, "Should have two imports")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewTracker()
			engine := NewInstrumentationEngine(tracker, "test.gno")

			result, err := engine.InstrumentFile([]byte(tt.input))
			require.NoError(t, err)

			resultStr := string(result)
			t.Logf("Input:\n%s\n\nResult:\n%s\n", tt.input, resultStr)

			tt.validate(t, resultStr)
		})
	}
}

func TestImport(t *testing.T) {
	// This is the exact case from avl/node.gno that's causing issues
	input := `package avl

import

//----------------------------------------
// Node

// Node represents a node in an AVL tree.
type Node struct {
	key       string // key is the unique identifier for the node.
	value     any    // value is the data stored in the node.
	height    int8   // height is the height of the node in the tree.
}

func (node *Node) Key() string {
	return node.key
}`

	tracker := NewTracker()
	engine := NewInstrumentationEngine(tracker, "node.gno")

	result, err := engine.InstrumentFile([]byte(input))
	require.NoError(t, err)

	resultStr := string(result)
	t.Logf("Result:\n%s", resultStr)

	// The result should be valid Go code
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "node.gno", resultStr, parser.ParseComments)
	if err != nil {
		t.Errorf("Failed to parse result: %v", err)
		t.Logf("Problematic result:\n%s", resultStr)
	}
	require.NoError(t, err, "Result should be valid Go code")

	// Should have added testing import
	assert.NotEmpty(t, f.Imports, "Should have imports")

	hasTestingImport := false
	for _, imp := range f.Imports {
		if imp.Path.Value == `"testing"` {
			hasTestingImport = true
			break
		}
	}
	assert.True(t, hasTestingImport, "Should have testing import")
}
