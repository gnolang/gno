package main

import (
	"testing"

	"github.com/gnolang/gno/contribs/gnodev/pkg/address"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNodeConfig_NoUnsafeRPC: DefaultNodeConfig starts from tm2's TestConfig,
// which turns the unsafe_* routes on. Neither mode may inherit that.
func TestNodeConfig_NoUnsafeRPC(t *testing.T) {
	for name, cfg := range map[string]AppConfig{
		"local":   defaultLocalAppConfig,
		"staging": defaultStagingOptions,
	} {
		t.Run(name, func(t *testing.T) {
			book := address.NewBook()
			book.Add(defaultDeployerAddress, DefaultDeployerName)

			nodeCfg, err := setupDevNodeConfig(&cfg, discardLogger(), nil, gnoland.NewBalances(), nil, book)
			require.NoError(t, err)
			assert.False(t, nodeCfg.TMConfig.RPC.Unsafe)
		})
	}
}
