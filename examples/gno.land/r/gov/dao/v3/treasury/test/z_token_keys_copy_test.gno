package test

import (
	"testing"

	trs_pkg "gno.land/p/nt/treasury/v0"
	"gno.land/p/nt/uassert/v0"

	"gno.land/r/gov/dao/v3/treasury"
)

// Named z_* so it runs after the tests that assert exact token counts, since
// this one replaces the treasury's token key list.

// SetTokenKeys stores its argument into realm state that the grc20Lister
// closure re-reads on every Balances() and every GRC20 payment. Held
// uncopied, an allowed DAO realm would keep a write handle on the treasury's
// live token set for good — able to swap it on any later block, with no
// proposal and no vote behind it.
//
// The govdao_treasury_token_keys_are_immutable.txtar cannot cover this: the
// copy in NewTreasuryGRC20TokensUpdate means the slice arriving here is
// already impl's own, so the hazard is only reachable by calling SetTokenKeys
// directly, as an allowed DAO realm does.
func TestZSetTokenKeysCopiesItsArgument(cur realm, t *testing.T) {
	testing.SetRealm(allowedRealm)

	keyed := registerGRC20Tokens(0, cur, t, []string{"CopyTok"}, treasuryAddr)
	grc20ID := trs_pkg.GRC20Banker{}.ID()

	// The caller's own slice, handed over the way an allowed DAO realm's
	// executor hands over the keys a proposal carried.
	keys := []string{keyed[0].key}
	treasury.SetTokenKeys(cross(cur), keys)

	balances := treasury.Balances(grc20ID)
	uassert.Equal(t, len(balances), 1)
	uassert.Equal(t, balances[0].Denom, keyed[0].key)

	// No proposal, no vote: the supplying realm rewrites its own array.
	keys[0] = "gno.land/r/never/voted.EVIL"

	// The treasury must still resolve the key it was given.
	after := treasury.Balances(grc20ID)
	uassert.Equal(t, len(after), 1,
		"the treasury's token set followed the caller's rewrite")
	uassert.Equal(t, after[0].Denom, keyed[0].key)
}
