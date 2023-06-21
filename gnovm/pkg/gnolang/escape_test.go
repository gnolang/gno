package gnolang

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type escapeTest struct {
	testName     string
	code         string
	declaration  func(*ast.File) *ast.FuncDecl
	expectedVars []string
}

var escapeTests = []escapeTest{
	{
		testName: "struct simple",
		code: `
		package p

		type Foo struct {
		  bar int
		}
		
		func main() {
			ff := &Foo{bar: 1}
			printit(ff)
		}`,
		declaration: func(f *ast.File) *ast.FuncDecl {
			return f.Decls[1].(*ast.FuncDecl)
		},
		expectedVars: []string{"ff"},
	},
	{
		testName: "variables simple",
		code: `
		package p
		func foo() {
			a := 5
			b := &a // both should escape
			c := b // should escape
			e := 5 // should not escape
		}`,
		declaration: func(f *ast.File) *ast.FuncDecl {
			return f.Decls[0].(*ast.FuncDecl)
		},
		expectedVars: []string{"a", "b", "c"},
	},
	{
		testName: "closures",
		code: `
		package p
		func foo() {
			a := new(int)
			*a = 5
			func(c int) *int {
				b := a // both should escape
				return b
			}
			e := 5 // should not escape
		}`,
		declaration: func(f *ast.File) *ast.FuncDecl {
			return f.Decls[0].(*ast.FuncDecl)
		},
		expectedVars: []string{"a", "b"},
	},
	{
		testName: "special built-in types",
		code: `
		package p
		func foo(x string) {
			var y string
			var z,zz map[int]int
			func(a, d string, b map[string]bool, c []string, e int, i interface{}) {
			}
		}`,
		declaration: func(f *ast.File) *ast.FuncDecl {
			return f.Decls[0].(*ast.FuncDecl)
		},
		expectedVars: []string{"a", "b", "c", "d", "x", "y", "z", "zz", "i"},
	},
	{
		testName: "goroutines",
		code: `
		package main

		func f() {
			i := new(int)
			*i = 1
			go func() {
				_ = i
			}()
			b := 4
			foo(&b)
		}
		`,
		declaration: func(f *ast.File) *ast.FuncDecl {
			var out *ast.FuncDecl
			for _, decl := range f.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "f" {
					out = fn
					break
				}
			}

			return out
		},
		expectedVars: []string{"i", "b"},
	},
}

func TestEscapeAnalysis(t *testing.T) {
	for _, et := range escapeTests {
		t.Run(et.testName, func(t *testing.T) {
			f, err := parser.ParseFile(token.NewFileSet(), "", et.code, 0)
			require.NoError(t, err)
			fn := et.declaration(f)
			ev := EscapeAnalysis(fn)
			assert.ElementsMatch(t, et.expectedVars, ev)
		})
	}
}
