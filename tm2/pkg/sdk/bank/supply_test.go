package bank

import (
	"math"
	"strings"
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
	neighbours := append(mainStoreNeighbours(addr), BalanceKey(addr, "atom"))
	assertKeyspaceIsDisjoint(t, SupplyPrefix, SupplyKey(testRealmDenom), neighbours)
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
	// The reason must survive %v. IssueCoin panics on this error and a recovered
	// panic is rendered with %v, which discards a std.Err* message — so wrapping it
	// that way would leave a realm author with only "invalid coins error".
	require.Contains(t, err.Error(), testRealmDenom,
		"the refusal must name the denom where the caller can see it")
	require.Equal(t, int64(math.MaxInt64), env.bankk.TotalSupply(ctx, testRealmDenom),
		"a refused mint must leave supply untouched")
	require.Zero(t, env.bankk.GetCoin(ctx, b, testRealmDenom),
		"a refused mint must credit nothing")
}

// Mint and burn reject a non-positive amount, which the .gno layer does not guard —
// IssueCoin passes the caller's int64 straight through. The reason has to survive %v
// for the same reason as the range error above.
func TestMintAndBurnSayWhyANonPositiveAmountIsRefused(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	a := crypto.AddressFromPreimage([]byte("nonpositive"))
	bad := std.Coins{{Denom: testRealmDenom, Amount: -5}}

	for name, err := range map[string]error{
		"mint": env.bankk.MintCoins(env.ctx, a, bad),
		"burn": env.bankk.BurnCoins(env.ctx, a, bad),
	} {
		require.Error(t, err, "%s of a negative amount must be refused", name)
		require.Contains(t, err.Error(), "positive",
			"%s must say what the rule is, not just that something was invalid", name)
		require.Contains(t, err.Error(), testRealmDenom,
			"%s must name the denom where the caller can see it", name)
	}
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

// A supply key whose denom cannot be recovered is only reachable past the keeper,
// which is the point: setSupply always builds a well-formed key, so nothing else
// exercises the report.
// The twin of TestOverlongDenomNeverReachesTheStore. TotalCoin is reachable from
// an unauthenticated qeval, so an unbounded denom here is a free way to make the
// node hash and copy megabytes per call.
func TestTotalSupplyRejectsAnOverlongDenomBeforeTheStore(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	huge := "/gno.land/r/x/" + strings.Repeat("z", 1<<20) + ":c"
	meter := store.NewGasMeter(100_000_000_000)
	cctx, _ := env.ctx.CacheContext()
	cctx = cctx.WithGasMeter(meter)

	require.Zero(t, env.bankk.TotalSupply(cctx, huge))
	require.Zero(t, meter.GasConsumed(),
		"a denom that cannot have a supply record must not reach the store")
}

// computeSupply builds a map keyed by denom while sweeping the whole balance
// keyspace, so without a cap a chain with enough denoms makes the sweep allocate
// without bound. It must fail rather than return a short answer, which would make
// the supply invariant compare against a total it silently truncated.
func TestComputeSupplyRefusesToAllocatePastItsCap(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	addr := crypto.AddressFromPreimage([]byte("many-denoms"))
	require.NoError(t, env.bankk.SetCoins(env.ctx, addr, std.Coins{
		{Denom: "/gno.land/r/a/x:one", Amount: 1},
		{Denom: "/gno.land/r/a/x:two", Amount: 1},
	}))

	_, err := computeSupply(env.ctx, env.bankk.ViewKeeper, 1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "distinct denoms")

	totals, err := computeSupply(env.ctx, env.bankk.ViewKeeper, 2)
	require.NoError(t, err, "the cap must admit exactly as many as it names")
	require.Len(t, totals, 2)
}

// A failure to total the balances must not blind the structural checks on the supply
// keyspace. Totalling bails past maxSupplyDenoms and on an untotallable sum, both of
// which an attacker influences — denom count is unbounded — so if the record sweep ran
// second behind an early return, inflating the denom count would also hide a malformed
// or corrupt supply record.
func TestSupplyInvariantChecksRecordsEvenWhenItCannotTotal(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx
	a := crypto.AddressFromPreimage([]byte("blind-a"))
	b := crypto.AddressFromPreimage([]byte("blind-b"))

	// A malformed record, reachable only past the keeper.
	rawSet(t, env, []byte(SupplyPrefix), encodeBalance(5))

	// And a state totalling cannot complete: two individually legal balances whose
	// sum leaves int64.
	require.NoError(t, env.bankk.SetCoins(ctx, a, std.Coins{{Denom: testRealmDenom, Amount: math.MaxInt64}}))
	require.NoError(t, env.bankk.SetCoins(ctx, b, std.Coins{{Denom: testRealmDenom, Amount: 1000}}))

	msg, broken := SupplyInvariant(env.bankk.ViewKeeper)(ctx)
	require.True(t, broken)
	require.Contains(t, msg, "cannot total the balances",
		"the invariant must say it could not verify the totals")
	require.Contains(t, msg, "malformed supply key",
		"and must still report the malformed record it can see without them")
}

func TestSupplyInvariantReportsAnUnparseableKey(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	// SupplyPrefix with no denom after it.
	rawSet(t, env, []byte(SupplyPrefix), encodeBalance(5))

	msg, broken := SupplyInvariant(env.bankk.ViewKeeper)(env.ctx)
	require.True(t, broken)
	require.Contains(t, msg, "denom")
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

// If the balances cannot be totalled, that must be an error, not truncated totals.
// The account-tier walk reports through a callback whose return value only stops the
// iteration, so an error raised inside it has to be carried out explicitly —
// otherwise computeSupply returns a partial sum with err == nil, RecomputeSupply
// seeds that as the genesis supply, and SupplyInvariant then agrees with it. The one
// redundancy check in the set would be self-consistently wrong.
func TestComputeSupplyReportsAnUntotallableSum(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx
	a := crypto.AddressFromPreimage([]byte("a"))
	b := crypto.AddressFromPreimage([]byte("b"))

	// Two account-tier balances that together exceed int64. Each is individually
	// legal, because AddCoins bounds each balance and nothing bounds the sum.
	require.NoError(t, env.bankk.SetCoins(ctx, a,
		std.Coins{{Denom: testAccountDenom, Amount: math.MaxInt64}}))
	require.NoError(t, env.bankk.SetCoins(ctx, b,
		std.Coins{{Denom: testAccountDenom, Amount: 1000}}))

	_, err := computeSupply(ctx, env.bankk.ViewKeeper, 0)
	require.Error(t, err, "an untotallable sum must be reported, not truncated")
	require.Contains(t, err.Error(), "sum past int64")

	// And it must refuse to seed rather than write the truncated number.
	require.Panics(t, func() { env.bankk.RecomputeSupply(ctx) },
		"RecomputeSupply must refuse to seed a supply it cannot compute")

	// The invariant must say it could not verify, rather than reporting healthy.
	msg, broken := SupplyInvariant(env.bankk.ViewKeeper)(ctx)
	require.True(t, broken, "the invariant must not report healthy here")
	require.Contains(t, msg, "supply was NOT verified")
}

// Both money-path helpers depend on their amount being strictly ascending, and both
// are internal, so the guards are driven directly. Every public entry point validates
// first, which is exactly what these must not rely on.
func TestInternalHelpersRejectDuplicateDenoms(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx
	addr := crypto.AddressFromPreimage([]byte("holder"))
	require.NoError(t, env.bankk.SetCoins(ctx, addr, std.Coins{{Denom: testRealmDenom, Amount: 100}}))
	env.bankk.RecomputeSupply(ctx)

	dup := std.Coins{{Denom: testRealmDenom, Amount: 5}, {Denom: testRealmDenom, Amount: 5}}

	// nextSupply: both entries would compute from the same old value, so the second
	// write would lose an increment and the counter would under-count the mint.
	_, err := env.bankk.nextSupply(ctx, dup, 1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate denom")

	// subtract: both entries would compute from the same starting balance, so only
	// one debit would land while the caller believes two did.
	err = env.bankk.subtract(ctx, env.acck.GetAccount(ctx, addr), addr, dup, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate denom")
	require.Equal(t, int64(100), env.bankk.GetCoin(ctx, addr, testRealmDenom),
		"a rejected debit must not have moved the balance")
}

// subtract takes the account as a parameter to save a read, and nothing in the types
// ties it to the address, so a mismatched pair would debit one address's account tier
// and another's split tier.
func TestSubtractRejectsAnAccountForTheWrongAddress(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx
	victim := crypto.AddressFromPreimage([]byte("victim"))
	attacker := crypto.AddressFromPreimage([]byte("attacker"))
	require.NoError(t, env.bankk.SetCoins(ctx, victim,
		std.Coins{{Denom: testAccountDenom, Amount: 500}}))
	require.NoError(t, env.bankk.SetCoins(ctx, attacker,
		std.Coins{{Denom: testRealmDenom, Amount: 10}}))

	err := env.bankk.subtract(ctx, env.acck.GetAccount(ctx, victim), attacker,
		std.Coins{{Denom: testAccountDenom, Amount: 500}}, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "was passed for address")
	require.Equal(t, int64(500), env.bankk.GetCoin(ctx, victim, testAccountDenom),
		"the victim's balance must be untouched")
}
