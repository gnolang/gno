package gnoweb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrustedPathsWildcard(t *testing.T) {
	t.Parallel()

	all := newTrustedPaths([]string{"*"})
	for _, pkg := range []string{"gnoland/home", "nym/app", "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5/app"} {
		assert.True(t, all.contains(pkg), pkg)
	}
	assert.False(t, newTrustedPaths(nil).contains("nym/app"), "no entry trusts nothing")
}

func TestDefaultAppConfigTrustsDefaults(t *testing.T) {
	t.Parallel()

	trusted := newTrustedPaths(NewDefaultAppConfig().TrustedPaths)
	assert.True(t, trusted.contains("gnoland/home"))
	assert.False(t, trusted.contains("nym/app"))
}
