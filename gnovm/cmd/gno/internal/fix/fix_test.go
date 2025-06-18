package fix

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_apply(t *testing.T) {
	t.Run("FuncLitPops", func(t *testing.T) {
		// Regression test for a bug whereby the FuncLit wouldn't pop the scope
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
		f, err := parser.ParseFile(token.NewFileSet(), "main.go", src, parser.ParseComments|parser.SkipObjectResolution)
		require.NoError(t, err)
		apply(f, nil, nil)
	})
}
