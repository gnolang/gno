package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGnoWebAppConfigTrustsEveryPathLocallyOnly(t *testing.T) {
	t.Parallel()

	local := defaultLocalAppConfig
	assert.Equal(t, []string{"*"}, gnoWebAppConfig(&local, "").TrustedPaths,
		"every package on a local chain is the developer's own")

	staging := defaultStagingOptions
	assert.NotContains(t, gnoWebAppConfig(&staging, "").TrustedPaths, "*",
		"a staging server hosts other deployers' realms")
}
