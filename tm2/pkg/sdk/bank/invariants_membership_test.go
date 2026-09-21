package bank

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
)

// AllInvariants has to actually run every member. Dropping one from its list
// compiles, leaves that member's own tests passing, and quietly stops the state
// from being checked. It matters most for the supply check, the only member that
// catches a mint which bypassed the counter -- the others all find such a mint
// perfectly well formed.
//
// Each case poisons state that exactly one member reports, and asserts the other
// two stay quiet. That is what makes it a membership test: with the uniqueness
// established, AllInvariants can only be reporting through the member named.
func TestAllInvariantsRunsEveryMember(t *testing.T) {
	t.Parallel()

	cases := []struct {
		member string
		want   string
		poison func(t *testing.T, env testEnv)
	}{
		{
			"balance-keys", "no account object",
			func(t *testing.T, env testEnv) {
				t.Helper()
				// A balance held by an address with no account object, so nothing
				// can sign for it. Supply is recomputed afterwards so that the
				// missing account is the only thing left to report.
				stray := crypto.AddressFromPreimage([]byte("no-account-holder"))
				rawSet(t, env, BalanceKey(stray, "atom"), encodeBalance(5))
				env.bankk.RecomputeSupply(env.ctx)
			},
		},
		{
			"account-tier", "must never be in the account tier",
			func(t *testing.T, env testEnv) {
				t.Helper()
				// A realm denom inside an account object, where the keeper routes
				// it to the split tier and would never put one.
				addr := crypto.AddressFromPreimage([]byte("stranded-holder"))
				acc := env.acck.NewAccountWithAddress(env.ctx, addr)
				require.NoError(t, acc.SetCoins(std.Coins{{Denom: testRealmDenom, Amount: 3}}))
				env.acck.SetAccount(env.ctx, acc)
				env.bankk.RecomputeSupply(env.ctx)
			},
		},
		{
			"total-supply", "no supply record",
			func(t *testing.T, env testEnv) {
				t.Helper()
				// SetCoins is supply-blind, so these balances exist with nothing
				// minted behind them -- the shape a counter-bypassing mint leaves.
				fundHealthy(t, env, crypto.AddressFromPreimage([]byte("unrecorded")))
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.member, func(t *testing.T) {
			t.Parallel()

			env := setupTestEnv()
			tt.poison(t, env)

			members := map[string]sdk.Invariant{
				"balance-keys": BalanceKeysInvariant(env.bankk.ViewKeeper),
				"account-tier": AccountTierInvariant(env.bankk.ViewKeeper),
				"total-supply": SupplyInvariant(env.bankk.ViewKeeper),
			}
			for name, inv := range members {
				msg, broken := inv(env.ctx)
				if name == tt.member {
					require.True(t, broken, "%s should report this state", name)
					require.Contains(t, msg, tt.want)
					continue
				}
				require.False(t, broken,
					"%s also reports this state, so the case does not pin %s:\n%s",
					name, tt.member, msg)
			}

			msg, broken := AllInvariants(env.bankk.ViewKeeper)(env.ctx)
			require.True(t, broken, "AllInvariants does not run %s", tt.member)
			require.Contains(t, msg, tt.want)
		})
	}
}
