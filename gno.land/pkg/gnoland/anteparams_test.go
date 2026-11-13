package gnoland

import (
	"testing"

	"github.com/stretchr/testify/require"

	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
)

func TestCheckTxParams(t *testing.T) {
	t.Parallel()

	// stored stands in for the params store; loads counts decodes.
	var stored int64
	loads := 0
	c := &checkTxParams{load: func(sdk.Context) anteParams {
		loads++
		var p anteParams
		p.auth.MaxMemoBytes = stored
		return p
	}}
	ctxAt := func(mode sdk.RunTxMode, height int64) sdk.Context {
		return sdk.NewContext(mode, nil, &bft.Header{ChainID: "test", Height: height}, log.NewNoopLogger())
	}
	get := func(mode sdk.RunTxMode, height int64) int64 {
		return c.get(ctxAt(mode, height)).auth.MaxMemoBytes
	}

	stored = 100
	require.EqualValues(t, 100, get(sdk.RunTxModeCheck, 5))
	require.EqualValues(t, 100, get(sdk.RunTxModeCheck, 5))
	require.EqualValues(t, 100, get(sdk.RunTxModeCheck, 5))
	require.Equal(t, 1, loads, "CheckTx at one height decodes once")

	// A tx in block 6 changes a param. DeliverTx always decodes, so the
	// rest of block 6 sees the new value at once.
	stored = 200
	require.EqualValues(t, 200, get(sdk.RunTxModeDeliver, 6))
	require.EqualValues(t, 200, get(sdk.RunTxModeDeliver, 6))
	require.Equal(t, 3, loads, "DeliverTx never uses the cache")

	// Simulate never uses it either.
	require.EqualValues(t, 200, get(sdk.RunTxModeSimulate, 5))
	require.Equal(t, 4, loads)

	// Block 6 is committed: the check state is now at height 6, so the
	// entry for 5 is not served.
	require.EqualValues(t, 200, get(sdk.RunTxModeCheck, 6))
	require.EqualValues(t, 200, get(sdk.RunTxModeCheck, 6))
	require.Equal(t, 5, loads, "a new height decodes again")

	// Genesis is never cached.
	require.EqualValues(t, 200, get(sdk.RunTxModeCheck, 0))
	require.EqualValues(t, 200, get(sdk.RunTxModeCheck, 0))
	require.Equal(t, 7, loads)
}
