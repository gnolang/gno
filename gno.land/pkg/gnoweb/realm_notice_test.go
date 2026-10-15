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

func TestTrustedPathProblem(t *testing.T) {
	t.Parallel()

	cases := map[string]bool{ // entry → has a problem
		"gnoland":             false,
		" /gnoswap/v1/pool/ ": false,
		"nym-sunny000":        false,
		"gno.land/r/gnoland":  true,
		"r/gnoland":           true,
		"/p/nt/":              true,
		"u/alice":             true,
		"GnoLand":             true,
		"*":                   false,
		"gnoland/*":           true,
		"*/home":              true,
	}
	for entry, want := range cases {
		assert.Equal(t, want, trustedPathProblem(entry) != "", entry)
	}
}
