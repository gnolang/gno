package run

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
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

func renderPage(t *testing.T, data RunData) string {
	t.Helper()
	var buf strings.Builder
	require.NoError(t, NewPageView(data).Render(&buf))
	return buf.String()
}

func TestPageGasDefaults(t *testing.T) {
	t.Parallel()

	page := renderPage(t, RunData{PkgPath: "gno.land/r/demo/counter", Gas: components.RunGas})
	assert.Contains(t, page, `value="`+components.RunGas.Wanted+`"`)
	assert.Contains(t, page, `value="`+components.RunGas.Fee+`"`)
	assert.NotEqual(t, components.CallGas, components.RunGas, "a script asks for less than $help's calls")
}
