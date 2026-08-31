package vm

import (
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// While the gas denom is restricted, freeing storage must not hand the deposit
// back to whoever released it — the whole point of the restriction is that
// locked tokens do not reach holders. The refund goes to the fee collector
// instead, and the unlock event says so via RefundWithheld, which is what the
// CLI reads to work out what a transaction actually cost.
func TestParamsDepositRefundIsWithheldWhileTheDenomIsRestricted(t *testing.T) {
	env, ctx, addr, pkgPath := setupParamsDepRealm(t)
	depAddr := gnolang.DeriveStorageDepositCryptoAddr(pkgPath)
	collector := env.vmk.GetParams(ctx).StorageFeeCollector
	require.False(t, collector.IsZero(), "the fee collector must be set for this to mean anything")
	require.NotEqual(t, addr, collector, "collector and releaser must differ, or the test proves nothing")

	// Lock a deposit while the denom is still unrestricted.
	callSet(t, env, ctx, addr, pkgPath, make([]byte, 100))
	lockedDep := env.bankk.GetCoins(ctx, depAddr).AmountOf(ugnot.Denom)

	// Now restrict the gas denom, then free the storage.
	env.bankk.SetRestrictedDenoms(ctx, []string{ugnot.Denom})
	callerBefore := env.bankk.GetCoin(ctx, addr, ugnot.Denom)
	collectorBefore := env.bankk.GetCoin(ctx, collector, ugnot.Denom)

	callDel(t, env, ctx, addr, pkgPath)

	released := lockedDep - env.bankk.GetCoins(ctx, depAddr).AmountOf(ugnot.Denom)
	require.Positive(t, released, "the delete must actually release a deposit")

	assert.Equal(t, callerBefore, env.bankk.GetCoin(ctx, addr, ugnot.Denom),
		"the releaser must not receive a withheld refund")
	assert.Equal(t, collectorBefore+released, env.bankk.GetCoin(ctx, collector, ugnot.Denom),
		"everything released must land with the fee collector")
}
