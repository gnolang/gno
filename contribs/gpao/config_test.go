package main

import (
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// gnolandGasPrice is the block gas price gno.land runs, read off gnoland-1:
// `params/auth:p:initial_gasprice` answers {"gas":"1000","price":"1ugnot"}.
var gnolandGasPrice = std.GasPrice{Gas: 1000, Price: std.MustParseCoin("1ugnot")}

// TestDefaultFeeCoversDefaultGasWanted pins the one relationship between two
// constants that can be edited independently and silently break every approval.
//
// The fee is flat but the chain's requirement is a ratio:
// EnsureSufficientMempoolFees compares GasFee/GasWanted against the block gas
// price, so -gas-fee has to cover the gas an approval ASKS for. -gas-wanted is
// what it asks for whenever the node will not simulate one -- the degraded case,
// where a second failure is least welcome -- and a fee below the ratio is
// refused at CheckTx, which handleCandidate counts against the PACKAGE rather
// than against the configuration that caused it.
//
// Asserted through std.GasPrice.IsGTE, the same comparison the ante makes, so
// the rule has one implementation and this test cannot drift from it.
func TestDefaultFeeCoversDefaultGasWanted(t *testing.T) {
	fee, err := std.ParseCoin(defaultGasFee)
	require.NoError(t, err)

	offered := std.GasPrice{Gas: defaultGasWanted, Price: fee}
	ok, err := offered.IsGTE(gnolandGasPrice)
	require.NoError(t, err)
	assert.True(t, ok,
		"the default fee (%s) does not cover the default gas-wanted (%d) at %d gas per %s: "+
			"every approval that falls back to it would be refused at CheckTx",
		defaultGasFee, defaultGasWanted, gnolandGasPrice.Gas, gnolandGasPrice.Price)

	// The ceiling the fee buys, stated so a future edit has to move a number
	// rather than merely stay above a floor. Measured enables are an order of
	// magnitude below it; see defaultGasFee.
	const ceiling = int64(100_000_000)
	atCeiling := std.GasPrice{Gas: ceiling, Price: fee}
	ok, err = atCeiling.IsGTE(gnolandGasPrice)
	require.NoError(t, err)
	assert.True(t, ok, "the documented ceiling has to be reachable with the default fee")

	overCeiling := std.GasPrice{Gas: ceiling + 1, Price: fee}
	ok, err = overCeiling.IsGTE(gnolandGasPrice)
	require.NoError(t, err)
	assert.False(t, ok, "and one gas past it has to be the point the node refuses")
}

// TestConfigRejectsANonPositivePrepareBudget: the prepare budget bounds the
// child until it reports ready, and zero would kill every child at spawn while
// reporting the node as unavailable.
func TestConfigRejectsANonPositivePrepareBudget(t *testing.T) {
	valid := config{
		chainID:       "test",
		key:           "approver",
		gnoRoot:       "/gno",
		verifyBudget:  10 * time.Second,
		prepareBudget: time.Minute,
		gasWanted:     1,
	}
	require.NoError(t, valid.validate())

	for _, budget := range []time.Duration{0, -time.Second} {
		cfg := valid
		cfg.prepareBudget = budget
		err := cfg.validate()
		require.Error(t, err)
		require.ErrorContains(t, err, "prepare-budget must be positive")
	}
}
