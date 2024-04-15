package gnolang

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsRealmPath(t *testing.T) {
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
