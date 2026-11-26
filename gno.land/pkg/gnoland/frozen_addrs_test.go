package gnoland

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
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
	env.prmk.SetStrings(env.ctx, "auth:p:frozen_addrs", addrStrs(addrs))
}

func addrStrs(addrs []crypto.Address) []string {
	strs := make([]string, len(addrs))
	for i, addr := range addrs {
		strs[i] = addr.String()
	}
	return strs
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

// The veto sees only the addresses a write adds, in list order, and refuses
// before any of them is frozen or given an account. Removals never reach it.
func TestFrozenAddrsVeto(t *testing.T) {
	t.Parallel()
	env := setupTestEnv()

	listed := crypto.AddressFromPreimage([]byte("listed"))
	a := crypto.AddressFromPreimage([]byte("a"))
	b := crypto.AddressFromPreimage([]byte("b"))
	env.setFrozenAddrs(listed)

	var vetoed [][]crypto.Address
	acck := env.acck.WithFreezeVeto(func(_ sdk.Context, addrs []crypto.Address) {
		vetoed = append(vetoed, addrs)
		panic("refused")
	})
	write := func(addrs ...crypto.Address) {
		acck.WillSetParam(env.ctx, "p:frozen_addrs", addrStrs(addrs))
	}

	assert.PanicsWithValue(t, "refused", func() { write(b, listed, a) })
	assert.Equal(t, [][]crypto.Address{{b, a}}, vetoed)
	assert.Nil(t, env.acck.GetAccount(env.ctx, a), "a refused write must not create accounts")
	assert.Nil(t, env.acck.GetAccount(env.ctx, b))

	assert.NotPanics(t, func() { write() }, "an unfreeze must not reach the veto")
	assert.Len(t, vetoed, 1)
	assert.False(t, env.acck.GetAccount(env.ctx, listed).(std.AccountFreezer).IsFrozen())
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

// Freezing a session key's own address stops that key signing for its master,
// though the master stays unfrozen: the list names keys, and a session key
// signs as its master.
func TestFrozenSessionKeyCannotSign(t *testing.T) {
	t.Parallel()
	env, ante, _, _, addr := setupSessionGnoEnv(t)
	ctx := env.ctx

	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	sa := createGnoSession(t, env, addr, sessionPub, 0, []string{"*"})
	msgs := []std.Msg{tu.NewTestMsg(addr)}
	fee := std.NewFee(1_000_000, std.NewCoin("atom", 150))
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sa.GetAccountNumber(), 0, fee)

	env.setFrozenAddrs(sessionAddr)
	balance := env.bankk.GetCoins(ctx, addr)
	_, res, abort := ante(ctx, tx, false)
	assert.True(t, abort)
	assert.Contains(t, res.Log, "is frozen")
	assert.Equal(t, balance, env.bankk.GetCoins(ctx, addr), "a refused tx must not charge the master a fee")

	env.setFrozenAddrs()
	_, res, abort = ante(ctx, tx, false)
	assert.False(t, abort, "an unfrozen session key signs again: %s", res.Log)
}
