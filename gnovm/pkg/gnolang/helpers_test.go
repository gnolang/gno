package gnolang

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsRealmPath(t *testing.T) {
	t.Parallel()
	tt := []struct {
		input  string
		result bool
	}{
		{"gno.land/r/demo/users", true},
		{"gno.land/r/hello", true},
		{"gno.land/p/demo/users", false},
		{"gno.land/p/hello", false},
		{"gno.land/x", false},
		{"std", false},
	}

	for _, tc := range tt {
		assert.Equal(
			t,
			tc.result,
			IsRealmPath(tc.input),
			"unexpected IsRealmPath(%q) result", tc.input,
		)
	}
}

func TestIsStdlib(t *testing.T) {
	t.Parallel()

	tt := []struct {
		s      string
		result bool
	}{
		{"std", true},
		{"math", true},
		{"very/long/path/with_underscores", true},
		{"gno.land/r/demo/users", false},
		{"gno.land/hello", false},
	}

	for _, tc := range tt {
		assert.Equal(
			t,
			tc.result,
			IsStdlib(tc.s),
			"IsStdlib(%q)", tc.s,
		)
	}
}
