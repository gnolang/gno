package run

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunDataPkgAlias(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		pkgPath string
		want    string
	}{
		"unversioned realm":  {"gno.land/r/demo/boards", "boards"},
		"versioned realm v0": {"gno.land/r/gnoland/boards2/v0", "boards2"},
		"versioned pure v1":  {"gno.land/p/nt/avl/v1", "avl"},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, RunData{PkgPath: tc.pkgPath}.PkgAlias())
		})
	}
}
