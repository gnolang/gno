package gnoweb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrustedPaths(t *testing.T) {
	t.Parallel()

	// Entries arrive unnormalized on purpose: the flag value is operator-typed.
	trusted := newTrustedPaths([]string{" gnoland ", "/nt/", "gnoswap/v1/pool", ""})

	cases := map[string]bool{
		"gnoland":             true,
		"gnoland/home":        true,
		"gnoland/home/":       true,
		"gnoland-evil/home":   false,
		"nt/avl":              true,
		"gnoswap/v1/pool":     true,
		"gnoswap/v1/pool/sub": true,
		"gnoswap/v1/router":   false,
		"gnoswap":             false,
		"":                    false,
	}
	for pkg, want := range cases {
		assert.Equal(t, want, trusted.contains(pkg), pkg)
	}
}

func TestTrustedPathsWildcard(t *testing.T) {
	t.Parallel()

	all := newTrustedPaths([]string{"*"})
	for _, pkg := range []string{"gnoland/home", "nym/app", "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5/app"} {
		assert.True(t, all.contains(pkg), pkg)
	}
	assert.False(t, newTrustedPaths(nil).contains("nym/app"), "no entry trusts nothing")
}

func TestMalformedTrustedPaths(t *testing.T) {
	t.Parallel()

	got := malformedTrustedPaths([]string{"gnoland", "/r/gnoland", "p/nt", "gno.land/r/x", "GnoLand", "nt/avl", "*", ""})
	assert.Equal(t, []string{"r/gnoland", "p/nt", "gno.land/r/x", "GnoLand"}, got)
}

func TestDefaultAppConfigTrustsDefaults(t *testing.T) {
	t.Parallel()

	trusted := newTrustedPaths(NewDefaultAppConfig().TrustedPaths)
	assert.True(t, trusted.contains("gnoland/home"))
	assert.False(t, trusted.contains("nym/app"))
}
