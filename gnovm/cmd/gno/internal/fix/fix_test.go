package fix

import (
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/ast/astutil"
)

func mustParse(src string) (*token.FileSet, *ast.File) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "main.go", src, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		panic(err)
	}
	return fset, f
}

func doFormat(fset *token.FileSet, f *ast.File) string {
	var buf strings.Builder
	format.Node(&buf, fset, f)
	return buf.String()
}

func Test_apply_FuncDeclPops(t *testing.T) {
	// Regression test for a bug whereby the FuncDecl wouldn't pop the scope
	// when exiting it, so declaring a variable with the same name as the
	// receiver or a parameter would panic
	const src = `
package main

type fooer struct {
	s string
}

func (f *fooer) Foo() {}

var f *fooer
`
	_, f := mustParse(src)
	apply(f, nil, nil)
}

func Test_apply_rename(t *testing.T) {
	tt := []struct {
		name     string
		src, res string
	}{
		{
			"base rename",
			`package main

func a(address string) {
	println(address)
}`,
			`package main

func a(address_ string) {
	println(address_)
}
`,
		},
		{
			"rename of global defined later",
			`package main

func a() {
	println(address)
}

var address = "123"
`,
			`package main

func a() {
	println(address_)
}

var address_ = "123"
`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			fset, f := mustParse(tc.src)
			apply(f, nil, func(c *astutil.Cursor, s scopes) bool {
				n := c.Node()
				if isBlockNode(n) {
					last := s[len(s)-1]
					if du := last["address"]; du != nil {
						du.rename("address_")
					}
				}
				return true
			})
			got := doFormat(fset, f)
			assert.Equal(t, tc.res, got)
		})
	}
}
