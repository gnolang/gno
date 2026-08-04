package bank

import (
	"math"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/stretchr/testify/require"
)

func TestSupplyKeyFormat(t *testing.T) {
	t.Parallel()

	require.Equal(t, "/supply/", SupplyPrefix)
	require.Equal(t, []byte("/supply/atom"), SupplyKey("atom"))

	denom, err := denomFromSupplyKey(SupplyKey(testRealmDenom))
	require.NoError(t, err)
	require.Equal(t, testRealmDenom, denom)

	_, err = denomFromSupplyKey([]byte("/supply/"))
	require.Error(t, err, "a key with no denom must be rejected")
	_, err = denomFromSupplyKey([]byte("/b/atom"))
	require.Error(t, err, "a key outside the prefix must be rejected")
}

// The supply keyspace must not overlap any other keyspace in the shared store, in
// either direction — a prefix sweep of one must never see a key of another.
func TestSupplyPrefixDoesNotCollide(t *testing.T) {
	t.Parallel()

	addr := crypto.AddressFromPreimage([]byte("a"))
	neighbours := [][]byte{
		[]byte("/a/" + string(addr[:])),         // auth accounts
		[]byte("/a/" + string(addr[:]) + "/s/"), // auth sessions
		BalanceKey(addr, "atom"),                // bank balances
		[]byte("/pv/bank:x"),                    // params
		[]byte("gasPrice"),
		[]byte("globalAccountNumber"),
		[]byte("consensus_params"),
		[]byte("pkg:gno.land/r/x"), // GnoVM mempackages
		[]byte("last_header"),
	}
	sup := SupplyKey(testRealmDenom)
	for _, n := range neighbours {
		require.False(t, hasPrefixBytes(sup, n), "%q must not be prefixed by %q", sup, n)
		require.False(t, hasPrefixBytes(n, []byte(SupplyPrefix)),
			"%q must not fall under %q", n, SupplyPrefix)
	}
	// And the range really is disjoint in a live store.
	env := setupTestEnv()
	stor := env.ctx.Store(env.key)
	for _, n := range neighbours {
		stor.Set(nil, n, []byte{1})
	}
	env.bankk.setSupply(env.ctx, testRealmDenom, 5)

	iter := store.PrefixIterator(nil, stor, []byte(SupplyPrefix))
	defer iter.Close()
	var seen int
	for ; iter.Valid(); iter.Next() {
		seen++
		_, err := denomFromSupplyKey(iter.Key())
		require.NoError(t, err, "a supply sweep saw a foreign key: %X", iter.Key())
	}
	require.Equal(t, 1, seen)
}

func hasPrefixBytes(b, prefix []byte) bool {
	return len(b) >= len(prefix) && string(b[:len(prefix)]) == string(prefix)
}

func TestMintAndBurnMoveSupplyButTransfersDoNot(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx
	a := crypto.AddressFromPreimage([]byte("a"))
	b := crypto.AddressFromPreimage([]byte("b"))

	for _, denom := range []string{testAccountDenom, testRealmDenom} {
		require.Zero(t, env.bankk.TotalSupply(ctx, denom))
		require.NoError(t, env.bankk.MintCoins(ctx, a, std.Coins{{Denom: denom, Amount: 100}}))
		require.Equal(t, int64(100), env.bankk.TotalSupply(ctx, denom))

		// Transfers are supply-neutral, on every entry point.
		require.NoError(t, env.bankk.SendCoins(ctx, a, b, std.Coins{{Denom: denom, Amount: 10}}))
		require.NoError(t, env.bankk.SendCoinsUnrestricted(ctx, a, b, std.Coins{{Denom: denom, Amount: 10}}))
		require.NoError(t, env.bankk.InputOutputCoins(ctx,
			[]Input{{Address: a, Coins: std.Coins{{Denom: denom, Amount: 5}}}},
			[]Output{{Address: b, Coins: std.Coins{{Denom: denom, Amount: 5}}}}))
		require.Equal(t, int64(100), env.bankk.TotalSupply(ctx, denom),
			"a transfer must not change supply")

		require.NoError(t, env.bankk.BurnCoins(ctx, b, std.Coins{{Denom: denom, Amount: 25}}))
		require.Equal(t, int64(75), env.bankk.TotalSupply(ctx, denom))

		// Burning the rest deletes the record rather than storing a zero.
		require.NoError(t, env.bankk.BurnCoins(ctx, a, std.Coins{{Denom: denom, Amount: 75}}))
		require.Zero(t, env.bankk.TotalSupply(ctx, denom))
		require.False(t, ctx.Store(env.key).Has(nil, SupplyKey(denom)),
			"a fully burned denom must leave no record")
	}
}

// The counter is what bounds total supply. Before it existed, two addresses could
// each hold MaxInt64 of one denom, because AddCoins bounds each balance and nothing
// bounded the sum.
func TestSupplyIsCappedAtMaxInt64(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx
	a := crypto.AddressFromPreimage([]byte("a"))
	b := crypto.AddressFromPreimage([]byte("b"))

	require.NoError(t, env.bankk.MintCoins(ctx, a,
		std.Coins{{Denom: testRealmDenom, Amount: math.MaxInt64}}))

	err := env.bankk.MintCoins(ctx, b, std.Coins{{Denom: testRealmDenom, Amount: 1}})
	require.Error(t, err, "a mint past MaxInt64 must be refused")
	require.Equal(t, int64(math.MaxInt64), env.bankk.TotalSupply(ctx, testRealmDenom),
		"a refused mint must leave supply untouched")
	require.Zero(t, env.bankk.GetCoin(ctx, b, testRealmDenom),
		"a refused mint must credit nothing")
}

// A burn of more than was ever minted means the counter already disagreed with the
// balances. It must be refused before the debit rather than making it worse.
func TestBurnBelowRecordedSupplyIsRefusedBeforeDebiting(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx
	a := crypto.AddressFromPreimage([]byte("a"))
	require.NoError(t, env.bankk.SetCoins(ctx, a, std.Coins{{Denom: testRealmDenom, Amount: 100}}))
	// SetCoins is supply-blind, so the counter reads zero while 100 is held.
	require.Zero(t, env.bankk.TotalSupply(ctx, testRealmDenom))

	require.Error(t, env.bankk.BurnCoins(ctx, a, std.Coins{{Denom: testRealmDenom, Amount: 10}}))
	require.Equal(t, int64(100), env.bankk.GetCoin(ctx, a, testRealmDenom),
		"a refused burn must not have debited")
}

// RecomputeSupply must be right regardless of how an account was funded, including
// the genesis shape where the account object is written before SetCoins.
func TestRecomputeSupplyCoversBothTiersAndGenesisShape(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx
	a := crypto.AddressFromPreimage([]byte("a"))
	b := crypto.AddressFromPreimage([]byte("b"))

	// The genesis shape: account object prewritten with the full pre-split set,
	// then SetCoins narrows it. A delta hook computes zero here.
	full := std.Coins{{Denom: testRealmDenom, Amount: 500}, {Denom: testAccountDenom, Amount: 250}}
	full.Sort()
	acc := env.acck.NewAccountWithAddress(ctx, a)
	require.NoError(t, acc.SetCoins(std.Coins{{Denom: testAccountDenom, Amount: 250}}))
	env.acck.SetAccount(ctx, acc)
	require.NoError(t, env.bankk.SetCoins(ctx, a, full))
	require.NoError(t, env.bankk.SetCoins(ctx, b, std.Coins{{Denom: testAccountDenom, Amount: 1000}}))

	env.bankk.RecomputeSupply(ctx)
	require.Equal(t, int64(500), env.bankk.TotalSupply(ctx, testRealmDenom))
	require.Equal(t, int64(1250), env.bankk.TotalSupply(ctx, testAccountDenom))

	// Idempotent, and it clears records for denoms that no longer exist.
	env.bankk.RecomputeSupply(ctx)
	require.Equal(t, int64(1250), env.bankk.TotalSupply(ctx, testAccountDenom))
	env.bankk.setSupply(ctx, "gone", 42)
	env.bankk.RecomputeSupply(ctx)
	require.Zero(t, env.bankk.TotalSupply(ctx, "gone"), "a stale record must be cleared")

	msg, broken := SupplyInvariant(env.bankk.ViewKeeper)(ctx)
	require.False(t, broken, "state seeded by RecomputeSupply must be healthy:\n%s", msg)
}

func TestSupplyInvariantReportsBothDirections(t *testing.T) {
	t.Parallel()

	a := crypto.AddressFromPreimage([]byte("a"))

	t.Run("held but never recorded", func(t *testing.T) {
		t.Parallel()
		env := setupTestEnv()
		// AddCoins credits without touching the counter — an unaccounted mint.
		require.NoError(t, env.bankk.AddCoins(env.ctx, a,
			std.Coins{{Denom: testRealmDenom, Amount: 7}}))
		msg, broken := SupplyInvariant(env.bankk.ViewKeeper)(env.ctx)
		require.True(t, broken)
		require.Contains(t, msg, "is held but there is no supply record")
	})

	t.Run("recorded but not held", func(t *testing.T) {
		t.Parallel()
		env := setupTestEnv()
		env.bankk.setSupply(env.ctx, "atom", 5)
		msg, broken := SupplyInvariant(env.bankk.ViewKeeper)(env.ctx)
		require.True(t, broken)
		require.Contains(t, msg, "no address holds any")
	})

	t.Run("recorded amount disagrees with the sum", func(t *testing.T) {
		t.Parallel()
		env := setupTestEnv()
		require.NoError(t, env.bankk.MintCoins(env.ctx, a,
			std.Coins{{Denom: testRealmDenom, Amount: 10}}))
		require.NoError(t, env.bankk.AddCoins(env.ctx, a,
			std.Coins{{Denom: testRealmDenom, Amount: 5}}))
		msg, broken := SupplyInvariant(env.bankk.ViewKeeper)(env.ctx)
		require.True(t, broken)
		require.Contains(t, msg, "recorded as 10 but 15 is held")
	})

	t.Run("corrupt supply value is reported, not panicked on", func(t *testing.T) {
		t.Parallel()
		env := setupTestEnv()
		env.ctx.Store(env.key).Set(nil, SupplyKey("atom"), []byte{1, 2, 3})
		var msg string
		var broken bool
		require.NotPanics(t, func() {
			msg, broken = SupplyInvariant(env.bankk.ViewKeeper)(env.ctx)
		})
		require.True(t, broken)
		require.Contains(t, msg, "expected 8 bytes, got 3")
	})
}
