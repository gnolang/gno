package bank

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"math"
	"math/rand"
	"slices"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	tu "github.com/gnolang/gno/tm2/pkg/sdk/testutils"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/require"
)

const testRealmDenom = "/gno.land/r/demo/foo:gold"

// TestBalanceKeyFormat pins the on-disk key layout. The format is consensus
// state, so a change here is a chain-breaking change and should be deliberate.
func TestBalanceKeyFormat(t *testing.T) {
	t.Parallel()

	addr := crypto.AddressFromPreimage([]byte("addr1"))
	key := BalanceKey(addr, testRealmDenom)

	want := append(append([]byte("/b/"), addr[:]...), testRealmDenom...)
	require.True(t, bytes.Equal(want, key),
		"balance key layout changed: got %X want %X", key, want)
	require.Len(t, key, len("/b/")+crypto.AddressSize+len(testRealmDenom))

	// The prefix must cover the key, so iterating it reaches this balance...
	require.True(t, bytes.HasPrefix(key, BalancePrefixKey(addr)))
	// ...and must be no broader than that, or one account's iteration would
	// return another's balances. HasPrefix alone is only a lower bound.
	require.Len(t, BalancePrefixKey(addr), len("/b/")+crypto.AddressSize)
}

// TestDenomFromBalanceKey checks the address/denom split. It works only because
// crypto.Address is fixed width; see the note in balance.go.
func TestDenomFromBalanceKey(t *testing.T) {
	t.Parallel()

	addr := crypto.AddressFromPreimage([]byte("addr1"))
	for _, denom := range []string{
		testRealmDenom,
		"/gno.land/r/my-org/token:gold",             // hyphen
		"/gno.land/r/demo/foo:a1b",                  // short base
		"/" + string(make([]byte, 0)) + "a.b:c/d:e", // separators in the denom
	} {
		got, err := denomFromBalanceKey(BalanceKey(addr, denom))
		require.NoError(t, err)
		require.Equal(t, denom, got)
	}

	// A key with no denom is malformed, not an empty denom.
	_, err := denomFromBalanceKey(BalancePrefixKey(addr))
	require.Error(t, err)
}

func TestEncodeDecodeBalance(t *testing.T) {
	t.Parallel()

	for _, amount := range []int64{1, 2, 255, 256, 1 << 40, 1<<63 - 1} {
		require.Equal(t, amount, decodeBalance(encodeBalance(amount)))
		require.Len(t, encodeBalance(amount), balanceAmountLen,
			"encoding must be fixed width so write gas does not depend on the amount")
	}

	// A stored balance is always positive: zero means "absent" and is deleted.
	require.Panics(t, func() { encodeBalance(0) })
	require.Panics(t, func() { encodeBalance(-1) })

	// Anything else in the value slot is state corruption, not a zero balance.
	require.Panics(t, func() { decodeBalance(nil) })
	require.Panics(t, func() { decodeBalance([]byte{1, 2, 3}) })
	require.Panics(t, func() { decodeBalance(make([]byte, balanceAmountLen)) })
}

// TestRealmBalanceRoundTrip covers the per-denom accessors, including that a
// balance drained to zero deletes its key rather than storing a zero.
func TestRealmBalanceRoundTrip(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx
	addr := crypto.AddressFromPreimage([]byte("addr1"))

	require.Zero(t, env.bankk.getSplitBalance(ctx, addr, testRealmDenom),
		"an unset balance reads as zero, not as an error")

	env.bankk.setSplitBalance(ctx, addr, testRealmDenom, 42)
	require.Equal(t, int64(42), env.bankk.getSplitBalance(ctx, addr, testRealmDenom))

	env.bankk.setSplitBalance(ctx, addr, testRealmDenom, 0)
	require.Zero(t, env.bankk.getSplitBalance(ctx, addr, testRealmDenom))
	require.Nil(t,
		ctx.Store(env.key).Get(ctx.GasContext(), BalanceKey(addr, testRealmDenom)),
		"a zero balance must delete the key, not store a zero")
}

// TestRealmCoinsOrdering pins that iteration yields denoms in ascending order,
// which is the precondition for GetCoins merging without revalidating.
func TestRealmCoinsOrdering(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx
	addr := crypto.AddressFromPreimage([]byte("addr1"))

	// Written out of order on purpose.
	for _, denom := range []string{
		"/gno.land/r/demo/z:zzz",
		"/gno.land/r/demo/a:aaa",
		"/gno.land/r/demo/m:mmm",
	} {
		env.bankk.setSplitBalance(ctx, addr, denom, 7)
	}

	coins := env.bankk.splitCoins(ctx, addr)
	require.Len(t, coins, 3)
	require.Equal(t, "/gno.land/r/demo/a:aaa", coins[0].Denom)
	require.Equal(t, "/gno.land/r/demo/m:mmm", coins[1].Denom)
	require.Equal(t, "/gno.land/r/demo/z:zzz", coins[2].Denom)
	require.True(t, coins.IsValid(), "iteration order must satisfy std.Coins")

	// Another address's balances must not leak into this one's prefix.
	other := crypto.AddressFromPreimage([]byte("addr2"))
	env.bankk.setSplitBalance(ctx, other, "/gno.land/r/demo/b:bbb", 1)
	require.Len(t, env.bankk.splitCoins(ctx, addr), 3)
	require.Len(t, env.bankk.splitCoins(ctx, other), 1)
}

// TestGetCoinsMergeIsSorted pins that GetCoins returns valid std.Coins when the
// two tiers interleave. With an allowlist the tiers are no longer separable by
// first byte — a split-tier "atom" sorts before an account-tier "ugnot", and a
// split-tier "zzz" sorts after it — so a concatenation in either direction would
// be unsorted and GetCoins must merge.
func TestGetCoinsMergeIsSorted(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx
	addr := crypto.AddressFromPreimage([]byte("addr1"))

	// One account-tier denom, and split-tier denoms that must land on both sides
	// of it: "atom" before "ugnot", "zeta" after, plus a "/"-prefixed one first.
	require.NoError(t, env.bankk.SetCoins(ctx, addr, std.Coins{
		{Denom: "/gno.land/r/demo/a:aaa", Amount: 4},
		{Denom: "atom", Amount: 1},
		{Denom: "ugnot", Amount: 2},
		{Denom: "zeta", Amount: 3},
	}))

	// Only ugnot is in the account object; the rest have their own keys.
	acc := env.acck.GetAccount(ctx, addr)
	require.Equal(t, std.Coins{{Denom: "ugnot", Amount: 2}}, acc.GetCoins())

	coins := env.bankk.GetCoins(ctx, addr)
	require.True(t, coins.IsValid(),
		"merged coins must be sorted and duplicate-free: %s", coins)
	require.Len(t, coins, 4)
	require.Equal(t, "/gno.land/r/demo/a:aaa", coins[0].Denom)
	require.Equal(t, "atom", coins[1].Denom)
	require.Equal(t, "ugnot", coins[2].Denom)
	require.Equal(t, "zeta", coins[3].Denom,
		"a split denom sorting after the account-tier denom must still be last")
}

// TestSetCoinsReplacesRatherThanMerges pins that SetCoins deletes realm balances
// absent from the new set. A merge would make genesis import non-idempotent and
// would leak balances across a replace.
func TestSetCoinsReplacesRatherThanMerges(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx
	addr := crypto.AddressFromPreimage([]byte("addr1"))

	require.NoError(t, env.bankk.SetCoins(ctx, addr, std.Coins{
		{Denom: "/gno.land/r/demo/a:aaa", Amount: 1},
		{Denom: "/gno.land/r/demo/b:bbb", Amount: 2},
	}))
	require.NoError(t, env.bankk.SetCoins(ctx, addr, std.Coins{
		{Denom: "/gno.land/r/demo/b:bbb", Amount: 9},
	}))

	coins := env.bankk.GetCoins(ctx, addr)
	require.Len(t, coins, 1, "the dropped denom must be deleted, not retained: %s", coins)
	require.Equal(t, "/gno.land/r/demo/b:bbb", coins[0].Denom)
	require.Equal(t, int64(9), coins[0].Amount)
}

// TestAddCoinsCreatesAccount pins that receiving coins still creates the
// account. Without it a recipient could not sign (auth.GetSignerAcc rejects an
// address with no account), so the funds would be visible but unspendable.
func TestAddCoinsCreatesAccount(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx

	// Realm-denom-only credit: nothing would otherwise write the account object.
	addr := crypto.AddressFromPreimage([]byte("fresh-realm"))
	require.Nil(t, env.acck.GetAccount(ctx, addr))
	require.NoError(t, env.bankk.AddCoins(ctx, addr, std.Coins{{Denom: testRealmDenom, Amount: 1}}))
	require.NotNil(t, env.acck.GetAccount(ctx, addr),
		"a realm-denom credit must still create the recipient's account")

	// Genesis-denom credit.
	addr2 := crypto.AddressFromPreimage([]byte("fresh-genesis"))
	require.Nil(t, env.acck.GetAccount(ctx, addr2))
	require.NoError(t, env.bankk.AddCoins(ctx, addr2, std.Coins{{Denom: "ugnot", Amount: 1}}))
	require.NotNil(t, env.acck.GetAccount(ctx, addr2))
}

// TestAccountIterationExcludesBalances pins that the balance keyspace is outside
// the account keyspace. "/b/" sorts after the "/a/" prefix range, so an account
// iteration must never observe a balance key — if it did, decodeAccount would be
// handed a balance value.
func TestAccountIterationExcludesBalances(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx

	addr := crypto.AddressFromPreimage([]byte("addr1"))
	env.acck.SetAccount(ctx, env.acck.NewAccountWithAddress(ctx, addr))
	env.bankk.setSplitBalance(ctx, addr, testRealmDenom, 1)

	count := 0
	env.acck.IterateAccounts(ctx, func(acc std.Account) bool {
		count++
		return false
	})
	require.Equal(t, 1, count, "account iteration picked up a balance key")

	// And the reverse: the balance prefix must not reach account keys.
	iter := store.PrefixIterator(ctx.GasContext(), ctx.Store(env.key), []byte(BalancePrefix))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		require.True(t, bytes.HasPrefix(iter.Key(), []byte(BalancePrefix)))
		require.False(t, bytes.HasPrefix(iter.Key(), []byte(auth.AddressStoreKeyPrefix)))
	}
}

// TestBalancePrefixDoesNotCollide guards the keyspace split at the byte level.
// mainStoreNeighbours are the foreign keyspaces sharing the main store, as real
// keys. One list so that a new keyspace cannot be checked against one bank prefix
// and forgotten for the other — the two collision tests below had drifted apart.
func mainStoreNeighbours(addr crypto.Address) [][]byte {
	return [][]byte{
		auth.AddressStoreKey(addr),  // /a/<addr>
		auth.SessionPrefixKey(addr), // /a/<addr>/s/
		[]byte("/pv/bank:x"),        // params
		[]byte(auth.GasPriceKey),
		[]byte(auth.GlobalAccountNumberKey),
		[]byte("consensus_params"),
		[]byte("last_header"),
		[]byte("pkg:gno.land/r/x"), // GnoVM mempackages
		[]byte("oid:x"),            // GnoVM object ids
	}
}

// assertKeyspaceIsDisjoint checks both that no neighbour key and sample share a
// prefix relation, and that a live sweep of prefix sees only sample. The prefix
// relation alone is not enough: bleed is a property of iteration, and that is what
// splitCoins and the invariants depend on.
func assertKeyspaceIsDisjoint(t *testing.T, prefix string, sample []byte, neighbours [][]byte) {
	t.Helper()

	for _, n := range neighbours {
		require.False(t, hasPrefixBytes(sample, n), "%X must not be prefixed by %X", sample, n)
		require.False(t, hasPrefixBytes(n, []byte(prefix)), "%X must not fall under %q", n, prefix)
	}

	env := setupTestEnv()
	stor := env.ctx.Store(env.key)
	for _, n := range neighbours {
		stor.Set(nil, n, []byte{1})
	}
	stor.Set(nil, sample, encodeBalance(5))

	iter := store.PrefixIterator(nil, stor, []byte(prefix))
	defer iter.Close()
	var seen int
	for ; iter.Valid(); iter.Next() {
		seen++
		require.Equal(t, sample, iter.Key(), "a sweep of %q saw a foreign key", prefix)
	}
	require.Equal(t, 1, seen)
}

// "/b/" must not be a prefix of, or prefixed by, any neighbouring keyspace,
// including the unprefixed keys: the shared store holds bare names as well as
// "/..." prefixes, so checking only the latter would be incomplete.
func TestBalancePrefixDoesNotCollide(t *testing.T) {
	t.Parallel()

	addr := crypto.AddressFromPreimage([]byte("addr1"))
	neighbours := append(mainStoreNeighbours(addr), SupplyKey(testRealmDenom))
	assertKeyspaceIsDisjoint(t, BalancePrefix, BalanceKey(addr, testRealmDenom), neighbours)

	// A balance key must sort strictly above the whole "/a/" iteration range,
	// whose upper bound is PrefixEndBytes("/a/").
	end := storetypes.PrefixEndBytes([]byte(auth.AddressStoreKeyPrefix))
	require.Positive(t, bytes.Compare(BalanceKey(addr, testRealmDenom), end),
		"balance keys must sort outside the account iteration range")
}

// TestSubtractRealmInsufficientFunds is a mint guard. If the realm-tier
// sufficiency check were dropped, SubtractCoins would silently clamp to zero and
// report success, and SendCoins would then credit the recipient the full amount —
// creating coins out of nothing.
func TestSubtractRealmInsufficientFunds(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	env.bankk.setSplitBalance(ctx, addr, testRealmDenom, 7)

	err := env.bankk.SubtractCoins(ctx, addr, std.Coins{{Denom: testRealmDenom, Amount: 100}})
	require.Error(t, err, "over-spending a realm denom must fail, not clamp")
	require.Equal(t, int64(7), env.bankk.GetCoin(ctx, addr, testRealmDenom),
		"a failed debit must leave the balance untouched")

	// Exactly one over must fail too — the gross case alone would not catch an
	// off-by-one in the comparison.
	require.Error(t, env.bankk.SubtractCoins(ctx, addr, std.Coins{{Denom: testRealmDenom, Amount: 8}}))
	require.Equal(t, int64(7), env.bankk.GetCoin(ctx, addr, testRealmDenom))

	// And exactly the whole balance must succeed, draining the key.
	require.NoError(t, env.bankk.SubtractCoins(ctx, addr, std.Coins{{Denom: testRealmDenom, Amount: 7}}))
	require.Zero(t, env.bankk.GetCoin(ctx, addr, testRealmDenom))
}

// TestAddCoinsRealmAmount pins the credited amount, not merely that the account
// was created. Deleting the realm credit loop entirely used to pass every unit
// test.
func TestAddCoinsRealmAmount(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx
	addr := crypto.AddressFromPreimage([]byte("addr1"))

	require.NoError(t, env.bankk.AddCoins(ctx, addr, std.Coins{{Denom: testRealmDenom, Amount: 40}}))
	require.Equal(t, int64(40), env.bankk.GetCoin(ctx, addr, testRealmDenom))
	require.NoError(t, env.bankk.AddCoins(ctx, addr, std.Coins{{Denom: testRealmDenom, Amount: 2}}))
	require.Equal(t, int64(42), env.bankk.GetCoin(ctx, addr, testRealmDenom),
		"a second credit must accumulate onto the first")
}

// TestAddCoinsRealmOverflowPanics pins that per-denom addition goes through
// overflow.Add. A plain + would wrap to a negative balance, which is a mint.
func TestAddCoinsRealmOverflowPanics(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	env.bankk.setSplitBalance(ctx, addr, testRealmDenom, math.MaxInt64)

	// Match the message: a plain + would also panic, but via encodeBalance's
	// non-positive guard after wrapping, so a bare require.Panics cannot tell
	// the two apart.
	require.PanicsWithValue(t,
		fmt.Sprintf("coin add overflow: %d%s + 1%s", int64(math.MaxInt64), testRealmDenom, testRealmDenom),
		func() {
			_ = env.bankk.AddCoins(ctx, addr, std.Coins{{Denom: testRealmDenom, Amount: 1}})
		}, "overflow must be caught by overflow.Add, not by the encoder")
}

// TestMixedTierFailureIsAtomic pins that a debit spanning both tiers leaves
// balances untouched when it fails. Writing realm debits before checking genesis
// solvency would leave the realm side already spent.
func TestMixedTierFailureIsAtomic(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	require.NoError(t, env.bankk.SetCoins(ctx, addr, std.Coins{
		{Denom: testRealmDenom, Amount: 100},
		{Denom: "ugnot", Amount: 10},
	}))

	// The realm half is affordable, the genesis half is not.
	err := env.bankk.SubtractCoins(ctx, addr, std.Coins{
		{Denom: testRealmDenom, Amount: 100},
		{Denom: "ugnot", Amount: 999},
	})
	require.Error(t, err)
	require.Equal(t, int64(100), env.bankk.GetCoin(ctx, addr, testRealmDenom),
		"the affordable realm debit must not have been applied")
	require.Equal(t, int64(10), env.bankk.GetCoin(ctx, addr, "ugnot"))

	// And the mirror: genesis affordable, realm not.
	err = env.bankk.SubtractCoins(ctx, addr, std.Coins{
		{Denom: testRealmDenom, Amount: 999},
		{Denom: "ugnot", Amount: 10},
	})
	require.Error(t, err)
	require.Equal(t, int64(100), env.bankk.GetCoin(ctx, addr, testRealmDenom))
	require.Equal(t, int64(10), env.bankk.GetCoin(ctx, addr, "ugnot"))
}

// TestBalanceAmountWidthIsPinned guards the encoded width separately from the
// key layout. Both are consensus state.
func TestBalanceAmountWidthIsPinned(t *testing.T) {
	t.Parallel()

	require.Equal(t, 8, balanceAmountLen,
		"the encoded balance width is consensus state; changing it is a chain break")
	// Byte golden: big-endian. Round-tripping alone would not notice a swap.
	require.Equal(t, []byte{0, 0, 0, 0, 0, 0, 1, 0}, encodeBalance(256),
		"the encoded balance is big-endian; the byte order is consensus state")

	// A value above MaxInt64 is corruption, not a large balance.
	tooBig := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	require.Panics(t, func() { decodeBalance(tooBig) })
}

// TestDeductFeesRealmDenom pins that a fee paid in a realm-issued denom works.
// The check lives here rather than in the auth package because auth's
// DummyBankKeeper keeps every balance in the account object and so cannot
// distinguish the two tiers — against it, the old acc.GetCoins() pre-check and
// the per-denom bank read are indistinguishable.
func TestDeductFeesRealmDenom(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx
	payer := crypto.AddressFromPreimage([]byte("payer"))
	collector := crypto.AddressFromPreimage([]byte("collector"))

	require.NoError(t, env.bankk.SetCoins(ctx, payer, std.Coins{{Denom: testRealmDenom, Amount: 100}}))
	acc := env.acck.GetAccount(ctx, payer)
	require.NotNil(t, acc)

	fees := std.Coins{{Denom: testRealmDenom, Amount: 30}}
	res := auth.DeductFees(env.bankk, ctx, acc, collector, fees)
	require.True(t, res.IsOK(),
		"a fee in a realm denom must be payable; the balance is not in the account object: %v", res.Error)
	require.Equal(t, int64(70), env.bankk.GetCoin(ctx, payer, testRealmDenom))
	require.Equal(t, int64(30), env.bankk.GetCoin(ctx, collector, testRealmDenom))

	// Insufficient must still surface as InsufficientFundsError, which gnokey
	// and the ante tests distinguish from InsufficientCoinsError.
	res = auth.DeductFees(env.bankk, ctx, acc, collector,
		std.Coins{{Denom: testRealmDenom, Amount: 1_000}})
	require.False(t, res.IsOK())
	require.IsType(t, std.InsufficientFundsError{}, res.Error)
}

// TestVictimCostIsIndependentOfJunkDenoms is the regression test for the attack
// the split exists to close: a realm can mint any denom into any address without
// the recipient's consent, and while those balances sat in the account object
// they made every transaction the victim sent more expensive, without bound.
//
// The assertion is comparative rather than absolute. Both arms mint the same
// number of junk denoms, so the store holds the same number of keys and the
// iavl depth term is the same; the only difference is *whose* balances they are.
// An absolute gas figure would drift with tree size in this test environment and
// would be brittle. On chain the depth term is pinned flat, so there the two
// figures coincide exactly.
func TestVictimCostIsIndependentOfJunkDenoms(t *testing.T) {
	t.Parallel()

	const junk = 200

	// transferCost mints `junk` realm denoms to `holder`, then measures the gas a
	// plain ugnot transfer from victim to dest costs.
	transferCost := func(t *testing.T, junkToVictim bool) store.Gas {
		t.Helper()
		env := setupTestEnv()
		victim := crypto.AddressFromPreimage([]byte("victim"))
		dest := crypto.AddressFromPreimage([]byte("dest"))
		bystander := crypto.AddressFromPreimage([]byte("bystander"))

		require.NoError(t, env.bankk.SetCoins(env.ctx, victim, std.Coins{{Denom: "ugnot", Amount: 1_000_000}}))

		holder := bystander
		if junkToVictim {
			holder = victim
		}
		for i := range junk {
			denom := fmt.Sprintf("/gno.land/r/spam/p%d:junk", i)
			require.NoError(t, env.bankk.AddCoins(env.ctx, holder, std.Coins{{Denom: denom, Amount: 1}}))
		}

		// Measure on a cache context, which is the layer that charges gas.
		meter := store.NewGasMeter(100_000_000_000)
		cctx, _ := env.ctx.CacheContext()
		cctx = cctx.WithGasMeter(meter)
		before := meter.GasConsumed()
		require.NoError(t, env.bankk.SendCoins(cctx, victim, dest, std.Coins{{Denom: "ugnot", Amount: 1}}))
		return meter.GasConsumed() - before
	}

	loaded := transferCost(t, true)
	control := transferCost(t, false)

	require.Positive(t, loaded, "the measurement harness must actually charge gas")
	require.LessOrEqual(t, loaded, control,
		"a ugnot transfer must not cost more because the sender was sent %d unsolicited "+
			"realm denoms (loaded=%d control=%d)", junk, loaded, control)
}

// TestSendCoinsMixedTierMovesBoth pins that a transfer spanning both tiers moves
// both halves. Skipping either half is a mint on one side and a loss on the
// other, and every other mixed-tier test here covers only the failing case.
func TestSendCoinsMixedTierMovesBoth(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx
	from := crypto.AddressFromPreimage([]byte("from"))
	to := crypto.AddressFromPreimage([]byte("to"))

	require.NoError(t, env.bankk.SetCoins(ctx, from, std.Coins{
		{Denom: testRealmDenom, Amount: 100},
		{Denom: "ugnot", Amount: 500},
	}))

	require.NoError(t, env.bankk.SendCoins(ctx, from, to, std.Coins{
		{Denom: testRealmDenom, Amount: 40},
		{Denom: "ugnot", Amount: 200},
	}))

	// Both tiers debited...
	require.Equal(t, int64(60), env.bankk.GetCoin(ctx, from, testRealmDenom))
	require.Equal(t, int64(300), env.bankk.GetCoin(ctx, from, "ugnot"))
	// ...and both credited. Totals conserved on each denom.
	require.Equal(t, int64(40), env.bankk.GetCoin(ctx, to, testRealmDenom))
	require.Equal(t, int64(200), env.bankk.GetCoin(ctx, to, "ugnot"))
}

// TestSendCoinsMultipleRealmDenoms pins that every realm denom in the set moves,
// not just the first. A loop that breaks early, or that writes one denom's
// balance to every key, is otherwise invisible.
func TestSendCoinsMultipleRealmDenoms(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx
	from := crypto.AddressFromPreimage([]byte("from"))
	to := crypto.AddressFromPreimage([]byte("to"))

	// Distinct amounts, so writing the wrong index is detectable.
	const a, b, c = "/gno.land/r/demo/a:aaa", "/gno.land/r/demo/b:bbb", "/gno.land/r/demo/c:ccc"
	require.NoError(t, env.bankk.SetCoins(ctx, from, std.Coins{
		{Denom: a, Amount: 100},
		{Denom: b, Amount: 200},
		{Denom: c, Amount: 300},
	}))

	require.NoError(t, env.bankk.SendCoins(ctx, from, to, std.Coins{
		{Denom: a, Amount: 1},
		{Denom: b, Amount: 2},
		{Denom: c, Amount: 3},
	}))

	require.Equal(t, int64(99), env.bankk.GetCoin(ctx, from, a))
	require.Equal(t, int64(198), env.bankk.GetCoin(ctx, from, b))
	require.Equal(t, int64(297), env.bankk.GetCoin(ctx, from, c))
	require.Equal(t, int64(1), env.bankk.GetCoin(ctx, to, a))
	require.Equal(t, int64(2), env.bankk.GetCoin(ctx, to, b))
	require.Equal(t, int64(3), env.bankk.GetCoin(ctx, to, c))
}

// TestVestingChecksEveryDenom pins that the lock check covers every denom in the
// amount, not just the first.
//
// This matters more than it looks: realm denoms sort before gas denoms, so in
// a mixed transfer the realm coin is always amt[0] and always has lockedAmt == 0.
// A loop that stopped at the first unlocked denom would therefore skip the lock
// check on every gas denom that followed — a complete vesting bypass — while
// every single-denom vesting test still passed.
func TestVestingChecksEveryDenom(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	addr := crypto.AddressFromPreimage([]byte("vester"))

	locked := std.Coins{{Denom: "ugnot", Amount: 1000}}
	baseAcc := std.NewBaseAccount(addr, locked, nil, 0, 0)
	cva, err := std.NewContinuousVestingAccount(baseAcc, std.VestingSchedule{
		OriginalVesting: locked,
		StartTime:       100,
		EndTime:         200,
	})
	require.NoError(t, err)
	env.acck.SetAccount(env.ctx, cva)
	require.NoError(t, env.bankk.SetCoins(env.ctx, addr, locked))
	env.bankk.setSplitBalance(env.ctx, addr, testRealmDenom, 50)

	// At t=100 nothing has vested. The realm coin is spendable and sorts first;
	// the ugnot is fully locked. The mixed transfer must still be refused.
	ctx := atTime(env, 100)
	err = env.bankk.SubtractCoins(ctx, addr, std.Coins{
		{Denom: testRealmDenom, Amount: 1},
		{Denom: "ugnot", Amount: 1000},
	})
	require.Error(t, err, "an unlocked leading denom must not let a locked one through")
	require.Equal(t, int64(1000), env.bankk.GetCoin(ctx, addr, "ugnot"))
	require.Equal(t, int64(50), env.bankk.GetCoin(ctx, addr, testRealmDenom))
}

// TestTierFollowsAllowlistNotDenomShape pins that storage tier is decided by the
// allowlist alone, never by what the denom looks like.
//
// This is what makes the design safe for denoms nobody has invented yet. An IBC
// voucher ("ibc/<hash>") and a plain "atom" are both unlike a realm denom, but
// both are permissionlessly creatable and unbounded in count, so both must be
// split. A shape-based rule would have filed them in the account object and
// reopened the griefing vector.
func TestTierFollowsAllowlistNotDenomShape(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx
	addr := crypto.AddressFromPreimage([]byte("addr1"))

	// None of these is in the allowlist, whatever they look like: an IBC-shaped
	// voucher, a bare denom, and denoms containing the separators a realm denom
	// uses. All must be split out.
	for _, denom := range []string{
		"ibc/07a1a2b3c4d5e6f7", "atom", "gno.land/r/demo/foo:gold", "a/b", "foo:bar",
	} {
		require.NoError(t, std.ValidateDenom(denom), "test denom must be valid: %s", denom)
		require.NoError(t, env.bankk.AddCoins(ctx, addr, std.Coins{{Denom: denom, Amount: 5}}))
		require.Equal(t, int64(5), env.bankk.GetCoin(ctx, addr, denom),
			"must be readable from the tier it was written to: %s", denom)
		require.Equal(t, int64(5), env.bankk.getSplitBalance(ctx, addr, denom),
			"a denom outside the allowlist must get its own key: %s", denom)

		if acc := env.acck.GetAccount(ctx, addr); acc != nil {
			require.Zero(t, acc.GetCoins().AmountOf(denom),
				"a denom outside the allowlist must not be in the account object: %s", denom)
		}
		require.NoError(t, env.bankk.SubtractCoins(ctx, addr, std.Coins{{Denom: denom, Amount: 5}}))
		require.Zero(t, env.bankk.GetCoin(ctx, addr, denom))
	}

	// And the allowlisted denom does live in the account object.
	require.NoError(t, env.bankk.AddCoins(ctx, addr, std.Coins{{Denom: "ugnot", Amount: 7}}))
	require.Equal(t, int64(7), env.acck.GetAccount(ctx, addr).GetCoins().AmountOf("ugnot"))
	require.Zero(t, env.bankk.getSplitBalance(ctx, addr, "ugnot"))
}

// TestAddCoinsGenesisCreditIgnoresRealmTier pins that the genesis credit is
// computed from the account object alone. Computing it from the full balance set
// would copy realm balances into the account object, so they would exist in both
// tiers at once: GetCoins would then return duplicate denoms — invalid
// std.Coins, which panics chain.Coins.AmountOf on the Gno side — and the phantom
// copy would be permanent.
func TestAddCoinsGenesisCreditIgnoresRealmTier(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx
	addr := crypto.AddressFromPreimage([]byte("addr1"))

	env.bankk.setSplitBalance(ctx, addr, testRealmDenom, 70)
	require.NoError(t, env.bankk.AddCoins(ctx, addr, std.Coins{{Denom: "ugnot", Amount: 5}}))

	acc := env.acck.GetAccount(ctx, addr)
	require.Zero(t, acc.GetCoins().AmountOf(testRealmDenom),
		"the realm balance leaked into the account object: %s", acc.GetCoins())
	require.Len(t, acc.GetCoins(), 1, "account object must hold only gas denoms: %s", acc.GetCoins())

	coins := env.bankk.GetCoins(ctx, addr)
	require.True(t, coins.IsValid(), "duplicate or unsorted coins: %s", coins)
	require.Len(t, coins, 2)
}

// TestConservation checks the keeper against an independent model of what a bank
// is: a map from address to denom to amount.
//
// Every other test here asserts a specific number at a specific address, which
// tells a reader whether the author's arithmetic was right, not whether the
// keeper is correct. This one asserts the properties that actually matter, and
// all four ways this design can lose money are violations of them:
//
//   - a denom written to one tier and read from the other,
//   - a partial write left behind by a failed multi-denom operation,
//   - a mis-indexed write putting one denom's balance under another's key,
//   - two call sites disagreeing about which tier a denom belongs to.
//
// The model is compared after *every* operation, including failed ones, so a
// divergence is reported at the operation that caused it rather than at the end.
// Comparing after a failure is what pins atomicity.
func TestConservation(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()

	addrs := []crypto.Address{
		crypto.AddressFromPreimage([]byte("c1")),
		crypto.AddressFromPreimage([]byte("c2")),
		crypto.AddressFromPreimage([]byte("c3")),
		crypto.AddressFromPreimage([]byte("c4")),
	}
	// Deliberately spans both tiers, and includes split-tier denoms that sort
	// either side of the account-tier one so the merge in GetCoins is exercised
	// in both directions.
	denoms := []string{"ugnot", "atom", "zeta", "/gno.land/r/a:x", "/gno.land/r/b:y"}

	// addrs[0] vests, so the restricted paths run the vesting check on every
	// operation it is the source of. At t=150 half the schedule has elapsed, so
	// vesterLocked of its gas denom can never be spent through SubtractCoins,
	// SendCoins or InputOutputCoins — an invariant the conservation oracle cannot
	// see on its own, since a bypassed lock still conserves the supply.
	// A cliff schedule that has not matured, so the whole original vesting stays
	// locked for the entire run. A linear schedule leaves so much spendable that
	// credits keep the balance clear of the threshold and the invariant is never
	// stressed — verified by instrumenting the minimum observed balance.
	const (
		vesterFunded = 1000
		vesterLocked = vesterFunded
	)
	vester := addrs[0]
	ctx := atTime(env, 150)

	model := map[crypto.Address]map[string]int64{}
	for _, a := range addrs {
		model[a] = map[string]int64{}
	}

	vesting := std.Coins{{Denom: testAccountDenom, Amount: vesterFunded}}
	dva, err := std.NewDelayedVestingAccount(
		// Draw a real account number: hardcoding 0 collides with the first
		// auto-allocated account and leaves the global counter behind the state.
		std.NewBaseAccount(vester, vesting, nil, env.acck.GetNextAccountNumber(ctx), 0),
		std.VestingSchedule{
			OriginalVesting: vesting,
			StartTime:       100,
			EndTime:         200,
			Type:            std.VestingDelayed,
		},
	)
	require.NoError(t, err)
	env.acck.SetAccount(ctx, dva)
	require.NoError(t, env.bankk.SetCoins(ctx, vester, vesting))
	model[vester][testAccountDenom] = vesterFunded
	// SetCoins does not maintain the counter, so seed it from what is now held.
	env.bankk.RecomputeSupply(ctx)

	// modelCoins renders one address's model row as std.Coins, which is what
	// GetCoins must return: ascending, positive, no duplicates.
	modelCoins := func(a crypto.Address) std.Coins {
		var out std.Coins
		for _, d := range slices.Sorted(maps.Keys(model[a])) {
			if amt := model[a][d]; amt > 0 {
				out = append(out, std.Coin{Denom: d, Amount: amt})
			}
		}
		if len(out) == 0 {
			return std.NewCoins()
		}
		return out
	}

	check := func(t *testing.T, step int, op string) {
		t.Helper()
		for _, a := range addrs {
			want := modelCoins(a)
			got := env.bankk.GetCoins(ctx, a)
			require.True(t, got.IsValid(), "step %d (%s): GetCoins invalid: %s", step, op, got)
			require.Equal(t, want.String(), got.String(),
				"step %d (%s): GetCoins disagrees with the model for %s", step, op, a)
			for _, d := range denoms {
				require.Equal(t, model[a][d], env.bankk.GetCoin(ctx, a, d),
					"step %d (%s): GetCoin(%s) disagrees with the model", step, op, d)
			}
		}
		// Structural invariants, after every operation including the rejected ones.
		// This runs both ways: the random walk is a fuzzer for the invariants — 500
		// keeper-produced states that must all be reported healthy, which is the
		// cheap defence against a check that fires on legitimate state — and the
		// invariants are extra oracles for the keeper, seeing what the map model
		// structurally cannot: whether a balance is in the right tier, whether its
		// key is well formed, and whether an address holding one can still sign.
		if msg, broken := AllInvariants(env.bankk.ViewKeeper)(ctx); broken {
			t.Fatalf("step %d (%s): invariant broken:\n%s", step, op, msg)
		}
		if msg, broken := auth.AllInvariants(env.acck)(ctx); broken {
			t.Fatalf("step %d (%s): auth invariant broken:\n%s", step, op, msg)
		}
		// The vesting lock is a separate invariant from conservation: dropping the
		// check entirely would keep every balance consistent with the model, so
		// only this assertion can catch it.
		require.GreaterOrEqual(t, env.bankk.GetCoin(ctx, vester, testAccountDenom),
			int64(vesterLocked),
			"step %d (%s): spent into the locked portion via a restricted path", step, op)
		// Per-denom supply must be conserved by transfers and match the model
		// for mints and burns.
		for _, d := range denoms {
			var want, got int64
			for _, a := range addrs {
				want += model[a][d]
				got += env.bankk.GetCoin(ctx, a, d)
			}
			require.Equal(t, want, got, "step %d (%s): total supply of %s diverged", step, op, d)
		}
	}

	// pick builds a valid std.Coins over a random subset of denoms. scale
	// controls how often an amount exceeds what the address holds, so that both
	// the success and the rejection paths are exercised.
	// Fixed seed: a failure must be reproducible from the reported step number.
	rng := rand.New(rand.NewSource(1))
	pick := func(a crypto.Address, overspend bool) std.Coins {
		var out std.Coins
		for _, d := range denoms { // denoms is not sorted; sort below
			if rng.Intn(3) == 0 {
				continue
			}
			amt := int64(rng.Intn(50) + 1)
			if overspend {
				amt += model[a][d] + 1
			}
			out = append(out, std.Coin{Denom: d, Amount: amt})
		}
		out.Sort()
		if !out.IsValid() {
			return nil
		}
		return out
	}

	// probeVestingLock attempts to spend exactly one coin more of the gas denom
	// than the schedule allows, and requires refusal. A directed probe rather than
	// a passive invariant: the random walk credits the vester often enough that
	// its balance drifts clear of the locked amount, so a breach is almost never
	// attempted — measured 17 successful vester debits with none of them binding.
	// Leads with an unlocked split-tier denom when one is held, so that a check
	// which stops at the first unlocked denom is caught too.
	probeVestingLock := func(t *testing.T, step int) {
		t.Helper()
		// Guarantee an unlocked denom to lead with. Without one the probe degenerates
		// to a single locked denom, and a check that stops at the first unlocked
		// denom would survive it.
		if env.bankk.GetCoin(ctx, vester, testRealmDenom) == 0 {
			require.NoError(t, env.bankk.MintCoins(ctx, vester,
				std.Coins{{Denom: testRealmDenom, Amount: 1}}))
			model[vester][testRealmDenom]++
		}
		spendable := env.bankk.GetCoin(ctx, vester, testAccountDenom) - vesterLocked
		amt := std.Coins{
			{Denom: testRealmDenom, Amount: 1},
			{Denom: testAccountDenom, Amount: spendable + 1},
		}
		amt.Sort()
		require.True(t, amt[0].Denom == testRealmDenom,
			"the unlocked denom must sort first for this probe to bind")
		require.Error(t, env.bankk.SubtractCoins(ctx, vester, amt),
			"step %d: spending %s must be refused; only %d of %s is unlocked",
			step, amt, spendable, testAccountDenom)
		check(t, step, "vesting probe (rejected)")
	}

	for step := range 500 {
		if step%25 == 0 {
			probeVestingLock(t, step)
		}
		from := addrs[rng.Intn(len(addrs))]
		to := addrs[rng.Intn(len(addrs))]
		overspend := rng.Intn(4) == 0

		switch rng.Intn(5) {
		case 0: // credit
			amt := pick(from, false)
			if amt == nil {
				continue
			}
			// MintCoins, not AddCoins: this arm creates value, so it must move the
			// supply counter or SupplyInvariant will (correctly) report the drift.
			require.NoError(t, env.bankk.MintCoins(ctx, from, amt))
			for _, c := range amt {
				model[from][c.Denom] += c.Amount
			}
			check(t, step, "MintCoins")

		case 1: // debit, sometimes unaffordable
			amt := pick(from, overspend)
			if amt == nil {
				continue
			}
			if err := env.bankk.BurnCoins(ctx, from, amt); err != nil {
				check(t, step, "BurnCoins(rejected)") // must have written nothing
				continue
			}
			for _, c := range amt {
				model[from][c.Denom] -= c.Amount
			}
			check(t, step, "BurnCoins")

		case 2: // transfer, sometimes unaffordable
			amt := pick(from, overspend)
			if amt == nil {
				continue
			}
			if err := env.bankk.SendCoins(ctx, from, to, amt); err != nil {
				check(t, step, "SendCoins(rejected)")
				continue
			}
			// A send to yourself debits then credits the same amount back, so it
			// nets to zero — the affordability check still runs against the
			// pre-debit balance.
			if from != to {
				for _, c := range amt {
					model[from][c.Denom] -= c.Amount
					model[to][c.Denom] += c.Amount
				}
			}
			check(t, step, "SendCoins")

		case 4: // the MsgMultiSend path, which has its own debit/credit loop
			amt := pick(from, overspend)
			if amt == nil || from == to {
				continue
			}
			in := []Input{{Address: from, Coins: amt}}
			out := []Output{{Address: to, Coins: amt}}
			if err := env.bankk.InputOutputCoins(ctx, in, out); err != nil {
				check(t, step, "InputOutputCoins(rejected)")
				continue
			}
			for _, c := range amt {
				model[from][c.Denom] -= c.Amount
				model[to][c.Denom] += c.Amount
			}
			check(t, step, "InputOutputCoins")

		case 3: // replace the whole balance set
			if from == vester {
				continue // unrestricted: allowed to spend into the locked portion
			}
			// SetCoins is replace-all and deliberately supply-blind (it cannot
			// compute its own delta — see RecomputeSupply), so reseed after it.
			amt := pick(from, false)
			if amt == nil {
				continue
			}
			require.NoError(t, env.bankk.SetCoins(ctx, from, amt))
			model[from] = map[string]int64{}
			for _, c := range amt {
				model[from][c.Denom] = c.Amount
			}
			env.bankk.RecomputeSupply(ctx)
			check(t, step, "SetCoins")
		}
	}

	// The run must actually have moved money, or the loop proved nothing.
	var total int64
	for _, a := range addrs {
		for _, d := range denoms {
			total += model[a][d]
		}
	}
	require.Positive(t, total, "the run ended with every balance zero")
}

// TestGetCoinCostIsFlat pins the property that justifies exposing a per-denom
// read to Gno at all: reading one denom must not get more expensive as the
// address accumulates others, whereas reading them all must.
//
// Realms had no way to do this before — the only Gno-visible accessor was
// GetCoins — so a realm interested in one denom paid for every denom the address
// had ever been sent, which it does not control.
//
// Measured comparatively, with the same number of keys in the store either way
// and only their owner differing. An absolute figure would not be flat here: the
// tm2 default gas config scales a read by tree depth, so any read costs more in a
// bigger store. gno.land pins depth flat (Fixed*Depth100), so on chain the two
// figures coincide exactly; holding the key count equal is what isolates the
// property from that artifact.
func TestGetCoinCostIsFlat(t *testing.T) {
	t.Parallel()

	const junk = 200
	const target = "/gno.land/r/spam/p0:junk"

	// onTarget: the address being read holds all the junk.
	// elsewhere: a bystander holds it, so the store is the same size but the
	// address being read holds only the one denom.
	measure := func(t *testing.T, onTarget bool, readAll bool) store.Gas {
		t.Helper()
		env := setupTestEnv()
		addr := crypto.AddressFromPreimage([]byte("holder"))
		bystander := crypto.AddressFromPreimage([]byte("bystander"))
		require.NoError(t, env.bankk.SetCoins(env.ctx, addr, std.Coins{{Denom: "ugnot", Amount: 1}}))
		require.NoError(t, env.bankk.AddCoins(env.ctx, addr, std.Coins{{Denom: target, Amount: 1}}))
		require.NoError(t, env.bankk.AddCoins(env.ctx, bystander, std.Coins{{Denom: "ugnot", Amount: 1}}))

		holder := bystander
		if onTarget {
			holder = addr
		}
		for i := 1; i < junk; i++ {
			denom := fmt.Sprintf("/gno.land/r/spam/p%d:junk", i)
			require.NoError(t, env.bankk.AddCoins(env.ctx, holder, std.Coins{{Denom: denom, Amount: 1}}))
		}

		meter := store.NewGasMeter(100_000_000_000)
		cctx, _ := env.ctx.CacheContext()
		cctx = cctx.WithGasMeter(meter)
		before := meter.GasConsumed()
		if readAll {
			env.bankk.GetCoins(cctx, addr)
		} else {
			env.bankk.GetCoin(cctx, addr, target)
		}
		return meter.GasConsumed() - before
	}

	oneLoaded := measure(t, true, false)
	oneClean := measure(t, false, false)
	allLoaded := measure(t, true, true)
	allClean := measure(t, false, true)

	require.Positive(t, oneLoaded, "the harness must actually charge gas")
	require.Equal(t, oneClean, oneLoaded,
		"reading one denom must cost the same whether or not the address holds %d "+
			"others (loaded=%d clean=%d)", junk, oneLoaded, oneClean)
	require.Greater(t, allLoaded, allClean,
		"reading every denom must get more expensive when the address holds more "+
			"(loaded=%d clean=%d)", allLoaded, allClean)
	require.Less(t, oneLoaded, allLoaded,
		"the per-denom read must be cheaper than enumerating (%d vs %d)", oneLoaded, allLoaded)
}

// A denom must live in exactly one tier. splitCoins asserts one direction (a
// split key for an allowlisted denom); these cover the other, which is the
// direction the allowlist is expected to move — shrinking it is how a chain
// migrates to a fully split layout, and running that binary against unmigrated
// state leaves account-object balances the keeper believes are split.
func TestAccountObjectHoldingASplitDenomFailsLoudly(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx
	addr := crypto.AddressFromPreimage([]byte("unmigrated"))

	// Reach past the keeper to build the state a shrunk allowlist would leave.
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	require.NoError(t, acc.SetCoins(std.NewCoins(
		std.NewCoin(testAccountDenom, 100),
		std.NewCoin("stranded", 500),
	)))
	env.acck.SetAccount(ctx, acc)

	// Enumerating must not report a balance that cannot be spent...
	require.PanicsWithValue(t,
		`denom "stranded" is in the account object but not in the account tier: `+
			`the allowlist changed without migrating existing balances`,
		func() { env.bankk.GetCoins(ctx, addr) })

	// ...and neither may the O(1) point read, even for a denom that is correctly
	// in the account tier: silently returning a wrong balance is the failure mode
	// this guard exists to prevent.
	require.Panics(t, func() { env.bankk.GetCoin(ctx, addr, testAccountDenom) })
}

func TestAccountTierAllowlistIsValidatedAtConstruction(t *testing.T) {
	t.Parallel()

	// A realm-issuable denom in the account tier would reinstate the unbounded
	// account-object growth the split exists to close, so the chain must refuse
	// to start rather than accept it.
	require.PanicsWithValue(t,
		`account-tier denom "/gno.land/r/x/tok:c" is realm-issuable`,
		func() { NewViewKeeper(auth.AccountKeeper{}, nil, []string{"/gno.land/r/x/tok:c"}) })

	require.Panics(t, func() { NewViewKeeper(auth.AccountKeeper{}, nil, []string{"UPPERCASE"}) })
	require.Panics(t, func() { NewViewKeeper(auth.AccountKeeper{}, nil, []string{""}) })
	require.NotPanics(t, func() { NewViewKeeper(auth.AccountKeeper{}, nil, []string{"ugnot"}) })
}

// Why GetCoin exists, and why realms should reach for banker.GetCoin rather
// than GetCoins().AmountOf(): a repeated point read is refunded by the cache
// store, but a repeated GetCoins is not, because splitCoins opens an iterator and
// iterator opens are charged every time. A realm reading balances in a loop
// therefore pays per iteration, where before the split it paid for one account
// read and got the rest free.
func TestRepeatedGetCoinsIsChargedButRepeatedGetCoinIsFree(t *testing.T) {
	t.Parallel()

	measure := func(t *testing.T, readAll bool) (first, repeat store.Gas) {
		t.Helper()
		env := setupTestEnv()
		addr := crypto.AddressFromPreimage([]byte("looper"))
		require.NoError(t, env.bankk.SetCoins(env.ctx, addr, std.Coins{
			{Denom: testRealmDenom, Amount: 100},
			{Denom: testAccountDenom, Amount: 100},
		}))

		meter := store.NewGasMeter(100_000_000_000)
		cctx, _ := env.ctx.CacheContext()
		cctx = cctx.WithGasMeter(meter)
		read := func() {
			if readAll {
				env.bankk.GetCoins(cctx, addr)
			} else {
				env.bankk.GetCoin(cctx, addr, testRealmDenom)
			}
		}
		read()
		first = meter.GasConsumed()
		read()
		return first, meter.GasConsumed() - first
	}

	_, repeatAll := measure(t, true)
	_, repeatOne := measure(t, false)

	require.Zero(t, repeatOne, "a repeated point read must be refunded by the cache store")
	require.Positive(t, repeatAll,
		"a repeated GetCoins must still be charged for the iterator open; if this "+
			"ever becomes free, the case for migrating realm callers to GetCoin weakens")
}

// The reported flake's first symptom was encodeBalance panicking on a negative
// balance while the guard three lines above compared the same two values. That is
// producible only if a non-positive amount reaches subtract: the guard is
// `old < coin.Amount`, which a negative amount passes, and `old - coin.Amount`
// then credits — or overflows to a negative and panics. Prove no public path
// admits one.
func TestNoPathAdmitsANonPositiveDebit(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx
	from := crypto.AddressFromPreimage([]byte("from"))
	to := crypto.AddressFromPreimage([]byte("to"))
	require.NoError(t, env.bankk.SetCoins(ctx, from, std.Coins{
		{Denom: testRealmDenom, Amount: 1000},
		{Denom: testAccountDenom, Amount: 1000},
	}))

	hostile := []std.Coins{
		{{Denom: testRealmDenom, Amount: -1}},
		{{Denom: testRealmDenom, Amount: math.MinInt64}},
		{{Denom: testAccountDenom, Amount: -1}},
		{{Denom: testAccountDenom, Amount: math.MinInt64}},
		{{Denom: testRealmDenom, Amount: -1}, {Denom: testAccountDenom, Amount: 1}},
		{{Denom: testRealmDenom, Amount: 1}, {Denom: testAccountDenom, Amount: math.MinInt64}},
	}
	for hi, amt := range hostile {
		for _, call := range []struct {
			name string
			fn   func() error
		}{
			{"SubtractCoins", func() error { return env.bankk.SubtractCoins(ctx, from, amt) }},
			{"subtractUnrestricted", func() error { return env.bankk.subtractCoinsUnrestricted(ctx, from, amt) }},
			{"SendCoins", func() error { return env.bankk.SendCoins(ctx, from, to, amt) }},
			{"InputOutputCoins", func() error {
				return env.bankk.InputOutputCoins(ctx,
					[]Input{{Address: from, Coins: amt}}, []Output{{Address: to, Coins: amt}})
			}},
		} {
			require.NotPanics(t, func() {
				require.Error(t, call.fn(), "case %d: %s must reject %#v", hi, call.name, amt)
			}, "%s panicked on %s", call.name, amt)
		}
	}

	// The guard must be in subtract itself, not only in its callers. Called
	// directly, bypassing the caller validation every public path performs: a
	// negative amount passes `old < coin.Amount` (old is never negative), so
	// without this the debit credits, and a large enough one overflows to a
	// negative that only encodeBalance catches — which is precisely the symptom
	// reported as an intermittent failure.
	for _, bad := range []int64{-1, 0, math.MinInt64} {
		acc := env.acck.GetAccount(ctx, from)
		require.NotPanics(t, func() {
			require.Error(t, env.bankk.subtract(ctx, acc, from,
				std.Coins{{Denom: testRealmDenom, Amount: bad}}, false),
				"subtract must reject a %d debit on its own", bad)
		})
		require.Equal(t, int64(1000), env.bankk.GetCoin(ctx, from, testRealmDenom),
			"a rejected %d debit must not move the balance", bad)
	}

	// The credit path too: AddCoins defends itself with Coins.IsValid rather than
	// an internal guard, so that gate is load-bearing and must be pinned. Without
	// it a negative credit silently debits.
	for _, bad := range []int64{-1, math.MinInt64} {
		for _, d := range []string{testRealmDenom, testAccountDenom} {
			require.NotPanics(t, func() {
				require.Error(t, env.bankk.AddCoins(ctx, from, std.Coins{{Denom: d, Amount: bad}}),
					"AddCoins must reject a %d credit of %s", bad, d)
			})
		}
	}
	require.Equal(t, int64(1000), env.bankk.GetCoin(ctx, from, testRealmDenom))

	// A zero amount is a documented no-op on SendCoins (pre-existing), not an
	// error. It must still never reach a write.
	require.NotPanics(t, func() {
		require.NoError(t, env.bankk.SendCoins(ctx, from, to,
			std.Coins{{Denom: testRealmDenom, Amount: 0}}))
	})

	// Nothing may have moved.
	require.Equal(t, int64(1000), env.bankk.GetCoin(ctx, from, testRealmDenom))
	require.Equal(t, int64(1000), env.bankk.GetCoin(ctx, from, testAccountDenom))
	require.Zero(t, env.bankk.GetCoin(ctx, to, testRealmDenom))
}

// Both tiers must reject a non-positive amount, on both directions of the money
// path. Neither defends itself otherwise: the split tier's check is
// `old < coin.Amount`, which a negative passes, and the account tier's
// SubUnsafe+IsValid validates the *result*, which a negative debit leaves
// perfectly valid but larger. Driven through the internal functions, since every
// public entry point validates first — that caller-side check is exactly what
// this is not allowed to depend on.
func TestNeitherTierInvertsOnANonPositiveAmount(t *testing.T) {
	t.Parallel()

	for _, tier := range []struct{ name, denom string }{
		{"account", testAccountDenom},
		{"split", testRealmDenom},
	} {
		for _, bad := range []int64{-1, 0, math.MinInt64} {
			env := setupTestEnv()
			ctx := env.ctx
			addr := crypto.AddressFromPreimage([]byte("holder"))
			require.NoError(t, env.bankk.SetCoins(ctx, addr,
				std.Coins{{Denom: tier.denom, Amount: 100}}))
			amt := std.Coins{{Denom: tier.denom, Amount: bad}}

			require.NotPanics(t, func() {
				require.Error(t, env.bankk.subtract(ctx, env.acck.GetAccount(ctx, addr), addr, amt, false),
					"%s tier: debit of %d must be refused", tier.name, bad)
			})
			require.NotPanics(t, func() {
				require.Error(t, env.bankk.AddCoins(ctx, addr, amt),
					"%s tier: credit of %d must be refused", tier.name, bad)
			})
			require.Equal(t, int64(100), env.bankk.GetCoin(ctx, addr, tier.denom),
				"%s tier: balance moved on a rejected %d", tier.name, bad)
		}
	}
}

// A denom longer than any that can exist must not reach the store. Nothing can
// hold one, so zero is the exact answer — and the length is the only bound on the
// work a caller can provoke: the tier lookup hashes the whole string, the key
// copies it, and the cache store retains that copy for the transaction while
// charging a flat rate. banker.GetCoin hands this a realm-supplied string.
func TestOverlongDenomNeverReachesTheStore(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	addr := crypto.AddressFromPreimage([]byte("holder"))
	require.NoError(t, env.bankk.SetCoins(env.ctx,
		addr, std.Coins{{Denom: testAccountDenom, Amount: 1}}))

	huge := "/gno.land/r/x/" + strings.Repeat("z", 1<<20) + ":c"
	meter := store.NewGasMeter(100_000_000_000)
	cctx, _ := env.ctx.CacheContext()
	cctx = cctx.WithGasMeter(meter)
	require.Zero(t, env.bankk.GetCoin(cctx, addr, huge))
	require.Zero(t, meter.GasConsumed(),
		"an impossible denom must not reach the store at all")
}

// parseBalanceKey checks the prefix where denomFromBalanceKey does not, because the
// invariants hand it keys from a sweep of the whole store rather than from an
// iterator already bounded by BalancePrefix.
func TestParseBalanceKeyRequiresThePrefix(t *testing.T) {
	t.Parallel()

	addr := crypto.AddressFromPreimage([]byte("holder"))
	foreign := append([]byte("/a/"), addr[:]...)
	_, _, err := parseBalanceKey(append(foreign, testRealmDenom...))
	require.Error(t, err, "a key outside /b/ must not be read as a balance key")
	require.Contains(t, err.Error(), BalancePrefix)
}

// refusingAccount holds coins but will not store them. No shipped account type
// does both — BaseSessionAccount refuses SetCoins but also reports no coins, so
// subtract fails its solvency check before reaching any write, which makes it
// useless for testing write order.
type refusingAccount struct{ std.Account }

func (refusingAccount) SetCoins(std.Coins) error { return errors.New("refuses to hold coins") }

// subtract writes the account tier before the split keys, because the account
// write is the only step that can still fail. With the order reversed, a failing
// account write leaves the split debits committed and returns an error — value
// destroyed by an operation that reported failure.
func TestSubtractWritesNothingWhenTheAccountWriteFails(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx
	addr := crypto.AddressFromPreimage([]byte("refuser"))
	require.NoError(t, env.bankk.SetCoins(ctx, addr, std.Coins{
		{Denom: testRealmDenom, Amount: 100},
		{Denom: testAccountDenom, Amount: 100},
	}))

	err := env.bankk.subtract(ctx, refusingAccount{env.acck.GetAccount(ctx, addr)}, addr, std.Coins{
		{Denom: testRealmDenom, Amount: 40},
		{Denom: testAccountDenom, Amount: 40},
	}, false)
	require.Error(t, err, "the account write must fail")
	require.Equal(t, int64(100), env.bankk.getSplitBalance(ctx, addr, testRealmDenom),
		"a failed subtract must not have debited the split tier")
}

// SetCoins replaces the account tier before touching the split keys, for the same
// reason subtract does. It cannot use refusingAccount, which SetCoins never sees —
// it reads the account itself — so this uses the only shipped type that refuses
// coins. A session account filed at a regular address path is itself a fault the
// auth keyspace invariant reports, and that is the point: a call that returns an
// error must not have moved money, even out of a state that should not exist.
func TestSetCoinsWritesNothingWhenTheAccountWriteFails(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx
	master := crypto.AddressFromPreimage([]byte("session-master"))
	_, sessPub, _ := tu.KeyTestPubAddr()
	sess := env.acck.NewSessionAccount(ctx, master, sessPub)
	addr := sess.GetAddress()

	// A split-only credit does not touch the account object's coins, so it cannot
	// fail here and leaves a balance for SetCoins to destroy.
	require.NoError(t, env.bankk.AddCoins(ctx, addr, std.Coins{{Denom: testRealmDenom, Amount: 100}}))
	env.acck.SetAccount(ctx, sess)

	err := env.bankk.SetCoins(ctx, addr, std.Coins{
		{Denom: testRealmDenom, Amount: 50},
		{Denom: testAccountDenom, Amount: 10},
	})
	require.Error(t, err, "a session account must refuse to hold the account-tier coins")
	require.Equal(t, int64(100), env.bankk.getSplitBalance(ctx, addr, testRealmDenom),
		"a failed SetCoins must not have replaced the split tier")
}

// The mirror of TestAccountObjectHoldingASplitDenomFailsLoudly: a split key for a
// denom that IS in the account tier. This is the direction the allowlist grows —
// adding a denom to it without migrating existing keys — where GetCoins would
// otherwise sum both homes and report a balance GetCoin disagrees with.
func TestSplitKeyForAnAccountTierDenomFailsLoudly(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx
	addr := crypto.AddressFromPreimage([]byte("grown"))

	// Reach past the keeper: write a split key for the gas denom, which the tier
	// router would never do.
	env.bankk.setSplitBalance(ctx, addr, testAccountDenom, 500)

	require.PanicsWithValue(t,
		`denom "`+testAccountDenom+`" has a split-tier key but is in the account tier: `+
			`the allowlist changed without migrating existing balances`,
		func() { env.bankk.GetCoins(ctx, addr) })
}

// The operator-visible half of the same mis-migration: before anything calls
// GetCoins and panics, the balance is simply frozen. Pinned because the failure mode
// if a future reader "fixes" the exclusivity assertion by summing both tiers is not
// a louder error — it is 500 spendable coins that the invariants call corrupt.
func TestAMisMigratedBalanceIsFrozenNotSpendable(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx
	holder := crypto.AddressFromPreimage([]byte("frozen"))
	dest := crypto.AddressFromPreimage([]byte("dest"))
	require.NoError(t, env.bankk.SetCoins(ctx, dest, std.Coins{{Denom: testAccountDenom, Amount: 1}}))

	// The state a node restarted on the old database after the allowlist grew: the
	// gas denom's balance is still under /b/, where the router no longer looks.
	env.bankk.setSplitBalance(ctx, holder, testAccountDenom, 500)

	err := env.bankk.SendCoins(ctx, holder, dest, std.Coins{{Denom: testAccountDenom, Amount: 10}})
	require.Error(t, err, "a balance in the wrong tier must not be spendable")
	// Asserted on the message rather than the type: the keeper returns it wrapped,
	// and the wrapper is what a caller reads. "coins" not "funds" — gnokey reports
	// InsufficientCoinsError and InsufficientFundsError differently.
	require.ErrorContains(t, err, "insufficient coins error")
	require.Equal(t, int64(500), env.bankk.getSplitBalance(ctx, holder, testAccountDenom),
		"and nothing was taken from the stranded balance")

	// A credit lands in the tier the router now prefers, giving the denom two homes.
	require.NoError(t, env.bankk.AddCoins(ctx, holder, std.Coins{{Denom: testAccountDenom, Amount: 7}}))
	require.Equal(t, "7"+testAccountDenom, env.acck.GetAccount(ctx, holder).GetCoins().String())
	require.Equal(t, int64(500), env.bankk.getSplitBalance(ctx, holder, testAccountDenom))
}

// The allowlist holds one denom today, so nothing otherwise exercises a second.
// A second gas denom is the expected way it grows — a chain accepting fees in an
// IBC voucher, for instance — and such a denom sorts *before* "ugnot", which makes
// this the first case where the account tier is neither a prefix nor a suffix of
// the sorted result. GetCoins must interleave three-ways: split, account, account.
func TestSecondGasDenomInAccountTier(t *testing.T) {
	t.Parallel()

	// Lowercase, because ValidateDenom's charset is lowercase-only — a cosmos-style
	// uppercase-hex IBC hash would have to be normalised before reaching here.
	const voucher = "ibc/" + "a1b2c3d4a1b2c3d4a1b2c3d4a1b2c3d4a1b2c3d4a1b2c3d4a1b2c3d4a1b2c3d4"
	require.NoError(t, std.ValidateDenom(voucher))
	require.Less(t, voucher, testAccountDenom, "the point of this test is that it sorts first")

	env := setupTestEnv()
	ctx := env.ctx
	k := NewBankKeeper(env.acck, env.prmk.ForModule(ModuleName), env.key,
		[]string{testAccountDenom, voucher})
	addr := crypto.AddressFromPreimage([]byte("holder"))

	// The split-tier denoms must straddle the account-tier ones. If every split
	// denom sorted before every account one, concatenating the two tiers would
	// coincidentally produce sorted output and the merge would not be under test.
	amt := std.Coins{
		{Denom: testRealmDenom, Amount: 5},    // split,   sorts first  ("/")
		{Denom: voucher, Amount: 20},          // account, sorts second ("i")
		{Denom: testAccountDenom, Amount: 50}, // account, sorts third  ("u")
		{Denom: "zeta", Amount: 7},            // split,   sorts last   ("z")
	}
	amt.Sort()
	require.NoError(t, k.SetCoins(ctx, addr, amt))

	require.Equal(t, amt.String(), k.GetCoins(ctx, addr).String(),
		"GetCoins must interleave the tiers, not concatenate them")
	require.True(t, k.GetCoins(ctx, addr).IsValid())

	// Both account-tier denoms live in the account object; the realm one does not.
	require.Equal(t, std.Coins{{Denom: voucher, Amount: 20}, {Denom: testAccountDenom, Amount: 50}}.String(),
		env.acck.GetAccount(ctx, addr).GetCoins().String())
	require.Zero(t, k.getSplitBalance(ctx, addr, voucher),
		"an allowlisted denom must not also have a split key")

	require.Equal(t, int64(20), k.GetCoin(ctx, addr, voucher))
	require.Equal(t, int64(5), k.GetCoin(ctx, addr, testRealmDenom))
	require.Equal(t, int64(7), k.GetCoin(ctx, addr, "zeta"))

	// And it spends like any other account-tier denom.
	to := crypto.AddressFromPreimage([]byte("dest"))
	require.NoError(t, k.SendCoins(ctx, addr, to, std.Coins{{Denom: voucher, Amount: 3}}))
	require.Equal(t, int64(17), k.GetCoin(ctx, addr, voucher))
	require.Equal(t, int64(3), k.GetCoin(ctx, to, voucher))
}
