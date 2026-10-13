package main

import (
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"

	"github.com/stretchr/testify/assert"
)

func TestGnoWebAppConfigTrustsEveryPathLocallyOnly(t *testing.T) {
	t.Parallel()

	local := defaultLocalAppConfig
	assert.Equal(t, []string{"*"}, gnoWebAppConfig(&local, "").TrustedPaths,
		"every package on a local chain is the developer's own")

	staging := defaultStagingOptions
	appcfg := gnoWebAppConfig(&staging, "")
	assert.Empty(t, appcfg.TrustedPaths, "a staging chain reserves no namespace")
	assert.Equal(t, gnoweb.IndexNoCommunity, appcfg.IndexCommunity)
}
