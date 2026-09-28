package bank

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
)

// computeSupply totals the account tier by walking account entries and skipping
// every kind but regular, on the grounds that only a regular entry can hold
// coins. A session object that held any would be invisible to the counter, so
// this funds a session ADDRESS and checks the money lands somewhere the sweep
// can see.
func TestSupplyCountsASessionAddressExactlyOnce(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx, _, da := setupSessionCtx(t, env,
		std.NewCoins(std.NewCoin("foo", 1000)),
		std.NewCoins(std.NewCoin("foo", 500)))
	sessionAddr := da.GetAddress()

	// The master's coins were seeded with SetCoins, which keeps no counter.
	env.bankk.RecomputeSupply(ctx)
	require.Equal(t, int64(1000), env.bankk.TotalSupply(ctx, "foo"))

	// Crediting the session's own address creates a regular account for it,
	// separate from the session entry filed under the master.
	require.NoError(t, env.bankk.MintCoins(ctx, sessionAddr, std.NewCoins(std.NewCoin("foo", 250))))
	require.Equal(t, int64(1250), env.bankk.TotalSupply(ctx, "foo"))
	require.Equal(t, int64(250), env.bankk.GetCoin(ctx, sessionAddr, "foo"))

	// The session object holds nothing, which is what makes the skip safe.
	var sessionCoins std.Coins
	require.NoError(t, env.acck.IterateAccountEntries(ctx, func(e auth.AccountEntry) bool {
		if e.Kind == auth.AccountKeySession && e.Account != nil {
			sessionCoins = e.Account.GetCoins()
		}
		return false
	}))
	require.True(t, sessionCoins.IsZero(),
		"a session object must never hold coins, or computeSupply would miss them")

	// The sweep must reach the same total as the incremental counter.
	env.bankk.RecomputeSupply(ctx)
	require.Equal(t, int64(1250), env.bankk.TotalSupply(ctx, "foo"))

	msg, broken := auth.AllInvariants(env.acck)(ctx)
	require.False(t, broken, "auth invariant broken: %s", msg)
	msg, broken = AllInvariants(env.bankk.ViewKeeper)(ctx)
	require.False(t, broken, "bank invariant broken: %s", msg)
}
