package test

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
)

// The harness must carry the call-credit ledger, seeded with the send, or
// PayCall/CallSend behave differently under `gno test` than on chain.
func TestContextSeedsCallCredits(t *testing.T) {
	send := std.Coins{{Denom: "ugnot", Amount: 42}}
	ctx := Context("", "gno.land/r/demo/pay", send)
	require.NotNil(t, ctx.CallCredits)
	require.Equal(t, send, ctx.CallCredits.Take("", "gno.land/r/demo/pay"))
	require.Empty(t, ctx.CallCredits.Take("", "gno.land/r/demo/pay"))
}
