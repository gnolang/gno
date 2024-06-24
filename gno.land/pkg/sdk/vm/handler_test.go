package vm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseQueryEvalData(t *testing.T) {
	t.Parallel()
	tt := []struct {
		input   string
		pkgpath string
		expr    string
	}{
		{
			"gno.land/r/realm.Expression()",
			"gno.land/r/realm",
			"Expression()",
		},
		{
			"a.b/c/d.e",
			"a.b/c/d",
			"e",
		},
		{
			"a.b.c.d.e/c/d.e",
			"a.b.c.d.e/c/d",
			"e",
		},
		{
			"abcde/c/d.e",
			"abcde/c/d",
			"e",
		},
	}
	for _, tc := range tt {
		path, expr := parseQueryEvalData(tc.input)
		assert.Equal(t, tc.pkgpath, path)
		assert.Equal(t, tc.expr, expr)
	}
}

func Test_parseQueryEval_panic(t *testing.T) {
	t.Parallel()

	assert.PanicsWithValue(t, panicInvalidQueryEvalData, func() {
		parseQueryEvalData("gno.land/r/demo/users")
	})
}
