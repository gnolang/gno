package gnoland

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	tu "github.com/gnolang/gno/tm2/pkg/sdk/testutils"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Tested here rather than in auth for the reason given on
// TestUnrestrictedAddrsParamGrantsAndRevokes: *GnoAccount carries the bit,
// std.BaseAccount does not.

// setFrozenAddrs makes the write governance makes. It goes through the params
// keeper rather than WillSetParam, so the old list is read back from the store
// before the new one replaces it, as on chain.
func (env testEnv) setFrozenAddrs(addrs ...crypto.Address) {
	strs := make([]string, len(addrs))
	for i, addr := range addrs {
		strs[i] = addr.String()
	}
	env.prmk.SetStrings(env.ctx, "auth:p:frozen_addrs", strs)
}

func TestFrozenAddrsParamFreezesAndUnfreezes(t *testing.T) {
	t.Parallel()
	env := setupTestEnv()

	existing := crypto.AddressFromPreimage([]byte("existing"))
	absent := crypto.AddressFromPreimage([]byte("absent"))
	env.acck.SetAccount(env.ctx, env.acck.NewAccountWithAddress(env.ctx, existing))

	frozen := func(t *testing.T, addr crypto.Address) bool {
		t.Helper()
		acc := env.acck.GetAccount(env.ctx, addr)
		require.NotNil(t, acc)
		return acc.(std.AccountFreezer).IsFrozen()
	}

	env.setFrozenAddrs(existing, absent)
	assert.True(t, frozen(t, existing))
	assert.True(t, frozen(t, absent),
		"an address with no account must get a frozen one, so coins it receives later cannot move")

	env.setFrozenAddrs(absent)
	assert.False(t, frozen(t, existing), "an address dropped from frozen_addrs must be unfrozen")
	assert.True(t, frozen(t, absent))

	assert.Panics(t, func() {
		env.prmk.SetStrings(env.ctx, "auth:p:frozen_addrs", []string{"not-an-address"})
	})
}

// A frozen account can sign nothing, whether with its own key or through a
// session created before the freeze, and is charged no fee for trying.
func TestFrozenAccountCannotSign(t *testing.T) {
	t.Parallel()
	env, ante, priv, _, addr := setupSessionGnoEnv(t)
	ctx := env.ctx

	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	sa := createGnoSession(t, env, addr, sessionPub, 0, []string{"*"})

	msgs := []std.Msg{tu.NewTestMsg(addr)}
	fee := std.NewFee(1_000_000, std.NewCoin("atom", 150))
	accNum := env.acck.GetAccount(ctx, addr).GetAccountNumber()
	direct := tu.NewTestTx(t, ctx.ChainID(), msgs, []crypto.PrivKey{priv}, []uint64{accNum}, []uint64{0}, fee)
	session := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sa.GetAccountNumber(), 0, fee)

	env.setFrozenAddrs(addr)
	balance := env.bankk.GetCoins(ctx, addr)
	for name, tx := range map[string]std.Tx{"direct": direct, "session": session} {
		_, res, abort := ante(ctx, tx, false)
		assert.True(t, abort, name)
		assert.Contains(t, res.Log, "is frozen", name)
	}
	assert.Equal(t, balance, env.bankk.GetCoins(ctx, addr), "a refused tx must not charge a fee")

	env.setFrozenAddrs()
	_, res, abort := ante(ctx, direct, false)
	assert.False(t, abort, "an unfrozen account signs again: %s", res.Log)
}
